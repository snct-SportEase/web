package handler

import (
	"backapp/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestPrivateReferenceJSONRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.NoStore())
	value := "original"
	router.GET("/reference", func(c *gin.Context) { privateReferenceJSON(c, gin.H{"name": value}) })
	get := func(etag string) *httptest.ResponseRecorder {
		response := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/reference", nil)
		req.Header.Set("If-None-Match", etag)
		router.ServeHTTP(response, req)
		return response
	}
	first := get("")
	if first.Code != 200 || first.Header().Get("Cache-Control") != "private, max-age=0, must-revalidate" || first.Header().Get("Pragma") != "" {
		t.Fatal("reference policy not applied")
	}
	etag := first.Header().Get("ETag")
	second := get(`"other", W/` + etag)
	if second.Code != 304 || second.Body.Len() != 0 {
		t.Fatal("matching ETag did not return an empty 304")
	}
	value = "changed"
	third := get(etag)
	if third.Code != 200 || third.Header().Get("ETag") == etag {
		t.Fatal("changed data was not returned")
	}
	router.GET("/personal", func(c *gin.Context) { c.JSON(200, gin.H{"user": "private"}) })
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/personal", nil))
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("personal data policy changed")
	}
}
