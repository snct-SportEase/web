package main

import (
	"backapp/internal/config"
	"backapp/internal/repository"
	"backapp/internal/router" // routerをインポート
	"database/sql"
	"log"
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

	// ルーターをセットアップ
	r := router.SetupRouter(db, cfg)

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

	// ルートユーザーを登録
	if err := whitelistRepo.AddWhitelistedEmail(cfg.InitRootUser, "root"); err != nil {
		return err
	}

	log.Printf("Successfully initialized root user: %s", cfg.InitRootUser)
	return nil
}
