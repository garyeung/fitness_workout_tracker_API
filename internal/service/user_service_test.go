package service_test

import (
	"context"
	"errors"
	"testing"
	"time"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of repository.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, data repository.UserCreate) (*repository.User, error) {
	args := m.Called(ctx, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*repository.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.User), args.Error(1)
}

func (m *MockUserRepository) ExistUser(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) DeleteUserByEmail(ctx context.Context, email string) error {
	args := m.Called(ctx, email)
	return args.Error(0)
}

// MockHashHelper is a mock implementation of encrypt.HashHelperInterface
type MockHashHelper struct {
	mock.Mock
}

func (m *MockHashHelper) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockHashHelper) CheckPasswordHash(hash, password string) bool {
	args := m.Called(hash, password)
	return args.Bool(0)
}

func TestUserService_SignupUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		input             service.UserSignup
		mockRepoSetup     func(*MockUserRepository)
		mockHashSetup     func(*MockHashHelper)
		expectedUser      *service.User
		expectedErrorType error
	}{
		{
			name: "Successful signup",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("ExistUser", ctx, "test@example.com").Return(false, nil).Once()
				mur.On("CreateUser", ctx, mock.AnythingOfType("repository.UserCreate")).Return(&repository.User{
					Id:           1,
					Name:         "Test User",
					Email:        "test@example.com",
					PasswordHash: "hashedpassword",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				mhh.On("HashPassword", "password123").Return("hashedpassword", nil).Once()
			},
			expectedUser: &service.User{
				Id:    1,
				Name:  "Test User",
				Email: "test@example.com",
			},
			expectedErrorType: nil,
		},
		{
			name: "Signup with invalid name",
			input: service.UserSignup{
				Name:     "", // Invalid name
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup:     func(mur *MockUserRepository) {}, // No repo calls for validation error
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: &apperrors.ValidationError{},
		},
		{
			name: "Signup with invalid email format",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "invalid-email", // Invalid email format
				Password: "password123",
			},
			mockRepoSetup:     func(mur *MockUserRepository) {},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: &apperrors.ValidationError{},
		},
		{
			name: "Signup with short password",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "short", // Short password
			},
			mockRepoSetup:     func(mur *MockUserRepository) {},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: &apperrors.ValidationError{},
		},
		{
			name: "Signup with already registered email",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "existing@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("ExistUser", ctx, "existing@example.com").Return(true, nil).Once()
			},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: &apperrors.ValidationError{},
		},
		{
			name: "Error checking user existence",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("ExistUser", ctx, "test@example.com").Return(false, errors.New("db error")).Once()
			},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: errors.New(""), // Generic error
		},
		{
			name: "Error hashing password",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("ExistUser", ctx, "test@example.com").Return(false, nil).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				mhh.On("HashPassword", "password123").Return("", errors.New("hash error")).Once()
			},
			expectedUser:      nil,
			expectedErrorType: errors.New(""), // Generic error
		},
		{
			name: "Error creating user",
			input: service.UserSignup{
				Name:     "Test User",
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("ExistUser", ctx, "test@example.com").Return(false, nil).Once()
				mur.On("CreateUser", ctx, mock.AnythingOfType("repository.UserCreate")).Return(nil, errors.New("db create error")).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				mhh.On("HashPassword", "password123").Return("hashedpassword", nil).Once()
			},
			expectedUser:      nil,
			expectedErrorType: errors.New(""), // Generic error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockHash := new(MockHashHelper)

			tt.mockRepoSetup(mockRepo)
			tt.mockHashSetup(mockHash)

			userService := service.NewUserService(mockRepo, mockHash)
			user, err := userService.SignupUser(ctx, tt.input)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.ErrorAs(t, err, &validationErr)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				// Don't assert on ID directly as it might be generated by the mock
			}

			mockRepo.AssertExpectations(t)
			mockHash.AssertExpectations(t)
		})
	}
}

func TestUserService_LoginUser(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name              string
		input             service.UserLogin
		mockRepoSetup     func(*MockUserRepository)
		mockHashSetup     func(*MockHashHelper)
		expectedUser      *service.User
		expectedErrorType error
	}{
		{
			name: "Successful login",
			input: service.UserLogin{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, "test@example.com").Return(&repository.User{
					Id:           1,
					Name:         "Test User",
					Email:        "test@example.com",
					PasswordHash: "hashedpassword",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				mhh.On("CheckPasswordHash", "hashedpassword", "password123").Return(true).Once()
			},
			expectedUser: &service.User{
				Id:    1,
				Name:  "Test User",
				Email: "test@example.com",
			},
			expectedErrorType: nil,
		},
		{
			name: "Login with non-existent user",
			input: service.UserLogin{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, "nonexistent@example.com").Return(nil, apperrors.ErrNotFound).Once()
			},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: apperrors.ErrNotFound,
		},
		{
			name: "Login with incorrect password",
			input: service.UserLogin{
				Email:    "test@example.com",
				Password: "wrongpassword",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, "test@example.com").Return(&repository.User{
					Id:           1,
					Name:         "Test User",
					Email:        "test@example.com",
					PasswordHash: "hashedpassword",
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				mhh.On("CheckPasswordHash", "hashedpassword", "wrongpassword").Return(false).Once()
			},
			expectedUser:      nil,
			expectedErrorType: &apperrors.ValidationError{},
		},
		{
			name: "Error fetching user",
			input: service.UserLogin{
				Email:    "test@example.com",
				Password: "password123",
			},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, "test@example.com").Return(nil, errors.New("db error")).Once()
			},
			mockHashSetup:     func(mhh *MockHashHelper) {},
			expectedUser:      nil,
			expectedErrorType: errors.New(""), // Generic error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockHash := new(MockHashHelper)

			tt.mockRepoSetup(mockRepo)
			tt.mockHashSetup(mockHash)

			userService := service.NewUserService(mockRepo, mockHash)
			user, err := userService.LoginUser(ctx, tt.input)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.ErrorAs(t, err, &validationErr)
				} else if errors.Is(tt.expectedErrorType, apperrors.ErrNotFound) {
					assert.True(t, errors.Is(err, apperrors.ErrNotFound))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
			mockHash.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()

	mockEmail := "tester@example.com"
	mockId := 1
	mockName := "tester"
	mockHashPassword := "hashedpassword"

	tests := []struct {
		name              string
		input             string
		mockRepoSetup     func(*MockUserRepository)
		mockHashSetup     func(*MockHashHelper)
		expectedUser      *service.User
		expectedErrorType error
	}{
		{
			name:  "Get user Successfully",
			input: mockEmail,
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, mockEmail).Return(&repository.User{
					Id:           mockId,
					Name:         mockName,
					Email:        mockEmail,
					PasswordHash: mockHashPassword,
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}, nil).Once()
			},
			mockHashSetup: func(mhh *MockHashHelper) {
				// don't need to set
			},
			expectedUser: &service.User{
				Id:    mockId,
				Name:  mockName,
				Email: mockEmail,
			},
			expectedErrorType: nil,
		},
		{
			name:          "failed to get non-existent user",
			input:         "non-existent@email.com",
			mockHashSetup: func(mhh *MockHashHelper) {},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, "non-existent@email.com").Return(
					nil, apperrors.ErrNotFound,
				)
			},
			expectedUser:      nil,
			expectedErrorType: apperrors.ErrNotFound,
		},
		{
			name:          "database error",
			input:         mockEmail,
			mockHashSetup: func(mhh *MockHashHelper) {},
			mockRepoSetup: func(mur *MockUserRepository) {
				mur.On("GetUserByEmail", ctx, mockEmail).Return(
					nil, errors.New("database failed to connect"),
				)
			},
			expectedUser:      nil,
			expectedErrorType: errors.New(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			mockHash := new(MockHashHelper)

			tt.mockRepoSetup(mockRepo)
			tt.mockHashSetup(mockHash)

			userService := service.NewUserService(mockRepo, mockHash)
			user, err := userService.GetUser(ctx, tt.input)

			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				var validationErr *apperrors.ValidationError
				if errors.As(err, &validationErr) {
					assert.ErrorAs(t, err, &validationErr)
				} else if errors.Is(tt.expectedErrorType, apperrors.ErrNotFound) {
					assert.True(t, errors.Is(err, apperrors.ErrNotFound))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
			}

			mockRepo.AssertExpectations(t)
			mockHash.AssertExpectations(t)
		})
	}
}

// TestValidate_UserSignup tests the Validate method of UserSignup
func TestValidate_UserSignup(t *testing.T) {
	tests := []struct {
		name              string
		input             service.UserSignup
		expectedErrorType error
	}{
		{
			name: "Valid UserSignup",
			input: service.UserSignup{
				Name:     "Valid Name",
				Email:    "valid@example.com",
				Password: "password123",
			},
			expectedErrorType: nil,
		},
		{
			name: "Name too short",
			input: service.UserSignup{
				Name:     "",
				Email:    "valid@example.com",
				Password: "password123",
			},
			expectedErrorType: apperrors.NewValidationError(apperrors.INVALID_NAME, ""),
		},
		{
			name: "Name too long",
			input: service.UserSignup{
				Name:     "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", // 101 chars
				Email:    "valid@example.com",
				Password: "password123",
			},
			expectedErrorType: apperrors.NewValidationError(apperrors.INVALID_NAME, ""),
		},
		{
			name: "Invalid email format",
			input: service.UserSignup{
				Name:     "Valid Name",
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedErrorType: apperrors.NewValidationError(apperrors.INVALID_EMAIL, ""),
		},
		{
			name: "Password too short",
			input: service.UserSignup{
				Name:     "Valid Name",
				Email:    "valid@example.com",
				Password: "short", // less than 8
			},
			expectedErrorType: apperrors.NewValidationError(apperrors.INVALID_PASSWORD, ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.expectedErrorType != nil {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedErrorType, err)
				// Optionally, check the error code
				if ve, ok := err.(*apperrors.ValidationError); ok {
					expectedVE := tt.expectedErrorType.(*apperrors.ValidationError)
					assert.Equal(t, expectedVE.Field, ve.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
