-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS chats (
    id           BIGSERIAL PRIMARY KEY,
    match_id     BIGINT NOT NULL UNIQUE,
    user_one_id  BIGINT NOT NULL,
    user_two_id  BIGINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chats_user_one_id ON chats (user_one_id);
CREATE INDEX IF NOT EXISTS idx_chats_user_two_id ON chats (user_two_id);

CREATE TABLE IF NOT EXISTS messages (
    id         BIGSERIAL PRIMARY KEY,
    chat_id    BIGINT NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
    sender_id  BIGINT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_messages_chat_id ON messages (chat_id);
CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chats;

-- +goose StatementEnd
