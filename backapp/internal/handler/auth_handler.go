package handler

import (
	"backapp/internal/config"
	"backapp/internal/middleware"
	"backapp/internal/models"
	"backapp/internal/repository"
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"google.golang.org/api/idtoken"
)

const (
	oauthStateCookieName = "oauthstate"
	oauthNonceCookieName = "oauthnonce"
	csrfTokenCookieName  = "csrf_token"
)

type googleIDTokenValidator interface {
	Validate(context.Context, string, string) (*idtoken.Payload, error)
}

type officialGoogleIDTokenValidator struct{}

func (officialGoogleIDTokenValidator) Validate(ctx context.Context, rawToken, audience string) (*idtoken.Payload, error) {
	return idtoken.Validate(ctx, rawToken, audience)
}

type AuthHandler struct {
	cfg          *config.Config
	userRepo     repository.UserRepository
	eventRepo    repository.EventRepository
	classRepo    repository.ClassRepository
	oauth2Config *oauth2.Config
	idTokens     googleIDTokenValidator
}

func NewAuthHandler(cfg *config.Config, userRepo repository.UserRepository, eventRepo repository.EventRepository, classRepo repository.ClassRepository) *AuthHandler {
	return &AuthHandler{
		cfg:       cfg,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		classRepo: classRepo,
		idTokens:  officialGoogleIDTokenValidator{},
		oauth2Config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}
}

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if isLINEInAppBrowser(c.GetHeader("User-Agent")) {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=line_inapp_browser_unsupported")
		return
	}

	state, err := setOAuthRandomCookie(c.Writer, c.Request, oauthStateCookieName)
	if err != nil {
		log.Printf("[auth] failed to generate OAuth state: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	nonce, err := setOAuthRandomCookie(c.Writer, c.Request, oauthNonceCookieName)
	if err != nil {
		log.Printf("[auth] failed to generate OAuth nonce: %T", err)
		clearOAuthCookie(c.Writer, c.Request, oauthStateCookieName)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	url := h.oauth2Config.AuthCodeURL(state, oauth2.SetAuthURLParam("nonce", nonce))
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if oauthError := c.Query("error"); oauthError != "" {
		clearOAuthCookies(c.Writer, c.Request)
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error="+url.QueryEscape(oauthError))
		return
	}

	oauthState, cookieErr := c.Cookie(oauthStateCookieName)
	oauthNonce, nonceCookieErr := c.Cookie(oauthNonceCookieName)
	queryState := c.Query("state")
	if cookieErr != nil || queryState == "" || subtle.ConstantTimeCompare([]byte(queryState), []byte(oauthState)) != 1 {
		log.Printf(
			"[auth] invalid oauth state: cookie_present=%v request_host=%s forwarded_host=%s forwarded_proto=%s query_state_len=%d cookie_state_len=%d",
			cookieErr == nil,
			c.Request.Host,
			c.Request.Header.Get("X-Forwarded-Host"),
			c.Request.Header.Get("X-Forwarded-Proto"),
			len(c.Query("state")),
			len(oauthState),
		)
		clearOAuthCookies(c.Writer, c.Request)
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=invalid_state")
		return
	}
	if nonceCookieErr != nil || oauthNonce == "" {
		clearOAuthCookies(c.Writer, c.Request)
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=invalid_nonce")
		return
	}
	// State and nonce are one-time values. Clear them before any network call
	// so a callback cannot be replayed after a partial failure.
	clearOAuthCookies(c.Writer, c.Request)

	code := c.Query("code")
	requestContext, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
	defer cancel()
	token, err := h.oauth2Config.Exchange(requestContext, code)
	if err != nil {
		log.Printf("[auth] OAuth token exchange failed: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	email, err := verifiedGoogleIDTokenEmail(requestContext, h.idTokens, token.Extra("id_token"), h.cfg.GoogleClientID, oauthNonce)
	if err != nil {
		log.Printf("[auth] Google ID token validation failed: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// ドメイン制限
	allowedDomains := []string{"sendai-nct.jp", "sendai-nct.ac.jp"}
	parts := strings.Split(email, "@")
	if len(parts) != 2 || parts[1] == "" {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=domain_not_allowed")
		return
	}
	emailDomain := parts[1]
	isAllowed := false
	for _, domain := range allowedDomains {
		if emailDomain == domain {
			isAllowed = true
			break
		}
	}
	if !isAllowed {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=domain_not_allowed")
		return
	}

	user, err := h.userRepo.GetUserByEmail(email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user by email"})
		return
	}

	if user == nil {
		// 新規ユーザー: InitRootUserと一致する場合はrootロール、それ以外はstudent
		role := "student"
		if h.cfg != nil && h.cfg.InitRootUser == email {
			role = "root"
		}

		newUser := &models.User{
			ID:                uuid.New().String(),
			Email:             email,
			IsProfileComplete: false,
		}
		if err := h.userRepo.CreateUser(newUser, role); err != nil {
			log.Printf("CreateUser error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		user = newUser
	}

	// Create session
	sessionToken := uuid.New().String()
	csrfToken, err := generateSecureRandomToken(32)
	if err != nil {
		log.Printf("[auth] Failed to generate CSRF token: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}
	if err := middleware.CreateSession(sessionToken, user.ID, csrfToken); err != nil {
		log.Printf("[auth] Failed to create session: %T", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// Set session and CSRF cookies only after both values are persisted.
	sessionExpiration := time.Now().Add(24 * time.Hour)
	setSessionTokenCookie(c.Writer, c.Request, sessionToken, sessionExpiration)
	setCSRFTokenCookie(c.Writer, c.Request, csrfToken, sessionExpiration)

	// Add a debug log to verify cookie flags
	log.Printf("[auth] Session created for user %s, secure=%v, origin=%s", user.Email, shouldUseSecureCookie(c.Request), c.Request.RemoteAddr)

	c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/dashboard")
}

func verifiedGoogleIDTokenEmail(ctx context.Context, validator googleIDTokenValidator, rawIDToken any, expectedAudience, expectedNonce string) (string, error) {
	idToken, ok := rawIDToken.(string)
	if !ok || idToken == "" {
		return "", fmt.Errorf("id_token is missing")
	}
	if expectedAudience == "" {
		return "", fmt.Errorf("expected audience is missing")
	}
	if expectedNonce == "" {
		return "", fmt.Errorf("expected nonce is missing")
	}

	payload, err := validator.Validate(ctx, idToken, expectedAudience)
	if err != nil {
		return "", fmt.Errorf("validate id_token: %w", err)
	}
	if payload == nil {
		return "", fmt.Errorf("id_token payload is missing")
	}
	if payload.Issuer != "accounts.google.com" && payload.Issuer != "https://accounts.google.com" {
		return "", fmt.Errorf("id_token issuer is invalid")
	}
	if payload.Audience != expectedAudience {
		return "", fmt.Errorf("id_token audience does not match client id")
	}
	if payload.Subject == "" {
		return "", fmt.Errorf("id_token subject is missing")
	}

	nonce, nonceOK := payload.Claims["nonce"].(string)
	if !nonceOK || subtle.ConstantTimeCompare([]byte(nonce), []byte(expectedNonce)) != 1 {
		return "", fmt.Errorf("id_token nonce does not match")
	}
	if rawAuthorizedParty, exists := payload.Claims["azp"]; exists {
		authorizedParty, ok := rawAuthorizedParty.(string)
		if !ok || authorizedParty != expectedAudience {
			return "", fmt.Errorf("id_token authorized party does not match client id")
		}
	}

	email, emailOK := payload.Claims["email"].(string)
	if !emailOK || email == "" {
		return "", fmt.Errorf("id_token did not include email")
	}
	emailVerified, verifiedOK := payload.Claims["email_verified"].(bool)
	if !verifiedOK || !emailVerified {
		return "", fmt.Errorf("id_token email is not verified")
	}

	return email, nil
}

func (h *AuthHandler) GetUser(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	// ロール情報を含むユーザー情報を取得
	userWithRoles, err := h.userRepo.GetUserWithRoles(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	if userWithRoles == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Determine if this user is the initial root user and hasn't completed profile
	if h.cfg != nil && h.cfg.InitRootUser != "" {
		userWithRoles.IsInitRootFirstLogin = (h.cfg.InitRootUser == userWithRoles.Email) && !userWithRoles.IsProfileComplete
	} else {
		userWithRoles.IsInitRootFirstLogin = false
	}

	c.JSON(http.StatusOK, userWithRoles)
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	user := userCtx.(*models.User)

	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len([]rune(req.DisplayName)) > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表示名は12文字以内で入力してください"})
		return
	}

	user.DisplayName = &req.DisplayName
	if user.IsProfileComplete {
		if user.ClassID == nil || *user.ClassID != req.ClassID {
			c.JSON(http.StatusForbidden, gin.H{"error": "Class cannot be changed after profile completion"})
			return
		}
	} else {
		activeEventID, err := h.eventRepo.GetActiveEvent()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
			return
		}

		class, err := h.classRepo.GetClassByID(req.ClassID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class details"})
			return
		}
		if class == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Selected class not found"})
			return
		}
		if class.EventID != nil && *class.EventID != activeEventID {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Selected class does not belong to the active event"})
			return
		}
		user.ClassID = &req.ClassID
		user.IsProfileComplete = true
	}

	if err := h.userRepo.UpdateUser(user); err != nil {
		log.Printf("UpdateUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// ... (Logout and other handlers are unchanged)
func (h *AuthHandler) Logout(c *gin.Context) {
	cookie, err := c.Cookie("session_token")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "not logged in"})
		return
	}

	middleware.DeleteSession(cookie)

	expired := time.Now().Add(-1 * time.Hour)
	setSessionTokenCookie(c.Writer, c.Request, "", expired)
	setCSRFTokenCookie(c.Writer, c.Request, "", expired)

	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

func (h *AuthHandler) FindUsersHandler(c *gin.Context) {
	query := c.Query("query")
	searchType := c.Query("searchType")

	users, err := h.userRepo.FindUsers(query, searchType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find users"})
		return
	}

	c.JSON(http.StatusOK, users)
}

type UpdateUserDisplayNameRequest struct {
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

func (h *AuthHandler) UpdateUserDisplayNameByAdmin(c *gin.Context) {
	var req UpdateUserDisplayNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len([]rune(req.DisplayName)) > 12 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "表示名は12文字以内で入力してください"})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(*models.User)

	userWithRoles, err := h.userRepo.GetUserWithRoles(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAuthorized := false
	for _, role := range userWithRoles.Roles {
		if role.Name == "root" || role.Name == "admin" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update user display name"})
		return
	}

	if err := h.userRepo.UpdateUserDisplayName(req.UserID, req.DisplayName); err != nil {
		log.Printf("UpdateUserDisplayName error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update display name"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User display name updated successfully"})
}

type UpdateUserRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Event  *int   `json:"event"`
}

func (h *AuthHandler) UpdateUserRoleByAdmin(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent modification of default roles
	defaultRoles := []string{"root", "admin", "student"}
	for _, r := range defaultRoles {
		if req.Role == r {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot assign default roles"})
			return
		}
	}

	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(*models.User)

	userWithRoles, err := h.userRepo.GetUserWithRoles(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAuthorized := false
	for _, role := range userWithRoles.Roles {
		if role.Name == "root" || role.Name == "admin" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to update user role"})
		return
	}

	if err := h.userRepo.UpdateUserRole(req.UserID, req.Role, req.Event); err != nil {
		log.Printf("UpdateUserRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

type PromoteUserRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"` // "student", "admin", or "root"
}

// PromoteUserByRoot はroot権限でユーザーのマスタロールを交換する
func (h *AuthHandler) PromoteUserByRoot(c *gin.Context) {
	var req PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != "student" && req.Role != "admin" && req.Role != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "設定できるマスタロールは 'student' / 'admin' / 'root' のみです"})
		return
	}

	if err := h.userRepo.ReplaceMasterRole(req.UserID, req.Role); err != nil {
		log.Printf("PromoteUserByRoot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to replace master role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Master role replaced successfully"})
}

// DemoteUserByRoot は廃止された付与・剥奪APIの互換エンドポイント
func (h *AuthHandler) DemoteUserByRoot(c *gin.Context) {
	c.JSON(http.StatusBadRequest, gin.H{"error": "マスタロールは剥奪ではなく交換してください"})
}

// isClassManagedRole returns true for class-managed roles that should not be
// deleted via role management. This includes {クラス名_競技名} and {クラス名_rep}
// formats — any role containing an underscore that is not a default role.
func isClassManagedRole(roleName string) bool {
	defaultRoles := []string{"root", "admin", "student"}
	for _, r := range defaultRoles {
		if roleName == r {
			return false
		}
	}
	for _, c := range roleName {
		if c == '_' {
			return true
		}
	}
	return false
}

func (h *AuthHandler) DeleteUserRoleByAdmin(c *gin.Context) {
	var req UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prevent modification of default roles
	defaultRoles := []string{"root", "admin", "student"}
	for _, r := range defaultRoles {
		if req.Role == r {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete default roles"})
			return
		}
	}

	// Prevent deletion of class-sport roles ({クラス名_競技名} format)
	// These roles must be managed via class/team management, not role management
	if isClassManagedRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete class sport roles. Please use class/team management page."})
		return
	}

	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(*models.User)

	userWithRoles, err := h.userRepo.GetUserWithRoles(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user roles"})
		return
	}

	isAuthorized := false
	for _, role := range userWithRoles.Roles {
		if role.Name == "root" || role.Name == "admin" {
			isAuthorized = true
			break
		}
	}

	if !isAuthorized {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to delete user role"})
		return
	}

	if err := h.userRepo.DeleteUserRole(req.UserID, req.Role); err != nil {
		log.Printf("DeleteUserRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role deleted successfully"})
}

func generateSecureRandomToken(size int) (string, error) {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("read secure random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(value), nil
}

func setOAuthRandomCookie(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	value, err := generateSecureRandomToken(32)
	if err != nil {
		return "", err
	}

	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().Add(20 * time.Minute),
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return value, nil
}

func setSessionTokenCookie(w http.ResponseWriter, r *http.Request, value string, expiration time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    value,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func setCSRFTokenCookie(w http.ResponseWriter, r *http.Request, value string, expiration time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     csrfTokenCookieName,
		Value:    value,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteStrictMode,
	})
}

func clearOAuthCookie(w http.ResponseWriter, r *http.Request, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearOAuthCookies(w http.ResponseWriter, r *http.Request) {
	clearOAuthCookie(w, r, oauthStateCookieName)
	clearOAuthCookie(w, r, oauthNonceCookieName)
}

func shouldUseSecureCookie(r *http.Request) bool {
	host := r.Host
	if forwardedHost := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-Host"), ",")[0]); forwardedHost != "" {
		host = forwardedHost
	}

	host = strings.ToLower(strings.TrimSpace(host))
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}
	host = strings.Trim(host, "[]")

	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return false
	}

	return middleware.IsRequestSecure(r)
}

func isLINEInAppBrowser(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	return strings.Contains(ua, "line/") || strings.Contains(ua, "liff")
}
