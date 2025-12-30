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

	eventHandler := handler.NewEventHandler(eventRepo, whitelistRepo, tournRepo)

	rainyModeRepo := repository.NewRainyModeRepository(db)
	rainyModeHandler := handler.NewRainyModeHandler(rainyModeRepo, eventRepo)

	tournHandler := handler.NewTournamentHandler(tournRepo, sportRepo, teamRepo, classRepo, eventRepo, hubManager)
	noonRepo := repository.NewNoonGameRepository(db)
	noonHandler := handler.NewNoonGameHandler(noonRepo, classRepo, eventRepo)

	notificationRepo := repository.NewNotificationRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	notificationHandler := handler.NewNotificationHandler(notificationRepo, eventRepo, roleRepo, cfg.WebPushPublicKey, cfg.WebPushPrivateKey)
	notificationRequestRepo := repository.NewNotificationRequestRepository(db)
	notificationRequestHandler := handler.NewNotificationRequestHandler(notificationRequestRepo, notificationRepo, roleRepo, cfg.WebPushPublicKey, cfg.WebPushPrivateKey)

	attendanceHandler := handler.NewAttendanceHandler(classRepo, eventRepo)

	qrCodeHandler := handler.NewQRCodeHandler(teamRepo, sportRepo, userRepo, eventRepo, classRepo)

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

			studentEvents := student.Group("/events")
			{
				studentEvents.GET("/:event_id/tournaments", tournHandler.GetTournamentsByEventHandler)
			}

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
			notifications.GET("/subscription", notificationHandler.GetSubscription)
			notifications.POST("/subscription", notificationHandler.SaveSubscription)
			notifications.DELETE("/subscription", notificationHandler.DeleteSubscription)
			notifications.GET("/debug", notificationHandler.GetNotificationDebugInfo)
		}

		admin := api.Group("/admin")
		{
			admin.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("admin", "root"))

			adminEvent := admin.Group("/events")
			{
				adminEvent.GET("/:event_id/tournaments", tournHandler.GetTournamentsByEventHandler)
				adminEvent.GET("/:event_id/noon-game/session", noonHandler.GetSession)
				// Templates (noon-game)
				adminEvent.POST("/:event_id/noon-game/templates/year-relay/run", noonHandler.CreateYearRelayRun)
			}

			admin.GET("/events", eventHandler.GetAllEvents)

			// Attendance routes
			attendance := admin.Group("/attendance")
			{
				attendance.GET("/class-details/:classID", attendanceHandler.GetClassDetailsHandler)
				attendance.POST("/register", attendanceHandler.RegisterAttendanceHandler)
			}

			// Assign a sport to a specific event
			admin.POST("/events/:event_id/sports", sportHandler.AssignSportToEventHandler)
			// Delete a sport from a specific event
			admin.DELETE("/events/:event_id/sports/:sport_id", sportHandler.DeleteSportFromEventHandler)

			admin.GET("/events/:event_id/sports/:sport_id/details", sportHandler.GetSportDetailsHandler)
			admin.PUT("/events/:event_id/sports/:sport_id/details", sportHandler.UpdateSportDetailsHandler)
			admin.PUT("/events/:event_id/sports/:sport_id/capacity", sportHandler.UpdateCapacityHandler)
			admin.PUT("/events/:event_id/sports/:sport_id/classes/:class_id/capacity", sportHandler.UpdateClassCapacityHandler)

			admin.PUT("/matches/:match_id/start-time", tournHandler.UpdateMatchStartTimeHandler)
			admin.PUT("/matches/:match_id/rainy-mode-start-time", tournHandler.UpdateMatchRainyModeStartTimeHandler)
			admin.PUT("/matches/:match_id/status", tournHandler.UpdateMatchStatusHandler)
			admin.PUT("/matches/:match_id/result", tournHandler.UpdateMatchResultHandler)
			admin.PUT("/noon-game/matches/:match_id/result", noonHandler.RecordMatchResult)
			admin.PUT("/noon-game/template-runs/:run_id/year-relay/blocks/:block/result", noonHandler.RecordYearRelayBlockResult)
			admin.PUT("/noon-game/template-runs/:run_id/year-relay/overall/result", noonHandler.RecordYearRelayOverallBonus)

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
			adminClassTeam.DELETE("/remove-member", classTeamHandler.RemoveTeamMemberHandler)
			adminClassTeam.GET("/sports/:sport_id/members", classTeamHandler.GetTeamMembersHandler)
			adminClassTeam.GET("/sports/:sport_id/confirmed-members", classTeamHandler.GetConfirmedTeamMembersHandler)
		}

		root := api.Group("/root")
		{
			root.Use(middleware.AuthMiddleware(userRepo), middleware.RoleRequired("root"))
			whitelist := root.Group("/whitelist")
			{
				whitelist.GET("", whitelistHandler.GetWhitelistHandler)
				whitelist.POST("", whitelistHandler.AddWhitelistedEmailHandler)
				whitelist.POST("/csv", whitelistHandler.BulkAddWhitelistedEmailsHandler)
				whitelist.DELETE("", whitelistHandler.DeleteWhitelistedEmailHandler)
				whitelist.DELETE("/bulk", whitelistHandler.DeleteWhitelistedEmailsHandler)
			}
			// Event management routes that require 'root' role
			rootEvents := root.Group("/events")
			{
				rootEvents.GET("", eventHandler.GetAllEvents)
				rootEvents.POST("", eventHandler.CreateEvent)
				rootEvents.PUT("/active", eventHandler.SetActiveEvent)
				// More specific routes must come before the generic :id route
				rootEvents.PUT("/:id/rainy-mode", eventHandler.SetRainyMode)
				rootEvents.GET("/:id/rainy-mode/settings", rainyModeHandler.GetRainyModeSettingsHandler)
				rootEvents.POST("/:id/rainy-mode/settings", rainyModeHandler.UpsertRainyModeSettingHandler)
				rootEvents.PUT("/:id/rainy-mode/settings", rainyModeHandler.UpsertRainyModeSettingHandler)
				rootEvents.DELETE("/:id/rainy-mode/settings/:sport_id/:class_id", rainyModeHandler.DeleteRainyModeSettingHandler)
				rootEvents.POST("/:id/tournaments/generate-all", tournHandler.GenerateAllTournamentsHandler)
				rootEvents.POST("/:id/tournaments/generate-preview", tournHandler.GenerateAllTournamentsPreviewHandler)
				rootEvents.POST("/:id/tournaments/bulk-create", tournHandler.BulkCreateTournamentsHandler)
				rootEvents.GET("/:id/tournaments", tournHandler.GetTournamentsByEventHandler)
				rootEvents.GET("/:id/noon-game/session", noonHandler.GetSession)
				rootEvents.POST("/:id/noon-game/session", noonHandler.UpsertSession)
				rootEvents.PUT("/:id/competition-guidelines", eventHandler.UpdateCompetitionGuidelines)
				// Generic :id route should be last
				rootEvents.PUT("/:id", eventHandler.UpdateEvent)
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

			rootNoon := root.Group("/noon-game")
			{
				rootNoon.POST("/sessions/:session_id/groups", noonHandler.SaveGroup)
				rootNoon.PUT("/sessions/:session_id/groups/:group_id", noonHandler.SaveGroup)
				rootNoon.DELETE("/sessions/:session_id/groups/:group_id", noonHandler.DeleteGroup)
				rootNoon.POST("/sessions/:session_id/matches", noonHandler.SaveMatch)
				rootNoon.PUT("/sessions/:session_id/matches/:match_id", noonHandler.SaveMatch)
				rootNoon.DELETE("/sessions/:session_id/matches/:match_id", noonHandler.DeleteMatch)
				rootNoon.POST("/sessions/:session_id/manual-points", noonHandler.AddManualPoint)
			}

			rootNotifications := root.Group("/notifications")
			{
				rootNotifications.POST("", notificationHandler.CreateNotification)
				rootNotifications.GET("/roles", notificationHandler.ListAvailableRoles)
			}

			rootUsers := root.Group("/users")
			{
				rootUsers.GET("", authHandler.FindUsersHandler)
				rootUsers.PUT("/display-name", authHandler.UpdateUserDisplayNameByAdmin)
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
