-- +goose Up
-- +goose StatementBegin
CREATE TABLE academic_profiles
(
    user_id        BIGINT PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    skills         JSONB       NOT NULL DEFAULT '[]'::jsonb,
    preferred_role VARCHAR(50) NOT NULL DEFAULT '',
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS academic_profiles;
-- +goose StatementEnd
