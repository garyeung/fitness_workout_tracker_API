package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)

	ctx := context.Background()

	// Test case 1: Successful user creation
	t.Run("success", func(t *testing.T) {
		userToCreate := repository.UserCreate{
			Name:         "Test User",
			Email:        "test@example.com",
			PasswordHash: "hashedpassword123",
		}

		// Expect both Prepare and QueryRow calls
		mock.ExpectPrepare(`INSERT INTO users \(name, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, name, email, password_hash, created_at, updated_at`).
			ExpectQuery(). // This expects the QueryRowContext call after preparation
			WithArgs(userToCreate.Name, userToCreate.Email, userToCreate.PasswordHash).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"}).
				AddRow(1, userToCreate.Name, userToCreate.Email, userToCreate.PasswordHash, time.Now(), time.Now()))

		createdUser, err := userRepo.CreateUser(ctx, userToCreate)

		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Equal(t, 1, createdUser.Id)
		assert.Equal(t, userToCreate.Email, createdUser.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: Duplicate email (simulating unique constraint violation)
	t.Run("duplicate email", func(t *testing.T) {
		userToCreate := repository.UserCreate{
			Name:         "Another User",
			Email:        "duplicate@example.com",
			PasswordHash: "hashedpassword456",
		}

		mockedPQError := &pq.Error{
			Code:     "23505", // SQLSTATE for unique_violation
			Severity: "ERROR",
			Message:  "duplicate key value violates unique constraint \"users_email_key\"",
			Detail:   "Key (email)=(duplicate@example.com) already exists.",
			Where:    "SQL statement \"INSERT INTO users ...\"",
		}
		mock.ExpectPrepare(`INSERT INTO users \(name, email, password_hash\) VALUES \(\$1, \$2, \$3\) RETURNING id, name, email, password_hash, created_at, updated_at`).
			ExpectQuery().
			WithArgs(userToCreate.Name, userToCreate.Email, userToCreate.PasswordHash).
			WillReturnError(mockedPQError)

		createdUser, err := userRepo.CreateUser(ctx, userToCreate)

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, errors.Is(err, apperrors.ErrAlreadyExists))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: Generic database error
	t.Run("database error", func(t *testing.T) {
		userToCreate := repository.UserCreate{
			Name:         "Error User",
			Email:        "error@example.com",
			PasswordHash: "hashedpassword789",
		}

		mock.ExpectPrepare(`INSERT INTO users`).
			ExpectQuery().
			WithArgs(userToCreate.Name, userToCreate.Email, userToCreate.PasswordHash).
			WillReturnError(errors.New("something went wrong with the database"))

		createdUser, err := userRepo.CreateUser(ctx, userToCreate)

		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.False(t, errors.Is(err, apperrors.ErrAlreadyExists))
		assert.False(t, errors.Is(err, apperrors.ErrNotFound))

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// The TestGetUserByEmail function should be fine as it already expects a Query before.
// No changes needed for TestGetUserByEmail if it already uses ExpectQuery.
// (Based on the previous successful run, it likely was already correct or didn't hit a Prepare issue).
func TestGetUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test case 1: User found
	t.Run("success", func(t *testing.T) {
		email := "findme@example.com"
		expectedUser := &repository.User{
			Id:           101,
			Name:         "Found User",
			Email:        email,
			PasswordHash: "abc",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		// This one might also need ExpectPrepare if executeQueryRow is used.
		// Let's assume it does, as your other helpers Prepare.
		mock.ExpectPrepare(`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
			ExpectQuery(). // Add this
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"}).
				AddRow(expectedUser.Id, expectedUser.Name, expectedUser.Email, expectedUser.PasswordHash, expectedUser.CreatedAt, expectedUser.UpdatedAt))

		user, err := userRepo.GetUserByEmail(ctx, email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: User not found
	t.Run("not found", func(t *testing.T) {
		email := "notfound@example.com"

		mock.ExpectPrepare(`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
			ExpectQuery(). // Add this
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		user, err := userRepo.GetUserByEmail(ctx, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: Generic database error
	t.Run("database error", func(t *testing.T) {
		email := "dberror@example.com"

		mock.ExpectPrepare(`SELECT id, name, email, password_hash, created_at, updated_at FROM users WHERE email = \$1`).
			ExpectQuery(). // Add this
			WithArgs(email).
			WillReturnError(errors.New("connection reset by peer"))

		user, err := userRepo.GetUserByEmail(ctx, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.False(t, errors.Is(err, apperrors.ErrNotFound))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestExistUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Test case 1: User exists
	t.Run("user exists", func(t *testing.T) {
		email := "existing@example.com"
		expectedUserID := 123

		// Expect prepare and query for the SELECT id FROM users
		mock.ExpectPrepare(`SELECT id FROM users WHERE email = \$1`).
			ExpectQuery().
			WithArgs(email).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(expectedUserID))

		exists, err := userRepo.ExistUser(ctx, email)

		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: User does not exist
	t.Run("user does not exist", func(t *testing.T) {
		email := "nonexistent@example.com"

		// Expect prepare and query, returning sql.ErrNoRows
		mock.ExpectPrepare(`SELECT id FROM users WHERE email = \$1`).
			ExpectQuery().
			WithArgs(email).
			WillReturnError(sql.ErrNoRows)

		exists, err := userRepo.ExistUser(ctx, email)

		assert.NoError(t, err) // ExistUser should return nil error if user not found, only false
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: Database error
	t.Run("database error", func(t *testing.T) {
		email := "error@example.com"
		expectedErr := errors.New("connection refused")

		// Expect prepare and query, returning a generic error
		mock.ExpectPrepare(`SELECT id FROM users WHERE email = \$1`).
			WillReturnError(expectedErr)

		exists, err := userRepo.ExistUser(ctx, email)

		assert.Error(t, err)
		assert.False(t, exists)
		// Check that the error is wrapped, not just the raw expectedErr
		assert.Contains(t, err.Error(), "failed to execute query for user")
		assert.True(t, errors.Is(err, expectedErr))
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteUserByEmail(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()
	userEmail := "deleteme@example.com"
	userID := 123

	// Test case 1: Successful deletion
	t.Run("success", func(t *testing.T) {
		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email (direct query on tx)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		// 3. Expect DELETE exercise plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1)`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// 4. Expect DELETE workout plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 2)) // 2 workout plans affected

		// 5. Expect DELETE user
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 user affected

		// 6. Expect CommitTx
		mock.ExpectCommit()

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 2: User not found (by SELECT id FROM users WHERE email)
	t.Run("user not found", func(t *testing.T) { // Renamed for clarity
		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email, returning sql.ErrNoRows
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnError(sql.ErrNoRows) // User does not exist

		// 3. Expect Rollback
		mock.ExpectRollback()

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 3: User found, but no rows affected by final user DELETE
	t.Run("user found but not deleted", func(t *testing.T) {
		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email (user exists)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		// 3. Expect DELETE exercise plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1)`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		// 4. Expect DELETE workout plans (succeeds)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 5. Expect DELETE user, but 0 rows affected
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // No rows affected

		// 6. Expect Rollback
		mock.ExpectRollback()

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 4: Database error during workout plan deletion
	t.Run("db error deleting workout plans", func(t *testing.T) {
		dbErr := errors.New("failed to connect to DB during workout plan delete")

		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email (user exists)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1)`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 2))

		mock.ExpectExec(regexp.QuoteMeta(`
		DELETE FROM workout_plans WHERE user_id = $1
`)).
			WithArgs(userID).
			WillReturnError(dbErr)

		// 4. Expect Rollback
		mock.ExpectRollback()

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, dbErr))
		assert.Contains(t, err.Error(), fmt.Sprintf("failed to delete workout plans for user ID %d", userID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 5: Database error during user deletion
	t.Run("db error deleting user", func(t *testing.T) {
		dbErr := errors.New("disk full")

		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email (user exists)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		// 3. Expect DELETE exercise plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 4. Expect DELETE user to return an error
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnError(dbErr)

		// 5. Expect Rollback
		mock.ExpectRollback()

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, dbErr))
		assert.Contains(t, err.Error(), fmt.Sprintf("failed to delete user with ID %d", userID))
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 6: Error during BeginTx
	t.Run("error during begin transaction", func(t *testing.T) {
		beginErr := errors.New("failed to open transaction connection")
		mock.ExpectBegin().WillReturnError(beginErr)

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, beginErr))
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	// Test case 7: Error during CommitTx
	t.Run("error during commit transaction", func(t *testing.T) {
		commitErr := errors.New("transaction commit failed")

		// 1. Expect BeginTx
		mock.ExpectBegin()

		// 2. Expect SELECT id FROM users WHERE email (user exists)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT id FROM users WHERE email = $1`)).
			WithArgs(userEmail).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(userID))

		// 3. Expect DELETE exercise plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id IN (SELECT id FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 4. Expect DELETE workout plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE user_id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 5. Expect DELETE user (succeeds)
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM users WHERE id = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// 6. Expect CommitTx to return an error
		mock.ExpectCommit().WillReturnError(commitErr)

		// 7. Expect Rollback (due to the deferred rollback in executeTransaction, after commit fails)

		err := userRepo.DeleteUserByEmail(ctx, userEmail)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, commitErr))
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
