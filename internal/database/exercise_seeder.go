package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"workout-tracker-api/internal/repository"
)

func ExtractSeed(seedFile string) (*[]repository.CreateExer, error) {
	var exercises []repository.CreateExer
	file, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read exercises.json: %w", err)
	}
	// extract data and fit the structure
	err = json.Unmarshal(file, &exercises)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal exercises.json: %w", err)
	}

	return &exercises, nil

}
func ExercisesSeeder(exercises []repository.CreateExer, db *sql.DB) error {
	exerciseRepo := repository.NewExerRepository(db)
	ctx := context.Background()
	for _, exer := range exercises {

		_, err := exerciseRepo.CreateExercise(ctx, exer)

		if err != nil {
			return fmt.Errorf("error seeding exercises: %w", err)
		}

	}

	return nil
}
