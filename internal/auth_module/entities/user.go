package entities

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email       string    `gorm:"unique" json:"email" binding:"required,email"`
	Password    string    `json:"password" binding:"required,min=6"`
	FullName    string    `gorm:"varchar(100)" json:"full_name" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required"`
	Role        string    `gorm:"varchar(5)" binding:"required" json:"role"`
	Created_At  time.Time `json:"created_at" gorm:"autoCreateTime"`
	Updated_At  time.Time `json:"updated_at" gorm:"autoCreateTime; autoUpdateTime"`
}
