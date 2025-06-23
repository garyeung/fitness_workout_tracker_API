package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"workout-tracker-api/internal/handler"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/pkg/api"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockExerciseService struct {
	mock.Mock
}

func (m *MockExerciseService) GetExerciseById(ctx context.Context, id int) (*service.Exercise, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Exercise), args.Error(1)
}
func (m *MockExerciseService) ListExercises(ctx context.Context) ([]service.Exercise, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.Exercise), args.Error(1)
}

func TestExerciseHandler(t *testing.T) {
	nonexisExercise := 999
	exercise1 := service.Exercise{
		Id:          1,
		Name:        "Best push up",
		Description: "very easy to do",
		MuscleGroup: service.Chest,
	}

	exercise2 := service.Exercise{
		Id:          2,
		Name:        "Good Leg day",
		Description: "not hard, trust me",
		MuscleGroup: service.Legs,
	}
	mockExercises := []service.Exercise{
		exercise1, exercise2,
	}

	t.Run("ListExercises", func(t *testing.T) {
		t.Run("successful non empty exercises 200", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("ListExercises", mock.Anything).Return(mockExercises, nil).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
			rr := httptest.NewRecorder()

			h.ListExercises(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			assert.Equal(t, "successfully fetch exercises", resp.Message)
			exs, ok := (*resp.Payload)["exercises"].([]any)
			assert.True(t, ok)
			assert.Len(t, exs, 2)
			mockService.AssertExpectations(t)
		})

		t.Run("successful empty exercises 200", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("ListExercises", mock.Anything).Return([]service.Exercise{}, nil).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
			rr := httptest.NewRecorder()

			h.ListExercises(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			exs, _ := (*resp.Payload)["exercises"].([]any)
			assert.Len(t, exs, 0)
			mockService.AssertExpectations(t)
		})

		t.Run("fail internal err 500", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("ListExercises", mock.Anything).Return(nil, errors.New("db error")).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises", nil)
			rr := httptest.NewRecorder()

			h.ListExercises(rr, req)

			assert.Equal(t, http.StatusInternalServerError, rr.Code)
			mockService.AssertExpectations(t)
		})
	})

	t.Run("GetExerciseByID", func(t *testing.T) {
		t.Run("successfully get exercise 200", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("GetExerciseById", mock.Anything, exercise1.Id).Return(&exercise1, nil).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises/1", nil)
			req.SetPathValue("exerciseId", strconv.Itoa(exercise1.Id))
			rr := httptest.NewRecorder()

			h.GetExerciseByID(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			var resp api.Success
			err := json.NewDecoder(rr.Body).Decode(&resp)
			assert.NoError(t, err)
			assert.Equal(t, api.FETCH, resp.Code)
			ex, ok := (*resp.Payload)["exercise"].(map[string]any)
			assert.True(t, ok)
			assert.Equal(t, float64(exercise1.Id), ex["id"])
			mockService.AssertExpectations(t)
		})

		t.Run("fail to get non existent exercise 404", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("GetExerciseById", mock.Anything, nonexisExercise).Return(nil, errors.New("not found")).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises/999", nil)
			req.SetPathValue("exerciseId", strconv.Itoa(nonexisExercise))
			rr := httptest.NewRecorder()

			h.GetExerciseByID(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code)
			mockService.AssertExpectations(t)
		})

		t.Run("fail to get invalid exercise id 400", func(t *testing.T) {
			mockService := new(MockExerciseService)
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises/abc", nil)
			req.SetPathValue("exerciseId", "abc")
			rr := httptest.NewRecorder()

			h.GetExerciseByID(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertExpectations(t)
		})

		t.Run("fail missing exercise id 400", func(t *testing.T) {
			mockService := new(MockExerciseService)
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises/", nil)
			rr := httptest.NewRecorder()

			h.GetExerciseByID(rr, req)

			assert.Equal(t, http.StatusBadRequest, rr.Code)
			mockService.AssertExpectations(t)
		})

		t.Run("database err 404", func(t *testing.T) {
			mockService := new(MockExerciseService)
			mockService.On("GetExerciseById", mock.Anything, exercise1.Id).Return(nil, errors.New("db error")).Once()
			h := handler.NewExerciseHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/exercises/1", nil)
			req.SetPathValue("exerciseId", strconv.Itoa(exercise1.Id))
			rr := httptest.NewRecorder()

			h.GetExerciseByID(rr, req)

			assert.Equal(t, http.StatusNotFound, rr.Code)
			mockService.AssertExpectations(t)
		})
	})
}
