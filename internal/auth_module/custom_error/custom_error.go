package customerror

import "errors"

var (
	ErrEmailExist        = errors.New("email already exists")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrDatabaseError     = errors.New("database operation failed")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrEmailNotFound     = errors.New("user with this email not found!")
	ErrWrongPassword     = errors.New("wrong password")
	ErrFailedSession     = errors.New("failed to store session in redis")
	ErrFailedCreateToken = errors.New("failed to create token")
)
