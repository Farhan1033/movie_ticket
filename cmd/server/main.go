package main

import (
	"movie-ticket/config"
	"movie-ticket/infra/postgres"
	redis_config "movie-ticket/infra/redis"
	"movie-ticket/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	postgres.InitDB()
	redis_config.InitRedis()

	r := gin.Default()

	router.InitRouter(r)

	r.Run(":" + config.Get("PORT"))
}
