# Pilot data store inventory

Scope: every place the WinX services keep data that belongs to a pilot user, who writes it, and
how account deletion (SCRUM-17) reaches it. Compiled from the migrations, service configs,
route tables and Kafka wiring on `main` plus the academic pilot changes (SCRUM-18 to SCRUM-45).

Use `scripts/pilot_user_footprint.sql` to list a user's PostgreSQL rows. A deleted user must
return 0 in every row.

## 1. PostgreSQL (database `winx-db`, shared by all services)

All services point at the same database, so `ON DELETE CASCADE` foreign keys to `users(id)`
span service boundaries.

| Service | Table | User key | Removed when `users` row is deleted |
|---|---|---|---|
| auth | `users` | `id`, `email` | the row itself |
| auth | `oauth_tokens` | `user_id` (TEXT, no FK) | **No** |
| auth | `password_resets` | `email` (no FK) | **No** |
| auth | `oauth_clients`, `roles` | none (catalogue) | n/a |
| profile | `profiles` | `user_id` | Yes (CASCADE) |
| profile | `profile_photos` | `user_id` | Yes (CASCADE), S3 objects are not |
| profile | `user_interests` | `user_id` | Yes (CASCADE) |
| profile | `matching_preferences` | `user_id` | Yes (CASCADE) |
| profile | `academic_profiles` | `user_id` | Yes (CASCADE) |
| profile | `user_courses` | `user_id` | Yes (CASCADE) |
| profile | `course_choices` | `user_id`, `candidate_id` | Yes (CASCADE on both) |
| profile | `team_members` | `user_id` | Yes (CASCADE); team size and status recalculate |
| profile | `teams` | `created_by` | Team survives, `created_by` becomes NULL; a team left with no members is **not** removed |
| profile | `courses`, `interests` | none (catalogue) | n/a |
| match | `swipes` | `swiper_id`, `swiped_id` (no FK) | **No** |
| match | `matches` | `user_one_id`, `user_two_id` (no FK) | **No** |
| chat | `chats` | `user_one_id`, `user_two_id` (no FK) | **No** |
| chat | `messages` | `sender_id` (no FK), `chat_id` CASCADE | Only via its chat |
| recommendation | `recommendations` | `user_id`, `recommended_id` (no FK) | **No** |
| notification | `notifications` | `recipient` = email (no FK), `payload` JSONB | **No** |

## 2. MongoDB (database `winx-profile`)

| Collection | Written by | Read by | User key | Removed on user delete |
|---|---|---|---|---|
| `profiles` | profile (`POST /profile/store`), recommendation seeder (manual) | profile, recommendation (`profilemeta`) | `user_id` | **No** |

## 3. Redis

All services use Redis database 0. The `cache_db: 2` value in the auth and notification configs is
not read by any code. The shared cache client prefixes every key with `users:`.

| Key | Written by | Meaning | Removed on user delete |
|---|---|---|---|
| `users:access_token:<token>` | auth (login, refresh) | token to user id | **No** until TTL (`access_token_exp`) |
| `users:user_id:<id>` | auth | current access token of the user | **No** until TTL |
| `users:user_email:<id>` | auth | email of the user | **No** until TTL |
| `pop:<id>`, `rec:<id>` | match | cached Social candidate lists, written through the raw client (prefix not verified) | **No** until TTL |
| `geoip:*` | profile | IP lookup cache, not user data | n/a |

Every service authenticates a request by reading the first three keys. Deleting them revokes
the session in all services at once. The refresh token lives only in `oauth_tokens`.

## 4. Object storage

`profile_photos` rows reference objects in the S3 bucket configured under `s3.bucket` in
`profile/config.dev.yaml`. Deleting the row does not delete the object.

## 5. Kafka topics

| Topic | Producer | Consumers | Effect |
|---|---|---|---|
| `user.registered` | auth | notification | creates a `notifications` row (welcome email) |
| `user.password` | auth | notification | creates a `notifications` row (reset PIN email) |
| `swipe.left`, `swipe.right` | none found in the code base | match | records a swipe |
| `match.created` | match | chat, notification | chat creates a chat, notification creates a row |

The recommendation service declares the `match.created` topic, but its consumer registration is
commented out, so it does not consume it today.

Consumers that create rows from `match.created` can recreate data for a user after deletion
(stale events, SCRUM-54).

## 6. Writers of sessions and events

- Sessions: only auth (`/login`, `/refresh`, `/logout`). The `notification` service contains a copy
  of the auth service code (register, login, token cache) but its router exposes only
  `/notifications`, so that code is not reachable.
- Account creation: only auth (`POST /register`, now requires consent, SCRUM-49).
- Events: auth publishes `user.registered` and `user.password`; match publishes `match.created`.
- Background work: the Kafka consumers listed above and a 5 second notification dispatch ticker.
  The recommendation `seed` binary ships in the image but is never started automatically.

## 7. Legacy Social surface (excluded from the pilot)

Exposed through the gateway today: photos (`/profile/photo`, `/profile/photo/store`), swipes,
matches, recommendations and chats (`/swipes/*`, `/matches`, `/recommendations`, `/chats`).
SCRUM-79 disables these at the gateway when `pilot.mode` is true (`api-gateway/config.dev.yaml`,
override with the `PILOT_MODE` environment variable). The setting defaults to true when it is
missing, so Social stays off unless someone sets `PILOT_MODE=false` on purpose. Disabled routes
answer 404 and never reach the downstream services. The matching tables (`swipes`, `matches`, `chats`,
`messages`, `recommendations`) must stay empty for pilot users.

## 8. Findings

1. Deleting the `users` row removes only the foreign-keyed profile and academic tables.
   Sessions, tokens, password resets, notifications, Mongo profile, Redis keys, S3 objects and the
   Social tables survive. Account deletion (SCRUM-53) must clean these explicitly.
2. A team can be left empty after its last member is deleted and must be removed explicitly.
3. Redis session keys are short lived but valid until TTL, so deletion must revoke them directly.
4. `match.created` consumers can recreate Social rows after deletion.
5. Out of scope here and to be documented separately: database backups, Kafka retention, and
   Graylog/OpenSearch logs, which can contain email addresses.

## 9. Manual checks for the stores SQL cannot see

```bash
# MongoDB
mongosh winx-profile --eval 'db.profiles.countDocuments({user_id: 42})'

# Redis database 0 (replace 42; the pattern also catches pop:42 and rec:42)
redis-cli --scan --pattern '*:42'
redis-cli EXISTS users:user_id:42 users:user_email:42
```
