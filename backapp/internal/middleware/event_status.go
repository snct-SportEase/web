package middleware

import (
	"backapp/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ActiveEventStatusRequired restricts an operation to the currently selected
// event when it has one of the allowed statuses. It is used for operations
// that change results, so a preparation event can be configured but cannot
// accidentally receive scores.
func ActiveEventStatusRequired(eventRepo repository.EventRepository, allowedStatuses ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		eventID, err := eventRepo.GetActiveEvent()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
			c.Abort()
			return
		}
		if eventID == 0 {
			c.JSON(http.StatusForbidden, gin.H{"error": "No active event found"})
			c.Abort()
			return
		}

		event, err := eventRepo.GetEventByID(eventID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event"})
			c.Abort()
			return
		}
		if event == nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "No active event found"})
			c.Abort()
			return
		}

		for _, status := range allowedStatuses {
			if event.Status == status {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Match results can only be entered while the event is active"})
		c.Abort()
	}
}
