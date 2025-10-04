package middleware

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"fmt"
	"net/http"

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

		user, err := userRepo.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("user", user)
		c.Next()
	}
}

func RootRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		userModel, ok := user.(*models.User)
		fmt.Printf("%+v\n", userModel)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
			c.Abort()
			return
		}

		isRoot := false
		for _, role := range userModel.Roles {
			if role.Name == "root" {
				isRoot = true
				break
			}
		}

		if !isRoot {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden: root access required"})
			c.Abort()
			return
		}

		c.Next()
	}
}
