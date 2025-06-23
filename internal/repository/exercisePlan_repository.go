package repository

import (
	"context"
	"database/sql"
	"fmt"
	"workout-tracker-api/internal/apperrors"
)

type WeightUnit string
type ExercisePlan struct {
	Id            int        `json:"id"`
	ExerciseId    int        `json:"exerciseId"`
	WorkoutPlanId int        `json:"workoutPlanId"`
	Sets          int        `json:"sets"`
	Repetitions   int        `json:"repetitions"`
	Weights       float32    `json:"weights"`
	WeightUnit    WeightUnit `json:"weightUnit"`
}

type CreateEP struct {
	ExerciseId  int        `json:"exerciseId"`
	Sets        int        `json:"sets"`
	Repetitions int        `json:"repetitions"`
	Weights     float32    `json:"weights"`
	WeightUnit  WeightUnit `json:"weightUnit"`
}

type UpdateEP struct {
	Id          int         `json:"id"`
	Sets        *int        `json:"sets,omitempty"`
	Repetitions *int        `json:"repetitions,omitempty"`
	Weights     *float32    `json:"weights,omitempty"`
	WeightUnit  *WeightUnit `json:"weightUnit,omitempty"`
}

const (
	KG    WeightUnit = "kg"
	LBS   WeightUnit = "lbs"
	OTHER WeightUnit = "other"
)

type ExercisePlanRepository interface {
	CreateExercisePlan(ctx context.Context, data CreateEP, workoutPlanID int) (*ExercisePlan, error)
	GetExercisePlanById(ctx context.Context, id int) (*ExercisePlan, error)
	UpdateExercisePlan(ctx context.Context, data UpdateEP) (*ExercisePlan, error)
	DeleteExercisePlanByID(ctx context.Context, id int) error
	ListExercisePlans(ctx context.Context, workoutID int) ([]ExercisePlan, error)
}

type postgresEPRepository struct {
	db *sql.DB
}

func NewEPRepository(db *sql.DB) ExercisePlanRepository {
	return &postgresEPRepository{
		db: db,
	}
}

func (r *postgresEPRepository) CreateExercisePlan(ctx context.Context, data CreateEP, workoutPlanId int) (*ExercisePlan, error) {
	var newExercisePlan ExercisePlan

	err := executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {
		// 1. Verify that the workout plan exists
		var currentWorkoutPlanID int
		err := tx.QueryRowContext(txCtx, "SELECT id FROM workout_plans WHERE id = $1", workoutPlanId).Scan(&currentWorkoutPlanID)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrForeignKeyViolation
			}
			return fmt.Errorf("failed to query workout plan id '%d' for creating exercise plan: %w", workoutPlanId, err)
		}

		// 2. Insert the new exercise plan
		insertQuery := `INSERT INTO exercise_plans (
			exercise_id, 
			workout_plan_id,
			sets,
			repetitions,
			weights,
			weight_unit
			) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, 
			exercise_id, 
			workout_plan_id,
			sets,
			repetitions,
			weights,
			weight_unit
			`

		// Use tx.QueryRow for INSERT ... RETURNING
		err = tx.QueryRowContext(txCtx,
			insertQuery,
			data.ExerciseId,
			workoutPlanId, // Use the validated workoutPlanID
			data.Sets,
			data.Repetitions,
			data.Weights,
			data.WeightUnit,
		).Scan(
			&newExercisePlan.Id,
			&newExercisePlan.ExerciseId,
			&newExercisePlan.WorkoutPlanId,
			&newExercisePlan.Sets,
			&newExercisePlan.Repetitions,
			&newExercisePlan.Weights,
			&newExercisePlan.WeightUnit,
		)
		if err != nil {
			return fmt.Errorf("failed to insert and scan new exercise plan: %w", err)
		}
		return nil // Commit transaction
	})

	if err != nil {
		return nil, err // Error from executeTransaction or the callback
	}

	return &newExercisePlan, nil

}

func (r *postgresEPRepository) GetExercisePlanById(ctx context.Context, id int) (*ExercisePlan, error) {
	var exercisePlan ExercisePlan

	query := `SELECT 
        id, 
		exercise_id, 
		workout_plan_id,
		sets,
		repetitions,
		weights,
		weight_unit FROM exercise_plans WHERE id = $1`

	row, err := executeQueryRow(ctx, r.db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query exercise plan by id '%v': %w", id, err)
	}

	err = row.Scan(
		&exercisePlan.Id,
		&exercisePlan.ExerciseId,
		&exercisePlan.WorkoutPlanId,
		&exercisePlan.Sets,
		&exercisePlan.Repetitions,
		&exercisePlan.Weights,
		&exercisePlan.WeightUnit)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to scan returned exercise plan data by id '%v': %w", id, err)
	}

	return &exercisePlan, nil

}

func (r *postgresEPRepository) UpdateExercisePlan(ctx context.Context, data UpdateEP) (*ExercisePlan, error) {
	var outsideUpdatedEP *ExercisePlan
	err := executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {
		var currentID int
		err := tx.QueryRowContext(txCtx, "SELECT id FROM exercise_plans WHERE id = $1", data.Id).Scan(&currentID)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrNotFound
			}

			return fmt.Errorf("failed to query exercise plan by id '%v' for update: %w", data.Id, err)
		}

		var updatedEP ExercisePlan
		query := `UPDATE exercise_plans
					SET sets = COALESCE($1, sets),
						repetitions = COALESCE($2, repetitions),
						weights = COALESCE($3, weights),
						weight_unit = COALESCE($4, weight_unit)
						WHERE id = $5
						RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`
		err = tx.QueryRowContext(txCtx,
			query,
			data.Sets,
			data.Repetitions,
			data.Weights,
			data.WeightUnit,
			data.Id).Scan(
			&updatedEP.Id,
			&updatedEP.ExerciseId,
			&updatedEP.WorkoutPlanId,
			&updatedEP.Sets,
			&updatedEP.Repetitions,
			&updatedEP.Weights,
			&updatedEP.WeightUnit,
		)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrNotFound
			}
			return fmt.Errorf("failed to update and scan exercise plan with id '%v': %w", data.Id, err)
		}
		outsideUpdatedEP = &updatedEP
		return nil
	})

	if err != nil {
		return nil, err
	}

	return outsideUpdatedEP, nil
}

func (r *postgresEPRepository) DeleteExercisePlanByID(ctx context.Context, id int) error {
	deleteQuery := `DELETE FROM exercise_plans WHERE id = $1`

	result, err := executeNonQuery(ctx, r.db, deleteQuery, id) // Use executeNonQuery
	if err != nil {
		return fmt.Errorf("failed to delete exercise plan with id '%v': %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting exercise plan with id '%v': %w", id, err)
	}

	if rowsAffected == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
func (r *postgresEPRepository) ListExercisePlans(ctx context.Context, workoutId int) ([]ExercisePlan, error) {
	query := `SELECT 
		id,
		exercise_id, 
		workout_plan_id,
		sets,
		repetitions,
		weights,
		weight_unit
	FROM exercise_plans WHERE workout_plan_id = $1`

	rows, err := executeQuery(ctx, r.db, query, workoutId)

	if err != nil {
		return nil, fmt.Errorf("failed to query exercise plans for workout plan id '%v': %w", workoutId, err)
	}
	defer rows.Close()

	var epsList []ExercisePlan
	for rows.Next() {
		var ep ExercisePlan
		if err := rows.Scan(
			&ep.Id,
			&ep.ExerciseId,
			&ep.WorkoutPlanId,
			&ep.Sets,
			&ep.Repetitions,
			&ep.Weights,
			&ep.WeightUnit); err != nil {
			return nil, fmt.Errorf("failed to scan exercise plan row: %w", err)
		}
		epsList = append(epsList, ep)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating exercise plan rows: %w", err)
	}

	return epsList, nil
}
