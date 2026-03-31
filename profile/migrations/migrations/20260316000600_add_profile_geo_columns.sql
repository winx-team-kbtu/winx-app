-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE profiles
    ADD COLUMN current_location GEOGRAPHY(POINT, 4326),
    ADD COLUMN city VARCHAR(100),
    ADD COLUMN country VARCHAR(100);

CREATE INDEX idx_profiles_current_location ON profiles USING GIST (current_location);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_profiles_current_location;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS current_location,
    DROP COLUMN IF EXISTS city,
    DROP COLUMN IF EXISTS country;
-- +goose StatementEnd
