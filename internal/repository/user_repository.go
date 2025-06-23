package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"workout-tracker-api/internal/apperrors"

	"github.com/lib/pq"
)

type User struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserCreate struct {
	Name         string `json:"name"`
	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}

type UserRepository interface {
	CreateUser(ctx context.Context, user UserCreate) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	DeleteUserByEmail(ctx context.Context, email string) error
	ExistUser(ctx context.Context, email string) (bool, error)
	// ... other user-related methods
}

// postgresUserRepository implements UserRepository
type postgresUserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{
		db: db,
	}
}

func (r *postgresUserRepository) CreateUser(ctx context.Context, data UserCreate) (*User, error) {

	// insert into users table, second
	query := `INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, name, email, password_hash, created_at, updated_at`

	row, err := executeQueryRow(ctx, r.db, query, data.Name, data.Email, data.PasswordHash)
	if err != nil {
		return nil, fmt.Errorf("failed to insert query for creating new user: %w", err)
	}

	// Scan the result into a User struct, third
	var newUser User
	err = row.Scan(&newUser.Id, &newUser.Name, &newUser.Email, &newUser.PasswordHash, &newUser.CreatedAt, &newUser.UpdatedAt)
	if err != nil {
		var pqErr *pq.Error
		// Check if the error is a PostgreSQL error
		if errors.As(err, &pqErr) {
			// SQLSTATE 23505 is the code for unique_violation
			if pqErr.Code == "23505" {
				return nil, fmt.Errorf("failed to create user: %w", apperrors.ErrAlreadyExists)
			}
		}
		return nil, fmt.Errorf("failed to scan returned new user data: %w", err)
	}

	return &newUser, nil
}

func (r *postgresUserRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	query := `SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = $1`

	row, err := executeQueryRow(ctx, r.db, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query for user: %w", err)
	}

	err = row.Scan(&user.Id, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperrors.ErrNotFound
		}

		return nil, fmt.Errorf("failed to scan returned user data by email '%v': %w", email, err)
	}
	return &user, nil

}

func (r *postgresUserRepository) DeleteUserByEmail(ctx context.Context, email string) error {
	return executeTransaction(ctx, r.db, func(txCtx context.Context, tx *sql.Tx) error {
		var userID int
		err := tx.QueryRowContext(txCtx, "SELECT id FROM users WHERE email = $1", email).Scan(&userID)

		if err != nil {
			if err == sql.ErrNoRows {
				return apperrors.ErrNotFound
			}

			return fmt.Errorf("failed to scan user data by email '%v' for deletion: %w", email, err)
		}

		// delete all exercise plans first associated with the user's workout plans
		deleteExercisePlansQuery := ` 
	DELETE FROM exercise_plans
	WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1)
	`
		_, err = tx.ExecContext(txCtx, deleteExercisePlansQuery, userID)
		if err != nil {
			return fmt.Errorf("failed to delete exercise plans for user ID %d: %w", userID, err)
		}

		// delete all workout plans second associated with the user
		deleteWorkoutPlansQuery := `
		DELETE FROM workout_plans WHERE user_id = $1
	`

		_, err = tx.ExecContext(txCtx, deleteWorkoutPlansQuery, userID)

		if err != nil {
			return fmt.Errorf("failed to delete workout plans for user ID %d: %w", userID, err)
		}

		// delete the user finally

		deleteUserQuery := `
		DELETE FROM users WHERE id = $1
	`

		result, err := tx.ExecContext(txCtx, deleteUserQuery, userID)

		if err != nil {
			return fmt.Errorf("failed to delete user with ID %d: %w", userID, err)
		}

		// ... after tx.Exec(deleteUserQuery, userID)
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected after deleting user with ID %d: %w", userID, err)
		}
		if rowsAffected == 0 {
			// This case should ideally not be reached if the initial SELECT found the user,
			// but it's a good safeguard or indicates an unexpected state.
			return fmt.Errorf("user with ID %d not found for final deletion step: %w", userID, sql.ErrNoRows)
		}

		return nil

	})

}

func (r *postgresUserRepository) ExistUser(ctx context.Context, email string) (bool, error) {
	query := `SELECT id FROM users WHERE email = $1`

	row, err := executeQueryRow(ctx, r.db, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to execute query for user: %w", err)
	}

	var ID int
	err = row.Scan(&ID)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, fmt.Errorf("failed to scan returned user data by email '%v': %w", email, err)
	}
	return true, nil

}
