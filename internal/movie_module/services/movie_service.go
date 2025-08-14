package services

import (
	"errors"
	"fmt"
	customerror "movie-ticket/internal/movie_module/custom_error"
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
	CreateMovie(role string, req *dto.CreateMovieRequest) (*dto.MovieResponse, error)
	GetMovies(page, limit int) ([]*dto.MovieResponse, error)
	GetMovieById(id string) (*dto.MovieResponse, error)
	UpdateMovie(role, id string, req *dto.UpdateMovieRequest) (*dto.MovieResponse, error)
	DeleteMovie(role, id string) error
	PatchStatus(role, id string, status *dto.StatusMovieRequest) error
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

func (s *movieSvc) GetMovies(page, limit int) ([]*dto.MovieResponse, error) {
	if page < 1 {
		page = 1
	}

	if limit < 1 || limit > 100 {
		limit = 10
	}

	movies, err := s.repo.GetMovies()

	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if len(movies) == 0 {
		return nil, customerror.ErrMovieNotFound
	}

	response := make([]*dto.MovieResponse, len(movies))
	for i, movie := range movies {
		response[i] = s.toMovieResponse(&movie)
	}

	return response, nil
}

func (s *movieSvc) GetMovieById(id string) (*dto.MovieResponse, error) {
	movieId, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidMovieId)
	}

	movie, err := s.repo.GetMovieById(movieId)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrDatabaseError)
	}

	if movie == nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrMovieNotFound, err)
	}

	return s.toMovieResponse(movie), nil
}

func (s *movieSvc) CreateMovie(role string, req *dto.CreateMovieRequest) (*dto.MovieResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	if req == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	if err := s.validator.Struct(req); err != nil {
		return nil, s.formatValidationError(err)
	}

	if err := s.validateBusinessRules(req); err != nil {
		return nil, err
	}

	existing, _ := s.repo.GetByTitle(req.Title)
	if existing != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrMovieExists, existing)
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
		return nil, fmt.Errorf("%w", customerror.ErrDatabaseError)
	}

	return s.toMovieResponse(movie), nil
}

func (s *movieSvc) UpdateMovie(role, id string, req *dto.UpdateMovieRequest) (*dto.MovieResponse, error) {
	if role != "admin" {
		return nil, fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	movieId, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidMovieId)
	}

	if req == nil {
		return nil, fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	if err := s.validator.Struct(req); err != nil {
		return nil, s.formatValidationError(err)
	}

	existingMovie, err := s.repo.GetMovieById(movieId)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if existingMovie == nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrMovieNotFound, existingMovie)
	}

	updateMovie := *existingMovie
	s.applyUpdates(&updateMovie, req)
	updateMovie.Updated_At = time.Now()

	if err := s.validateUpdatedMovie(req); err != nil {
		return nil, err
	}

	if err := s.repo.UpdateMovies(movieId, &updateMovie); err != nil {
		return nil, fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return s.toMovieResponse(&updateMovie), nil
}

func (s *movieSvc) DeleteMovie(role, id string) error {
	if role != "admin" {
		return fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	movieId, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrInvalidMovieId, err)
	}

	movie, err := s.repo.GetMovieById(movieId)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if movie == nil {
		return customerror.ErrMovieNotFound
	}

	if err := s.repo.DeleteMovie(movieId); err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return nil
}

func (s *movieSvc) PatchStatus(role, id string, status *dto.StatusMovieRequest) error {
	if role != "admin" {
		return fmt.Errorf("%w", customerror.ErrUnauthorizedUser)
	}

	parseId, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrInvalidMovieId, err)
	}

	checkMovie, err := s.repo.GetMovieById(parseId)
	if err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	if checkMovie == nil {
		return fmt.Errorf("%w", customerror.ErrMovieNotFound)
	}

	// pastikan status tidak nil
	if status.Status == nil {
		return fmt.Errorf("%w", customerror.ErrInvalidInput)
	}

	if err := s.repo.UpdateStatus(parseId, *status.Status); err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrDatabaseError, err)
	}

	return nil
}

// Helper
func (s *movieSvc) validateBusinessRules(req *dto.CreateMovieRequest) error {
	if _, err := url.Parse(req.Poster_Url); err != nil {
		return fmt.Errorf("%w: %v", customerror.ErrInvalidPosterUrl, err)
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
		Status:           movie.Status,
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

func (s *movieSvc) validateUpdatedMovie(req *dto.UpdateMovieRequest) error {
	if req.Poster_Url != nil {
		if _, err := url.Parse(*req.Poster_Url); err != nil {
			return fmt.Errorf("%w: %v", customerror.ErrInvalidPosterUrl, err)
		}
	}

	return nil
}

func (s *movieSvc) applyUpdates(movie *entities.Movies, req *dto.UpdateMovieRequest) {
	if req.Title != nil {
		movie.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		movie.Description = strings.TrimSpace(*req.Description)
	}
	if req.Genre != nil {
		movie.Genre = strings.TrimSpace(*req.Genre)
	}
	if req.Duration_Minutes != nil {
		movie.Duration_Minutes = *req.Duration_Minutes
	}
	if req.Rating != nil {
		movie.Rating = *req.Rating
	}
	if req.Poster_Url != nil {
		movie.Poster_Url = *req.Poster_Url
	}
}
