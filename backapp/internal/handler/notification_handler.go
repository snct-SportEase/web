package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	NotificationRepo repository.NotificationRepository
	EventRepo        repository.EventRepository
	RoleRepo         repository.RoleRepository
	VAPIDPublicKey   string
	VAPIDPrivateKey  string
}

func NewNotificationHandler(notificationRepo repository.NotificationRepository, eventRepo repository.EventRepository, roleRepo repository.RoleRepository, vapidPublicKey, vapidPrivateKey string) *NotificationHandler {
	return &NotificationHandler{
		NotificationRepo: notificationRepo,
		EventRepo:        eventRepo,
		RoleRepo:         roleRepo,
		VAPIDPublicKey:   vapidPublicKey,
		VAPIDPrivateKey:  vapidPrivateKey,
	}
}

type createNotificationRequest struct {
	Title       string   `json:"title"`
	Body        string   `json:"body"`
	TargetRoles []string `json:"target_roles"`
}

func (h *NotificationHandler) CreateNotification(c *gin.Context) {
	var req createNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	req.Body = strings.TrimSpace(req.Body)

	if req.Title == "" || req.Body == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "タイトルと本文は必須です"})
		return
	}

	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return
	}

	availableRoles, err := h.RoleRepo.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ロール情報の取得に失敗しました"})
		return
	}

	targetRoles, err := normalizeTargetRoles(req.TargetRoles, availableRoles)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activeEventID, err := h.EventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "アクティブな大会情報の取得に失敗しました"})
		return
	}

	var eventIDPtr *int
	if activeEventID != 0 {
		eventIDPtr = &activeEventID
	}

	notificationID, err := h.NotificationRepo.CreateNotification(req.Title, req.Body, user.ID, eventIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "通知の作成に失敗しました"})
		return
	}

	if err := h.NotificationRepo.AddNotificationTargets(notificationID, targetRoles); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "通知の対象ロール登録に失敗しました"})
		return
	}

	go h.dispatchPushNotifications(int(notificationID), req.Title, req.Body, targetRoles)

	c.JSON(http.StatusCreated, gin.H{
		"message":        "通知を送信しました",
		"notificationId": notificationID,
	})
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return
	}

	var userRoles []string
	for _, role := range user.Roles {
		userRoles = append(userRoles, role.Name)
	}

	includeAuthored := false
	if flag := c.Query("include_authored"); flag != "" {
		if v, err := strconv.ParseBool(flag); err == nil {
			includeAuthored = v
		}
	}

	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	notifications, err := h.NotificationRepo.GetNotificationsForAccess(userRoles, user.ID, includeAuthored, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "通知の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *NotificationHandler) ListAvailableRoles(c *gin.Context) {
	roles, err := h.RoleRepo.GetAllRoles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ロール一覧の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

type subscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	Keys     struct {
		Auth   string `json:"auth"`
		P256dh string `json:"p256dh"`
	} `json:"keys"`
}

func (h *NotificationHandler) SaveSubscription(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return
	}

	var req subscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正な購読情報です"})
		return
	}

	req.Endpoint = strings.TrimSpace(req.Endpoint)
	req.Keys.Auth = strings.TrimSpace(req.Keys.Auth)
	req.Keys.P256dh = strings.TrimSpace(req.Keys.P256dh)

	if req.Endpoint == "" || req.Keys.Auth == "" || req.Keys.P256dh == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "購読情報が不足しています"})
		return
	}

	if err := h.NotificationRepo.UpsertPushSubscription(user.ID, req.Endpoint, req.Keys.Auth, req.Keys.P256dh); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "購読情報の保存に失敗しました"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "購読情報を保存しました"})
}

func (h *NotificationHandler) DeleteSubscription(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}

	req.Endpoint = strings.TrimSpace(req.Endpoint)
	if req.Endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "エンドポイントは必須です"})
		return
	}

	if err := h.NotificationRepo.DeletePushSubscription(user.ID, req.Endpoint); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "購読情報の削除に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "購読情報を削除しました"})
}

func (h *NotificationHandler) GetSubscription(c *gin.Context) {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return
	}

	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return
	}

	subs, err := h.NotificationRepo.GetPushSubscriptionsByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "購読情報の取得に失敗しました"})
		return
	}

	// エンドポイントのみを返す（セキュリティのため）
	endpoints := make([]string, 0, len(subs))
	for _, sub := range subs {
		endpoints = append(endpoints, sub.Endpoint)
	}

	c.JSON(http.StatusOK, gin.H{
		"subscribed": len(subs) > 0,
		"endpoints":  endpoints,
		"count":      len(subs),
	})
}

func (h *NotificationHandler) dispatchPushNotifications(notificationID int, title, body string, targetRoles []string) {
	if h.VAPIDPrivateKey == "" || h.VAPIDPublicKey == "" {
		log.Println("[notification] VAPIDキーが設定されていないためPush通知をスキップします")
		return
	}

	userIDs, err := h.NotificationRepo.GetUserIDsByRoles(targetRoles)
	if err != nil {
		log.Printf("[notification] ユーザー抽出に失敗しました: %v\n", err)
		return
	}
	if len(userIDs) == 0 {
		return
	}

	subs, err := h.NotificationRepo.GetPushSubscriptionsByUserIDs(userIDs)
	if err != nil {
		log.Printf("[notification] 購読情報の取得に失敗しました: %v\n", err)
		return
	}

	if len(subs) == 0 {
		return
	}

	payload, err := json.Marshal(gin.H{
		"title": title,
		"body":  body,
		"data": gin.H{
			"notificationId": notificationID,
		},
	})
	if err != nil {
		log.Printf("[notification] Pushペイロード生成に失敗しました: %v\n", err)
		return
	}

	for _, sub := range subs {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256dhKey,
			},
		}

		resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
			Subscriber:      "mailto:notifications@sportease.local",
			VAPIDPublicKey:  h.VAPIDPublicKey,
			VAPIDPrivateKey: h.VAPIDPrivateKey,
			TTL:             60,
		})

		if err != nil {
			log.Printf("[notification] Push送信に失敗しました: endpoint=%s, error=%v\n", sub.Endpoint, err)
			continue
		}

		if resp != nil {
			if resp.StatusCode >= 400 {
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("[notification] Push送信が失敗しました: endpoint=%s, status=%d, body=%s\n", sub.Endpoint, resp.StatusCode, string(bodyBytes))
			} else {
				log.Printf("[notification] Push送信成功: endpoint=%s, status=%d\n", sub.Endpoint, resp.StatusCode)
			}
			resp.Body.Close()
		}
	}
}

func normalizeTargetRoles(requested []string, available []models.Role) ([]string, error) {
	if len(requested) == 0 {
		return nil, errors.New("少なくとも1つのロールを選択してください")
	}

	if len(available) == 0 {
		return nil, errors.New("ロールが登録されていません")
	}

	mapping := make(map[string]string, len(available))
	for _, role := range available {
		name := strings.TrimSpace(role.Name)
		if name == "" {
			continue
		}
		mapping[strings.ToLower(name)] = name
	}

	unique := make(map[string]struct{})
	for _, roleName := range requested {
		roleName = strings.TrimSpace(roleName)
		if roleName == "" {
			continue
		}

		normalized, ok := mapping[strings.ToLower(roleName)]
		if !ok {
			return nil, errors.New("存在しないロールが含まれています")
		}

		unique[normalized] = struct{}{}
	}

	if len(unique) == 0 {
		return nil, errors.New("少なくとも1つのロールを選択してください")
	}

	result := make([]string, 0, len(unique))
	for role := range unique {
		result = append(result, role)
	}
	sort.Strings(result)
	return result, nil
}
