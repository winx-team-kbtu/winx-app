import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';

export interface Profile {
  id?: number;
  user_id?: number;
  first_name?: string;
  birth_date?: string;
  gender?: string;
  show_gender?: boolean;
  show_age?: boolean;
  about_me?: string;
  job_title?: string;
  company?: string;
  school?: string;
  interest_ids?: number[];
  interested_in?: string[];
  looking_for?: string;
  lifestyle?: Record<string, string>;
  preferences?: Record<string, unknown>;
  location?: {
    city: string;
    country?: string;
    current_location?: { latitude: number; longitude: number };
  };
  photos?: Photo[];
}

export interface Photo {
  id: number;
  url: string;
  is_primary: boolean;
}

export interface Interest {
  id: number;
  name: string;
}

export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}

@Injectable({ providedIn: 'root' })
export class ProfileService {
  constructor(private http: HttpClient) {}

  getMe() {
    return this.http.get<ApiResponse<Profile>>('/api/v1/profile/me');
  }

  store(profile: Partial<Profile>) {
    return this.http.post<ApiResponse<Profile>>('/api/v1/profile/store', profile);
  }

  getPhotos() {
    return this.http.get<ApiResponse<Photo[]>>('/api/v1/profile/photo');
  }

  storePhoto(file: File) {
    const formData = new FormData();
    formData.append('photo', file);
    return this.http.post<ApiResponse<Photo>>('/api/v1/profile/photo/store', formData);
  }

  getInterests() {
    return this.http.get<ApiResponse<Interest[]>>('/api/v1/profile/interests');
  }

  lookupLocationByIP() {
    return this.http.get<ApiResponse<{ city: string; country: string; current_location: { latitude: number; longitude: number } }>>('/api/v1/profile/location/ip');
  }
}
