package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleCreateRequest struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	MovieID   uuid.UUID `json:"movie_id" validate:"required"`
	StudioID  uuid.UUID `json:"studio_id" validate:"required"`
	StartTime string    `json:"start_time" validate:"required"`
	EndTime   string    `json:"end_time" validate:"required"`
	Price     int       `json:"price" validate:"required,min=1,max=255"`
	CreatedAt time.Time `json:"created_at" validate:"required"`
	UpdatedAt time.Time `json:"updated_at" validate:"required"`
}

type ScheduleUpdateRequest struct {
	MovieID   *uuid.UUID `json:"movie_id,omitempty" validate:"omitempty"`
	StudioID  *uuid.UUID `json:"studio_id,omitempty" validate:"omitempty"`
	StartTime *string    `json:"start_time,omitempty" validate:"omitempty"`
	EndTime   *string    `json:"end_time,omitempty" validate:"omitempty"`
	Price     *int       `json:"price,omitempty" validate:"omitempty,min=1,max=255"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" validate:"omitempty"`
}
