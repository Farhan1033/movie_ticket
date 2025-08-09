package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email       string    `gorm:"unique" json:"email" binding:"required,email"`
	Password    string    `json:"password" binding:"required,min=6"`
	FullName    string    `gorm:"type:varchar(100)" json:"full_name" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required"`
	Role        string    `gorm:"type:varchar(5)" json:"role"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
