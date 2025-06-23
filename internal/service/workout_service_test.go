package service_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/service" // Your service package
	// Assuming this utility exists
)

// --- Mocks ---

// MockWorkoutRepository is a mock implementation of repository.WorkoutRepository
type MockWorkoutRepository struct {
	mock.Mock
}

func (m *MockWorkoutRepository) CreateWorkout(ctx context.Context, data repository.CreateWP) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepository) GetWorkoutById(ctx context.Context, id int) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepository) UpdateWorkout(ctx context.Context, data repository.UpdateWP) (*repository.WorkoutPlan, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepository) DeleteWorkoutById(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockWorkoutRepository) ListWorkoutsByStatus(ctx context.Context, userID int, status repository.WPStatus, asc bool) ([]repository.WorkoutPlan, error) {
	args := m.Called(ctx, userID, status, asc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WorkoutPlan), args.Error(1)
}
func (m *MockWorkoutRepository) ListUserWorkouts(ctx context.Context, userID int) ([]repository.WorkoutPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.WorkoutPlan), args.Error(1)
}

// MockExercisePlanRepository is a mock implementation of repository.ExercisePlanRepository
type MockExercisePlanRepository struct {
	mock.Mock
}

func (m *MockExercisePlanRepository) CreateExercisePlan(ctx context.Context, data repository.CreateEP, workoutPlanID int) (*repository.ExercisePlan, error) {
	args := m.Called(ctx, data, workoutPlanID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ExercisePlan), args.Error(1)
}
func (m *MockExercisePlanRepository) GetExercisePlanById(ctx context.Context, id int) (*repository.ExercisePlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ExercisePlan), args.Error(1)
}
func (m *MockExercisePlanRepository) UpdateExercisePlan(ctx context.Context, data repository.UpdateEP) (*repository.ExercisePlan, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ExercisePlan), args.Error(1)
}
func (m *MockExercisePlanRepository) DeleteExercisePlanByID(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *MockExercisePlanRepository) ListExercisePlans(ctx context.Context, workoutID int) ([]repository.ExercisePlan, error) {
	args := m.Called(ctx, workoutID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.ExercisePlan), args.Error(1)
}

// --- Tests ---

func TestWorkoutService_CreateWorkout(t *testing.T) {
	ctx := context.Background()
	now := time.Now().Truncate(time.Second)
	scheduledDate := now.UTC().Add(24 * time.Hour)

	tests := []struct {
		name              string
		input             service.WorkoutPlanCreate
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkout   *service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name: "Successful workout creation with exercise plans",
			input: service.WorkoutPlanCreate{
				UserId:        1,
				ScheduledDate: &scheduledDate,
				ExercisePlans: []service.ExercisePlanCreate{
					{ExerciseId: 10, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: service.KG},
					{ExerciseId: 20, Sets: 4, Repetitions: 8, Weights: 70, WeightUnit: service.LBS},
				},
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("CreateWorkout", ctx, repository.CreateWP{
					UserId:        1,
					ScheduledDate: scheduledDate,
					Comment:       nil,
				}).Return(&repository.WorkoutPlan{
					Id:            1,
					UserId:        1,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					Comment:       sql.NullString{Valid: false},
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("CreateExercisePlan", ctx, repository.CreateEP{
					ExerciseId: 10, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: repository.KG,
				}, 1).Return(&repository.ExercisePlan{
					Id: 1, ExerciseId: 10, WorkoutPlanId: 1, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: repository.KG,
				}, nil).Once()
				mer.On("CreateExercisePlan", ctx, repository.CreateEP{
					ExerciseId: 20, Sets: 4, Repetitions: 8, Weights: 70, WeightUnit: repository.LBS,
				}, 1).Return(&repository.ExercisePlan{
					Id: 2, ExerciseId: 20, WorkoutPlanId: 1, Sets: 4, Repetitions: 8, Weights: 70, WeightUnit: repository.LBS,
				}, nil).Once()
			},
			expectedWorkout: &service.WorkoutPlan{
				Id:            1,
				UserId:        1,
				Status:        service.PENDING,
				ScheduledDate: scheduledDate,
				Comment:       nil,
				ExercisePlans: []service.ExercisePlan{
					{Id: 1, ExerciseId: 10, WorkoutPlanId: 1, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: service.KG},
					{Id: 2, ExerciseId: 20, WorkoutPlanId: 1, Sets: 4, Repetitions: 8, Weights: 70, WeightUnit: service.LBS},
				},
			},
			expectedErrorType: nil,
		},
		{
			name: "Invalid workout plan user ID",
			input: service.WorkoutPlanCreate{
				UserId:        0, // Invalid
				ScheduledDate: &scheduledDate,
				ExercisePlans: []service.ExercisePlanCreate{},
			},
			mockWPRepoSetup:   func(mwr *MockWorkoutRepository) {},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: &apperrors.ValidationError{Field: apperrors.INVALID_ID},
		},
		{
			name: "Invalid scheduled date format",
			input: service.WorkoutPlanCreate{
				UserId:        1,
				ScheduledDate: nil, // Invalid
				ExercisePlans: []service.ExercisePlanCreate{},
			},
			mockWPRepoSetup:   func(mwr *MockWorkoutRepository) {},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: &apperrors.ValidationError{Field: apperrors.INVALID_DATE},
		},
		{
			name: "Invalid exercise plan sets",
			input: service.WorkoutPlanCreate{
				UserId:        1,
				ScheduledDate: &scheduledDate,
				ExercisePlans: []service.ExercisePlanCreate{
					{ExerciseId: 10, Sets: 0, Repetitions: 10, Weights: 50, WeightUnit: service.KG}, // Invalid sets
				},
			},
			mockWPRepoSetup:   func(mwr *MockWorkoutRepository) {},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: &apperrors.ValidationError{Field: apperrors.INVALID_SETTING},
		},
		{
			name: "Error from WPRepo.CreateWorkout",
			input: service.WorkoutPlanCreate{
				UserId:        1,
				ScheduledDate: &scheduledDate,
				ExercisePlans: []service.ExercisePlanCreate{},
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("CreateWorkout", ctx, mock.AnythingOfType("repository.CreateWP")).
					Return(nil, errors.New("db error creating workout")).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to create workout plan: db error creating workout"), // Wrapped error
		},
		{
			name: "Error from EPRepo.CreateExercisePlan",
			input: service.WorkoutPlanCreate{
				UserId:        1,
				ScheduledDate: &scheduledDate,
				ExercisePlans: []service.ExercisePlanCreate{
					{ExerciseId: 10, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: service.KG},
				},
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("CreateWorkout", ctx, mock.AnythingOfType("repository.CreateWP")).Return(&repository.WorkoutPlan{
					Id:            1,
					UserId:        1,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					Comment:       sql.NullString{Valid: false},
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("CreateExercisePlan", ctx, mock.AnythingOfType("repository.CreateEP"), 1).
					Return(nil, errors.New("db error creating exercise plan")).Once()
			},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to create exercise plan: db error creating exercise plan"), // Wrapped error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workout, err := workoutService.CreateWorkout(ctx, tt.input)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.IsType(t, tt.expectedErrorType, validationErr) // Check type of validation error
					assert.Equal(t, tt.expectedErrorType.(*apperrors.ValidationError).Field, validationErr.Field)
				} else {
					assert.EqualError(t, err, tt.expectedErrorType.Error()) // For wrapped generic errors
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workout)
				assert.Equal(t, tt.expectedWorkout.Id, workout.Id)
				assert.Equal(t, tt.expectedWorkout.UserId, workout.UserId)
				assert.Equal(t, tt.expectedWorkout.Status, workout.Status)
				assert.Equal(t, tt.expectedWorkout.ScheduledDate, workout.ScheduledDate)
				assert.Equal(t, tt.expectedWorkout.Comment, workout.Comment)
				// Deep compare exercise plans
				assert.Equal(t, len(tt.expectedWorkout.ExercisePlans), len(workout.ExercisePlans))
				for i := range tt.expectedWorkout.ExercisePlans {
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].ExerciseId, workout.ExercisePlans[i].ExerciseId)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].Sets, workout.ExercisePlans[i].Sets)
					// ... add more assertions for other fields of ExercisePlan
				}
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_GetWorkoutById(t *testing.T) {
	ctx := context.Background()
	now := time.Now().Truncate(time.Second)
	scheduledDate := now.UTC().Add(24 * time.Hour)

	tests := []struct {
		name              string
		workoutID         int
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkout   *service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name:      "Successful retrieval with exercise plans",
			workoutID: 1,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, 1).Return(&repository.WorkoutPlan{
					Id:            1,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					Comment:       sql.NullString{Valid: false},
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, 1).Return([]repository.ExercisePlan{
					{Id: 1, ExerciseId: 10, WorkoutPlanId: 1, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: repository.KG},
				}, nil).Once()
			},
			expectedWorkout: &service.WorkoutPlan{
				Id:            1,
				UserId:        100,
				Status:        service.PENDING,
				ScheduledDate: scheduledDate,
				Comment:       nil,
				CreatedAt:     now,
				UpdatedAt:     now,
				ExercisePlans: []service.ExercisePlan{
					{Id: 1, ExerciseId: 10, WorkoutPlanId: 1, Sets: 3, Repetitions: 10, Weights: 50, WeightUnit: service.KG},
				},
			},
			expectedErrorType: nil,
		},
		{
			name:      "Workout plan not found",
			workoutID: 99,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, 99).Return(nil, apperrors.ErrNotFound).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to get workout plan: resource not found"), // Wrapped apperrors.ErrNotFound
		},
		{
			name:      "DB error getting workout plan",
			workoutID: 1,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, 1).Return(nil, errors.New("db error")).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to get workout plan: db error"), // Wrapped generic error
		},
		{
			name:      "DB error listing exercise plans",
			workoutID: 1,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, 1).Return(&repository.WorkoutPlan{
					Id:            1,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					Comment:       sql.NullString{Valid: false},
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, 1).Return(nil, errors.New("db error listing exercise plans")).Once()
			},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to get exercise plans: db error listing exercise plans"), // Wrapped generic error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workout, err := workoutService.GetWorkoutById(ctx, tt.workoutID)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workout)
				assert.Equal(t, tt.expectedWorkout.Id, workout.Id)
				// Add more assertions for other fields and exercise plans
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_ListWorkouts(t *testing.T) {
	ctx := context.Background()
	userID := 100
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name              string
		userID            int
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkouts  []service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name:   "Successful list with multiple workouts and exercise plans",
			userID: userID,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListUserWorkouts", ctx, userID).Return([]repository.WorkoutPlan{
					{Id: 1, UserId: userID, Status: repository.PENDING, ScheduledDate: now.Add(24 * time.Hour), CreatedAt: now, UpdatedAt: now},
					{Id: 2, UserId: userID, Status: repository.COMPLETED, ScheduledDate: now.Add(48 * time.Hour), CreatedAt: now, UpdatedAt: now},
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, 1).Return([]repository.ExercisePlan{
					{Id: 10, ExerciseId: 1, WorkoutPlanId: 1, Sets: 3},
				}, nil).Once()
				mer.On("ListExercisePlans", ctx, 2).Return([]repository.ExercisePlan{
					{Id: 20, ExerciseId: 2, WorkoutPlanId: 2, Sets: 2},
				}, nil).Once()
			},
			expectedWorkouts: []service.WorkoutPlan{
				{
					Id:            1,
					UserId:        userID,
					Status:        service.PENDING,
					ScheduledDate: now.Add(24 * time.Hour),
					CreatedAt:     now,
					UpdatedAt:     now,
					ExercisePlans: []service.ExercisePlan{{Id: 10, ExerciseId: 1, WorkoutPlanId: 1, Sets: 3}},
				},
				{
					Id:            2,
					UserId:        userID,
					Status:        service.COMPLETED,
					ScheduledDate: now.Add(48 * time.Hour),
					CreatedAt:     now,
					UpdatedAt:     now,
					ExercisePlans: []service.ExercisePlan{{Id: 20, ExerciseId: 2, WorkoutPlanId: 2, Sets: 2}},
				},
			},
			expectedErrorType: nil,
		},
		{
			name:   "Successful list with no workouts",
			userID: userID,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListUserWorkouts", ctx, userID).Return([]repository.WorkoutPlan{}, nil).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {}, // No EP calls if no WP
			expectedWorkouts:  []service.WorkoutPlan{},
			expectedErrorType: nil,
		},
		{
			name:   "Error from WPRepo.ListUserWorkouts",
			userID: userID,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListUserWorkouts", ctx, userID).Return(nil, errors.New("db error listing workouts")).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkouts:  nil,
			expectedErrorType: errors.New("failed to fetched workout plans: db error listing workouts"),
		},
		{
			name:   "Error from EPRepo.ListExercisePlans",
			userID: userID,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListUserWorkouts", ctx, userID).Return([]repository.WorkoutPlan{
					{Id: 1, UserId: userID, Status: repository.PENDING, ScheduledDate: now.Add(24 * time.Hour), CreatedAt: now, UpdatedAt: now},
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, 1).Return(nil, errors.New("db error listing exercise plans")).Once()
			},
			expectedWorkouts:  nil,
			expectedErrorType: errors.New("failed to fetched exercise plans: db error listing exercise plans"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workouts, err := workoutService.ListWorkouts(ctx, tt.userID)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedWorkouts), len(workouts))
				for i := range tt.expectedWorkouts {
					assert.Equal(t, tt.expectedWorkouts[i].Id, workouts[i].Id)
					assert.Equal(t, tt.expectedWorkouts[i].UserId, workouts[i].UserId)
					assert.Equal(t, tt.expectedWorkouts[i].Status, workouts[i].Status)
					assert.True(t, workouts[i].ScheduledDate.Equal(tt.expectedWorkouts[i].ScheduledDate)) // String comparison
					assert.Equal(t, tt.expectedWorkouts[i].Comment, workouts[i].Comment)
					assert.Equal(t, len(tt.expectedWorkouts[i].ExercisePlans), len(workouts[i].ExercisePlans))
					// Add deep comparison for exercise plans if needed
				}
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_ListWorkoutsByStatus(t *testing.T) {
	ctx := context.Background()
	userID := 100
	status := service.PENDING
	asc := true
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name              string
		userID            int
		status            service.WPStatus
		asc               bool
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkouts  []service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name:   "Successful list by status (PENDING, ASC)",
			userID: userID,
			status: status,
			asc:    asc,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.WPStatus(status), asc).Return([]repository.WorkoutPlan{
					{Id: 1, UserId: userID, Status: repository.PENDING, ScheduledDate: now.Add(24 * time.Hour), CreatedAt: now, UpdatedAt: now},
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, 1).Return([]repository.ExercisePlan{
					{Id: 10, ExerciseId: 1, WorkoutPlanId: 1, Sets: 3},
				}, nil).Once()
			},
			expectedWorkouts: []service.WorkoutPlan{
				{
					Id:            1,
					UserId:        userID,
					Status:        service.PENDING,
					ScheduledDate: now.Add(24 * time.Hour),
					CreatedAt:     now,
					UpdatedAt:     now,
					ExercisePlans: []service.ExercisePlan{{Id: 10, ExerciseId: 1, WorkoutPlanId: 1, Sets: 3}},
				},
			},
			expectedErrorType: nil,
		},
		{
			name:   "Error from WPRepo.ListWorkoutsByStatus",
			userID: userID,
			status: status,
			asc:    asc,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("ListWorkoutsByStatus", ctx, userID, repository.WPStatus(status), asc).Return(nil, errors.New("db error listing by status")).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkouts:  nil,
			expectedErrorType: errors.New("failed to fetched workout plans with filters: db error listing by status"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workouts, err := workoutService.ListWorkoutsByStatus(ctx, tt.userID, tt.status, tt.asc)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workouts)
				assert.Equal(t, len(tt.expectedWorkouts), len(workouts))
				// Add more detailed assertions for content
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_CompleteWorkout(t *testing.T) {
	ctx := context.Background()
	workoutID := 1
	comment := "Great session!"

	tests := []struct {
		name              string
		workoutID         int
		comment           *string
		mockWPRepoSetup   func(*MockWorkoutRepository)
		expectedErrorType error
	}{
		{
			name:      "Successful completion with comment",
			workoutID: workoutID,
			comment:   &comment,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, repository.UpdateWP{
					Id:      workoutID,
					Status:  repository.COMPLETED,
					Comment: &comment,
				}).Return(&repository.WorkoutPlan{}, nil).Once()
			},
			expectedErrorType: nil,
		},
		{
			name:      "Successful completion without comment",
			workoutID: workoutID,
			comment:   nil,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, repository.UpdateWP{
					Id:      workoutID,
					Status:  repository.COMPLETED,
					Comment: nil,
				}).Return(&repository.WorkoutPlan{}, nil).Once()
			},
			expectedErrorType: nil,
		},
		{
			name:      "Workout not found",
			workoutID: 99,
			comment:   nil,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, mock.AnythingOfType("repository.UpdateWP")).Return(nil, apperrors.ErrNotFound).Once()
			},
			expectedErrorType: errors.New("failed to set complete status to workout plan: resource not found"),
		},
		{
			name:      "DB error updating workout",
			workoutID: 1,
			comment:   nil,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, mock.AnythingOfType("repository.UpdateWP")).Return(nil, errors.New("db update error")).Once()
			},
			expectedErrorType: errors.New("failed to set complete status to workout plan: db update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository) // Still need to pass it, even if not used

			tt.mockWPRepoSetup(mockWPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			err := workoutService.CompleteWorkout(ctx, tt.workoutID, tt.comment)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
			} else {
				assert.NoError(t, err)
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_ScheduleWorkout(t *testing.T) {
	ctx := context.Background()
	workoutID := 1
	scheduledDateValid := time.Now().UTC().Add(7 * 24 * time.Hour).Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	tests := []struct {
		name              string
		workoutID         int
		scheduledDate     *time.Time
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkout   *service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name:          "Successful schedule",
			workoutID:     workoutID,
			scheduledDate: &scheduledDateValid,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, repository.UpdateWP{
					Id:            workoutID,
					Status:        repository.PENDING,
					ScheduledDate: &scheduledDateValid,
				}).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDateValid,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, workoutID).Return([]repository.ExercisePlan{}, nil).Once()
			},
			expectedWorkout: &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        100,
				Status:        service.PENDING,
				ScheduledDate: scheduledDateValid,
				CreatedAt:     now,
				UpdatedAt:     now,
				ExercisePlans: []service.ExercisePlan{},
			},
			expectedErrorType: nil,
		},
		{
			name:          "not set scheduled date",
			workoutID:     workoutID,
			scheduledDate: nil, // not set
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, repository.UpdateWP{
					Id:            workoutID,
					Status:        repository.PENDING,
					ScheduledDate: nil,
				}).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDateValid,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, workoutID).Return([]repository.ExercisePlan{}, nil).Once()

			},
			expectedWorkout: &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        100,
				Status:        service.PENDING,
				ScheduledDate: scheduledDateValid,
				CreatedAt:     now,
				UpdatedAt:     now,
				ExercisePlans: []service.ExercisePlan{},
			},
			expectedErrorType: nil,
		},
		{
			name:          "Workout not found during update",
			workoutID:     99,
			scheduledDate: &scheduledDateValid,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, mock.AnythingOfType("repository.UpdateWP")).Return(nil, apperrors.ErrNotFound).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to update workout schedule: resource not found"),
		},
		{
			name:          "DB error updating workout",
			workoutID:     1,
			scheduledDate: &scheduledDateValid,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, mock.AnythingOfType("repository.UpdateWP")).Return(nil, errors.New("db update error")).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to update workout schedule: db update error"),
		},
		{
			name:          "DB error listing exercise plans after update",
			workoutID:     1,
			scheduledDate: &scheduledDateValid,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("UpdateWorkout", ctx, mock.AnythingOfType("repository.UpdateWP")).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDateValid,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				mer.On("ListExercisePlans", ctx, workoutID).Return(nil, errors.New("db error listing exercise plans")).Once()
			},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to fech exercise plans: db error listing exercise plans"), // Typo 'fech' will match the error string
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workout, err := workoutService.ScheduleWorkout(ctx, tt.workoutID, tt.scheduledDate)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.IsType(t, tt.expectedErrorType, validationErr)
					assert.Equal(t, tt.expectedErrorType.(*apperrors.ValidationError).Field, validationErr.Field)
				} else {
					assert.EqualError(t, err, tt.expectedErrorType.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workout)
				assert.Equal(t, tt.expectedWorkout.Id, workout.Id)
				assert.Equal(t, tt.expectedWorkout.ScheduledDate, workout.ScheduledDate)
				// Add more detailed assertions for content
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_UpdateExercisePlans(t *testing.T) {
	ctx := context.Background()
	workoutID := 1
	now := time.Now().UTC().Truncate(time.Second)
	scheduledDate := now.Add(24 * time.Hour).UTC()

	tests := []struct {
		name              string
		workoutID         int
		epsUpdate         []service.ExercisePlanUpdate
		mockWPRepoSetup   func(*MockWorkoutRepository)
		mockEPRepoSetup   func(*MockExercisePlanRepository)
		expectedWorkout   *service.WorkoutPlan
		expectedErrorType error
	}{
		{
			name:      "Successful update of exercise plans",
			workoutID: workoutID,
			epsUpdate: []service.ExercisePlanUpdate{
				{Id: 10, Sets: 5, Repetitions: 15, Weights: 60, WeightUnit: service.KG},
				{Id: 20, Sets: 6, Repetitions: 10, Weights: 80, WeightUnit: service.LBS},
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, workoutID).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				sets1, reps1, weights1, unit1 := 5, 15, float32(60), repository.KG
				sets2, reps2, weights2, unit2 := 6, 10, float32(80), repository.LBS

				mer.On("UpdateExercisePlan", ctx, repository.UpdateEP{
					Id: 10, Sets: &sets1, Repetitions: &reps1, Weights: &weights1, WeightUnit: &unit1,
				}).Return(&repository.ExercisePlan{
					Id: 10, ExerciseId: 101, WorkoutPlanId: workoutID, Sets: sets1, Repetitions: reps1, Weights: weights1, WeightUnit: unit1,
				}, nil).Once()

				mer.On("UpdateExercisePlan", ctx, repository.UpdateEP{
					Id: 20, Sets: &sets2, Repetitions: &reps2, Weights: &weights2, WeightUnit: &unit2,
				}).Return(&repository.ExercisePlan{
					Id: 20, ExerciseId: 102, WorkoutPlanId: workoutID, Sets: sets2, Repetitions: reps2, Weights: weights2, WeightUnit: unit2,
				}, nil).Once()
			},
			expectedWorkout: &service.WorkoutPlan{
				Id:            workoutID,
				UserId:        100,
				Status:        service.PENDING,
				ScheduledDate: scheduledDate,
				CreatedAt:     now,
				UpdatedAt:     now,
				ExercisePlans: []service.ExercisePlan{
					{Id: 10, ExerciseId: 101, WorkoutPlanId: workoutID, Sets: 5, Repetitions: 15, Weights: 60, WeightUnit: service.KG},
					{Id: 20, ExerciseId: 102, WorkoutPlanId: workoutID, Sets: 6, Repetitions: 10, Weights: 80, WeightUnit: service.LBS},
				},
			},
			expectedErrorType: nil,
		},
		{
			name:      "Workout plan not found during initial fetch",
			workoutID: 99,
			epsUpdate: []service.ExercisePlanUpdate{},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, 99).Return(nil, apperrors.ErrNotFound).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to fetch workout plan id '99': resource not found"),
		},
		{
			name:      "Invalid exercise plan update (e.g., negative weights)",
			workoutID: workoutID,
			epsUpdate: []service.ExercisePlanUpdate{
				{Id: 10, Sets: 3, Repetitions: 10, Weights: -10, WeightUnit: service.KG}, // Invalid
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, workoutID).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup:   func(mer *MockExercisePlanRepository) {}, // No EP repo call due to validation error
			expectedWorkout:   nil,
			expectedErrorType: &apperrors.ValidationError{Field: apperrors.INVALID_SETTING},
		},
		{
			name:      "DB error updating exercise plan",
			workoutID: workoutID,
			epsUpdate: []service.ExercisePlanUpdate{
				{Id: 10, Sets: 5, Repetitions: 15, Weights: 60, WeightUnit: service.KG},
			},
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("GetWorkoutById", ctx, workoutID).Return(&repository.WorkoutPlan{
					Id:            workoutID,
					UserId:        100,
					Status:        repository.PENDING,
					ScheduledDate: scheduledDate,
					CreatedAt:     now,
					UpdatedAt:     now,
				}, nil).Once()
			},
			mockEPRepoSetup: func(mer *MockExercisePlanRepository) {
				sets, reps, weights, unit := 5, 15, float32(60), repository.KG
				mer.On("UpdateExercisePlan", ctx, repository.UpdateEP{
					Id: 10, Sets: &sets, Repetitions: &reps, Weights: &weights, WeightUnit: &unit,
				}).Return(nil, errors.New("db error updating exercise plan")).Once()
			},
			expectedWorkout:   nil,
			expectedErrorType: errors.New("failed to update exercise id '10': db error updating exercise plan"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository)

			tt.mockWPRepoSetup(mockWPRepo)
			tt.mockEPRepoSetup(mockEPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			workout, err := workoutService.UpdateExercisePlans(ctx, tt.workoutID, tt.epsUpdate)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.IsType(t, tt.expectedErrorType, validationErr)
					assert.Equal(t, tt.expectedErrorType.(*apperrors.ValidationError).Field, validationErr.Field)
				} else {
					assert.EqualError(t, err, tt.expectedErrorType.Error())
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, workout)
				assert.Equal(t, tt.expectedWorkout.Id, workout.Id)
				assert.Equal(t, tt.expectedWorkout.UserId, workout.UserId)
				assert.Equal(t, tt.expectedWorkout.Status, workout.Status)
				assert.Equal(t, tt.expectedWorkout.ScheduledDate, workout.ScheduledDate)
				assert.Equal(t, tt.expectedWorkout.Comment, workout.Comment)
				assert.Equal(t, len(tt.expectedWorkout.ExercisePlans), len(workout.ExercisePlans))
				for i := range tt.expectedWorkout.ExercisePlans {
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].Id, workout.ExercisePlans[i].Id)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].ExerciseId, workout.ExercisePlans[i].ExerciseId)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].WorkoutPlanId, workout.ExercisePlans[i].WorkoutPlanId)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].Sets, workout.ExercisePlans[i].Sets)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].Repetitions, workout.ExercisePlans[i].Repetitions)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].Weights, workout.ExercisePlans[i].Weights)
					assert.Equal(t, tt.expectedWorkout.ExercisePlans[i].WeightUnit, workout.ExercisePlans[i].WeightUnit)
				}
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}

func TestWorkoutService_DeleteWorkoutById(t *testing.T) {
	ctx := context.Background()
	workoutID := 1

	tests := []struct {
		name              string
		workoutID         int
		mockWPRepoSetup   func(*MockWorkoutRepository)
		expectedErrorType error
	}{
		{
			name:      "Successful deletion",
			workoutID: workoutID,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("DeleteWorkoutById", ctx, workoutID).Return(nil).Once()
			},
			expectedErrorType: nil,
		},
		{
			name:      "Workout not found",
			workoutID: 99,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("DeleteWorkoutById", ctx, 99).Return(apperrors.ErrNotFound).Once()
			},
			expectedErrorType: errors.New("failed to delete workout plan id '99': resource not found"),
		},
		{
			name:      "DB error during deletion",
			workoutID: 1,
			mockWPRepoSetup: func(mwr *MockWorkoutRepository) {
				mwr.On("DeleteWorkoutById", ctx, 1).Return(errors.New("db delete error")).Once()
			},
			expectedErrorType: errors.New("failed to delete workout plan id '1': db delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWPRepo := new(MockWorkoutRepository)
			mockEPRepo := new(MockExercisePlanRepository) // Still need to pass it, even if not used by Delete method

			tt.mockWPRepoSetup(mockWPRepo)

			workoutService := service.NewWPService(mockWPRepo, mockEPRepo)
			err := workoutService.DeleteWorkoutById(ctx, tt.workoutID)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErrorType.Error())
			} else {
				assert.NoError(t, err)
			}

			mockWPRepo.AssertExpectations(t)
			mockEPRepo.AssertExpectations(t)
		})
	}
}
