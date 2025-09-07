package postgres

import (
	"log"
	"movie-ticket/config"

	// user "movie-ticket/internal/auth_module/entities"
	// movie "movie-ticket/internal/movie_module/entities"
	// reservation "movie-ticket/internal/reservation_module/entities"
	// schedule "movie-ticket/internal/schedule_module/entities"
	// studio "movie-ticket/internal/studio_module/entities"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require",
	// 	config.Get("DB_HOST"),
	// 	config.Get("DB_USER"),
	// 	config.Get("DB_PASS"),
	// 	config.Get("DB_NAME"),
	// 	config.Get("DB_PORT"),
	// )

	dsn := config.Get("DATABASE_URL")

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("failed to connect database :", err)
	}

	// err = DB.AutoMigrate(
	// 	&user.User{},
	// 	&movie.Movies{},
	// 	&studio.Studio{},
	// 	&schedule.Schedules{},
	// 	&reservation.Reservation{},
	// 	&reservation.ReservationSeat{})
	if err != nil {
		log.Fatal("failed to migrate :", err)
	}

	log.Println("✅ Postgres terkoneksi")
}
