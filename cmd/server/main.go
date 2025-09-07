package main

import (
	"log"
	"movie-ticket/config"
	"movie-ticket/infra/postgres"
	redis_config "movie-ticket/infra/redis"
	"movie-ticket/internal/router"
	"net/http"
	"time"

	_ "movie-ticket/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Movie Ticket API
// @version 1.0
// @description Dokumentasi API untuk aplikasi Movie Ticket
// @host movieticket-farhan10335643-qxvhtr05.leapcell.dev
// @BasePath /api/v1
func main() {
	config.LoadEnv()

	gin.SetMode(gin.ReleaseMode)
	log.Println("Gin mode: RELEASE")

	postgres.InitDB()
	redis_config.InitRedis()

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.InitRouter(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.GET("/kaithheathcheck", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Movie Ticket API is running smoothly 🚀",
		})
	})

	address := "0.0.0.0:" + config.Get("PORT")
	log.Printf("Server berjalan di %s", address)

	if err := r.Run(address); err != nil {
		log.Fatalf("Gagal menjalankan server: %v", err)
	}
}
