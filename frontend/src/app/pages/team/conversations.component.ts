import { Component, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { Subscription, interval } from 'rxjs';
import {
  AcademicService,
  AcademicConversation,
  Course,
  academicError,
} from '../../core/services/academic.service';
import { CoursePickerComponent } from '../../shared/course-picker/course-picker.component';
import { IconComponent } from '../../shared/icon/icon.component';

@Component({
  selector: 'app-team-conversations',
  standalone: true,
  imports: [CommonModule, RouterLink, CoursePickerComponent, IconComponent],
  templateUrl: './conversations.component.html',
  styleUrl: './conversations.component.css',
})
export class ConversationsComponent implements OnDestroy {
  course: Course | null = null;
  rooms: AcademicConversation[] = [];
  loading = false;
  error = '';
  offset = 0;
  private request = 0;
  private active?: Subscription;
  private poll = interval(5000).subscribe(() => {
    if (!document.hidden && !this.loading && this.course) this.load(true);
  });
  constructor(private api: AcademicService) {}
  selectCourse(course: Course | null): void {
    this.course = course;
    this.rooms = [];
    this.offset = 0;
    this.load();
  }
  load(quiet = false, offset = this.offset): void {
    const request = ++this.request;
    this.active?.unsubscribe();
    if (!this.course) {
      this.loading = false;
      return;
    }
    this.loading = true;
    if (!quiet) this.error = '';
    this.active = this.api.conversations(this.course.id, offset).subscribe({
      next: (res) => {
        if (request !== this.request) return;
        this.rooms = res.data ?? [];
        this.offset = offset;
        this.loading = false;
        this.error = '';
        if (!this.rooms.length && offset > 0) this.load(false, Math.max(0, offset - 50));
      },
      error: (err) => {
        if (request !== this.request) return;
        this.loading = false;
        if (!quiet) this.error = academicError(err);
      },
    });
  }
  initials(name: string): string {
    return (
      name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map((s) => s[0])
        .join('')
        .toUpperCase() || '?'
    );
  }
  preview(room: AcademicConversation): string {
    if (room.invitation_status === 'pending') {
      return room.inviter_id === room.peer_id ? 'Приглашение в команду' : 'Приглашение отправлено';
    }
    return (
      room.last_message ||
      (room.status === 'joined'
        ? 'Вы в одной команде'
        : room.initiated_by === room.peer_id
          ? 'Хочет обсудить проект'
          : 'Начните общение')
    );
  }
  dateFormat(value: string): string {
    return new Date(value).toDateString() === new Date().toDateString() ? 'HH:mm' : 'dd.MM';
  }
  ngOnDestroy(): void {
    this.request++;
    this.active?.unsubscribe();
    this.poll.unsubscribe();
  }
}
