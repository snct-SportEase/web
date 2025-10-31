package handler

import (
	"backapp/internal/models"
	"backapp/internal/repository"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SportHandler handles sport-related API requests.
type SportHandler struct {
	sportRepo repository.SportRepository
	classRepo repository.ClassRepository
	teamRepo  repository.TeamRepository
	eventRepo repository.EventRepository
	tournRepo repository.TournamentRepository
}

// NewSportHandler creates a new instance of SportHandler.
func NewSportHandler(sportRepo repository.SportRepository, classRepo repository.ClassRepository, teamRepo repository.TeamRepository, eventRepo repository.EventRepository, tournRepo repository.TournamentRepository) *SportHandler {
	return &SportHandler{
		sportRepo: sportRepo,
		classRepo: classRepo,
		teamRepo:  teamRepo,
		eventRepo: eventRepo,
		tournRepo: tournRepo,
	}
}

// GetAllSportsHandler handles the request to get all sports.
func (h *SportHandler) GetAllSportsHandler(c *gin.Context) {
	sports, err := h.sportRepo.GetAllSports()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sports"})
		return
	}

	if sports == nil {
		c.JSON(http.StatusOK, []*models.Sport{})
		return
	}

	c.JSON(http.StatusOK, sports)
}

// CreateSportHandler handles the request to create a new sport.
func (h *SportHandler) CreateSportHandler(c *gin.Context) {
	var sport models.Sport
	if err := c.ShouldBindJSON(&sport); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if sport.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sport name is required"})
		return
	}

	// Create the sport
	id, err := h.sportRepo.CreateSport(&sport)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sport"})
		return
	}
	sport.ID = int(id)

	c.JSON(http.StatusCreated, sport)
}

// GetSportsByEventHandler handles the request to get all sports for a specific event.
func (h *SportHandler) GetSportsByEventHandler(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	eventSports, err := h.sportRepo.GetSportsByEventID(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sports for the event"})
		return
	}

	c.JSON(http.StatusOK, eventSports)
}

// AssignSportToEventHandler handles the request to assign a sport to an event.
func (h *SportHandler) AssignSportToEventHandler(c *gin.Context) {
	eventIDStr := c.Param("id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var eventSport models.EventSport
	if err := c.ShouldBindJSON(&eventSport); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Set the event_id from the URL parameter
	eventSport.EventID = eventID

	if err := h.sportRepo.AssignSportToEvent(&eventSport); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign sport to the event"})
		return
	}

	// Get sport details
	sport, err := h.sportRepo.GetSportByID(eventSport.SportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sport details"})
		return
	}

	// Get all classes for the event
	classes, err := h.classRepo.GetAllClasses(eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve classes"})
		return
	}

	// Create teams for each class
	for _, class := range classes {
		team := &models.Team{
			Name:    fmt.Sprintf("%s", class.Name),
			ClassID: class.ID,
			SportID: sport.ID,
			EventID: eventID,
		}
		if _, err := h.teamRepo.CreateTeam(team); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Sport assigned to event successfully"})
}

// DeleteSportFromEventHandler handles the request to delete a sport from an event.
func (h *SportHandler) DeleteSportFromEventHandler(c *gin.Context) {
	eventIDStr := c.Param("event_id")
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	sportIDStr := c.Param("sport_id")
	sportID, err := strconv.Atoi(sportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	// Delete tournaments associated with the sport and event
	if err := h.tournRepo.DeleteTournamentsByEventAndSportID(eventID, sportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tournaments for the sport"})
		return
	}

	// Delete teams associated with the sport and event
	if err := h.teamRepo.DeleteTeamsByEventAndSportID(eventID, sportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete teams for the sport"})
		return
	}

	// Finally, delete the sport from the event
	if err := h.sportRepo.DeleteSportFromEvent(eventID, sportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sport from the event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sport deleted from event successfully"})
}

// GetTeamsBySportHandler handles the request to get all teams for a specific sport.
func (h *SportHandler) GetTeamsBySportHandler(c *gin.Context) {
	sportIDStr := c.Param("id")
	sportID, err := strconv.Atoi(sportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	teams, err := h.sportRepo.GetTeamsBySportID(sportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve teams for the sport"})
		return
	}

	c.JSON(http.StatusOK, teams)
}

type UpdateSportDetailsRequest struct {
	Description string `json:"description"`
	Rules       string `json:"rules"`
}

func (h *SportHandler) GetSportDetailsHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	sportID, err := strconv.Atoi(c.Param("sport_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	details, err := h.sportRepo.GetSportDetails(eventID, sportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sport details"})
		return
	}

	c.JSON(http.StatusOK, details)
}

func (h *SportHandler) UpdateSportDetailsHandler(c *gin.Context) {
	eventID, err := strconv.Atoi(c.Param("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}
	sportID, err := strconv.Atoi(c.Param("sport_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sport ID"})
		return
	}

	var req UpdateSportDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if err := h.sportRepo.UpdateSportDetails(eventID, sportID, req.Description, req.Rules); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sport details"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sport details updated successfully"})
}
