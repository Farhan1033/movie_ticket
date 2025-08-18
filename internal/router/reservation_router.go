package router

import (
	"movie-ticket/infra/postgres"
	redis_config "movie-ticket/infra/redis"
	"movie-ticket/internal/middleware"
	"movie-ticket/internal/reservation_module/handler"
	repository "movie-ticket/internal/reservation_module/repositories"
	service "movie-ticket/internal/reservation_module/services"

	"github.com/gin-gonic/gin"
)

func InitReservationRouter(c *gin.Engine) {
	repoDB := repository.NewReservationRepository(postgres.DB)
	repoRedis := repository.NewSeatRedisRepository(redis_config.RedisClient)
	svc := service.NewReservationService(repoDB, repoRedis)

	api := c.Group("/api/v1")
	api.Use(middleware.JwtMiddleware(), middleware.RequireRole("user", "admin"))
	{
		handler.NewReservationHandler(api, svc)
	}
}
