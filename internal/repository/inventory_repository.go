package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"strconv"
	"time"
)

type inventoryRepository struct {
	db    *sql.DB
	redis *redis.Client
}

func NewInventoryRepository(db *sql.DB, redis *redis.Client) InventoryRepository {
	return &inventoryRepository{db: db, redis: redis}
}

func (r *inventoryRepository) CreateReading(ctx context.Context, reading *domain.InventoryReading) error {
	query := `
        INSERT INTO inventory_readings (device_id, weight, item_count, timestamp)
        VALUES ($1, $2, $3, $4)
        RETURNING id`

	err := r.db.QueryRowContext(ctx, query,
		reading.DeviceID, reading.Weight, reading.ItemCount, reading.Timestamp,
	).Scan(&reading.ID)

	return err
}

func (r *inventoryRepository) GetLatestReading(ctx context.Context, deviceID uuid.UUID) (*domain.InventoryReading, error) {
	reading := &domain.InventoryReading{}
	query := `
        SELECT id, device_id, weight, item_count, timestamp
        FROM inventory_readings
        WHERE device_id = $1
        ORDER BY timestamp DESC
        LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, deviceID).Scan(
		&reading.ID, &reading.DeviceID, &reading.Weight,
		&reading.ItemCount, &reading.Timestamp,
	)

	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	}
	return reading, err
}

func (r *inventoryRepository) GetReadingsByDevice(ctx context.Context, deviceID uuid.UUID, limit int) ([]*domain.InventoryReading, error) {
	query := `
        SELECT id, device_id, weight, item_count, timestamp
        FROM inventory_readings
        WHERE device_id = $1
        ORDER BY timestamp DESC
        LIMIT $2`

	rows, err := r.db.QueryContext(ctx, query, deviceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []*domain.InventoryReading
	for rows.Next() {
		reading := &domain.InventoryReading{}
		err := rows.Scan(
			&reading.ID, &reading.DeviceID, &reading.Weight,
			&reading.ItemCount, &reading.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	return readings, nil
}

func (r *inventoryRepository) CacheLatestWeight(ctx context.Context, deviceID string, weight float64) error {
	key := fmt.Sprintf("device:%s:weight", deviceID)
	return r.redis.Set(ctx, key, weight, 24*time.Hour).Err()
}

func (r *inventoryRepository) GetCachedWeight(ctx context.Context, deviceID string) (float64, error) {
	key := fmt.Sprintf("device:%s:weight", deviceID)
	val, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}
