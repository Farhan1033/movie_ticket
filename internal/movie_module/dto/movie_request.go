package dto

import (
	"time"

	"github.com/google/uuid"
)

type CreateMovieRequest struct {
	Title            string `json:"title" validate:"required,min=1,max=200"`
	Description      string `json:"description" validate:"required,min=1,max=200"`
	Genre            string `json:"genre" validate:"required,min=1,max=100"`
	Duration_Minutes int    `json:"duration_minutes" validate:"required,min=1,max=600"`
	Rating           string `json:"rating" validate:"required,oneof=G PG PG-13 R NC-17"`
	Poster_Url       string `json:"poster_url" validate:"required,url"`
}

type UpdateMovieRequest struct {
	Title            *string `json:"title,omitempty" validate:"omitempty,min=1,max=200"`
	Description      *string `json:"description,omitempty" validate:"omitempty,min=1,max=200"`
	Genre            *string `json:"genre,omitempty" validate:"omitempty,min=1,max=100"`
	Duration_Minutes *int    `json:"duration_minutes,omitempty" validate:"omitempty,min=1,max=600"`
	Rating           *string `json:"rating,omitempty" validate:"omitempty,oneof=G PG PG-13 R NC-17"`
	Poster_Url       *string `json:"poster_url,omitempty" validate:"omitempty,url"`
}

// Response DTOs
type MovieResponse struct {
	ID               uuid.UUID `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	Genre            string    `json:"genre"`
	Duration_Minutes int       `json:"duration_minutes"`
	Rating           string    `json:"rating"`
	Poster_Url       string    `json:"poster_url"`
	Created_At       time.Time `json:"created_at"`
	Updated_At       time.Time `json:"updated_at"`
}

type MoviesResponse struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}
