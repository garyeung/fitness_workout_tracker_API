package service

import (
	"context"
	"fmt"
	"workout-tracker-api/internal/repository"
)

type ExerciseServiceInterface interface {
	GetExerciseById(ctx context.Context, id int) (*Exercise, error)
	ListExercises(ctx context.Context) ([]Exercise, error)
}

type ExerciseService struct {
	exerciseRepo repository.ExerciseRepository
}

func NewExerciseService(r repository.ExerciseRepository) ExerciseServiceInterface {
	return &ExerciseService{
		exerciseRepo: r,
	}
}

type MuscleGroup string

type Exercise struct {
	Id          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	MuscleGroup MuscleGroup `json:"muscleGroup"`
}

const (
	Chest     MuscleGroup = "chest"
	Legs      MuscleGroup = "legs"
	Back      MuscleGroup = "back"
	Shoulders MuscleGroup = "shoulders"
	Arms      MuscleGroup = "arms"
	Core      MuscleGroup = "core"
	Glutes    MuscleGroup = "glutes"
)

// list exercises
func (s *ExerciseService) ListExercises(ctx context.Context) ([]Exercise, error) {
	exercisesList, err := s.exerciseRepo.ListExercises(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list exercises: %w", err)
	}

	var result []Exercise

	for _, e := range exercisesList {
		serviceExercise := toServiceExercise(&e)

		result = append(result, *serviceExercise)
	}

	return result, nil
}

// get exercise
func (s *ExerciseService) GetExerciseById(ctx context.Context, id int) (*Exercise, error) {
	exercise, err := s.exerciseRepo.GetExerciseById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("failed to get exercise by id '%v': %w", id, err)
	}

	result := toServiceExercise(exercise)

	return result, nil
}

func toServiceExercise(mu *repository.Exercise) *Exercise {
	if mu == nil {
		return nil
	}

	return &Exercise{
		Id:          mu.Id,
		Name:        mu.Name,
		Description: mu.Description,
		MuscleGroup: MuscleGroup(mu.MuscleGroup),
	}
}
