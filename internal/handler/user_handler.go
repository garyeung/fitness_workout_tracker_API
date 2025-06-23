package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util"
	"workout-tracker-api/internal/util/auth"
	"workout-tracker-api/internal/util/helper"
	"workout-tracker-api/pkg/api"

	"github.com/oapi-codegen/runtime/types"
)

type UserHandler struct {
	UserService    service.UserServiceInterface
	WorkoutService service.WorkoutServiceInterface
	TokenService   auth.TokenInterface
}

func NewUserHandler(us service.UserServiceInterface, ws service.WorkoutServiceInterface, ts auth.TokenInterface) *UserHandler {
	return &UserHandler{
		UserService:    us,
		WorkoutService: ws,
		TokenService:   ts,
	}
}

// signup
func (h *UserHandler) SignupUser(w http.ResponseWriter, r *http.Request) {
	var req api.SignupUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding signup request: %v", err)
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return
	}

	ctx := r.Context()

	_, err := h.UserService.SignupUser(ctx, service.UserSignup{
		Name:     req.Name,
		Email:    string(req.Email),
		Password: *req.Password,
	})
	if err != nil {

		var validationErr *apperrors.ValidationError
		if errors.As(err, &validationErr) {
			helper.SendErrorResponse(w, err)
			return
		}

		helper.SendErrorResponse(w, fmt.Errorf("error during user signup: %v", err))
		return
	}

	resposne := api.Success{
		Code:    api.CREATED,
		Message: "successfully create user",
	}
	helper.SendSuccessResponse(w, http.StatusCreated, &resposne)
}

// LoginUser handles POST /user/login requests.
func (h *UserHandler) LoginUser(w http.ResponseWriter, r *http.Request) {
	var loginRequest api.LoginUserJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_INPUT, "invalid request body"))
		return
	}

	// Map API request to service layer input
	serviceLogin := service.UserLogin{
		Email:    string(loginRequest.Email),
		Password: loginRequest.Password,
	}

	// Call the service layer
	user, err := h.UserService.LoginUser(r.Context(), serviceLogin)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			helper.SendErrorResponse(w, apperrors.NewValidationError(apperrors.INVALID_EMAIL, "the email is not registered"))
			return
		}
		var ValidationErr *apperrors.ValidationError
		if errors.As(err, &ValidationErr) {
			helper.SendErrorResponse(w, err)
			return
		}

		helper.SendErrorResponse(w, fmt.Errorf("failed to log in:%v", err))
		return

	}

	// Generate JWT token upon successful login
	token, err := h.TokenService.GenerateToken(auth.Claims{
		Payload: auth.Payload{
			Id:    &user.Id,
			Email: user.Email,
			Name:  user.Name,
		},
	})

	if err != nil {
		helper.SendErrorResponse(w, fmt.Errorf("failed to generate token after login: %w", err))
		return
	}

	var accessToken api.UserToken = token

	response := api.Success{
		Code:    api.FETCH,
		Message: "successfull login",
		Payload: &map[string]any{
			"accessToken": accessToken,
		},
	}
	// Map service user to API response struct
	helper.SendSuccessResponse(w, http.StatusOK, &response)
}

// LogoutUser handles POST /user/logout requests.
func (h *UserHandler) LogoutUser(w http.ResponseWriter, r *http.Request) {
	jti, ok := helper.GetJTIFromContext(r.Context())
	if !ok {
		log.Printf("Failed to get jti from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return
	}

	if jti != nil {
		err := h.TokenService.BlacklistToken(r.Context(), jti.Id, jti.ExpirationTime)
		if err != nil {
			helper.SendErrorResponse(w, fmt.Errorf("failed to blacklist token %v", err))
			return
		}
		helper.SendSuccessResponse(w, http.StatusNoContent, nil)
	} else {
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
	}
}

// GetUserStatus handles GET /user/status requests.
func (h *UserHandler) GetUserStatus(w http.ResponseWriter, r *http.Request) {
	userInfo, ok := helper.GetUserInfoFromContext(r.Context())

	if !ok {
		log.Printf("Failed to get user info from context")
		helper.SendErrorResponse(w, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.UserService.GetUser(r.Context(), userInfo.Email)

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			helper.SendErrorResponse(w, err)
			return
		}
		helper.SendErrorResponse(w, fmt.Errorf("failed to fetch user info :%v", err))
		return
	}

	workouts, err := h.WorkoutService.ListWorkouts(r.Context(), user.Id)
	if err != nil {
		helper.SendErrorResponse(w, fmt.Errorf("failed to fetch workout plans :%v", err))
		return
	}

	var apiWorkouts []api.WorkoutPlan

	for _, wp := range workouts {
		apiWp := toAPIWorkout(&wp)
		apiWorkouts = append(apiWorkouts, *apiWp)
	}
	userStatus := api.UserStatus{
		Email:        (*types.Email)(&user.Email),
		Id:           util.IntTo64(user.Id),
		WorkoutPlans: &apiWorkouts,
	}

	helper.SendSuccessResponse(w, http.StatusOK, &api.Success{
		Code:    api.FETCH,
		Message: "successfully fetch user status",
		Payload: &map[string]interface{}{
			"userStatus": userStatus,
		},
	})

}
