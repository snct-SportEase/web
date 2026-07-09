package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatisticsHandler struct {
	classRepo repository.ClassRepository
	eventRepo repository.EventRepository
	sportRepo repository.SportRepository
	tournRepo repository.TournamentRepository
}

func NewStatisticsHandler(classRepo repository.ClassRepository, eventRepo repository.EventRepository, sportRepo repository.SportRepository, tournRepo repository.TournamentRepository) *StatisticsHandler {
	return &StatisticsHandler{
		classRepo: classRepo,
		eventRepo: eventRepo,
		sportRepo: sportRepo,
		tournRepo: tournRepo,
	}
}

func (h *StatisticsHandler) GetOverallAttendanceRate(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
		return
	}

	totalAttendance := 0
	totalStudents := 0
	for _, class := range classes {
		totalAttendance += class.AttendCount
		totalStudents += class.StudentCount
	}

	rate := 0.0
	if totalStudents > 0 {
		rate = float64(totalAttendance) / float64(totalStudents) * 100
	}

	c.JSON(http.StatusOK, gin.H{"attendance_rate": rate})
}

func (h *StatisticsHandler) GetParticipationRateBySport(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	sports, err := h.sportRepo.GetAllSports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sports"})
		return
	}

	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get classes"})
		return
	}
	totalPossible := len(classes)

	participantCounts, err := h.tournRepo.CountTeamsBySportForEvent(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get participation counts"})
		return
	}

	result := make(map[string]float64)
	for _, sport := range sports {
		participantCount := participantCounts[sport.ID]
		rate := 0.0
		if totalPossible > 0 {
			rate = float64(participantCount) / float64(totalPossible) * 100
		}
		result[sport.Name] = rate
	}

	c.JSON(http.StatusOK, result)
}

func (h *StatisticsHandler) GetClassScoreTrends(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	event, err := h.eventRepo.GetEventByID(eventID)
	if err != nil || event == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event"})
		return
	}

	if event.HideScores {
		c.JSON(http.StatusOK, gin.H{"message": "Scores are hidden for this event"})
		return
	}

	events, err := h.eventRepo.GetAllEvents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get events"})
		return
	}

	eventIDs := make([]int, 0, len(events))
	for _, event := range events {
		eventIDs = append(eventIDs, event.ID)
	}

	// イベントごとのスコア取得を1回のIN句クエリにまとめ、統計画面のN+1を避ける。
	scoresByEventID, err := h.classRepo.GetClassScoresByEvents(eventIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get class scores"})
		return
	}

	result := make(map[string][]*models.ClassScore)
	for _, event := range events {
		result[event.Name] = scoresByEventID[event.ID]
	}

	c.JSON(http.StatusOK, result)
}

func (h *StatisticsHandler) GetRealtimeEventProgress(c *gin.Context) {
	eventID, err := h.eventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
		return
	}

	// 進捗表示ではトーナメント詳細の組み立ては不要なので、sport名だけJOINで軽く取得する。
	sportNames, err := h.tournRepo.GetTournamentSportNamesByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tournaments"})
		return
	}

	progress := make(map[string]string)
	for _, sportName := range sportNames {
		progress[sportName] = "進行中" // Assuming status
	}

	c.JSON(http.StatusOK, progress)
}
