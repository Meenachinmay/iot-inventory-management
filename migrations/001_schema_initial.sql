-- +goose Up
-- +goose StatementBegin
-- Initialize database schema with all tables including item_weight column

CREATE TABLE IF NOT EXISTS clients (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    total_device INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    client_id UUID REFERENCES clients(id) ON DELETE CASCADE NOT NULL,
    current_item_count INT DEFAULT 0,
    max_capacity DECIMAL(10, 2) DEFAULT 100.00 NOT NULL,
    total_item_sold_count INT DEFAULT 0,
    item_weight DECIMAL DEFAULT 0.0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_client_id ON devices(client_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_devices_client_id;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS clients;
-- +goose StatementEnd