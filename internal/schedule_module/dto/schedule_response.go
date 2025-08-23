package dto

import (
	"time"

	"github.com/google/uuid"
)

type ScheduleResponse struct {
	ID             uuid.UUID `json:"id"`
	MovieTitle     string    `json:"movie_title"`
	MovieDesc      string    `jsono:"movie_desc"`
	MovieGenre     string    `json:"movie_genre"`
	MoviePoster    string    `json:"movie_poster"`
	MovieRating    string    `json:"movie_rating"`
	StudioName     string    `json:"studio_name"`
	StudioLocation string    `json:"studio_location"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
	Price          int       `json:"price"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type MessageResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}
