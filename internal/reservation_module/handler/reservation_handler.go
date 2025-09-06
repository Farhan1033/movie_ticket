package handler

import (
	"fmt"
	"net/http"
	"strings"

	"movie-ticket/internal/middleware"
	"movie-ticket/internal/reservation_module/dto"
	service "movie-ticket/internal/reservation_module/services"

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
	r.GET("/reservation/history", h.GetHistory)
}

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

// CreateReservation godoc
// @Summary Membuat reservasi tiket baru
// @Description Membuat reservasi tiket untuk jadwal dan kursi tertentu. Reservasi akan memiliki waktu expired untuk konfirmasi
// @Tags Reservations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param request body dto.CreateReservationRequest true "Reservation creation data"
// @Success 201 {object} SuccessResponse{data=ReservationResponse} "Reservation created successfully"
// @Failure 400 {object} ErrorResponse "Bad Request - Validation error, invalid user ID, schedule ID, atau seats"
// @Failure 409 {object} ErrorResponse "Conflict - Seats unavailable atau sudah diambil"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /reservation/create [post]
// @Security BearerAuth
func (h *ReservationHandler) CreateReservation(c *gin.Context) {
	var req *dto.CreateReservationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	userID, err := middleware.GetUserIDFromRedis(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID format",
		})
		return
	}

	scheduleID, err := uuid.Parse(req.ScheduleID)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "invalid_schedule_id",
			Message: "Invalid schedule ID format",
		})
		return
	}

	for _, seat := range req.Seats {
		if strings.TrimSpace(seat) == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "invalid_seats",
				Message: "Seat codes cannot be empty",
			})
			return
		}
	}

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

// ConfirmReservation godoc
// @Summary Konfirmasi reservasi tiket
// @Description Mengkonfirmasi reservasi yang sebelumnya dibuat. Reservasi harus dalam status pending dan belum expired
// @Tags Reservations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Reservation ID" format(uuid)
// @Success 200 {object} SuccessResponse "Reservation confirmed successfully"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid reservation ID, invalid status transition, atau reservation expired"
// @Failure 404 {object} ErrorResponse "Not Found - Reservation tidak ditemukan"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /reservation/{id}/confirm [put]
// @Security BearerAuth
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

// CancelReservation godoc
// @Summary Membatalkan reservasi tiket
// @Description Membatalkan reservasi yang sebelumnya dibuat. Kursi yang dibatalkan akan kembali tersedia
// @Tags Reservations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Reservation ID" format(uuid)
// @Success 200 {object} SuccessResponse "Reservation canceled successfully"
// @Failure 400 {object} ErrorResponse "Bad Request - Invalid reservation ID atau invalid status transition"
// @Failure 404 {object} ErrorResponse "Not Found - Reservation tidak ditemukan"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /reservation/{id}/cancel [put]
// @Security BearerAuth
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

// GetReservation godoc
// @Summary Mendapatkan detail reservasi berdasarkan ID
// @Description Mengambil informasi lengkap reservasi termasuk detail kursi, status, dan waktu expired
// @Tags Reservations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Param id path string true "Reservation ID" format(uuid)
// @Success 200 {object} SuccessResponse{data=ReservationResponse} "Reservation retrieved successfully"
// @Failure 400 {object} ErrorResponse "Bad Request - Missing atau invalid reservation ID"
// @Failure 404 {object} ErrorResponse "Not Found - Reservation tidak ditemukan"
// @Failure 500 {object} ErrorResponse "Internal Server Error"
// @Router /reservation/{id} [get]
// @Security BearerAuth
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

// GetHistory godoc
// @Summary Mendapatkan riwayat reservasi user
// @Description Mengambil semua riwayat reservasi yang pernah dibuat oleh user yang sedang login
// @Tags Reservations
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token" default(Bearer <token>)
// @Success 200 {object} map[string]interface{} "History retrieved successfully"
// @Failure 400 {object} map[string]interface{} "Bad Request - Invalid user ID"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Router /reservation/history [get]
// @Security BearerAuth
func (h *ReservationHandler) GetHistory(c *gin.Context) {
	userID, err := middleware.GetUserIDFromRedis(c)
	fmt.Print(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	history, err := h.reservationService.GetHistory(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": history})
}
