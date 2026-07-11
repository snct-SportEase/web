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
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type AuthHandler struct {
	cfg          *config.Config
	userRepo     repository.UserRepository
	eventRepo    repository.EventRepository
	classRepo    repository.ClassRepository
	oauth2Config *oauth2.Config
}

func NewAuthHandler(cfg *config.Config, userRepo repository.UserRepository, eventRepo repository.EventRepository, classRepo repository.ClassRepository) *AuthHandler {
	return &AuthHandler{
		cfg:       cfg,
		userRepo:  userRepo,
		eventRepo: eventRepo,
		classRepo: classRepo,
		oauth2Config: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			Scopes:       []string{"openid", "email", "https://www.googleapis.com/auth/userinfo.email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}
}

// ... (GoogleLogin, GoogleCallback, GetUser are unchanged)

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	if isLINEInAppBrowser(c.GetHeader("User-Agent")) {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=line_inapp_browser_unsupported")
		return
	}

	state := generateStateOauthCookie(c.Writer, c.Request)
	url := h.oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	if oauthError := c.Query("error"); oauthError != "" {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error="+url.QueryEscape(oauthError))
		return
	}

	oauthState, cookieErr := c.Cookie("oauthstate")
	if subtle.ConstantTimeCompare([]byte(c.Query("state")), []byte(oauthState)) != 1 {
		log.Printf(
			"[auth] invalid oauth state: cookie_present=%v request_host=%s forwarded_host=%s forwarded_proto=%s query_state_len=%d cookie_state_len=%d",
			cookieErr == nil,
			c.Request.Host,
			c.Request.Header.Get("X-Forwarded-Host"),
			c.Request.Header.Get("X-Forwarded-Proto"),
			len(c.Query("state")),
			len(oauthState),
		)
		clearOauthStateCookie(c.Writer, c.Request)
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=invalid_state")
		return
	}
	clearOauthStateCookie(c.Writer, c.Request)

	code := c.Query("code")
	token, err := h.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauth2 token exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	email, err := fetchGoogleUserEmail(context.Background(), token.AccessToken)
	if err != nil {
		log.Printf("failed to fetch google userinfo, falling back to id_token: %v", err)
		email, err = emailFromGoogleIDToken(token.Extra("id_token"), h.cfg.GoogleClientID)
		if err != nil {
			log.Printf("failed to read email from google id_token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
			return
		}
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
	middleware.CreateSession(sessionToken, user.ID)

	// Set session cookie
	setSessionTokenCookie(c.Writer, c.Request, sessionToken, time.Now().Add(24*time.Hour))

	// Add a debug log to verify cookie flags
	log.Printf("[auth] Session created for user %s, secure=%v, origin=%s", user.Email, shouldUseSecureCookie(c.Request), c.Request.RemoteAddr)

	c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/dashboard")
}

func fetchGoogleUserEmail(ctx context.Context, accessToken string) (string, error) {
	if accessToken == "" {
		return "", fmt.Errorf("access token is empty")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return "", fmt.Errorf("create userinfo request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("userinfo request failed: %w", err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read userinfo response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("userinfo returned %s: %s", response.Status, strings.TrimSpace(string(data)))
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		return "", fmt.Errorf("parse userinfo response: %w", err)
	}
	if userInfo.Email == "" {
		return "", fmt.Errorf("userinfo response did not include email")
	}

	return userInfo.Email, nil
}

func emailFromGoogleIDToken(rawIDToken any, expectedAudience string) (string, error) {
	idToken, ok := rawIDToken.(string)
	if !ok || idToken == "" {
		return "", fmt.Errorf("id_token is missing")
	}

	parts := strings.Split(idToken, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("id_token has invalid format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode id_token payload: %w", err)
	}

	var claims struct {
		Email         string          `json:"email"`
		EmailVerified bool            `json:"email_verified"`
		Audience      json.RawMessage `json:"aud"`
		ExpiresAt     int64           `json:"exp"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("parse id_token payload: %w", err)
	}

	if claims.Email == "" {
		return "", fmt.Errorf("id_token did not include email")
	}
	if !claims.EmailVerified {
		return "", fmt.Errorf("id_token email is not verified")
	}
	if claims.ExpiresAt != 0 && time.Now().Unix() > claims.ExpiresAt {
		return "", fmt.Errorf("id_token is expired")
	}
	if expectedAudience != "" && !idTokenAudienceMatches(claims.Audience, expectedAudience) {
		return "", fmt.Errorf("id_token audience does not match client id")
	}

	return claims.Email, nil
}

func idTokenAudienceMatches(rawAudience json.RawMessage, expectedAudience string) bool {
	var audience string
	if err := json.Unmarshal(rawAudience, &audience); err == nil {
		return audience == expectedAudience
	}

	var audiences []string
	if err := json.Unmarshal(rawAudience, &audiences); err != nil {
		return false
	}

	for _, audience := range audiences {
		if audience == expectedAudience {
			return true
		}
	}
	return false
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

	setSessionTokenCookie(c.Writer, c.Request, "", time.Now().Add(-1*time.Hour))

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

type UpdateUserClassRepRequest struct {
	UserID  string `json:"user_id"`
	ClassID int    `json:"class_id"`
}

func (h *AuthHandler) UpdateUserClassRepByRoot(c *gin.Context) {
	var req UpdateUserClassRepRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(req.UserID) == "" || req.ClassID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and class_id are required"})
		return
	}

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
		c.JSON(http.StatusNotFound, gin.H{"error": "Class not found"})
		return
	}
	if class.EventID != nil && *class.EventID != activeEventID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Selected class does not belong to the active event"})
		return
	}

	roleName := class.Name + "_rep"
	if err := h.userRepo.ReplaceClassRepRole(req.UserID, roleName, req.ClassID, &activeEventID); err != nil {
		log.Printf("ReplaceClassRepRole error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to replace class role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Class role replaced successfully"})
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

func generateStateOauthCookie(w http.ResponseWriter, r *http.Request) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Expires:  expiration,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return state
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

func clearOauthStateCookie(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauthstate",
		Value:    "",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   shouldUseSecureCookie(r),
		SameSite: http.SameSiteLaxMode,
	})
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
