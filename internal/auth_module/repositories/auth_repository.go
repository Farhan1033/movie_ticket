package repositories

import (
	"errors"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/auth_module/entities"

	"gorm.io/gorm"
)

type AuthRepository interface {
	Create(user *entities.User) error
	FindByEmail(email string) (*entities.User, error)
}

type authRepo struct{}

func NewAuthRepo() AuthRepository {
	return &authRepo{}
}

func (r *authRepo) Create(user *entities.User) error {
	return postgres.DB.Create(user).Error
}

func (r *authRepo) FindByEmail(email string) (*entities.User, error) {
	var user entities.User

	err := postgres.DB.Where("email = ?", email).First(&user).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("email not found")
	}

	return &user, err
}
