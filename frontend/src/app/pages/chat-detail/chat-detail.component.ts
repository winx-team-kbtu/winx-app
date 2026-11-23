import { AfterViewChecked, Component, ElementRef, OnDestroy, OnInit, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, RouterLink } from '@angular/router';
import { forkJoin, Subject, Subscription, takeUntil } from 'rxjs';
import { ChatService, Message } from '../../core/services/chat.service';
import { ProfileService } from '../../core/services/profile.service';
import { academicError } from '../../core/services/academic.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { ChatStartersComponent } from '../../shared/chat-starters/chat-starters.component';

@Component({
  selector: 'app-chat-detail',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, IconComponent, ChatStartersComponent],
  templateUrl: './chat-detail.component.html',
  styleUrl: './chat-detail.component.css',
})
export class ChatDetailComponent implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('messagesEnd') messagesEnd?: ElementRef<HTMLElement>;
  @ViewChild('composer') composer?: ElementRef<HTMLInputElement>;

  readonly starters = [
    'Как любишь проводить выходные?',
    'Что тебя недавно порадовало?',
    'Какое место в городе стоит посетить?',
  ];
  readonly peerName = typeof history.state?.peerName === 'string' ? history.state.peerName : '';
  messages: Message[] = [];
  newMessage = '';
  error = '';
  loading = true;
  ready = false;
  pending = '';
  myId: number | null = null;
  private chatId = 0;
  private readonly destroyed = new Subject<void>();
  private wsSub?: Subscription;
  private pendingDraft = '';
  private shouldScroll = false;
  private ackTimer?: ReturnType<typeof setTimeout>;

  constructor(
    private route: ActivatedRoute,
    public chat: ChatService,
    private profileService: ProfileService,
  ) {}

  ngOnInit(): void {
    this.chatId = Number(this.route.snapshot.paramMap.get('id'));
    this.loadMessages();
    this.connect();
  }

  connect(): void {
    this.wsSub?.unsubscribe();
    this.wsSub = this.chat.connectWs().subscribe((message) => {
      if (message.type === 'error' && this.pending) {
        this.failSend();
      } else if (message.type === 'message' && Number(message.chat_id) === this.chatId && message.id) {
        this.merge([{
          id: message.id,
          chat_id: this.chatId,
          sender_id: message.sender_id ?? 0,
          content: message.content ?? '',
          created_at: message.created_at ?? new Date().toISOString(),
        }]);
      }
    });
  }

  isMine(message: Message): boolean {
    return this.myId !== null && message.sender_id === this.myId;
  }

  loadMessages(): void {
    this.loading = true;
    this.error = '';
    forkJoin({ messages: this.chat.getMessages(this.chatId), viewer: this.profileService.getMe() })
      .pipe(takeUntil(this.destroyed)).subscribe({
      next: ({ messages, viewer }) => {
        this.myId = viewer.data?.user_id ?? null;
        this.merge(messages.data ?? []);
        this.loading = false;
        this.ready = true;
      },
      error: (err) => {
        this.loading = false;
        this.ready = false;
        this.error = academicError(err, 'Не удалось загрузить переписку.');
      },
    });
  }

  draftStarter(text: string): void {
    this.newMessage = text;
    this.composer?.nativeElement.focus();
  }

  send(): void {
    const content = this.newMessage.trim();
    if (!this.ready || this.pending || !content) return;
    this.error = '';
    if (!this.chat.sendMessage(this.chatId, content)) {
      this.error = 'Нет соединения с чатом. Сообщение осталось в черновике.';
      return;
    }
    this.pending = content;
    this.pendingDraft = this.newMessage;
    this.ackTimer = setTimeout(() => this.failSend(), 8000);
  }

  private merge(incoming: Message[]): void {
    const ids = new Set(this.messages.map((message) => message.id));
    const added = incoming.filter((message) => !ids.has(message.id));
    if (!added.length) return;
    this.messages = [...this.messages, ...added].sort((a, b) => a.id - b.id);
    if (added.some((message) => this.isMine(message) && message.content === this.pending)) {
      if (this.newMessage === this.pendingDraft) this.newMessage = '';
      this.pending = '';
      this.pendingDraft = '';
      clearTimeout(this.ackTimer);
    }
    this.shouldScroll = true;
  }

  private failSend(): void {
    clearTimeout(this.ackTimer);
    this.pending = '';
    this.error = 'Отправка не подтверждена. Проверьте переписку перед повтором.';
  }

  ngAfterViewChecked(): void {
    if (this.shouldScroll && this.messagesEnd) {
      this.messagesEnd.nativeElement.scrollIntoView({ block: 'end' });
      this.shouldScroll = false;
    }
  }

  ngOnDestroy(): void {
    clearTimeout(this.ackTimer);
    this.destroyed.next();
    this.destroyed.complete();
    this.wsSub?.unsubscribe();
    this.chat.disconnectWs();
  }
}
