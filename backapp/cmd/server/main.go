package main

import (
	"backapp/internal/config"
	"backapp/internal/models"
	"backapp/internal/repository"
	"backapp/internal/router"
	"backapp/internal/websocket"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"
)

func main() {
	log.Println("Starting the application...")
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := repository.NewDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection successful.")

	// 初期ルートユーザーを登録
	if err := initializeRootUser(db, cfg); err != nil {
		log.Printf("Warning: Failed to initialize root user: %v", err)
	}

	// 初期イベントを作成
	eventID, err := initializeEvent(db, cfg)
	if err != nil {
		log.Printf("Warning: Failed to initialize event: %v", err)
	}

	// class_scores テーブルを初期化
	if err := initializeClassScores(db, eventID); err != nil {
		log.Printf("Warning: Failed to initialize class scores: %v", err)
	}

	hubManager := websocket.NewHubManager()

	// ルーターをセットアップ
	r := router.SetupRouter(db, cfg, hubManager)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// initializeRootUser は初期ルートユーザーをwhitelisted_emailsテーブルに登録する
func initializeRootUser(db *sql.DB, cfg *config.Config) error {
	if cfg.InitRootUser == "" {
		log.Println("INIT_ROOT_USER is not set, skipping root user initialization")
		return nil
	}

	whitelistRepo := repository.NewWhitelistRepository(db)

	// 既に登録されているかチェック
	isWhitelisted, err := whitelistRepo.IsEmailWhitelisted(cfg.InitRootUser)
	if err != nil {
		return err
	}

	if isWhitelisted {
		log.Printf("Root user %s is already whitelisted", cfg.InitRootUser)
		return nil
	}

	// ルートユーザーを登録 (event_id は NULL)
	if err := whitelistRepo.AddWhitelistedEmail(cfg.InitRootUser, "root", nil); err != nil {
		return err
	}

	log.Printf("Successfully initialized root user: %s", cfg.InitRootUser)
	return nil
}

// initializeEvent は初期イベントと関連クラスを登録する
func initializeEvent(db *sql.DB, cfg *config.Config) (int64, error) {
	if cfg.InitEventName == "" {
		log.Println("INIT_EVENT_NAME is not set, skipping event initialization")
		return 0, nil
	}

	// 既存イベントのチェック
	var existingEventID int64
	err := db.QueryRow("SELECT id FROM events WHERE name = ?", cfg.InitEventName).Scan(&existingEventID)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to check for existing event: %w", err)
	}

	// イベントが既に存在する場合は、そのIDを返してスキップ
	if err != sql.ErrNoRows {
		log.Printf("Event '%s' already exists with ID: %d, skipping initialization.", cfg.InitEventName, existingEventID)
		return existingEventID, nil
	}

	// --- イベント作成 ---
	eventRepo := repository.NewEventRepository(db)

	// season の値チェック
	season := cfg.InitEventSeason
	if season != "spring" && season != "autumn" {
		return 0, fmt.Errorf("invalid season value: '%s'. must be 'spring' or 'autumn'", season)
	}

	year, err := strconv.Atoi(cfg.InitEventYear)
	if err != nil {
		return 0, err
	}
	layout := "2006-01-02"
	startDate, err := time.Parse(layout, cfg.InitEventStartDate)
	if err != nil {
		return 0, err
	}
	endDate, err := time.Parse(layout, cfg.InitEventEndDate)
	if err != nil {
		return 0, err
	}
	event := &models.Event{
		Name:       cfg.InitEventName,
		Year:       year,
		Season:     season,
		Start_date: &startDate,
		End_date:   &endDate,
	}

	eventID, err := eventRepo.CreateEvent(event)
	if err != nil {
		return 0, fmt.Errorf("failed to create event: %w", err)
	}
	log.Printf("Successfully created event '%s' with ID: %d", cfg.InitEventName, eventID)

	// --- クラス作成 ---
	classRepo := repository.NewClassRepository(db)
	classNames := []string{
		"1-1", "1-2", "1-3", "IS2", "IS3",
		"IS4", "IS5", "IT2", "IT3", "IT4",
		"IT5", "IE2", "IE3", "IE4", "IE5",
		"専教",
	}

	if err := classRepo.CreateClasses(int(eventID), classNames); err != nil {
		return 0, fmt.Errorf("failed to create classes: %w", err)
	}
	log.Printf("Successfully created classes for event ID: %d", eventID)

	// --- アクティブイベント設定 ---
	activeEventID := int(eventID)
	if err := eventRepo.SetActiveEvent(&activeEventID); err != nil {
		return 0, fmt.Errorf("failed to set active event: %w", err)
	}
	log.Printf("Successfully set active event to ID: %d", activeEventID)

	return eventID, nil
}

// initializeClassScores は class_scores テーブルを初期化する
func initializeClassScores(db *sql.DB, eventID int64) error {
	if eventID == 0 {
		log.Println("Event ID is 0, skipping class scores initialization")
		return nil
	}

	classRepo := repository.NewClassRepository(db)
	classes, err := classRepo.GetAllClasses(int(eventID))
	if err != nil {
		return fmt.Errorf("failed to get classes for event ID %d: %w", eventID, err)
	}

	if len(classes) == 0 {
		log.Printf("No classes found for event ID %d, skipping class scores initialization.", eventID)
		return nil
	}

	var classIDs []int
	for _, class := range classes {
		classIDs = append(classIDs, class.ID)
	}

	classScoreRepo := repository.NewClassScoreRepository(db)
	if err := classScoreRepo.InitializeClassScores(int(eventID), classIDs); err != nil {
		return fmt.Errorf("failed to initialize class scores: %w", err)
	}

	log.Printf("Successfully initialized class scores for event ID: %d", eventID)
	return nil
}
