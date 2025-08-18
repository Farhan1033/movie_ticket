package entities

import (
	user "movie-ticket/internal/auth_module/entities"
	schedule "movie-ticket/internal/schedule_module/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationStatus string

const (
	StatusPending  ReservationStatus = "PENDING"
	StatusPaid     ReservationStatus = "PAID"
	StatusCanceled ReservationStatus = "CANCELED"
	StatusExpired  ReservationStatus = "EXPIRED"
)

// IsValidTransition checks if status transition is valid
func (r ReservationStatus) IsValidTransition(to ReservationStatus) bool {
	switch r {
	case StatusPending:
		return to == StatusPaid || to == StatusCanceled || to == StatusExpired
	case StatusPaid:
		return false // Paid reservations cannot be changed
	case StatusCanceled, StatusExpired:
		return false // Final states
	default:
		return false
	}
}

type Reservation struct {
	ID         uuid.UUID         `gorm:"type:uuid;primaryKey" json:"id"`
	UserID     uuid.UUID         `gorm:"type:uuid;not null" json:"user_id" binding:"required"`
	ScheduleID uuid.UUID         `gorm:"type:uuid;not null" json:"schedule_id" binding:"required"`
	TotalPrice int               `gorm:"not null" json:"total_price" binding:"required"`
	Status     ReservationStatus `gorm:"type:varchar(20);not null;default:'PENDING'" json:"status"`
	CreatedAt  time.Time         `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time         `gorm:"autoUpdateTime" json:"updated_at"`
	ExpiresAt  time.Time         `gorm:"not null" json:"expires_at"`

	User     user.User          `gorm:"foreignKey:UserID;references:ID" json:"user"`
	Schedule schedule.Schedules `gorm:"foreignKey:ScheduleID;references:ID" json:"schedule"`
	Seats    []ReservationSeat  `gorm:"foreignKey:ReservationID;references:ID" json:"seats"`
}

func (Reservation) TableName() string {
	return "reservations"
}

func (r *Reservation) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	if r.Status == "" {
		r.Status = StatusPending
	}
	if r.ExpiresAt.IsZero() {
		r.ExpiresAt = time.Now().Add(5 * time.Minute)
	}
	return nil
}

func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt) && r.Status == StatusPending
}

func (r *Reservation) CanTransitionTo(newStatus ReservationStatus) bool {
	return r.Status.IsValidTransition(newStatus)
}

type ReservationSeat struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	ReservationID uuid.UUID `gorm:"type:uuid;not null" json:"reservation_id"`
	SeatCode      string    `gorm:"type:varchar(10);not null" json:"seat_code"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	Reservation Reservation `gorm:"foreignKey:ReservationID;references:ID" json:"reservation"`
}

func (ReservationSeat) TableName() string {
	return "reservation_seats"
}

func (rs *ReservationSeat) BeforeCreate(tx *gorm.DB) error {
	if rs.ID == uuid.Nil {
		rs.ID = uuid.New()
	}
	return nil
}
