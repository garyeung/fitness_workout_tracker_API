package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/service"
)

// (Copy MockWorkoutRepository here if report_service_test.go is in a different package,
// otherwise, it should already be available if both are in service_test package)

// MockWorkoutForReportRepository is a mock implementation of repository.WorkoutRepository
// (This definition should be present in service_test package, ideally once)
type MockWorkoutForReportRepository struct {
	mock.Mock
}

func (m *MockWorkoutForReportRepository) CreateWorkout(ctx context.Context, data repository.CreateWP) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutForReportRepository) GetWorkoutById(ctx context.Context, id int) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutForReportRepository) UpdateWorkout(ctx context.Context, data repository.UpdateWP) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutForReportRepository) DeleteWorkoutById(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockWorkoutForReportRepository) ListWorkoutsByStatus(ctx context.Context, userID int, status repository.WPStatus, asc bool) ([]repository.WorkoutPlan, error) {
	args := m.Called(ctx, userID, status, asc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutForReportRepository) ListUserWorkouts(ctx context.Context, userID int) ([]repository.WorkoutPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WorkoutPlan), args.Error(1)
}

// --- Tests ---

func TestReportService_Progress(t *testing.T) {
	ctx := context.Background()
	userID := 123

	tests := []struct {
		name              string
		userID            int
		mockRepoSetup     func(*MockWorkoutForReportRepository)
		expectedProgress  *service.ProgressStatus
		expectedErrorType error
	}{
		{
			name:   "Successful progress report with completed and total workouts",
			userID: userID,
			mockRepoSetup: func(mwr *MockWorkoutForReportRepository) {
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.COMPLETED, true).Return([]repository.WorkoutPlan{
					{Id: 1, Status: repository.COMPLETED, UserId: userID},
					{Id: 2, Status: repository.COMPLETED, UserId: userID},
				}, nil).Once()
				mwr.On("ListUserWorkouts", ctx, userID).Return([]repository.WorkoutPlan{
					{Id: 1, Status: repository.COMPLETED, UserId: userID},
					{Id: 2, Status: repository.COMPLETED, UserId: userID},
					{Id: 3, Status: repository.PENDING, UserId: userID},
				}, nil).Once()
			},
			expectedProgress: &service.ProgressStatus{
				CompleteWorkouts: 2,
				TotalWorkouts:    3,
			},
			expectedErrorType: nil,
		},
		{
			name:   "Successful progress report with no workouts",
			userID: userID,
			mockRepoSetup: func(mwr *MockWorkoutForReportRepository) {
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.COMPLETED, true).Return([]repository.WorkoutPlan{}, nil).Once()
				mwr.On("ListUserWorkouts", ctx, userID).Return([]repository.WorkoutPlan{}, nil).Once()
			},
			expectedProgress: &service.ProgressStatus{
				CompleteWorkouts: 0,
				TotalWorkouts:    0,
			},
			expectedErrorType: nil,
		},
		{
			name:   "Error from ListWorkoutsByStatus",
			userID: userID,
			mockRepoSetup: func(mwr *MockWorkoutForReportRepository) {
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.COMPLETED, true).Return(nil, errors.New("db error completed workouts")).Once()
			},
			expectedProgress:  nil,
			expectedErrorType: errors.New("failed to fetch workout plans by filter: db error completed workouts"),
		},
		{
			name:   "Error from ListUserWorkouts",
			userID: userID,
			mockRepoSetup: func(mwr *MockWorkoutForReportRepository) {
				// Completed workouts call succeeds
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.COMPLETED, true).Return([]repository.WorkoutPlan{
					{Id: 1, Status: repository.COMPLETED, UserId: userID},
				}, nil).Once()
				// All workouts call fails
				mwr.On("ListUserWorkouts", ctx, userID).Return(nil, errors.New("db error all workouts")).Once()
			},
			expectedProgress:  nil,
			expectedErrorType: errors.New("failed to fetch all workout plans: db error all workouts"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWorkoutRepo := new(MockWorkoutForReportRepository)
			tt.mockRepoSetup(mockWorkoutRepo)

			reportService := service.NewReportService(mockWorkoutRepo)
			progress, err := reportService.Progress(ctx, tt.userID)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
				assert.Nil(t, progress)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, progress)
				assert.Equal(t, tt.expectedProgress.CompleteWorkouts, progress.CompleteWorkouts)
				assert.Equal(t, tt.expectedProgress.TotalWorkouts, progress.TotalWorkouts)
			}

			mockWorkoutRepo.AssertExpectations(t)
		})
	}
}
