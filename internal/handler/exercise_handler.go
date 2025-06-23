package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"
)

type ExerciseHandler struct {
	ExerciseService service.ExerciseServiceInterface
}

func NewExerciseHandler(es service.ExerciseServiceInterface) *ExerciseHandler {
	return &ExerciseHandler{
		ExerciseService: es,
	}
}

// ListExercises
func (ec *ExerciseHandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	exerciseList, err := ec.ExerciseService.ListExercises(r.Context())

	if err != nil {
		helper.SendErrorResponse(w, fmt.Errorf("failed to fetch exercises: %w", err))
		return
	}

	var exercises []api.Exercise

	for _, exer := range exerciseList {
		apiExer := toAPIExercise(&exer)
		exercises = append(exercises, *apiExer)
	}

	response := api.Success{
		Code:    api.FETCH,
		Message: "successfully fetch exercises",
		Payload: &map[string]any{
			"exercises": exercises,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)
}

// GetExerciseByID
func (ec *ExerciseHandler) GetExerciseByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("exerciseId")
	if id == "" {
		helper.SendErrorResponse(w, apperrors.ErrInvalidInput)
		return
	}

	exerciseID, err := strconv.Atoi(id)
	if err != nil {
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_ID, "exercise id is not valid"))
		return
	}

	exer, err := ec.ExerciseService.GetExerciseById(r.Context(), exerciseID)

	if err != nil {
		helper.SendErrorResponse(w, apperrors.ErrNotFound)
		return
	}

	exericse := toAPIExercise(exer)

	response := api.Success{
		Code:    api.FETCH,
		Message: "successfully fetch exercise",
		Payload: &map[string]interface{}{
			"exercise": exericse,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

func toAPIExercise(serviceExers *service.Exercise) *api.Exercise {
	if serviceExers == nil {
		return nil
	}

	return &api.Exercise{
		Description: &serviceExers.Description,
		Id:          util.IntTo64(serviceExers.Id),
		MuscleGroup: (*api.MuscleGroup)(&serviceExers.MuscleGroup),
		Name:        &serviceExers.Name,
	}
}
