package domain

import (
	"github.com/google/uuid"
	"time"
)

type Device struct {
	ID                 uuid.UUID `json:"id"`
	ClientID           uuid.UUID `json:"client_id"`
	CurrentItemCount   int       `json:"current_item_count"`
	ItemWeight         float64   `json:"item_weight"`
	MaxCapacity        float64   `json:"max_capacity"`
	TotalItemSoldCount int       `json:"total_item_sold_count"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type DeviceMessage struct {
	DeviceID         string    `json:"device_id"`
	CurrentTotalItem float64   `json:"current_total_item"`
	ItemWeight       float64   `json:"item_weight"`
	Timestamp        time.Time `json:"timestamp"`
}
