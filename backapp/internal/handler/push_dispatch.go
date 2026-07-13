package handler

import (
	"context"
	"log"
	"time"

	"backapp/internal/models"
	"backapp/internal/push"
	"backapp/internal/repository"
)

func dispatchPushBatch(sender push.Sender, repo repository.NotificationRepository, payload []byte, subscriptions []models.PushSubscription, ttl int, logPrefix string) {
	if sender == nil || !sender.Enabled() || len(subscriptions) == 0 {
		return
	}

	log.Printf("[%s] %d件の購読に対してPush通知を送信します\n", logPrefix, len(subscriptions))
	batchContext, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	results := sender.SendBatch(batchContext, payload, subscriptions, ttl)
	for index, result := range results {
		endpointID := push.EndpointLogID(result.Subscription.Endpoint)
		if result.InvalidSubscription {
			log.Printf("[%s] [%d/%d] 不正な購読情報を削除します: userID=%s, endpointID=%s\n", logPrefix, index+1, len(results), result.Subscription.UserID, endpointID)
			if err := repo.DeletePushSubscription(result.Subscription.UserID, result.Subscription.Endpoint); err != nil {
				log.Printf("[%s] 不正な購読情報の削除に失敗しました: userID=%s, endpointID=%s, errorType=%T\n", logPrefix, result.Subscription.UserID, endpointID, err)
			}
			continue
		}
		if result.Err != nil {
			// Network errors can embed the capability URL, so log only the type.
			log.Printf("[%s] [%d/%d] Push送信に失敗しました: userID=%s, endpointID=%s, errorType=%T\n", logPrefix, index+1, len(results), result.Subscription.UserID, endpointID, result.Err)
			continue
		}
		if result.StatusCode >= 400 {
			log.Printf("[%s] [%d/%d] Pushサービスがエラーを返しました: userID=%s, endpointID=%s, status=%d\n", logPrefix, index+1, len(results), result.Subscription.UserID, endpointID, result.StatusCode)
			cleanupExpiredPushSubscription(repo, result.Subscription, result.StatusCode, logPrefix)
			continue
		}
		log.Printf("[%s] [%d/%d] Push送信成功: userID=%s, endpointID=%s, status=%d\n", logPrefix, index+1, len(results), result.Subscription.UserID, endpointID, result.StatusCode)
	}
}
