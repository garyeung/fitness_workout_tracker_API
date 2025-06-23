package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/service"
)

// MockExerciseRepository is a mock implementation of repository.ExerciseRepository
type MockExerciseRepository struct {
	mock.Mock
}

func (m *MockExerciseRepository) CreateExercise(ctx context.Context, data repository.CreateExer) (*repository.Exercise, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) DeleteExercise(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockExerciseRepository) GetExerciseById(ctx context.Context, id int) (*repository.Exercise, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Exercise), args.Error(1)
}

func (m *MockExerciseRepository) ListExercises(ctx context.Context) ([]repository.Exercise, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.Exercise), args.Error(1)
}

func TestListExercises(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockExerciseRepository)
	exerciseService := service.NewExerciseService(mockRepo)

	tests := []struct {
		name              string
		mockRepoSetup     func()
		expectedExercises []service.Exercise
		expectedError     error
	}{
		{
			name: "Successful list of exercises",
			mockRepoSetup: func() {
				mockRepo.On("ListExercises", ctx).Return([]repository.Exercise{
					{Id: 1, Name: "Push-ups", Description: "Bodyweight exercise", MuscleGroup: "chest"},
					{Id: 2, Name: "Squats", Description: "Lower body exercise", MuscleGroup: "legs"},
				}, nil).Once()
			},
			expectedExercises: []service.Exercise{
				{Id: 1, Name: "Push-ups", Description: "Bodyweight exercise", MuscleGroup: "chest"},
				{Id: 2, Name: "Squats", Description: "Lower body exercise", MuscleGroup: "legs"},
			},
			expectedError: nil,
		},
		{
			name: "No exercises found",
			mockRepoSetup: func() {
				mockRepo.On("ListExercises", ctx).Return([]repository.Exercise{}, nil).Once()
			},
			expectedExercises: []service.Exercise(nil),
			expectedError:     nil,
		},
		{
			name: "Repository error",
			mockRepoSetup: func() {
				mockRepo.On("ListExercises", ctx).Return(nil, errors.New("database error")).Once()
			},
			expectedExercises: nil,
			expectedError:     errors.New("failed to list exercises: database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup()
			exercises, err := exerciseService.ListExercises(ctx)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, exercises)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExercises, exercises)
			}
			mockRepo.AssertExpectations(t) // Verify that expectations were met
		})
	}
}

func TestGetExerciseById(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockExerciseRepository)
	exerciseService := service.NewExerciseService(mockRepo)

	tests := []struct {
		name             string
		exerciseID       int
		mockRepoSetup    func(id int)
		expectedExercise *service.Exercise
		expectedError    error
	}{
		{
			name:       "Successful retrieval by ID",
			exerciseID: 1,
			mockRepoSetup: func(id int) {
				mockRepo.On("GetExerciseById", ctx, id).Return(&repository.Exercise{
					Id: 1, Name: "Push-ups", Description: "Bodyweight exercise", MuscleGroup: "chest",
				}, nil).Once()
			},
			expectedExercise: &service.Exercise{
				Id: 1, Name: "Push-ups", Description: "Bodyweight exercise", MuscleGroup: "chest",
			},
			expectedError: nil,
		},
		{
			name:       "Exercise not found",
			exerciseID: 99,
			mockRepoSetup: func(id int) {
				mockRepo.On("GetExerciseById", ctx, id).Return(nil, apperrors.ErrNotFound).Once()
			},
			expectedExercise: nil,
			expectedError:    errors.New("failed to get exercise by id '99': resource not found"),
		},
		{
			name:       "Repository error",
			exerciseID: 1,
			mockRepoSetup: func(id int) {
				mockRepo.On("GetExerciseById", ctx, id).Return(nil, errors.New("database error")).Once()
			},
			expectedExercise: nil,
			expectedError:    errors.New("failed to get exercise by id '1': database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockRepoSetup(tt.exerciseID)
			exercise, err := exerciseService.GetExerciseById(ctx, tt.exerciseID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, exercise)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedExercise, exercise)
			}
			mockRepo.AssertExpectations(t)
		})
	}
}
