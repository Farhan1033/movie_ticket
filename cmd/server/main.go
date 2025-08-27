package main

import (
	"movie-ticket/config"
	"movie-ticket/infra/postgres"
	redis_config "movie-ticket/infra/redis"
	"movie-ticket/internal/router"

	_ "movie-ticket/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Movie Ticket API
// @version 1.0
// @description Dokumentasi API untuk aplikasi Movie Ticket
// @host localhost:8080
// @BasePath /api/v1
func main() {
	config.LoadEnv()
	postgres.InitDB()
	redis_config.InitRedis()

	r := gin.Default()

	router.InitRouter(r)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":" + config.Get("PORT"))
}
