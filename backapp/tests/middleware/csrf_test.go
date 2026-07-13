package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backapp/internal/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
)

func newCSRFTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	redisServer := miniredis.RunT(t)
	middleware.InitSessionStore(redisServer.Addr())
	if err := middleware.CreateSession("session-token", "user-1", "expected-csrf-token"); err != nil {
		t.Fatalf("create session: %v", err)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.CSRFProtection())
	router.Any("/resource", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func performCSRFRequest(router *gin.Engine, method, sessionToken, csrfToken string) int {
	request := httptest.NewRequest(method, "/resource", nil)
	if sessionToken != "" {
		request.AddCookie(&http.Cookie{Name: "session_token", Value: sessionToken})
	}
	if csrfToken != "" {
		request.Header.Set("X-CSRF-Token", csrfToken)
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code
}

func TestCSRFProtection(t *testing.T) {
	router := newCSRFTestRouter(t)

	if status := performCSRFRequest(router, http.MethodPost, "session-token", "expected-csrf-token"); status != http.StatusNoContent {
		t.Fatalf("valid token status = %d, want %d", status, http.StatusNoContent)
	}
	if status := performCSRFRequest(router, http.MethodPut, "session-token", ""); status != http.StatusForbidden {
		t.Fatalf("missing token status = %d, want %d", status, http.StatusForbidden)
	}
	if status := performCSRFRequest(router, http.MethodDelete, "session-token", "attacker-token"); status != http.StatusForbidden {
		t.Fatalf("mismatched token status = %d, want %d", status, http.StatusForbidden)
	}
	if status := performCSRFRequest(router, http.MethodPost, "unknown-session", "expected-csrf-token"); status != http.StatusForbidden {
		t.Fatalf("unbound token status = %d, want %d", status, http.StatusForbidden)
	}
	if status := performCSRFRequest(router, http.MethodGet, "session-token", ""); status != http.StatusNoContent {
		t.Fatalf("safe method status = %d, want %d", status, http.StatusNoContent)
	}
	if status := performCSRFRequest(router, http.MethodPost, "", ""); status != http.StatusNoContent {
		t.Fatalf("unauthenticated request status = %d, want downstream handling", status)
	}
}
