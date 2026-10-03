package middleware

import (
	"backapp/internal/models"
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type TestRunStatusReader interface {
	GetStatus(ctx context.Context) (*models.EventTestRunStatus, error)
}

// TestRunIsolation keeps real student accounts out of the mutable application
// while a production rehearsal or recovery is in progress. Admin and root
// accounts are the only test participants.
func TestRunIsolation(statusReader TestRunStatusReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		user, ok := userValue.(*models.User)
		if !exists || !ok || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}

		for _, role := range user.Roles {
			if role.Name == "admin" || role.Name == "root" {
				c.Next()
				return
			}
		}

		status, err := statusReader.GetStatus(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "テスト試行状態を確認できないため、学生向け機能を一時停止しています",
				"code":  "test_run_state_unavailable",
			})
			c.Abort()
			return
		}
		if status != nil && status.IsIsolatingUsers() {
			c.JSON(http.StatusLocked, gin.H{
				"error":          "テスト試行中のため、学生向け機能はメンテナンス中です",
				"code":           "test_run_maintenance",
				"test_run_state": status.State,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
