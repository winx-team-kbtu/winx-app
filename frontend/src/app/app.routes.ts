import { Routes } from '@angular/router';
import { authGuard, guestGuard } from './core/guards/auth.guard';

export const routes: Routes = [
  {
    path: 'admin',
    loadComponent: () => import('./pages/admin/admin.component').then(m => m.AdminComponent),
    canActivate: [authGuard],
  },
  {
    path: 'onboarding',
    loadComponent: () =>
      import('./pages/onboarding/onboarding.component').then((m) => m.OnboardingComponent),
    canActivate: [authGuard],
  },
  {
    path: '',
    loadComponent: () =>
      import('./pages/home-redirect.component').then((m) => m.HomeRedirectComponent),
    pathMatch: 'full',
    canActivate: [authGuard],
  },
  { path: 'team', redirectTo: 'team/discover', pathMatch: 'full' },
  {
    path: 'team/chats',
    loadComponent: () =>
      import('./pages/team/conversations.component').then((m) => m.ConversationsComponent),
    canActivate: [authGuard],
  },
  {
    path: 'team/chats/:id',
    loadComponent: () =>
      import('./pages/team/conversation-detail.component').then(
        (m) => m.ConversationDetailComponent,
      ),
    canActivate: [authGuard],
  },
  {
    path: 'team/profile',
    loadComponent: () =>
      import('./pages/team/academic-profile.component').then((m) => m.AcademicProfileComponent),
    canActivate: [authGuard],
  },
  {
    path: 'team/discover',
    loadComponent: () =>
      import('./pages/team/course-deck.component').then((m) => m.CourseDeckComponent),
    canActivate: [authGuard],
  },
  {
    path: 'team/shortlist',
    redirectTo: 'team/chats',
    pathMatch: 'full',
  },
  {
    path: 'team/teams',
    loadComponent: () => import('./pages/team/teams.component').then((m) => m.TeamsComponent),
    canActivate: [authGuard],
  },
  {
    path: 'login',
    canActivate: [guestGuard],
    loadComponent: () => import('./pages/login/login.component').then((m) => m.LoginComponent),
  },
  {
    path: 'register',
    canActivate: [guestGuard],
    loadComponent: () =>
      import('./pages/register/register.component').then((m) => m.RegisterComponent),
  },
  {
    path: 'discover',
    loadComponent: () =>
      import('./pages/discover/discover.component').then((m) => m.DiscoverComponent),
    canActivate: [authGuard],
  },
  {
    path: 'matches',
    loadComponent: () =>
      import('./pages/matches/matches.component').then((m) => m.MatchesComponent),
    canActivate: [authGuard],
  },
  {
    path: 'chats',
    loadComponent: () => import('./pages/chats/chats.component').then((m) => m.ChatsComponent),
    canActivate: [authGuard],
  },
  {
    path: 'chats/:id',
    loadComponent: () =>
      import('./pages/chat-detail/chat-detail.component').then((m) => m.ChatDetailComponent),
    canActivate: [authGuard],
  },
  {
    path: 'profile',
    loadComponent: () =>
      import('./pages/profile/profile.component').then((m) => m.ProfileComponent),
    canActivate: [authGuard],
  },
  {
    path: 'notifications',
    loadComponent: () =>
      import('./pages/notifications/notifications.component').then((m) => m.NotificationsComponent),
    canActivate: [authGuard],
  },
  { path: '**', redirectTo: '' },
];
