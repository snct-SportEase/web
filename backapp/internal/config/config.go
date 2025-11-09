package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost, DBPort, DBUser, DBPassword, DBName                           string
	GoogleClientID, GoogleClientSecret, GoogleRedirectURL                string
	FrontendURL                                                          string
	InitRootUser                                                         string
	InitEventName, InitEventSeason, InitEventStartDate, InitEventEndDate string
	InitEventYear                                                        string
	WebPushPublicKey                                                     string
	WebPushPrivateKey                                                    string
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
		InitEventName:      os.Getenv("INIT_EVENT_NAME"),
		InitEventYear:      os.Getenv("INIT_EVENT_YEAR"),
		InitEventSeason:    os.Getenv("INIT_EVENT_SEASON"),
		InitEventStartDate: os.Getenv("INIT_EVENT_START_DATE"),
		InitEventEndDate:   os.Getenv("INIT_EVENT_END_DATE"),
		WebPushPublicKey:   os.Getenv("WEBPUSH_PUBLIC_KEY"),
		WebPushPrivateKey:  os.Getenv("WEBPUSH_PRIVATE_KEY"),
	}
	return cfg, nil
}
