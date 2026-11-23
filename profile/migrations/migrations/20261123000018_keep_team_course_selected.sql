-- +goose Up
INSERT INTO user_courses (user_id, course_id)
SELECT user_id, course_id FROM team_members
ON CONFLICT (user_id, course_id) DO NOTHING;

UPDATE teams t SET created_by = (
    SELECT m.user_id FROM team_members m
    JOIN users u ON u.id=m.user_id AND u.deletion_requested_at IS NULL
    WHERE m.team_id=t.id ORDER BY m.joined_at,m.user_id LIMIT 1
) WHERE t.created_by IS NULL;

ALTER TABLE team_members ADD CONSTRAINT team_members_enrollment_fkey
    FOREIGN KEY (user_id, course_id) REFERENCES user_courses (user_id, course_id)
    DEFERRABLE INITIALLY DEFERRED;

-- +goose Down
ALTER TABLE team_members DROP CONSTRAINT team_members_enrollment_fkey;
