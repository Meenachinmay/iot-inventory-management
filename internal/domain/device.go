package domain

import (
	"github.com/google/uuid"
	"time"
)

type Device struct {
	ID          uuid.UUID `json:"id"`
	DeviceID    string    `json:"device_id"`
	ClientID    uuid.UUID `json:"client_id"`
	Location    string    `json:"location"`
	MaxCapacity float64   `json:"max_capacity"`
	ItemWeight  float64   `json:"item_weight"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DeviceMessage struct {
	DeviceID  string    `json:"device_id"`
	Weight    float64   `json:"weight"`
	Timestamp time.Time `json:"timestamp"`
}
