package router

import (
	"github.com/gin-gonic/gin"
	"smat/iot/simulation/iot-inventory-management/internal/handler"
	"smat/iot/simulation/iot-inventory-management/internal/middleware"
)

func SetupRouter(
	deviceHandler *handler.DeviceHandler,
	wsHandler *handler.WebSocketHandler,
	healthHandler *handler.HealthHandler,
	simulationHandler *handler.SimulationHandler,
) *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORS())

	api := router.Group("/server/v1")
	{
		devices := api.Group("/devices")
		{
			devices.GET("", deviceHandler.GetAllDevices)
			devices.GET("/:deviceId", deviceHandler.GetDevice)
			devices.POST("/initialize", deviceHandler.InitializeDevices)
		}

		clients := api.Group("/clients")
		{
			clients.GET("/:clientId/devices", deviceHandler.GetDevicesByClient)
		}

		simulation := api.Group("/simulation")
		{
			simulation.POST("/device/:deviceId/sale", simulationHandler.SimulateSale)
		}

		api.GET("/queue/stats", healthHandler.QueueStats)
	}

	router.GET("/ws", wsHandler.HandleWebSocket)
	router.GET("/health", healthHandler.HealthCheck)

	return router
}
