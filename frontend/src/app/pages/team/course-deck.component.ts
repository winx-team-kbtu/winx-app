import { Component, ElementRef, OnDestroy, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { Subject, forkJoin, takeUntil, timer } from 'rxjs';
import {
  AcademicService,
  Candidate,
  Course,
  academicError,
} from '../../core/services/academic.service';
import { CoursePickerComponent } from '../../shared/course-picker/course-picker.component';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { SwipeCardDirective, SwipeDirection } from '../../shared/swipe-card/swipe-card.directive';

@Component({
  selector: 'app-course-deck',
  standalone: true,
  imports: [
    CommonModule,
    RouterLink,
    CoursePickerComponent,
    IconComponent,
    SheetComponent,
    SwipeCardDirective,
  ],
  templateUrl: './course-deck.component.html',
  styleUrls: ['./team.css', './course-deck.component.css'],
})
export class CourseDeckComponent implements OnDestroy {
  course: Course | null = null;
  candidates: Candidate[] = [];
  selected: Candidate | null = null;
  loading = false;
  saving = false;
  error = '';
  feedback = '';
  chatFeedback = false;
  dragX = 0;
  leaving: SwipeDirection | null = null;
  private requestId = 0;
  private focusNext = false;
  @ViewChild('swipeCard') set card(element: ElementRef<HTMLElement> | undefined) {
    if (this.focusNext && element) {
      element.nativeElement.focus({ preventScroll: true });
      this.focusNext = false;
    }
  }
  private readonly destroyed = new Subject<void>();
  private feedbackTimer?: ReturnType<typeof setTimeout>;
  constructor(private api: AcademicService) {}
  get current(): Candidate | null {
    return this.candidates[0] ?? null;
  }
  get cardTransform(): string {
    if (this.leaving)
      return (
        'translateX(' +
        (this.leaving === 'right' ? '120%' : '-120%') +
        ') rotate(' +
        (this.leaving === 'right' ? 14 : -14) +
        'deg)'
      );
    return 'translateX(' + this.dragX + 'px) rotate(' + this.dragX / 22 + 'deg)';
  }
  trackCandidate(_: number, candidate: Candidate): number {
    return candidate.user_id;
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
  selectCourse(course: Course | null): void {
    this.course = course;
    this.selected = null;
    this.feedback = '';
    this.dragX = 0;
    this.leaving = null;
    this.load();
  }
  load(): void {
    const request = ++this.requestId;
    this.error = '';
    this.candidates = [];
    if (!this.course) {
      this.loading = false;
      return;
    }
    this.loading = true;
    // Choices are excluded by the API, so each new batch starts at offset zero.
    this.api
      .deck(this.course.id)
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          if (request !== this.requestId) return;
          this.candidates = res.data ?? [];
          this.loading = false;
        },
        error: (error) => {
          if (request !== this.requestId) return;
          this.loading = false;
          this.error = academicError(error);
        },
      });
  }
  swipe(direction: SwipeDirection): void {
    this.choose(direction === 'right' ? 'like' : 'pass');
  }
  choose(choice: 'like' | 'pass'): void {
    const candidate = this.current;
    if (this.saving || this.loading || !this.course || !candidate) return;
    this.saving = true;
    this.selected = null;
    this.error = '';
    this.feedback = '';
    this.focusNext = document.activeElement?.classList.contains('teammate-card') ?? false;
    this.leaving = choice === 'like' ? 'right' : 'left';
    const delay = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 220;
    forkJoin({
      saved: this.api.choose(this.course.id, candidate.user_id, choice),
      animation: timer(delay),
    })
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: () => {
          this.candidates = this.candidates.filter((c) => c.user_id !== candidate.user_id);
          this.saving = false;
          this.leaving = null;
          this.dragX = 0;
          this.chatFeedback = choice === 'like';
          this.feedback = this.chatFeedback ? 'Диалог добавлен в чаты' : 'Пропущено';
          clearTimeout(this.feedbackTimer);
          if (!this.chatFeedback) this.feedbackTimer = setTimeout(() => (this.feedback = ''), 1800);
          if (!this.candidates.length) this.load();
        },
        error: (error) => {
          this.saving = false;
          this.leaving = null;
          this.dragX = 0;
          this.focusNext = false;
          this.error = academicError(error);
        },
      });
  }
  ngOnDestroy(): void {
    clearTimeout(this.feedbackTimer);
    this.requestId++;
    this.destroyed.next();
    this.destroyed.complete();
  }
}
