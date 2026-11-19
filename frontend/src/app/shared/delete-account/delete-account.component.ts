import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AccountService, deletionError } from '../../core/services/account.service';
import { AuthService } from '../../core/services/auth.service';
import { SheetComponent } from '../sheet/sheet.component';

/**
 * Self-service account deletion (SCRUM-55). The user must tick the confirmation and re-enter
 * their password. Outcomes are shown truthfully: wrong password and failures keep the account,
 * a pending cleanup is reported as pending, and only a finished deletion is reported as done.
 */
@Component({
  selector: 'app-delete-account',
  standalone: true,
  imports: [CommonModule, FormsModule, SheetComponent],
  template: `
    <button class="logout-link danger-link" type="button" (click)="show()">Удалить аккаунт</button>
    <app-sheet *ngIf="open" title="Удалить аккаунт" [locked]="submitting" (closed)="close()">
      <form class="sheet-form" (ngSubmit)="submit()" #form="ngForm">
        <div class="sheet-body">
          <p class="delete-lead">Удаление нельзя отменить. Будут удалены:</p>
          <ul class="delete-list">
            <li>почта, пароль и данные входа;</li>
            <li>профиль, выбранные предметы, навыки и роль;</li>
            <li>ваши выборы «нравится» и «пропустить»;</li>
            <li>участие в командах: остальные участники остаются, размер команды пересчитывается.</li>
          </ul>
          <p class="field-note">
            Вход и открытые сессии перестанут работать. Резервные копии и технические журналы могут
            храниться дольше и не стираются сразу.
          </p>
          <label class="consent-field"
            ><input
              type="checkbox"
              [(ngModel)]="understood"
              name="understood"
              [disabled]="submitting"
            /><span>Понимаю, что удаление необратимо.</span></label
          >
          <div class="form-group">
            <label for="delete-password">Пароль для подтверждения</label
            ><input
              id="delete-password"
              type="password"
              class="form-control"
              [(ngModel)]="password"
              name="password"
              autocomplete="current-password"
              maxlength="255"
              required
              [disabled]="submitting"
            />
          </div>
          <p *ngIf="submitting" class="field-note" role="status">
            Удаляем аккаунт. Не закрывайте страницу.
          </p>
          <p *ngIf="error" class="error-msg" role="alert">{{ error }}</p>
        </div>
        <div class="sheet-actions">
          <button class="btn btn-outline" type="button" [disabled]="submitting" (click)="close()">
            Отмена</button
          ><button
            class="btn btn-danger"
            type="submit"
            [disabled]="submitting || !understood || !password"
          >
            {{ submitting ? 'Удаляем…' : 'Удалить навсегда' }}
          </button>
        </div>
      </form>
    </app-sheet>
  `,
  styles: [
    `
      .danger-link {
        margin-left: 16px;
      }
      .consent-field {
        display: flex;
        align-items: flex-start;
        gap: 10px;
        margin: 12px 0 16px;
        font-size: 13px;
        line-height: 1.5;
        cursor: pointer;
      }
      .consent-field input {
        width: 16px;
        height: 16px;
        flex-shrink: 0;
        margin-top: 2px;
        accent-color: var(--accent);
      }
      .delete-lead {
        font-weight: 600;
        margin-bottom: 8px;
      }
      .delete-list {
        margin: 0 0 12px 18px;
        padding: 0;
        font-size: 14px;
        line-height: 1.6;
      }
    `,
  ],
})
export class DeleteAccountComponent {
  open = false;
  understood = false;
  password = '';
  submitting = false;
  error = '';

  constructor(
    private accounts: AccountService,
    private auth: AuthService,
    private router: Router,
  ) {}

  show(): void {
    this.reset();
    this.open = true;
  }

  close(): void {
    if (this.submitting) return;
    this.reset();
    this.open = false;
  }

  submit(): void {
    if (this.submitting || !this.understood || !this.password) return;
    const session = this.auth.sessionId();
    const password = this.password;
    this.submitting = true;
    this.error = '';
    this.accounts.delete(password).subscribe({
      next: (outcome) => {
        // The server already revoked the sessions in both outcomes; clear browser state and
        // show the result where the user lands. Pending is never presented as deleted.
        this.auth.endSessionLocally(session);
        void this.router.navigate(['/login'], {
          replaceUrl: true,
          queryParams: { reason: outcome === 'deleted' ? 'deleted' : 'deletion-pending' },
        });
      },
      error: (error) => {
        this.error = deletionError(error);
        this.password = '';
        this.submitting = false;
      },
    });
  }

  private reset(): void {
    this.understood = false;
    this.password = '';
    this.error = '';
  }
}
