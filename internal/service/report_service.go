package service

import (
	"context"
	"fmt"
	"workout-tracker-api/internal/repository"
)

type ReportServiceInterface interface {
	Progress(ctx context.Context, userID int) (*ProgressStatus, error)
}

type ReportService struct {
	workoutRepo repository.WorkoutRepository
}

func NewReportService(wr repository.WorkoutRepository) ReportServiceInterface {
	return &ReportService{
		workoutRepo: wr,
	}
}

type ProgressStatus struct {
	CompleteWorkouts int `json:"completedWorkouts"`
	TotalWorkouts    int `json:"totalWorkouts"`
}

func (s *ReportService) Progress(ctx context.Context, userID int) (*ProgressStatus, error) {
	completed, err := s.workoutRepo.ListWorkoutsByStatus(ctx, userID, repository.COMPLETED, true)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workout plans by filter: %w", err)
	}

	all, err := s.workoutRepo.ListUserWorkouts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all workout plans: %w", err)
	}

	return &ProgressStatus{
		CompleteWorkouts: len(completed),
		TotalWorkouts:    len(all),
	}, nil

}
