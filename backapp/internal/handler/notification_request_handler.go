package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

type NotificationRequestHandler struct {
	RequestRepo       repository.NotificationRequestRepository
	NotificationRepo  repository.NotificationRepository
	RoleRepo          repository.RoleRepository
	WebPushPublicKey  string
	WebPushPrivateKey string
}

func NewNotificationRequestHandler(
	requestRepo repository.NotificationRequestRepository,
	notificationRepo repository.NotificationRepository,
	roleRepo repository.RoleRepository,
	publicKey string,
	privateKey string,
) *NotificationRequestHandler {
	return &NotificationRequestHandler{
		RequestRepo:       requestRepo,
		NotificationRepo:  notificationRepo,
		RoleRepo:          roleRepo,
		WebPushPublicKey:  publicKey,
		WebPushPrivateKey: privateKey,
	}
}

type createNotificationRequestPayload struct {
	Title      string `json:"title"`
	Body       string `json:"body"`
	TargetText string `json:"target_text"`
}

func (h *NotificationRequestHandler) CreateRequest(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		return
	}

	var payload createNotificationRequestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}

	payload.Title = strings.TrimSpace(payload.Title)
	payload.Body = strings.TrimSpace(payload.Body)
	payload.TargetText = strings.TrimSpace(payload.TargetText)

	if payload.Title == "" || payload.Body == "" || payload.TargetText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "タイトル・内容・対象情報は必須です"})
		return
	}

	req := &models.NotificationRequest{
		Title:       payload.Title,
		Body:        payload.Body,
		TargetText:  payload.TargetText,
		Status:      models.NotificationRequestStatusPending,
		RequesterID: user.ID,
	}

	requestID, err := h.RequestRepo.CreateRequest(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請の作成に失敗しました"})
		return
	}

	go h.notifyRootsOfNewRequest(int(requestID), req)

	c.JSON(http.StatusCreated, gin.H{"request_id": requestID})
}

func (h *NotificationRequestHandler) ListStudentRequests(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		return
	}

	requests, err := h.RequestRepo.GetRequestsByRequester(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請一覧の取得に失敗しました"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func (h *NotificationRequestHandler) ListRootRequests(c *gin.Context) {
	requests, err := h.RequestRepo.GetAllRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請一覧の取得に失敗しました"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

func (h *NotificationRequestHandler) GetRequestDetail(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		return
	}

	requestID, ok := parseIDParam(c, "request_id")
	if !ok {
		return
	}

	req, err := h.RequestRepo.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請の取得に失敗しました"})
		return
	}
	if req == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "対象の申請が見つかりません"})
		return
	}

	if !isRoot(user) && req.RequesterID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "申請へのアクセス権限がありません"})
		return
	}

	messages, err := h.RequestRepo.GetMessages(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "メッセージの取得に失敗しました"})
		return
	}

	req.Messages = messages

	c.JSON(http.StatusOK, gin.H{"request": req})
}

type addMessagePayload struct {
	Message string `json:"message"`
}

func (h *NotificationRequestHandler) AddMessage(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		return
	}

	requestID, ok := parseIDParam(c, "request_id")
	if !ok {
		return
	}

	req, err := h.RequestRepo.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請の取得に失敗しました"})
		return
	}
	if req == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "対象の申請が見つかりません"})
		return
	}

	isRootUser := isRoot(user)
	if !isRootUser && req.RequesterID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "メッセージを投稿する権限がありません"})
		return
	}

	var payload addMessagePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}
	payload.Message = strings.TrimSpace(payload.Message)
	if payload.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "メッセージは必須です"})
		return
	}

	_, err = h.RequestRepo.AddMessage(requestID, user.ID, payload.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "メッセージの投稿に失敗しました"})
		return
	}

	go h.notifyParticipantsOfMessage(req, user, payload.Message, isRootUser)

	c.JSON(http.StatusCreated, gin.H{"message": "メッセージを送信しました"})
}

type decisionPayload struct {
	Status string `json:"status"`
}

func (h *NotificationRequestHandler) DecideRequest(c *gin.Context) {
	user := currentUser(c)
	if user == nil {
		return
	}

	requestID, ok := parseIDParam(c, "request_id")
	if !ok {
		return
	}

	req, err := h.RequestRepo.GetRequestByID(requestID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "申請の取得に失敗しました"})
		return
	}
	if req == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "対象の申請が見つかりません"})
		return
	}

	var payload decisionPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なリクエスト形式です"})
		return
	}

	status := models.NotificationRequestStatus(payload.Status)
	if status != models.NotificationRequestStatusApproved && status != models.NotificationRequestStatusRejected {
		c.JSON(http.StatusBadRequest, gin.H{"error": "承認または否認のみ指定できます"})
		return
	}

	resolverID := user.ID
	if err := h.RequestRepo.UpdateRequestStatus(requestID, status, &resolverID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ステータス更新に失敗しました"})
		return
	}

	go h.notifyRequesterDecision(req, status, user)

	c.JSON(http.StatusOK, gin.H{"message": "処理を完了しました"})
}

func (h *NotificationRequestHandler) notifyRootsOfNewRequest(requestID int, req *models.NotificationRequest) {
	if h.WebPushPrivateKey == "" || h.WebPushPublicKey == "" {
		return
	}

	rootIDs, err := h.NotificationRepo.GetUserIDsByRoles([]string{"root"})
	if err != nil || len(rootIDs) == 0 {
		return
	}

	payload := gin.H{
		"title": "新しい通知申請",
		"body":  req.Title,
		"data": gin.H{
			"type":        "notification_request",
			"requestId":   requestID,
			"description": req.TargetText,
		},
	}
	h.sendPushToUsers(rootIDs, payload)
}

func (h *NotificationRequestHandler) notifyParticipantsOfMessage(req *models.NotificationRequest, sender *models.User, message string, senderIsRoot bool) {
	if h.WebPushPrivateKey == "" || h.WebPushPublicKey == "" {
		return
	}

	targetIDs := make([]string, 0)
	if senderIsRoot {
		targetIDs = append(targetIDs, req.RequesterID)
	} else {
		rootIDs, err := h.NotificationRepo.GetUserIDsByRoles([]string{"root"})
		if err == nil {
			targetIDs = append(targetIDs, rootIDs...)
		}
	}

	if len(targetIDs) == 0 {
		return
	}

	payload := gin.H{
		"title": "申請チャットに新着メッセージ",
		"body":  message,
		"data": gin.H{
			"type":      "notification_request_message",
			"requestId": req.ID,
			"sender":    userDisplayName(sender),
		},
	}
	h.sendPushToUsers(targetIDs, payload)
}

func (h *NotificationRequestHandler) notifyRequesterDecision(req *models.NotificationRequest, status models.NotificationRequestStatus, resolver *models.User) {
	if h.WebPushPrivateKey == "" || h.WebPushPublicKey == "" {
		return
	}

	message := "申請が承認されました"
	if status == models.NotificationRequestStatusRejected {
		message = "申請が否認されました"
	}

	payload := gin.H{
		"title": "通知申請の結果",
		"body":  message,
		"data": gin.H{
			"type":      "notification_request_decision",
			"requestId": req.ID,
			"status":    status,
		},
	}
	h.sendPushToUsers([]string{req.RequesterID}, payload)

	resultText := "承認"
	if status == models.NotificationRequestStatusRejected {
		resultText = "否認"
	}

	systemMessage := fmt.Sprintf("rootが申請を%sしました。", resultText)
	_, _ = h.RequestRepo.AddMessage(req.ID, resolver.ID, systemMessage)
}

func (h *NotificationRequestHandler) sendPushToUsers(userIDs []string, payload gin.H) {
	if len(userIDs) == 0 {
		return
	}

	subscriptions, err := h.NotificationRepo.GetPushSubscriptionsByUserIDs(userIDs)
	if err != nil || len(subscriptions) == 0 {
		return
	}

	bodyBytes, err := jsonMarshal(payload)
	if err != nil {
		return
	}

	for _, sub := range subscriptions {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256dhKey,
			},
		}
		resp, err := webpush.SendNotification(bodyBytes, subscription, &webpush.Options{
			Subscriber:      "mailto:notifications@sportease.local",
			VAPIDPublicKey:  h.WebPushPublicKey,
			VAPIDPrivateKey: h.WebPushPrivateKey,
			TTL:             60,
		})
		if err == nil && resp != nil {
			resp.Body.Close()
		}
	}
}

func userDisplayName(user *models.User) string {
	if user == nil {
		return ""
	}
	if user.DisplayName != nil && *user.DisplayName != "" {
		return *user.DisplayName
	}
	if user.Email != "" {
		return user.Email
	}
	return user.ID
}

func parseIDParam(c *gin.Context, param string) (int, bool) {
	value := c.Param(param)
	id, err := strconv.Atoi(value)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不正なIDです"})
		return 0, false
	}
	return id, true
}

func currentUser(c *gin.Context) *models.User {
	userValue, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "ユーザー情報を取得できませんでした"})
		return nil
	}
	user, ok := userValue.(*models.User)
	if !ok || user == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ユーザー情報の解析に失敗しました"})
		return nil
	}
	return user
}

func isRoot(user *models.User) bool {
	for _, role := range user.Roles {
		if role.Name == "root" {
			return true
		}
	}
	return false
}

func jsonMarshal(v interface{}) ([]byte, error) {
	type jsonMarshaler interface {
		MarshalJSON() ([]byte, error)
	}
	if m, ok := v.(jsonMarshaler); ok {
		return m.MarshalJSON()
	}
	return json.Marshal(v)
}
