package handler_test

import (
	"backapp/internal/handler"
	"backapp/internal/models"
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateStudentCountsRejectsNegativeCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eventRepo := new(MockEventRepository)
	classRepo := new(MockClassRepository)
	h := handler.NewClassHandler(classRepo, eventRepo, nil, nil)
	eventRepo.On("GetActiveEvent").Return(7, nil).Once()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/api/root/classes/student-counts", bytes.NewBufferString(`[{"class_id":1,"student_count":-1}]`))
	c.Request.Header.Set("Content-Type", "application/json")
	h.UpdateStudentCountsHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	classRepo.AssertNotCalled(t, "UpdateStudentCounts", mock.Anything, mock.Anything)
	eventRepo.AssertExpectations(t)
}

func TestUpdateStudentCountsCSVRejectsUnknownClassWithoutPartialUpdate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	eventRepo := new(MockEventRepository)
	classRepo := new(MockClassRepository)
	h := handler.NewClassHandler(classRepo, eventRepo, nil, nil)
	eventRepo.On("GetActiveEvent").Return(7, nil).Once()
	classRepo.On("GetAllClasses", 7).Return([]*models.Class{{ID: 1, Name: "1-A"}}, nil).Once()

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("csv", "counts.csv")
	assert.NoError(t, err)
	_, err = part.Write([]byte("class_name,student_count\n1-A,30\n存在しないクラス,20\n"))
	assert.NoError(t, err)
	assert.NoError(t, writer.Close())

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/root/classes/student-counts/csv", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())
	h.UpdateStudentCountsFromCSVHandler(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "unknown class")
	classRepo.AssertNotCalled(t, "UpdateStudentCounts", mock.Anything, mock.Anything)
	eventRepo.AssertExpectations(t)
	classRepo.AssertExpectations(t)
}
