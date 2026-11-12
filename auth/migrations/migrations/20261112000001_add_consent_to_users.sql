-- +goose Up
-- +goose StatementBegin
-- Informed consent given at pilot registration. NULL for accounts created
-- before the pilot or by an administrator.
ALTER TABLE users
    ADD COLUMN consent_notice_version VARCHAR(32) NULL,
    ADD COLUMN consent_accepted_at    TIMESTAMPTZ NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS consent_accepted_at,
    DROP COLUMN IF EXISTS consent_notice_version;
-- +goose StatementEnd
