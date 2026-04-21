-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS recommendations (
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT         NOT NULL,
    recommended_id   BIGINT         NOT NULL,
    score            DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ    NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, recommended_id)
);

CREATE INDEX IF NOT EXISTS idx_recommendations_user_id ON recommendations (user_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_score   ON recommendations (user_id, score DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS recommendations;

-- +goose StatementEnd
