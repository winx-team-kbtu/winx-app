# Account deletion in the app

SCRUM-55 · [S2-22] · Winx Sprint 2. Backend behaviour is described in `docs/pilot-data-inventory.md`, section 8.

## Where it is

**Profile → «Удалить аккаунт»**, below «Выйти из аккаунта». It is on the Team Mode profile (`/team/profile`) and on the legacy profile (`/profile`). The entry is outside the profile loading block, so it stays available if the profile fails to load.

## Flow

1. The user opens the dialog. It lists what is deleted and says that backups and technical logs may be kept longer.
2. The user ticks «Понимаю, что удаление необратимо» and enters their password. The red button stays disabled until both are done.
3. The app calls `DELETE /api/v1/account` with the password. While it runs, the dialog is locked and says so.

## Outcomes

| Server answer | What the user sees | Session |
|---|---|---|
| 200, `status: deleted` | Sign-in page: account deleted, sign-in and old sessions no longer work, backups and logs may be kept longer | cleared locally |
| 202, `status: pending` | Sign-in page: request accepted but cleanup not finished, sign-in is blocked, the server retries automatically. Never shown as deleted | cleared locally (server already revoked it) |
| 403 | Stays in the dialog: «Неверный пароль. Аккаунт не удалён.» Password field is cleared | kept |
| 401 | «Сессия завершена. Войдите снова и повторите удаление.» | handled by the existing session recovery |
| 404 | Account not found, may already be deleted; sign out and try to sign in | kept |
| 429 | Too many attempts | kept |
| 0, timeout, 5xx | Deletion not confirmed. For network errors and timeouts the text says that if sign-in stops working later, deletion is already running | kept |
| anything else, or a malformed 200/202 | Generic failure, never success | kept |

A 200 counts as success only with `status: deleted`, and a 202 as pending only with `status: pending`.

## Code

- `core/services/account.service.ts`: the request, outcome mapping and error texts.
- `shared/delete-account/delete-account.component.ts`: the dialog.
- `AuthService.endSessionLocally()`: clears browser state without calling the server again.
- `pages/login/login.component.ts`: the two notices (`reason=deleted`, `reason=deletion-pending`).

Tests: `npm run test:account` from `frontend`.
