# Sprint 2 regression run: API, mobile deck and privacy

SCRUM-63 · [S2-27] · Winx Sprint 2

**Status: partial.** All automated tests passed: frontend (45 tests and the build), Go unit tests (56 packages) and Go database integration tests (54 packages, 224 tests, 0 failures, 0 skipped). The live API flow, the manual store checks and the device checks did not run, so the task is not complete and must not be marked Done on this evidence alone. Section 4 lists exactly what is left.

## 1. Run details

| Item | Value |
|---|---|
| Date | frontend 2026-11-18, Go unit and integration tests 2026-11-19 |
| Code under test | branch `scrum-48-55-63-privacy-deletion`, commit `f9d7ecb` (`1e43199` `implemented onboarding` plus the SCRUM-48 and SCRUM-55 changes); the later commits only change this report |
| Where it ran | Frontend: Linux cloud sandbox (Node 22.23). Go tests: the developer's Mac. Integration tests ran against a throwaway copy of a migrated, unseeded Postgres 17 database with PostGIS (compose uses Postgres 16) and a local Redis, started by `run-integration-tests.sh`. No Kafka, MongoDB or application containers, no real mobile devices |
| Test account | none; no live stack was used |

## 2. What ran and the result

| Check | Command | Result |
|---|---|---|
| Browser sessions and session cleanup | `npm run test:session` | 16 of 16 passed |
| Swipe card gestures (mobile deck input) | `npm run test:swipes` | 11 of 11 passed |
| Onboarding | `npm run test:onboarding` | 5 of 5 passed |
| Account deletion UI logic (SCRUM-55) | `npm run test:account` | 8 of 8 passed |
| Notice and consent (SCRUM-48) | `npm run test:privacy` | 5 of 5 passed |
| Production build | `npx ng build` | succeeded, no errors |
| Formatting of the changed Go file | `gofmt -l` | no differences |
| Go unit tests, all services | `go test -short ./...` in `auth`, `api-gateway`, `profile`, `chat`, `match`, `notification`, `recommendation` | 56 packages `ok`, no failures. Every line says `(cached)`: Go reused an earlier passing result for identical code and inputs, it did not rerun them. `recommendation` has no tests |
| Go integration tests | `go test -count=1 -v -tags integration ./...` in `auth`, `profile`, `chat`, `notification` | 54 packages `ok`, 224 tests passed, 0 failed, 0 skipped, no cache. Includes all database tests listed in section 3 |

## 3. Coverage of the task scope

| Scope item | Tests | Result |
|---|---|---|
| Mobile deck | `frontend/tests/swipe-card.test.cjs`; `profile` `handlers/deck`, `services/deck`, `repositories/deck` (course filter, self excluded, teamed students leave the deck) | Passed |
| Course isolation | `profile` `handlers/course`, `services/course` (`TestSelectionsAreIsolatedBetweenUsers`), `repositories/course`, `repositories/deck` | Passed, including the database queries |
| Social regression | `api-gateway/api/pilot_test.go` (social routes answer 404 in pilot mode and are reached when it is off); `match`, `chat` and `api-gateway` module tests; `chat` `repositories/message` (`TestAcademicMessagePermissionsAndSocialIsolation`) | Passed |
| Membership constraints | `profile` `handlers/team`, `services/team`, `repositories/team` (five-member limit under race, duplicate membership, join code collision, roster visible to members only), `repositories/conversation` | Passed |
| Privacy checklist | consent: `auth` `handlers`, `services`, `repositories/user`; deletion: `auth` `services/deletion` (including the retry worker and blocked sign-in), `handlers/account`, `profile` `repositories/erasure`, `services/erasure`; late events after deletion: `chat` and `notification` `accountguard`, `notification` consumers; frontend consent and notice tests | Passed |
| Session cleanup | `frontend/tests/auth-session.test.cjs`; `auth` token and middleware tests; `auth` deletion tests (sessions revoked) | Passed |

"Passed" means the test ran in section 2 and succeeded. It does not replace a run against the live stack.

## 4. Not run, and why

1. **Live API flow** (`httprequests/e2e-flow.http`): register without consent (expect rejection), register with consent, unauthorized profile edit, deletion with a wrong and a correct password, rejection of the old session after deletion, retry of a failed cleanup. Needs the full stack (services, Kafka, MongoDB).
2. **Manual store checks after deletion** from `docs/pilot-data-inventory.md` section 9 (`scripts/pilot_user_footprint.sql`, MongoDB, Redis). Backups, Kafka and logs are outside the deletion and are not tested.
3. **Real mobile devices**: keyboard behaviour in the new deletion dialog, swipe gestures on iOS and Android. `docs/ux-audit.md` already lists device checks as open.
4. **Deletion dialog in a real browser**: only its logic and wiring are covered by tests. The visual states (confirmation, progress, error, pending notice on the sign-in page) were not looked at.
5. **Seeded data and Postgres 16**: the integration run used a schema-only copy on Postgres 17, not the demo seed and not the compose version.

## 5. Evidence still to attach in Jira

For each item in section 4 once run: build or commit, date, test account, action, expected and actual result, and a link to the log or PR (checklist item 10). Approval from Alisher for release readiness is separate and is not claimed here.
