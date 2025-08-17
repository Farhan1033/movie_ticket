package services

import (
	"fmt"
	movieError "movie-ticket/internal/movie_module/custom_error"
	movie "movie-ticket/internal/movie_module/repositories"
	customerror "movie-ticket/internal/schedule_module/custom_errors"
	"movie-ticket/internal/schedule_module/dto"
	"movie-ticket/internal/schedule_module/entities"
	"movie-ticket/internal/schedule_module/repositories"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ScheduleServices interface {
	Create(role string, req *dto.ScheduleCreateRequest) (*dto.ScheduleResponse, error)
	Get() ([]*dto.ScheduleResponse, error)
	GetById(id string) (*dto.ScheduleResponse, error)
	Update(role, id string, req *dto.ScheduleUpdateRequest) (*dto.ScheduleResponse, error)
	Delete(role, id string) error
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

	layout := "15:04:05" // format jam:menit:detik

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

	if req.Price <= 0 {
		return nil, fmt.Errorf("%w", customerror.ErrPriceInput)
	}

	existingSchedules, err := svc.repo.GetSchedulesByStudioID(req.StudioID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", movieError.ErrDatabaseError, err)
	}

	if existingSchedules != nil && len(existingSchedules) > 0 {
		for _, s := range existingSchedules {
			existingStart, err := time.Parse(layout, s.StartTime)
			if err != nil {
				return nil, fmt.Errorf("invalid existing start time format: %v", err)
			}

			existingEnd, err := time.Parse(layout, s.EndTime)
			if err != nil {
				return nil, fmt.Errorf("invalid existing end time format: %v", err)
			}

			if (start.Before(existingEnd) && end.After(existingStart)) ||
				(start.Equal(existingStart) || end.Equal(existingEnd)) {
				return nil, fmt.Errorf("%w", customerror.ErrScheduleConflict)
			}
		}
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

func (svc *svcSchedule) Get() ([]*dto.ScheduleResponse, error) {
	schedules, err := svc.repo.Get()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err.Error())
	}

	if len(schedules) == 0 {
		return nil, fmt.Errorf("%w", customerror.ErrScheduleNotFound)
	}

	response := make([]*dto.ScheduleResponse, len(schedules))
	for i, schedule := range schedules {
		response[i] = svc.toScheduleResponse(&schedule)
	}

	return response, nil
}

func (svc *svcSchedule) GetById(id string) (*dto.ScheduleResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidScheduleId)
	}

	idParse, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidScheduleId)
	}

	schedule, err := svc.repo.GetById(idParse)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err.Error())
	}
	if schedule == nil {
		return nil, fmt.Errorf("%w", customerror.ErrScheduleNotFound)
	}

	return svc.toScheduleResponse(schedule), nil
}

func (svc *svcSchedule) Update(role, id string, req *dto.ScheduleUpdateRequest) (*dto.ScheduleResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if id == "" {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	idParse, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidScheduleId)
	}

	layout := "15:04:05" // format jam:menit:detik

	start, err := time.Parse(layout, *req.StartTime)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format")
	}

	end, err := time.Parse(layout, *req.EndTime)
	if err != nil {
		return nil, fmt.Errorf("invalid end time format")
	}

	if !start.Before(end) {
		return nil, fmt.Errorf("%w", customerror.ErrTimeStart)
	}

	if start.After(end) || start.Equal(end) {
		return nil, fmt.Errorf("%w", customerror.ErrTimeStart)
	}

	existingSchedule, err := svc.repo.GetById(idParse)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingSchedule == nil {
		return nil, fmt.Errorf("%w", customerror.ErrScheduleNotFound)
	}

	existingShedules, err := svc.repo.GetSchedulesByStudioID(*req.StudioID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", movieError.ErrDatabaseError, err)
	}

	if existingShedules == nil {
		return nil, fmt.Errorf("%w", movieError.ErrMovieNotFound)
	}

	for _, s := range existingShedules {
		if s.ID == idParse {
			continue
		}

		existingStart, _ := time.Parse(layout, s.StartTime)
		existingEnd, _ := time.Parse(layout, s.EndTime)

		if start.Before(existingEnd) && end.After(existingStart) {
			return nil, fmt.Errorf("%w", customerror.ErrScheduleConflict)
		}
	}

	if *req.Price < 0 {
		return nil, fmt.Errorf("%w", customerror.ErrPriceInput)
	}

	scheduleUpdate := *existingSchedule
	svc.applyUpdates(&scheduleUpdate, req)
	scheduleUpdate.UpdatedAt = time.Now()

	if err := svc.repo.Update(idParse, &scheduleUpdate); err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return svc.toScheduleResponse(&scheduleUpdate), nil
}

func (svc *svcSchedule) Delete(role, id string) error {
	if role != "admin" {
		return fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if id == "" {
		return fmt.Errorf("%w", customerror.ErrInvalidScheduleId)
	}

	idParse, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w", customerror.ErrInvalidScheduleId)
	}

	existingSchedule, err := svc.repo.GetById(idParse)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingSchedule == nil {
		return fmt.Errorf("%w", customerror.ErrScheduleNotFound)
	}

	if err := svc.repo.Delete(idParse); err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return nil
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

func (svc *svcSchedule) applyUpdates(schedule *entities.Schedules, req *dto.ScheduleUpdateRequest) {
	if req.MovieID != nil {
		schedule.MovieID = *req.MovieID
	}
	if req.StudioID != nil {
		schedule.StudioID = *req.StudioID
	}
	if req.StartTime != nil {
		schedule.StartTime = strings.TrimSpace(*req.StartTime)
	}
	if req.EndTime != nil {
		schedule.EndTime = strings.TrimSpace(*req.EndTime)
	}
	if req.Price != nil {
		schedule.Price = *req.Price
	}
}
