package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"encoding/csv"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	eventRepo        repository.EventRepository
	whitelistRepo    repository.WhitelistRepository
	tournamentRepo   repository.TournamentRepository
	classRepo        repository.ClassRepository
	notificationRepo repository.NotificationRepository
	vapidPublicKey   string
	vapidPrivateKey  string
}

func NewEventHandler(eventRepo repository.EventRepository, whitelistRepo repository.WhitelistRepository, tournamentRepo repository.TournamentRepository, classRepo repository.ClassRepository, notificationRepo repository.NotificationRepository, vapidPublicKey string, vapidPrivateKey string) *EventHandler {
	return &EventHandler{
		eventRepo:        eventRepo,
		whitelistRepo:    whitelistRepo,
		tournamentRepo:   tournamentRepo,
		classRepo:        classRepo,
		notificationRepo: notificationRepo,
		vapidPublicKey:   vapidPublicKey,
		vapidPrivateKey:  vapidPrivateKey,
	}
}

func (h *EventHandler) CreateEvent(c *gin.Context) {
	var req struct {
		Name      string  `json:"name"`
		Year      int     `json:"year"`
		Season    string  `json:"season"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		SurveyUrl *string `json:"survey_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Date parsing
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
		endDate = &t
	}

	event := &models.Event{
		Name:       req.Name,
		Year:       req.Year,
		Season:     req.Season,
		Start_date: startDate,
		End_date:   endDate,
		SurveyUrl:  req.SurveyUrl,
	}

	id, err := h.eventRepo.CreateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	event.ID = int(id)

	if event.Season == "autumn" {
		springEvent, err := h.eventRepo.GetEventByYearAndSeason(event.Year, "spring")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if springEvent != nil {
			err := h.eventRepo.CopyClassScores(springEvent.ID, event.ID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
	}

	c.JSON(http.StatusCreated, event)
}

func (h *EventHandler) GetAllEvents(c *gin.Context) {
	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, events)
}

func (h *EventHandler) UpdateEvent(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		Name      string  `json:"name"`
		Year      int     `json:"year"`
		Season    string  `json:"season"`
		StartDate string  `json:"start_date"`
		EndDate   string  `json:"end_date"`
		SurveyUrl *string `json:"survey_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Date parsing
	var startDate, endDate *time.Time
	if req.StartDate != "" {
		t, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD."})
			return
		}
		startDate = &t
	}
	if req.EndDate != "" {
		t, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD."})
			return
		}
		endDate = &t
	}

	event := &models.Event{
		ID:         id,
		Name:       req.Name,
		Year:       req.Year,
		Season:     req.Season,
		Start_date: startDate,
		End_date:   endDate,
		SurveyUrl:  req.SurveyUrl,
	}

	err = h.eventRepo.UpdateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) GetActiveEvent(c *gin.Context) {
	event_id, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event_id == 0 {
		c.JSON(http.StatusOK, gin.H{"event_id": nil, "event_name": nil, "competition_guidelines_pdf_url": nil})
		return
	}

	event, err := h.eventRepo.GetEventByID(event_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusOK, gin.H{"event_id": nil, "event_name": nil, "competition_guidelines_pdf_url": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event_id":                       event.ID,
		"event_name":                     event.Name,
		"competition_guidelines_pdf_url": event.CompetitionGuidelinesPdfUrl,
		"survey_url":                     event.SurveyUrl,
		"is_survey_published":            event.IsSurveyPublished,
	})
}

func (h *EventHandler) SetActiveEvent(c *gin.Context) {
	// 1. リクエストボディの構造体を定義
	req := models.SetActiveEventRequest{}

	// 2. リクエストボディをパースし、バリデーション
	if err := c.ShouldBindJSON(&req); err != nil {
		// バリデーションエラーやJSONパースエラーの場合、400 Bad Requestを返す
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or missing event_id", "details": err.Error()})
		return
	}

	// Pass pointer to allow nil -> clearing active event
	err := h.eventRepo.SetActiveEvent(req.EventID)
	if err != nil {
		// DB更新に失敗した場合、500 Internal Server Errorを返す
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to set active event", "details": err.Error()})
		return
	}

	// If a new event is being activated (not cleared), update the whitelist
	if req.EventID != nil {
		if err := h.whitelistRepo.UpdateNullEventIDs(*req.EventID); err != nil {
			// このエラーはクリティカルではないので、ログには残すがクライアントには成功を返す
			// c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update whitelist event IDs", "details": err.Error()})
			// return
			// TODO: Add proper logging here
		}
	}

	// 4. 成功レスポンスを返す
	c.JSON(http.StatusOK, gin.H{
		"message":  "Active event set successfully",
		"event_id": req.EventID,
	})
}

func (h *EventHandler) SetRainyMode(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		IsRainyMode bool `json:"is_rainy_mode"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.eventRepo.SetRainyMode(eventID, req.IsRainyMode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If rainy mode is being enabled, apply rainy mode start times to tournaments
	if req.IsRainyMode {
		err = h.tournamentRepo.ApplyRainyModeStartTimes(eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to apply rainy mode start times", "details": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rainy mode updated successfully", "is_rainy_mode": req.IsRainyMode})
}

func (h *EventHandler) UpdateCompetitionGuidelines(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var req struct {
		PdfUrl *string `json:"pdf_url"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event, err := h.eventRepo.GetEventByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	event.CompetitionGuidelinesPdfUrl = req.PdfUrl
	err = h.eventRepo.UpdateEvent(event)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *EventHandler) NotifySurvey(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil || event.SurveyUrl == nil || *event.SurveyUrl == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Event has no survey URL configured"})
		return
	}

	// 1. Create a notification record
	title := "大会アンケートのお願い"
	body := "大会に関するアンケート機能が公開されました。「" + event.Name + "」についてダッシュボードの一番上のリンクからアンケートにご協力ください。"
	createdBy := "" // System notification
	userIDVal, exists := c.Get("user_id")
	if exists {
		createdBy = userIDVal.(string)
	}

	notifID, err := h.notificationRepo.CreateNotification(title, body, createdBy, &event.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	// Target all basic roles
	targetRoles := []string{"student", "admin", "root"}
	if err := h.notificationRepo.AddNotificationTargets(notifID, targetRoles); err != nil {
		// Log error but don't fail the request completely
	}

	// Publish the survey
	if err := h.eventRepo.PublishSurvey(eventID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish survey"})
		return
	}

	// Send push notifications
	go h.dispatchPushNotifications(int(notifID), "アンケート回答のお願い", "アンケートページが公開されました。期間内に回答してください。", targetRoles)

	c.JSON(http.StatusOK, gin.H{"message": "Survey notification sent successfully"})
}

func (h *EventHandler) dispatchPushNotifications(notificationID int, title, body string, targetRoles []string) {
	log.Printf("[event-notification] 通知送信開始: notificationID=%d, title=%s, targetRoles=%v\n", notificationID, title, targetRoles)

	if h.vapidPrivateKey == "" || h.vapidPublicKey == "" {
		log.Println("[event-notification] VAPIDキーが設定されていないためPush通知をスキップします")
		return
	}

	userIDs, err := h.notificationRepo.GetUserIDsByRoles(targetRoles)
	if err != nil {
		log.Printf("[event-notification] ユーザー抽出に失敗しました: %v\n", err)
		return
	}
	log.Printf("[event-notification] 対象ユーザー数: %d, userIDs=%v\n", len(userIDs), userIDs)
	if len(userIDs) == 0 {
		log.Println("[event-notification] 対象ユーザーが0人のためPush通知をスキップします")
		return
	}

	subs, err := h.notificationRepo.GetPushSubscriptionsByUserIDs(userIDs)
	if err != nil {
		log.Printf("[event-notification] 購読情報の取得に失敗しました: %v\n", err)
		return
	}

	log.Printf("[event-notification] 購読情報数: %d\n", len(subs))
	if len(subs) == 0 {
		log.Println("[event-notification] 購読情報が0件のためPush通知をスキップします")
		return
	}

	payload, err := json.Marshal(gin.H{
		"title": title,
		"body":  body,
		"data": gin.H{
			"url": "/dashboard",
		},
	})
	if err != nil {
		log.Printf("[event-notification] Pushペイロード生成に失敗しました: %v\n", err)
		return
	}

	log.Printf("[event-notification] %d件の購読に対してPush通知を送信します\n", len(subs))
	for i, sub := range subs {
		subscription := &webpush.Subscription{
			Endpoint: sub.Endpoint,
			Keys: webpush.Keys{
				Auth:   sub.AuthKey,
				P256dh: sub.P256dhKey,
			},
		}

		log.Printf("[event-notification] [%d/%d] Push送信試行: userID=%s, endpoint=%s\n", i+1, len(subs), sub.UserID, sub.Endpoint)

		resp, err := webpush.SendNotification(payload, subscription, &webpush.Options{
			Subscriber:      "mailto:notifications@sportease.local",
			VAPIDPublicKey:  h.vapidPublicKey,
			VAPIDPrivateKey: h.vapidPrivateKey,
			TTL:             60,
		})

		if err != nil {
			log.Printf("[event-notification] [%d/%d] Push送信に失敗しました: userID=%s, endpoint=%s, error=%v\n", i+1, len(subs), sub.UserID, sub.Endpoint, err)
			continue
		}

		if resp != nil {
			if resp.StatusCode >= 400 {
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("[event-notification] [%d/%d] Push送信が失敗しました: userID=%s, endpoint=%s, status=%d, body=%s\n", i+1, len(subs), sub.UserID, sub.Endpoint, resp.StatusCode, string(bodyBytes))
			} else {
				log.Printf("[event-notification] [%d/%d] Push送信成功: userID=%s, endpoint=%s, status=%d\n", i+1, len(subs), sub.UserID, sub.Endpoint, resp.StatusCode)
			}
			resp.Body.Close()
		}
	}
	log.Printf("[event-notification] 通知送信完了: notificationID=%d\n", notificationID)
}

func (h *EventHandler) ImportSurveyScores(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	event, err := h.eventRepo.GetEventByID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if event == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Season != "spring" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Survey score import is only allowed for the spring event"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	reader := csv.NewReader(src)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV file"})
		return
	}

	if len(records) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file does not contain enough data"})
		return
	}

	headers := records[0]
	classNameColIdx := -1
	for i, header := range headers {
		if strings.Contains(strings.ToLower(header), "クラス") || strings.Contains(header, "class") {
			classNameColIdx = i
			break
		}
	}

	if classNameColIdx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find class name column in CSV"})
		return
	}

	submissionCounts := make(map[string]int)
	for _, row := range records[1:] {
		if len(row) > classNameColIdx {
			className := strings.TrimSpace(row[classNameColIdx])
			if className != "" {
				submissionCounts[className]++
			}
		}
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch class info"})
		return
	}

	surveyPointsData := make(map[int]int)
	for _, class := range classes {
		if class.StudentCount > 0 {
			submissions := submissionCounts[class.Name]
			rate := float64(submissions) / float64(class.StudentCount)
			var points int
			switch {
			case rate >= 0.9:
				points = 10
			case rate >= 0.8:
				points = 9
			case rate >= 0.7:
				points = 8
			case rate >= 0.6:
				points = 7
			case rate >= 0.5:
				points = 6
			default:
				points = 5
			}
			surveyPointsData[class.ID] = points
		}
	}

	// Assuming a method like SetSurveyPoints exists or can be created to update 'survey_points'
	// Since we need to update score_logs, we'll implement it within this handler or class repository.
	// For simplicity, we can reuse SetNoonGamePoints logic or require a new method `SetSurveyPoints`.
	// Let's create `SetSurveyPoints` in ClassRepository just to be clean, or execute DB queries directly

	if err := h.classRepo.SetSurveyPoints(eventID, surveyPointsData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update survey points"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Survey scores imported successfully", "imported_classes_count": len(surveyPointsData)})
}
