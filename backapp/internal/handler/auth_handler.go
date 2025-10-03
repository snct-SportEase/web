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
	oauth2Config *oauth2.Config
}

func NewAuthHandler(cfg *config.Config, userRepo repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		cfg:      cfg,
		userRepo: userRepo,
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
		c.Redirect(http.StatusTemporaryRedirect, strings.TrimSuffix(h.cfg.FrontendURL, "/")+"/?error=access_denied")
		return
	}

	user, err := h.userRepo.GetUserByEmail(userInfo.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user == nil {
		newUser := &models.User{
			ID:                uuid.New().String(),
			Email:             userInfo.Email,
			IsProfileComplete: false,
		}
		if err := h.userRepo.CreateUser(newUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		user = newUser
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
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	c.JSON(http.StatusOK, user)
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

	user.DisplayName = &req.DisplayName
	user.ClassID = &req.ClassID
	user.IsProfileComplete = true

	if err := h.userRepo.UpdateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)

	return state
}
