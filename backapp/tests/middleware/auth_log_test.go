package middleware_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backapp/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddlewareDoesNotLogRejectedSessionToken(t *testing.T) {
	redisServer := miniredis.RunT(t)
	middleware.InitSessionStore(redisServer.Addr())

	var logs bytes.Buffer
	previousWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() {
		log.SetOutput(previousWriter)
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.AuthMiddleware(nil))
	router.GET("/private", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	const sessionToken = "session-secret-that-must-never-be-logged"
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
	}
	if strings.Contains(logs.String(), sessionToken) {
		t.Fatal("rejected session token was written to logs")
	}
	if !strings.Contains(logs.String(), "Session token rejected") {
		t.Fatalf("expected a non-sensitive rejection log, got %q", logs.String())
	}
}
