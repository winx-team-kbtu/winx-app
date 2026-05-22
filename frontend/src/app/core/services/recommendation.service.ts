import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ApiResponse } from './profile.service';

export interface RecommendationProfile {
  user_id: number;
  name?: string;
  photo_url?: string;
  age?: number;
  city?: string;
  country?: string;
  about_me?: string;
  gender?: string;
  looking_for?: string;
  score: number;
  shared_interests: number;
  has_photo: boolean;
  distance_km?: number;
}

@Injectable({ providedIn: 'root' })
export class RecommendationService {
  constructor(private http: HttpClient) {}

  list(limit = 20, offset = 0) {
    return this.http.get<ApiResponse<RecommendationProfile[]>>('/api/v1/recommendations', {
      params: { limit, offset },
    });
  }
}
