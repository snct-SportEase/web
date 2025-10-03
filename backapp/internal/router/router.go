package router

import (
	"backapp/internal/config"
	"backapp/internal/handler"
	"backapp/internal/repository"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter はGinルーターをセットアップし、ルーティングを定義します
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "http://localhost:3300" || origin == "http://localhost:3000" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, Cookie")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	userRepo := repository.NewUserRepository(db)
	authHandler := handler.NewAuthHandler(cfg, userRepo)

	classRepo := repository.NewClassRepository(db)
	classHandler := handler.NewClassHandler(classRepo)

	// ヘルスチェック用のエンドポイント
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	api := router.Group("/api")
	{
		api.GET("/classes", classHandler.GetAllClasses)

		auth := api.Group("/auth")
		{
			google := auth.Group("/google")
			{
				google.GET("/login", authHandler.GoogleLogin)
				google.GET("/callback", authHandler.GoogleCallback)
			}
			auth.GET("/user", authHandler.AuthMiddleware(), authHandler.GetUser)
			auth.GET("/logout", authHandler.Logout)
		}

		user := api.Group("/user")
		{
			user.Use(authHandler.AuthMiddleware())
			user.PUT("/profile", authHandler.UpdateProfile)
		}
	}

	return router
}
