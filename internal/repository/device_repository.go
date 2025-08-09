package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
)

type deviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(ctx context.Context, device *domain.Device) error {
	query := `
        INSERT INTO devices (device_id, client_id, location, max_capacity, item_weight)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		device.DeviceID, device.ClientID, device.Location,
		device.MaxCapacity, device.ItemWeight,
	).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)

	return err
}

func (r *deviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*domain.Device, error) {
	device := &domain.Device{}
	query := `
        SELECT id, device_id, client_id, location, max_capacity, item_weight, created_at, updated_at
        FROM devices WHERE device_id = $1`

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&device.ID, &device.DeviceID, &device.ClientID, &device.Location,
		&device.MaxCapacity, &device.ItemWeight, &device.CreatedAt, &device.UpdatedAt,
	)

	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	}
	return device, err
}

func (r *deviceRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Device, error) {
	query := `
        SELECT id, device_id, client_id, location, max_capacity, item_weight, created_at, updated_at
        FROM devices WHERE client_id = $1`

	rows, err := r.db.QueryContext(ctx, query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		device := &domain.Device{}
		err := rows.Scan(
			&device.ID, &device.DeviceID, &device.ClientID, &device.Location,
			&device.MaxCapacity, &device.ItemWeight, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *deviceRepository) GetAll(ctx context.Context) ([]*domain.Device, error) {
	query := `
        SELECT id, device_id, client_id, location, max_capacity, item_weight, created_at, updated_at
        FROM devices ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.Device
	for rows.Next() {
		device := &domain.Device{}
		err := rows.Scan(
			&device.ID, &device.DeviceID, &device.ClientID, &device.Location,
			&device.MaxCapacity, &device.ItemWeight, &device.CreatedAt, &device.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		devices = append(devices, device)
	}

	return devices, nil
}

func (r *deviceRepository) Update(ctx context.Context, device *domain.Device) error {
	query := `
        UPDATE devices 
        SET location = $1, max_capacity = $2, item_weight = $3, updated_at = CURRENT_TIMESTAMP
        WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query,
		device.Location, device.MaxCapacity, device.ItemWeight, device.ID,
	)
	return err
}

func (r *deviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
