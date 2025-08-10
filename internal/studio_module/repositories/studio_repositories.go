package repositories

import (
	"errors"
	"fmt"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/studio_module/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudioRepository interface {
	Create(input *entities.Studio) error
	Get() ([]entities.Studio, error)
	GetByName(name string) (*entities.Studio, error)
	GetById(id uuid.UUID) (*entities.Studio, error)
	Update(id uuid.UUID, input *entities.Studio) error
	Delete(id uuid.UUID) error
}

type studioRepo struct{}

func NewStudioRepo() StudioRepository {
	return &studioRepo{}
}

func (r *studioRepo) Create(input *entities.Studio) error {
	if err := postgres.DB.Create(&input).Error; err != nil {
		return fmt.Errorf("failed to create studio: %w", err)
	}

	return nil
}

func (r *studioRepo) Get() ([]entities.Studio, error) {
	var studios []entities.Studio

	err := postgres.DB.Find(studios).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("data not found: %w", err)
	}

	return studios, nil
}

func (r *studioRepo) GetByName(name string) (*entities.Studio, error) {
	var studio *entities.Studio

	err := postgres.DB.Where("name = ?", name).First(studio).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("data not found: %w", err)
	}

	return studio, nil
}

func (r *studioRepo) GetById(id uuid.UUID) (*entities.Studio, error) {
	var studio *entities.Studio

	err := postgres.DB.Where("id = ?", id).First(studio).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("data not found: %w", err)
	}

	return studio, nil
}

func (r *studioRepo) Update(id uuid.UUID, input *entities.Studio) error {
	updates := map[string]interface{}{
		"name":          input.Name,
		"seat_capacity": input.Seat_Capacity,
		"location":      input.Location,
		"updated_at":    time.Now(),
	}

	result := postgres.DB.Model(&entities.Studio{}).
		Where("id = ?", input.ID).
		Updates(updates)

	if result.Error != nil {
		return fmt.Errorf("failed during data update: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("no changed data")
	}

	return nil
}

func (r *studioRepo) Delete(id uuid.UUID) error {
	if err := postgres.DB.Delete(&entities.Studio{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete studio: %w", err)
	}

	return nil
}
