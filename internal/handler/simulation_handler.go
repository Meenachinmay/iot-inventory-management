package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"smat/iot/simulation/iot-inventory-management/internal/service"
	"smat/iot/simulation/iot-inventory-management/pkg/utils"
)

type SimulationHandler struct {
	mqttService       service.MQTTService
	deviceService     service.DeviceService
	simulationService service.SimulationService
}

func NewSimulationHandler(
	mqttService service.MQTTService,
	deviceService service.DeviceService,
	simulationService service.SimulationService,
) *SimulationHandler {
	return &SimulationHandler{
		mqttService:       mqttService,
		deviceService:     deviceService,
		simulationService: simulationService,
	}
}

func (h *SimulationHandler) SimulateSale(c *gin.Context) {
	deviceID := c.Param("deviceId")

	var req struct {
		ItemsSold int `json:"items_sold" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	// get the message
	message, err := h.simulationService.SimulateSale(c.Request.Context(), deviceID, float64(req.ItemsSold))
	if err != nil {
		if err.Error() == "device not found" {
			utils.ErrorResponse(c, http.StatusNotFound, "Device not found")
			return
		}
		if err.Error() == "invalid device ID format" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Invalid device ID format")
			return
		}
		if err.Error() == "insufficient stock for requested sale" {
			utils.ErrorResponse(c, http.StatusBadRequest, "Insufficient stock for requested sale")
			return
		}
		if len(err.Error()) > 23 && err.Error()[:23] == "failed to update device:" {
			utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to simulate sale")
		return
	}

	// once you the get the message publish that message to mqtt service
	if err := h.mqttService.PublishDeviceMessage(message); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to publish sale event")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   message,
	})
}
