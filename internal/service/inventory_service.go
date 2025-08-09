package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"smat/iot/simulation/iot-inventory-management/internal/repository"
)

type inventoryService struct {
	deviceRepo    repository.DeviceRepository
	inventoryRepo repository.InventoryRepository
	rabbitMQ      RabbitMQService
	wsHub         *WebSocketHub
}

func NewInventoryService(
	deviceRepo repository.DeviceRepository,
	inventoryRepo repository.InventoryRepository,
	rabbitMQ RabbitMQService,
	wsHub *WebSocketHub,
) InventoryService {
	return &inventoryService{
		deviceRepo:    deviceRepo,
		inventoryRepo: inventoryRepo,
		rabbitMQ:      rabbitMQ,
		wsHub:         wsHub,
	}
}

func (s *inventoryService) ProcessWeightUpdate(ctx context.Context, update *domain.DeviceMessage) error {
	device, err := s.deviceRepo.GetByDeviceID(ctx, update.DeviceID)
	if err != nil {
		return fmt.Errorf("failed to get device: %w", err)
	}

	if device == nil {
		return fmt.Errorf("device not found: %s", update.DeviceID)
	}

	itemCount := int(math.Round(update.Weight / device.ItemWeight))

	prevWeight, _ := s.inventoryRepo.GetCachedWeight(ctx, update.DeviceID)

	reading := &domain.InventoryReading{
		DeviceID:  device.ID,
		Weight:    update.Weight,
		ItemCount: itemCount,
		Timestamp: update.Timestamp,
	}

	if err := s.inventoryRepo.CreateReading(ctx, reading); err != nil {
		return fmt.Errorf("failed to save reading: %w", err)
	}

	if err := s.inventoryRepo.CacheLatestWeight(ctx, update.DeviceID, update.Weight); err != nil {
		log.Printf("Failed to cache weight: %v", err)
	}

	// Prepare update for WebSocket broadcast
	inventoryUpdate := &domain.InventoryUpdate{
		DeviceID:    update.DeviceID,
		Weight:      update.Weight,
		ItemCount:   itemCount,
		PrevWeight:  prevWeight,
		WeightDelta: prevWeight - update.Weight,
		Timestamp:   update.Timestamp,
		ClientID:    device.ClientID.String(),
		Location:    device.Location,
	}

	if s.wsHub != nil {
		updateJSON, _ := json.Marshal(inventoryUpdate)
		s.wsHub.Broadcast(updateJSON)
	}

	return nil
}

func (s *inventoryService) GetLatestReading(ctx context.Context, deviceID uuid.UUID) (*domain.InventoryReading, error) {
	return s.inventoryRepo.GetLatestReading(ctx, deviceID)
}

func (s *inventoryService) GetDeviceHistory(ctx context.Context, deviceID uuid.UUID, limit int) ([]*domain.InventoryReading, error) {
	return s.inventoryRepo.GetReadingsByDevice(ctx, deviceID, limit)
}
