-- +goose Up
SELECT setval(pg_get_serial_sequence('roles','id'), GREATEST((SELECT MAX(id) FROM roles), (SELECT last_value FROM roles_id_seq)));

CREATE TABLE admin_audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    entity TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    record_key JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE admin_audit_logs;
