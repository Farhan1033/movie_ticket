package services

import (
	"movie-ticket/internal/movie_module/entities"
	"movie-ticket/internal/movie_module/repositories"

	"github.com/google/uuid"
)

type MoviesService interface {
	CreateMovies(input *entities.Movies) error
	GetMovies() ([]entities.Movies, error)
	GetMovieById(id uuid.UUID) ([]entities.Movies, error)
	UpdateMovies(id uuid.UUID, input *entities.Movies) error
	DeleteMovie(id uuid.UUID) error
}

type movieSvc struct {
	repo repositories.MovieRepository
}

func NewMoviesService(r repositories.MovieRepository) MoviesService {
	return &movieSvc{repo: r}
}


