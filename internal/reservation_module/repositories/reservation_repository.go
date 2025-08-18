package repository

import (
	"context"
	"movie-ticket/internal/reservation_module/dto"
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
	HistoryReservations(ctx context.Context, userID uuid.UUID) ([]*dto.ReservationHistory, error)
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

func (r *reservationRepository) HistoryReservations(ctx context.Context, userID uuid.UUID) ([]*dto.ReservationHistory, error) {
	var reservations []*dto.ReservationHistory

	// Tahap 1: Ambil reservasi (tanpa seat)
	queryReservations := `
		SELECT 
			r.id,
			r.user_id,
			r.schedule_id,
			r.total_price,
			r.status,
			r.created_at,
			r.updated_at,
			r.expires_at,

			s.start_time,
			s.end_time,
			s.price,

			m.title AS movie_title,
			m.genre AS movie_genre,
			m.poster_url AS movie_poster,

			st.name AS studio_name,
			st.location AS studio_location
		FROM reservations r
		JOIN schedules s ON r.schedule_id = s.id
		JOIN movies m ON s.movie_id = m.id
		JOIN studios st ON s.studio_id = st.id
		WHERE r.user_id = ? 
		  AND r.status IN ('PENDING', 'CANCEL', 'PAID')
		ORDER BY r.created_at DESC;
	`

	if err := r.db.WithContext(ctx).Raw(queryReservations, userID).Scan(&reservations).Error; err != nil {
		return nil, err
	}

	if len(reservations) == 0 {
		return reservations, nil
	}

	var reservationIDs []uuid.UUID
	for _, res := range reservations {
		reservationIDs = append(reservationIDs, res.ID)
	}

	type SeatRow struct {
		ReservationID uuid.UUID
		SeatCode      string
	}
	var seatRows []SeatRow

	querySeats := `
		SELECT reservation_id, seat_code
		FROM reservation_seats
		WHERE reservation_id IN ?;
	`

	if err := r.db.WithContext(ctx).Raw(querySeats, reservationIDs).Scan(&seatRows).Error; err != nil {
		return nil, err
	}

	seatsMap := make(map[uuid.UUID][]string)
	for _, seat := range seatRows {
		seatsMap[seat.ReservationID] = append(seatsMap[seat.ReservationID], seat.SeatCode)
	}

	for _, res := range reservations {
		if seatList, ok := seatsMap[res.ID]; ok {
			res.Seats = seatList
		}
	}

	return reservations, nil
}
