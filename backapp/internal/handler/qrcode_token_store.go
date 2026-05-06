package handler

import (
	"backapp/internal/middleware"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type QRCodeTokenStore interface {
	SaveActiveToken(userID string, eventID, sportID int, token string, ttl time.Duration) error
	GetActiveToken(userID string, eventID, sportID int) (string, error)
}

type redisQRCodeTokenStore struct{}

func NewRedisQRCodeTokenStore() QRCodeTokenStore {
	return &redisQRCodeTokenStore{}
}

func (s *redisQRCodeTokenStore) SaveActiveToken(userID string, eventID, sportID int, token string, ttl time.Duration) error {
	return middleware.SetRedisValue(qrCodeTokenKey(userID, eventID, sportID), token, ttl)
}

func (s *redisQRCodeTokenStore) GetActiveToken(userID string, eventID, sportID int) (string, error) {
	token, err := middleware.GetRedisValue(qrCodeTokenKey(userID, eventID, sportID))
	if err == redis.Nil {
		return "", nil
	}
	return token, err
}

func qrCodeTokenKey(userID string, eventID, sportID int) string {
	return fmt.Sprintf("qr_code:active:%s:%d:%d", userID, eventID, sportID)
}
