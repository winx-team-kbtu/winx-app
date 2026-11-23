import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterLink } from '@angular/router';
import { AcademicService, Course, academicError } from '../../core/services/academic.service';
import { courseName } from '../text';

@Component({
  selector: 'app-course-picker',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink],
  template: ` <div class="course-picker">
    <p *ngIf="loading" class="loading-label" role="status">Загрузка…</p>
    <div *ngIf="error" class="error-msg" role="alert">
      {{ error }} <button class="btn btn-ghost" (click)="load()">Повторить</button>
    </div>
    <ng-container *ngIf="!loading && !error">
      <label *ngIf="courses.length" class="course-label"
        ><span class="sr-only">Предмет</span
        ><select
          class="form-control"
          aria-label="Предмет"
          [ngModel]="selectedId"
          (ngModelChange)="selectId(+$event)"
          [disabled]="disabled"
        >
          <option *ngFor="let course of courses" [ngValue]="course.id">
            {{ label(course.code, course.title) }}
          </option>
        </select></label
      >
      <div *ngIf="!courses.length" class="empty-state">
        <h3>Выберите предметы</h3>
        <p>Чтобы видеть однокурсников.</p>
        <a routerLink="/team/profile" class="btn btn-primary mt-16">Выбрать предметы</a>
      </div>
    </ng-container>
  </div>`,
  styles: [
    `
      :host {
        display: block;
        min-width: 0;
      }
      .course-picker {
        margin-bottom: 20px;
      }
      .course-label {
        display: block;
      }
      select {
        max-width: var(--course-picker-width, 320px);
        width: 100%;
        font-size: 14px;
        font-weight: 500;
        background: #fff;
      }
      .loading-label {
        color: var(--muted);
        font-size: 13px;
        padding: 12px 0;
      }
      @media (max-width: 640px) {
        select {
          max-width: none;
          font-size: 16px;
        }
      }
    `,
  ],
})
export class CoursePickerComponent implements OnInit {
  @Input() disabled = false;
  @Output() courseChange = new EventEmitter<Course | null>();
  courses: Course[] = [];
  selectedId = 0;
  loading = true;
  error = '';
  label = courseName;
  constructor(
    private api: AcademicService,
    private route: ActivatedRoute,
    private router: Router,
  ) {}
  ngOnInit(): void {
    this.load();
  }
  load(): void {
    this.loading = true;
    this.error = '';
    this.api.myCourses().subscribe({
      next: (res) => {
        this.courses = res.data ?? [];
        this.loading = false;
        const preferred = Number(
          this.route.snapshot.queryParamMap.get('course') ||
            localStorage.getItem('winx_team_course'),
        );
        this.selectId(
          this.courses.some((c) => c.id === preferred) ? preferred : (this.courses[0]?.id ?? 0),
        );
      },
      error: (err) => {
        this.loading = false;
        this.error = academicError(err);
        this.courseChange.emit(null);
      },
    });
  }
  selectId(id: number): void {
    this.selectedId = id;
    const course = this.courses.find((c) => c.id === id) ?? null;
    if (course) {
      localStorage.setItem('winx_team_course', String(id));
      if (this.route.snapshot.queryParamMap.get('course') !== String(id)) {
        void this.router.navigate([], {
          relativeTo: this.route,
          queryParams: { course: id },
          queryParamsHandling: 'merge',
          replaceUrl: true,
          preserveFragment: true,
        });
      }
    }
    this.courseChange.emit(course);
  }
}
