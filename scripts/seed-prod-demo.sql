-- Demo data for Team Mode: 40 students, courses, skills, teams in every status,
-- saved candidates and academic chats with messages.
--
-- All accounts use the e-mail pattern demoNN@demo.winx.local and share one password,
-- passed as the psql variable demo_password. Run through scripts/seed-prod-demo.sh,
-- which first runs demo-clean.sql, so the seed can be repeated safely.
\set ON_ERROR_STOP on

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- 1. Accounts -------------------------------------------------------------------------
CREATE TEMP TABLE demo_people (seq int PRIMARY KEY, full_name text NOT NULL) ON COMMIT DROP;
INSERT INTO demo_people (seq, full_name) VALUES
 (1,'Aigerim Nurlanova'),(2,'Daniyar Seitkali'),(3,'Madina Kassymova'),(4,'Yerlan Ospanov'),
 (5,'Aruzhan Bekova'),(6,'Timur Khasanov'),(7,'Dana Zhumabayeva'),(8,'Alikhan Nurpeisov'),
 (9,'Zarina Abenova'),(10,'Bekzat Sadykov'),(11,'Amina Tleubayeva'),(12,'Nurlan Akhmetov'),
 (13,'Kamila Iskakova'),(14,'Rustem Dzhaksybekov'),(15,'Ayana Mukhametova'),(16,'Samat Baizhanov'),
 (17,'Lyazzat Omarova'),(18,'Arsen Kenzhebaev'),(19,'Dilnaz Rakhimova'),(20,'Miras Beisenov'),
 (21,'Saltanat Zhaksylykova'),(22,'Eldar Smagulov'),(23,'Assel Nurmagambetova'),(24,'Dias Akhanov'),
 (25,'Meruert Sagyndykova'),(26,'Kanat Bekmuratov'),(27,'Zhanna Abdrakhmanova'),(28,'Askar Sultanov'),
 (29,'Alua Kuanyshbekova'),(30,'Yernur Tokhtarov'),(31,'Balnur Serikbayeva'),(32,'Ilyas Nurbekov'),
 (33,'Gulnaz Dauletova'),(34,'Adilet Zhunusov'),(35,'Karina Petrova'),(36,'Maxim Volkov'),
 (37,'Sofia Ivanova'),(38,'Alexandr Kim'),(39,'Victoria Lee'),(40,'Dmitri Sokolov');

INSERT INTO users (email, password, role_id, consent_notice_version, consent_accepted_at)
SELECT format('demo%s@demo.winx.local', lpad(seq::text, 2, '0')),
       crypt(:'demo_password', gen_salt('bf', 10)),
       1, 'pilot-2026-11-v2', NOW()
FROM demo_people
ON CONFLICT (email) DO NOTHING;

-- seq -> user id for every following step
CREATE TEMP TABLE demo_users ON COMMIT DROP AS
SELECT p.seq, u.id, p.full_name
FROM demo_people p
JOIN users u ON u.email = format('demo%s@demo.winx.local', lpad(p.seq::text, 2, '0'));

INSERT INTO profiles (user_id, email, name, city, country)
SELECT u.id, format('demo%s@demo.winx.local', lpad(u.seq::text, 2, '0')), u.full_name, 'Almaty', 'Kazakhstan'
FROM demo_users u
ON CONFLICT (user_id) DO NOTHING;

-- 2. Academic profiles ----------------------------------------------------------------
INSERT INTO academic_profiles (user_id, skills, preferred_role)
SELECT id,
  CASE seq % 4
    WHEN 0 THEN '["Figma","UX design","Presentations"]'::jsonb
    WHEN 1 THEN '["Go","PostgreSQL","API design"]'::jsonb
    WHEN 2 THEN '["Angular","TypeScript","CSS"]'::jsonb
    ELSE '["Python","Data analysis","Testing"]'::jsonb
  END,
  CASE seq % 4
    WHEN 0 THEN 'Designer'
    WHEN 1 THEN 'Backend developer'
    WHEN 2 THEN 'Frontend developer'
    ELSE 'QA / Data analyst'
  END
FROM demo_users
ON CONFLICT (user_id) DO NOTHING;

-- 3. Courses: everyone takes Algorithms, plus one more; every fifth student also takes ML
INSERT INTO user_courses (user_id, course_id)
SELECT u.id, c.id FROM demo_users u JOIN courses c ON c.code = 'ALG201'
ON CONFLICT DO NOTHING;

INSERT INTO user_courses (user_id, course_id)
SELECT u.id, c.id FROM demo_users u
JOIN courses c ON c.code = CASE u.seq % 3 WHEN 0 THEN 'DBS202' WHEN 1 THEN 'WEB203' ELSE 'SWE204' END
ON CONFLICT DO NOTHING;

INSERT INTO user_courses (user_id, course_id)
SELECT u.id, c.id FROM demo_users u JOIN courses c ON c.code = 'MLR207'
WHERE u.seq % 5 = 0
ON CONFLICT DO NOTHING;

-- 4. Teams in every status ------------------------------------------------------------
-- members are demo seq ranges; a student belongs to one team per course
CREATE TEMP TABLE demo_teams (code text, course text, team_name text, first_seq int, step int, cnt int) ON COMMIT DROP;
INSERT INTO demo_teams VALUES
 ('DEMO01','ALG201','Algorithm Explorers', 1, 1, 2),   -- Forming (2)
 ('DEMO02','ALG201','Code Collective',     3, 1, 3),   -- Ready   (3)
 ('DEMO03','ALG201','Big O Five',          6, 1, 5),   -- Full    (5)
 ('DEMO04','ALG201','Binary Birds',       11, 1, 4),   -- Ready   (4)
 ('DEMO05','ALG201','Solo Starter',       15, 1, 1),   -- Forming (1)
 ('DEMO06','WEB203','Web Wizards',         1, 3, 3),   -- members 1, 4, 7  (Ready)
 ('DEMO07','DBS202','Query Squad',         3, 3, 4);   -- members 3, 6, 9, 12 (Ready)

INSERT INTO teams (course_id, name, join_code, created_by)
SELECT c.id, t.team_name, t.code, u.id
FROM demo_teams t
JOIN courses c ON c.code = t.course
JOIN demo_users u ON u.seq = t.first_seq
ON CONFLICT (join_code) DO NOTHING;

INSERT INTO team_members (team_id, user_id, course_id)
SELECT tm.id, u.id, tm.course_id
FROM demo_teams t
JOIN teams tm ON tm.join_code = t.code
CROSS JOIN LATERAL generate_series(0, t.cnt - 1) AS g(i)
JOIN demo_users u ON u.seq = t.first_seq + g.i * t.step
ON CONFLICT DO NOTHING;

-- 5. Saved candidates (Like / Pass) in Algorithms among students without a team ------
-- students 16..25 look at 26..40: likes for 3 of every 4, passes for the rest
INSERT INTO course_choices (user_id, candidate_id, course_id, choice)
SELECT a.id, b.id, c.id, CASE WHEN (a.seq + b.seq) % 4 = 0 THEN 'pass' ELSE 'like' END
FROM demo_users a
JOIN demo_users b ON b.seq BETWEEN 26 AND 40 AND (b.seq - a.seq) BETWEEN 10 AND 14
JOIN courses c ON c.code = 'ALG201'
WHERE a.seq BETWEEN 16 AND 25
ON CONFLICT DO NOTHING;

-- 6. Academic chats with messages -----------------------------------------------------
-- (initiator, other, status, messages: sender is the initiator when 'i', the other when 'o')
CREATE TEMP TABLE demo_chats (initiator int, other int, chat_status text, msgs jsonb) ON COMMIT DROP;
INSERT INTO demo_chats VALUES
 (16, 26, 'open',
  '[["i","Привет! Тоже на Algorithms, ищешь команду на проект?"],
    ["o","Привет, да. Нам нужен кто-то на бэкенд, ты на чём пишешь?"],
    ["i","Go и PostgreSQL, могу взять API."],
    ["o","Отлично, давай создадим команду и пригласим ещё двоих."]]'),
 (17, 27, 'open',
  '[["i","Добрый день! Видел твой профиль, нужен фронтенд-разработчик."],
    ["o","Да, я на Angular. Когда хотите созвониться?"]]'),
 (18, 28, 'open',
  '[["i","Привет! Ты уже в команде?"]]'),
 (19, 29, 'open',
  '[["i","Салем! Давай объединимся на проект по алгоритмам."],
    ["o","Салем, давай. Скинь, что умеешь."],
    ["i","Тестирование и анализ данных, плюс немного Python."],
    ["o","Круто, нам как раз не хватает QA."],
    ["i","Тогда ждём ещё одного дизайнера и собираем пятёрку."]]'),
 (20, 30, 'declined',
  '[["i","Привет! Присоединишься к нам?"],
    ["o","Спасибо, мы уже собрали команду."]]'),
 (21, 31, 'open', '[]');

INSERT INTO chats (course_id, user_one_id, user_two_id, initiated_by, status, closed_by, created_at, updated_at)
SELECT c.id,
       LEAST(a.id, b.id), GREATEST(a.id, b.id),
       a.id, d.chat_status,
       CASE WHEN d.chat_status = 'declined' THEN b.id END,
       NOW() - interval '3 hours', NOW() - interval '10 minutes'
FROM demo_chats d
JOIN demo_users a ON a.seq = d.initiator
JOIN demo_users b ON b.seq = d.other
JOIN courses c ON c.code = 'ALG201'
ON CONFLICT (course_id, user_one_id, user_two_id) DO NOTHING;

INSERT INTO messages (chat_id, sender_id, content, created_at)
SELECT ch.id,
       CASE WHEN m.elem->>0 = 'i' THEN a.id ELSE b.id END,
       m.elem->>1,
       NOW() - interval '3 hours' + (m.ord * interval '7 minutes')
FROM demo_chats d
JOIN demo_users a ON a.seq = d.initiator
JOIN demo_users b ON b.seq = d.other
JOIN courses c ON c.code = 'ALG201'
JOIN chats ch ON ch.course_id = c.id
             AND ch.user_one_id = LEAST(a.id, b.id)
             AND ch.user_two_id = GREATEST(a.id, b.id)
CROSS JOIN LATERAL jsonb_array_elements(d.msgs) WITH ORDINALITY AS m(elem, ord)
WHERE NOT EXISTS (SELECT 1 FROM messages x WHERE x.chat_id = ch.id);

COMMIT;

SELECT 'users' AS what, count(*) FROM users WHERE email LIKE '%@demo.winx.local'
UNION ALL SELECT 'teams', count(*) FROM teams WHERE join_code LIKE 'DEMO%'
UNION ALL SELECT 'team members', count(*) FROM team_members m JOIN teams t ON t.id = m.team_id WHERE t.join_code LIKE 'DEMO%'
UNION ALL SELECT 'saved candidates', count(*) FROM course_choices WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@demo.winx.local')
UNION ALL SELECT 'chats', count(*) FROM chats WHERE initiated_by IN (SELECT id FROM users WHERE email LIKE '%@demo.winx.local');
