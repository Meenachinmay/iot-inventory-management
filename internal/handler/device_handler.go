package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"smat/iot/simulation/iot-inventory-management/internal/service"
	"smat/iot/simulation/iot-inventory-management/pkg/utils"
)

var ErrDeviceNotFound = errors.New("device not found")

type DeviceHandler struct {
	deviceService service.DeviceService
}

func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{deviceService: deviceService}
}

func (h *DeviceHandler) GetAllDevices(c *gin.Context) {
	devices, err := h.deviceService.GetAllDevices(c.Request.Context())
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch devices")
		return
	}

	utils.SuccessResponse(c, "Devices fetched successfully", devices)
}

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("deviceId")

	parsedDeviceID, err := uuid.Parse(deviceID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid device ID format")
		return
	}

	device, err := h.deviceService.GetDevice(c.Request.Context(), parsedDeviceID)
	if err != nil {
		if errors.Is(err, ErrDeviceNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Device not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch device")
		return
	}

	utils.SuccessResponse(c, "Device fetched successfully", device)
}

func (h *DeviceHandler) GetDevicesByClient(c *gin.Context) {
	clientIDStr := c.Param("clientId")
	clientID, err := uuid.Parse(clientIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid client ID")
		return
	}

	devices, err := h.deviceService.GetDevicesByClient(c.Request.Context(), clientID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch devices")
		return
	}

	utils.SuccessResponse(c, "Devices fetched successfully", devices)
}

func (h *DeviceHandler) InitializeDevices(c *gin.Context) {
	if err := h.deviceService.InitializeDevices(c.Request.Context()); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to initialize devices")
		return
	}

	utils.SuccessResponse(c, "Devices initialized successfully", nil)
}
