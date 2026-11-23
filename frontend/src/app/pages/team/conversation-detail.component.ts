import {
  AfterViewChecked,
  Component,
  ElementRef,
  OnDestroy,
  OnInit,
  ViewChild,
} from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { Subject, forkJoin, interval, takeUntil } from 'rxjs';
import {
  AcademicService,
  AcademicConversation,
  ConversationDecision,
  academicError,
} from '../../core/services/academic.service';
import { ChatService, Message } from '../../core/services/chat.service';
import { ProfileService } from '../../core/services/profile.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { ChatStartersComponent } from '../../shared/chat-starters/chat-starters.component';
import { courseName } from '../../shared/text';

@Component({
  selector: 'app-team-conversation',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, IconComponent, SheetComponent, ChatStartersComponent],
  templateUrl: './conversation-detail.component.html',
  styleUrls: ['../chat-detail/chat-detail.component.css', './conversation-detail.component.css'],
})
export class ConversationDetailComponent implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('messagesEnd') messagesEnd?: ElementRef<HTMLElement>;
  @ViewChild('messageList') messageList?: ElementRef<HTMLElement>;
  @ViewChild('composer') composer?: ElementRef<HTMLInputElement>;
  room: AcademicConversation | null = null;
  messages: Message[] = [];
  myId = 0;
  loading = true;
  error = '';
  newMessage = '';
  panel: 'invite' | 'decline' | null = null;
  busy = false;
  decisionError = '';
  teamName = '';
  pending = '';
  loadingOlder = false;
  canLoadOlder = false;
  courseName = courseName;
  readonly starters = ['Привет! Какую роль хочешь взять в проекте?', 'Есть идеи для проекта?', 'Когда удобно обсудить проект?'];
  private chatId = 0;
  private refreshBusy = false;
  private roomRevision = 0;
  private pendingDraft = '';
  private shouldScroll = false;
  private readonly destroyed = new Subject<void>();
  private ackTimer?: ReturnType<typeof setTimeout>;
  constructor(
    private api: AcademicService,
    public chat: ChatService,
    private profiles: ProfileService,
    private route: ActivatedRoute,
    private router: Router,
  ) {}
  ngOnInit(): void {
    this.chatId = Number(this.route.snapshot.paramMap.get('id'));
    forkJoin({
      room: this.api.conversation(this.chatId),
      user: this.profiles.getMe(),
      messages: this.api.conversationMessages(this.chatId),
    })
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          this.room = res.room.data;
          this.myId = res.user.data?.user_id ?? 0;
          this.messages = res.messages.data ?? [];
          this.canLoadOlder = this.messages.length === 50;
          this.loading = false;
          this.shouldScroll = true;
          this.chat
            .connectWs()
            .pipe(takeUntil(this.destroyed))
            .subscribe((msg) => {
              if (msg.type === 'message' && Number(msg.chat_id) === this.chatId && msg.id) {
                this.merge([
                  {
                    id: msg.id,
                    chat_id: this.chatId,
                    sender_id: msg.sender_id ?? 0,
                    content: msg.content ?? '',
                    created_at: msg.created_at ?? new Date().toISOString(),
                  },
                ]);
              } else if (msg.type === 'error' && this.pending) {
                this.failSend();
              }
            });
          interval(3000)
            .pipe(takeUntil(this.destroyed))
            .subscribe(() => {
              if (!document.hidden) this.refresh();
            });
        },
        error: (err) => {
          this.loading = false;
          this.error = academicError(err);
        },
      });
  }
  refresh(): void {
    if (this.refreshBusy || this.busy) return;
    this.refreshBusy = true;
    const revision = this.roomRevision;
    forkJoin({
      room: this.api.conversation(this.chatId),
      messages: this.api.conversationMessages(this.chatId),
    })
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          this.refreshBusy = false;
          if (revision !== this.roomRevision || this.busy) return;
          this.room = res.room.data;
          this.merge(res.messages.data ?? []);
          this.refreshBusy = false;
        },
        error: () => {
          this.refreshBusy = false;
        },
      });
  }
  private merge(incoming: Message[]): void {
    const known = new Set(this.messages.map((m) => m.id));
    const added = incoming.filter((m) => !known.has(m.id));
    if (!added.length) return;
    const nearBottom =
      !this.messageList ||
      this.messageList.nativeElement.scrollHeight -
        this.messageList.nativeElement.scrollTop -
        this.messageList.nativeElement.clientHeight <
        100;
    this.messages = [...this.messages, ...added].sort((a, b) => a.id - b.id);
    for (const msg of added) {
      if (msg.sender_id === this.myId && msg.content === this.pending) {
        if (this.newMessage === this.pendingDraft) this.newMessage = '';
        this.pendingDraft = '';
        this.pending = '';
        clearTimeout(this.ackTimer);
        this.error = '';
      }
    }
    this.shouldScroll = nearBottom || added.some((m) => m.sender_id === this.myId);
  }
  loadOlder(): void {
    if (this.loadingOlder || !this.messages.length) return;
    this.loadingOlder = true;
    this.api
      .conversationMessages(this.chatId, this.messages[0].id)
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          const older = res.data ?? [];
          const known = new Set(this.messages.map((m) => m.id));
          this.messages = [...older.filter((m) => !known.has(m.id)), ...this.messages];
          this.canLoadOlder = older.length === 50;
          this.loadingOlder = false;
        },
        error: (err) => {
          this.loadingOlder = false;
          this.error = academicError(err);
        },
      });
  }
  send(): void {
    const content = this.newMessage.trim();
    if (!content || this.pending || !this.room?.can_write) return;
    this.error = '';
    if (!this.chat.sendMessage(this.chatId, content)) {
      this.error = 'Нет соединения с чатом. Попробуйте ещё раз.';
      return;
    }
    this.pending = content;
    this.pendingDraft = this.newMessage;
    this.ackTimer = setTimeout(() => this.failSend(), 8000);
  }
  draftStarter(text: string): void {
    this.newMessage = text;
    this.composer?.nativeElement.focus();
  }
  private failSend(): void {
    clearTimeout(this.ackTimer);
    this.pending = '';
    this.error = 'Не удалось отправить сообщение. Попробуйте ещё раз.';
    this.refresh();
  }
  open(panel: 'invite' | 'decline'): void {
    this.panel = panel;
    this.decisionError = '';
  }
  close(): void {
    if (!this.busy) {
      this.panel = null;
      this.decisionError = '';
    }
  }
  get otherTeam(): boolean {
    return !!this.room?.peer_team_id && this.room.peer_team_id !== this.room.my_team_id;
  }
  decide(): void {
    if (!this.room || !this.panel || this.busy) return;
    if (this.panel === 'invite' && !this.room.my_team_id && this.teamName.trim().length < 2) return;
    this.submitDecision(this.panel, this.teamName.trim());
  }
  answerInvitation(decision: 'accept' | 'reject'): void {
    if (this.busy || !this.room?.invitation_id || this.room.invitation_status !== 'pending') return;
    this.submitDecision(decision, '', this.room.invitation_id);
  }
  private submitDecision(decision: ConversationDecision, teamName: string, invitationID = 0): void {
    this.roomRevision++;
    this.busy = true;
    this.decisionError = '';
    this.api
      .decideConversation(this.chatId, decision, teamName, invitationID)
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          this.room = res.data;
          this.busy = false;
          this.panel = null;
          if (decision === 'decline')
            void this.router.navigate(['/team/chats'], {
              queryParams: { course: this.room.course_id },
              replaceUrl: true,
            });
        },
        error: (err) => {
          this.busy = false;
          this.decisionError = academicError(err);
          this.refresh();
        },
      });
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
    this.chat.disconnectWs();
  }
}
