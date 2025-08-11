-- +goose Up
-- +goose StatementBegin
ALTER TABLE devices ADD COLUMN current_weight DECIMAL(10, 2) NOT NULL DEFAULT 0.0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE devices DROP COLUMN current_weight;
-- +goose StatementEnd