import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  { path: '', redirectTo: '/discover', pathMatch: 'full' },
  {
    path: 'login',
    loadComponent: () => import('./pages/login/login.component').then(m => m.LoginComponent),
  },
  {
    path: 'register',
    loadComponent: () => import('./pages/register/register.component').then(m => m.RegisterComponent),
  },
  {
    path: 'discover',
    loadComponent: () => import('./pages/discover/discover.component').then(m => m.DiscoverComponent),
    canActivate: [authGuard],
  },
  {
    path: 'matches',
    loadComponent: () => import('./pages/matches/matches.component').then(m => m.MatchesComponent),
    canActivate: [authGuard],
  },
  {
    path: 'chats',
    loadComponent: () => import('./pages/chats/chats.component').then(m => m.ChatsComponent),
    canActivate: [authGuard],
  },
  {
    path: 'chats/:id',
    loadComponent: () => import('./pages/chat-detail/chat-detail.component').then(m => m.ChatDetailComponent),
    canActivate: [authGuard],
  },
  {
    path: 'profile',
    loadComponent: () => import('./pages/profile/profile.component').then(m => m.ProfileComponent),
    canActivate: [authGuard],
  },
  {
    path: 'notifications',
    loadComponent: () => import('./pages/notifications/notifications.component').then(m => m.NotificationsComponent),
    canActivate: [authGuard],
  },
  { path: '**', redirectTo: '/discover' },
];
