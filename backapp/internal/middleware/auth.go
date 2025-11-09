package middleware

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Note: In a production environment, use a persistent session store like Redis.
var sessionStore = make(map[string]string)

func CreateSession(token, userID string) {
	sessionStore[token] = userID
}

func GetUserIDFromSession(token string) (string, bool) {
	userID, exists := sessionStore[token]
	return userID, exists
}

func DeleteSession(token string) {
	delete(sessionStore, token)
}

func IsRequestSecure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	proto := strings.ToLower(r.Header.Get("X-Forwarded-Proto"))
	if proto == "https" {
		return true
	}

	if strings.ToLower(r.Header.Get("X-Forwarded-Ssl")) == "on" {
		return true
	}

	return false
}

func AuthMiddleware(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		cookie, err := c.Cookie("session_token")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		userID, exists := GetUserIDFromSession(cookie)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserWithRoles(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func RoleRequired(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			c.Abort()
			return
		}

		hasRole := false
		for _, userRole := range userModel.Roles {
			for _, requiredRole := range roles {
				if userRole.Name == requiredRole {
					hasRole = true
					break
				}
			}
			if hasRole {
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminOrClassRepRequired allows access to users with admin/root roles or class_name_rep roles
func AdminOrClassRepRequired(userRepo repository.UserRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(*models.User)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			c.Abort()
			return
		}

		// Check if user has admin or root role
		hasAdminRole := false
		for _, userRole := range userModel.Roles {
			if userRole.Name == "admin" || userRole.Name == "root" {
				hasAdminRole = true
				break
			}
		}

		if hasAdminRole {
			c.Next()
			return
		}

		// Check if user has class_name_rep role
		hasClassRepRole := false
		for _, userRole := range userModel.Roles {
			if len(userRole.Name) > 4 && userRole.Name[len(userRole.Name)-4:] == "_rep" {
				hasClassRepRole = true
				break
			}
		}

		if !hasClassRepRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func RootRequired() gin.HandlerFunc {
	return RoleRequired("root")
}
