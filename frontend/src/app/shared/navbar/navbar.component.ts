import { Component } from '@angular/core';
import { RouterLink, RouterLinkActive } from '@angular/router';
import { AuthService } from '../../core/services/auth.service';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive],
  template: `
    <nav class="navbar">
      <div class="navbar-brand">
        <span class="sparkle">✦</span> Winx
      </div>
      <div class="navbar-links">
        <a routerLink="/discover" routerLinkActive="active" class="nav-link">
          <span class="nav-icon">🔥</span>
          <span class="nav-label">Explore</span>
        </a>
        <a routerLink="/matches" routerLinkActive="active" class="nav-link">
          <span class="nav-icon">💕</span>
          <span class="nav-label">Matches</span>
        </a>
        <a routerLink="/chats" routerLinkActive="active" class="nav-link">
          <span class="nav-icon">💬</span>
          <span class="nav-label">Chats</span>
        </a>
        <a routerLink="/notifications" routerLinkActive="active" class="nav-link">
          <span class="nav-icon">🔔</span>
          <span class="nav-label">Notifs</span>
        </a>
        <a routerLink="/profile" routerLinkActive="active" class="nav-link">
          <span class="nav-icon">👤</span>
          <span class="nav-label">Profile</span>
        </a>
      </div>
      <button class="logout-btn" (click)="auth.logout()">Logout</button>
    </nav>
  `,
  styleUrl: './navbar.component.css',
})
export class NavbarComponent {
  constructor(public auth: AuthService) {}
}
