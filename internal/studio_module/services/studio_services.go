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
	GetById(id string) (*dto.StudioResponse, error)
	Update(role, id string, input *dto.UpdateStudioRequest) (*dto.StudioResponse, error)
	Delete(role, id string) error
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

func (s *studioSvc) GetByName(name string) (*dto.StudioResponse, error) {
	if name == "" {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	studio, err := s.repo.GetByName(name)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if studio == nil {
		return nil, fmt.Errorf("%w", customerror.ErrStudioNotFound)
	}

	return s.toStudioResponse(studio), nil
}

func (s *studioSvc) GetById(id string) (*dto.StudioResponse, error) {
	idParse, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	studio, err := s.repo.GetById(idParse)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if studio == nil {
		return nil, fmt.Errorf("%w", customerror.ErrStudioNotFound)
	}

	return s.toStudioResponse(studio), nil
}

func (s *studioSvc) Update(role, id string, input *dto.UpdateStudioRequest) (*dto.StudioResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if input == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	idParse, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidStudioId)
	}

	if err := s.validate.Struct(input); err != nil {
		return nil, s.formatValidationError(err)
	}

	existingStudio, err := s.repo.GetById(idParse)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingStudio == nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrStudioNotFound, existingStudio)
	}

	updateStudio := *existingStudio
	s.applyUpdates(&updateStudio, input)
	updateStudio.Updated_At = time.Now()

	if err := s.repo.Update(idParse, &updateStudio); err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return s.toStudioResponse(&updateStudio), nil

}
func (s *studioSvc) Delete(role, id string) error {
	if role != "admin" {
		return fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	idParse, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w", customerror.ErrInvalidStudioId)
	}

	studio, err := s.repo.GetById(idParse)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if studio == nil {
		return fmt.Errorf("%w", customerror.ErrStudioNotFound)
	}

	deleteStudio := s.repo.Delete(studio.ID)
	if deleteStudio != nil {
		return fmt.Errorf("%w", customerror.ErrDatabaseError)
	}

	return nil
}

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

func (s *studioSvc) applyUpdates(studio *entities.Studio, req *dto.UpdateStudioRequest) {
	if req.Name != nil {
		studio.Name = strings.TrimSpace(*req.Name)
	}
	if req.Location != nil {
		studio.Location = strings.TrimSpace(*req.Location)
	}
	if req.Seat_Capacity != nil {
		studio.Seat_Capacity = *req.Seat_Capacity
	}
}
