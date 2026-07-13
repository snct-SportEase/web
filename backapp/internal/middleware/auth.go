package middleware

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const sessionTTL = 24 * time.Hour
const sessionKeyPrefix = "session:"
const csrfKeyPrefix = "csrf:"

var redisClient *redis.Client

func InitSessionStore(addr string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func CreateSession(token, userID, csrfToken string) error {
	if redisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}
	if token == "" || userID == "" || csrfToken == "" {
		return fmt.Errorf("session token, user ID, and CSRF token are required")
	}

	ctx := context.Background()
	_, err := redisClient.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		pipe.Set(ctx, sessionKeyPrefix+token, userID, sessionTTL)
		pipe.Set(ctx, csrfKeyPrefix+token, csrfToken, sessionTTL)
		return nil
	})
	return err
}

func GetUserIDFromSession(token string) (string, bool) {
	if redisClient == nil {
		return "", false
	}
	val, err := redisClient.Get(context.Background(), sessionKeyPrefix+token).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("[redis] Redis error during session lookup: %v", err)
		}
		return "", false
	}
	return val, true
}

func GetCSRFTokenForSession(token string) (string, bool) {
	if redisClient == nil {
		return "", false
	}

	value, err := redisClient.Get(context.Background(), csrfKeyPrefix+token).Result()
	if err != nil {
		if err != redis.Nil {
			log.Printf("[redis] Redis error during CSRF lookup: %T", err)
		}
		return "", false
	}
	return value, true
}

func DeleteSession(token string) {
	if redisClient == nil {
		return
	}
	redisClient.Del(context.Background(), sessionKeyPrefix+token, csrfKeyPrefix+token)
}

func SetRedisValue(key, value string, ttl time.Duration) error {
	if redisClient == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	return redisClient.Set(context.Background(), key, value, ttl).Err()
}

func GetRedisValue(key string) (string, error) {
	if redisClient == nil {
		return "", fmt.Errorf("redis client is not initialized")
	}

	return redisClient.Get(context.Background(), key).Result()
}

func CompareAndDeleteRedisValue(key, expectedValue string) (bool, error) {
	if redisClient == nil {
		return false, fmt.Errorf("redis client is not initialized")
	}

	const script = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`
	result, err := redisClient.Eval(context.Background(), script, []string{key}, expectedValue).Int()
	if err != nil {
		return false, err
	}
	return result == 1, nil
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
			log.Printf("[auth] No session_token cookie found for path: %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: No session cookie"})
			c.Abort()
			return
		}

		userID, exists := GetUserIDFromSession(cookie)
		if !exists {
			// Session cookies are bearer credentials. Never include the value in
			// logs, even when the lookup fails.
			log.Printf("[auth] Session token rejected for path: %s", c.Request.URL.Path)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Session expired or invalid"})
			c.Abort()
			return
		}

		user, err := userRepo.GetUserWithRoles(userID)
		if err != nil {
			log.Printf("[auth] Failed to fetch user from DB for ID %s: %v", userID, err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User fetch failed"})
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

func RootRequired() gin.HandlerFunc {
	return RoleRequired("root")
}
