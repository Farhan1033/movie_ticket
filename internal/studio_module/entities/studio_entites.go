package entities

import (
	"time"

	"github.com/google/uuid"
)

type Studio struct {
	ID            uuid.UUID `gorm:"type:uuid; primaryKey" json:"id"`
	Name          string    `gorm:"type:varhcar(100); not null" json:"name" binding:"required"`
	Seat_Capacity int       `gorm:"type:int" json:"seat_capacity" binding:"required"`
	Location      string    `gorm:"type:varchar(100); not null" json:"location" binding:"required"`
	Created_At    time.Time `json:"created_at" gorm:"autoCreateTime"`
	Updated_At    time.Time `json:"updated_at" gorm:"autoCreateTime; autoUpdateTime"`
}
