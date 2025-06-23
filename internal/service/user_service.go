package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
	"workout-tracker-api/internal/apperrors"
	"workout-tracker-api/internal/repository"
	"workout-tracker-api/internal/util/encrypt"
)

type User struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserSignup struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (data *UserSignup) Validate() error {
	if len(data.Name) < 1 || len(data.Name) > 100 {
		return apperrors.NewValidationError(apperrors.INVALID_NAME, "Set the name length between 1 and 100")
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,8}$`)
	if !emailRegex.MatchString(data.Email) {
		return apperrors.NewValidationError(apperrors.INVALID_EMAIL, "The email cannot match the format")
	}

	if len(data.Password) < 8 {
		return apperrors.NewValidationError(apperrors.INVALID_PASSWORD, "The password must be at least 8 characters long")
	}

	return nil

}

type UserLogin struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type UserServiceInterface interface {
	SignupUser(ctx context.Context, input UserSignup) (*User, error)
	LoginUser(ctx context.Context, input UserLogin) (*User, error)
	GetUser(ctx context.Context, userEmail string) (*User, error)
}

type UserService struct {
	userRepo repository.UserRepository
	hash     encrypt.HashHelperInterface
}

func NewUserService(ur repository.UserRepository, h encrypt.HashHelperInterface) UserServiceInterface {
	return &UserService{
		userRepo: ur,
		hash:     h,
	}
}

func (s *UserService) SignupUser(ctx context.Context, input UserSignup) (*User, error) {

	if err := input.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate: %w", err)
	}

	exist, err := s.userRepo.ExistUser(ctx, input.Email)

	if err != nil {
		return nil, fmt.Errorf("failed to check user registration status: %w", err)
	}

	if exist {
		return nil, apperrors.NewValidationError(apperrors.INVALID_EMAIL, "email have already registered")
	}

	hashPS, err := s.hash.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	data := repository.UserCreate{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: hashPS,
	}
	createdUser, err := s.userRepo.CreateUser(ctx, data)

	if err != nil {
		return nil, fmt.Errorf("failed to create user for sign up: %w", err)
	}

	result := toServiceUser(createdUser)

	return result, nil
}

func (s *UserService) LoginUser(ctx context.Context, input UserLogin) (*User, error) {

	fetchedUser, err := s.userRepo.GetUserByEmail(ctx, input.Email)

	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	correct := s.hash.CheckPasswordHash(fetchedUser.PasswordHash, input.Password)
	if !correct {
		return nil, apperrors.NewValidationError(apperrors.INVALID_PASSWORD, "invalid password")
	}

	result := toServiceUser(fetchedUser)

	return result, nil
}
func (s *UserService) GetUser(ctx context.Context, userEmail string) (*User, error) {
	userInfo, err := s.userRepo.GetUserByEmail(ctx, userEmail)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	return toServiceUser(userInfo), nil

}

func toServiceUser(ru *repository.User) *User {
	if ru == nil {
		return nil
	}

	return &User{
		Id:        ru.Id,
		Name:      ru.Name,
		Email:     ru.Email,
		CreatedAt: ru.CreatedAt,
		UpdatedAt: ru.UpdatedAt,
	}
}
