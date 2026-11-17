package middleware

import (
	"backapp/internal/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventVisibilityReader interface {
	GetActiveEvent() (int, error)
	GetEventByID(id int) (*models.Event, error)
}

// EventVisibilityRequired prevents students from enumerating resources that
// belong to an event which has not been announced yet. It supports the event
// identifiers used by the public read routes: :event_id, :id, or ?event_id=.
// When none is present, the currently active event is checked.
func EventVisibilityRequired(eventRepo EventVisibilityReader) gin.HandlerFunc {
	return func(c *gin.Context) {
		userValue, exists := c.Get("user")
		user, ok := userValue.(*models.User)
		if !exists || !ok || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in context"})
			c.Abort()
			return
		}
		if hasEventManagementRole(user) {
			c.Next()
			return
		}

		eventIDText := c.Param("event_id")
		if eventIDText == "" {
			eventIDText = c.Param("id")
		}
		if eventIDText == "" {
			eventIDText = c.Query("event_id")
		}

		var eventID int
		var err error
		if eventIDText == "" {
			eventID, err = eventRepo.GetActiveEvent()
		} else {
			eventID, err = strconv.Atoi(eventIDText)
		}
		if err != nil || eventID <= 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			c.Abort()
			return
		}

		event, err := eventRepo.GetEventByID(eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get event"})
			c.Abort()
			return
		}
		if event == nil || event.Status == models.EventStatusPreparing || event.Status == models.EventStatusTesting {
			c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func hasEventManagementRole(user *models.User) bool {
	for _, role := range user.Roles {
		if role.Name == "admin" || role.Name == "root" {
			return true
		}
	}
	return false
}
