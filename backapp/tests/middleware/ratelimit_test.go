package middleware_test

import (
	"backapp/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newIPRateLimitRouter(t *testing.T, trustedProxies []string, limit int) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	require.NoError(t, router.SetTrustedProxies(trustedProxies))
	router.RemoteIPHeaders = []string{"X-Forwarded-For"}
	router.GET("/limited", middleware.RateLimit(limit, time.Minute), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	return router
}

func performRateLimitedRequest(router *gin.Engine, remoteAddr, forwardedFor string) int {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/limited", nil)
	request.RemoteAddr = remoteAddr
	request.Header.Set("X-Forwarded-For", forwardedFor)
	router.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestRateLimitIgnoresForwardedForFromUntrustedPeer(t *testing.T) {
	router := newIPRateLimitRouter(t, []string{"192.0.2.10/32"}, 2)

	require.Equal(t, http.StatusNoContent, performRateLimitedRequest(router, "203.0.113.5:41000", "198.51.100.1"))
	require.Equal(t, http.StatusNoContent, performRateLimitedRequest(router, "203.0.113.5:41001", "198.51.100.2"))
	require.Equal(t, http.StatusTooManyRequests, performRateLimitedRequest(router, "203.0.113.5:41002", "198.51.100.3"))
}

func TestRateLimitUsesCanonicalAddressFromTrustedProxy(t *testing.T) {
	router := newIPRateLimitRouter(t, []string{"192.0.2.10/32"}, 1)

	require.Equal(t, http.StatusNoContent, performRateLimitedRequest(router, "192.0.2.10:41000", "198.51.100.1"))
	require.Equal(t, http.StatusTooManyRequests, performRateLimitedRequest(router, "192.0.2.10:41001", "198.51.100.1"))
	require.Equal(t, http.StatusNoContent, performRateLimitedRequest(router, "192.0.2.10:41002", "198.51.100.2"))
}
