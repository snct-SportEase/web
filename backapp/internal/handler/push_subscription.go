package handler

import (
	"backapp/internal/models"
	"backapp/internal/push"
	"backapp/internal/repository"
	"log"
	"net/http"
)

func cleanupExpiredPushSubscription(repo repository.NotificationRepository, sub models.PushSubscription, statusCode int, logPrefix string) {
	if statusCode != http.StatusGone && statusCode != http.StatusNotFound {
		return
	}

	endpointID := push.EndpointLogID(sub.Endpoint)
	if err := repo.DeletePushSubscription(sub.UserID, sub.Endpoint); err != nil {
		log.Printf("[%s] 失効した購読情報の削除に失敗しました: userID=%s, endpointID=%s, status=%d, errorType=%T\n", logPrefix, sub.UserID, endpointID, statusCode, err)
		return
	}

	log.Printf("[%s] 失効した購読情報を削除しました: userID=%s, endpointID=%s, status=%d\n", logPrefix, sub.UserID, endpointID, statusCode)
}
