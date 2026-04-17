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
	"io"
	"log"
	"net/http"
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
			Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		},
	}
}

// ... (GoogleLogin, GoogleCallback, GetUser are unchanged)

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := generateStateOauthCookie(c.Writer, c.Request)
	url := h.oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	oauthState, _ := c.Cookie("oauthstate")
	if subtle.ConstantTimeCompare([]byte(c.Query("state")), []byte(oauthState)) != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	code := c.Query("code")
	token, err := h.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauth2 token exchange error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create userinfo request"})
		return
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)

	response, err := http.DefaultClient.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user info"})
		return
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Printf("failed to read userinfo response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		log.Printf("failed to parse userinfo: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication failed"})
		return
	}

	// ドメイン制限
	allowedDomains := []string{"sendai-nct.jp", "sendai-nct.ac.jp"}
	parts := strings.Split(userInfo.Email, "@")
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

	user, err := h.userRepo.GetUserByEmail(userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user by email"})
		return
	}

	if user == nil {
		// 新規ユーザー: InitRootUserと一致する場合はrootロール、それ以外はstudent
		role := "student"
		if h.cfg != nil && h.cfg.InitRootUser == userInfo.Email {
			role = "root"
		}

		newUser := &models.User{
			ID:                uuid.New().String(),
			Email:             userInfo.Email,
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

	isSecure := middleware.IsRequestSecure(c.Request)
	
	// Set session cookie
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   isSecure,
		SameSite: http.SameSiteLaxMode,
	})

	// Add a debug log to verify cookie flags
	log.Printf("[auth] Session created for user %s, secure=%v, origin=%s", user.Email, isSecure, c.Request.RemoteAddr)

	c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/dashboard")
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

	// Check if it's the first time the profile is being completed
	isFirstCompletion := !user.IsProfileComplete

	user.DisplayName = &req.DisplayName
	user.ClassID = &req.ClassID
	user.IsProfileComplete = true

	if err := h.userRepo.UpdateUser(user); err != nil {
		log.Printf("UpdateUser error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	// Assign class representative role on first completion
	if isFirstCompletion {
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
		if class != nil {
			// Verify the class belongs to the active event
			if class.EventID != nil && *class.EventID != activeEventID {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Selected class does not belong to the active event"})
				return
			}

			roleName := class.Name + "_rep"
			if err := h.userRepo.UpdateUserRole(user.ID, roleName, &activeEventID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign class representative role"})
				return
			}
		}
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

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   middleware.IsRequestSecure(c.Request),
		SameSite: http.SameSiteLaxMode,
	})

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
	Role   string `json:"role"` // "admin" or "root"
}

// PromoteUserByRoot はroot権限でユーザーをadminまたはrootに昇格させる
func (h *AuthHandler) PromoteUserByRoot(c *gin.Context) {
	var req PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != "admin" && req.Role != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "昇格できるロールは 'admin' または 'root' のみです"})
		return
	}

	if err := h.userRepo.AddUserRoleIfNotExists(req.UserID, req.Role); err != nil {
		log.Printf("PromoteUserByRoot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to promote user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User promoted successfully"})
}

// DemoteUserByRoot はroot権限でユーザーからadminまたはrootロールを剥奪する
func (h *AuthHandler) DemoteUserByRoot(c *gin.Context) {
	var req PromoteUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Role != "admin" && req.Role != "root" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "降格できるロールは 'admin' または 'root' のみです"})
		return
	}

	if err := h.userRepo.DeleteUserRole(req.UserID, req.Role); err != nil {
		log.Printf("DemoteUserByRoot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to demote user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User demoted successfully"})
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
		Secure:   middleware.IsRequestSecure(r),
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	return state
}
