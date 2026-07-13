package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backapp/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestNoStoreSetsCachePreventionHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.NoStore())
	router.GET("/api/private", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	request := httptest.NewRequest(http.MethodGet, "/api/private", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if got := response.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q, want %q", got, "no-store")
	}
	if got := response.Header().Get("Pragma"); got != "no-cache" {
		t.Fatalf("Pragma = %q, want %q", got, "no-cache")
	}
	if got := response.Header().Get("Expires"); got != "0" {
		t.Fatalf("Expires = %q, want %q", got, "0")
	}
}
