package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"
)

type WorkoutHandler struct {
	WorkoutService service.WorkoutServiceInterface
}

func NewWorkoutHandler(ws service.WorkoutServiceInterface) *WorkoutHandler {
	return &WorkoutHandler{
		WorkoutService: ws,
	}
}

// ListWorkoutPlans
func (h *WorkoutHandler) ListWorkoutPlans(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	userInfo, ok := helper.GetUserInfoFromContext(r.Context())

	if !ok {
		log.Printf("Failed to get user info from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return
	}

	if query.Has("status") {
		status := query.Get("status")
		switch status {
		case string(api.Pending):
		case string(api.Completed):
		case string(api.Missed):
		default:
			helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "status not valid"))
			return
		}
		var Isasc bool
		sort := query.Get("sort")
		switch sort {
		case "asc":
			Isasc = true
		default:
			Isasc = false
		}

		wpList, err := h.WorkoutService.ListWorkoutsByStatus(r.Context(), userInfo.Id, service.WPStatus(status), Isasc)

		if err != nil {
			helper.SendErrorResponse(w, fmt.Errorf("failed to fetch workout plans: %w", err))
			return
		}

		var workoutPlans []api.WorkoutPlan
		for _, wp := range wpList {
			apiWP := toAPIWorkout(&wp)
			workoutPlans = append(workoutPlans, *apiWP)
		}

		response := api.Success{
			Code:    api.FETCH,
			Message: "successfully fetch workout plans",
			Payload: &map[string]any{
				"workoutPlans": workoutPlans,
			},
		}

		helper.SendSuccessResponse(w, http.StatusOK, &response)
		return
	}

	wpList, err := h.WorkoutService.ListWorkouts(r.Context(), userInfo.Id)

	if err != nil {
		helper.SendErrorResponse(w, fmt.Errorf("failed to fetch workout plans: %w", err))
		return
	}
	response := api.Success{
		Code:    api.FETCH,
		Message: "successfully fetch workout plans",
		Payload: &map[string]any{
			"workoutPlans": wpList,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

// CreateWorkoutPlan
func (h *WorkoutHandler) CreateWorkoutPlan(w http.ResponseWriter, r *http.Request) {
	var req api.CreateWorkoutPlanJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding creating request: %v", err)
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return
	}

	userInfo, ok := helper.GetUserInfoFromContext(r.Context())

	if !ok {
		log.Printf("Failed to get user info from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return
	}

	var createEPs []service.ExercisePlanCreate
	for _, ep := range *req.ExercisePlans {
		createEP := toServiceCreateEP(&ep)
		createEPs = append(createEPs, *createEP)
	}

	var input = service.WorkoutPlanCreate{
		UserId:        userInfo.Id,
		ScheduledDate: req.ScheduledDate,
		ExercisePlans: createEPs,
	}

	wp, err := h.WorkoutService.CreateWorkout(r.Context(), input)
	if err != nil {

		var validationErr *apperrors.ValidationError

		if errors.As(err, &validationErr) {
			helper.SendErrorResponse(w, err)
			return
		}

		helper.SendErrorResponse(w, fmt.Errorf("error creating workout plan: %w", err))
		return
	}

	workoutPlan := toAPIWorkout(wp)

	response := api.Success{
		Code:    api.CREATED,
		Message: "successfully create workout plan",
		Payload: &map[string]interface{}{
			"workoutPlan": workoutPlan,
		},
	}

	helper.SendSuccessResponse(w, http.StatusCreated, &response)
}

// GetWorkoutPlanById
func (h *WorkoutHandler) GetWorkoutPlanById(w http.ResponseWriter, r *http.Request) {

	wpId, err := doubleAuth(w, r, h.WorkoutService)

	if err != nil {
		log.Print(err)
		return
	}

	wp, err := h.WorkoutService.GetWorkoutById(r.Context(), wpId)

	if err != nil {
		helper.SendErrorResponse(w, apperrors.ErrNotFound)
		return
	}

	workoutPlan := toAPIWorkout(wp)
	response := api.Success{
		Code:    api.FETCH,
		Message: "successfully fetch workout plan",
		Payload: &map[string]any{
			"workoutPlan": workoutPlan,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

// UpdateExercisPlans
func (h *WorkoutHandler) UpdateExercisePlansInWorkoutPlan(w http.ResponseWriter, r *http.Request) {
	wpId, err := doubleAuth(w, r, h.WorkoutService)
	if err != nil {
		log.Print(err)
		return
	}

	var req api.UpdateExercisePlansInWorkoutPlanJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup request: %v", err)
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return

	}

	var serviceUpdateEPs []service.ExercisePlanUpdate

	for _, ep := range *req.ExercisePlans {
		serviceEP := toServiceUpdateEP(&ep)

		serviceUpdateEPs = append(serviceUpdateEPs, *serviceEP)
	}

	updatedWP, err := h.WorkoutService.UpdateExercisePlans(r.Context(), wpId, serviceUpdateEPs)

	if err != nil {
		var validationErr *apperrors.ValidationError

		if errors.As(err, &validationErr) {
			helper.SendErrorResponse(w, err)
			return
		}

		helper.SendErrorResponse(w, fmt.Errorf("failed to update exercise plans: %w", err))
		return
	}

	workoutPlan := toAPIWorkout(updatedWP)

	response := api.Success{
		Code:    api.UPDATE,
		Message: "successfully update exercise plans",
		Payload: &map[string]interface{}{
			"workoutPlan": workoutPlan,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

// CompleteWorkoutPlanById
func (h *WorkoutHandler) CompleteWorkoutPlanById(w http.ResponseWriter, r *http.Request) {
	wpId, err := doubleAuth(w, r, h.WorkoutService)
	if err != nil {
		log.Print(err)
		return
	}

	var req api.CompleteWorkoutPlanByIdJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup request: %v", err)
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return

	}

	err = h.WorkoutService.CompleteWorkout(r.Context(), wpId, req.Comment)

	if err != nil {
		helper.SendErrorResponse(w, fmt.Errorf("failed to complete workout plan: %w", err))
		return
	}
	helper.SendSuccessResponse(w, http.StatusNoContent, nil)
}

// ScheduleWorkoutPlanById
func (h *WorkoutHandler) ScheduleWorkoutPlanById(w http.ResponseWriter, r *http.Request) {

	wpId, err := doubleAuth(w, r, h.WorkoutService)
	if err != nil {
		log.Print(err)
		return
	}

	var req api.ScheduleWorkoutPlanByIdJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup request: %v", err)
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return

	}

	updatedWP, err := h.WorkoutService.ScheduleWorkout(r.Context(), wpId, req.ScheduledDate)
	if err != nil {
		var validationErr *apperrors.ValidationError
		if errors.As(err, &validationErr) {
			helper.SendErrorResponse(w, err)
			return
		}

		helper.SendErrorResponse(w, fmt.Errorf("failed to schedule workout plan: %w", err))
		return
	}

	workoutPlan := toAPIWorkout(updatedWP)

	response := api.Success{
		Code:    api.UPDATE,
		Message: "successfully schedule workout plan",
		Payload: &map[string]interface{}{
			"workoutPlan": workoutPlan,
		},
	}

	helper.SendSuccessResponse(w, http.StatusOK, &response)

}

// DeleteWorkoutPlanById
func (h *WorkoutHandler) DeleteWoroutPlanById(w http.ResponseWriter, r *http.Request) {
	wpId, err := doubleAuth(w, r, h.WorkoutService)

	if err != nil {
		log.Print(err)
		return
	}

	err = h.WorkoutService.DeleteWorkoutById(r.Context(), wpId)

	if err != nil {

		helper.SendErrorResponse(w, fmt.Errorf("failed to delete workout plan: %w", err))
		return
	}

	helper.SendSuccessResponse(w, http.StatusNoContent, nil)

}

func toAPIWorkout(workout *service.WorkoutPlan) *api.WorkoutPlan {
	if workout == nil {
		return nil
	}

	createAt := workout.CreatedAt
	updatedAt := workout.UpdatedAt
	sheduledDate := workout.ScheduledDate
	var exercisePlans []api.ExercisePlan

	for _, ep := range workout.ExercisePlans {
		apiEP := toAPIExercisePlan(&ep)

		exercisePlans = append(exercisePlans, *apiEP)
	}

	return &api.WorkoutPlan{
		Comment:       workout.Comment,
		CreatedAt:     &createAt,
		UpdatedAt:     &updatedAt,
		Id:            util.IntTo64(workout.Id),
		ScheduledDate: &sheduledDate,
		Status:        (*api.WorkoutPlanStatus)(&workout.Status),
		UserId:        util.IntTo64(workout.UserId),
		ExercisePlans: &exercisePlans,
	}
}

func toAPIExercisePlan(exercisePlan *service.ExercisePlan) *api.ExercisePlan {
	if exercisePlan == nil {
		return nil
	}

	return &api.ExercisePlan{
		ExerciseId:    util.IntTo64(exercisePlan.ExerciseId),
		Id:            util.IntTo64(exercisePlan.Id),
		Repetitions:   &exercisePlan.Repetitions,
		Sets:          &exercisePlan.Sets,
		WeightUnit:    (*api.WeightUnit)(&exercisePlan.WeightUnit),
		WorkoutPlanId: util.IntTo64(exercisePlan.WorkoutPlanId),
		Weights:       &exercisePlan.Weights,
	}

}

func toServiceCreateEP(apiEP *api.CreateExercisePlan) *service.ExercisePlanCreate {
	if apiEP == nil {
		return nil
	}

	return &service.ExercisePlanCreate{
		ExerciseId:  int(*apiEP.ExerciseId),
		Repetitions: *apiEP.Repetitions,
		Sets:        *apiEP.Sets,
		Weights:     *apiEP.Weights,
		WeightUnit:  service.WeightUnit(*apiEP.WeightUnit),
	}
}

func toServiceUpdateEP(apiEP *api.UpdateExercisePlan) *service.ExercisePlanUpdate {
	if apiEP == nil {
		return nil
	}

	return &service.ExercisePlanUpdate{
		Id:          int(*apiEP.Id),
		Sets:        *apiEP.Sets,
		Repetitions: *apiEP.Repetitions,
		Weights:     *apiEP.Weights,
		WeightUnit:  service.WeightUnit(*apiEP.WeightUnit),
	}
}

func doubleAuth(w http.ResponseWriter, r *http.Request, workoutService service.WorkoutServiceInterface) (workoutId int, err error) {
	id := r.PathValue("workoutId")
	if id == "" {
		err := apperrors.NewValidationError(apperrors.INVALID_INPUT, "workout id not set in path")
		helper.SendErrorResponse(w, err)
		return -1, err
	}

	wpId, err := strconv.Atoi(id)
	if err != nil {
		err := apperrors.NewValidationError(apperrors.INVALID_ID, "workout id not valid")
		helper.SendErrorResponse(w, err)
		return -1, err
	}

	userInfo, ok := helper.GetUserInfoFromContext(r.Context())

	if !ok {
		err := fmt.Errorf("failed to get user info from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return -1, err
	}

	exsitingWP, err := workoutService.GetWorkoutById(r.Context(), wpId)

	if err != nil {
		err := fmt.Errorf("error fetching workout plan %d for operation by user %d", wpId, userInfo.Id)
		helper.SendErrorResponse(w, apperrors.ErrNotFound)
		return -1, err
	}

	if exsitingWP.UserId != userInfo.Id {
		err := fmt.Errorf("unauthorized update attempt: User %d tried to operate workout %d", userInfo.Id, wpId)
		helper.SendErrorResponse(w, apperrors.ErrForbidden)
		return 0, err
	}

	return exsitingWP.Id, nil

}
