package handler

import (
	"backapp/internal/middleware"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type BarcodeTokenStore interface {
	SaveActiveToken(userID string, eventID, sportID int, token string, ttl time.Duration) error
	GetActiveToken(userID string, eventID, sportID int) (string, error)
	ConsumeActiveToken(userID string, eventID, sportID int, token string) (bool, error)
}

type redisBarcodeTokenStore struct{}

func NewRedisBarcodeTokenStore() BarcodeTokenStore {
	return &redisBarcodeTokenStore{}
}

func (s *redisBarcodeTokenStore) SaveActiveToken(userID string, eventID, sportID int, token string, ttl time.Duration) error {
	return middleware.SetRedisValue(barcodeTokenKey(userID, eventID, sportID), token, ttl)
}

func (s *redisBarcodeTokenStore) GetActiveToken(userID string, eventID, sportID int) (string, error) {
	token, err := middleware.GetRedisValue(barcodeTokenKey(userID, eventID, sportID))
	if err == redis.Nil {
		return "", nil
	}
	return token, err
}

func (s *redisBarcodeTokenStore) ConsumeActiveToken(userID string, eventID, sportID int, token string) (bool, error) {
	return middleware.CompareAndDeleteRedisValue(barcodeTokenKey(userID, eventID, sportID), token)
}

func barcodeTokenKey(userID string, eventID, sportID int) string {
	return fmt.Sprintf("barcode:active:%s:%d:%d", userID, eventID, sportID)
}
