import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { RouterLink } from '@angular/router';
import { firstValueFrom } from 'rxjs';
import { ApiResponse } from '../../core/services/profile.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';

interface Entity { name: string; label: string; read_only?: boolean }
interface Field { name: string; type: string; key: boolean; required: boolean; generated: boolean; read_only: boolean }
type Row = Record<string, unknown>;
interface RecordPage { rows: Row[]; fields: Field[]; page: number; has_more: boolean }
interface Day { date: string; [key: string]: string | number }
interface Analytics { current: Record<string, number>; previous: Record<string, number>; series: Day[]; from: string; to: string }

@Component({
  selector: 'app-admin',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterLink, IconComponent, SheetComponent],
  templateUrl: './admin.component.html',
  styleUrl: './admin.component.css',
})
export class AdminComponent implements OnInit {
  readonly periods = [7, 30, 90];
  readonly kpis = ['registrations', 'active_users', 'matches', 'memberships'];
  readonly social = ['social_likes', 'matches', 'social_messages'];
  readonly team = ['teams', 'memberships', 'team_chats', 'team_messages'];
  readonly labels: Record<string, string> = {
    registrations:'Новые аккаунты', active_users:'Активные участники', matches:'Мэтчи',
    memberships:'Участники команд', social_likes:'Лайки', social_messages:'Сообщения в знакомствах',
    team_messages:'Сообщения в командах', team_chats:'Диалоги по предметам', teams:'Команды',
    name:'Название', first_name:'Имя', email:'Почта', id:'ID', role_id:'ID роли', user_id:'ID пользователя',
    course_id:'ID предмета', team_id:'ID команды', created_at:'Создано', updated_at:'Изменено',
    title:'Название', code:'Код', status:'Статус', content:'Текст', action:'Действие',
    actor_id:'ID администратора', entity:'Сущность', record_key:'Ключ записи',
  };
  readonly groups = {
    conversations: ['social_messages','team_messages'],
    connections: ['matches','team_chats'],
    growth: ['registrations','memberships'],
  };
  nav: 'overview' | 'data' | 'audit' = 'overview';
  graph: keyof typeof this.groups = 'conversations';
  days = 30;
  entities: Entity[] = [];
  selected = 'users';
  analytics?: Analytics;
  data?: RecordPage;
  query = '';
  appliedQuery = '';
  page = 1;
  loading = true;
  accessDenied = false;
  error = '';
  notice = '';
  panel: 'create' | 'edit' | 'delete' | null = null;
  row?: Row;
  values: Record<string, string> = {};
  password = '';
  saving = false;
  panelError = '';
  private revision = 0;

  constructor(private http: HttpClient) {}

  ngOnInit(): void { void this.initialize(); }
  async initialize(): Promise<void> {
    this.loading = true;
    this.error = '';
    try {
      this.entities = (await this.get<Entity[]>('catalog')).data;
      await this.load();
    } catch (error) {
      this.accessDenied = error instanceof HttpErrorResponse && error.status === 403;
      this.error = this.message(error);
    } finally { this.loading = false; }
  }
  get entity(): Entity | undefined {
    return this.entities.find(e => e.name === this.selected);
  }
  get editableEntities(): Entity[] {
    return this.entities.filter(entity => !entity.read_only);
  }
  get columns(): Field[] {
    return (this.data?.fields ?? []).slice(0, 7);
  }
  get chartKeys(): string[] { return this.groups[this.graph]; }
  get chartMax(): number {
    return Math.max(1, ...(this.analytics?.series ?? []).flatMap(day => this.chartKeys.map(key => Number(day[key]) || 0)));
  }
  get formFields(): Field[] {
    return (this.data?.fields ?? []).filter(f =>
      this.entity?.read_only || this.panel === 'edit' || !f.generated && !f.read_only,
    );
  }
  choose(nav: typeof this.nav, selected = 'users'): void {
    this.nav = nav;
    this.selectEntity(nav === 'audit' ? 'admin_audit_logs' : selected);
  }
  selectEntity(name: string): void {
    this.selected = name;
    this.query = this.appliedQuery = '';
    this.page = 1;
    this.data = undefined;
    this.notice = '';
    void this.load();
  }
  search(): void {
    this.appliedQuery = this.query.trim();
    this.page = 1;
    void this.load();
  }
  turn(page: number): void { this.page = page; void this.load(); }
  setPeriod(days: number): void { this.days = days; void this.load(); }
  async load(): Promise<void> {
    const current = ++this.revision;
    this.loading = true;
    this.error = '';
    try {
      if (this.nav === 'overview') {
        const result = await this.get<Analytics>('analytics', { days: this.days });
        if (current === this.revision) this.analytics = result.data;
      } else {
        const result = await this.get<RecordPage>('records', { entity: this.selected, page: this.page, q: this.appliedQuery });
        if (current === this.revision) this.data = result.data;
      }
    } catch (error) {
      if (current === this.revision) {
        this.error = this.message(error);
        if (error instanceof HttpErrorResponse && error.status === 403) {
          this.accessDenied = true;
          this.entities = [];
          this.analytics = this.data = undefined;
        }
      }
    } finally {
      if (current === this.revision) this.loading = false;
    }
  }
  open(panel: 'create' | 'edit' | 'delete', row?: Row): void {
    this.panel = panel;
    this.panelError = '';
    this.row = row;
    this.password = '';
    this.values = {};
    for (const field of this.formFields) {
      const value = row?.[field.name];
      this.values[field.name] = value == null ? '' : typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value);
    }
  }
  close(): void { if (!this.saving) this.panel = null; }
  disabled(field: Field): boolean {
    return !!this.entity?.read_only || field.read_only || field.generated || (field.key && this.panel === 'edit');
  }
  jsonField(field: Field): boolean {
    return ['json', 'jsonb', 'ARRAY'].includes(field.type);
  }
  label(field: string): string { return this.labels[field] ?? field.replace(/_/g, ' '); }
  formatDay(value: string): string {
    return value.split('-').reverse().join('.');
  }
  cell(value: unknown): string {
    if (value == null) return '—';
    if (typeof value === 'boolean') return value ? 'Да' : 'Нет';
    return typeof value === 'object' ? JSON.stringify(value) : String(value);
  }
  trend(metric: string): string {
    const before = this.analytics?.previous[metric] ?? 0;
    const current = this.analytics?.current[metric] ?? 0;
    const difference = current - before;
    return (difference > 0 ? '+' : '') + difference.toLocaleString('ru-RU') + ' к прошлому периоду';
  }
  async save(): Promise<void> {
    const panel = this.panel;
    if (!panel || this.saving || this.entity?.read_only) return;
    this.panelError = '';
    const key: Row = {};
    const values: Row = {};
    for (const field of this.data?.fields ?? []) {
      const existing = this.row?.[field.name];
      if (field.key && existing != null) key[field.name] = existing;
      if (this.disabled(field) || panel === 'delete') continue;
      const raw = this.values[field.name] ?? '';
      if (panel === 'edit' && raw === (existing == null ? '' : typeof existing === 'object' ? JSON.stringify(existing,null,2) : String(existing))) continue;
      if (panel === 'create' && raw === '' && !field.required) continue;
      try {
        if (raw === '' && field.required) throw Error();
        if (raw === '') values[field.name] = null;
        else if (this.jsonField(field) || field.type === 'boolean') values[field.name] = JSON.parse(raw);
        else if (['smallint','integer','bigint','real','double precision','numeric'].includes(field.type)) {
          const number = Number(raw);
          if (!Number.isFinite(number)) throw Error();
          values[field.name] = number;
        } else values[field.name] = raw;
      } catch {
        this.panelError = 'Проверьте поле «' + this.label(field.name) + '».';
        return;
      }
    }
    if (panel === 'edit' && !this.password && !Object.keys(values).length) {
      this.close();
      return;
    }
    this.saving = true;
    try {
      const result = await firstValueFrom(this.http.request<ApiResponse<Row>>(
        panel === 'create' ? 'POST' : panel === 'edit' ? 'PATCH' : 'DELETE',
        '/api/v1/admin/records',
        { params: { entity: this.selected }, body: { key, values, password: this.password } },
      ));
      this.panel = null;
      this.notice = result.message !== 'OK' && result.message !== 'ok' ? result.message : panel === 'delete' ? 'Запись удалена' : 'Запись сохранена';
      await this.load();
      if (panel === 'delete' && this.data?.rows.length === 0 && this.page > 1) {
        this.page--;
        await this.load();
      }
    } catch (error) { this.panelError = this.message(error); }
    finally { this.saving = false; this.password = ''; }
  }
  private get<T>(path: string, params: Record<string,string|number> = {}) {
    return firstValueFrom(this.http.get<ApiResponse<T>>('/api/v1/admin/' + path, { params }));
  }
  private message(error: unknown): string {
    if (error instanceof HttpErrorResponse) {
      if (error.status === 403) return 'Для доступа нужна роль администратора.';
      if (typeof error.error?.message === 'string' && error.status < 500) return error.error.message;
    }
    return 'Не удалось выполнить запрос. Попробуйте ещё раз.';
  }
}
