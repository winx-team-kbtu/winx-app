-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_interests
(
    user_id     BIGINT    NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id BIGINT    NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    added_at    TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (user_id, interest_id)
);

CREATE INDEX idx_user_interests_user_id ON user_interests (user_id);
CREATE INDEX idx_user_interests_interest_id ON user_interests (interest_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_interests;
-- +goose StatementEnd
