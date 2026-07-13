package middleware

import "github.com/gin-gonic/gin"

// NoStore prevents browsers and intermediary caches from retaining API
// responses that can contain authenticated or otherwise sensitive data.
func NoStore() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Next()
	}
}
