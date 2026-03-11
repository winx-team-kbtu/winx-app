-- +goose Up
-- +goose StatementBegin
CREATE TABLE oauth_tokens
(
    id                 BIGSERIAL PRIMARY KEY,
    client_id          TEXT      NOT NULL,
    user_id            TEXT,
    redirect_uri       TEXT,
    scope              TEXT,
    code               TEXT UNIQUE,
    code_expires_at    TIMESTAMP,
    access             TEXT UNIQUE,
    access_expires_at  TIMESTAMP,
    code_expires_in    BIGINT,
    refresh            TEXT UNIQUE,
    refresh_expires_at TIMESTAMP,

    payload            JSONB     NOT NULL DEFAULT '{}'::jsonb,

    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_tokens;
-- +goose StatementEnd
