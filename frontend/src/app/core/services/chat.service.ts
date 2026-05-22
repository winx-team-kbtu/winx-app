import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, Subject } from 'rxjs';
import { ApiResponse } from './profile.service';

export interface Chat {
  id: number;
  match_id: number;
  user_one_id: number;
  user_two_id: number;
  created_at: string;
}

export interface Message {
  id: number;
  chat_id: number;
  sender_id: number;
  content: string;
  created_at: string;
}

export interface WsMessage {
  type: string;
  chat_id?: number;
  content?: string;
  sender_id?: number;
  created_at?: string;
}

@Injectable({ providedIn: 'root' })
export class ChatService {
  private ws: WebSocket | null = null;
  private messagesSubject = new Subject<WsMessage>();

  constructor(private http: HttpClient) {}

  listChats(limit = 20, offset = 0) {
    return this.http.get<ApiResponse<Chat[]>>('/api/v1/chats', {
      params: { limit, offset },
    });
  }

  getMessages(chatId: number, limit = 50, offset = 0) {
    return this.http.get<ApiResponse<Message[]>>(`/api/v1/chats/${chatId}/messages`, {
      params: { limit, offset },
    });
  }

  connectWs(token: string): Observable<WsMessage> {
    // Create a fresh Subject for each connection
    this.messagesSubject = new Subject<WsMessage>();
    const wsUrl = `ws://${window.location.hostname}:3006/ws?token=${token}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data);
        this.messagesSubject.next(msg);
      } catch {}
    };

    this.ws.onclose = () => this.messagesSubject.complete();
    this.ws.onerror = () => this.messagesSubject.complete();

    return this.messagesSubject.asObservable();
  }

  sendMessage(chatId: number, content: string): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ type: 'message', chat_id: chatId, content }));
    }
  }

  disconnectWs(): void {
    this.ws?.close();
    this.ws = null;
  }
}
