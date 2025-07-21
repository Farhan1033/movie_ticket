package repositories

import (
	"context"
	"errors"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/movie_module/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MovieRepository interface {
	CreateMovies(input *entities.Movies) error
	GetMovies() ([]entities.Movies, error)
	GetMovieById(id uuid.UUID) ([]entities.Movies, error)
	GetComingSoonMovies(ctx context.Context, limit int) ([]entities.Movies, error)
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

func (r *movieRepo) GetMovieById(id uuid.UUID) ([]entities.Movies, error) {
	var movies []entities.Movies

	err := postgres.DB.Where("id = ?", id).Find(&movies).Error

	if err != nil {
		return nil, errors.New("no available movies")
	}

	return movies, nil
}

func (r *movieRepo) GetComingSoonMovies(ctx context.Context, limit int) ([]entities.Movies, error) {
	var movies []entities.Movies

	query := postgres.DB.WithContext(ctx).
		Where("is_active = ?", true).
		Where("release_date > ?", time.Now()).
		Order("release_date ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&movies).Error

	return movies, err
}
