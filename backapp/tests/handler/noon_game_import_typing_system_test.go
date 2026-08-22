package handler_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"backapp/internal/handler"
	"backapp/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type typingSystemTeam struct {
	TeamName    string `json:"team_name"`
	Match1Score int    `json:"match_1_score"`
	Match2Score int    `json:"match_2_score"`
	Match3Score int    `json:"match_3_score"`
	TotalScore  int    `json:"total_score"`
	Rank        int    `json:"rank"`
}

type typingSystemPayload struct {
	SchemaVersion string             `json:"schema_version"`
	ExportID      string             `json:"export_id"`
	Teams         []typingSystemTeam `json:"teams"`
}

func buildTypingSystemImportClasses(eventID int) []*models.Class {
	return []*models.Class{
		{ID: 1, Name: "1-1", EventID: &eventID},
		{ID: 2, Name: "1-2", EventID: &eventID},
		{ID: 3, Name: "1-3", EventID: &eventID},
		{ID: 4, Name: "IS2", EventID: &eventID},
		{ID: 5, Name: "IT2", EventID: &eventID},
		{ID: 6, Name: "IE2", EventID: &eventID},
		{ID: 7, Name: "IS3", EventID: &eventID},
		{ID: 8, Name: "IT3", EventID: &eventID},
		{ID: 9, Name: "IE3", EventID: &eventID},
		{ID: 10, Name: "IS4", EventID: &eventID},
		{ID: 11, Name: "IT4", EventID: &eventID},
		{ID: 12, Name: "IE4", EventID: &eventID},
		{ID: 13, Name: "IS5", EventID: &eventID},
		{ID: 14, Name: "IT5", EventID: &eventID},
		{ID: 15, Name: "IE5", EventID: &eventID},
		{ID: 16, Name: "専教", EventID: &eventID},
	}
}

func buildTypingSystemImportPayload(exportID string) typingSystemPayload {
	return typingSystemPayload{
		SchemaVersion: "typing-results-v1",
		ExportID:      exportID,
		Teams: []typingSystemTeam{
			{
				TeamName:    "1年生",
				Match1Score: 11,
				Match2Score: 11,
				Match3Score: 11,
				TotalScore:  33,
				Rank:        1,
			},
			{
				TeamName:    "2年生",
				Match1Score: 10,
				Match2Score: 10,
				Match3Score: 10,
				TotalScore:  30,
				Rank:        2,
			},
			{
				TeamName:    "3年生",
				Match1Score: 9,
				Match2Score: 9,
				Match3Score: 9,
				TotalScore:  27,
				Rank:        3,
			},
			{
				TeamName:    "4年生",
				Match1Score: 8,
				Match2Score: 8,
				Match3Score: 8,
				TotalScore:  24,
				Rank:        4,
			},
			{
				TeamName:    "5年生",
				Match1Score: 7,
				Match2Score: 7,
				Match3Score: 7,
				TotalScore:  21,
				Rank:        5,
			},
			{
				TeamName:    "専攻科・教員",
				Match1Score: 6,
				Match2Score: 6,
				Match3Score: 6,
				TotalScore:  18,
				Rank:        6,
			},
		},
	}
}

func buildTypingSystemImportRequest(t *testing.T, payload typingSystemPayload) *http.Request {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "typing-results.json")
	assert.NoError(t, err)

	content, err := json.Marshal(payload)
	assert.NoError(t, err)
	_, err = part.Write(content)
	assert.NoError(t, err)
	err = writer.Close()
	assert.NoError(t, err)

	req, _ := http.NewRequest(http.MethodPost, "/api/root/noon-game/sessions/10/typing-system/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req
}

func TestNoonGameHandler_ImportTypingSystemResults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := "00000000-0000-0000-0000-000000000001"
	eventID := 1
	sessionID := 10
	exportID := "4ef87d60-2e74-477a-9c16-a93423d04c20"

	t.Run("Success", func(t *testing.T) {
		payload := buildTypingSystemImportPayload(exportID)

		mockNoonRepo := new(MockNoonGameRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

		hash := sha256.Sum256(func() []byte {
			s, err := json.Marshal(payload)
			assert.NoError(t, err)
			return s
		}())
		expectedHash := hex.EncodeToString(hash[:])

		mockNoonRepo.On("GetSessionByID", sessionID).Return(&models.NoonGameSession{ID: sessionID, EventID: eventID}, nil).Once()
		mockNoonRepo.On("GetTypingSystemImportsBySessionAndExportID", sessionID, exportID).Return([]*models.NoonGameTypingSystemImportRecord{}, nil).Once()
		mockNoonRepo.On("GetActiveTypingSystemImport", sessionID).Return(nil, nil).Once()
		mockClassRepo.On("GetAllClasses", eventID).Return(buildTypingSystemImportClasses(eventID), nil).Once()
		mockNoonRepo.On("ApplyTypingSystemResultImport", sessionID, mock.AnythingOfType("[]*models.NoonGamePoint"), false, mock.MatchedBy(func(item *models.NoonGameTypingSystemImportRecord) bool {
			if item == nil {
				return false
			}
			if item.ExportID != exportID || item.SHA256 != expectedHash || item.SessionID != sessionID {
				return false
			}
			if item.Action != "import" || !item.IsActive {
				return false
			}
			if item.RequestedBy != userID {
				return false
			}
			if len(item.Results) != 6 || item.Results[0].TeamName != "1年生" || item.Results[0].Points != 40 || item.Results[1].Points != 30 || item.Results[2].Points != 25 || item.Results[3].Points != 0 {
				return false
			}
			return true
		})).Run(func(args mock.Arguments) {
			points := args.Get(1).([]*models.NoonGamePoint)
			assert.Equal(t, 16, len(points))
			for _, p := range points {
				assert.Equal(t, "typing_system", p.Source)
				assert.Equal(t, userID, p.CreatedBy)
			}
			assert.Equal(t, 40, points[0].Points)
			assert.Equal(t, 30, points[3].Points)
			assert.Equal(t, 25, points[6].Points)
		}).Return(nil).Once()
		mockNoonRepo.On("SumConfirmedPointsByEvent", eventID).Return(map[int]int{}, nil).Once()
		mockClassRepo.On("SetNoonGamePoints", eventID, map[int]int{}).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "session_id", Value: "10"}}
		c.Set("user", &models.User{ID: userID})
		c.Request = buildTypingSystemImportRequest(t, payload)

		h.ImportTypingSystemResults(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "imported", response["message"])
		assert.Equal(t, "import", response["action"])
		assert.Equal(t, float64(6), response["team_count"])
		assert.Equal(t, float64(16), response["class_count"])

		mockNoonRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})

	t.Run("Success - already imported same export_id and same content", func(t *testing.T) {
		payload := buildTypingSystemImportPayload(exportID)

		mockNoonRepo := new(MockNoonGameRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

		content, err := json.Marshal(payload)
		assert.NoError(t, err)
		hash := sha256.Sum256(content)
		existingHash := hex.EncodeToString(hash[:])

		mockNoonRepo.On("GetSessionByID", sessionID).Return(&models.NoonGameSession{ID: sessionID, EventID: eventID}, nil).Once()
		mockClassRepo.On("GetAllClasses", eventID).Return(buildTypingSystemImportClasses(eventID), nil).Once()
		mockNoonRepo.On("GetTypingSystemImportsBySessionAndExportID", sessionID, exportID).Return([]*models.NoonGameTypingSystemImportRecord{{
			SessionID: sessionID,
			ExportID:  exportID,
			SHA256:    existingHash,
			Status:    "success",
		}}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "session_id", Value: "10"}}
		c.Set("user", &models.User{ID: userID})
		c.Request = buildTypingSystemImportRequest(t, payload)

		h.ImportTypingSystemResults(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "already imported", response["message"])

		mockNoonRepo.AssertExpectations(t)
		mockNoonRepo.AssertNotCalled(t, "ApplyTypingSystemResultImport", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockNoonRepo.AssertNotCalled(t, "SumConfirmedPointsByEvent", mock.Anything)
		mockClassRepo.AssertExpectations(t)
	})

	t.Run("Conflict - same export_id but different content", func(t *testing.T) {
		payload := buildTypingSystemImportPayload(exportID)
		payloadCopy := buildTypingSystemImportPayload(exportID)
		payloadCopy.Teams[0].TotalScore = 999
		payloadCopy.Teams[0].Match1Score = 999

		mockNoonRepo := new(MockNoonGameRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

		content, err := json.Marshal(payload)
		assert.NoError(t, err)
		hash := sha256.Sum256(content)
		existingHash := hex.EncodeToString(hash[:])
		otherPayload, err := json.Marshal(payloadCopy)
		assert.NoError(t, err)
		otherHash := sha256.Sum256(otherPayload)
		otherHashHex := hex.EncodeToString(otherHash[:])
		assert.NotEqual(t, existingHash, otherHashHex)

		mockNoonRepo.On("GetSessionByID", sessionID).Return(&models.NoonGameSession{ID: sessionID, EventID: eventID}, nil).Once()
		mockClassRepo.On("GetAllClasses", eventID).Return(buildTypingSystemImportClasses(eventID), nil).Once()
		mockNoonRepo.On("GetTypingSystemImportsBySessionAndExportID", sessionID, payload.ExportID).Return([]*models.NoonGameTypingSystemImportRecord{{
			SessionID: sessionID,
			ExportID:  payload.ExportID,
			SHA256:    otherHashHex,
			Status:    "success",
		}}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "session_id", Value: "10"}}
		c.Set("user", &models.User{ID: userID})
		c.Request = buildTypingSystemImportRequest(t, payload)

		h.ImportTypingSystemResults(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		var response map[string]interface{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "Import with same export_id already exists but file content differs", response["error"])

		mockNoonRepo.AssertNotCalled(t, "ApplyTypingSystemResultImport", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockNoonRepo.AssertNotCalled(t, "SumConfirmedPointsByEvent", mock.Anything)
		mockNoonRepo.AssertExpectations(t)
	})

	t.Run("Conflict - active different export_id and replace not specified", func(t *testing.T) {
		payload := buildTypingSystemImportPayload(exportID)

		mockNoonRepo := new(MockNoonGameRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

		mockNoonRepo.On("GetSessionByID", sessionID).Return(&models.NoonGameSession{ID: sessionID, EventID: eventID}, nil).Once()
		mockClassRepo.On("GetAllClasses", eventID).Return(buildTypingSystemImportClasses(eventID), nil).Once()
		mockNoonRepo.On("GetTypingSystemImportsBySessionAndExportID", sessionID, exportID).Return([]*models.NoonGameTypingSystemImportRecord{}, nil).Once()
		mockNoonRepo.On("GetActiveTypingSystemImport", sessionID).Return(&models.NoonGameTypingSystemImportRecord{
			SessionID:   sessionID,
			ExportID:    "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
			SHA256:      "previous-hash",
			Status:      "success",
			IsActive:    true,
			Action:      "import",
			RequestedBy: "00000000-0000-0000-0000-000000000002",
		}, nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "session_id", Value: "10"}}
		c.Set("user", &models.User{ID: userID})
		c.Request = buildTypingSystemImportRequest(t, payload)

		h.ImportTypingSystemResults(c)

		assert.Equal(t, http.StatusConflict, w.Code)
		var response map[string]interface{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "Existing import already exists for this session. Use replace=true to overwrite", response["error"])

		mockNoonRepo.AssertNotCalled(t, "ApplyTypingSystemResultImport", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		mockNoonRepo.AssertExpectations(t)
	})

	t.Run("Success - replace existing import", func(t *testing.T) {
		payload := buildTypingSystemImportPayload(exportID)

		mockNoonRepo := new(MockNoonGameRepository)
		mockClassRepo := new(MockClassRepository)
		mockEventRepo := new(MockEventRepository)

		h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

		hash := sha256.Sum256(func() []byte {
			s, err := json.Marshal(payload)
			assert.NoError(t, err)
			return s
		}())
		expectedHash := hex.EncodeToString(hash[:])

		mockNoonRepo.On("GetSessionByID", sessionID).Return(&models.NoonGameSession{ID: sessionID, EventID: eventID}, nil).Once()
		mockNoonRepo.On("GetTypingSystemImportsBySessionAndExportID", sessionID, exportID).Return([]*models.NoonGameTypingSystemImportRecord{}, nil).Once()
		mockNoonRepo.On("GetActiveTypingSystemImport", sessionID).Return(&models.NoonGameTypingSystemImportRecord{
			SessionID:   sessionID,
			ExportID:    "00000000-0000-0000-0000-000000000099",
			SHA256:      "previous-hash",
			Status:      "success",
			IsActive:    true,
			Action:      "import",
			RequestedBy: "00000000-0000-0000-0000-000000000002",
		}, nil).Once()
		mockClassRepo.On("GetAllClasses", eventID).Return(buildTypingSystemImportClasses(eventID), nil).Once()
		mockNoonRepo.On("ApplyTypingSystemResultImport", sessionID, mock.AnythingOfType("[]*models.NoonGamePoint"), true, mock.MatchedBy(func(item *models.NoonGameTypingSystemImportRecord) bool {
			return item != nil && item.Action == "replace" && item.ReplacedExportID != nil && *item.ReplacedExportID == "00000000-0000-0000-0000-000000000099" && item.SHA256 == expectedHash
		})).Return(nil).Once()
		mockNoonRepo.On("SumConfirmedPointsByEvent", eventID).Return(map[int]int{}, nil).Once()
		mockClassRepo.On("SetNoonGamePoints", eventID, map[int]int{}).Return(nil).Once()

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "session_id", Value: "10"}}
		c.Set("user", &models.User{ID: userID})
		c.Request = buildTypingSystemImportRequest(t, payload)
		c.Request.URL.RawQuery = "replace=true"

		h.ImportTypingSystemResults(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &response))
		assert.Equal(t, "imported", response["message"])
		assert.Equal(t, "replace", response["action"])
		assert.Equal(t, float64(6), response["team_count"])

		mockNoonRepo.AssertExpectations(t)
		mockClassRepo.AssertExpectations(t)
	})
}
