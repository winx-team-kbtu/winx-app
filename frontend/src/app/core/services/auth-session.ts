export const AUTH_STORAGE_KEY = 'winx_auth';

export interface AuthSession {
  id: string | null;
  access: string | null;
  refresh: string | null;
  expiresAt: number | null;
  email: string | null;
}

export const emptySession = (): AuthSession => ({
  id: null,
  access: null,
  refresh: null,
  expiresAt: null,
  email: null,
});

function token(value: unknown): string | null {
  return typeof value === 'string' && value.trim() && value !== 'undefined' && value !== 'null'
    ? value.trim()
    : null;
}

function normalize(value: Partial<AuthSession> | null): AuthSession {
  const access = token(value?.access),
    refresh = token(value?.refresh);
  if (!access && !refresh) return emptySession();
  return {
    id: token(value?.id) || crypto.randomUUID(),
    access,
    refresh,
    expiresAt:
      typeof value?.expiresAt === 'number' && Number.isFinite(value.expiresAt)
        ? value.expiresAt
        : null,
    email: token(value?.email),
  };
}

function parse(raw: string): AuthSession {
  try {
    return normalize(JSON.parse(raw));
  } catch {
    return emptySession();
  }
}

function clearLegacy(): void {
  for (const key of ['winx_access_token', 'winx_refresh_token']) {
    try {
      localStorage.removeItem(key);
    } catch {}
  }
}

export function saveAuthSession(session: AuthSession): boolean {
  const value = JSON.stringify(session);
  try {
    // One write keeps rotated access/refresh tokens and account identity together.
    // Keep the empty record on logout so old tabs cannot restore legacy tokens.
    localStorage.setItem(AUTH_STORAGE_KEY, value);
    clearLegacy();
    try {
      sessionStorage.removeItem(AUTH_STORAGE_KEY);
    } catch {}
    return true;
  } catch {
    try {
      localStorage.removeItem(AUTH_STORAGE_KEY);
    } catch {}
    clearLegacy();
    try {
      sessionStorage.setItem(AUTH_STORAGE_KEY, value);
      return true;
    } catch {
      return false;
    }
  }
}

// null means unavailable storage; it must not erase an in-memory login.
export function readAuthSession(): AuthSession | null {
  let readable = false;
  try {
    const raw = localStorage.getItem(AUTH_STORAGE_KEY);
    readable = true;
    if (raw !== null) {
      clearLegacy();
      return parse(raw);
    }
    const access = token(localStorage.getItem('winx_access_token'));
    const refresh = token(localStorage.getItem('winx_refresh_token'));
    if (access || refresh) {
      const session = normalize({ access, refresh });
      saveAuthSession(session);
      return session;
    }
  } catch {}
  try {
    const raw = sessionStorage.getItem(AUTH_STORAGE_KEY);
    if (raw !== null) return parse(raw);
    readable = true;
  } catch {}
  return readable ? emptySession() : null;
}

export function safeReturnUrl(value: string | null | undefined): string | null {
  if (!value || value.includes('\\')) return null;
  return /^\/(?:admin|discover|matches|chats(?:\/\d+)?|notifications|profile|team\/(?:discover|shortlist|chats(?:\/\d+)?|teams|profile))(?:\?[^#]*)?(?:#.*)?$/.test(
    value,
  )
    ? value
    : null;
}

export class SessionChangedError extends Error {
  constructor() {
    super('Session changed');
    this.name = 'SessionChangedError';
  }
}
