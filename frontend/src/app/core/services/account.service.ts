import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Observable, map, timeout } from 'rxjs';

/**
 * `deleted`: the server finished the cleanup (HTTP 200).
 * `pending`: the account is blocked and its sessions are revoked, but cleanup is still
 * being retried on the server (HTTP 202). It must never be shown as a success.
 */
export type DeletionOutcome = 'deleted' | 'pending';

/** Maps the DELETE /api/v1/account response to an outcome; anything unexpected throws. */
export function deletionOutcome(status: number, body: unknown): DeletionOutcome {
  const reported = (body as { data?: { status?: unknown } } | null)?.data?.status;
  if (status === 200 && reported === 'deleted') return 'deleted';
  if (status === 202 && reported === 'pending') return 'pending';
  throw new HttpErrorResponse({ status: 502, error: { message: 'Unexpected deletion response' } });
}

/** Honest, user-facing text for a failed deletion request. The account is not deleted. */
export function deletionError(error: unknown): string {
  if (error instanceof HttpErrorResponse) {
    if (error.status === 403) return 'Неверный пароль. Аккаунт не удалён.';
    if (error.status === 401) return 'Сессия завершена. Войдите снова и повторите удаление.';
    if (error.status === 404)
      return 'Аккаунт не найден. Возможно, он уже удалён: выйдите и попробуйте войти снова.';
    if (error.status === 429) return 'Слишком много попыток. Попробуйте немного позже.';
    if (error.status === 0)
      return 'Нет связи с сервером. Удаление не подтверждено. Если позже войти не получится, удаление уже выполняется.';
    if (error.status >= 500)
      return 'Сервер временно недоступен. Удаление не подтверждено, попробуйте ещё раз.';
    return 'Не удалось удалить аккаунт. Попробуйте ещё раз.';
  }
  if ((error as { name?: unknown } | null)?.name === 'TimeoutError')
    return 'Сервер не ответил вовремя. Удаление не подтверждено. Если позже войти не получится, удаление уже выполняется.';
  return 'Не удалось удалить аккаунт. Попробуйте ещё раз.';
}

@Injectable({ providedIn: 'root' })
export class AccountService {
  constructor(private http: HttpClient) {}

  /** Deletes the signed-in account after the user confirmed with their password. */
  delete(password: string): Observable<DeletionOutcome> {
    return this.http
      .request<unknown>('DELETE', '/api/v1/account', {
        body: { password },
        observe: 'response',
      })
      .pipe(
        timeout(30000),
        map((response) => deletionOutcome(response.status, response.body)),
      );
  }
}
