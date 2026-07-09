package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost, DBPort, DBUser, DBPassword, DBName                           string
	GoogleClientID, GoogleClientSecret, GoogleRedirectURL                string
	FrontendURL                                                          string
	AppEnv                                                               string
	InitRootUser                                                         string
	InitEventName, InitEventSeason, InitEventStartDate, InitEventEndDate string
	InitEventYear                                                        string
	WebPushPublicKey                                                     string
	WebPushPrivateKey                                                    string
	RedisAddr                                                            string
}

func Load() (*Config, error) {
	loadEnv()

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
		AppEnv:             os.Getenv("APP_ENV"),
		InitRootUser:       os.Getenv("INIT_ROOT_USER"),
		InitEventName:      os.Getenv("INIT_EVENT_NAME"),
		InitEventYear:      os.Getenv("INIT_EVENT_YEAR"),
		InitEventSeason:    os.Getenv("INIT_EVENT_SEASON"),
		InitEventStartDate: os.Getenv("INIT_EVENT_START_DATE"),
		InitEventEndDate:   os.Getenv("INIT_EVENT_END_DATE"),
		WebPushPublicKey:   os.Getenv("WEBPUSH_PUBLIC_KEY"),
		WebPushPrivateKey:  os.Getenv("WEBPUSH_PRIVATE_KEY"),
		RedisAddr:          os.Getenv("REDIS_ADDR"),
	}
	return cfg, nil
}

func loadEnv() {
	root, err := findProjectRoot()
	if err != nil {
		return
	}
	_ = godotenv.Load(filepath.Join(root, ".env"))
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if hasProjectRootMarker(dir) {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func hasProjectRootMarker(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "backapp", "go.mod")); err != nil {
		return false
	}

	composeFiles := []string{"docker-compose.yml", "docker-compose.production.yml"}
	for _, file := range composeFiles {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}
