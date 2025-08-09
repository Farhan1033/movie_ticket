package dto

import "github.com/google/uuid"

type RegisterRequest struct {
	Email       string `json:"email" validate:"required"`
	Password    string `json:"password" validate:"required,min=6,max=100"`
	FullName    string `json:"full_name" validate:"required,min=1,max=100"`
	PhoneNumber string `json:"phone_number" validate:"required,min=1,max=13"`
	Role        string `json:"role" validate:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type UserSession struct {
	ID    uuid.UUID `json:"id"`
	Role  string    `json:"role"`
	Email string    `json:"email"`
}
