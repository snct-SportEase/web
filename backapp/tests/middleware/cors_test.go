package middleware_test

import (
	"backapp/internal/config"
	"backapp/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func performCORSRequest(appEnv, origin string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(middleware.CORSMiddleware(&config.Config{
		FrontendURL: "https://frontend.example",
		AppEnv:      appEnv,
	}))
	router.GET("/ping", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set("Origin", origin)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func TestCORSMiddlewareAllowsConfiguredFrontendInProduction(t *testing.T) {
	recorder := performCORSRequest("production", "https://frontend.example")

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://frontend.example" {
		t.Fatalf("expected configured frontend origin, got %q", got)
	}
}

func TestCORSMiddlewareRejectsLocalhostInProduction(t *testing.T) {
	recorder := performCORSRequest("production", "http://localhost:5000")

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected localhost origin to be rejected in production, got %q", got)
	}
}

func TestCORSMiddlewareAllowsLocalhostOutsideProduction(t *testing.T) {
	recorder := performCORSRequest("development", "http://localhost:5000")

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5000" {
		t.Fatalf("expected localhost origin in development, got %q", got)
	}
}
