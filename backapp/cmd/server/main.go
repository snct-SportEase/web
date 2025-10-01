package main

import (
	"log"
	"backapp/internal/config"
	"backapp/internal/repository"
	"backapp/internal/router" // routerをインポート
)

func main() {
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

	// ルーターをセットアップ
	r := router.SetupRouter(db)

	log.Println("Starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}