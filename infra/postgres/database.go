package postgres

import (
	"fmt"
	"log"
	"movie-ticket/config"
	"movie-ticket/internal/auth_module/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		config.Get("DB_HOST"),
		config.Get("DB_USER"),
		config.Get("DB_PASS"),
		config.Get("DB_NAME"),
		config.Get("DB_PORT"),
		config.Get("TIMEZONE"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect database :", err)
	}

	err = DB.AutoMigrate(&entities.User{})
	if err != nil {
		log.Fatal("failed to migrate :", err)
	}
}
