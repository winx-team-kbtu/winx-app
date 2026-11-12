-- +goose Up
-- +goose StatementBegin
-- Set when the user asked to delete the account. While it is set the account cannot sign in
-- and its cleanup is pending; the row itself is removed once every store has been cleaned.
ALTER TABLE users
    ADD COLUMN deletion_requested_at TIMESTAMPTZ NULL;

CREATE INDEX idx_users_deletion_requested_at ON users (deletion_requested_at)
    WHERE deletion_requested_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_deletion_requested_at;
ALTER TABLE users
    DROP COLUMN IF EXISTS deletion_requested_at;
-- +goose StatementEnd
