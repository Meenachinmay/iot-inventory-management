package domain

import (
	"github.com/google/uuid"
	"time"
)

type Client struct {
	ID           uuid.UUID `json:"id"`
	TotalDevices int       `json:"total_devices"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
