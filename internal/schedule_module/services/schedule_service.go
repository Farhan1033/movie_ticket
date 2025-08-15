package services

import (
	"fmt"
	movieError "movie-ticket/internal/movie_module/custom_error"
	movie "movie-ticket/internal/movie_module/repositories"
	customerror "movie-ticket/internal/schedule_module/custom_errors"
	"movie-ticket/internal/schedule_module/dto"
	"movie-ticket/internal/schedule_module/entities"
	"movie-ticket/internal/schedule_module/repositories"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ScheduleServices interface {
	Create(role string, req *dto.ScheduleCreateRequest) (*dto.ScheduleResponse, error)
}

type svcSchedule struct {
	repo      repositories.ScheduleRepository
	Validate  *validator.Validate
	movieRepo movie.MovieRepository
}

func NewShceduleSvc(r repositories.ScheduleRepository) ScheduleServices {
	return &svcSchedule{
		repo:      r,
		Validate:  validator.New(),
		movieRepo: movie.NewMovieRepo(),
	}
}

func (svc *svcSchedule) Create(role string, req *dto.ScheduleCreateRequest) (*dto.ScheduleResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if req == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	layout := "15:04" // format jam:menit

	start, err := time.Parse(layout, req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format")
	}

	end, err := time.Parse(layout, req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format")
	}

	if !start.Before(end) {
		return nil, fmt.Errorf("%w", customerror.ErrTimeStart)
	}

	if start.After(end) || start.Equal(end) {
		return nil, fmt.Errorf("%w", customerror.ErrTimeStart)
	}

	checkMovie, err := svc.movieRepo.GetMovieById(req.MovieID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", movieError.ErrDatabaseError, err)
	}

	if checkMovie == nil {
		return nil, fmt.Errorf("%w", movieError.ErrMovieNotFound)
	}

	if !checkMovie.Status {
		return nil, fmt.Errorf("%w", customerror.ErrInactiveMovie)
	}

	schedule := &entities.Schedules{
		ID:        uuid.New(),
		MovieID:   req.MovieID,
		StudioID:  req.StudioID,
		StartTime: start.Format(layout),
		EndTime:   end.Format(layout),
		Price:     req.Price,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := svc.repo.Create(schedule); err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return svc.toScheduleResponse(schedule), nil
}

// Helper
func (svc *svcSchedule) toScheduleResponse(model *entities.Schedules) *dto.ScheduleResponse {
	return &dto.ScheduleResponse{
		ID:        model.ID,
		MovieID:   model.MovieID,
		StudioID:  model.StudioID,
		StartTime: model.StartTime,
		EndTime:   model.EndTime,
		Price:     model.Price,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
