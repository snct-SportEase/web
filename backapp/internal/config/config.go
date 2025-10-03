package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost, DBPort, DBUser, DBPassword, DBName            string
	GoogleClientID, GoogleClientSecret, GoogleRedirectURL string
	FrontendURL                                           string
	InitRootUser                                          string
}

func Load() (*Config, error) {
	godotenv.Load("../../.env")
	cfg := &Config{
		DBHost:             os.Getenv("DB_HOST"),
		DBPort:             os.Getenv("DB_PORT"),
		DBUser:             os.Getenv("DB_USER"),
		DBPassword:         os.Getenv("DB_PASSWORD"),
		DBName:             os.Getenv("DB_DATABASE"),
		GoogleClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		GoogleRedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		FrontendURL:        os.Getenv("FRONTEND_URL"),
		InitRootUser:       os.Getenv("INIT_ROOT_USER"),
	}
	return cfg, nil
}
