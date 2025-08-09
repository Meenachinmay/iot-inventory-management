package service

import (
	"context"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
)

type DeviceService interface {
	RegisterDevice(ctx context.Context, device *domain.Device) error
	GetDevice(ctx context.Context, deviceID string) (*domain.Device, error)
	GetDevicesByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.Device, error)
	GetAllDevices(ctx context.Context) ([]*domain.Device, error)
	InitializeDevices(ctx context.Context) error
}

type InventoryService interface {
	ProcessWeightUpdate(ctx context.Context, update *domain.DeviceMessage) error
	GetLatestReading(ctx context.Context, deviceID uuid.UUID) (*domain.InventoryReading, error)
	GetDeviceHistory(ctx context.Context, deviceID uuid.UUID, limit int) ([]*domain.InventoryReading, error)
}

type MQTTService interface {
	Connect() error
	Subscribe(topic string) error
	Disconnect()
}

type RabbitMQService interface {
	Connect() error
	PublishMessage(message []byte) error
	PublishMessageWithContext(ctx context.Context, message []byte) error
	PublishJSON(v interface{}) error
	ConsumeMessages() (<-chan []byte, error)
	HealthCheck() error
	GetQueueInfo() (*amqp.Queue, error)
	Close()
}
