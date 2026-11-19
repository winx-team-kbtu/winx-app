# Sprint 2 regression run: API, mobile deck and privacy

SCRUM-63 · [S2-27] · Winx Sprint 2

**Status: partial.** Frontend tests and the build passed, and the Go unit tests of all services passed (56 packages, `-short`). The database integration tests, the live API flow and the device checks did not run, so the task is not complete and must not be marked Done on this evidence alone. Section 4 lists exactly what is left.

## 1. Run details

| Item | Value |
|---|---|
| Date | frontend 2026-11-18, Go unit tests 2026-11-19 |
| Code under test | branch `scrum-48-55-63-privacy-deletion` (`1e43199` `implemented onboarding` plus the SCRUM-48 and SCRUM-55 changes) |
| Where it ran | Frontend: Linux cloud sandbox (Node 22.23). Go tests: the developer's Mac (`go test -short`, output in `go-test-full.txt`; Go version not recorded). No Postgres, Redis, Kafka or MongoDB stack was started. No real mobile devices |
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

## 3. Coverage of the task scope

"Passed" means the test ran in section 2 and succeeded. "Not run" means the test exists in the repository but did not run.

| Scope item | Tests in the repository | Result |
|---|---|---|
| Mobile deck | `frontend/tests/swipe-card.test.cjs`; `profile` `handlers/deck`, `services/deck` | Passed. `repositories/deck` (database) not run |
| Course isolation | `profile` `handlers/course`, `services/course` | Passed at handler and service level. `repositories/course` and `repositories/deck` (database queries) not run, so isolation in the SQL itself is not verified |
| Social regression | `api-gateway/api/pilot_test.go` (social routes answer 404 in pilot mode and are reached when it is off); `match`, `chat` and `api-gateway` module tests | Passed |
| Membership constraints | `profile` `handlers/team`, `services/team`, `models` | Passed. `repositories/team` (database) not run |
| Privacy checklist | consent: `auth` `handlers`, `services`; deletion: `auth` `services/deletion`, `handlers/account`, `profile` `handlers/erasure`, `services/erasure`; frontend consent and notice tests | Passed. Database tests for deletion (`auth` deletion integration, `profile` `repositories/erasure`) and for late events after deletion (`chat` and `notification` `accountguard`, `notification` consumers) not run |
| Session cleanup | `frontend/tests/auth-session.test.cjs`; `auth` token and middleware tests | Passed |

## 4. Not run, and why

1. **Database integration tests** (`*_integration_test.go`): `auth` repository and deletion, `profile` repositories (`academic`, `choice`, `conversation`, `course`, `deck`, `erasure`, `team`), `chat` repository and `accountguard`, `notification` `accountguard` and consumers. They are excluded from `-short` and need Postgres and Redis. Start the stack from `docs/team-chats.md`, then run `go test -tags integration ./...` in each service.
2. **Live API flow** (`httprequests/e2e-flow.http`): register without consent (expect rejection), register with consent, unauthorized profile edit, deletion with a wrong and a correct password, rejection of the old session after deletion, retry of a failed cleanup. Needs the docker stack.
3. **Manual store checks after deletion** from `docs/pilot-data-inventory.md` section 9 (`scripts/pilot_user_footprint.sql`, MongoDB, Redis).
4. **Real mobile devices**: keyboard behaviour in the new deletion dialog, swipe gestures on iOS and Android. `docs/ux-audit.md` already lists device checks as open.
5. **Deletion dialog in a real browser**: only its logic and wiring are covered by tests. The visual states (confirmation, progress, error, pending notice on the sign-in page) were not looked at.

## 5. Evidence still to attach in Jira

For each item in section 4 once run: build or commit, date, test account, action, expected and actual result, and a link to the log or PR (checklist item 10). Approval from Alisher for release readiness is separate and is not claimed here.
