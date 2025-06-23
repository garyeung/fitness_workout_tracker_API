package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"workout-tracker-api/internal/handler"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockReportService implements service.ReportServiceInterface
type MockReportService struct {
	mock.Mock
}

func (m *MockReportService) Progress(ctx context.Context, userID int) (*service.ProgressStatus, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.ProgressStatus), args.Error(1)
}

func TestReportHandler_ReportProgress(t *testing.T) {
	const testUserID = 42

	t.Run("successfully fetch progress", func(t *testing.T) {
		mockService := new(MockReportService)
		handlerObj := handler.NewReportHandler(mockService)

		progress := &service.ProgressStatus{
			CompleteWorkouts: 5,
			TotalWorkouts:    10,
		}
		mockService.On("Progress", mock.Anything, testUserID).Return(progress, nil).Once()

		// Add user info to context
		req := httptest.NewRequest(http.MethodGet, "/report/progress", nil)
		ctx := helper.SetUserInfoToContext(req.Context(), &helper.UserInfo{Id: testUserID})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handlerObj.ReportProgress(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp api.Success
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, api.FETCH, resp.Code)
		assert.Equal(t, "successsfully fetch progress report", resp.Message)
		progressPayload, ok := (*resp.Payload)["progress"].(map[string]any)
		assert.True(t, ok)
		assert.Equal(t, float64(5), progressPayload["completedWorkouts"])
		assert.Equal(t, float64(10), progressPayload["totalWorkouts"])
		mockService.AssertExpectations(t)
	})

	t.Run("unauthorized if no user in context", func(t *testing.T) {
		mockService := new(MockReportService)
		handlerObj := handler.NewReportHandler(mockService)

		req := httptest.NewRequest(http.MethodGet, "/report/progress", nil)
		rr := httptest.NewRecorder()

		handlerObj.ReportProgress(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockService.AssertNotCalled(t, "Progress")
	})

	t.Run("service error returns 500", func(t *testing.T) {
		mockService := new(MockReportService)
		handlerObj := handler.NewReportHandler(mockService)

		mockService.On("Progress", mock.Anything, testUserID).Return(nil, errors.New("db error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/report/progress", nil)
		ctx := helper.SetUserInfoToContext(req.Context(), &helper.UserInfo{Id: testUserID})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handlerObj.ReportProgress(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		mockService.AssertExpectations(t)
	})
}
