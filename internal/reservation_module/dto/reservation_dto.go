package dto

import (
	"time"

	"github.com/google/uuid"
)

type ReservationCreateRequest struct {
	UserID     uuid.UUID `json:"user_id" validation:"required"`
	ScheduleID uuid.UUID `json:"schedule_id" validation:"required"`
	TotalPrice int       `json:"total_price" validation:"required,min=0"`
	Status     string    `json:"status" validation:"required"` // PENDING | PAID | CANCELED
}

type ReservationResponse struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`
	TotalPrice int       `json:"total_price"`
	Status     string    `json:"status"` // PENDING | PAID | CANCELED
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type MessageResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
}
