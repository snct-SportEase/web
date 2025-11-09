package router

import (
	"backapp/internal/config"
	"backapp/internal/handler"
	"backapp/internal/middleware"
	"backapp/internal/repository"
	"backapp/internal/websocket"
	"database/sql"

	"github.com/gin-gonic/gin"
)

// SetupRouter はGinルーターをセットアップし、ルーティングを定義します
func SetupRouter(db *sql.DB, cfg *config.Config, hubManager *websocket.HubManager) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(middleware.CORSMiddleware())

	// Serve static files for uploaded images
	router.Static("/uploads", "./uploads")

	userRepo := repository.NewUserRepository(db)
	eventRepo := repository.NewEventRepository(db)

	classRepo := repository.NewClassRepository(db)
	teamRepo := repository.NewTeamRepository(db)
	tournRepo := repository.NewTournamentRepository(db)
	classHandler := handler.NewClassHandler(classRepo, eventRepo, teamRepo, tournRepo)

	authHandler := handler.NewAuthHandler(cfg, userRepo, eventRepo, classRepo)

	whitelistRepo := repository.NewWhitelistRepository(db)
	whitelistHandler := handler.NewWhitelistHandler(whitelistRepo, eventRepo)

	sportRepo := repository.NewSportRepository(db)
	sportHandler := handler.NewSportHandler(sportRepo, classRepo, teamRepo, eventRepo, tournRepo)

	eventHandler := handler.NewEventHandler(eventRepo, whitelistRepo)

	tournHandler := handler.NewTournamentHandler(tournRepo, sportRepo, teamRepo, classRepo, hubManager)

	notificationRepo := repository.NewNotificationRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	notificationHandler := handler.NewNotificationHandler(notificationRepo, eventRepo, roleRepo, cfg.WebPushPublicKey, cfg.WebPushPrivateKey)
	notificationRequestRepo := repository.NewNotificationRequestRepository(db)
	notificationRequestHandler := handler.NewNotificationRequestHandler(notificationRequestRepo, notificationRepo, roleRepo, cfg.WebPushPublicKey, cfg.WebPushPrivateKey)

	attendanceHandler := handler.NewAttendanceHandler(classRepo, eventRepo)

	qrCodeHandler := handler.NewQRCodeHandler(teamRepo, sportRepo, userRepo, eventRepo)

	classTeamHandler := handler.NewClassTeamHandler(classRepo, teamRepo, userRepo, eventRepo, sportRepo)

	imageHandler := handler.NewImageHandler()
	pdfHandler := handler.NewPdfHandler()

	mvpRepo := repository.NewMVPRepository(db)
	mvpHandler := handler.NewMVPHandler(mvpRepo)

	wsHandler := handler.NewWebSocketHandler(hubManager)

	// ヘルスチェック用のエンドポイント
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "UP",
		})
	})

	api := router.Group("/api")
	{
		api.GET("/ws/tournaments/:tournament_id", wsHandler.ServeTournamentWebSocket)

		api.GET("/classes", classHandler.GetAllClasses)
		api.GET("/scores/class", middleware.AuthMiddleware(userRepo), classHandler.GetClassScores)

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

		// QR Code routes accessible to authenticated users
		qrcode := api.Group("/qrcode")
		{
			qrcode.Use(middleware.AuthMiddleware(userRepo))
			qrcode.GET("/teams", qrCodeHandler.GetUserTeamsHandler)
			qrcode.POST("/generate", qrCodeHandler.GenerateQRCodeHandler)
			qrcode.POST("/verify", qrCodeHandler.VerifyQRCodeHandler)
		}

		student := api.Group("/student")
		{
			student.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("student", "admin", "root"))
			student.GET("/class-progress", classHandler.GetClassProgress)

			studentNotificationRequests := student.Group("/notification-requests")
			{
				studentNotificationRequests.GET("", notificationRequestHandler.ListStudentRequests)
				studentNotificationRequests.POST("", notificationRequestHandler.CreateRequest)
				studentNotificationRequests.GET("/:request_id", notificationRequestHandler.GetRequestDetail)
				studentNotificationRequests.POST("/:request_id/messages", notificationRequestHandler.AddMessage)
			}
		}

		notifications := api.Group("/notifications")
		{
			notifications.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("student", "admin", "root"))
			notifications.GET("", notificationHandler.ListNotifications)
			notifications.POST("/subscription", notificationHandler.SaveSubscription)
			notifications.DELETE("/subscription", notificationHandler.DeleteSubscription)
		}

		admin := api.Group("/admin")
		{
			admin.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("admin", "root"))

			adminEvent := admin.Group("/events")
			{
				adminEvent.GET("/:event_id/tournaments", tournHandler.GetTournamentsByEventHandler)
			}

			admin.GET("/events", eventHandler.GetAllEvents)

			// Attendance routes
			attendance := admin.Group("/attendance")
			{
				attendance.GET("/class-details/:classID", attendanceHandler.GetClassDetailsHandler)
				attendance.POST("/register", attendanceHandler.RegisterAttendanceHandler)
			}

			// Assign a sport to a specific event
			admin.POST("/events/:id/sports", sportHandler.AssignSportToEventHandler)
			// Delete a sport from a specific event
			admin.DELETE("/events/:event_id/sports/:sport_id", sportHandler.DeleteSportFromEventHandler)

			admin.GET("/events/:event_id/sports/:sport_id/details", sportHandler.GetSportDetailsHandler)
			admin.PUT("/events/:event_id/sports/:sport_id/details", sportHandler.UpdateSportDetailsHandler)

			admin.PUT("/matches/:match_id/start-time", tournHandler.UpdateMatchStartTimeHandler)
			admin.PUT("/matches/:match_id/status", tournHandler.UpdateMatchStatusHandler)
			admin.PUT("/matches/:match_id/result", tournHandler.UpdateMatchResultHandler)

			adminUsers := admin.Group("/users")
			{
				adminUsers.GET("", authHandler.FindUsersHandler)
				adminUsers.PUT("/display-name", authHandler.UpdateUserDisplayNameByAdmin)
				adminUsers.PUT("/role", authHandler.UpdateUserRoleByAdmin)
				adminUsers.DELETE("/role", authHandler.DeleteUserRoleByAdmin)
			}

			admin.GET("/allsports", sportHandler.GetAllSportsHandler)

			admin.POST("/images", imageHandler.UploadImageHandler)
			admin.POST("/pdfs", pdfHandler.UploadPdfHandler)

			mvp := api.Group("/mvp")
			{
				mvp.Use(middleware.AuthMiddleware(userRepo))
				mvp.GET("/class", mvpHandler.GetMVPClass)
			}

			adminMvp := admin.Group("/mvp")
			{
				adminMvp.GET("/eligible-classes", mvpHandler.GetEligibleClasses)
				adminMvp.POST("/vote", mvpHandler.VoteMVP)
				adminMvp.GET("/votes", mvpHandler.GetMVPVotes)
				adminMvp.GET("/user-vote", mvpHandler.GetUserVote)
			}

		}

		// Class and team management routes (accessible to admin/root or class_name_rep)
		adminClassTeam := api.Group("/admin/class-team")
		{
			adminClassTeam.Use(middleware.AuthMiddleware(userRepo), middleware.AdminOrClassRepRequired(userRepo))
			adminClassTeam.GET("/managed-class", classTeamHandler.GetManagedClassHandler)
			adminClassTeam.GET("/classes/:class_id/members", classTeamHandler.GetClassMembersHandler)
			adminClassTeam.POST("/assign-members", classTeamHandler.AssignTeamMembersHandler)
			adminClassTeam.GET("/sports/:sport_id/members", classTeamHandler.GetTeamMembersHandler)
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
				rootEvents.POST("/:event_id/tournaments/generate-all", tournHandler.GenerateAllTournamentsHandler)
				rootEvents.POST("/:event_id/tournaments/generate-preview", tournHandler.GenerateAllTournamentsPreviewHandler)
				rootEvents.POST("/:event_id/tournaments/bulk-create", tournHandler.BulkCreateTournamentsHandler)
				rootEvents.GET("/:event_id/tournaments", tournHandler.GetTournamentsByEventHandler)
			}
			// Sport management routes that require 'root' role
			rootClasses := root.Group("/classes")
			{
				rootClasses.PUT("/student-counts", classHandler.UpdateStudentCountsHandler)
				rootClasses.POST("/student-counts/csv", classHandler.UpdateStudentCountsFromCSVHandler)
			}
			rootSports := root.Group("/sports")
			{
				rootSports.GET("", sportHandler.GetAllSportsHandler)
				rootSports.POST("", sportHandler.CreateSportHandler)
				rootSports.GET("/:id/teams", sportHandler.GetTeamsBySportHandler)
			}

			rootNotifications := root.Group("/notifications")
			{
				rootNotifications.POST("", notificationHandler.CreateNotification)
				rootNotifications.GET("/roles", notificationHandler.ListAvailableRoles)
			}

			rootNotificationRequests := root.Group("/notification-requests")
			{
				rootNotificationRequests.GET("", notificationRequestHandler.ListRootRequests)
				rootNotificationRequests.GET("/:request_id", notificationRequestHandler.GetRequestDetail)
				rootNotificationRequests.POST("/:request_id/messages", notificationRequestHandler.AddMessage)
				rootNotificationRequests.POST("/:request_id/decision", notificationRequestHandler.DecideRequest)
			}
		}
	}

	return router
}
