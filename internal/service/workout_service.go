package service

import (
	"context"
	"fmt"
	"time"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
)

type WeightUnit string

const (
	KG    WeightUnit = "kg"
	LBS   WeightUnit = "lbs"
	OTHER WeightUnit = "other"
)

type ExercisePlan struct {
	Id            int        `json:"id"`
	ExerciseId    int        `json:"exerciseId"`
	WorkoutPlanId int        `json:"workoutPlanId"`
	Sets          int        `json:"sets"`
	Repetitions   int        `json:"repetitions"`
	Weights       float32    `json:"weights"`
	WeightUnit    WeightUnit `json:"weightUnit"`
}

type ExercisePlanCreate struct {
	ExerciseId  int        `json:"exerciseId"`
	Sets        int        `json:"sets"`
	Repetitions int        `json:"repetitions"`
	Weights     float32    `json:"weights"`
	WeightUnit  WeightUnit `json:"weightUnit"`
}

func (data *ExercisePlanCreate) Validate() error {
	if data.Sets <= 0 || data.Sets > 999 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "sets can not be zero or too large")
	}

	if data.Repetitions < 0 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "repetitions can not be negative")
	}

	if data.Weights < 0 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "weights can not be negative")
	}

	// if WeightUnit out of type WeightUnit
	switch data.WeightUnit {
	case KG, LBS, OTHER:
	default:
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "invalide weight unit")
	}

	return nil
}

type ExercisePlanUpdate struct {
	Id          int        `json:"id"`
	Sets        int        `json:"sets"`
	Repetitions int        `json:"repetitions"`
	Weights     float32    `json:"weights"`
	WeightUnit  WeightUnit `json:"weightUnit"`
}

func (data *ExercisePlanUpdate) Validate() error {
	if data.Sets <= 0 || data.Sets > 999 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "sets can not be zero or too large")
	}

	if data.Repetitions < 0 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "repetitions can not be negative")
	}

	if data.Weights < 0 {
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "weights can not be negative")
	}

	// if WeightUnit out of type WeightUnit
	switch data.WeightUnit {
	case KG, LBS, OTHER:
	default:
		return apperrors.NewValidationError(apperrors.INVALID_SETTING, "invalide weight unit")
	}

	return nil
}

type WPStatus string

const (
	PENDING   WPStatus = "pending"
	COMPLETED WPStatus = "completed"
	MISSED    WPStatus = "missed"
)

type WorkoutPlan struct {
	Id            int            `json:"id"`
	UserId        int            `json:"userId"`
	Status        WPStatus       `json:"status"`
	ScheduledDate time.Time      `json:"scheduledDate"`
	Comment       *string        `json:"comment,omitempty"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	ExercisePlans []ExercisePlan `json:"exercisePlans"`
}

type WorkoutPlanCreate struct {
	UserId        int                  `json:"userId"`
	ScheduledDate *time.Time           `json:"scheduledDate"`
	ExercisePlans []ExercisePlanCreate `json:"exercisePlans"`
}

func (data *WorkoutPlanCreate) Validate() error {
	// check user id valid or not
	if data.UserId <= 0 {
		return apperrors.NewValidationError(apperrors.INVALID_ID, "not a valid user id")
	}
	// check scheduled date valid or not
	if data.ScheduledDate == nil {
		return apperrors.NewValidationError(apperrors.INVALID_DATE, "not valid scheduled date, date is not set")

	}

	for _, ep := range data.ExercisePlans {
		if err := ep.Validate(); err != nil {

			return err
		}
	}

	return nil
}

type WorkoutServiceInterface interface {
	CreateWorkout(ctx context.Context, data WorkoutPlanCreate) (*WorkoutPlan, error)
	DeleteWorkoutById(ctx context.Context, id int) error
	GetWorkoutById(ctx context.Context, id int) (*WorkoutPlan, error)
	ListWorkouts(ctx context.Context, userId int) ([]WorkoutPlan, error)
	ListWorkoutsByStatus(ctx context.Context, userId int, status WPStatus, asc bool) ([]WorkoutPlan, error)
	CompleteWorkout(ctx context.Context, id int, comment *string) error
	ScheduleWorkout(ctx context.Context, id int, scheduledDate *time.Time) (*WorkoutPlan, error)
	UpdateExercisePlans(ctx context.Context, workoutId int, epsUpdate []ExercisePlanUpdate) (*WorkoutPlan, error)
}

type WorkoutService struct {
	WPRepo repository.WorkoutRepository
	EPRepo repository.ExercisePlanRepository
}

func NewWPService(wr repository.WorkoutRepository, er repository.ExercisePlanRepository) WorkoutServiceInterface {
	return &WorkoutService{
		WPRepo: wr,
		EPRepo: er,
	}
}

func (ws *WorkoutService) ListWorkoutsByStatus(ctx context.Context, userId int, status WPStatus, asc bool) ([]WorkoutPlan, error) {
	wpList, err := ws.WPRepo.ListWorkoutsByStatus(ctx, userId, repository.WPStatus(status), asc)

	if err != nil {
		return nil, fmt.Errorf("failed to fetched workout plans with filters: %w", err)
	}

	var result []WorkoutPlan
	for _, wp := range wpList {
		epList, err := ws.EPRepo.ListExercisePlans(ctx, wp.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetched exercise plans: %w", err)
		}

		serviceWP := toServiceWP(&wp, epList)
		result = append(result, *serviceWP)
	}
	return result, nil

}

func (ws *WorkoutService) ListWorkouts(ctx context.Context, userId int) ([]WorkoutPlan, error) {
	wpList, err := ws.WPRepo.ListUserWorkouts(ctx, userId)

	if err != nil {
		return nil, fmt.Errorf("failed to fetched workout plans: %w", err)
	}

	var result []WorkoutPlan
	for _, wp := range wpList {
		epList, err := ws.EPRepo.ListExercisePlans(ctx, wp.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to fetched exercise plans: %w", err)
		}

		serviceWP := toServiceWP(&wp, epList)
		result = append(result, *serviceWP)
	}
	return result, nil

}

func (ws *WorkoutService) CreateWorkout(ctx context.Context, data WorkoutPlanCreate) (*WorkoutPlan, error) {
	if err := data.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate: %w", err)
	}

	for _, ep := range data.ExercisePlans {
		if err := ep.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate: %w", err)
		}
	}

	// alrealy validate
	workout, err := ws.WPRepo.CreateWorkout(ctx, repository.CreateWP{
		UserId:        data.UserId,
		ScheduledDate: *data.ScheduledDate,
		Comment:       nil,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create workout plan: %w", err)
	}

	var exercisePlans []repository.ExercisePlan
	for _, ep := range data.ExercisePlans {
		exercisePlan, err := ws.EPRepo.CreateExercisePlan(ctx, repository.CreateEP{
			ExerciseId:  ep.ExerciseId,
			Sets:        ep.Sets,
			Repetitions: ep.Repetitions,
			Weights:     ep.Weights,
			WeightUnit:  repository.WeightUnit(ep.WeightUnit),
		}, workout.Id)
		if err != nil {
			return nil, fmt.Errorf("failed to create exercise plan: %w", err)
		}

		exercisePlans = append(exercisePlans, *exercisePlan)
	}

	result := toServiceWP(workout, exercisePlans)

	return result, nil
}

func toServiceWP(wp *repository.WorkoutPlan, eps []repository.ExercisePlan) *WorkoutPlan {
	if wp == nil {
		return nil
	}

	id := wp.Id
	userID := wp.UserId
	status := WPStatus(wp.Status)
	shcduledDate := wp.ScheduledDate
	createdAt := wp.CreatedAt
	updatedAt := wp.UpdatedAt
	var comment *string
	if !wp.Comment.Valid {
		comment = nil
	} else {
		comment = &wp.Comment.String
	}

	var epList []ExercisePlan
	for _, ep := range eps {
		ServiceEP := toServiceEP(&ep)
		epList = append(epList, *ServiceEP)
	}

	return &WorkoutPlan{
		Id:            id,
		UserId:        userID,
		Status:        status,
		ScheduledDate: shcduledDate,
		Comment:       comment,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
		ExercisePlans: epList,
	}

}

func (ws *WorkoutService) GetWorkoutById(ctx context.Context, id int) (*WorkoutPlan, error) {
	workoutPlan, err := ws.WPRepo.GetWorkoutById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get workout plan: %w", err)
	}

	exercisePlans, err := ws.EPRepo.ListExercisePlans(ctx, workoutPlan.Id)

	if err != nil {
		return nil, fmt.Errorf("failed to get exercise plans: %w", err)
	}

	return toServiceWP(workoutPlan, exercisePlans), nil

}
func (ws *WorkoutService) CompleteWorkout(ctx context.Context, id int, comment *string) error {
	_, err := ws.WPRepo.UpdateWorkout(ctx, repository.UpdateWP{
		Id:      id,
		Status:  repository.COMPLETED,
		Comment: comment,
	})

	if err != nil {
		return fmt.Errorf("failed to set complete status to workout plan: %w", err)
	}

	return nil
}

func (ws *WorkoutService) ScheduleWorkout(ctx context.Context, id int, scheduledDate *time.Time) (*WorkoutPlan, error) {
	workout, err := ws.WPRepo.UpdateWorkout(ctx, repository.UpdateWP{
		Id:            id,
		Status:        repository.PENDING,
		ScheduledDate: scheduledDate,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to update workout schedule: %w", err)
	}

	exercisePlans, err := ws.EPRepo.ListExercisePlans(ctx, workout.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to fech exercise plans: %w", err)
	}

	return toServiceWP(workout, exercisePlans), nil

}
func (ws *WorkoutService) UpdateExercisePlans(ctx context.Context, workoutId int, epsUpdate []ExercisePlanUpdate) (*WorkoutPlan, error) {

	workoutPlan, err := ws.WPRepo.GetWorkoutById(ctx, workoutId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workout plan id '%v': %w", workoutId, err)
	}

	var exercisePlans []repository.ExercisePlan
	for _, ep := range epsUpdate {
		if err := ep.Validate(); err != nil {
			return nil, fmt.Errorf("failed to validate exercise plan id '%v': %w", ep.Id, err)
		}

		exercisePlan, err := ws.EPRepo.UpdateExercisePlan(ctx, repository.UpdateEP{
			Id:          ep.Id,
			Sets:        &ep.Sets,
			Repetitions: &ep.Repetitions,
			Weights:     &ep.Weights,
			WeightUnit:  (*repository.WeightUnit)(&ep.WeightUnit),
		})

		if err != nil {
			return nil, fmt.Errorf("failed to update exercise id '%v': %w", ep.Id, err)
		}

		exercisePlans = append(exercisePlans, *exercisePlan)
	}

	return toServiceWP(workoutPlan, exercisePlans), nil

}

func (ws *WorkoutService) DeleteWorkoutById(ctx context.Context, id int) error {
	// already including delete exercise plans
	err := ws.WPRepo.DeleteWorkoutById(ctx, id)

	if err != nil {
		return fmt.Errorf("failed to delete workout plan id '%v': %w", id, err)
	}

	return nil
}

func toServiceEP(ep *repository.ExercisePlan) *ExercisePlan {

	if ep == nil {
		return nil
	}

	return &ExercisePlan{
		Id:            ep.Id,
		ExerciseId:    ep.ExerciseId,
		WorkoutPlanId: ep.WorkoutPlanId,
		Sets:          ep.Sets,
		Repetitions:   ep.Repetitions,
		Weights:       ep.Weights,
		WeightUnit:    WeightUnit(ep.WeightUnit),
	}
}
