import { Component } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { NavbarComponent } from './shared/navbar/navbar.component';
import { AuthService } from './core/services/auth.service';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, NavbarComponent, CommonModule],
  template: `
    <app-navbar *ngIf="auth.isLoggedIn()"></app-navbar>
    <router-outlet></router-outlet>
  `,
  styles: [`
    :host { display: block; min-height: 100vh; }
  `]
})
export class AppComponent {
  constructor(public auth: AuthService) {}
}
