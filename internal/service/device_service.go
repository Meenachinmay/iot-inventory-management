package service

import (
	"context"
	"github.com/google/uuid"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"smat/iot/simulation/iot-inventory-management/internal/repository"
)

type deviceService struct {
	repo repository.DeviceRepository
}

func NewDeviceService(repo repository.DeviceRepository) DeviceService {
	return &deviceService{repo: repo}
}

func (s *deviceService) RegisterDevice(ctx context.Context, device *domain.Device) error {
	return s.repo.Create(ctx, device)
}

func (s *deviceService) GetDevice(ctx context.Context, deviceID uuid.UUID) (*domain.Device, error) {
	return s.repo.GetByDeviceID(ctx, deviceID)
}

func (s *deviceService) GetDevicesByClient(ctx context.Context, clientID uuid.UUID) ([]*domain.Device, error) {
	return s.repo.GetByClientID(ctx, clientID)
}

func (s *deviceService) GetAllDevices(ctx context.Context) ([]*domain.Device, error) {
	return s.repo.GetAll(ctx)
}

func (s *deviceService) InitializeDevices(ctx context.Context) error {
	clientIDs := []string{
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a12",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a13",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a14",
		"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a15",
	}

	for _, clientIDStr := range clientIDs {
		clientID, _ := uuid.Parse(clientIDStr)
		for j := 0; j < 100; j++ {
			device := &domain.Device{
				ClientID: clientID,
			}

			existing, _ := s.repo.GetByDeviceID(ctx, device.ID)
			if existing == nil {
				if err := s.repo.Create(ctx, device); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *deviceService) Update(ctx context.Context, device *domain.Device) error {
	return s.repo.Update(ctx, device)
}
