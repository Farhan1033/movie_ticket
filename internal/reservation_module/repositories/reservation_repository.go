package repository

import (
	"context"
	"movie-ticket/internal/reservation_module/entities"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReservationRepository interface {
	Create(ctx context.Context, reservation *entities.Reservation, seats []string) error
	UpdateStatus(ctx context.Context, reservationID uuid.UUID, status entities.ReservationStatus) error
	FindByID(ctx context.Context, id uuid.UUID) (*entities.Reservation, error)
	FindExpiredReservations(ctx context.Context) ([]*entities.Reservation, error)
	UpdateExpiredReservations(ctx context.Context) error
}

type reservationRepository struct {
	db *gorm.DB
}

func NewReservationRepository(db *gorm.DB) ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) Create(ctx context.Context, reservation *entities.Reservation, seats []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create reservation
		if err := tx.Create(reservation).Error; err != nil {
			return err
		}

		// Create seat entities
		var seatEntities []entities.ReservationSeat
		for _, seat := range seats {
			seatEntities = append(seatEntities, entities.ReservationSeat{
				ReservationID: reservation.ID,
				SeatCode:      seat,
			})
		}

		if len(seatEntities) > 0 {
			if err := tx.Create(&seatEntities).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *reservationRepository) UpdateStatus(ctx context.Context, reservationID uuid.UUID, status entities.ReservationStatus) error {
	result := r.db.WithContext(ctx).Model(&entities.Reservation{}).
		Where("id = ?", reservationID).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *reservationRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Reservation, error) {
	var reservation entities.Reservation
	if err := r.db.WithContext(ctx).Preload("Seats").First(&reservation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &reservation, nil
}

func (r *reservationRepository) FindExpiredReservations(ctx context.Context) ([]*entities.Reservation, error) {
	var reservations []*entities.Reservation
	err := r.db.WithContext(ctx).
		Preload("Seats").
		Where("status = ? AND expires_at < ?", entities.StatusPending, time.Now()).
		Find(&reservations).Error

	return reservations, err
}

func (r *reservationRepository) UpdateExpiredReservations(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&entities.Reservation{}).
		Where("status = ? AND expires_at < ?", entities.StatusPending, time.Now()).
		Update("status", entities.StatusExpired).Error
}
