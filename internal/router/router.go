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
	uiHandler *handler.UIHandler,
) *gin.Engine {
	router := gin.Default()

	// Static files
	router.Static("/static", "./web/static")

	router.Use(middleware.CORS())

	// UI Routes
	ui := router.Group("/ui")
	{
		ui.GET("/login", uiHandler.LoginPage)
		ui.POST("/login", uiHandler.HandleLogin)
		ui.GET("/dashboard", uiHandler.Dashboard)
		ui.GET("/devices/:clientId", uiHandler.GetDevices)
		ui.GET("/device/:deviceId", uiHandler.GetDeviceModal)
		ui.GET("/logout", uiHandler.Logout)
	}

	// Redirect root to login
	router.GET("/", func(c *gin.Context) {
		c.Redirect(301, "/ui/login")
	})

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
			clients.GET("/:clientId/devices", deviceHandler.GetAllDevices)
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
