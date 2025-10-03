package middleware

import (
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
