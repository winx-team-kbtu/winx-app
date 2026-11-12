-- +goose Up
-- +goose StatementBegin
-- Course-scoped Like/Pass decisions. Deliberately separate from the Social
-- swipes/matches tables: a Like is a private shortlist entry, not a match.
CREATE TABLE course_choices
(
    user_id      BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    candidate_id BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    course_id    BIGINT      NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    choice       VARCHAR(4)  NOT NULL CHECK (choice IN ('like', 'pass')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, candidate_id, course_id),
    CHECK (user_id <> candidate_id)
);

CREATE INDEX idx_course_choices_shortlist ON course_choices (user_id, course_id, choice);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS course_choices;
-- +goose StatementEnd
