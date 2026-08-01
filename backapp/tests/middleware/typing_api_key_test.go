package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backapp/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func performTypingAPIKeyRequest(configuredKey, providedKey string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.TypingAPIKeyAuth(configuredKey))
	router.GET("/typing", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodGet, "/typing", nil)
	if providedKey != "" {
		request.Header.Set("X-API-Key", providedKey)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestTypingAPIKeyAuthAcceptsMatchingKey(t *testing.T) {
	recorder := performTypingAPIKeyRequest("correct-secret", "correct-secret")
	assert.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestTypingAPIKeyAuthRejectsMissingOrInvalidKey(t *testing.T) {
	for _, providedKey := range []string{"", "wrong-secret", "correct-secret-with-suffix"} {
		recorder := performTypingAPIKeyRequest("correct-secret", providedKey)
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.NotEmpty(t, recorder.Header().Get("WWW-Authenticate"))
	}
}

func TestTypingAPIKeyAuthFailsClosedWhenNotConfigured(t *testing.T) {
	recorder := performTypingAPIKeyRequest("", "any-key")
	assert.Equal(t, http.StatusServiceUnavailable, recorder.Code)
}
