import {
  AfterViewInit,
  Component,
  ElementRef,
  EventEmitter,
  Input,
  OnDestroy,
  Output,
  ViewChild,
} from '@angular/core';
import { IconComponent } from '../icon/icon.component';

@Component({
  selector: 'app-sheet',
  standalone: true,
  imports: [IconComponent],
  template: `
    <dialog
      #dialog
      [class.cover-sheet]="cover"
      aria-labelledby="sheet-title"
      (cancel)="dismiss($event)"
      (click)="backdrop($event)"
    >
      <section class="sheet-frame">
        <header>
          <h2 id="sheet-title" [class.sr-only]="cover">{{ title }}</h2>
          <ng-content select="[sheet-cover]" />
          <button
            type="button"
            class="icon-button"
            aria-label="Закрыть"
            [disabled]="locked"
            (click)="dismiss()"
          >
            <app-icon name="close" [size]="20" />
          </button>
        </header>
        <div class="sheet-content"><ng-content /></div>
      </section>
    </dialog>
  `,
  styles: [
    `
      dialog {
        width: min(520px, calc(100% - 40px));
        max-height: calc(100dvh - 40px);
        margin: auto;
        padding: 0;
        border: 1px solid var(--line);
        border-radius: 14px;
        background: var(--surface);
        color: var(--ink);
      }
      dialog::backdrop {
        background: rgba(22, 20, 15, 0.32);
      }
      .sheet-frame {
        display: flex;
        flex-direction: column;
        max-height: calc(100dvh - 42px);
      }
      header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        gap: 16px;
        padding: 18px 24px;
        border-bottom: 1px solid var(--line);
        flex-shrink: 0;
      }
      h2 {
        font-size: 18px;
        font-weight: 700;
        letter-spacing: -0.025em;
      }
      header .icon-button {
        width: 32px;
        height: 32px;
        border: 0;
      }
      .sheet-content {
        display: flex;
        flex-direction: column;
        min-height: 0;
        overflow: hidden;
      }
      .cover-sheet {
        width: min(600px, calc(100% - 32px));
        border-radius: 22px;
        box-shadow: 0 20px 64px rgba(30, 20, 13, 0.16);
      }
      .cover-sheet .sheet-frame {
        display: block;
        overflow: auto;
        overscroll-behavior: contain;
      }
      .cover-sheet header {
        display: block;
        position: relative;
        padding: 0;
      }
      .cover-sheet header .icon-button {
        position: absolute;
        top: 14px;
        right: 14px;
        width: 40px;
        height: 40px;
        border: 1px solid var(--line);
        border-radius: 50%;
        background: var(--surface);
        color: var(--ink);
      }
      .cover-sheet header .icon-button:hover {
        background: var(--soft);
      }
      @media (max-width: 640px) {
        dialog {
          width: 100%;
          max-width: 100%;
          margin: auto 0 0;
          border-radius: 16px 16px 0 0;
          max-height: calc(100dvh - 20px);
        }
        .sheet-frame {
          max-height: calc(100dvh - 22px);
          padding-bottom: env(safe-area-inset-bottom);
        }
        .cover-sheet {
          width: 100%;
          border-radius: 22px 22px 0 0;
        }
        header {
          padding: 16px 20px;
        }
      }
    `,
  ],
})
export class SheetComponent implements AfterViewInit, OnDestroy {
  @Input() title = '';
  @Input() cover = false;
  @Input() locked = false;
  @Output() closed = new EventEmitter<void>();
  @ViewChild('dialog', { static: true }) dialog!: ElementRef<HTMLDialogElement>;
  ngAfterViewInit(): void {
    this.dialog.nativeElement.showModal();
  }
  ngOnDestroy(): void {
    this.dialog.nativeElement.close();
  }
  dismiss(event?: Event): void {
    event?.preventDefault();
    if (!this.locked) this.closed.emit();
  }
  backdrop(event: MouseEvent): void {
    if (event.target === this.dialog.nativeElement) this.dismiss();
  }
}
