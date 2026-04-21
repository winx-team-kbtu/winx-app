-- +goose Up
-- +goose StatementBegin
CREATE TABLE profile_photos
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT       NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    email      VARCHAR(255) NOT NULL,
    provider   VARCHAR(50)  NOT NULL DEFAULT 's3',
    bucket     VARCHAR(255) NOT NULL,
    object_key TEXT         NOT NULL,
    url        TEXT         NOT NULL,
    mime_type  VARCHAR(255),
    size_bytes BIGINT       NOT NULL DEFAULT 0,
    width      INT,
    height     INT,
    created_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_profile_photos_object_key ON profile_photos (object_key);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profile_photos;
-- +goose StatementEnd
