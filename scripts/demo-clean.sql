-- Removes demo data created by seed-prod-demo.sql.
-- Touches only accounts with an e-mail ending in @demo.winx.local; real users are never affected.
BEGIN;

CREATE TEMP TABLE demo_ids ON COMMIT DROP AS
SELECT id FROM users WHERE email LIKE '%@demo.winx.local';

-- chat tables have no foreign keys to users, so they are cleaned explicitly
DELETE FROM chats
WHERE user_one_id IN (SELECT id FROM demo_ids)
   OR user_two_id IN (SELECT id FROM demo_ids);

-- profiles, academic data, course choices and team memberships are removed by ON DELETE CASCADE
DELETE FROM users WHERE id IN (SELECT id FROM demo_ids);

-- teams left without members after the accounts were removed
DELETE FROM teams t
WHERE NOT EXISTS (SELECT 1 FROM team_members m WHERE m.team_id = t.id)
  AND t.join_code LIKE 'DEMO%';

COMMIT;
