-- +goose Up
-- +goose StatementBegin
CREATE TABLE oauth_clients
(
    id            TEXT PRIMARY KEY,
    secret        TEXT      NOT NULL,
    redirect_uris TEXT[] NOT NULL DEFAULT '{}',
    scopes        TEXT[] NOT NULL DEFAULT '{}',
    created_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS oauth_clients;
-- +goose StatementEnd
