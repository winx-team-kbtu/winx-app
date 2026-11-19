# Sprint 2 regression run: API, mobile deck and privacy

SCRUM-63 · [S2-27] · Winx Sprint 2

**Status: partial.** The frontend part ran and passed. The backend (Go) part and the live-stack checks did not run, so the task is not complete and must not be marked Done on this evidence alone. Section 4 lists exactly what is left.

## 1. Run details

| Item | Value |
|---|---|
| Date | 2026-11-18 |
| Base commit | `1e43199` (`implemented onboarding`) plus the uncommitted SCRUM-48 and SCRUM-55 changes in the same working tree |
| Where it ran | Linux cloud sandbox (Node 22.23, Go 1.24.7). No Postgres, Redis, Kafka or MongoDB stack was started. No real mobile devices |
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

## 3. Coverage of the task scope

"Covered by tests" means a test exists in the repository. It is not a pass unless section 2 says it ran.

| Scope item | Tests in the repository | Ran here? |
|---|---|---|
| Mobile deck | `frontend/tests/swipe-card.test.cjs`; `profile/.../handlers/deck`, `services/deck`, `repositories/deck` | Frontend: yes, passed. Backend: no |
| Course isolation | `profile/.../repositories/deck` and `course` integration tests; `services/course` | No (Go, needs database) |
| Social regression | `api-gateway/api/pilot_test.go` (Social routes answer 404 in pilot mode and are reached when it is off); `match`, `chat` and `recommendation` unit tests | No (Go) |
| Membership constraints | `profile/.../services/team`, `handlers/team`, `repositories/team`, `models/team_status_test.go` | No (Go, needs database for the repository tests) |
| Privacy checklist | consent: `auth/.../register_consent_test.go` and handler test; deletion: `auth/.../services/deletion/*`, `handlers/account`, `profile/.../erasure/*`; late events: `chat` and `notification` `accountguard` tests | Frontend consent and notice: yes, passed. Backend: no |
| Session cleanup | `frontend/tests/auth-session.test.cjs`; deletion tests above | Frontend: yes, passed. Backend: no |

## 4. Not run, and why

1. **Go unit and integration tests for all seven services** (about 286 test functions). `go test` needs the Go 1.25 toolchain declared in `go.mod`, and downloading it from `proxy.golang.org` was refused by the sandbox network policy. This was not worked around. Run it on a machine with Go 1.25 and network access:
   ```bash
   cd auth && go test -short ./...   # repeat for api-gateway, profile, chat, match, notification, recommendation
   go test -tags integration ./...   # needs WINX_TEST_DSN, Postgres, Redis (see docs/team-chats.md)
   ```
2. **Live API flow** (`httprequests/e2e-flow.http`): register without consent (expect rejection), register with consent, unauthorized profile edit, deletion with a wrong and a correct password, rejection of the old session after deletion, retry of a failed cleanup. Needs the docker stack.
3. **Manual store checks after deletion** from `docs/pilot-data-inventory.md` section 9 (`scripts/pilot_user_footprint.sql`, MongoDB, Redis).
4. **Real mobile devices**: keyboard behaviour in the new deletion dialog, swipe gestures on iOS and Android. `docs/ux-audit.md` already lists device checks as open.
5. **Deletion dialog in a real browser**: only its logic and wiring are covered by tests. The visual states (confirmation, progress, error, pending notice on the sign-in page) were not looked at.

## 5. Evidence still to attach in Jira

For each item in section 4 once run: build or commit, date, test account, action, expected and actual result, and a link to the log or PR (checklist item 10). Approval from Alisher for release readiness is separate and is not claimed here.
