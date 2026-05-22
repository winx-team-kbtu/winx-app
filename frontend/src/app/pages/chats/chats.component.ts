import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { forkJoin } from 'rxjs';
import { ChatService, Chat } from '../../core/services/chat.service';
import { MatchService } from '../../core/services/match.service';

@Component({
  selector: 'app-chats',
  standalone: true,
  imports: [CommonModule, RouterLink],
  templateUrl: './chats.component.html',
  styleUrl: './chats.component.css',
})
export class ChatsComponent implements OnInit {
  chats: Chat[] = [];
  matchNameMap: Record<number, string> = {};
  loading = true;

  constructor(
    private chatService: ChatService,
    private matchService: MatchService,
  ) {}

  ngOnInit(): void {
    forkJoin({
      chats: this.chatService.listChats(),
      matches: this.matchService.list(),
    }).subscribe({
      next: ({ chats, matches }) => {
        this.chats = chats.data ?? [];
        this.matchNameMap = {};
        for (const m of matches.data ?? []) {
          this.matchNameMap[m.id] = m.name;
        }
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  getChatName(chat: Chat): string {
    return this.matchNameMap[chat.match_id] || 'Chat';
  }

  getInitial(chat: Chat): string {
    return (this.getChatName(chat)[0] || '?').toUpperCase();
  }
}
