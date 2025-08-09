package domain

import (
	"github.com/google/uuid"
	"time"
)

type InventoryReading struct {
	ID        uuid.UUID `json:"id"`
	DeviceID  uuid.UUID `json:"device_id"`
	Weight    float64   `json:"weight"`
	ItemCount int       `json:"item_count"`
	Timestamp time.Time `json:"timestamp"`
}

type InventoryUpdate struct {
	DeviceID    string    `json:"device_id"`
	ClientID    string    `json:"client_id"`
	Weight      float64   `json:"weight"`
	ItemCount   int       `json:"item_count"`
	PrevWeight  float64   `json:"prev_weight"`
	WeightDelta float64   `json:"weight_delta"`
	Timestamp   time.Time `json:"timestamp"`
	Location    string    `json:"location"`
}
