-- +goose Up
-- +goose StatementBegin
CREATE TABLE notifications
(
    id                   BIGSERIAL PRIMARY KEY,
    notification_type_id BIGINT      NOT NULL REFERENCES notification_types (id),
    recipient            VARCHAR(255) NOT NULL,
    subject              VARCHAR(255) NOT NULL,
    body                 TEXT         NOT NULL,
    payload              JSONB        NOT NULL DEFAULT '{}'::jsonb,
    status               VARCHAR(50)  NOT NULL DEFAULT 'pending',
    channel              VARCHAR(50)  NOT NULL DEFAULT 'email',
    error_message        TEXT,
    sent_at              TIMESTAMP,
    created_at           TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_notifications_status_id ON notifications (status, id);
CREATE INDEX idx_notifications_type_id ON notifications (notification_type_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS notifications;
-- +goose StatementEnd
