package repositories

import (
	"fmt"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/schedule_module/entities"
)

type ScheduleRepository interface {
	Create(req *entities.Schedules) error
}

type scheduleRepo struct{}

func NewScheduleRepo() ScheduleRepository {
	return &scheduleRepo{}
}

func (repo *scheduleRepo) Create(req *entities.Schedules) error {
	if err := postgres.DB.Create(req).Error; err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	return nil
}
