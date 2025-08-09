-- +goose Up
CREATE TABLE IF NOT EXISTS devices (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       device_id VARCHAR(100) UNIQUE NOT NULL,
                                       client_id UUID REFERENCES clients(id) ON DELETE CASCADE NOT NULL,
                                       location VARCHAR(255) NOT NULL,
                                       max_capacity DECIMAL(10, 2) DEFAULT 100.00 NOT NULL,
                                       item_weight DECIMAL(10, 2) DEFAULT 1.00 NOT NULL,
                                       created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_devices_client_id ON devices(client_id);

-- +goose Down
DROP TABLE IF EXISTS devices CASCADE;