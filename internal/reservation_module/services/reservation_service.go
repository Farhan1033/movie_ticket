package service

import (
	"context"
	"errors"
	"fmt"
	"movie-ticket/internal/reservation_module/entities"
	repository "movie-ticket/internal/reservation_module/repositories"
	"time"

	"github.com/google/uuid"
)

type ReservationService interface {
	CreateReservation(ctx context.Context, userID uuid.UUID, scheduleID uuid.UUID, seats []string, totalPrice int) (*entities.Reservation, error)
	ConfirmReservation(ctx context.Context, reservationID uuid.UUID) error
	CancelReservation(ctx context.Context, reservationID uuid.UUID) error
	GetReservation(ctx context.Context, reservationID uuid.UUID) (*entities.Reservation, error)
	CleanupExpiredReservations(ctx context.Context) error
}

type reservationService struct {
	reservationRepo repository.ReservationRepository
	seatRedisRepo   repository.SeatRedisRepository
}

func NewReservationService(resRepo repository.ReservationRepository, redisRepo repository.SeatRedisRepository) ReservationService {
	return &reservationService{
		reservationRepo: resRepo,
		seatRedisRepo:   redisRepo,
	}
}

func (s *reservationService) CreateReservation(ctx context.Context, userID uuid.UUID, scheduleID uuid.UUID, seats []string, totalPrice int) (*entities.Reservation, error) {
	if len(seats) == 0 {
		return nil, errors.New("seats required")
	}

	if totalPrice <= 0 {
		return nil, errors.New("invalid total price")
	}

	// Validate seat codes
	for _, seat := range seats {
		if seat == "" {
			return nil, errors.New("seat code cannot be empty")
		}
	}

	// Hold seats in Redis with 5 minute TTL
	if err := s.seatRedisRepo.HoldSeats(ctx, scheduleID.String(), userID.String(), seats, 5*time.Minute); err != nil {
		return nil, fmt.Errorf("failed to hold seats: %w", err)
	}

	// Create reservation
	reservation := &entities.Reservation{
		UserID:     userID,
		ScheduleID: scheduleID,
		TotalPrice: totalPrice,
		Status:     entities.StatusPending,
		ExpiresAt:  time.Now().Add(5 * time.Minute),
	}

	if err := s.reservationRepo.Create(ctx, reservation, seats); err != nil {
		// Rollback: release seats in Redis
		_ = s.seatRedisRepo.ReleaseSeats(ctx, scheduleID.String(), seats)
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Load the created reservation with seats
	createdReservation, err := s.reservationRepo.FindByID(ctx, reservation.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load created reservation: %w", err)
	}

	return createdReservation, nil
}

func (s *reservationService) ConfirmReservation(ctx context.Context, reservationID uuid.UUID) error {
	// Get reservation
	reservation, err := s.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("reservation not found: %w", err)
	}

	// Check if reservation can be confirmed
	if !reservation.CanTransitionTo(entities.StatusPaid) {
		return fmt.Errorf("cannot confirm reservation with status %s", reservation.Status)
	}

	// Check if reservation is expired
	if reservation.IsExpired() {
		return errors.New("reservation has expired")
	}

	// Extract seat codes
	seatCodes := extractSeatCodes(reservation)

	// Update status in database
	if err := s.reservationRepo.UpdateStatus(ctx, reservationID, entities.StatusPaid); err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Move seats from temporary hold to confirmed in Redis
	if err := s.seatRedisRepo.ConfirmSeats(ctx, reservation.ScheduleID.String(), seatCodes); err != nil {
		// Log error but don't fail the operation as DB is already updated
		// In production, you might want to implement compensation logic here
		fmt.Printf("Warning: failed to confirm seats in Redis: %v\n", err)
	}

	return nil
}

func (s *reservationService) CancelReservation(ctx context.Context, reservationID uuid.UUID) error {
	// Get reservation
	reservation, err := s.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return fmt.Errorf("reservation not found: %w", err)
	}

	// Check if reservation can be canceled
	if !reservation.CanTransitionTo(entities.StatusCanceled) {
		return fmt.Errorf("cannot cancel reservation with status %s", reservation.Status)
	}

	// Extract seat codes
	seatCodes := extractSeatCodes(reservation)

	// Update status in database
	if err := s.reservationRepo.UpdateStatus(ctx, reservationID, entities.StatusCanceled); err != nil {
		return fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Release seats in Redis
	if err := s.seatRedisRepo.ReleaseSeats(ctx, reservation.ScheduleID.String(), seatCodes); err != nil {
		// Log error but don't fail the operation as DB is already updated
		fmt.Printf("Warning: failed to release seats in Redis: %v\n", err)
	}

	return nil
}

func (s *reservationService) GetReservation(ctx context.Context, reservationID uuid.UUID) (*entities.Reservation, error) {
	reservation, err := s.reservationRepo.FindByID(ctx, reservationID)
	if err != nil {
		return nil, fmt.Errorf("reservation not found: %w", err)
	}

	return reservation, nil
}

func (s *reservationService) CleanupExpiredReservations(ctx context.Context) error {
	// Find expired reservations
	expiredReservations, err := s.reservationRepo.FindExpiredReservations(ctx)
	if err != nil {
		return fmt.Errorf("failed to find expired reservations: %w", err)
	}

	// Update expired reservations status
	if err := s.reservationRepo.UpdateExpiredReservations(ctx); err != nil {
		return fmt.Errorf("failed to update expired reservations: %w", err)
	}

	// Release seats in Redis for expired reservations
	for _, reservation := range expiredReservations {
		seatCodes := extractSeatCodes(reservation)
		if len(seatCodes) > 0 {
			if err := s.seatRedisRepo.ReleaseSeats(ctx, reservation.ScheduleID.String(), seatCodes); err != nil {
				fmt.Printf("Warning: failed to release seats for expired reservation %s: %v\n", reservation.ID, err)
			}
		}
	}

	return nil
}

func extractSeatCodes(reservation *entities.Reservation) []string {
	seats := make([]string, 0, len(reservation.Seats))
	for _, seat := range reservation.Seats {
		seats = append(seats, seat.SeatCode)
	}
	return seats
}
