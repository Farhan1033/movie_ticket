package services

import (
	"errors"
	"movie-ticket/internal/auth_module/entities"
	"movie-ticket/internal/auth_module/repositories"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(user *entities.User) error
	Login(email, password string) (*entities.User, error)
}

type authSvc struct {
	repo repositories.AuthRepository
}

func NewAuthSvc(r repositories.AuthRepository) AuthService {
	return &authSvc{repo: r}
}

func (s *authSvc) Register(user *entities.User) error {
	existingUser, err := s.repo.FindByEmail(user.Email)
	if err != nil {
		return err
	}

	if existingUser != nil {
		return errors.New("email is already in use")
	}

	user.ID = uuid.New()

	hash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)

	user.Password = string(hash)

	return s.repo.Create(user)
}

func (s *authSvc) Login(email, password string) (*entities.User, error) {
	existingUser, err := s.repo.FindByEmail(email)
	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		return nil, errors.New("email not registered")
	}

	if bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(password)) != nil {
		return nil, errors.New("wrong password")
	}

	return existingUser, nil
}
