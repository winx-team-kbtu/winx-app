import { Component, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { firstValueFrom } from 'rxjs';
import { Router, RouterLink } from '@angular/router';
import {
  AcademicService,
  Candidate,
  Course,
  Roster,
  Team,
  academicError,
} from '../../core/services/academic.service';
import { CoursePickerComponent } from '../../shared/course-picker/course-picker.component';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { courseName } from '../../shared/text';

@Component({
  selector: 'app-teams',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, CoursePickerComponent, IconComponent, SheetComponent],
  templateUrl: './teams.component.html',
  styleUrls: ['./team.css', './teams.component.css'],
})
export class TeamsComponent {
  @ViewChild(CoursePickerComponent) picker!: CoursePickerComponent;
  course: Course | null = null;
  team: Team | null = null;
  roster: Roster | null = null;
  member: Candidate | null = null;
  panel: 'create' | 'join' | 'code' | 'remove' | null = null;
  removalID = 0;
  name = '';
  joinCode = '';
  loading = false;
  busy = false;
  error = '';
  copyMessage = '';
  loadFailed = false;
  myId = 0;
  courseName = courseName;
  private requestId = 0;
  constructor(
    private api: AcademicService,
    private router: Router,
  ) {}
  openMember(person: Candidate): void {
    this.error = '';
    this.member = person;
  }
  confirmRemoval(memberId: number): void {
    if (!memberId || this.busy) return;
    this.removalID = memberId;
    this.error = '';
    this.panel = 'remove';
  }
  async remove(): Promise<void> {
    if (!this.course || !this.team || !this.removalID || this.busy) return;
    this.busy = true;
    this.error = '';
    try {
      await firstValueFrom(this.api.removeMember(this.course.id, this.removalID, this.team.id));
      this.panel = null;
      this.member = null;
      await this.load();
    } catch (error) {
      this.error = academicError(error);
    } finally {
      this.busy = false;
    }
  }
  async openChat(): Promise<void> {
    if (!this.course || !this.member || this.busy) return;
    this.busy = true;
    this.error = '';
    try {
      const result = await firstValueFrom(this.api.openTeammateChat(this.course.id, this.member.user_id));
      await this.router.navigate(['/team/chats', result.data.id]);
    } catch (err) {
      this.error = academicError(err);
    } finally {
      this.busy = false;
    }
  }
  selectCourse(course: Course | null): void {
    this.course = course;
    this.panel = null;
    this.member = null;
    this.name = '';
    this.copyMessage = '';
    void this.load();
  }
  open(panel: 'create' | 'join' | 'code'): void {
    this.error = '';
    this.copyMessage = '';
    this.panel = panel;
  }
  close(): void {
    if (!this.busy) {
      if (this.panel !== 'remove') this.member = null;
      this.panel = null;
      this.error = '';
    }
  }
  async load(): Promise<void> {
    const request = ++this.requestId;
    this.team = null;
    this.roster = null;
    this.error = '';
    this.loadFailed = false;
    if (!this.course) {
      this.loading = false;
      return;
    }
    const courseId = this.course.id;
    this.loading = true;
    try {
      const result = await firstValueFrom(this.api.teamPage(courseId));
      if (request !== this.requestId) return;
      this.team = result.data;
      this.roster = result.data.roster;
      this.myId = result.data.viewer_id;
    } catch (err) {
      if (request !== this.requestId) return;
      if (!(err instanceof HttpErrorResponse && err.status === 404 && !this.team)) {
        this.error = academicError(err);
        this.loadFailed = true;
      }
    } finally {
      if (request === this.requestId) this.loading = false;
    }
  }
  async create(): Promise<void> {
    if (!this.course || this.busy || this.name.trim().length < 2 || this.name.trim().length > 50)
      return;
    this.busy = true;
    this.error = '';
    try {
      await firstValueFrom(this.api.createTeam(this.course.id, this.name.trim()));
      this.panel = null;
      await this.load();
    } catch (err) {
      this.error = academicError(err);
    } finally {
      this.busy = false;
    }
  }
  async join(): Promise<void> {
    if (this.busy || !this.joinCode.trim()) return;
    this.busy = true;
    this.error = '';
    try {
      const result = await firstValueFrom(this.api.joinTeam(this.joinCode.trim().toUpperCase()));
      this.joinCode = '';
      this.panel = null;
      if (this.course?.id === result.data.course_id) await this.load();
      else this.picker.selectId(result.data.course_id);
    } catch (err) {
      this.error = academicError(err);
    } finally {
      this.busy = false;
    }
  }
  async copyCode(): Promise<void> {
    if (!this.team) return;
    try {
      await navigator.clipboard.writeText(this.team.join_code);
      this.copyMessage = 'Код скопирован';
    } catch {
      this.copyMessage = 'Выделите код и скопируйте вручную.';
    }
  }
  initials(name: string): string {
    return (
      name
        .trim()
        .split(/\s+/)
        .slice(0, 2)
        .map((s) => s[0])
      .join('').toUpperCase() || '?'
    );
  }
  statusLabel(status: string): string {
    return status === 'full'
      ? 'Команда собрана'
      : status === 'ready'
        ? 'Готовы к работе'
        : 'Набираем людей';
  }
}
