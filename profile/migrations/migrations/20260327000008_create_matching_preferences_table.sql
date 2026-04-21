-- +goose Up
-- +goose StatementBegin
CREATE TABLE matching_preferences
(
    user_id            BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    min_age            INT         NOT NULL DEFAULT 18,
    max_age            INT         NOT NULL DEFAULT 99,
    max_distance_km    INT         NOT NULL DEFAULT 50,
    interested_in      VARCHAR(50)[]         DEFAULT ARRAY[]::VARCHAR(50)[],
    show_me_global     BOOLEAN     NOT NULL DEFAULT FALSE,
    only_show_verified BOOLEAN     NOT NULL DEFAULT FALSE,
    updated_at         TIMESTAMP   NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS matching_preferences;
-- +goose StatementEnd
