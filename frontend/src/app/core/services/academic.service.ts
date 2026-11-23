import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse } from '@angular/common/http';
import { ApiResponse } from './profile.service';

export interface Course {
  id: number;
  code: string;
  title: string;
}
export interface AcademicProfile {
  skills: string[];
  preferred_role: string;
}
export interface Candidate extends AcademicProfile {
  user_id: number;
  name: string;
}
export type ConversationDecision = 'invite' | 'accept' | 'reject' | 'decline';

export interface AcademicConversation {
  id: number;
  course_id: number;
  course_code: string;
  course_title: string;
  peer_id: number;
  peer_name: string;
  skills: string[];
  preferred_role: string;
  status: 'open' | 'declined' | 'joined';
  initiated_by: number | null;
  closed_by: number | null;
  team_id: number | null;
  my_team_id: number | null;
  my_team_name: string;
  my_team_size: number;
  peer_team_id: number | null;
  last_message: string;
  last_message_at: string | null;
  created_at: string;
  can_write: boolean;
  invitation_id: number | null;
  inviter_id: number | null;
  invitation_status: 'pending' | 'accepted' | 'declined' | null;
  invitation_team_name: string | null;
}
export interface Team {
  created_by?: number;
  id: number;
  course_id: number;
  name: string;
  join_code: string;
  member_count: number;
  status: 'forming' | 'ready' | 'full';
}
export interface Roster {
  team_id: number;
  team_name: string;
  size: number;
  status: Team['status'];
  course: Course;
  members: Candidate[];
}
export interface TeamPage extends Team {
  viewer_id: number;
  roster: Roster;
}

// Academic pages handle their own errors; existing Social flows are unchanged.
export function academicError(
  error: unknown,
  fallback = 'Не удалось загрузить данные. Попробуйте ещё раз.',
): string {
  if (!(error instanceof HttpErrorResponse)) return fallback;
  if (error.status === 0) return 'Нет связи с сервером. Проверьте подключение.';
  if (error.status === 401) return 'Сессия завершена. Войдите снова.';
  if (error.status >= 500) return 'Сервер временно недоступен. Попробуйте ещё раз.';
  const message = error.error?.message;
  if (typeof message === 'string' && message.trim()) return message;
  if (message && typeof message === 'object') {
    return (
      Object.values(message)
        .filter((value) => typeof value === 'string')
        .join(' · ') || fallback
    );
  }
  return fallback;
}

@Injectable({ providedIn: 'root' })
export class AcademicService {
  private readonly base = '/api/v1/profile/academic';
  constructor(private http: HttpClient) {}
  courses() {
    return this.http.get<ApiResponse<Course[]>>(`${this.base}/courses`);
  }
  myCourses() {
    return this.http.get<ApiResponse<Course[]>>(`${this.base}/courses/me`);
  }
  saveCourses(course_ids: number[]) {
    return this.http.put<ApiResponse<Course[]>>(`${this.base}/courses/me`, {
      course_ids,
    });
  }
  profile() {
    return this.http.get<ApiResponse<AcademicProfile>>(`${this.base}/me`);
  }
  saveProfile(profile: AcademicProfile) {
    return this.http.put<ApiResponse<AcademicProfile>>(`${this.base}/me`, profile);
  }
  deck(course_id: number, offset = 0) {
    return this.http.get<ApiResponse<Candidate[]>>(`${this.base}/deck`, {
      params: { course_id, limit: 20, offset },
    });
  }
  choose(course_id: number, candidate_id: number, choice: 'like' | 'pass') {
    return this.http.put<ApiResponse<null>>(`${this.base}/choices`, {
      course_id,
      candidate_id,
      choice,
    });
  }
  shortlist(course_id: number, offset = 0) {
    return this.http.get<ApiResponse<Candidate[]>>(`${this.base}/shortlist`, {
      params: { course_id, limit: 50, offset },
    });
  }
  myTeam(course_id: number) {
    return this.http.get<ApiResponse<Team>>(`${this.base}/teams/me`, {
      params: { course_id },
    });
  }
  conversations(course_id: number, offset = 0) {
    return this.http.get<ApiResponse<AcademicConversation[]>>(`${this.base}/conversations`, {
      params: { course_id, limit: 50, offset },
    });
  }
  openTeammateChat(course_id: number, peer_id: number) {
    return this.http.post<ApiResponse<{ id: number }>>(`${this.base}/conversations/teammate`, {
      course_id,
      peer_id,
    });
  }
  teamPage(course_id: number) {
    return this.http.get<ApiResponse<TeamPage>>(`${this.base}/teams/me`, {
      params: { course_id, include_roster: true },
    });
  }
  conversation(chat_id: number) {
    return this.http.get<ApiResponse<AcademicConversation>>(`${this.base}/conversations/detail`, {
      params: { chat_id },
    });
  }
  conversationMessages(chat_id: number, before_id = 0) {
    return this.http.get<ApiResponse<import('./chat.service').Message[]>>(
      `${this.base}/conversations/messages`,
      { params: { chat_id, before_id } },
    );
  }
  decideConversation(
    chat_id: number,
    decision: ConversationDecision,
    team_name: string,
    invitation_id: number,
  ) {
    return this.http.post<ApiResponse<AcademicConversation>>(
      `${this.base}/conversations/decision`,
      { chat_id, decision, team_name, invitation_id },
    );
  }
  roster(course_id: number) {
    return this.http.get<ApiResponse<Roster>>(`${this.base}/teams/roster`, {
      params: { course_id },
    });
  }
  createTeam(course_id: number, name: string) {
    return this.http.post<ApiResponse<Team>>(`${this.base}/teams`, {
      course_id,
      name,
    });
  }
  joinTeam(join_code: string) {
    return this.http.post<ApiResponse<Team>>(`${this.base}/teams/join`, {
      join_code,
    });
  }
  removeMember(course_id: number, member_id: number, team_id: number) {
    return this.http.post(`${this.base}/teams/member`, {
      course_id,
      member_id,
      team_id,
    });
  }
}
