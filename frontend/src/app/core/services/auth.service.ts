import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { tap } from 'rxjs/operators';

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
}

export interface AuthResponse {
  success: boolean;
  message: string;
  data: AuthTokens;
}

@Injectable({ providedIn: 'root' })
export class AuthService {
  private readonly TOKEN_KEY = 'winx_access_token';
  private readonly REFRESH_KEY = 'winx_refresh_token';

  constructor(private http: HttpClient, private router: Router) {}

  login(email: string, password: string) {
    return this.http.post<AuthResponse>('/api/v1/login', { email, password }).pipe(
      tap(res => this.saveTokens(res.data))
    );
  }

  register(email: string, password: string) {
    return this.http.post<void>('/api/v1/register', { email, password });
  }

  logout() {
    this.http.post('/api/v1/logout', {}).subscribe({ error: () => {} });
    localStorage.removeItem(this.TOKEN_KEY);
    localStorage.removeItem(this.REFRESH_KEY);
    this.router.navigate(['/login']);
  }

  refreshToken() {
    const refresh_token = localStorage.getItem(this.REFRESH_KEY);
    return this.http.post<AuthResponse>('/api/v1/refresh', { refresh_token }).pipe(
      tap(res => this.saveTokens(res.data))
    );
  }

  getToken(): string | null {
    const token = localStorage.getItem(this.TOKEN_KEY);
    return token && token !== 'undefined' && token !== 'null' ? token : null;
  }

  isLoggedIn(): boolean {
    return !!this.getToken();
  }

  // Decode JWT payload to extract user ID (sub claim)
  getUserId(): number | null {
    const token = this.getToken();
    if (!token) return null;
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return Number(payload.sub ?? payload.user_id) || null;
    } catch {
      return null;
    }
  }

  private saveTokens(tokens: AuthTokens): void {
    if (!tokens?.access_token || !tokens?.refresh_token) {
      console.error('saveTokens: invalid tokens', tokens);
      return;
    }
    localStorage.setItem(this.TOKEN_KEY, tokens.access_token);
    localStorage.setItem(this.REFRESH_KEY, tokens.refresh_token);
  }
}
