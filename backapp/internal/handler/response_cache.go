package handler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Only use for non-personal reference data, after authentication and isolation
// checks. Always revalidate so a role change cannot bypass those checks.
func privateReferenceJSON(c *gin.Context, value any) {
	data, err := json.Marshal(value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode response"})
		return
	}
	etag := fmt.Sprintf(`"%x"`, sha256.Sum256(data))
	c.Header("Cache-Control", "private, max-age=0, must-revalidate")
	c.Writer.Header().Del("Pragma")
	c.Writer.Header().Del("Expires")
	c.Header("ETag", etag)
	c.Header("Vary", "Cookie, Authorization")
	ifNoneMatch := ""
	if c.Request != nil {
		ifNoneMatch = c.GetHeader("If-None-Match")
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		candidate = strings.TrimSpace(candidate)
		if candidate == "*" || strings.TrimPrefix(candidate, "W/") == etag {
			c.Status(http.StatusNotModified)
			return
		}
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}
