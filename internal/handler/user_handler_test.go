package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/handler"
	"workout-tracker-api/internal/service"
	"workout-tracker-api/internal/util/auth"
	"workout-tracker-api/internal/util/helper" // Assuming this helper exists for context
	"workout-tracker-api/pkg/api"
)

// --- Mock Implementations ---

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) SignupUser(ctx context.Context, input service.UserSignup) (*service.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
}

func (m *MockUserService) LoginUser(ctx context.Context, input service.UserLogin) (*service.User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, email string) (*service.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.User), args.Error(1)
}

type MockUserWorkoutService struct {
	mock.Mock
}

func (m *MockUserWorkoutService) CreateWorkout(ctx context.Context, input service.WorkoutPlanCreate) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockUserWorkoutService) GetWorkoutById(ctx context.Context, id int) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockUserWorkoutService) ListWorkouts(ctx context.Context, userID int) ([]service.WorkoutPlan, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.WorkoutPlan), args.Error(1)
}

func (m *MockUserWorkoutService) ListWorkoutsByStatus(ctx context.Context, userId int, status service.WPStatus, asc bool) ([]service.WorkoutPlan, error) {
	args := m.Called(ctx, userId, status, asc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]service.WorkoutPlan), args.Error(1)
}

func (m *MockUserWorkoutService) UpdateExercisePlans(ctx context.Context, workoutID int, epsUpdate []service.ExercisePlanUpdate) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, workoutID, epsUpdate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

func (m *MockUserWorkoutService) DeleteWorkoutById(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserWorkoutService) CompleteWorkout(ctx context.Context, id int, comment *string) error {
	args := m.Called(ctx, id, comment)
	return args.Error(0)
}

func (m *MockUserWorkoutService) ScheduleWorkout(ctx context.Context, id int, scheduledDate *time.Time) (*service.WorkoutPlan, error) {
	args := m.Called(ctx, id, scheduledDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.WorkoutPlan), args.Error(1)
}

type MockTokenService struct {
	mock.Mock
}

func (m *MockTokenService) GenerateToken(claims auth.Claims) (string, error) {
	args := m.Called(claims)
	return args.String(0), args.Error(1)
}

func (m *MockTokenService) ParseToken(ctx context.Context, tokenString string) (*auth.Claims, error) {
	args := m.Called(ctx, tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockTokenService) BlacklistToken(ctx context.Context, jti string, expirationTime time.Time) error {
	args := m.Called(ctx, jti, expirationTime)
	return args.Error(0)
}

func (m *MockTokenService) CheckBlacklist(ctx context.Context, jti string) (bool, error) {
	args := m.Called(ctx, jti)
	return args.Bool(0), args.Error(1)
}

// --- Test Suite ---

func TestUserHandler(t *testing.T) {
	mockUserService := new(MockUserService)
	mockWorkoutService := new(MockUserWorkoutService)
	mockTokenService := new(MockTokenService)

	userHandler := handler.NewUserHandler(mockUserService, mockWorkoutService, mockTokenService)

	// --- Test SignupUser ---
	t.Run("SignupUser - Success", func(t *testing.T) {
		reqBody := `{"name": "John Doe", "email": "john@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/signup", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		expectedServiceInput := service.UserSignup{
			Name:     "John Doe",
			Email:    "john@example.com",
			Password: "password123",
		}
		mockUserService.On("SignupUser", mock.Anything, expectedServiceInput).Return(&service.User{
			Id:        1,
			Name:      "John Doe",
			Email:     "john@example.com",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil).Once()

		userHandler.SignupUser(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		var resp api.Success
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, api.CREATED, resp.Code)
		assert.Equal(t, "successfully create user", resp.Message)
		mockUserService.AssertExpectations(t)
	})

	t.Run("SignupUser - Invalid Input (empty email)", func(t *testing.T) {
		reqBody := `{"name": "John Doe", "email": "", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/signup", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		userHandler.SignupUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.INVALID_INPUT), resp.Code)
		assert.Contains(t, resp.Message, "invalid request body")
		mockUserService.AssertNotCalled(t, "SignupUser")

	})

	t.Run("SignupUser - Email Already Registered", func(t *testing.T) {
		reqBody := `{"name": "Jane Doe", "email": "jane@example.com", "password": "password123"}`
		req := httptest.NewRequest(http.MethodPost, "/user/signup", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		expectedServiceInput := service.UserSignup{
			Name:     "Jane Doe",
			Email:    "jane@example.com",
			Password: "password123",
		}
		mockUserService.On("SignupUser", mock.Anything, expectedServiceInput).Return(nil, apperrors.NewValidationError(apperrors.INVALID_EMAIL, "email have already registered")).Once()

		userHandler.SignupUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.INVALID_EMAIL), resp.Code)
		assert.Contains(t, resp.Message, "email have already registered")
		mockUserService.AssertExpectations(t)
	})

	// --- Test LoginUser ---
	t.Run("LoginUser - Success", func(t *testing.T) {
		mockId := 10
		reqBody := `{"email": "user@example.com", "password": "correctpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		expectedUserServiceLoginInput := service.UserLogin{
			Email:    "user@example.com",
			Password: "correctpassword",
		}
		returnedUser := &service.User{
			Id:    mockId,
			Email: "user@example.com",
			Name:  "Test User",
		}
		mockUserService.On("LoginUser", mock.Anything, expectedUserServiceLoginInput).Return(returnedUser, nil).Once()

		expectedTokenClaims := auth.Claims{
			Payload: auth.Payload{
				Id:    &mockId,
				Email: returnedUser.Email,
				Name:  returnedUser.Name,
			},
		}
		mockTokenService.On("GenerateToken", expectedTokenClaims).Return("mock_jwt_token", nil).Once()

		userHandler.LoginUser(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp api.Success
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, api.FETCH, resp.Code)
		assert.Equal(t, "successfull login", resp.Message)
		assert.Contains(t, (*resp.Payload)["accessToken"].(string), "mock_jwt_token") // Check token in payload
		mockUserService.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
	})

	t.Run("LoginUser - Invalid Credentials (User Not Found)", func(t *testing.T) {
		reqBody := `{"email": "nonexistent@example.com", "password": "anypassword"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		expectedUserServiceLoginInput := service.UserLogin{
			Email:    "nonexistent@example.com",
			Password: "anypassword",
		}
		mockUserService.On("LoginUser", mock.Anything, expectedUserServiceLoginInput).Return(nil, apperrors.ErrNotFound).Once()

		userHandler.LoginUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.INVALID_EMAIL), resp.Code)
		assert.Contains(t, resp.Message, "the email is not registered")
		mockUserService.AssertExpectations(t)
		mockTokenService.AssertNotCalled(t, "GenerateToken")
	})

	t.Run("LoginUser - Invalid Credentials (Incorrect Password)", func(t *testing.T) {
		reqBody := `{"email": "user@example.com", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		expectedUserServiceLoginInput := service.UserLogin{
			Email:    "user@example.com",
			Password: "wrongpassword",
		}
		mockUserService.On("LoginUser", mock.Anything, expectedUserServiceLoginInput).Return(nil, apperrors.NewValidationError(apperrors.INVALID_PASSWORD, "invalid password")).Once()

		userHandler.LoginUser(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.INVALID_PASSWORD), resp.Code)
		assert.Contains(t, resp.Message, "invalid password")
		mockUserService.AssertExpectations(t)
		mockTokenService.AssertNotCalled(t, "GenerateToken")
	})

	t.Run("LoginUser - Token Generation Error", func(t *testing.T) {
		reqBody := `{"email": "user@example.com", "password": "correctpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/user/login", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		returnedUser := &service.User{
			Id:    10,
			Email: "user@example.com",
			Name:  "Test User",
		}
		mockUserService.On("LoginUser", mock.Anything, mock.Anything).Return(returnedUser, nil).Once()
		mockTokenService.On("GenerateToken", mock.Anything).Return("", errors.New("token generation failed")).Once()

		userHandler.LoginUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "Internal Server Error")
		mockUserService.AssertExpectations(t)
		mockTokenService.AssertExpectations(t)
	})

	// --- Test LogoutUser ---
	t.Run("LogoutUser - Success", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rr := httptest.NewRecorder()

		// Simulate JTI in context
		testJTI := helper.JTIInfo{
			Id:             "test-jti-123",
			ExpirationTime: time.Now().Add(time.Hour).UTC(),
		}
		ctx := context.WithValue(req.Context(), helper.JTIContextKey, &testJTI)
		req = req.WithContext(ctx)

		mockTokenService.On("BlacklistToken", req.Context(), testJTI.Id, testJTI.ExpirationTime).Return(nil).Once()

		userHandler.LogoutUser(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Empty(t, rr.Body.String()) // No content for 204
		mockTokenService.AssertExpectations(t)
	})

	t.Run("LogoutUser - No JTI in Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rr := httptest.NewRecorder()

		userHandler.LogoutUser(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
		mockTokenService.AssertNotCalled(t, "BlacklistToken")
	})

	t.Run("LogoutUser - Blacklist Token Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/user/logout", nil)
		rr := httptest.NewRecorder()

		testJTI := helper.JTIInfo{
			Id:             "test-jti-456",
			ExpirationTime: time.Now().Add(time.Hour).UTC(),
		}
		ctx := context.WithValue(req.Context(), helper.JTIContextKey, &testJTI)
		req = req.WithContext(ctx)

		mockTokenService.On("BlacklistToken", mock.Anything, testJTI.Id, testJTI.ExpirationTime).Return(errors.New("blacklist failed")).Once()

		userHandler.LogoutUser(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "Internal Server Error")
		mockTokenService.AssertExpectations(t)
	})

	// --- Test GetUserStatus ---
	t.Run("GetUserStatus - Success", func(t *testing.T) {
		mockId := 10
		mockComment := "Frist workout"
		req := httptest.NewRequest(http.MethodGet, "/user/status", nil)
		rr := httptest.NewRecorder()

		testUserInfo := helper.UserInfo{
			Id:    mockId,
			Email: "status@example.com",
			Name:  "Status User",
		}
		ctx := context.WithValue(req.Context(), helper.UserContextKey, &testUserInfo)
		req = req.WithContext(ctx)

		fetchedUser := &service.User{
			Id:        mockId,
			Email:     "status@example.com",
			Name:      "Status User",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockUserService.On("GetUser", mock.Anything, testUserInfo.Email).Return(fetchedUser, nil).Once()

		mockWorkouts := []service.WorkoutPlan{
			{
				Id:            1,
				UserId:        fetchedUser.Id,
				Status:        "completed",
				ScheduledDate: time.Now().Add(-24 * time.Hour),
				Comment:       &mockComment,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
				ExercisePlans: []service.ExercisePlan{
					{Id: 101, ExerciseId: 1, Sets: 3, Repetitions: 10, Weights: 50.0, WeightUnit: "kg"},
				},
			},
		}
		mockWorkoutService.On("ListWorkouts", mock.Anything, fetchedUser.Id).Return(mockWorkouts, nil).Once()

		userHandler.GetUserStatus(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp api.Success
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, api.FETCH, resp.Code)
		assert.Equal(t, "successfully fetch user status", resp.Message)

		userStatus, ok := (*resp.Payload)["userStatus"].(map[string]interface{})
		assert.True(t, ok)

		assert.Equal(t, "status@example.com", userStatus["email"])
		assert.Equal(t, float64(mockId), userStatus["id"]) // JSON unmarshals int64 to float64
		assert.NotNil(t, userStatus["workoutPlans"])

		mockUserService.AssertExpectations(t)
		mockWorkoutService.AssertExpectations(t)
	})

	t.Run("GetUserStatus - No User Info in Context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/status", nil) // No user info in context
		rr := httptest.NewRecorder()

		userHandler.GetUserStatus(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.UNAUTHORIZED), resp.Code)
		mockUserService.AssertNotCalled(t, "GetUser")
		mockWorkoutService.AssertNotCalled(t, "ListWorkouts")
	})

	t.Run("GetUserStatus - User Service Error (Not Found)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/status", nil)
		rr := httptest.NewRecorder()

		testUserInfo := helper.UserInfo{
			Email: "unknown@example.com",
			Name:  "Unknown User",
		}
		ctx := context.WithValue(req.Context(), helper.UserContextKey, &testUserInfo)
		req = req.WithContext(ctx)

		mockUserService.On("GetUser", mock.Anything, testUserInfo.Email).Return(nil, apperrors.ErrNotFound).Once()

		userHandler.GetUserStatus(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Equal(t, string(apperrors.NOT_FOUND), resp.Code)
		mockUserService.AssertExpectations(t)
		mockWorkoutService.AssertNotCalled(t, "ListWorkouts")
	})

	t.Run("GetUserStatus - Workout Service Error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/user/status", nil)
		rr := httptest.NewRecorder()

		testUserInfo := helper.UserInfo{
			Email: "status2@example.com",
			Name:  "Status User 2",
		}
		ctx := context.WithValue(req.Context(), helper.UserContextKey, &testUserInfo)
		req = req.WithContext(ctx)

		fetchedUser := &service.User{
			Id:    21,
			Email: "status2@example.com",
			Name:  "Status User 2",
		}
		mockUserService.On("GetUser", mock.Anything, testUserInfo.Email).Return(fetchedUser, nil).Once()
		mockWorkoutService.On("ListWorkouts", mock.Anything, fetchedUser.Id).Return(nil, errors.New("db error listing workouts")).Once()

		userHandler.GetUserStatus(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		var resp api.Error
		err := json.NewDecoder(rr.Body).Decode(&resp)
		assert.NoError(t, err)
		assert.Contains(t, resp.Message, "Internal Server Error")
		mockUserService.AssertExpectations(t)
		mockWorkoutService.AssertExpectations(t)
	})
}
