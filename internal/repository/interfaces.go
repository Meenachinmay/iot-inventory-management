package repository

import (
	"context"
	"github.com/google/uuid"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetByDeviceID(ctx context.Context, deviceID string) (*domain.Device, error)
	GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Device, error)
	GetAll(ctx context.Context) ([]*domain.Device, error)
	Update(ctx context.Context, device *domain.Device) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type InventoryRepository interface {
	CreateReading(ctx context.Context, reading *domain.InventoryReading) error
	GetLatestReading(ctx context.Context, deviceID uuid.UUID) (*domain.InventoryReading, error)
	GetReadingsByDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]*domain.InventoryReading, error)
	CacheLatestWeight(ctx context.Context, deviceID string, weight float64) error
	GetCachedWeight(ctx context.Context, deviceID string) (float64, error)
}
