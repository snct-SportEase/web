package repository

import (
	"database/sql"
	"fmt"
	"backapp/internal/config"
	"github.com/go-sql-driver/mysql"
)

func NewDB(cfg *config.Config) (*sql.DB, error) {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.User = cfg.DBUser
	mysqlConfig.Passwd = cfg.DBPassword
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%s", cfg.DBHost, cfg.DBPort)
	mysqlConfig.DBName = cfg.DBName
	mysqlConfig.ParseTime = true

	dsn := mysqlConfig.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil { return nil, err }

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
