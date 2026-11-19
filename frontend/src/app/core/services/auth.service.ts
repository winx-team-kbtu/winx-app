import { Injectable, NgZone, OnDestroy, signal } from '@angular/core';
import { HttpBackend, HttpClient, HttpErrorResponse } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, Subject, defer, firstValueFrom, from, of, tap, timeout } from 'rxjs';
import {
  AUTH_STORAGE_KEY,
  AuthSession,
  SessionChangedError,
  emptySession,
  readAuthSession,
  safeReturnUrl,
  saveAuthSession,
} from './auth-session';

export interface AuthTokens {
  access_token: string;
  refresh_token?: string;
  token_type: string;
  expires_in?: number;
}
export interface AuthResponse {
  success: boolean;
  message: string;
  data: AuthTokens;
}

@Injectable({ providedIn: 'root' })
export class AuthService implements OnDestroy {
  private readonly http: HttpClient;
  private readonly session = signal<AuthSession>(readAuthSession() ?? emptySession());
  private readonly tokenEvents = new Subject<string | null>();
  readonly tokenChanges$ = this.tokenEvents.asObservable();
  private refreshTask: { id: string; promise: Promise<string> } | null = null;
  private revision = 0;
  private stored = true;
  private readonly onStorage = (event: StorageEvent) => {
    if (event.key === AUTH_STORAGE_KEY || event.key === null) this.syncSession();
  };
  private readonly onFocus = () => this.syncSession();

  constructor(
    backend: HttpBackend,
    private router: Router,
    private zone: NgZone,
  ) {
    // Auth endpoints bypass recovery to avoid refresh loops and logout races.
    this.http = new HttpClient(backend);
    window.addEventListener('storage', this.onStorage);
    window.addEventListener('focus', this.onFocus);
    window.addEventListener('pageshow', this.onFocus);
  }

  ngOnDestroy(): void {
    window.removeEventListener('storage', this.onStorage);
    window.removeEventListener('focus', this.onFocus);
    window.removeEventListener('pageshow', this.onFocus);
    this.tokenEvents.complete();
  }

  isLoggedIn(): boolean {
    return !!this.session().id;
  }
  getToken(): string | null {
    return this.session().access;
  }
  getEmail(): string | null {
    return this.session().email;
  }
  sessionId(): string | null {
    this.syncSession();
    return this.session().id;
  }

  login(email: string, password: string): Observable<AuthResponse> {
    return defer(() => {
      this.syncSession();
      const revision = ++this.revision;
      return this.http.post<AuthResponse>('/api/v1/login', { email, password }).pipe(
        timeout(15000),
        tap((response) => {
          if (revision !== this.revision) throw new SessionChangedError();
          const session = this.fromTokens(response.data, {
            ...emptySession(),
            id: crypto.randomUUID(),
            email: email.trim(),
          });
          if (!session.refresh) throw this.invalidResponse();
          this.commit(session);
        }),
      );
    });
  }

  register(email: string, password: string, consentAccepted: boolean) {
    return this.http
      .post<void>('/api/v1/register', { email, password, consent_accepted: consentAccepted })
      .pipe(timeout(15000));
  }

  accessForRequest(): Observable<string | null> {
    return defer(() => {
      this.syncSession();
      const session = this.session();
      if (
        session.refresh &&
        (!session.access || (session.expiresAt !== null && session.expiresAt <= Date.now() + 30000))
      ) {
        return this.refreshSession();
      }
      return of(session.access);
    });
  }

  refreshSession(failedAccess?: string | null): Observable<string> {
    return defer(() => from(this.refreshAccess(failedAccess)));
  }

  private refreshAccess(failedAccess?: string | null): Promise<string> {
    this.syncSession();
    const snapshot = this.session();
    if (failedAccess && snapshot.access && snapshot.access !== failedAccess)
      return Promise.resolve(snapshot.access);
    if (!snapshot.id || !snapshot.refresh) {
      this.expireSession(snapshot.id);
      return Promise.reject(new HttpErrorResponse({ status: 401 }));
    }
    if (this.refreshTask?.id === snapshot.id) return this.refreshTask.promise;
    const promise = this.withRefreshLock(async () => {
      this.assertSession(snapshot.id);
      const current = this.session();
      // Another tab may have refreshed while this one waited for the lock.
      if (current.access && current.access !== snapshot.access) return current.access;
      try {
        const response = await firstValueFrom(
          this.http
            .post<AuthResponse>('/api/v1/refresh', { refresh_token: current.refresh })
            .pipe(timeout(15000)),
        );
        const refreshed = this.fromTokens(response.data, current);
        if (this.sessionId() !== snapshot.id) {
          // A completed refresh must never revive a logged-out session.
          void this.endServerSession(refreshed);
          throw new SessionChangedError();
        }
        this.commit(refreshed);
        return refreshed.access!;
      } catch (error) {
        this.assertSession(snapshot.id);
        if (error instanceof HttpErrorResponse && [400, 401, 403, 404, 422].includes(error.status))
          this.expireSession(snapshot.id);
        throw error;
      }
    }).finally(() => {
      if (this.refreshTask?.promise === promise) this.refreshTask = null;
    });
    this.refreshTask = { id: snapshot.id, promise };
    return promise;
  }

  assertSession(id: string | null): void {
    if (this.sessionId() !== id) throw new SessionChangedError();
  }

  expireSession(id: string | null): void {
    if (this.sessionId() !== id) return;
    const returnUrl = safeReturnUrl(this.router.url);
    this.revision++;
    this.commit(emptySession());
    this.zone.run(
      () =>
        void this.router.navigate(['/login'], {
          replaceUrl: true,
          queryParams: { reason: 'expired', returnUrl },
        }),
    );
  }

  /**
   * Clears browser state after the server has already revoked this session (account deletion).
   * Unlike logout() it does not call the server again and does not navigate.
   */
  endSessionLocally(id: string | null): void {
    if (this.sessionId() !== id) return;
    this.revision++;
    this.commit(emptySession());
  }

  logout(): void {
    this.syncSession();
    const snapshot = this.session();
    this.revision++;
    this.commit(emptySession());
    // Local sign-out is immediate even if the server is offline.
    void this.router.navigate(['/login'], { replaceUrl: true });
    void this.withRefreshLock(() => this.endServerSession(snapshot));
  }

  private async endServerSession(session: AuthSession): Promise<void> {
    const revoke = (access: string) =>
      firstValueFrom(
        this.http
          .post('/api/v1/logout', {}, { headers: { Authorization: 'Bearer ' + access } })
          .pipe(timeout(8000)),
      );
    try {
      if (session.access) {
        try {
          await revoke(session.access);
          return;
        } catch (error) {
          if (!(error instanceof HttpErrorResponse && error.status === 401)) return;
        }
      }
      // An expired access token can still have a live refresh credential.
      // Use it only for revocation, never to restore browser state.
      if (session.refresh) {
        const response = await firstValueFrom(
          this.http
            .post<AuthResponse>('/api/v1/refresh', { refresh_token: session.refresh })
            .pipe(timeout(8000)),
        );
        if (response.data?.access_token) await revoke(response.data.access_token);
      }
    } catch {
      /* Signing out never depends on a successful network request. */
    }
  }

  private withRefreshLock<T>(action: () => Promise<T>): Promise<T> {
    // Web Locks callbacks are not patched by Zone.js in Angular 17.
    return navigator.locks
      ? navigator.locks.request('winx-auth-refresh', () => this.zone.run(action))
      : action();
  }

  private fromTokens(tokens: AuthTokens, previous: AuthSession): AuthSession {
    if (!tokens || typeof tokens.access_token !== 'string' || !tokens.access_token.trim())
      throw this.invalidResponse();
    return {
      ...previous,
      access: tokens.access_token,
      refresh:
        typeof tokens.refresh_token === 'string' && tokens.refresh_token.trim()
          ? tokens.refresh_token
          : previous.refresh,
      expiresAt:
        typeof tokens.expires_in === 'number' && tokens.expires_in > 0
          ? Date.now() + tokens.expires_in * 1000
          : null,
    };
  }

  private invalidResponse(): HttpErrorResponse {
    return new HttpErrorResponse({
      status: 502,
      error: { message: 'Не удалось получить данные сессии. Попробуйте ещё раз.' },
    });
  }

  private commit(session: AuthSession): void {
    this.zone.run(() => {
      const previous = this.session();
      this.stored = saveAuthSession(session);
      this.session.set(session);
      if (previous.access !== session.access) this.tokenEvents.next(session.access);
    });
  }

  private syncSession(): void {
    if (!this.stored) return;
    const saved = readAuthSession();
    if (!saved) return;
    const previous = this.session();
    if (JSON.stringify(saved) === JSON.stringify(previous)) return;
    this.session.set(saved);
    if (previous.access !== saved.access) this.tokenEvents.next(saved.access);
    if (previous.id !== saved.id) {
      this.revision++;
      if (!saved.id) {
        this.zone.run(
          () =>
            void this.router.navigate(['/login'], {
              replaceUrl: true,
              queryParams: { reason: 'signed-out', returnUrl: safeReturnUrl(this.router.url) },
            }),
        );
      } else {
        // Drop account-specific views and pending requests after an account switch.
        window.location.reload();
      }
    }
  }
}
