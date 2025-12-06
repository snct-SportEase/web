package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWhitelistHandler_DeleteWhitelistedEmailHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Delete single email", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Email string `json:"email"`
		}{
			Email: "test@example.com",
		}

		mockWhitelistRepo.On("DeleteWhitelistedEmail", "test@example.com").Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Email deleted from whitelist successfully", response["message"])
		mockWhitelistRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid request body")
		mockWhitelistRepo.AssertNotCalled(t, "DeleteWhitelistedEmail", mock.Anything)
	})

	t.Run("Error - Empty email", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Email string `json:"email"`
		}{
			Email: "",
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		// binding:"required"が先にチェックされるため、"Invalid request body"が返される
		assert.Contains(t, response["error"], "Invalid request body")
		mockWhitelistRepo.AssertNotCalled(t, "DeleteWhitelistedEmail", mock.Anything)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Email string `json:"email"`
		}{
			Email: "test@example.com",
		}

		mockWhitelistRepo.On("DeleteWhitelistedEmail", "test@example.com").Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to delete email from whitelist", response["error"])
		mockWhitelistRepo.AssertExpectations(t)
	})
}

func TestWhitelistHandler_DeleteWhitelistedEmailsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Delete multiple emails", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Emails []string `json:"emails"`
		}{
			Emails: []string{"test1@example.com", "test2@example.com", "test3@example.com"},
		}

		mockWhitelistRepo.On("DeleteWhitelistedEmails", reqBody.Emails).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist/bulk", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Emails deleted from whitelist successfully", response["message"])
		mockWhitelistRepo.AssertExpectations(t)
	})

	t.Run("Success - Delete single email in bulk", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Emails []string `json:"emails"`
		}{
			Emails: []string{"test@example.com"},
		}

		mockWhitelistRepo.On("DeleteWhitelistedEmails", reqBody.Emails).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist/bulk", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailsHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Emails deleted from whitelist successfully", response["message"])
		mockWhitelistRepo.AssertExpectations(t)
	})

	t.Run("Error - Invalid request body", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist/bulk", bytes.NewBuffer([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailsHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, response["error"], "Invalid request body")
		mockWhitelistRepo.AssertNotCalled(t, "DeleteWhitelistedEmails", mock.Anything)
	})

	t.Run("Error - Empty emails array", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Emails []string `json:"emails"`
		}{
			Emails: []string{},
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist/bulk", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailsHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "At least one email is required", response["error"])
		mockWhitelistRepo.AssertNotCalled(t, "DeleteWhitelistedEmails", mock.Anything)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		reqBody := struct {
			Emails []string `json:"emails"`
		}{
			Emails: []string{"test1@example.com", "test2@example.com"},
		}

		mockWhitelistRepo.On("DeleteWhitelistedEmails", reqBody.Emails).Return(assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		jsonBody, _ := json.Marshal(reqBody)
		c.Request, _ = http.NewRequest(http.MethodDelete, "/api/root/whitelist/bulk", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		h.DeleteWhitelistedEmailsHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to delete emails from whitelist", response["error"])
		mockWhitelistRepo.AssertExpectations(t)
	})
}

func TestWhitelistHandler_GetWhitelistHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Success - Get all whitelisted emails", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		expectedEntries := []repository.WhitelistEntry{
			{Email: "test1@example.com", Role: "student", EventID: nil},
			{Email: "test2@example.com", Role: "admin", EventID: nil},
		}

		mockWhitelistRepo.On("GetAllWhitelistedEmails").Return(expectedEntries, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/api/root/whitelist", nil)

		h.GetWhitelistHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []repository.WhitelistEntry
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedEntries, response)
		mockWhitelistRepo.AssertExpectations(t)
	})

	t.Run("Error - Repository error", func(t *testing.T) {
		mockWhitelistRepo := new(MockWhitelistRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewWhitelistHandler(mockWhitelistRepo, mockEventRepo)

		// nilの代わりに空のスライスを返す必要がある
		mockWhitelistRepo.On("GetAllWhitelistedEmails").Return([]repository.WhitelistEntry{}, assert.AnError).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest(http.MethodGet, "/api/root/whitelist", nil)

		h.GetWhitelistHandler(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Failed to retrieve whitelist", response["error"])
		mockWhitelistRepo.AssertExpectations(t)
	})
}
