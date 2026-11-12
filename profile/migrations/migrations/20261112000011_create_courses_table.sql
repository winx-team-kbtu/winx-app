-- +goose Up
-- +goose StatementBegin
CREATE TABLE courses
(
    id         BIGSERIAL PRIMARY KEY,
    code       VARCHAR(32)  NOT NULL UNIQUE,
    title      VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Seeded pilot catalogue. Course selection is self-reported, not official enrolment.
INSERT INTO courses (code, title)
VALUES ('ALG201', 'Algorithms and Data Structures'),
       ('DBS202', 'Database Systems'),
       ('WEB203', 'Web Development'),
       ('SWE204', 'Software Engineering'),
       ('NET205', 'Computer Networks'),
       ('OSY206', 'Operating Systems'),
       ('MLR207', 'Machine Learning'),
       ('MOB208', 'Mobile Development');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS courses;
-- +goose StatementEnd
