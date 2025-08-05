package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateStudioRequest struct {
	Name          string `json:"name" validate:"required,min=1,max=100"`
	Seat_Capacity int    `json:"seat_capacity" validate:"required,min=1,max=600"`
	Location      string `json:"location" validate:"required,min=1"`
}

type UpdateStudioRequest struct {
	Name          *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Seat_Capacity *int    `json:"seat_capacity,omitempty" validate:"omitempty,min=1,max=600"`
	Location      string  `json:"location,omitempty" validate:"omitempty,min=1"`
}

type StudioResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Seat_Capacity int       `json:"seat_capacity"`
	Location      string    `json:"location"`
	Created_At    time.Time `json:"created_at"`
	Updated_At    time.Time `json:"updated_at"`
}
