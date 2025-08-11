package dto

import (
	"time"

	"github.com/google/uuid"
)

type StudioResponse struct {
	ID            uuid.UUID `json:"id"`
	Name          string    `json:"name"`
	Seat_Capacity int       `json:"seat_capacity"`
	Location      string    `json:"location"`
	Created_At    time.Time `json:"created_at"`
	Updated_At    time.Time `json:"updated_at"`
}

type ResponseMessage struct {
	Message string `json:"message"`
}
