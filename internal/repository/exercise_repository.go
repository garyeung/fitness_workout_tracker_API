package repository

import (
	"context"
	"database/sql"
	"fmt"
	"workout-tracker-api/internal/apperrors"
)

type MuscleGroup string

type Exercise struct {
	Id          int         `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	MuscleGroup MuscleGroup `json:"muscleGroup"`
}

type CreateExer struct {
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

type ExerciseRepository interface {
	CreateExercise(ctx context.Context, data CreateExer) (*Exercise, error)
	DeleteExercise(ctx context.Context, id int) error
	GetExerciseById(ctx context.Context, id int) (*Exercise, error)
	ListExercises(ctx context.Context) ([]Exercise, error)
}

type postgresExerRepository struct {
	db *sql.DB
}

func NewExerRepository(db *sql.DB) ExerciseRepository {
	return &postgresExerRepository{
		db: db,
	}
}

func (r *postgresExerRepository) GetExerciseById(ctx context.Context, id int) (*Exercise, error) {
	var exercise Exercise

	query := `SELECT id, name, description, muscle_group FROM exercises WHERE id = $1`

	row, err := executeQueryRow(ctx, r.db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query exercise by id '%v': %w", id, err)
	}

	err = row.Scan(&exercise.Id, &exercise.Name, &exercise.Description, &exercise.MuscleGroup)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to get exercise by id '%v': %w", id, err)
	}

	return &exercise, nil

}

func (r *postgresExerRepository) ListExercises(ctx context.Context) ([]Exercise, error) {
	query := `SELECT id, name, description, muscle_group FROM exercises`

	rows, err := executeQuery(ctx, r.db, query)

	if err != nil {
		return nil, fmt.Errorf("failed to query all exercises: %w", err)
	}
	defer rows.Close()

	var esList []Exercise
	for rows.Next() {
		var exercise Exercise
		if err := rows.Scan(
			&exercise.Id,
			&exercise.Name,
			&exercise.Description,
			&exercise.MuscleGroup); err != nil {
			return nil, fmt.Errorf("failed to scan exercise row: %w", err)
		}
		esList = append(esList, exercise)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating exercise rows: %w", err)
	}

	return esList, nil
}

func (r *postgresExerRepository) CreateExercise(ctx context.Context, data CreateExer) (*Exercise, error) {
	var newExercise Exercise

	err := executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {
		insertQuery := `INSERT INTO exercises (name, description, muscle_group) VALUES ($1, $2, $3) RETURNING id, name, description, muscle_group`

		err := tx.QueryRowContext(txCtx, insertQuery,
			data.Name,
			data.Description,
			data.MuscleGroup,
		).Scan(
			&newExercise.Id,
			&newExercise.Name,
			&newExercise.Description,
			&newExercise.MuscleGroup,
		)
		if err != nil {
			return fmt.Errorf("failed to insert and scan new exercise : %w", err)
		}
		return nil // Commit transaction
	})
	if err != nil {
		return nil, err // Error from executeTransaction or the callback
	}

	return &newExercise, nil
}

func (r *postgresExerRepository) DeleteExercise(ctx context.Context, id int) error {
	deleteQuery := `DELETE FROM exercises WHERE id = $1`

	result, err := executeNonQuery(ctx, r.db, deleteQuery, id) // Use executeNonQuery
	if err != nil {
		return fmt.Errorf("failed to delete exercise with id '%v': %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting exercise with id '%v': %w", id, err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	return nil

}
