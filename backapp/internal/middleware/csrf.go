package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

const csrfHeaderName = "X-CSRF-Token"

// CSRFProtection validates state-changing requests that carry a session cookie.
// The expected token is bound to that session in Redis and is never exposed to
// browser JavaScript; the trusted frontend proxy copies it from an HttpOnly
// cookie into the request header after enforcing a same-origin request.
func CSRFProtection() gin.HandlerFunc {
	return func(c *gin.Context) {
		if isSafeMethod(c.Request.Method) {
			c.Next()
			return
		}

		sessionToken, err := c.Cookie("session_token")
		if err != nil || sessionToken == "" {
			// Authentication middleware remains responsible for requests without
			// a session. CSRF is relevant only when ambient credentials are sent.
			c.Next()
			return
		}

		providedToken := c.GetHeader(csrfHeaderName)
		expectedToken, exists := GetCSRFTokenForSession(sessionToken)
		if !exists || !constantTimeStringEqual(providedToken, expectedToken) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid CSRF token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func isSafeMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func constantTimeStringEqual(left, right string) bool {
	if left == "" || right == "" || len(left) != len(right) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(left), []byte(right)) == 1
}
