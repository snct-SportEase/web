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

	sportRepo := repository.NewSportRepository(db)
	sportHandler := handler.NewSportHandler(sportRepo)

	eventHandler := handler.NewEventHandler(eventRepo, whitelistRepo)

	// ヘルスチェック用のエンドポイント
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	api := router.Group("/api")
	{
		api.GET("/classes", classHandler.GetAllClasses)

		api.GET("/events/active", middleware.AuthMiddleware(userRepo), eventHandler.GetActiveEvent)

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

		// Events accessible to any authenticated user
		events := api.Group("/events")
		{
			events.Use(middleware.AuthMiddleware(userRepo))
			// Get sports for a specific event
			events.GET("/:id/sports", sportHandler.GetSportsByEventHandler)
		}

		admin := api.Group("/admin")
		{
			admin.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("admin", "root"))
			// Assign a sport to a specific event
			admin.POST("/events/:id/sports", sportHandler.AssignSportToEventHandler)
			// Delete a sport from a specific event
			admin.DELETE("/events/:event_id/sports/:sport_id", sportHandler.DeleteSportFromEventHandler)
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
			// Event management routes that require 'root' role
			rootEvents := root.Group("/events")
			{
				rootEvents.GET("", eventHandler.GetAllEvents)
				rootEvents.POST("", eventHandler.CreateEvent)
				rootEvents.PUT("/:id", eventHandler.UpdateEvent)
				rootEvents.PUT("/active", eventHandler.SetActiveEvent)
			}
			// Sport management routes that require 'root' role
			rootSports := root.Group("/sports")
			{
				rootSports.GET("", sportHandler.GetAllSportsHandler)
				rootSports.POST("", sportHandler.CreateSportHandler)
			}
			// User management routes that require 'root' role
			rootUsers := root.Group("/users")
			{
				rootUsers.GET("", authHandler.FindUsersHandler)
				rootUsers.PUT("/display-name", authHandler.UpdateUserDisplayNameByRoot)
			}
		}
	}

	return router
}