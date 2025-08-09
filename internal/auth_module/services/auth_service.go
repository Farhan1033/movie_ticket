package services

import (
	"fmt"
	customerror "movie-ticket/internal/auth_module/custom_error"
	"movie-ticket/internal/auth_module/dto"
	"movie-ticket/internal/auth_module/entities"
	"movie-ticket/internal/auth_module/repositories"
	"movie-ticket/internal/middleware"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(user *dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(req *dto.LoginRequest) (*dto.UserResponse, error)
	Logout(tokenString string) error
	RefreshToken(oldToken string) (*dto.UserResponse, error)
}

type authSvc struct {
	repo repositories.AuthRepository
}

func NewAuthSvc(r repositories.AuthRepository) AuthService {
	return &authSvc{repo: r}
}

func (s *authSvc) Register(user *dto.RegisterRequest) (*dto.RegisterResponse, error) {
	if user == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	existingUser, err := s.repo.FindByEmail(user.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingUser != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrEmailNotFound, existingUser)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	user.Password = string(hash)

	users := *&entities.User{
		ID:          uuid.New(),
		Email:       strings.TrimSpace(user.Email),
		Password:    user.Password,
		FullName:    strings.TrimSpace(user.FullName),
		PhoneNumber: user.PhoneNumber,
		Role:        strings.TrimSpace(user.Role),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(&users); err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrDatabaseError)
	}

	return s.responseAuth(&users), nil
}

func (s *authSvc) Login(req *dto.LoginRequest) (*dto.UserResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	existingUser, err := s.repo.FindByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingUser == nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrEmailNotFound, existingUser)
	}

	if bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)) != nil {
		return nil, fmt.Errorf("%w", customerror.ErrWrongPassword)
	}

	token, err := middleware.CreateToken(existingUser.ID, existingUser.Role, existingUser.Email)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrFailedCreateToken, err)
	}

	return &dto.UserResponse{
		Message: "Success login",
		Token:   token,
	}, nil
}

func (s *authSvc) Logout(tokenString string) error {
	// Revoke token dari Redis
	if err := middleware.RevokeToken(tokenString); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}
	return nil
}

func (s *authSvc) RefreshToken(oldToken string) (*dto.UserResponse, error) {
	// Refresh token
	newToken, err := middleware.RefreshToken(oldToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	return &dto.UserResponse{
		Message: "Token refreshed successfully",
		Token:   newToken,
	}, nil
}

func (s *authSvc) responseAuth(user *entities.User) *dto.RegisterResponse {
	return &dto.RegisterResponse{
		ID:          user.ID,
		Email:       user.Email,
		FullName:    user.FullName,
		PhoneNumber: user.PhoneNumber,
		Role:        user.Role,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
