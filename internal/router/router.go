package router

import (
	"github.com/gin-gonic/gin"
	"smat/iot/simulation/iot-inventory-management/internal/handler"
	"smat/iot/simulation/iot-inventory-management/internal/middleware"
)

func SetupRouter(
	deviceHandler *handler.DeviceHandler,
	inventoryHandler *handler.InventoryHandler,
	wsHandler *handler.WebSocketHandler,
	healthHandler *handler.HealthHandler, // Add this parameter
) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())

	// API routes
	api := router.Group("/server/v1")
	{
		// Device routes
		devices := api.Group("/devices")
		{
			devices.GET("", deviceHandler.GetAllDevices)
			devices.GET("/:deviceId", deviceHandler.GetDevice)
			devices.POST("/initialize", deviceHandler.InitializeDevices)
		}

		// Client routes
		clients := api.Group("/clients")
		{
			clients.GET("/:clientId/devices", deviceHandler.GetDevicesByClient)
		}

		// Inventory routes
		inventory := api.Group("/inventory")
		{
			inventory.GET("/device/:deviceId/latest", inventoryHandler.GetLatestReading)
			inventory.GET("/device/:deviceId/history", inventoryHandler.GetDeviceHistory)
		}

		// Queue monitoring
		api.GET("/queue/stats", healthHandler.QueueStats)
	}

	// WebSocket endpoint
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Health check endpoints
	router.GET("/health", healthHandler.HealthCheck)

	return router
}
