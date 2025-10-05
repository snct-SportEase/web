package handler

import (
	"backapp/internal/repository"
	"encoding/csv"
	"net/http"

	"github.com/gin-gonic/gin"
)

type WhitelistHandler struct {
	WhitelistRepo repository.WhitelistRepository
	EventRepo     repository.EventRepository
}

func NewWhitelistHandler(whitelistRepo repository.WhitelistRepository, eventRepo repository.EventRepository) *WhitelistHandler {
	return &WhitelistHandler{
		WhitelistRepo: whitelistRepo,
		EventRepo:     eventRepo,
	}
}

func (h *WhitelistHandler) GetWhitelistHandler(c *gin.Context) {
	entries, err := h.WhitelistRepo.GetAllWhitelistedEmails()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve whitelist"})
		return
	}
	c.JSON(http.StatusOK, entries)
}

func (h *WhitelistHandler) AddWhitelistedEmailHandler(c *gin.Context) {
	var entry repository.WhitelistEntry
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if entry.Email == "" || entry.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email and role are required"})
		return
	}

	// Get active event_id
	activeEventID, err := h.EventRepo.GetActiveEvent()
	if err != nil {
		// If there is an error other than "no rows", it's a server error.
		// "no rows" is handled by GetActiveEvent returning 0, so we don't need to check sql.ErrNoRows here.
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event", "details": err.Error()})
		return
	}

	var eventIDPtr *int
	if activeEventID != 0 {
		eventIDPtr = &activeEventID
	}

	// Add email with event_id
	if err := h.WhitelistRepo.AddWhitelistedEmail(entry.Email, entry.Role, eventIDPtr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add email to whitelist"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Email added to whitelist successfully"})
}

func (h *WhitelistHandler) BulkAddWhitelistedEmailsHandler(c *gin.Context) {
	file, _, err := c.Request.FormFile("csvfile")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "CSV file not found in request"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse CSV file"})
		return
	}

	// Get active event_id
	activeEventID, err := h.EventRepo.GetActiveEvent()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get active event", "details": err.Error()})
		return
	}

	var eventIDPtr *int
	if activeEventID != 0 {
		eventIDPtr = &activeEventID
	}

	var entries []repository.WhitelistEntry
	for i, record := range records {
		if i == 0 { // Skip header row
			continue
		}
		if len(record) < 2 {
			continue // Skip empty or invalid rows
		}
		entries = append(entries, repository.WhitelistEntry{
			Email:   record[0],
			Role:    record[1],
			EventID: eventIDPtr, // Set the active event ID
		})
	}

	if err := h.WhitelistRepo.AddWhitelistedEmails(entries); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add emails to whitelist from CSV"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully added emails from CSV"})
}
