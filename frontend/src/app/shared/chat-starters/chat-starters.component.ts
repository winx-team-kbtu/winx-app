import { Component, EventEmitter, Input, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { IconComponent } from '../icon/icon.component';

@Component({
  selector: 'app-chat-starters',
  standalone: true,
  imports: [CommonModule, IconComponent],
  template: `
    <section aria-label="Подсказки для первого сообщения">
      <app-icon name="chat" [size]="22" />
      <h2>С чего начать?</h2>
      <button *ngFor="let starter of starters" type="button" (click)="picked.emit(starter)">
        {{ starter }} <app-icon name="arrow" [size]="15" />
      </button>
    </section>
  `,
  styles: [`
    :host { display: block; width: min(100%, 370px); margin: auto; }
    section { display: flex; flex-direction: column; gap: 9px; color: var(--ink); }
    h2 { font-size: 19px; margin-bottom: 5px; }
    section > app-icon { align-self: flex-start; color: #8c2c22; }
    button { display: flex; align-items: center; justify-content: space-between; gap: 8px;
      min-height: 44px; border: 1px solid var(--line); border-radius: 10px;
      padding: 9px 12px; text-align: left; font-size: 13px; background: var(--surface); }
    button:hover { border-color: #d8a69b; color: #8c2c22; }
  `],
})
export class ChatStartersComponent {
  @Input() starters: readonly string[] = [];
  @Output() picked = new EventEmitter<string>();
}
