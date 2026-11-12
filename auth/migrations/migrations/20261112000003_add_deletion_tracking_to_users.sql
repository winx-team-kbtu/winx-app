-- +goose Up
-- +goose StatementBegin
-- Bookkeeping for pending account deletions: how often cleanup was attempted, when, and why
-- the last attempt failed. The retry worker uses it to back off between attempts.
ALTER TABLE users
    ADD COLUMN deletion_attempts        INT         NOT NULL DEFAULT 0,
    ADD COLUMN deletion_last_attempt_at TIMESTAMPTZ NULL,
    ADD COLUMN deletion_last_error      TEXT        NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS deletion_last_error,
    DROP COLUMN IF EXISTS deletion_last_attempt_at,
    DROP COLUMN IF EXISTS deletion_attempts;
-- +goose StatementEnd
