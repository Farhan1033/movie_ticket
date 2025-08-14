package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleResponse struct {
	ID        uuid.UUID `json:"id"`
	MovieID   uuid.UUID `json:"movie_id"`
	StudioID  uuid.UUID `json:"studio_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Price     int       `json:"price"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
