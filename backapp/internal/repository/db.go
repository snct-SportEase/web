package repository

import (
	"backapp/internal/config"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/panjf2000/ants/v2"
	"github.com/patrickmn/go-cache"
	"golang.org/x/sync/singleflight"
)

var (
	// Global resources for optimization
	GlobalCache   *cache.Cache
	GlobalSFGroup singleflight.Group
	GlobalAnts    *ants.Pool
)

func InitOptimizationResources() {
	// Initialize in-memory cache with 5 min default expiration
	GlobalCache = cache.New(5*time.Minute, 10*time.Minute)

	// Initialize worker pool for high-concurrency tasks
	var err error
	GlobalAnts, err = ants.NewPool(1000) // Adjust pool size as needed
	if err != nil {
		log.Fatalf("failed to create ants pool: %v", err)
	}
}

func NewDB(cfg *config.Config) (*sql.DB, error) {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = cfg.DBUser
	mysqlConfig.Passwd = cfg.DBPassword
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)
	mysqlConfig.DBName = cfg.DBName
	mysqlConfig.AllowNativePasswords = true
	mysqlConfig.ParseTime = true
	mysqlConfig.Collation = "utf8mb4_unicode_ci"
	mysqlConfig.Params = map[string]string{"charset": "utf8mb4"}

	dsn := mysqlConfig.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
