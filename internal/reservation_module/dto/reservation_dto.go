package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateReservationRequest struct {
	ScheduleID string   `json:"schedule_id" validate:"required"`
	Seats      []string `json:"seats" validate:"required"`
	TotalPrice int      `json:"total_price" validate:"required"`
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

type ReservationHistory struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	ScheduleID uuid.UUID `json:"schedule_id"`
	TotalPrice int       `json:"total_price"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	ExpiresAt  time.Time `json:"expires_at"`

	// Schedule
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Price     int    `json:"price"`

	// Movie
	MovieTitle  string `json:"movie_title"`
	MovieGenre  string `json:"movie_genre"`
	MoviePoster string `json:"movie_poster"`

	// Studio
	StudioName     string `json:"studio_name"`
	StudioLocation string `json:"studio_location"`

	// Seats (akan diisi manual setelah query kedua)
	Seats []string `json:"seats"`
}
