package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
)

func TestCreateWorkout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		userID := 1
		scheduledDate := time.Now().Add(24 * time.Hour).Truncate(time.Second) // Truncate to match DB precision
		comment := "Leg day workout"
		newWP := repository.CreateWP{
			UserId:        userID,
			ScheduledDate: scheduledDate,
			Comment:       &comment,
		}
		expectedID := 1

		mock.ExpectPrepare(regexp.QuoteMeta(
			`INSERT INTO workout_plans (
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
	updated_at`,
		)).
			ExpectQuery().
			WithArgs(newWP.UserId, newWP.ScheduledDate, repository.PENDING, sql.NullString{String: *newWP.Comment, Valid: true}).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(expectedID, newWP.UserId, repository.PENDING, newWP.ScheduledDate, sql.NullString{String: *newWP.Comment, Valid: true}, time.Now(), time.Now()))

		workoutPlan, err := wpRepo.CreateWorkout(ctx, newWP)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlan)
		assert.Equal(t, expectedID, workoutPlan.Id)
		assert.Equal(t, newWP.UserId, workoutPlan.UserId)
		assert.Equal(t, repository.PENDING, workoutPlan.Status)
		assert.True(t, workoutPlan.ScheduledDate.Equal(newWP.ScheduledDate))
		assert.Equal(t, sql.NullString{String: *newWP.Comment, Valid: true}, workoutPlan.Comment)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with null comment", func(t *testing.T) {
		userID := 1
		scheduledDate := time.Now().Add(48 * time.Hour).Truncate(time.Second)
		newWP := repository.CreateWP{
			UserId:        userID,
			ScheduledDate: scheduledDate,
			Comment:       nil,
		}
		expectedID := 2

		mock.ExpectPrepare(regexp.QuoteMeta(
			`INSERT INTO workout_plans (
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
	updated_at`,
		)).
			ExpectQuery().
			WithArgs(newWP.UserId, newWP.ScheduledDate, repository.PENDING, sql.NullString{Valid: false}).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(expectedID, newWP.UserId, repository.PENDING, newWP.ScheduledDate, sql.NullString{Valid: false}, time.Now(), time.Now()))

		workoutPlan, err := wpRepo.CreateWorkout(ctx, newWP)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlan)
		assert.Equal(t, expectedID, workoutPlan.Id)
		assert.False(t, workoutPlan.Comment.Valid)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during insert", func(t *testing.T) {
		userID := 1
		scheduledDate := time.Now().Truncate(time.Second)
		newWP := repository.CreateWP{
			UserId:        userID,
			ScheduledDate: scheduledDate,
			Comment:       nil,
		}
		dbError := errors.New("database is down")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`INSERT INTO workout_plans (
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
	updated_at`,
		)).
			WillReturnError(dbError)

		workoutPlan, err := wpRepo.CreateWorkout(ctx, newWP)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute insert query for creating new workout plan")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestGetWorkoutById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		wpID := 1
		scheduledDate := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
		comment := sql.NullString{String: "My workout plan", Valid: true}
		expectedWP := repository.WorkoutPlan{
			Id:            wpID,
			UserId:        101,
			Status:        repository.PENDING,
			ScheduledDate: scheduledDate,
			Comment:       comment,
			CreatedAt:     time.Now().Truncate(time.Second),
			UpdatedAt:     time.Now().Truncate(time.Second),
		}

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(expectedWP.Id, expectedWP.UserId, expectedWP.Status, expectedWP.ScheduledDate, expectedWP.Comment, expectedWP.CreatedAt, expectedWP.UpdatedAt))

		workoutPlan, err := wpRepo.GetWorkoutById(ctx, wpID)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlan)
		assert.Equal(t, expectedWP, *workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		wpID := 99
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(wpID).
			WillReturnError(sql.ErrNoRows)

		workoutPlan, err := wpRepo.GetWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		wpID := 1
		dbError := errors.New("database connection lost")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE id = $1`,
		)).
			WillReturnError(dbError)

		workoutPlan, err := wpRepo.GetWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query workout plan by id")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateWorkout(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	t.Run("success with all fields updated", func(t *testing.T) {
		wpID := 1
		status := repository.COMPLETED
		scheduledDate := time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC)
		comment := "Completed with high intensity"

		updateData := repository.UpdateWP{
			Id:            wpID,
			Status:        status,
			ScheduledDate: &scheduledDate,
			Comment:       &comment,
		}

		expectedWP := repository.WorkoutPlan{
			Id:            wpID,
			UserId:        101, // Assuming these remain unchanged
			Status:        status,
			ScheduledDate: scheduledDate,
			Comment:       sql.NullString{String: comment, Valid: true},
			CreatedAt:     time.Now().Truncate(time.Second),
			UpdatedAt:     time.Now().Truncate(time.Second),
		}

		// Mock the initial SELECT to get existing values

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wpID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE workout_plans
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
				updated_at`,
		)).
			WithArgs(updateData.Status, *updateData.ScheduledDate, sql.NullString{String: *updateData.Comment, Valid: true}, updateData.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(expectedWP.Id, expectedWP.UserId, expectedWP.Status, expectedWP.ScheduledDate, expectedWP.Comment, expectedWP.CreatedAt, expectedWP.UpdatedAt))
		mock.ExpectCommit()

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlan)
		assert.Equal(t, expectedWP, *workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with partial fields updated (e.g., only status)", func(t *testing.T) {
		wpID := 1
		status := repository.MISSED
		// Original values for other fields
		originalScheduledDate := time.Date(2025, 6, 15, 10, 0, 0, 0, time.UTC)
		originalComment := sql.NullString{String: "original comment", Valid: true}
		originalCreatedAt := time.Now().Add(-24 * time.Hour).Truncate(time.Second)
		originalUpdatedAt := time.Now().Add(-12 * time.Hour).Truncate(time.Second)

		updateData := repository.UpdateWP{
			Id:     wpID,
			Status: status,
			// ScheduledDate, Comment are nil
		}

		expectedWP := repository.WorkoutPlan{
			Id:            wpID,
			UserId:        101,
			Status:        status,
			ScheduledDate: originalScheduledDate, // Should be original value
			Comment:       originalComment,       // Should be original value
			CreatedAt:     originalCreatedAt,
			UpdatedAt:     originalUpdatedAt,
		}

		// Mock the initial SELECT to get existing values

		// Now expect the update query
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wpID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE workout_plans
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
				updated_at`,
		)).
			WithArgs(updateData.Status, nil, nil, updateData.Id). // Corrected args to include original values for non-updated fields
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(expectedWP.Id, expectedWP.UserId, expectedWP.Status, expectedWP.ScheduledDate, expectedWP.Comment, expectedWP.CreatedAt, expectedWP.UpdatedAt))
		mock.ExpectCommit()

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlan)
		assert.Equal(t, expectedWP, *workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found during initial select", func(t *testing.T) {
		wpID := 99
		status := repository.COMPLETED
		updateData := repository.UpdateWP{
			Id:     wpID,
			Status: status,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnError(sql.ErrNoRows)

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during query and scan", func(t *testing.T) {
		wpID := 1
		status := repository.COMPLETED
		updateData := repository.UpdateWP{
			Id:     wpID,
			Status: status,
		}
		dbError := errors.New("db update error")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wpID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE workout_plans
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
				updated_at`,
		)).
			WithArgs(updateData.Status, nil, nil, updateData.Id). // Corrected args to include original values for non-updated fields
			WillReturnError(dbError)
		mock.ExpectRollback()

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update and scan workout plan with id")
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during update", func(t *testing.T) {
		wpID := 1
		status := repository.COMPLETED
		scheduledDate := time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC)
		comment := "Completed with high intensity"

		updateData := repository.UpdateWP{
			Id:            wpID,
			Status:        status,
			ScheduledDate: &scheduledDate,
			Comment:       &comment,
		}

		dbError := errors.New("database update failed")

		// Mock the initial SELECT to get existing values

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wpID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE workout_plans
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
				updated_at`,
		)).
			WithArgs(updateData.Status, *updateData.ScheduledDate, sql.NullString{String: *updateData.Comment, Valid: true}, updateData.Id).
			WillReturnError(dbError)
		mock.ExpectRollback()

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update and scan workout plan")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during begin transaction", func(t *testing.T) {
		wpID := 1
		status := repository.COMPLETED
		updateData := repository.UpdateWP{
			Id:     wpID,
			Status: status,
		}
		beginErr := errors.New("failed to begin transaction")

		// Mock the initial SELECT to get existing values
		mock.ExpectBegin().WillReturnError(beginErr)

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.Contains(t, err.Error(), beginErr.Error())
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during commit transaction", func(t *testing.T) {
		wpID := 1
		status := repository.COMPLETED
		scheduledDate := time.Date(2025, 6, 20, 12, 0, 0, 0, time.UTC)
		comment := "Completed with high intensity"

		updateData := repository.UpdateWP{
			Id:            wpID,
			Status:        status,
			ScheduledDate: &scheduledDate,
			Comment:       &comment,
		}
		commitErr := errors.New("transaction commit failed")

		// Mock the initial SELECT to get existing values
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT id FROM workout_plans WHERE id = $1")).
			WithArgs(wpID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(wpID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE workout_plans
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
				updated_at`,
		)).
			WithArgs(updateData.Status, *updateData.ScheduledDate, sql.NullString{String: *updateData.Comment, Valid: true}, updateData.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
				AddRow(wpID, 101, status, scheduledDate, sql.NullString{String: comment, Valid: true}, time.Now().Truncate(time.Second), time.Now().Truncate(time.Second)))
		mock.ExpectCommit().WillReturnError(commitErr)

		workoutPlan, err := wpRepo.UpdateWorkout(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.Contains(t, err.Error(), commitErr.Error())
		assert.Nil(t, workoutPlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteWorkoutById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		wpID := 1

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected in exercise_plans

		// Finally, expect the DELETE for workout_plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected in workout_plans

		mock.ExpectCommit()

		err := wpRepo.DeleteWorkoutById(ctx, wpID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found workout plan", func(t *testing.T) {
		wpID := 99
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 row affected in exercise_plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 row affected in workout_plans

		err := wpRepo.DeleteWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error deleting exercise plans", func(t *testing.T) {
		wpID := 1
		dbError := errors.New("error deleting exercise plans")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id = $1`)).
			WithArgs(wpID).
			WillReturnError(dbError)

		mock.ExpectRollback() // Expect rollback because of transaction error

		err := wpRepo.DeleteWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete exercise plans for workout plan id")
		assert.Contains(t, err.Error(), dbError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error deleting workout plan", func(t *testing.T) {
		wpID := 1
		dbError := errors.New("error deleting workout plan")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 row affected in exercise_plans

		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE id = $1`)).
			WithArgs(wpID).
			WillReturnError(dbError)

		mock.ExpectRollback() // Expect rollback because of transaction error

		err := wpRepo.DeleteWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete workout plan with id")
		assert.Contains(t, err.Error(), dbError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error for workout plan delete", func(t *testing.T) {
		wpID := 1
		rowsAffectedError := errors.New("cannot get rows affected")

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM exercise_plans WHERE workout_plan_id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 0 row affected in exercise_plans
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM workout_plans WHERE id = $1`)).
			WithArgs(wpID).
			WillReturnResult(sqlmock.NewErrorResult(rowsAffectedError))

		mock.ExpectRollback()

		err := wpRepo.DeleteWorkoutById(ctx, wpID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected after deleting workout plan with id")
		assert.Contains(t, err.Error(), rowsAffectedError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestListUserWorkouts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	userID := 10

	t.Run("success with multiple workout plans", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		expectedWPs := []repository.WorkoutPlan{
			{Id: 1, UserId: userID, Status: repository.PENDING, ScheduledDate: now.Add(24 * time.Hour), Comment: sql.NullString{String: "Morning run", Valid: true}, CreatedAt: now, UpdatedAt: now},
			{Id: 2, UserId: userID, Status: repository.COMPLETED, ScheduledDate: now.Add(48 * time.Hour), Comment: sql.NullString{Valid: false}, CreatedAt: now, UpdatedAt: now},
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
			AddRow(expectedWPs[0].Id, expectedWPs[0].UserId, expectedWPs[0].Status, expectedWPs[0].ScheduledDate, expectedWPs[0].Comment, expectedWPs[0].CreatedAt, expectedWPs[0].UpdatedAt).
			AddRow(expectedWPs[1].Id, expectedWPs[1].UserId, expectedWPs[1].Status, expectedWPs[1].ScheduledDate, expectedWPs[1].Comment, expectedWPs[1].CreatedAt, expectedWPs[1].UpdatedAt)

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 ORDER BY scheduled_date ASC`,
		)).
			ExpectQuery().
			WithArgs(userID).
			WillReturnRows(rows)

		workoutPlans, err := wpRepo.ListUserWorkouts(ctx, userID)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlans)
		assert.Len(t, workoutPlans, 2)
		assert.Equal(t, expectedWPs, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no workout plans", func(t *testing.T) {
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 ORDER BY scheduled_date ASC`,
		)).
			ExpectQuery().
			WithArgs(userID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"})) // No rows

		workoutPlans, err := wpRepo.ListUserWorkouts(ctx, userID)
		assert.NoError(t, err)
		assert.Empty(t, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		dbError := errors.New("network error")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 ORDER BY scheduled_date ASC`,
		)).
			ExpectQuery().
			WithArgs(userID).
			WillReturnError(dbError)

		workoutPlans, err := wpRepo.ListUserWorkouts(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query workout plan for user id")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestListWorkoutsByStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	wpRepo := repository.NewWorkoutRepository(db)
	ctx := context.Background()

	userID := 10
	status := repository.PENDING

	t.Run("success with multiple workout plans", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		expectedWPs := []repository.WorkoutPlan{
			{Id: 1, UserId: userID, Status: status, ScheduledDate: now.Add(24 * time.Hour), Comment: sql.NullString{String: "Morning run", Valid: true}, CreatedAt: now, UpdatedAt: now},
			{Id: 2, UserId: userID, Status: status, ScheduledDate: now.Add(48 * time.Hour), Comment: sql.NullString{Valid: false}, CreatedAt: now, UpdatedAt: now},
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"}).
			AddRow(expectedWPs[0].Id, expectedWPs[0].UserId, expectedWPs[0].Status, expectedWPs[0].ScheduledDate, expectedWPs[0].Comment, expectedWPs[0].CreatedAt, expectedWPs[0].UpdatedAt).
			AddRow(expectedWPs[1].Id, expectedWPs[1].UserId, expectedWPs[1].Status, expectedWPs[1].ScheduledDate, expectedWPs[1].Comment, expectedWPs[1].CreatedAt, expectedWPs[1].UpdatedAt)

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 AND status = $2 ORDER BY scheduled_date ASC`,
		)).
			ExpectQuery().
			WithArgs(userID, status).
			WillReturnRows(rows)

		workoutPlans, err := wpRepo.ListWorkoutsByStatus(ctx, userID, status, true)
		assert.NoError(t, err)
		assert.NotNil(t, workoutPlans)
		assert.Len(t, workoutPlans, 2)
		assert.Equal(t, expectedWPs, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no workout plans", func(t *testing.T) {
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 AND status = $2 ORDER BY scheduled_date DESC`,
		)).
			ExpectQuery().
			WithArgs(userID, status).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "status", "scheduled_date", "comment", "created_at", "updated_at"})) // No rows

		workoutPlans, err := wpRepo.ListWorkoutsByStatus(ctx, userID, status, false)
		assert.NoError(t, err)
		assert.Empty(t, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		dbError := errors.New("network error")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, user_id, status, scheduled_date, comment, created_at, updated_at FROM workout_plans WHERE user_id = $1 AND status = $2 ORDER BY scheduled_date ASC`,
		)).
			WillReturnError(dbError)

		workoutPlans, err := wpRepo.ListWorkoutsByStatus(ctx, userID, status, true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query workout plan by status")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, workoutPlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
