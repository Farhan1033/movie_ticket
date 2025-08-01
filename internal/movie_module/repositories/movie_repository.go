package repositories

import (
	"errors"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/movie_module/entities"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovieRepository interface {
	CreateMovies(input *entities.Movies) error
	GetMovies() ([]entities.Movies, error)
	GetByTitle(input string) ([]entities.Movies, error)
	GetMovieById(id uuid.UUID) ([]entities.Movies, error)
	UpdateMovies(id uuid.UUID, input *entities.Movies) error
	DeleteMovie(id uuid.UUID) error
}

type movieRepo struct{}

func NewMovieRepo() MovieRepository {
	return &movieRepo{}
}

func (r *movieRepo) CreateMovies(input *entities.Movies) error {
	return postgres.DB.Create(input).Error
}

func (r *movieRepo) GetMovies() ([]entities.Movies, error) {
	var movies []entities.Movies

	err := postgres.DB.Find(&movies).Error

	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("no available movies")
	}

	return movies, nil
}

func (r *movieRepo) GetByTitle(input string) ([]entities.Movies, error) {
	var movies []entities.Movies

	if err := postgres.DB.Where("title = ?", input).Find(&movies).Error; err != nil {
		return nil, errors.New("no available movies")
	}

	return movies, nil
}

func (r *movieRepo) GetMovieById(id uuid.UUID) ([]entities.Movies, error) {
	var movies []entities.Movies

	err := postgres.DB.Where("id = ?", id).Find(&movies).Error

	if err != nil {
		return nil, errors.New("no available movies")
	}

	return movies, nil
}

func (r *movieRepo) UpdateMovies(id uuid.UUID, input *entities.Movies) error {
	updates := map[string]interface{}{
		"title":            input.Title,
		"description":      input.Description,
		"genre":            input.Genre,
		"duration_minutes": input.Duration_Minutes,
		"rating":           input.Rating,
		"postre_url":       input.Poster_Url,
	}

	result := postgres.DB.Model(&entities.Movies{}).
		Where("id = ?", id).
		Updates(updates)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("failed when updating data")
	}

	return nil
}

func (r *movieRepo) DeleteMovie(id uuid.UUID) error {
	return postgres.DB.Delete(&entities.Movies{}, "id = ?", id).Error
}
