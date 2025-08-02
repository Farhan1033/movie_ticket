package dto

import "errors"

var (
	ErrMovieNotFound  = errors.New("movie not found")
	ErrMovieExists    = errors.New("movie with this title already exists")
	ErrInvalidInput   = errors.New("invalid input data")
	ErrDatabaseError  = errors.New("database operation failed")
	ErrInvalidMovieId = errors.New("invalid movie id format")
)
