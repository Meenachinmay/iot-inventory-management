-- +goose Up
CREATE TABLE IF NOT EXISTS inventory_readings (
                                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                                  device_id UUID REFERENCES devices(id) ON DELETE CASCADE NOT NULL,
                                                  weight DECIMAL(10, 2) NOT NULL,
                                                  item_count INTEGER NOT NULL,
                                                  timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_inventory_readings_device_id ON inventory_readings(device_id);
CREATE INDEX IF NOT EXISTS idx_inventory_readings_timestamp ON inventory_readings(timestamp);

-- +goose Down
DROP TABLE IF EXISTS inventory_readings CASCADE;