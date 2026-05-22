import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { NotificationService, Notification } from '../../core/services/notification.service';

@Component({
  selector: 'app-notifications',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './notifications.component.html',
  styleUrl: './notifications.component.css',
})
export class NotificationsComponent implements OnInit {
  notifications: Notification[] = [];
  loading = true;

  constructor(private notifService: NotificationService) {}

  ngOnInit(): void {
    this.notifService.list().subscribe({
      next: (res) => {
        this.notifications = res.data ?? [];
        this.loading = false;
      },
      error: () => (this.loading = false),
    });
  }

  delete(notif: Notification): void {
    this.notifService.delete(notif.id).subscribe({
      next: () => {
        this.notifications = this.notifications.filter(n => n.id !== notif.id);
      },
    });
  }

  getIcon(type: string): string {
    const icons: Record<string, string> = {
      match: '💕',
      message: '💬',
      like: '❤️',
      default: '🔔',
    };
    return icons[type] ?? icons['default'];
  }
}
