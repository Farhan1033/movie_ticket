package services

import (
	"fmt"
	customerror "movie-ticket/internal/studio_module/custom_error"
	"movie-ticket/internal/studio_module/dto"
	"movie-ticket/internal/studio_module/entities"
	"movie-ticket/internal/studio_module/repositories"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type StudioService interface {
	Create(role string, req *dto.CreateStudioRequest) (*dto.StudioResponse, error)
	Get() ([]*dto.StudioResponse, error)
	GetByName(name string) (*dto.StudioResponse, error)
	GetById(id uuid.UUID) (*dto.StudioResponse, error)
	Update(id uuid.UUID, input *entities.Studio) error
	Delete(id uuid.UUID) error
}

type studioSvc struct {
	repo     repositories.StudioRepository
	validate *validator.Validate
}

func NewStudioService(r repositories.StudioRepository) StudioService {
	return &studioSvc{
		repo:     r,
		validate: validator.New(),
	}
}

func (s *studioSvc) Create(role string, req *dto.CreateStudioRequest) (*dto.StudioResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if req == nil {
		return nil, customerror.ErrInvalidInput
	}

	if err := s.validate.Struct(req); err != nil {
		return nil, s.formatValidationError(err)
	}

	existingStudio, err := s.repo.GetByName(req.Name)

	if err != nil {
		return nil, customerror.ErrDatabaseError
	}

	if existingStudio != nil {
		return nil, customerror.ErrStudioExists
	}

	studios := &entities.Studio{
		ID:            uuid.New(),
		Name:          strings.TrimSpace(req.Name),
		Seat_Capacity: req.Seat_Capacity,
		Location:      strings.TrimSpace(req.Location),
		Created_At:    time.Now(),
		Updated_At:    time.Now(),
	}

	if err := s.repo.Create(studios); err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return s.toStudioResponse(studios), nil
}

func (s *studioSvc) Get() ([]*dto.StudioResponse, error) {
	studios, err := s.repo.Get()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if len(studios) == 0 {
		return nil, fmt.Errorf("%w", customerror.ErrStudioNotFound)
	}

	response := make([]*dto.StudioResponse, len(studios))
	for i, studio := range studios {
		response[i] = s.toStudioResponse(&studio)
	}

	return response, nil
}

func (s *studioSvc) GetByName(name string) (*dto.StudioResponse, error)
func (s *studioSvc) GetById(id uuid.UUID) (*dto.StudioResponse, error)
func (s *studioSvc) Update(id uuid.UUID, input *entities.Studio) error
func (s *studioSvc) Delete(id uuid.UUID) error

// Helper Service
func (s *studioSvc) toStudioResponse(studio *entities.Studio) *dto.StudioResponse {
	return &dto.StudioResponse{
		ID:            studio.ID,
		Name:          studio.Name,
		Seat_Capacity: studio.Seat_Capacity,
		Location:      studio.Location,
		Created_At:    studio.Created_At,
		Updated_At:    studio.Updated_At,
	}
}

func (s *studioSvc) formatValidationError(err error) error {
	var errorMessages []string

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required", strings.ToLower(err.Field())))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters/value", strings.ToLower(err.Field()), err.Param()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at most %s characters/value", strings.ToLower(err.Field()), err.Param()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("%s is invalid", strings.ToLower(err.Field())))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
}
