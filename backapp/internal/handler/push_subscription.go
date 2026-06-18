package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"log"
	"net/http"
)

func cleanupExpiredPushSubscription(repo repository.NotificationRepository, sub models.PushSubscription, statusCode int, logPrefix string) {
	if statusCode != http.StatusGone && statusCode != http.StatusNotFound {
		return
	}

	if err := repo.DeletePushSubscription(sub.UserID, sub.Endpoint); err != nil {
		log.Printf("[%s] 失効した購読情報の削除に失敗しました: userID=%s, endpoint=%s, status=%d, error=%v\n", logPrefix, sub.UserID, sub.Endpoint, statusCode, err)
		return
	}

	log.Printf("[%s] 失効した購読情報を削除しました: userID=%s, endpoint=%s, status=%d\n", logPrefix, sub.UserID, sub.Endpoint, statusCode)
}
