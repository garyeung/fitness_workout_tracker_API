// workout_handler_test.go
package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/handler"     // Your handler package
	"workout-tracker-api/internal/service"     // Your service package
	"workout-tracker-api/internal/util/helper" // Assuming this helper exists for context
	"workout-tracker-api/pkg/api"
)

// MockWorkoutService is a mock implementation of service.WorkoutServiceInterface
type MockWorkoutService struct {
	mock.Mock
}

// Implement all methods of service.WorkoutServiceInterface
func (m *MockWorkoutService) CreateWorkout(ctx context.Context, input service.WorkoutPlanCreate) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockWorkoutService) GetWorkoutById(ctx context.Context, id int) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockWorkoutService) ListWorkouts(ctx context.Context, userID int) ([]service.WorkoutPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.WorkoutPlan), args.Error(1)
}

func (m *MockWorkoutService) ListWorkoutsByStatus(ctx context.Context, userId int, status service.WPStatus, asc bool) ([]service.WorkoutPlan, error) {
	args := m.Called(ctx, userId, status, asc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.WorkoutPlan), args.Error(1)
}

func (m *MockWorkoutService) UpdateExercisePlans(ctx context.Context, workoutID int, epsUpdate []service.ExercisePlanUpdate) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, workoutID, epsUpdate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockWorkoutService) DeleteWorkoutById(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWorkoutService) CompleteWorkout(ctx context.Context, id int, comment *string) error {
	args := m.Called(ctx, id, comment)
	return args.Error(0)
}

func (m *MockWorkoutService) ScheduleWorkout(ctx context.Context, id int, scheduledDate *time.Time) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, id, scheduledDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func TestWorkoutHandler(t *testing.T) {
	testUserID := 123
	testUserEmail := "test@example.com"
	testUserName := "Test User"

	// Helper to create a request with user info in context
	createRequestWithUser := func(method, url string, body []byte) *http.Request {
		req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		userInfo := helper.UserInfo{
			Email: testUserEmail,
			Name:  testUserName,
			Id:    testUserID,
		}
		ctx := context.WithValue(req.Context(), helper.UserContextKey, &userInfo)
		return req.WithContext(ctx)
	}

	// Helper to create a request with user info and workoutId in path
	createRequestWithUserAndWorkoutID := func(method, url string, workoutID int, body []byte) *http.Request {
		req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.SetPathValue("workoutId", strconv.Itoa(workoutID))

		userInfo := helper.UserInfo{
			Email: testUserEmail,
			Name:  testUserName,
			Id:    testUserID,
		}
		ctx := context.WithValue(req.Context(), helper.UserContextKey, &userInfo)
		return req.WithContext(ctx)
	}

	t.Run("ListWorkoutPlans", func(t *testing.T) {

		t.Run("Successful listing without filters", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			expectedWorkouts := []service.WorkoutPlan{
				{Id: 1, UserId: testUserID, Status: service.PENDING, ScheduledDate: time.Now().Add(24 * time.Hour)},
				{Id: 2, UserId: testUserID, Status: service.COMPLETED, ScheduledDate: time.Now().Add(48 * time.Hour)},
			}
			mockWorkoutService.On("ListWorkouts", mock.Anything, testUserID).Return(expectedWorkouts, nil).Once()

			req := createRequestWithUser(http.MethodGet, "/workouts", nil)
			rr := httptest.NewRecorder()

			workoutHandler.ListWorkoutPlans(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			assert.Equal(t, "successfully fetch workout plans", resp.Message)
			workoutPlansRaw, ok := (*resp.Payload)["workoutPlans"].([]any)
			assert.True(t, ok)
			assert.Len(t, workoutPlansRaw, 2)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Successful listing with status filter (completed, asc)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			expectedWorkouts := []service.WorkoutPlan{
				{Id: 2, UserId: testUserID, Status: service.COMPLETED, ScheduledDate: time.Now().Add(48 * time.Hour)},
				{Id: 3, UserId: testUserID, Status: service.COMPLETED, ScheduledDate: time.Now().Add(24 * time.Hour)},
			}
			mockWorkoutService.On("ListWorkoutsByStatus", mock.Anything, testUserID, service.COMPLETED, true).Return(expectedWorkouts, nil).Once()

			req := createRequestWithUser(http.MethodGet, "/workouts?status=completed&sort=asc", nil)
			rr := httptest.NewRecorder()

			workoutHandler.ListWorkoutPlans(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			assert.Equal(t, "successfully fetch workout plans", resp.Message)
			workoutPlansRaw, ok := (*resp.Payload)["workoutPlans"].([]any)
			assert.True(t, ok)
			assert.Len(t, workoutPlansRaw, 2)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Unauthorized access (no user in context)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			req := httptest.NewRequest(http.MethodGet, "/workouts", nil)
			rr := httptest.NewRecorder()

			workoutHandler.ListWorkoutPlans(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "ListWorkouts")
			mockWorkoutService.AssertNotCalled(t, "ListWorkoutsByStatus")
		})

		t.Run("Invalid status query parameter", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			req := createRequestWithUser(http.MethodGet, "/workouts?status=invalid_status", nil)
			rr := httptest.NewRecorder()

			workoutHandler.ListWorkoutPlans(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "ListWorkouts")
			mockWorkoutService.AssertNotCalled(t, "ListWorkoutsByStatus")
		})

		t.Run("Service returns an error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			serviceErr := errors.New("database connection failed")
			mockWorkoutService.On("ListWorkouts", mock.Anything, testUserID).Return(nil, serviceErr).Once()

			req := createRequestWithUser(http.MethodGet, "/workouts", nil)
			rr := httptest.NewRecorder()

			workoutHandler.ListWorkoutPlans(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})

	t.Run("CreateWorkoutPlan", func(t *testing.T) {

		mockScheduledDate := time.Date(2025, time.December, 25, 10, 30, 0, 0, time.Local)
		var mockExerciseId int64 = 1
		mockRepetion := 2
		mockSet := 3
		mockWeightUnit := api.Kg
		var mockWeight float32 = 60.2
		mockCreateExercisePlans := []api.CreateExercisePlan{
			{
				ExerciseId:  &mockExerciseId,
				Repetitions: &mockRepetion,
				Sets:        &mockSet,
				WeightUnit:  &mockWeightUnit,
				Weights:     &mockWeight,
			},
		}
		expectedExercisePlansCreate := []service.ExercisePlanCreate{
			{
				ExerciseId:  int(mockExerciseId),
				Repetitions: mockRepetion,
				Sets:        mockSet,
				WeightUnit:  service.WeightUnit(mockWeightUnit),
				Weights:     mockWeight,
			},
		}
		t.Run("Successful creation", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			reqBody := api.CreateWorkoutPlan{
				ScheduledDate: &mockScheduledDate,
				ExercisePlans: &mockCreateExercisePlans,
			}
			expectedServiceData := service.WorkoutPlanCreate{
				UserId:        testUserID,
				ScheduledDate: &mockScheduledDate,
				ExercisePlans: expectedExercisePlansCreate,
			}
			createdWorkout := &service.WorkoutPlan{
				Id:            101,
				UserId:        testUserID,
				Status:        service.PENDING,
				ScheduledDate: mockScheduledDate,
			}

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUser(http.MethodPost, "/workouts", body)
			rr := httptest.NewRecorder()

			mockWorkoutService.On("CreateWorkout", req.Context(), expectedServiceData).Return(createdWorkout, nil).Once()

			workoutHandler.CreateWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusCreated, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.CREATED, resp.Code)
			assert.Equal(t, "successfully create workout plan", resp.Message)
			createdWpRaw, ok := (*resp.Payload)["workoutPlan"].(map[string]interface{})
			assert.True(t, ok)
			assert.Equal(t, float64(101), createdWpRaw["id"].(float64)) // JSON unmarshals numbers to float64
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Invalid JSON body", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			req := createRequestWithUser(http.MethodPost, "/workouts", []byte(`{"scheduledDate": "invalid"}`))
			rr := httptest.NewRecorder()

			workoutHandler.CreateWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "CreateWorkout")
		})

		t.Run("Unauthorized access (no user in context)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			reqBody := api.CreateWorkoutPlanJSONRequestBody{
				ScheduledDate: &mockScheduledDate,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/workouts", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			workoutHandler.CreateWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "CreateWorkout")
		})

		t.Run("Service returns error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			reqBody := api.CreateWorkoutPlanJSONRequestBody{
				ScheduledDate: &mockScheduledDate,
				ExercisePlans: &mockCreateExercisePlans,
			}
			serviceErr := errors.New("failed to save to database")

			mockWorkoutService.On("CreateWorkout", mock.Anything, mock.Anything).Return(nil, serviceErr).Once()

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUser(http.MethodPost, "/workouts", body)
			rr := httptest.NewRecorder()

			workoutHandler.CreateWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})

	t.Run("GetWorkoutPlanbyID", func(t *testing.T) {
		workoutID := 123

		t.Run("Successful retrieval", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			expectedWorkout := &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        testUserID,
				Status:        service.PENDING,
				ScheduledDate: time.Now().Add(24 * time.Hour),
			}

			//double Auth
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(expectedWorkout, nil).Twice()

			req := createRequestWithUserAndWorkoutID(http.MethodGet, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			assert.Equal(t, "successfully fetch workout plan", resp.Message)
			fetchedWpRaw, ok := (*resp.Payload)["workoutPlan"].(map[string]any)
			assert.True(t, ok)
			assert.Equal(t, float64(workoutID), fetchedWpRaw["id"].(float64))
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Invalid workout ID in path", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			req := createRequestWithUser(http.MethodGet, "/workouts/abc", nil)
			rr := httptest.NewRecorder()
			// Manually set path value, as createRequestWithUser doesn't handle it for non-numeric cases
			req.SetPathValue("workoutId", "abc")

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_ID), resp.Code)

			mockWorkoutService.AssertNumberOfCalls(t, "GetWorkoutById", 0)
		})

		t.Run("Unauthorized access (no user in context)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/workouts/%d", workoutID), nil)
			// Manually set path value, no user info
			req.SetPathValue("workoutId", strconv.Itoa(workoutID))
			rr := httptest.NewRecorder()

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
			mockWorkoutService.AssertNumberOfCalls(t, "GetWorkoutById", 0)
		})

		t.Run("Workout not found", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(nil, apperrors.ErrNotFound).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodGet, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.NOT_FOUND), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("User not authorized for this workout (e.g., different user ID)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)
			// Simulate a workout that belongs to a different user
			anotherUserID := 456
			workoutBelongingToAnotherUser := &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        anotherUserID, // Different user
				Status:        service.PENDING,
				ScheduledDate: time.Now().Add(24 * time.Hour),
			}
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(workoutBelongingToAnotherUser, nil).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodGet, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusForbidden, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.FORBIDDEN), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Service returns generic error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			serviceErr := errors.New("db error")
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(nil, serviceErr).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodGet, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.GetWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.NOT_FOUND), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})

	t.Run("DeleteWorkoutPlanbyID", func(t *testing.T) {
		workoutID := 123

		// Setup for successful doubleAuth check
		existingWorkout := &service.WorkoutPlan{
			Id:            workoutID,
			UserId:        testUserID,
			Status:        service.PENDING,
			ScheduledDate: time.Now(),
		}

		t.Run("Successful deletion", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("DeleteWorkoutById", mock.Anything, workoutID).Return(nil).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodDelete, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.DeleteWoroutPlanById(rr, req)

			assert.Equal(t, http.StatusNoContent, rr.Code)
			assert.Empty(t, rr.Body.String()) // No content for 204
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Deletion service returns error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			serviceErr := errors.New("db delete error")
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("DeleteWorkoutById", mock.Anything, workoutID).Return(serviceErr).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodDelete, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.DeleteWoroutPlanById(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Workout not found for deletion", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(nil, apperrors.ErrNotFound).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodDelete, fmt.Sprintf("/workouts/%d", workoutID), workoutID, nil)
			rr := httptest.NewRecorder()

			workoutHandler.DeleteWoroutPlanById(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})

	t.Run("CompleteWorkoutPlanbyID", func(t *testing.T) {
		workoutID := 123
		now := time.Now()

		// Setup for successful doubleAuth check
		existingWorkout := &service.WorkoutPlan{
			Id:            workoutID,
			UserId:        testUserID,
			Status:        service.PENDING,
			ScheduledDate: now,
		}
		comment := "Great workout!"
		reqBody := api.CompleteWorkoutPlan{Comment: &comment}

		t.Run("Successful completion", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("CompleteWorkout", mock.Anything, existingWorkout.Id, &comment).Return(nil).Once()

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/complete", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.CompleteWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusNoContent, rr.Code)
			assert.Empty(t, rr.Body)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Invalid JSON body", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/complete", workoutID), workoutID, []byte(`{"comment":123}`))
			rr := httptest.NewRecorder()

			workoutHandler.CompleteWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "CompleteWorkout")
		})

		t.Run("Update service returns error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			serviceErr := errors.New("failed to update status")

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("CompleteWorkout", mock.Anything, workoutID, &comment).Return(serviceErr).Once()

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/complete", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.CompleteWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})

	t.Run("ScheduleWorkoutPlanbyId", func(t *testing.T) {
		workoutID := 123
		now := time.Now()
		newScheduledDate := time.Date(2024, 12, 13, 4, 4, 0, 0, time.Local)

		// Setup for successful doubleAuth check
		existingWorkout := &service.WorkoutPlan{
			Id:            workoutID,
			UserId:        testUserID,
			Status:        service.PENDING,
			ScheduledDate: now,
		}
		reqBody := api.ScheduleWorkoutPlan{ScheduledDate: &newScheduledDate}

		t.Run("Successful scheduling", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			updatedWorkout := &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        testUserID,
				Status:        service.PENDING, // Status should not change
				ScheduledDate: newScheduledDate,
			}

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("ScheduleWorkout", mock.Anything, workoutID, &newScheduledDate).Return(updatedWorkout, nil).Once()

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/schedule", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.ScheduleWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.UPDATE, resp.Code)
			assert.Equal(t, "successfully schedule workout plan", resp.Message)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Invalid JSON body", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/schedule", workoutID), workoutID, []byte(`{"scheduledDate": "not-a-date"}`))
			rr := httptest.NewRecorder()

			workoutHandler.ScheduleWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "ScheduleWorkout")
		})

		t.Run("Update service returns error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			serviceErr := errors.New("failed to update scheduled date")

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("ScheduleWorkout", mock.Anything, workoutID, &newScheduledDate).Return(nil, serviceErr).Once()

			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/schedule", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.ScheduleWorkoutPlanById(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})
	// Add inside TestWorkoutHandler(t *testing.T)
	t.Run("UpdateExercisePlansInWorkoutPlan", func(t *testing.T) {
		workoutID := 123
		var mockEPId int64 = 33
		mockSet := 3
		mockRepetition := 10
		var mockWeights float32 = 50.2
		mockWeightUnit := api.Kg
		now := time.Now()

		// Setup for successful doubleAuth check
		existingWorkout := &service.WorkoutPlan{
			Id:            workoutID,
			UserId:        testUserID,
			Status:        service.PENDING,
			ScheduledDate: now,
		}
		mockUpdateEP := api.UpdateExercisePlan{
			Id:          &mockEPId,
			Sets:        &mockSet,
			Repetitions: &mockRepetition,
			Weights:     &mockWeights,
			WeightUnit:  &mockWeightUnit,
		}
		mockUpdateEPs := []api.UpdateExercisePlan{mockUpdateEP}
		serviceUpdateEPs := []service.ExercisePlanUpdate{
			{
				Id:          int(mockEPId),
				Sets:        mockSet,
				Repetitions: mockRepetition,
				Weights:     mockWeights,
				WeightUnit:  service.WeightUnit(mockWeightUnit),
			},
		}
		updatedWorkout := &service.WorkoutPlan{
			Id:            workoutID,
			UserId:        testUserID,
			Status:        service.PENDING,
			ScheduledDate: now,
			ExercisePlans: []service.ExercisePlan{},
		}

		t.Run("Successful update", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			mockWorkoutService.On("UpdateExercisePlans", mock.Anything, workoutID, serviceUpdateEPs).Return(updatedWorkout, nil).Once()

			reqBody := api.UpdateExercisePlansInWorkoutPlanJSONBody{
				ExercisePlans: &mockUpdateEPs,
			}
			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/exercise-plans", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.UpdateExercisePlansInWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.UPDATE, resp.Code)
			assert.Equal(t, "successfully update exercise plans", resp.Message)
			mockWorkoutService.AssertExpectations(t)
		})

		t.Run("Invalid JSON body", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()

			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/exercise-plans", workoutID), workoutID, []byte(`{"exercisePlans": "not-an-array"}`))
			rr := httptest.NewRecorder()

			workoutHandler.UpdateExercisePlansInWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "UpdateExercisePlans")
		})

		t.Run("Unauthorized access (no user in context)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			reqBody := api.UpdateExercisePlansInWorkoutPlanJSONBody{
				ExercisePlans: &mockUpdateEPs,
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/workouts/%d/exercise-plans", workoutID), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.SetPathValue("workoutId", strconv.Itoa(workoutID))
			rr := httptest.NewRecorder()

			workoutHandler.UpdateExercisePlansInWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusUnauthorized, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "UpdateExercisePlans")
		})

		t.Run("Forbidden (user does not own workout)", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			otherUserWorkout := &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        9999, // not testUserID
				Status:        service.PENDING,
				ScheduledDate: now,
			}
			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(otherUserWorkout, nil).Once()

			reqBody := api.UpdateExercisePlansInWorkoutPlanJSONBody{
				ExercisePlans: &mockUpdateEPs,
			}
			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/exercise-plans", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.UpdateExercisePlansInWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusForbidden, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.FORBIDDEN), resp.Code)
			mockWorkoutService.AssertNotCalled(t, "UpdateExercisePlans")
		})

		t.Run("Service returns error", func(t *testing.T) {
			mockWorkoutService := new(MockWorkoutService)
			workoutHandler := handler.NewWorkoutHandler(mockWorkoutService)

			mockWorkoutService.On("GetWorkoutById", mock.Anything, workoutID).Return(existingWorkout, nil).Once()
			serviceErr := errors.New("db error")
			mockWorkoutService.On("UpdateExercisePlans", mock.Anything, workoutID, serviceUpdateEPs).Return(nil, serviceErr).Once()

			reqBody := api.UpdateExercisePlansInWorkoutPlanJSONBody{
				ExercisePlans: &mockUpdateEPs,
			}
			body, _ := json.Marshal(reqBody)
			req := createRequestWithUserAndWorkoutID(http.MethodPut, fmt.Sprintf("/workouts/%d/exercise-plans", workoutID), workoutID, body)
			rr := httptest.NewRecorder()

			workoutHandler.UpdateExercisePlansInWorkoutPlan(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			var resp api.Error
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, string(apperrors.INTERNAL_ERROR), resp.Code)
			mockWorkoutService.AssertExpectations(t)
		})
	})
}
