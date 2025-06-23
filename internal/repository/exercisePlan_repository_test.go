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

func TestCreateExercisePlan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	epRepo := repository.NewEPRepository(db)
	ctx := context.Background()

	workoutPlanID := 10

	t.Run("success", func(t *testing.T) {
		newEP := repository.CreateEP{
			ExerciseId:  1,
			Sets:        3,
			Repetitions: 10,
			Weights:     50.0,
			WeightUnit:  repository.KG,
		}
		expectedID := 1

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id FROM workout_plans WHERE id = $1`)).
			WithArgs(workoutPlanID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(workoutPlanID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercise_plans ( exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).
			WithArgs(newEP.ExerciseId, workoutPlanID, newEP.Sets, newEP.Repetitions, newEP.Weights, newEP.WeightUnit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(expectedID, newEP.ExerciseId, workoutPlanID, newEP.Sets, newEP.Repetitions, newEP.Weights, newEP.WeightUnit))
		mock.ExpectCommit()

		exercisePlan, err := epRepo.CreateExercisePlan(ctx, newEP, workoutPlanID)
		assert.NoError(t, err)
		assert.NotNil(t, exercisePlan)
		assert.Equal(t, expectedID, exercisePlan.Id)
		assert.Equal(t, newEP.ExerciseId, exercisePlan.ExerciseId)
		assert.Equal(t, workoutPlanID, exercisePlan.WorkoutPlanId)
		assert.Equal(t, newEP.Sets, exercisePlan.Sets)
		assert.Equal(t, newEP.Repetitions, exercisePlan.Repetitions)
		assert.Equal(t, newEP.Weights, exercisePlan.Weights)
		assert.Equal(t, newEP.WeightUnit, exercisePlan.WeightUnit)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during insert", func(t *testing.T) {
		newEP := repository.CreateEP{
			ExerciseId:  1,
			Sets:        3,
			Repetitions: 10,
			Weights:     50.0,
			WeightUnit:  repository.KG,
		}
		dbError := errors.New("database connection lost")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id FROM workout_plans WHERE id = $1`)).
			WithArgs(workoutPlanID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(workoutPlanID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercise_plans ( exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).
			WithArgs(newEP.ExerciseId, workoutPlanID, newEP.Sets, newEP.Repetitions, newEP.Weights, newEP.WeightUnit).
			WillReturnError(dbError)
		mock.ExpectRollback()

		exercisePlan, err := epRepo.CreateExercisePlan(ctx, newEP, workoutPlanID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to insert and scan new exercise plan")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during begin transaction", func(t *testing.T) {
		newEP := repository.CreateEP{
			ExerciseId:  1,
			Sets:        3,
			Repetitions: 10,
			Weights:     50.0,
			WeightUnit:  repository.KG,
		}
		beginErr := errors.New("failed to begin transaction")

		mock.ExpectBegin().WillReturnError(beginErr)

		exercisePlan, err := epRepo.CreateExercisePlan(ctx, newEP, workoutPlanID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.Contains(t, err.Error(), beginErr.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during commit transaction", func(t *testing.T) {
		newEP := repository.CreateEP{
			ExerciseId:  1,
			Sets:        3,
			Repetitions: 10,
			Weights:     50.0,
			WeightUnit:  repository.KG,
		}
		commitErr := errors.New("transaction commit failed")
		expectedID := 1

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(
			`SELECT id FROM workout_plans WHERE id = $1`)).
			WithArgs(workoutPlanID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(workoutPlanID))

		mock.ExpectQuery(regexp.QuoteMeta(
			`INSERT INTO exercise_plans ( exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit ) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).
			WithArgs(newEP.ExerciseId, workoutPlanID, newEP.Sets, newEP.Repetitions, newEP.Weights, newEP.WeightUnit).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(expectedID, newEP.ExerciseId, workoutPlanID, newEP.Sets, newEP.Repetitions, newEP.Weights, newEP.WeightUnit))
		mock.ExpectCommit().WillReturnError(commitErr)

		exercisePlan, err := epRepo.CreateExercisePlan(ctx, newEP, workoutPlanID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.Contains(t, err.Error(), commitErr.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestGetExercisePlanById(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	epRepo := repository.NewEPRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		epID := 1
		expectedEP := repository.ExercisePlan{
			Id:            epID,
			ExerciseId:    101,
			WorkoutPlanId: 201,
			Sets:          4,
			Repetitions:   8,
			Weights:       70.0,
			WeightUnit:    repository.LBS,
		}

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(epID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(expectedEP.Id, expectedEP.ExerciseId, expectedEP.WorkoutPlanId, expectedEP.Sets, expectedEP.Repetitions, expectedEP.Weights, expectedEP.WeightUnit))

		exercisePlan, err := epRepo.GetExercisePlanById(ctx, epID)
		assert.NoError(t, err)
		assert.NotNil(t, exercisePlan)
		assert.Equal(t, expectedEP, *exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		epID := 99
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE id = $1`,
		)).
			ExpectQuery().
			WithArgs(epID).
			WillReturnError(sql.ErrNoRows)

		exercisePlan, err := epRepo.GetExercisePlanById(ctx, epID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		epID := 1
		dbError := errors.New("database connection lost")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE id = $1`,
		)).
			WillReturnError(dbError)

		exercisePlan, err := epRepo.GetExercisePlanById(ctx, epID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query exercise plan by id")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestUpdateExercisePlan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	epRepo := repository.NewEPRepository(db)
	ctx := context.Background()

	t.Run("success with all fields updated", func(t *testing.T) {
		epID := 1
		sets := 4
		repetitions := 12
		weights := float32(60.5)
		weightUnit := repository.KG

		updateData := repository.UpdateEP{
			Id:          epID,
			Sets:        &sets,
			Repetitions: &repetitions,
			Weights:     &weights,
			WeightUnit:  &weightUnit,
		}

		expectedEP := repository.ExercisePlan{
			Id:            epID,
			ExerciseId:    101, // Assuming these remain unchanged
			WorkoutPlanId: 201, // Assuming these remain unchanged
			Sets:          sets,
			Repetitions:   repetitions,
			Weights:       weights,
			WeightUnit:    weightUnit,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1")).WithArgs(epID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(epID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE exercise_plans
					SET sets = COALESCE($1, sets),
						repetitions = COALESCE($2, repetitions),
						weights = COALESCE($3, weights),
						weight_unit = COALESCE($4, weight_unit)
						WHERE id = $5
						RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).
			WithArgs(*updateData.Sets, *updateData.Repetitions, *updateData.Weights, *updateData.WeightUnit, updateData.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(expectedEP.Id, expectedEP.ExerciseId, expectedEP.WorkoutPlanId, expectedEP.Sets, expectedEP.Repetitions, expectedEP.Weights, expectedEP.WeightUnit))
		mock.ExpectCommit()

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.NoError(t, err)
		assert.NotNil(t, exercisePlan)
		assert.Equal(t, expectedEP, *exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with partial fields updated (e.g., only sets)", func(t *testing.T) {
		epID := 1
		sets := 5
		// Original values for other fields
		originalRepetitions := 10
		originalWeights := float32(50.0)
		originalWeightUnit := repository.KG

		updateData := repository.UpdateEP{
			Id:   epID,
			Sets: &sets,
			// Repetitions, Weights, WeightUnit are nil
		}

		expectedEP := repository.ExercisePlan{
			Id:            epID,
			ExerciseId:    101,
			WorkoutPlanId: 201,
			Sets:          sets,
			Repetitions:   originalRepetitions, // Should be original value
			Weights:       originalWeights,     // Should be original value
			WeightUnit:    originalWeightUnit,  // Should be original value
		}

		// Now expect the update
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1")).WithArgs(epID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(epID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE exercise_plans
					SET sets = COALESCE($1, sets),
						repetitions = COALESCE($2, repetitions),
						weights = COALESCE($3, weights),
						weight_unit = COALESCE($4, weight_unit)
						WHERE id = $5
						RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).
			WithArgs(*updateData.Sets, nil, nil, nil, updateData.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(expectedEP.Id, expectedEP.ExerciseId, expectedEP.WorkoutPlanId, expectedEP.Sets, expectedEP.Repetitions, expectedEP.Weights, expectedEP.WeightUnit))
		mock.ExpectCommit()

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.NoError(t, err)
		assert.NotNil(t, exercisePlan)
		assert.Equal(t, expectedEP, *exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		epID := 99
		sets := 4
		updateData := repository.UpdateEP{
			Id:   epID,
			Sets: &sets,
		}

		// Mock the initial SELECT to return no rows
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1",
		)).
			WithArgs(epID).
			WillReturnError(sql.ErrNoRows)

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during initial select", func(t *testing.T) {
		epID := 1
		sets := 4
		updateData := repository.UpdateEP{
			Id:   epID,
			Sets: &sets,
		}
		dbError := errors.New("db select error")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1")).WithArgs(epID).
			WillReturnError(dbError)

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query exercise plan by id")
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during update", func(t *testing.T) {
		epID := 1
		sets := 4
		repetitions := 12
		weights := float32(60.5)
		weightUnit := repository.KG

		updateData := repository.UpdateEP{
			Id:          epID,
			Sets:        &sets,
			Repetitions: &repetitions,
			Weights:     &weights,
			WeightUnit:  &weightUnit,
		}

		dbError := errors.New("database update failed")

		// Mock the initial SELECT to get existing values
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1")).WithArgs(epID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(epID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE exercise_plans
					SET sets = COALESCE($1, sets),
						repetitions = COALESCE($2, repetitions),
						weights = COALESCE($3, weights),
						weight_unit = COALESCE($4, weight_unit)
						WHERE id = $5
						RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).WithArgs(*updateData.Sets, *updateData.Repetitions, *updateData.Weights, *updateData.WeightUnit, updateData.Id).
			WillReturnError(dbError)
		mock.ExpectRollback()

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to update and scan exercise plan")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("db error during begin transaction", func(t *testing.T) {
		epID := 1
		sets := 4
		updateData := repository.UpdateEP{
			Id:   epID,
			Sets: &sets,
		}
		beginErr := errors.New("failed to begin transaction")

		// Mock the initial SELECT to get existing values
		mock.ExpectBegin().WillReturnError(beginErr)

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to begin transaction")
		assert.Contains(t, err.Error(), beginErr.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error during commit transaction", func(t *testing.T) {
		epID := 1
		sets := 4
		repetitions := 12
		weights := float32(60.5)
		weightUnit := repository.KG

		updateData := repository.UpdateEP{
			Id:          epID,
			Sets:        &sets,
			Repetitions: &repetitions,
			Weights:     &weights,
			WeightUnit:  &weightUnit,
		}
		commitErr := errors.New("transaction commit failed")

		// Mock the initial SELECT to get existing values
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(
			"SELECT id FROM exercise_plans WHERE id = $1")).WithArgs(epID).WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(epID))
		mock.ExpectQuery(regexp.QuoteMeta(
			`UPDATE exercise_plans
					SET sets = COALESCE($1, sets),
						repetitions = COALESCE($2, repetitions),
						weights = COALESCE($3, weights),
						weight_unit = COALESCE($4, weight_unit)
						WHERE id = $5
						RETURNING id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit`,
		)).WithArgs(*updateData.Sets, *updateData.Repetitions, *updateData.Weights, *updateData.WeightUnit, updateData.Id).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
				AddRow(epID, 101, 201, 3, 10, 50.0, repository.KG))

		mock.ExpectCommit().WillReturnError(commitErr)

		exercisePlan, err := epRepo.UpdateExercisePlan(ctx, updateData)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to commit transaction")
		assert.Contains(t, err.Error(), commitErr.Error())
		assert.Nil(t, exercisePlan)

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDeleteExercisePlanByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	epRepo := repository.NewEPRepository(db)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		epID := 1
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercise_plans WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(epID).
			WillReturnResult(sqlmock.NewResult(0, 1)) // 1 row affected

		err := epRepo.DeleteExercisePlanByID(ctx, epID)
		assert.NoError(t, err)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not found", func(t *testing.T) {
		epID := 99
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercise_plans WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(epID).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		err := epRepo.DeleteExercisePlanByID(ctx, epID)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrNotFound))

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		epID := 1
		dbError := errors.New("database down")
		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercise_plans WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(epID).
			WillReturnError(dbError)

		err := epRepo.DeleteExercisePlanByID(ctx, epID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to delete exercise plan with id")
		assert.Contains(t, err.Error(), dbError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("rows affected error", func(t *testing.T) {
		epID := 1
		rowsAffectedError := errors.New("cannot get rows affected")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`DELETE FROM exercise_plans WHERE id = $1`,
		)).
			ExpectExec().
			WithArgs(epID).
			WillReturnResult(sqlmock.NewErrorResult(rowsAffectedError))

		err := epRepo.DeleteExercisePlanByID(ctx, epID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get rows affected after deleting exercise plan with id")
		assert.Contains(t, err.Error(), rowsAffectedError.Error())

		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestListExercisePlans(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	epRepo := repository.NewEPRepository(db)
	ctx := context.Background()

	workoutID := 10

	t.Run("success with multiple exercise plans", func(t *testing.T) {
		expectedEPs := []repository.ExercisePlan{
			{Id: 1, ExerciseId: 101, WorkoutPlanId: workoutID, Sets: 3, Repetitions: 10, Weights: 50.0, WeightUnit: repository.KG},
			{Id: 2, ExerciseId: 102, WorkoutPlanId: workoutID, Sets: 4, Repetitions: 8, Weights: 70.0, WeightUnit: repository.LBS},
		}

		rows := sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"}).
			AddRow(expectedEPs[0].Id, expectedEPs[0].ExerciseId, expectedEPs[0].WorkoutPlanId, expectedEPs[0].Sets, expectedEPs[0].Repetitions, expectedEPs[0].Weights, expectedEPs[0].WeightUnit).
			AddRow(expectedEPs[1].Id, expectedEPs[1].ExerciseId, expectedEPs[1].WorkoutPlanId, expectedEPs[1].Sets, expectedEPs[1].Repetitions, expectedEPs[1].Weights, expectedEPs[1].WeightUnit)

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE workout_plan_id = $1`,
		)).
			ExpectQuery().
			WithArgs(workoutID).
			WillReturnRows(rows)

		exercisePlans, err := epRepo.ListExercisePlans(ctx, workoutID)
		assert.NoError(t, err)
		assert.NotNil(t, exercisePlans)
		assert.Len(t, exercisePlans, 2)
		assert.Equal(t, expectedEPs, exercisePlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with no exercise plans", func(t *testing.T) {
		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE workout_plan_id = $1`,
		)).
			ExpectQuery().
			WithArgs(workoutID).
			WillReturnRows(sqlmock.NewRows([]string{"id", "exercise_id", "workout_plan_id", "sets", "repetitions", "weights", "weight_unit"})) // No rows

		exercisePlans, err := epRepo.ListExercisePlans(ctx, workoutID)
		assert.NoError(t, err)
		assert.Empty(t, exercisePlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("db error", func(t *testing.T) {
		dbError := errors.New("network error")

		mock.ExpectPrepare(regexp.QuoteMeta(
			`SELECT id, exercise_id, workout_plan_id, sets, repetitions, weights, weight_unit FROM exercise_plans WHERE workout_plan_id = $1`,
		)).
			WillReturnError(dbError)

		exercisePlans, err := epRepo.ListExercisePlans(ctx, workoutID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to query exercise plans for workout plan id")
		assert.Contains(t, err.Error(), dbError.Error())
		assert.Nil(t, exercisePlans)

		assert.NoError(t, mock.ExpectationsWereMet())
	})

}
