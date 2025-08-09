package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"smat/iot/simulation/iot-inventory-management/internal/service"
	"smat/iot/simulation/iot-inventory-management/pkg/utils"
	"strconv"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

func (h *InventoryHandler) GetLatestReading(c *gin.Context) {
	deviceIDStr := c.Param("deviceId")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid device ID")
		return
	}

	reading, err := h.inventoryService.GetLatestReading(c.Request.Context(), deviceID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch reading")
		return
	}

	if reading == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "No readings found")
		return
	}

	utils.SuccessResponse(c, "Reading fetched successfully", reading)
}

func (h *InventoryHandler) GetDeviceHistory(c *gin.Context) {
	deviceIDStr := c.Param("deviceId")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid device ID")
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	history, err := h.inventoryService.GetDeviceHistory(c.Request.Context(), deviceID, limit)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch history")
		return
	}

	utils.SuccessResponse(c, "History fetched successfully", history)
}
