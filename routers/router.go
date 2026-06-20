package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/naki/location-service/controllers"
	"github.com/naki/location-service/functions/api_functions"
	"github.com/naki/location-service/transport/middlewares"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.RedirectTrailingSlash = false
	r.RedirectFixedPath = false

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "location"})
	})

	api := r.Group("/api/v1")
	api.Use(middlewares.AuthMiddleware())
	{
		api.POST("/location/update",
			middlewares.RequireRole("nurse"),
			controllers.UpdateLocation)

		api.GET("/location/nurse/:nurse_id",
			middlewares.RequireRole("nurse", "customer", "super_admin"),
			controllers.GetNurseLocation)

		api.GET("/location/track/:visit_id",
			middlewares.RequireRole("nurse", "customer", "super_admin"),
			controllers.TrackVisit)

		api.GET("/location/track/:visit_id/ws",
			middlewares.RequireRole("nurse", "customer", "super_admin"),
			api_functions.HandleTrackingWebSocket)

		api.GET("/location/active",
			middlewares.RequireRole("super_admin"),
			controllers.GetActiveNurses)

		api.GET("/location/visits",
			middlewares.RequireRole("super_admin"),
			controllers.GetActiveVisits)

		api.GET("/location/history/:nurse_id",
			middlewares.RequireRole("nurse", "super_admin"),
			controllers.GetLocationHistory)

		api.GET("/location/visit/:visit_id/trail",
			middlewares.RequireRole("nurse", "customer", "super_admin"),
			controllers.GetVisitTrail)
	}

	return r
}
