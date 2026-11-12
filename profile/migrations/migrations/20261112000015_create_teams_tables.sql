-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams
(
    id         BIGSERIAL PRIMARY KEY,
    course_id  BIGINT       NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    name       VARCHAR(100) NOT NULL,
    join_code  VARCHAR(8)   NOT NULL,
    created_by BIGINT       REFERENCES users (id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT teams_join_code_key UNIQUE (join_code),
    -- target of the composite FK below, keeps a member's course equal to the team's course
    CONSTRAINT teams_id_course_id_key UNIQUE (id, course_id)
);

CREATE TABLE team_members
(
    team_id   BIGINT      NOT NULL,
    user_id   BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    course_id BIGINT      NOT NULL,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (team_id, user_id),
    -- one team per user per course
    CONSTRAINT team_members_user_course_key UNIQUE (user_id, course_id),
    CONSTRAINT team_members_team_course_fkey FOREIGN KEY (team_id, course_id)
        REFERENCES teams (id, course_id) ON DELETE CASCADE
);

CREATE INDEX idx_team_members_course_id ON team_members (course_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_members;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
