package router

import "github.com/gin-gonic/gin"

func InitRouter(r *gin.Engine) {
	InitAuthRoutes(r)
	InitMovieRoute(r)
	InitStudioRouter(r)
	InitialScheduleRouter(r)
}
