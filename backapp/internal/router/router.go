package router

import (
	"database/sql"
	"github.com/gin-gonic/gin"
)

// SetupRouter はGinルーターをセットアップし、ルーティングを定義します
func SetupRouter(db *sql.DB) *gin.Engine {
	router := gin.Default()

	// ヘルスチェック用のエンドポイント
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	// 今後ここに他のルーティングを追加していく
	// 例: v1 := router.Group("/api/v1") { ... }

	return router
}
