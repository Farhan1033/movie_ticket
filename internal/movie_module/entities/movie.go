package entities

import (
	"time"

	"github.com/google/uuid"
)

type Movies struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Title            string    `gorm:"type:varchar(200); not null" json:"title" binding:"required"`
	Description      string    `gorm:"type:text" json:"description" binding:"required"`
	Genre            string    `gorm:"type:varchar(100)" json:"genre" binding:"required"`
	Duration_Minutes int       `gorm:"type:int; not null" json:"duration_minutes" binding:"required"`
	Rating           string    `gorm:"type:varchar(10)" json:"rating" binding:"required"`
	Poster_Url       string    `gorm:"type:varchar(500)" json:"poster_url" binding:"required"`
	Created_At       time.Time `gorm:"autoCreateTime" json:"created_at"`
	Updated_At       time.Time `gorm:"autoCreateTime; autoUpdateTime" json:"updated_at"`
}
