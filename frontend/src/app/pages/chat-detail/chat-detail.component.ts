import { Component, OnInit, OnDestroy, ViewChild, ElementRef, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute } from '@angular/router';
import { Subscription } from 'rxjs';
import { ChatService, Message } from '../../core/services/chat.service';
import { AuthService } from '../../core/services/auth.service';
import { ProfileService } from '../../core/services/profile.service';

@Component({
  selector: 'app-chat-detail',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './chat-detail.component.html',
  styleUrl: './chat-detail.component.css',
})
export class ChatDetailComponent implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('messagesEnd') messagesEnd!: ElementRef;

  chatId!: number;
  messages: Message[] = [];
  newMessage = '';
  loading = true;
  myId: number | null = null;
  private wsSub?: Subscription;
  private shouldScroll = false;

  constructor(
    private route: ActivatedRoute,
    private chatService: ChatService,
    private auth: AuthService,
    private profileService: ProfileService,
  ) {}

  isMine(msg: Message): boolean {
    return this.myId !== null && msg.sender_id === this.myId;
  }

  ngOnInit(): void {
    this.chatId = Number(this.route.snapshot.paramMap.get('id'));

    this.profileService.getMe().subscribe({
      next: (res) => {
        this.myId = res.data?.user_id ?? null;
      },
    });

    this.loadMessages();

    const token = this.auth.getToken() ?? '';
    this.wsSub = this.chatService.connectWs(token).subscribe({
      next: (msg) => {
        if (msg.type === 'message' && Number(msg.chat_id) === this.chatId && Number(msg.sender_id) !== this.myId) {
          this.messages.push({
            id: Date.now(),
            chat_id: this.chatId,
            sender_id: msg.sender_id ?? 0,
            content: msg.content ?? '',
            created_at: msg.created_at ?? new Date().toISOString(),
          });
          this.shouldScroll = true;
        }
      },
    });
  }

  ngAfterViewChecked(): void {
    if (this.shouldScroll) {
      this.scrollToBottom();
      this.shouldScroll = false;
    }
  }

  ngOnDestroy(): void {
    this.wsSub?.unsubscribe();
    this.chatService.disconnectWs();
  }

  loadMessages(): void {
    this.chatService.getMessages(this.chatId).subscribe({
      next: (res) => {
        this.messages = res.data ?? [];
        this.loading = false;
        this.shouldScroll = true;
      },
      error: () => (this.loading = false),
    });
  }

  send(): void {
    const text = this.newMessage.trim();
    if (!text) return;
    this.chatService.sendMessage(this.chatId, text);
    this.newMessage = '';
    this.messages.push({
      id: Date.now(),
      chat_id: this.chatId,
      sender_id: this.myId ?? 0,
      content: text,
      created_at: new Date().toISOString(),
    });
    this.shouldScroll = true;
  }

  private scrollToBottom(): void {
    try {
      this.messagesEnd.nativeElement.scrollIntoView({ behavior: 'smooth' });
    } catch {}
  }
}
