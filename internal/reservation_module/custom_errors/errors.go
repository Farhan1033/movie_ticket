package customerrors

import "errors"

var (
	ErrUnauthorizedUser       = errors.New("unauthorized user")
	ErrOnlyUserCanReserve     = errors.New("only user role can create reservation")
	ErrInvalidInput           = errors.New("invalid input")
	ErrInvalidID              = errors.New("invalid id")
	ErrScheduleNotFound       = errors.New("schedule not found")
	ErrScheduleInactive       = errors.New("schedule is inactive")
	ErrScheduleAlreadyStarted = errors.New("cannot book for a schedule that already started")
	ErrSeatsRequired          = errors.New("at least one seat is required")
	ErrSeatAlreadyBooked      = errors.New("one or more seats already booked")
	ErrSeatOnHold             = errors.New("one or more seats currently on hold")
	ErrDatabaseError          = errors.New("database error")
	ErrReservationNotFound    = errors.New("reservation not found")
	ErrForbidden              = errors.New("forbidden")
	ErrAlreadyPaid            = errors.New("reservation already paid")
	ErrAlreadyCanceled        = errors.New("reservation already canceled")
)
