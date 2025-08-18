package handler

import (
	"net/http"
	"strings"

	"movie-ticket/internal/reservation_module/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ReservationHandler struct {
	reservationService service.ReservationService
}

func NewReservationHandler(r *gin.RouterGroup, reservationService service.ReservationService) {
	h := ReservationHandler{reservationService: reservationService}
	r.POST("/reservation/create", h.CreateReservation)
	r.PUT("/reservation/:id/confirm", h.ConfirmReservation)
	r.PUT("/reservation/:id/cancel", h.CancelReservation)
	r.GET("/reservation/:id", h.GetReservation)
}

// CreateReservationRequest represents the request payload for creating a reservation
type CreateReservationRequest struct {
	UserID     string   `json:"user_id" binding:"required"`
	ScheduleID string   `json:"schedule_id" binding:"required"`
	Seats      []string `json:"seats" binding:"required,min=1"`
	TotalPrice int      `json:"total_price" binding:"required,min=1"`
}

type ReservationResponse struct {
	ID         string   `json:"id"`
	UserID     string   `json:"user_id"`
	ScheduleID string   `json:"schedule_id"`
	Seats      []string `json:"seats"`
	TotalPrice int      `json:"total_price"`
	Status     string   `json:"status"`
	ExpiresAt  string   `json:"expires_at"`
	CreatedAt  string   `json:"created_at"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ---------------- CREATE RESERVATION ----------------
func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	var req CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// Validate user ID
	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	// Validate schedule ID
	scheduleID, err := uuid.Parse(req.ScheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_schedule_id",
			Message: "Invalid schedule ID format",
		})
		return
	}

	// Validate seats
	for _, seat := range req.Seats {
		if strings.TrimSpace(seat) == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_seats",
				Message: "Seat codes cannot be empty",
			})
			return
		}
	}

	// Create reservation
	reservation, err := h.reservationService.CreateReservation(
		c.Request.Context(),
		userID,
		scheduleID,
		req.Seats,
		req.TotalPrice,
	)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		if strings.Contains(err.Error(), "seats required") {
			statusCode = http.StatusBadRequest
			errorType = "seats_required"
		} else if strings.Contains(err.Error(), "invalid total price") {
			statusCode = http.StatusBadRequest
			errorType = "invalid_total_price"
		} else if strings.Contains(err.Error(), "seat") && strings.Contains(err.Error(), "taken") {
			statusCode = http.StatusConflict
			errorType = "seats_unavailable"
		} else if strings.Contains(err.Error(), "hold seats") {
			statusCode = http.StatusConflict
			errorType = "seats_unavailable"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	// Extract seat codes
	seatCodes := make([]string, 0, len(reservation.Seats))
	for _, seat := range reservation.Seats {
		seatCodes = append(seatCodes, seat.SeatCode)
	}

	c.JSON(http.StatusCreated, SuccessResponse{
		Message: "Reservation created successfully",
		Data: ReservationResponse{
			ID:         reservation.ID.String(),
			UserID:     reservation.UserID.String(),
			ScheduleID: reservation.ScheduleID.String(),
			Seats:      seatCodes,
			TotalPrice: reservation.TotalPrice,
			Status:     string(reservation.Status),
			ExpiresAt:  reservation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:  reservation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}

// ---------------- CONFIRM RESERVATION ----------------
func (h *ReservationHandler) ConfirmReservation(c *gin.Context) {
	reservationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_reservation_id",
			Message: "Invalid reservation ID format",
		})
		return
	}

	if err := h.reservationService.ConfirmReservation(c.Request.Context(), reservationID); err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errorType = "reservation_not_found"
		} else if strings.Contains(err.Error(), "cannot confirm") {
			statusCode = http.StatusBadRequest
			errorType = "invalid_status_transition"
		} else if strings.Contains(err.Error(), "expired") {
			statusCode = http.StatusBadRequest
			errorType = "reservation_expired"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Reservation confirmed successfully",
	})
}

// ---------------- CANCEL RESERVATION ----------------
func (h *ReservationHandler) CancelReservation(c *gin.Context) {
	reservationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_reservation_id",
			Message: "Invalid reservation ID format",
		})
		return
	}

	if err := h.reservationService.CancelReservation(c.Request.Context(), reservationID); err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errorType = "reservation_not_found"
		} else if strings.Contains(err.Error(), "cannot cancel") {
			statusCode = http.StatusBadRequest
			errorType = "invalid_status_transition"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Reservation canceled successfully",
	})
}

// ---------------- GET RESERVATION ----------------
func (h *ReservationHandler) GetReservation(c *gin.Context) {
	reservationIDStr := c.Param("id")
	if reservationIDStr == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "missing_reservation_id",
			Message: "Reservation ID is required",
		})
		return
	}

	reservationID, err := uuid.Parse(reservationIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_reservation_id",
			Message: "Invalid reservation ID format",
		})
		return
	}

	reservation, err := h.reservationService.GetReservation(c.Request.Context(), reservationID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errorType = "reservation_not_found"
		}

		c.JSON(statusCode, ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	// Extract seat codes
	seatCodes := make([]string, 0, len(reservation.Seats))
	for _, seat := range reservation.Seats {
		seatCodes = append(seatCodes, seat.SeatCode)
	}

	c.JSON(http.StatusOK, SuccessResponse{
		Message: "Reservation retrieved successfully",
		Data: ReservationResponse{
			ID:         reservation.ID.String(),
			UserID:     reservation.UserID.String(),
			ScheduleID: reservation.ScheduleID.String(),
			Seats:      seatCodes,
			TotalPrice: reservation.TotalPrice,
			Status:     string(reservation.Status),
			ExpiresAt:  reservation.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
			CreatedAt:  reservation.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		},
	})
}
