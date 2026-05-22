import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ProfileService, Profile, Interest } from '../../core/services/profile.service';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css',
})
export class ProfileComponent implements OnInit {
  profile: Profile = { location: { city: '' } };
  interests: Interest[] = [];
  loading = true;
  saving = false;
  saveError = '';
  saveSuccess = false;
  photoUploading = false;
  photoUrl = '';

  genders = ['man', 'woman', 'non_binary', 'other'];
  interestedInOptions = ['man', 'woman', 'non_binary', 'everyone'];
  lookingForOptions = [
    'long_term_partner', 'long_term_open_to_short',
    'short_term_open_to_long_term', 'short_term', 'new_friends', 'still_figuring_it_out',
  ];

  constructor(private profileService: ProfileService) {}

  ngOnInit(): void {
    this.profileService.getMe().subscribe({
      next: (res) => {
        if (res.data) {
          this.profile = res.data;
          if (!this.profile.location) this.profile.location = { city: '' };
        }
        this.loading = false;
      },
      error: () => (this.loading = false),
    });

    this.profileService.getInterests().subscribe({
      next: (res) => (this.interests = res.data ?? []),
    });

    this.profileService.getPhotos().subscribe({
      next: (res) => {
        const primary = (res.data ?? []).find(p => p.is_primary);
        if (primary) this.photoUrl = primary.url;
      },
    });
  }

  formatOption(opt: string): string {
    return opt.split('_').map(w => w.charAt(0).toUpperCase() + w.slice(1)).join(' ');
  }

  isInterestSelected(id: number): boolean {
    return this.profile.interest_ids?.includes(id) ?? false;
  }

  isInterestedInSelected(opt: string): boolean {
    return this.profile.interested_in?.includes(opt) ?? false;
  }

  toggleInterestedIn(opt: string): void {
    const current = this.profile.interested_in ?? [];
    if (current.includes(opt)) {
      this.profile.interested_in = current.filter(v => v !== opt);
    } else {
      this.profile.interested_in = [...current, opt];
    }
  }

  toggleInterest(id: number): void {
    const ids = this.profile.interest_ids ?? [];
    if (ids.includes(id)) {
      this.profile.interest_ids = ids.filter(i => i !== id);
    } else if (ids.length < 20) {
      this.profile.interest_ids = [...ids, id];
    }
  }

  detectLocation(): void {
    this.profileService.lookupLocationByIP().subscribe({
      next: (res) => {
        if (res.data) {
          if (!this.profile.location) this.profile.location = { city: '' };
          this.profile.location.city = res.data.city;
          this.profile.location.country = res.data.country;
          this.profile.location.current_location = {
            latitude: res.data.current_location.latitude,
            longitude: res.data.current_location.longitude,
          };
        } else {
          this.setDefaultLocation();
        }
      },
      error: () => this.setDefaultLocation(),
    });
  }

  private setDefaultLocation(): void {
    if (!this.profile.location) this.profile.location = { city: '' };
    this.profile.location.city = 'Almaty';
    this.profile.location.country = 'Kazakhstan';
    this.profile.location.current_location = { latitude: 43.34536794589393, longitude: 76.89939104069116 };
  }

  onPhotoSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    this.photoUploading = true;
    this.profileService.storePhoto(file).subscribe({
      next: (res) => {
        if (res.data?.url) this.photoUrl = res.data.url;
        this.photoUploading = false;
      },
      error: () => (this.photoUploading = false),
    });
  }

  save(): void {
    this.saving = true;
    this.saveError = '';
    this.saveSuccess = false;
    this.profileService.store(this.profile).subscribe({
      next: () => {
        this.saveSuccess = true;
        this.saving = false;
        setTimeout(() => (this.saveSuccess = false), 3000);
      },
      error: (err) => {
        this.saveError = err?.error?.message || 'Failed to save. Try again.';
        this.saving = false;
      },
    });
  }
}
