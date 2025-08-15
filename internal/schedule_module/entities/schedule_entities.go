package entities

import (
	"movie-ticket/internal/movie_module/entities"
	studio "movie-ticket/internal/studio_module/entities"
	"time"

	"github.com/google/uuid"
)

type Schedules struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	MovieID   uuid.UUID `gorm:"type:uuid;not null" json:"movie_id"`
	StudioID  uuid.UUID `gorm:"type:uuid;not null" json:"studio_id"`
	StartTime string    `gorm:"type:time;not null" json:"start_time"`
	EndTime   string    `gorm:"type:time;not null" json:"end_time"`
	Price     int       `gorm:"not null" json:"price"`
	CreatedAt time.Time `gorm:"autoCreateTime;" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoCreateTime;autoUpdateTime" json:"updated_at"`

	//Relation
	Movie  entities.Movies `gorm:"foreignKey:MovieID;references:ID" json:"movies"`
	Studio studio.Studio   `gorm:"foreignKey:StudioID;references:ID" json:"studio"`
}
