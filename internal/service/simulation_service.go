package service

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"log"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"smat/iot/simulation/iot-inventory-management/internal/repository"
	"time"
)

type SimulationService interface {
	SimulateSale(ctx context.Context, deviceID string, itemsSold float64) (*domain.DeviceMessage, error)
}

type simulationService struct {
	deviceRepo repository.DeviceRepository
}

func NewSimulationService(deviceRepo repository.DeviceRepository) SimulationService {
	return &simulationService{
		deviceRepo: deviceRepo,
	}
}

func (s *simulationService) SimulateSale(ctx context.Context, deviceID string, itemsSold float64) (*domain.DeviceMessage, error) {
	deviceUUID, err := uuid.Parse(deviceID)
	if err != nil {
		log.Printf("SimulateSale: ERROR - Failed to parse device UUID: %v", err)
		return nil, errors.New("invalid device ID format")
	}

	device, err := s.deviceRepo.GetByDeviceID(ctx, deviceUUID)
	if err != nil {
		log.Printf("SimulateSale: ERROR - Failed to retrieve device from repository: %v", err)
		return nil, err
	}

	if device == nil {
		log.Printf("SimulateSale: ERROR - Device with ID %s not found", deviceUUID.String())
		return nil, errors.New("device not found")
	}

	if float64(device.CurrentItemCount) < itemsSold {
		log.Printf("SimulateSale: ERROR - Insufficient stock. Current items: %d, Requested: %.2f",
			device.CurrentItemCount, itemsSold)
		return nil, errors.New("insufficient stock for requested sale")
	}

	newItemCount := device.CurrentItemCount - int(itemsSold)
	newWeight := float64(newItemCount) * device.ItemWeight

	if newWeight > device.MaxCapacity {
		log.Printf("SimulateSale: WARNING - New weight %.2f exceeds max capacity %.2f",
			newWeight, device.MaxCapacity)
		return nil, errors.New("weight exceeds maximum capacity")
	}

	device.CurrentItemCount = newItemCount
	device.CurrentWeight = newWeight
	device.TotalItemSoldCount = device.TotalItemSoldCount + int(itemsSold)
	device.UpdatedAt = time.Now()

	log.Printf("SimulateSale: Updating device %s - Items: %d -> %d, Weight: %.2f -> %.2f",
		deviceUUID.String(),
		device.CurrentItemCount+int(itemsSold),
		device.CurrentItemCount,
		device.CurrentWeight+(itemsSold*device.ItemWeight),
		device.CurrentWeight)

	err = s.deviceRepo.Update(ctx, device)
	if err != nil {
		log.Printf("SimulateSale: ERROR - Failed to update device in repository: %v", err)
		return nil, errors.New("failed to update device: " + err.Error())
	}

	message := &domain.DeviceMessage{
		DeviceID:         deviceUUID.String(),
		CurrentTotalItem: float64(device.CurrentItemCount),
		CurrentWeight:    device.CurrentWeight, // Include current weight in response
		ItemWeight:       device.ItemWeight,
		Timestamp:        time.Now(),
	}

	return message, nil
}
