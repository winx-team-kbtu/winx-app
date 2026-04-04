-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS swipes (
    id         BIGSERIAL PRIMARY KEY,
    swiper_id  BIGINT    NOT NULL,
    swiped_id  BIGINT    NOT NULL,
    direction  VARCHAR(5) NOT NULL CHECK (direction IN ('left', 'right')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (swiper_id, swiped_id)
);

CREATE INDEX IF NOT EXISTS idx_swipes_swiper_id ON swipes (swiper_id);
CREATE INDEX IF NOT EXISTS idx_swipes_swiped_id ON swipes (swiped_id);

CREATE TABLE IF NOT EXISTS matches (
    id           BIGSERIAL PRIMARY KEY,
    user_one_id  BIGINT NOT NULL,
    user_two_id  BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_one_id, user_two_id)
);

CREATE INDEX IF NOT EXISTS idx_matches_user_one_id ON matches (user_one_id);
CREATE INDEX IF NOT EXISTS idx_matches_user_two_id ON matches (user_two_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS matches;
DROP TABLE IF EXISTS swipes;

-- +goose StatementEnd
