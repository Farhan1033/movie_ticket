package customerror

import "errors"

var (
	ErrStudioNotFound  = errors.New("studio not found")
	ErrStudioExists    = errors.New("studio with this title already exists")
	ErrInvalidInput    = errors.New("invalid input data")
	ErrDatabaseError   = errors.New("database operation failed")
	ErrInvalidStudioId = errors.New("invalid studio id format")
)
