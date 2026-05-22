import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ApiResponse } from './profile.service';

export interface Match {
  id: number;
  matched_user_id: number;
  name: string;
  city?: string;
  country?: string;
  photo_url?: string;
  created_at: string;
}

@Injectable({ providedIn: 'root' })
export class MatchService {
  constructor(private http: HttpClient) {}

  list(limit = 20, offset = 0) {
    return this.http.get<ApiResponse<Match[]>>('/api/v1/matches', {
      params: { limit, offset },
    });
  }

  delete(id: number) {
    return this.http.delete<ApiResponse<null>>(`/api/v1/matches/${id}`);
  }

  swipeLeft(swiped_id: number) {
    return this.http.post<ApiResponse<null>>('/api/v1/swipes/left', { swiped_id });
  }

  swipeRight(swiped_id: number) {
    return this.http.post<ApiResponse<null>>('/api/v1/swipes/right', { swiped_id });
  }
}
