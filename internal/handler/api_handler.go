package handler

import (
	"net/http"
	"strconv"
	"workout-tracker-api/pkg/api"
)

type APIhandler struct {
	UserHandler     *UserHandler
	WorkoutHandler  *WorkoutHandler
	ExerciseHandler *ExerciseHandler
	ReportHandler   *ReportHandler
}

// CompleteWorkoutPlanById implements api.ServerInterface.
func (a *APIhandler) CompleteWorkoutPlanById(w http.ResponseWriter, r *http.Request, workoutId int64) {
	r.SetPathValue("workoutId", strconv.Itoa(int(workoutId)))
	a.WorkoutHandler.CompleteWorkoutPlanById(w, r)
}

// CreateWorkoutPlan implements api.ServerInterface.
func (a *APIhandler) CreateWorkoutPlan(w http.ResponseWriter, r *http.Request) {
	a.WorkoutHandler.CreateWorkoutPlan(w, r)
}

// DeleteWorkoutPlanById implements api.ServerInterface.
func (a *APIhandler) DeleteWorkoutPlanById(w http.ResponseWriter, r *http.Request, workoutId int64) {
	r.SetPathValue("workoutId", strconv.Itoa(int(workoutId)))
	a.WorkoutHandler.DeleteWoroutPlanById(w, r)
}

// GetExerciseById implements api.ServerInterface.
func (a *APIhandler) GetExerciseById(w http.ResponseWriter, r *http.Request, exerciseId int64) {
	r.SetPathValue("exerciseId", strconv.Itoa(int(exerciseId)))
	a.ExerciseHandler.GetExerciseByID(w, r)
}

// GetUserStatus implements api.ServerInterface.
func (a *APIhandler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	a.UserHandler.GetUserStatus(w, r)
}

// GetWorkoutPlanById implements api.ServerInterface.
func (a *APIhandler) GetWorkoutPlanById(w http.ResponseWriter, r *http.Request, workoutId int64) {
	r.SetPathValue("workoutId", strconv.Itoa(int(workoutId)))
	a.WorkoutHandler.GetWorkoutPlanById(w, r)
}

// ListExercises implements api.ServerInterface.
func (a *APIhandler) ListExercises(w http.ResponseWriter, r *http.Request) {
	a.ExerciseHandler.ListExercises(w, r)
}

// ListWorkoutPlans implements api.ServerInterface.
func (a *APIhandler) ListWorkoutPlans(w http.ResponseWriter, r *http.Request, params api.ListWorkoutPlansParams) {

	a.WorkoutHandler.ListWorkoutPlans(w, r)
}

// LoginUser implements api.ServerInterface.
func (a *APIhandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	a.UserHandler.LoginUser(w, r)
}

// LogoutUser implements api.ServerInterface.
func (a *APIhandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	a.UserHandler.LogoutUser(w, r)
}

// ReportProgress implements api.ServerInterface.
func (a *APIhandler) ReportProgress(w http.ResponseWriter, r *http.Request) {
	a.ReportHandler.ReportProgress(w, r)
}

// ScheduleWorkoutPlanById implements api.ServerInterface.
func (a *APIhandler) ScheduleWorkoutPlanById(w http.ResponseWriter, r *http.Request, workoutId int64) {
	r.SetPathValue("workoutId", strconv.Itoa(int(workoutId)))
	a.WorkoutHandler.ScheduleWorkoutPlanById(w, r)
}

// SignupUser implements api.ServerInterface.
func (a *APIhandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	a.UserHandler.SignupUser(w, r)
}

// UpdateExercisePlansInWorkoutPlan implements api.ServerInterface.
func (a *APIhandler) UpdateExercisePlansInWorkoutPlan(w http.ResponseWriter, r *http.Request, workoutId int64) {

	r.SetPathValue("workoutId", strconv.Itoa(int(workoutId)))
	a.WorkoutHandler.UpdateExercisePlansInWorkoutPlan(w, r)
}

func NewAPIHandler(
	userH *UserHandler,
	workoutH *WorkoutHandler,
	exerciseH *ExerciseHandler,
	reportH *ReportHandler,
) api.ServerInterface {
	return &APIhandler{
		UserHandler:     userH,
		WorkoutHandler:  workoutH,
		ExerciseHandler: exerciseH,
		ReportHandler:   reportH,
	}
}
