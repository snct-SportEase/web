package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backapp/internal/models"

	"github.com/gin-gonic/gin"
)

// UserRateLimit applies an atomic Redis-backed limit after AuthMiddleware has
// populated the authenticated user. It is not affected by forwarded IP headers.
func UserRateLimit(limit int, window time.Duration, namespace string) gin.HandlerFunc {
	const incrementScript = `
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`

	return func(c *gin.Context) {
		value, exists := c.Get("user")
		user, ok := value.(*models.User)
		if !exists || !ok || user == nil || user.ID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}
		if redisClient == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limit service unavailable"})
			c.Abort()
			return
		}

		key := fmt.Sprintf("rate_limit:%s:%s", namespace, user.ID)
		redisContext, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		count, err := redisClient.Eval(
			redisContext,
			incrementScript,
			[]string{key},
			window.Milliseconds(),
		).Int64()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Rate limit service unavailable"})
			c.Abort()
			return
		}
		if count > int64(limit) {
			retryAfter := int64(window.Seconds())
			if ttl, err := redisClient.TTL(redisContext, key).Result(); err == nil && ttl > 0 {
				retryAfter = int64(ttl.Seconds()) + 1
			}
			c.Header("Retry-After", strconv.FormatInt(retryAfter, 10))
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many subscription requests"})
			c.Abort()
			return
		}

		c.Next()
	}
}
