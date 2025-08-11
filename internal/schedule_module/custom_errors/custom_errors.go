package customerrors

import "errors"

var (
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrScheduleExists    = errors.New("schedule with this title already exists")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrDatabaseError     = errors.New("database operation failed")
	ErrInvalidScheduleId = errors.New("invalid schedule id format")
	ErrUnauthorizedUser  = errors.New("forbidden user")
)
