import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { firstValueFrom, forkJoin } from 'rxjs';
import { AcademicService, Course, academicError } from '../../core/services/academic.service';
import { Profile, ProfileService } from '../../core/services/profile.service';
import { AuthService } from '../../core/services/auth.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { DeleteAccountComponent } from '../../shared/delete-account/delete-account.component';
import { courseName } from '../../shared/text';

type Section = 'basics' | 'courses' | 'skills' | 'role';
@Component({
  selector: 'app-academic-profile',
  standalone: true,
  imports: [CommonModule, FormsModule, IconComponent, SheetComponent, DeleteAccountComponent],
  templateUrl: './academic-profile.component.html',
  styleUrl: './team.css',
})
export class AcademicProfileComponent implements OnInit {
  courses: Course[] = [];
  selectedIds: number[] = [];
  skillsText = '';
  role = '';
  name = 'Мой профиль';
  city = '';
  private userProfile: Profile = {};
  loading = true;
  ready = false;
  saving = false;
  error = '';
  success = '';
  editing: Section | null = null;
  titles = {
    basics: 'Основное',
    courses: 'Предметы',
    skills: 'Навыки',
    role: 'Роль в команде',
  };
  label = courseName;
  private snapshot = {
    name: '',
    city: '',
    ids: [] as number[],
    skills: '',
    role: '',
  };
  constructor(
    private api: AcademicService,
    private social: ProfileService,
    public auth: AuthService,
  ) {}
  ngOnInit(): void {
    this.load();
  }
  load(): void {
    this.loading = true;
    this.error = '';
    this.ready = false;
    forkJoin({
      courses: this.api.courses(),
      selected: this.api.myCourses(),
      profile: this.api.profile(),
      user: this.social.getMe(),
    }).subscribe({
      next: ({ courses, selected, profile, user }) => {
        this.courses = courses.data ?? [];
        this.selectedIds = (selected.data ?? []).map((c) => c.id);
        this.skillsText = (profile.data?.skills ?? []).join(', ');
        this.role = profile.data?.preferred_role ?? '';
        this.userProfile = user.data;
        this.name = user.data.first_name || 'Мой профиль';
        this.city = user.data.location?.city || '';
        this.loading = false;
        this.ready = true;
      },
      error: (err) => {
        this.loading = false;
        this.error = academicError(err);
      },
    });
  }
  get skills(): string[] {
    const seen = new Set<string>();
    return this.skillsText
      .split(',')
      .map((s) => s.trim())
      .filter((s) => {
        if (!s || seen.has(s.toLowerCase())) return false;
        seen.add(s.toLowerCase());
        return true;
      });
  }
  get skillsError(): string {
    if (this.skills.length > 20) return 'Можно добавить до 20 навыков.';
    if (this.skills.some((s) => [...s].length > 30)) return 'Название навыка — до 30 символов.';
    return '';
  }
  get selectedNames(): string {
    return (
      this.courses
        .filter((c) => this.selectedIds.includes(c.id))
        .map((c) => courseName(c.code, c.title))
        .join(', ') || 'Выбрать'
    );
  }
  edit(section: Section): void {
    this.snapshot = {
      name: this.name,
      city: this.city,
      ids: [...this.selectedIds],
      skills: this.skillsText,
      role: this.role,
    };
    this.editing = section;
    this.error = '';
    this.success = '';
  }
  cancel(): void {
    this.name = this.snapshot.name;
    this.city = this.snapshot.city;
    this.selectedIds = [...this.snapshot.ids];
    this.skillsText = this.snapshot.skills;
    this.role = this.snapshot.role;
    this.editing = null;
    this.error = '';
  }
  toggleCourse(id: number): void {
    this.selectedIds = this.selectedIds.includes(id)
      ? this.selectedIds.filter((value) => value !== id)
      : [...this.selectedIds, id];
  }
  async save(): Promise<void> {
    if (
      this.saving ||
      !this.editing ||
      (this.editing === 'basics' && (!this.name.trim() || !this.city.trim())) ||
      (this.editing === 'skills' && this.skillsError) ||
      [...this.role.trim()].length > 50
    )
      return;
    this.saving = true;
    this.error = '';
    try {
      if (this.editing === 'basics') {
        const result = await firstValueFrom(
          this.social.store({
            first_name: this.name.trim(),
            location: { ...this.userProfile.location, city: this.city.trim() },
          }),
        );
        this.userProfile = result.data;
        this.name = result.data.first_name || 'Мой профиль';
        this.city = result.data.location?.city || '';
      } else if (this.editing === 'courses') {
        const result = await firstValueFrom(this.api.saveCourses(this.selectedIds));
        this.selectedIds = result.data.map((c) => c.id);
      } else {
        const result = await firstValueFrom(
          this.api.saveProfile({
            skills: this.skills,
            preferred_role: this.role.trim(),
          }),
        );
        this.skillsText = result.data.skills.join(', ');
        this.role = result.data.preferred_role;
      }
      this.editing = null;
      this.success = 'Сохранено';
    } catch (err) {
      this.error = academicError(err);
    } finally {
      this.saving = false;
    }
  }
}
