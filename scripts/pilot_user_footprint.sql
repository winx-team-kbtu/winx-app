-- Lists every PostgreSQL row that belongs to one pilot user.
-- Use it to inventory a user's data and to prove that account deletion left nothing behind.
--
--   psql "$DSN" -v uid=42 -v email="'student@example.com'" -f scripts/pilot_user_footprint.sql
--
-- A fully deleted user returns 0 in every row. MongoDB and Redis are not covered here;
-- see docs/pilot-data-inventory.md for the equivalent mongosh and redis-cli checks.
SELECT store, table_name, rows
FROM (
    SELECT 1 AS ord, 'auth' AS store, 'users' AS table_name, COUNT(*) AS rows
      FROM users WHERE id = :uid
    UNION ALL SELECT 2, 'auth', 'oauth_tokens', COUNT(*) FROM oauth_tokens WHERE user_id = :uid::text
    UNION ALL SELECT 3, 'auth', 'password_resets', COUNT(*) FROM password_resets WHERE email = :email

    UNION ALL SELECT 10, 'profile', 'profiles', COUNT(*) FROM profiles WHERE user_id = :uid
    UNION ALL SELECT 11, 'profile', 'profile_photos', COUNT(*) FROM profile_photos WHERE user_id = :uid
    UNION ALL SELECT 12, 'profile', 'user_interests', COUNT(*) FROM user_interests WHERE user_id = :uid
    UNION ALL SELECT 13, 'profile', 'matching_preferences', COUNT(*) FROM matching_preferences WHERE user_id = :uid
    UNION ALL SELECT 14, 'profile', 'academic_profiles', COUNT(*) FROM academic_profiles WHERE user_id = :uid
    UNION ALL SELECT 15, 'profile', 'user_courses', COUNT(*) FROM user_courses WHERE user_id = :uid
    UNION ALL SELECT 16, 'profile', 'course_choices', COUNT(*) FROM course_choices WHERE user_id = :uid OR candidate_id = :uid
    UNION ALL SELECT 17, 'profile', 'team_members', COUNT(*) FROM team_members WHERE user_id = :uid
    UNION ALL SELECT 18, 'profile', 'teams (created_by)', COUNT(*) FROM teams WHERE created_by = :uid

    UNION ALL SELECT 20, 'match', 'swipes', COUNT(*) FROM swipes WHERE swiper_id = :uid OR swiped_id = :uid
    UNION ALL SELECT 21, 'match', 'matches', COUNT(*) FROM matches WHERE user_one_id = :uid OR user_two_id = :uid

    UNION ALL SELECT 30, 'chat', 'chats', COUNT(*) FROM chats WHERE user_one_id = :uid OR user_two_id = :uid
    UNION ALL SELECT 31, 'chat', 'messages', COUNT(*) FROM messages WHERE sender_id = :uid

    UNION ALL SELECT 40, 'recommendation', 'recommendations', COUNT(*) FROM recommendations WHERE user_id = :uid OR recommended_id = :uid

    UNION ALL SELECT 50, 'notification', 'notifications', COUNT(*) FROM notifications WHERE recipient = :email
) AS footprint
ORDER BY ord;
