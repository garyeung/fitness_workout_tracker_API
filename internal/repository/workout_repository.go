package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"workout-tracker-api/internal/apperrors"
)

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
	Comment       sql.NullString `json:"comment"` // could not set
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
}

type CreateWP struct {
	UserId        int       `json:"userId"`
	ScheduledDate time.Time `json:"scheduledDate"`
	Comment       *string   `json:"comment,omitempty"`
}

type UpdateWP struct {
	Id            int        `json:"id"`
	Status        WPStatus   `json:"status"`
	ScheduledDate *time.Time `json:"scheduledDate,omitempty"`
	Comment       *string    `json:"comment,omitempty"` // poniter for optional update
}

type WorkoutRepository interface {
	CreateWorkout(ctx context.Context, data CreateWP) (*WorkoutPlan, error)
	GetWorkoutById(ctx context.Context, id int) (*WorkoutPlan, error)
	UpdateWorkout(ctx context.Context, data UpdateWP) (*WorkoutPlan, error)
	DeleteWorkoutById(ctx context.Context, id int) error
	ListWorkoutsByStatus(ctx context.Context, userId int, status WPStatus, asc bool) ([]WorkoutPlan, error)
	ListUserWorkouts(ctx context.Context, userId int) ([]WorkoutPlan, error)
}

type postgresWorkoutRepository struct {
	db *sql.DB
}

func NewWorkoutRepository(db *sql.DB) WorkoutRepository {
	return &postgresWorkoutRepository{
		db: db,
	}
}

func (r *postgresWorkoutRepository) CreateWorkout(ctx context.Context, data CreateWP) (*WorkoutPlan, error) {
	var status = PENDING
	if data.Comment != nil && *data.Comment == "" {
		data.Comment = nil
	}

	query := `INSERT INTO workout_plans (
	user_id, 
	scheduled_date, 
	status, 
	comment) VALUES ($1, $2, $3, $4)
	RETURNING
	id,
	user_id,
	status,
	scheduled_date,
	comment,
	created_at,
	updated_at`

	row, err := executeQueryRow(ctx, r.db,
		query,
		data.UserId,
		data.ScheduledDate,
		status,
		data.Comment)

	if err != nil {
		return nil, fmt.Errorf("failed to execute insert query for creating new workout plan: %w", err)
	}

	var newWP WorkoutPlan
	err = row.Scan(
		&newWP.Id,
		&newWP.UserId,
		&newWP.Status,
		&newWP.ScheduledDate,
		&newWP.Comment,
		&newWP.CreatedAt,
		&newWP.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan returned new workout plan: %w", err)
	}

	return &newWP, nil

}

func (r *postgresWorkoutRepository) GetWorkoutById(ctx context.Context, id int) (*WorkoutPlan, error) {
	var workoutPlan WorkoutPlan
	query := `SELECT
	id,
	user_id,
	status,
	scheduled_date,
	comment,
	created_at, 
	updated_at FROM workout_plans WHERE id = $1`

	row, err := executeQueryRow(ctx, r.db, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query workout plan by id '%v': %w", id, err)
	}

	err = row.Scan(
		&workoutPlan.Id,
		&workoutPlan.UserId,
		&workoutPlan.Status,
		&workoutPlan.ScheduledDate,
		&workoutPlan.Comment,
		&workoutPlan.CreatedAt,
		&workoutPlan.UpdatedAt)
	if err != nil {

		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to scan returned workout plan by id '%v': %w", id, err)
	}

	return &workoutPlan, nil
}

func (r *postgresWorkoutRepository) UpdateWorkout(ctx context.Context, data UpdateWP) (*WorkoutPlan, error) {
	var result *WorkoutPlan
	err := executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {
		var currentID int
		err := tx.QueryRowContext(txCtx, "SELECT id FROM workout_plans WHERE id = $1", data.Id).Scan(&currentID)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrNotFound
			}

			return fmt.Errorf("failed to query workout plan by id '%v' for update: %w", data.Id, err)
		}

		var updatedWP WorkoutPlan
		query := `UPDATE workout_plans
				SET status = COALESCE($1, status),
					scheduled_date = COALESCE($2, scheduled_date),
					comment = COALESCE($3, comment),
					updated_at = CURRENT_TIMESTAMP 
				WHERE id = $4 RETURNING
				id,
				user_id,
				status,
				scheduled_date,
				comment,
				created_at, 
				updated_at`

		err = tx.QueryRowContext(txCtx,
			query, // Use the corrected query
			data.Status,
			data.ScheduledDate,
			data.Comment,
			data.Id).Scan(
			&updatedWP.Id,
			&updatedWP.UserId,
			&updatedWP.Status,
			&updatedWP.ScheduledDate,
			&updatedWP.Comment,
			&updatedWP.CreatedAt,
			&updatedWP.UpdatedAt)
		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrNotFound
			}
			return fmt.Errorf("failed to update and scan workout plan with id '%v': %w", data.Id, err)
		}
		result = &updatedWP
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *postgresWorkoutRepository) DeleteWorkoutById(ctx context.Context, id int) error {
	return executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {

		deleteExercisePlansQuery := `DELETE FROM exercise_plans WHERE workout_plan_id = $1`

		_, err := tx.ExecContext(txCtx, deleteExercisePlansQuery, id)

		if err != nil {
			return fmt.Errorf("failed to delete exercise plans for workout plan id %d: %w", id, err)
		}

		deleteWorkoutPlanQuery := `DELETE FROM workout_plans WHERE id = $1`
		result, err := tx.ExecContext(txCtx, deleteWorkoutPlanQuery, id)

		if err != nil {
			return fmt.Errorf("failed to delete workout plan with id %d: %w", id, err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			// This error is about getting RowsAffected, not the delete operation itself.
			return fmt.Errorf("failed to get rows affected after deleting workout plan with id %d: %w", id, err)
		}

		if rowsAffected == 0 {
			// If no rows were affected, the workout plan with the given ID was not found.
			// Wrapping sql.ErrNoRows provides a conventional way to check for this specific error type upstream.
			return apperrors.ErrNotFound
		}

		return nil
	})
}
func (r *postgresWorkoutRepository) ListWorkoutsByStatus(ctx context.Context, userID int, status WPStatus, asc bool) ([]WorkoutPlan, error) {
	query := `SELECT
		id,
		user_id,
		status,
		scheduled_date,
		comment,
		created_at,
		updated_at 
	FROM workout_plans WHERE user_id = $1 AND status = $2 
	`
	if asc {
		query = fmt.Sprintf("%s ORDER BY scheduled_date ASC", query)
	} else {
		query = fmt.Sprintf("%s ORDER BY scheduled_date DESC", query)
	}
	rows, err := executeQuery(ctx, r.db, query, userID, status)

	if err != nil {
		return nil, fmt.Errorf("failed to query workout plan by status: %w", err)
	}

	defer rows.Close()

	var wpList []WorkoutPlan
	for rows.Next() {
		var wp WorkoutPlan
		if err := rows.Scan(
			&wp.Id,
			&wp.UserId,
			&wp.Status,
			&wp.ScheduledDate,
			&wp.Comment,
			&wp.CreatedAt,
			&wp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workout plan: %w", err)
		}
		wpList = append(wpList, wp)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workout plans: %w", err)
	}

	return wpList, nil
}

func (r *postgresWorkoutRepository) ListUserWorkouts(ctx context.Context, userID int) ([]WorkoutPlan, error) {
	query := `SELECT
		id,
		user_id,
		status,
		scheduled_date,
		comment,
		created_at,
		updated_at 
	FROM workout_plans WHERE user_id = $1 ORDER BY scheduled_date ASC 
	`
	rows, err := executeQuery(ctx, r.db, query, userID)

	if err != nil {
		return nil, fmt.Errorf("failed to query workout plan for user id '%v': %w", userID, err)
	}

	defer rows.Close()

	var wpList []WorkoutPlan
	for rows.Next() {
		var wp WorkoutPlan
		if err := rows.Scan(
			&wp.Id,
			&wp.UserId,
			&wp.Status,
			&wp.ScheduledDate,
			&wp.Comment,
			&wp.CreatedAt,
			&wp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workout plan row: %w", err)
		}
		wpList = append(wpList, wp)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workout plan rows: %w", err)
	}

	return wpList, nil

}
