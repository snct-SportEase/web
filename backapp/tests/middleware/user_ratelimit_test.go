package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backapp/internal/middleware"
	"backapp/internal/models"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func TestUserRateLimitUsesAuthenticatedUser(t *testing.T) {
	redisServer := miniredis.RunT(t)
	middleware.InitSessionStore(redisServer.Addr())

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &models.User{ID: "user-1"})
		c.Next()
	})
	router.POST("/subscription", middleware.UserRateLimit(2, time.Minute, "test-push"), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	for requestNumber := 1; requestNumber <= 3; requestNumber++ {
		request := httptest.NewRequest(http.MethodPost, "/subscription", nil)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)

		if requestNumber <= 2 && response.Code != http.StatusCreated {
			t.Fatalf("request %d status = %d, want %d", requestNumber, response.Code, http.StatusCreated)
		}
		if requestNumber == 3 {
			if response.Code != http.StatusTooManyRequests {
				t.Fatalf("request 3 status = %d, want %d", response.Code, http.StatusTooManyRequests)
			}
			if response.Header().Get("Retry-After") == "" {
				t.Fatal("Retry-After header is missing")
			}
		}
	}
}

func TestUserRateLimitSeparatesUsers(t *testing.T) {
	redisServer := miniredis.RunT(t)
	middleware.InitSessionStore(redisServer.Addr())

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user", &models.User{ID: c.GetHeader("X-Test-User")})
		c.Next()
	})
	router.POST("/subscription", middleware.UserRateLimit(1, time.Minute, "test-separate"), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})

	for _, userID := range []string{"user-1", "user-2"} {
		request := httptest.NewRequest(http.MethodPost, "/subscription", nil)
		request.Header.Set("X-Test-User", userID)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, request)
		if response.Code != http.StatusCreated {
			t.Fatalf("user %s status = %d, want %d", userID, response.Code, http.StatusCreated)
		}
	}
}
