package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

type RegisterResponse struct {
	ID          uuid.UUID `json:"id"`
	Email       string    `json:"email"`
	FullName    string    `json:"full_name"`
	PhoneNumber string    `json:"phone_number"`
	Role        string    `json:"role"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
