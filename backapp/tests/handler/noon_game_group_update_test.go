package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNoonGameHandler_SaveGroup_AutoRenamesDerivedGroupAndSyncsMatchLabels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockNoonRepo := new(MockNoonGameRepository)
	mockClassRepo := new(MockClassRepository)
	mockEventRepo := new(MockEventRepository)

	h := handler.NewNoonGameHandler(mockNoonRepo, mockClassRepo, mockEventRepo)

	session := &models.NoonGameSession{ID: 1, EventID: 1, Name: "昼休み競技 2025"}
	existingGroup := &models.NoonGameGroupWithMembers{
		NoonGameGroup: &models.NoonGameGroup{
			ID:        10,
			SessionID: 1,
			Name:      "1-1 & IEコース",
		},
		Members: []*models.NoonGameGroupMember{
			{ID: 1, GroupID: 10, ClassID: 101, Class: &models.Class{ID: 101, Name: "1-1"}},
			{ID: 2, GroupID: 10, ClassID: 201, Class: &models.Class{ID: 201, Name: "IE2"}},
		},
	}
	updatedGroup := &models.NoonGameGroupWithMembers{
		NoonGameGroup: &models.NoonGameGroup{
			ID:        10,
			SessionID: 1,
			Name:      "1-2 & IEコース",
		},
		Members: []*models.NoonGameGroupMember{
			{ID: 3, GroupID: 10, ClassID: 102, Class: &models.Class{ID: 102, Name: "1-2"}},
			{ID: 4, GroupID: 10, ClassID: 201, Class: &models.Class{ID: 201, Name: "IE2"}},
		},
	}

	oldDisplayName := "1-1 & IEコース"
	matchEntries := []*models.NoonGameMatchEntry{
		{
			ID:          1,
			MatchID:     99,
			EntryIndex:  0,
			SideType:    "group",
			GroupID:     intPtrNoonGroupUpdate(10),
			DisplayName: &oldDisplayName,
		},
	}
	match := &models.NoonGameMatchWithResult{
		NoonGameMatch: &models.NoonGameMatch{
			ID:        99,
			SessionID: 1,
			Entries:   matchEntries,
		},
		Entries: matchEntries,
	}

	mockNoonRepo.On("GetSessionByID", 1).Return(session, nil).Once()
	mockNoonRepo.On("GetGroupWithMembers", 1, 10).Return(existingGroup, nil).Once()
	mockClassRepo.On("GetClassByID", 102).Return(&models.Class{ID: 102, Name: "1-2"}, nil).Once()
	mockClassRepo.On("GetClassByID", 201).Return(&models.Class{ID: 201, Name: "IE2"}, nil).Once()
	mockNoonRepo.
		On("SaveGroup", mock.MatchedBy(func(group *models.NoonGameGroup) bool {
			return group != nil &&
				group.ID == 10 &&
				group.SessionID == 1 &&
				group.Name == "1-2 & IEコース"
		}), []int{102, 201}).
		Return(updatedGroup, nil).
		Once()
	mockNoonRepo.On("GetMatchesWithResults", 1).Return([]*models.NoonGameMatchWithResult{match}, nil).Once()
	mockNoonRepo.
		On("SaveMatch", mock.MatchedBy(func(saved *models.NoonGameMatch) bool {
			if saved == nil || len(saved.Entries) != 1 || saved.Entries[0].DisplayName == nil {
				return false
			}
			return *saved.Entries[0].DisplayName == "1-2 & IEコース"
		})).
		Return(match.NoonGameMatch, nil).
		Once()

	body, err := json.Marshal(map[string]interface{}{
		"name":        "1-1 & IEコース",
		"description": "",
		"class_ids":   []int{102, 201},
	})
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPut, "/noon-game/sessions/1/groups/10", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{
		{Key: "session_id", Value: "1"},
		{Key: "group_id", Value: "10"},
	}
	c.Request = req

	h.SaveGroup(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "1-2 \\u0026 IEコース")

	mockNoonRepo.AssertExpectations(t)
	mockClassRepo.AssertExpectations(t)
}

func intPtrNoonGroupUpdate(v int) *int {
	return &v
}
