package repositories

import (
	"fmt"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/schedule_module/entities"

	"github.com/google/uuid"
)

type ScheduleRepository interface {
	Create(req *entities.Schedules) error
	Get() ([]entities.Schedules, error)
	GetById(id uuid.UUID) (*entities.Schedules, error)
	Update(id uuid.UUID, req *entities.Schedules) error
	Delete(id uuid.UUID) error
	GetSchedulesByStudioID(studioID uuid.UUID) ([]*entities.Schedules, error)
}

type scheduleRepo struct{}

func NewScheduleRepo() ScheduleRepository {
	return &scheduleRepo{}
}

func (repo *scheduleRepo) Create(req *entities.Schedules) error {
	if err := postgres.DB.Preload("Movie").Preload("Studio").Create(req).Error; err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}

func (repo *scheduleRepo) Get() ([]entities.Schedules, error) {
	var schedules []entities.Schedules

	err := postgres.DB.Preload("Movie").Preload("Studio").Find(&schedules).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	return schedules, nil
}

func (repo *scheduleRepo) GetById(id uuid.UUID) (*entities.Schedules, error) {
	var schedule entities.Schedules

	err := postgres.DB.Preload("Movie").Preload("Studio").Where("id = ?", id).Find(&schedule).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	return &schedule, nil
}

func (repo *scheduleRepo) Update(id uuid.UUID, req *entities.Schedules) error {
	update := map[string]interface{}{
		"movie_id":   req.MovieID,
		"studio_id":  req.StudioID,
		"start_time": req.StartTime,
		"end_time":   req.EndTime,
		"price":      req.Price,
		"created_at": req.CreatedAt,
		"updated_at": req.UpdatedAt,
	}

	updateSchedule := postgres.DB.Model(&entities.Schedules{}).Preload("Movie").Preload("Studio").Where("id = ?", id).Updates(update)

	if updateSchedule.Error != nil {
		return fmt.Errorf("failed to updated schedule: %w", updateSchedule.Error)
	}

	if updateSchedule.RowsAffected == 0 {
		return fmt.Errorf("no data updated: %v", updateSchedule)
	}

	return nil
}

func (repo *scheduleRepo) Delete(id uuid.UUID) error {
	if err := postgres.DB.Delete(&entities.Schedules{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete schedule: %w", err)
	}

	return nil
}

func (repo *scheduleRepo) GetSchedulesByStudioID(studioID uuid.UUID) ([]*entities.Schedules, error) {
	var schedulesByStudioId []*entities.Schedules

	err := postgres.DB.Preload("Studio").Where("studio_id = ?", studioID).
		Order("start_time ASC").
		Find(&schedulesByStudioId).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	return schedulesByStudioId, nil
}
