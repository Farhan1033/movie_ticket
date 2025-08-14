package customerrors

import "errors"

var (
	ErrScheduleNotFound  = errors.New("schedule not found")
	ErrScheduleExists    = errors.New("schedule with this title already exists")
	ErrInvalidInput      = errors.New("invalid input data")
	ErrDatabaseError     = errors.New("database operation failed")
	ErrInvalidScheduleId = errors.New("invalid schedule id format")
	ErrUnauthorizedUser  = errors.New("forbidden user")
	ErrTimeStart         = errors.New("the start time must not be earlier than the end time")
	ErrInactiveMovie     = errors.New("unable to create a schedule because the movie is inactive")
)
