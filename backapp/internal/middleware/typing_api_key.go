package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const typingAPIKeyHeader = "X-API-Key"

// TypingAPIKeyAuth protects the server-to-server API used by the typing system.
// An empty configured key fails closed so a deployment mistake never exposes
// participant data publicly.
func TypingAPIKeyAuth(expectedKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expectedKey == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Typing integration is not configured"})
			c.Abort()
			return
		}

		providedKey := c.GetHeader(typingAPIKeyHeader)
		if !secureAPIKeyEqual(providedKey, expectedKey) {
			c.Header("WWW-Authenticate", `ApiKey realm="typing-integration"`)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func secureAPIKeyEqual(provided, expected string) bool {
	if provided == "" || len(provided) != len(expected) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
