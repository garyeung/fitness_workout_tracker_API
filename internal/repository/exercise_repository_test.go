package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
)

func TestCreateExercise(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	exerciseRepo := repository.NewExerRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		newExer := repository.CreateExer{
			Name:        "Push-up",
			Description: "Bodyweight exercise",
			MuscleGroup: repository.Chest,
		}
		expectedID := 1

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercises (name, description, muscle_group) VALUES ($1, $2, $3) RETURNING id, name, description, muscle_group`,
		)).
			WithArgs(newExer.Name, newExer.Description, newExer.MuscleGroup).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "muscle_group"}).
				AddRow(expectedID, newExer.Name, newExer.Description, newExer.MuscleGroup))
		mock.ExpectCommit()

		exercise, err := exerciseRepo.CreateExercise(ctx, newExer)
		assert.NoError(t, err)
		assert.NotNil(t, exercise)
		assert.Equal(t, expectedID, exercise.Id)
		assert.Equal(t, newExer.Name, exercise.Name)
		assert.Equal(t, newExer.Description, exercise.Description)
		assert.Equal(t, newExer.MuscleGroup, exercise.MuscleGroup)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during insert", func(t *testing.T) {
		newExer := repository.CreateExer{
			Name:        "Push-up",
			Description: "Bodyweight exercise",
			MuscleGroup: repository.Chest,
		}
		dbError := errors.New("database connection lost")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercises (name, description, muscle_group) VALUES ($1, $2, $3) RETURNING id, name, description, muscle_group`,
		)).
			WithArgs(newExer.Name, newExer.Description, newExer.MuscleGroup).
			WillReturnError(dbError)
		mock.ExpectRollback() // Expect rollback on query error

		exercise, err := exerciseRepo.CreateExercise(ctx, newExer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert and scan new exercise")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercise)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during begin transaction", func(t *testing.T) {
		newExer := repository.CreateExer{
			Name:        "Push-up",
			Description: "Bodyweight exercise",
			MuscleGroup: repository.Chest,
		}
		beginErr := errors.New("failed to begin transaction")

		mock.ExpectBegin().WillReturnError(beginErr)

		exercise, err := exerciseRepo.CreateExercise(ctx, newExer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.Contains(t, err.Error(), beginErr.Error())
		assert.Nil(t, exercise)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during commit transaction", func(t *testing.T) {
		newExer := repository.CreateExer{
			Name:        "Push-up",
			Description: "Bodyweight exercise",
			MuscleGroup: repository.Chest,
		}
		commitErr := errors.New("transaction commit failed")
		expectedID := 1

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercises (name, description, muscle_group) VALUES ($1, $2, $3) RETURNING id, name, description, muscle_group`,
		)).
			WithArgs(newExer.Name, newExer.Description, newExer.MuscleGroup).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "muscle_group"}).
				AddRow(expectedID, newExer.Name, newExer.Description, newExer.MuscleGroup))
		mock.ExpectCommit().WillReturnError(commitErr)

		exercise, err := exerciseRepo.CreateExercise(ctx, newExer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.Contains(t, err.Error(), commitErr.Error())
		assert.Nil(t, exercise) // Should be nil because commit failed

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetExerciseById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	exerciseRepo := repository.NewExerRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		exerciseID := 1
		expectedExercise := repository.Exercise{
			Id:          exerciseID,
			Name:        "Squat",
			Description: "Lower body exercise",
			MuscleGroup: repository.Legs,
		}

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(exerciseID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "muscle_group"}).
				AddRow(expectedExercise.Id, expectedExercise.Name, expectedExercise.Description, expectedExercise.MuscleGroup))

		exercise, err := exerciseRepo.GetExerciseById(ctx, exerciseID)
		assert.NoError(t, err)
		assert.NotNil(t, exercise)
		assert.Equal(t, expectedExercise, *exercise)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		exerciseID := 99
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(exerciseID).
			WillReturnError(sql.ErrNoRows) // Simulate no rows found

		exercise, err := exerciseRepo.GetExerciseById(ctx, exerciseID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.Nil(t, exercise)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		exerciseID := 1
		dbError := errors.New("database connection lost")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises WHERE id = $1`,
		)).
			WillReturnError(dbError)

		exercise, err := exerciseRepo.GetExerciseById(ctx, exerciseID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query exercise by id")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercise)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestListExercises(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	exerciseRepo := repository.NewExerRepository(db)
	ctx := context.Background()

	t.Run("success with multiple exercises", func(t *testing.T) {
		expectedExercises := []repository.Exercise{
			{Id: 1, Name: "Squat", Description: "Lower body", MuscleGroup: repository.Legs},
			{Id: 2, Name: "Bench Press", Description: "Upper body", MuscleGroup: repository.Chest},
		}

		rows := sqlmock.NewRows([]string{"id", "name", "description", "muscle_group"}).
			AddRow(expectedExercises[0].Id, expectedExercises[0].Name, expectedExercises[0].Description, expectedExercises[0].MuscleGroup).
			AddRow(expectedExercises[1].Id, expectedExercises[1].Name, expectedExercises[1].Description, expectedExercises[1].MuscleGroup)

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises`,
		)).
			ExpectQuery().
			WillReturnRows(rows)

		exercises, err := exerciseRepo.ListExercises(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, exercises)
		assert.Len(t, exercises, 2)
		assert.Equal(t, expectedExercises, exercises)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no exercises", func(t *testing.T) {
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises`,
		)).
			ExpectQuery().
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "description", "muscle_group"})) // No rows

		exercises, err := exerciseRepo.ListExercises(ctx)
		assert.NoError(t, err)
		assert.Empty(t, exercises)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		dbError := errors.New("network error")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, name, description, muscle_group FROM exercises`,
		)).
			ExpectQuery().
			WillReturnError(dbError)

		exercises, err := exerciseRepo.ListExercises(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query all exercises")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercises)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}

func TestDeleteExercise(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	exerciseRepo := repository.NewExerRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		exerciseID := 1
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercises WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(exerciseID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

		err := exerciseRepo.DeleteExercise(ctx, exerciseID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		exerciseID := 99
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercises WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(exerciseID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		err := exerciseRepo.DeleteExercise(ctx, exerciseID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		exerciseID := 1
		dbError := errors.New("database down")
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercises WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(exerciseID).
			WillReturnError(dbError)

		err := exerciseRepo.DeleteExercise(ctx, exerciseID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete exercise with id")
		assert.Contains(t, err.Error(), dbError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		exerciseID := 1
		rowsAffectedError := errors.New("cannot get rows affected")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercises WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(exerciseID).
			WillReturnResult(sqlmock.NewErrorResult(rowsAffectedError))

		err := exerciseRepo.DeleteExercise(ctx, exerciseID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected after deleting exercise with id")
		assert.Contains(t, err.Error(), rowsAffectedError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
