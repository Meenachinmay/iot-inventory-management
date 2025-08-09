package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"smat/iot/simulation/iot-inventory-management/internal/service"
)

type HealthHandler struct {
	rabbitMQ service.RabbitMQService
}

func NewHealthHandler(rabbitMQ service.RabbitMQService) *HealthHandler {
	return &HealthHandler{
		rabbitMQ: rabbitMQ,
	}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	health := gin.H{
		"status":   "healthy",
		"services": gin.H{},
	}

	statusCode := http.StatusOK

	if err := h.rabbitMQ.HealthCheck(); err != nil {
		health["services"].(gin.H)["rabbitmq"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		health["status"] = "unhealthy"
		statusCode = http.StatusServiceUnavailable
	} else {
		health["services"].(gin.H)["rabbitmq"] = gin.H{
			"status": "healthy",
		}

		if queueInfo, err := h.rabbitMQ.GetQueueInfo(); err == nil {
			health["services"].(gin.H)["rabbitmq"].(gin.H)["queue_messages"] = queueInfo.Messages
			health["services"].(gin.H)["rabbitmq"].(gin.H)["queue_consumers"] = queueInfo.Consumers
		}
	}

	c.JSON(statusCode, health)
}

func (h *HealthHandler) QueueStats(c *gin.Context) {
	queueInfo, err := h.rabbitMQ.GetQueueInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to get queue statistics",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"queue_name": queueInfo.Name,
		"messages":   queueInfo.Messages,
		"consumers":  queueInfo.Consumers,
	})
}
