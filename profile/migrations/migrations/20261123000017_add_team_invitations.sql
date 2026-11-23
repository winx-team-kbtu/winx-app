-- +goose Up
CREATE TABLE team_invitations (
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT NOT NULL UNIQUE REFERENCES chats(id) ON DELETE CASCADE,
    team_id BIGINT NOT NULL,
    inviter_id BIGINT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    FOREIGN KEY (team_id, inviter_id) REFERENCES team_members(team_id, user_id) ON DELETE CASCADE
);

CREATE INDEX team_invitations_inviter_membership_idx ON team_invitations(team_id, inviter_id);

-- +goose Down
DROP TABLE team_invitations;
