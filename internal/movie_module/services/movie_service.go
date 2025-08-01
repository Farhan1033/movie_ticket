package services

import (
	"errors"
	"fmt"
	"movie-ticket/internal/movie_module/dto"
	"movie-ticket/internal/movie_module/entities"
	"movie-ticket/internal/movie_module/repositories"
	"net/url"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type MoviesService interface {
	CreateMovie(req *dto.CreateMovieRequest) (*dto.MovieResponse, error)
}

type movieSvc struct {
	repo      repositories.MovieRepository
	validator *validator.Validate
}

func NewMoviesService(r repositories.MovieRepository) MoviesService {
	return &movieSvc{
		repo:      r,
		validator: validator.New(),
	}
}

func (s *movieSvc) CreateMovie(req *dto.CreateMovieRequest) (*dto.MovieResponse, error) {
	if req == nil {
		return nil, dto.ErrInvalidInput
	}

	if err := s.validator.Struct(req); err != nil {
		return nil, s.formatValidationError(err)
	}

	if err := s.validateBusinessRules(req); err != nil {
		return nil, err
	}

	existing, _ := s.repo.GetByTitle(req.Title)
	if existing != nil {
		return nil, dto.ErrMovieExists
	}

	movie := &entities.Movies{
		ID:               uuid.New(),
		Title:            strings.TrimSpace(req.Title),
		Description:      strings.TrimSpace(req.Description),
		Genre:            strings.TrimSpace(req.Genre),
		Duration_Minutes: req.Duration_Minutes,
		Rating:           req.Rating,
		Poster_Url:       req.Poster_Url,
		Created_At:       time.Now(),
		Updated_At:       time.Now(),
	}

	if err := s.repo.CreateMovies(movie); err != nil {
		return nil, fmt.Errorf("%w: %v", dto.ErrDatabaseError, err)
	}

	return s.toMovieResponse(movie), nil
}

func (s *movieSvc) validateBusinessRules(req *dto.CreateMovieRequest) error {
	if _, err := url.Parse(req.Poster_Url); err != nil {
		return fmt.Errorf("invalid poster URL format: %w", err)
	}

	if req.Duration_Minutes < 1 || req.Duration_Minutes > 600 {
		return errors.New("duration must be between 1 and 600 minutes")
	}

	return nil
}

func (s *movieSvc) toMovieResponse(movie *entities.Movies) *dto.MovieResponse {
	return &dto.MovieResponse{
		ID:               movie.ID,
		Title:            movie.Title,
		Description:      movie.Description,
		Genre:            movie.Genre,
		Duration_Minutes: movie.Duration_Minutes,
		Rating:           movie.Rating,
		Poster_Url:       movie.Poster_Url,
		Created_At:       movie.Created_At,
		Updated_At:       movie.Updated_At,
	}
}

func (s *movieSvc) formatValidationError(err error) error {
	var errorMessages []string

	for _, err := range err.(validator.ValidationErrors) {
		switch err.Tag() {
		case "required":
			errorMessages = append(errorMessages, fmt.Sprintf("%s is required", strings.ToLower(err.Field())))
		case "min":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at least %s characters/value", strings.ToLower(err.Field()), err.Param()))
		case "max":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be at most %s characters/value", strings.ToLower(err.Field()), err.Param()))
		case "url":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be a valid URL", strings.ToLower(err.Field())))
		case "oneof":
			errorMessages = append(errorMessages, fmt.Sprintf("%s must be one of: %s", strings.ToLower(err.Field()), err.Param()))
		default:
			errorMessages = append(errorMessages, fmt.Sprintf("%s is invalid", strings.ToLower(err.Field())))
		}
	}

	return fmt.Errorf("validation failed: %s", strings.Join(errorMessages, ", "))
}
