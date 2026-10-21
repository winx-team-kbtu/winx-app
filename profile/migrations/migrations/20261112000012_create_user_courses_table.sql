-- +goose Up
-- +goose StatementBegin
CREATE TABLE user_courses
(
    user_id    BIGINT      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    course_id  BIGINT      NOT NULL REFERENCES courses (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, course_id)
);

CREATE INDEX idx_user_courses_course_id ON user_courses (course_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS user_courses;
-- +goose StatementEnd
