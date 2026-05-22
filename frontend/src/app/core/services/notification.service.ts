import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ApiResponse } from './profile.service';

export interface Notification {
  id: number;
  type: string;
  channel: string;
  recipient: string;
  subject: string;
  body: string;
  status: string;
  created_at: string;
  sent_at?: string;
  error_message?: string;
}

@Injectable({ providedIn: 'root' })
export class NotificationService {
  constructor(private http: HttpClient) {}

  list() {
    return this.http.get<ApiResponse<Notification[]>>('/api/v1/notifications');
  }

  delete(id: number) {
    return this.http.delete<ApiResponse<null>>(`/api/v1/notifications/${id}`);
  }
}
