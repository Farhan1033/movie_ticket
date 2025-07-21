package main

import (
	"movie-ticket/config"
	"movie-ticket/infra/postgres"
	"movie-ticket/internal/router"

	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadEnv()
	postgres.InitDB()

	r := gin.Default()

	router.InitRouter(r)

	r.Run(":" + config.Get("PORT"))
}
