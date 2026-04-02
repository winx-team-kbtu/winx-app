-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL;
UPDATE users SET role_id = 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN IF EXISTS role_id;
-- +goose StatementEnd
