package helper

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/pkg/api"
)

func SendSuccessResponse(w http.ResponseWriter, statusCode int, data *api.Success) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func SendErrorResponse(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	message := "Internal Server Error"
	errorCode := apperrors.INTERNAL_ERROR // Default error code

	var valErr *apperrors.ValidationError

	if errors.As(err, &valErr) {
		statusCode = http.StatusBadRequest
		message = valErr.Error()
		errorCode = apperrors.ErrorCode(valErr.Field) // Use ValidationField as error code
	} else if errors.Is(err, apperrors.ErrNotFound) {
		statusCode = http.StatusNotFound
		message = err.Error()
		errorCode = apperrors.NOT_FOUND
	} else if errors.Is(err, apperrors.ErrAlreadyExists) {
		statusCode = http.StatusConflict
		message = err.Error()
		errorCode = apperrors.ALREADY_EXISTS
	} else if errors.Is(err, apperrors.ErrUnauthorized) {
		statusCode = http.StatusUnauthorized
		message = err.Error()
		errorCode = apperrors.UNAUTHORIZED
	} else if errors.Is(err, apperrors.ErrForbidden) {
		statusCode = http.StatusForbidden
		message = err.Error()
		errorCode = apperrors.FORBIDDEN
	} else if errors.Is(err, apperrors.ErrInvalidInput) {
		statusCode = http.StatusBadRequest
		message = err.Error()
		errorCode = apperrors.BAD_REQUEST
	} else {
		// Log unhandled errors for debugging
		log.Printf("Unhandled error in handler: %v", err)
	}

	// Create a generic error response structure
	errorResponse := api.Error{
		Code:    string(errorCode),
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}
