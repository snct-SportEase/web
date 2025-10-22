package handler

import (
	"backapp/internal/config"
	"backapp/internal/middleware"
	"backapp/internal/models"
	"backapp/internal/repository"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
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

func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := generateStateOauthCookie(c.Writer)
	url := h.oauth2Config.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	oauthState, _ := c.Cookie("oauthstate")
	if c.Query("state") != oauthState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state"})
		return
	}

	code := c.Query("code")
	token, err := h.oauth2Config.Exchange(context.Background(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var userInfo struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(data, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// ドメイン制限
	allowedDomains := []string{"sendai-nct.jp", "sendai-nct.ac.jp"}
	emailDomain := strings.Split(userInfo.Email, "@")[1]
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

	// ホワイトリストチェック
	isWhitelisted, err := h.userRepo.IsEmailWhitelisted(userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !isWhitelisted {
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=email_not_whitelisted")
		return
	}

	user, err := h.userRepo.GetUserByEmail(userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user by email"})
		return
	}

	if user == nil {
		// User does not exist, create a new one
		newUser := &models.User{
			ID:                uuid.New().String(),
			Email:             userInfo.Email,
			IsProfileComplete: false,
		}
		if err := h.userRepo.CreateUser(newUser); err != nil {
			if err.Error() == "email not in whitelist" {
				c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=email_not_whitelisted")
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}
		user = newUser // Continue with the newly created user
	} else {
		// User exists, sync role from whitelist
		whitelistRole, err := h.userRepo.GetRoleByEmail(user.Email)
		if err == nil {
			// Add role from whitelist if it doesn't exist
			if err := h.userRepo.AddUserRoleIfNotExists(user.ID, whitelistRole); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user role"})
				return
			}
		}
	}

	// Create session
	sessionToken := uuid.New().String()
	middleware.CreateSession(sessionToken, user.ID)

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "session_token",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		Path:     "/",
		HttpOnly: false,
		Secure:   c.Request.TLS != nil,
		SameSite: http.SameSiteLaxMode,
	})

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

	// Check if it's the first time the profile is being completed
	isFirstCompletion := !user.IsProfileComplete

	user.DisplayName = &req.DisplayName
	user.ClassID = &req.ClassID
	user.IsProfileComplete = true

	if err := h.userRepo.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Assign class representative role on first completion
	if isFirstCompletion {
		class, err := h.classRepo.GetClassByID(req.ClassID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class details"})
			return
		}
		if class != nil {
			roleName := class.Name + "_rep"
			if err := h.userRepo.UpdateUserRole(user.ID, roleName, nil); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign class representative role"})
				return
			}
		}
	}

	c.JSON(http.StatusOK, user)
}

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
		HttpOnly: false,
		Secure:   c.Request.TLS != nil,
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User role deleted successfully"})
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
