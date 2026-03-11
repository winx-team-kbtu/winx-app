-- +goose Up
-- +goose StatementBegin
CREATE TABLE notification_types
(
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(100) NOT NULL UNIQUE,
    name       VARCHAR(255) NOT NULL,
    channel    VARCHAR(50)  NOT NULL DEFAULT 'email',
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO notification_types (code, name, channel)
VALUES ('welcome', 'Welcome notification', 'email'),
       ('password_reset', 'Password reset notification', 'email'),
       ('match', 'Match notification', 'email')
ON CONFLICT (code) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notification_types;
-- +goose StatementEnd
