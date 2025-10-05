package router

import (
	"backapp/internal/config"
	"backapp/internal/handler"
	"backapp/internal/middleware"
	"backapp/internal/repository"
	"database/sql"

	"github.com/gin-gonic/gin"
)

// SetupRouter はGinルーターをセットアップし、ルーティングを定義します
func SetupRouter(db *sql.DB, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	userRepo := repository.NewUserRepository(db)
	authHandler := handler.NewAuthHandler(cfg, userRepo)

	classRepo := repository.NewClassRepository(db)
	classHandler := handler.NewClassHandler(classRepo)

	whitelistRepo := repository.NewWhitelistRepository(db)
	eventRepo := repository.NewEventRepository(db)
	whitelistHandler := handler.NewWhitelistHandler(whitelistRepo, eventRepo)

	eventHandler := handler.NewEventHandler(eventRepo)

	// ヘルスチェック用のエンドポイント
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	api := router.Group("/api")
	{
		api.GET("/classes", classHandler.GetAllClasses)

		api.GET("/events/active", eventHandler.GetActiveEvent)

		auth := api.Group("/auth")
		{
			google := auth.Group("/google")
			{
				google.GET("/login", authHandler.GoogleLogin)
				google.GET("/callback", authHandler.GoogleCallback)
			}
			auth.GET("/user", middleware.AuthMiddleware(userRepo), authHandler.GetUser)
			auth.POST("/logout", authHandler.Logout)
		}

		user := api.Group("/user")
		{
			user.Use(middleware.AuthMiddleware(userRepo))
			user.PUT("/profile", authHandler.UpdateProfile)
		}

		root := api.Group("/root")
		{
			root.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("root"))
			whitelist := root.Group("/whitelist")
			{
				whitelist.GET("", whitelistHandler.GetWhitelistHandler)
				whitelist.POST("", whitelistHandler.AddWhitelistedEmailHandler)
				whitelist.POST("/csv", whitelistHandler.BulkAddWhitelistedEmailsHandler)
			}
			events := root.Group("/events")
			{
				events.GET("", eventHandler.GetAllEvents)
				events.POST("", eventHandler.CreateEvent)
				events.PUT("/:id", eventHandler.UpdateEvent)
				events.PUT("/active", eventHandler.SetActiveEvent)
			}
		}
	}

	return router
}
