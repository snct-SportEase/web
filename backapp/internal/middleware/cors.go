package middleware

import (
	"backapp/internal/config"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		frontendURL := strings.TrimSuffix(cfg.FrontendURL, "/")
		isProduction := strings.EqualFold(cfg.AppEnv, "production") || strings.EqualFold(gin.Mode(), gin.ReleaseMode)
		isLocalhost := strings.HasPrefix(origin, "http://localhost:") || strings.HasPrefix(origin, "http://127.0.0.1:")

		// Allow configured frontend URL, with localhost limited to development.
		if origin == frontendURL || (!isProduction && isLocalhost) {
			c.Header("Access-Control-Allow-Origin", origin)
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
