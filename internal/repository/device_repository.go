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
        INSERT INTO devices (client_id, item_weight)
        VALUES ($1, $2)
        RETURNING id, created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		device.ClientID, device.ItemWeight).Scan(&device.ID, &device.CreatedAt, &device.UpdatedAt)

	return err
}

func (r *deviceRepository) GetByDeviceID(ctx context.Context, id uuid.UUID) (*domain.Device, error) {
	device := &domain.Device{}
	query := `
        SELECT id, client_id, current_item_count, max_capacity, total_item_sold_count, item_weight, created_at, updated_at
        FROM devices WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&device.ID, &device.ClientID, &device.CurrentItemCount,
		&device.MaxCapacity, &device.TotalItemSoldCount, &device.ItemWeight, &device.CreatedAt, &device.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return device, err
}

func (r *deviceRepository) GetByClientID(ctx context.Context, clientID uuid.UUID) ([]*domain.Device, error) {
	query := `
        SELECT id, client_id, current_item_count, max_capacity, total_item_sold_count, item_weight, current_weight, created_at, updated_at
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
			&device.ID, &device.ClientID, &device.CurrentItemCount,
			&device.MaxCapacity, &device.TotalItemSoldCount, &device.ItemWeight, &device.CurrentWeight, &device.CreatedAt, &device.UpdatedAt,
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
        SELECT id, client_id, current_item_count, max_capacity, total_item_sold_count, item_weight, current_weight, created_at, updated_at
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
			&device.ID, &device.ClientID, &device.CurrentItemCount,
			&device.MaxCapacity, &device.TotalItemSoldCount, &device.ItemWeight, &device.CurrentWeight, &device.CreatedAt, &device.UpdatedAt,
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
        SET current_item_count = $1, max_capacity = $2, total_item_sold_count = $3, item_weight = $4, current_weight = $5, updated_at = CURRENT_TIMESTAMP
        WHERE id = $6`

	_, err := r.db.ExecContext(ctx, query,
		device.CurrentItemCount, device.MaxCapacity, device.TotalItemSoldCount, device.ItemWeight, device.CurrentWeight, device.ID,
	)
	return err
}

func (r *deviceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM devices WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
