import { academicError } from '../../core/services/academic.service';
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ProfileService, Profile, Interest } from '../../core/services/profile.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { DeleteAccountComponent } from '../../shared/delete-account/delete-account.component';
import { AuthService } from '../../core/services/auth.service';
import { optionName } from '../../shared/text';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, FormsModule, IconComponent, SheetComponent, DeleteAccountComponent],
  templateUrl: './profile.component.html',
  styleUrl: './profile.component.css',
})
export class ProfileComponent implements OnInit {
  profile: Profile = { location: { city: '' } };
  interests: Interest[] = [];
  loading = true;
  loadError = '';
  saving = false;
  saveError = '';
  saveSuccess = false;
  editing: 'basics' | 'about' | 'preferences' | 'interests' | null = null;
  titles = {
    basics: 'Основное',
    about: 'О себе',
    preferences: 'Предпочтения',
    interests: 'Интересы',
  };
  private snapshot: Profile | null = null;

  genders = ['man', 'woman', 'non_binary', 'other'];
  interestedInOptions = ['man', 'woman', 'non_binary', 'everyone'];
  lookingForOptions = [
    'long_term_partner',
    'long_term_open_to_short',
    'short_term_open_to_long_term',
    'short_term',
    'new_friends',
    'still_figuring_it_out',
  ];

  constructor(
    private profileService: ProfileService,
    public auth: AuthService,
  ) {}

  edit(section: 'basics' | 'about' | 'preferences' | 'interests'): void {
    this.snapshot = JSON.parse(JSON.stringify(this.profile));
    this.editing = section;
    this.saveError = '';
    this.saveSuccess = false;
  }

  cancel(): void {
    if (this.snapshot) this.profile = this.snapshot;
    this.editing = null;
    this.saveError = '';
  }

  get interestSummary(): string {
    return (
      this.interests
        .filter((i) => this.profile.interest_ids?.includes(i.id))
        .map((i) => i.name)
        .slice(0, 3)
        .join(', ') || 'Выбрать'
    );
  }

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.loadError = '';
    this.profileService.getMe().subscribe({
      next: (res) => {
        if (res.data) {
          this.profile = res.data;
          if (!this.profile.location) this.profile.location = { city: '' };
        }
        this.loading = false;
      },
      error: (error) => {
        this.loading = false;
        this.loadError = academicError(error);
      },
    });

    this.profileService.getInterests().subscribe({
      next: (res) => (this.interests = res.data ?? []),
      error: () => {},
    });
  }

  formatOption(opt: string): string {
    return optionName(opt);
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
      this.profile.interested_in = current.filter((v) => v !== opt);
    } else {
      this.profile.interested_in = [...current, opt];
    }
  }

  toggleInterest(id: number): void {
    const ids = this.profile.interest_ids ?? [];
    if (ids.includes(id)) {
      this.profile.interest_ids = ids.filter((i) => i !== id);
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
          this.saveError = 'Не удалось определить город. Введите его вручную.';
        }
      },
      error: () => {
        this.saveError = 'Не удалось определить город. Введите его вручную.';
      },
    });
  }

  save(): void {
    this.saving = true;
    this.saveError = '';
    this.saveSuccess = false;

    const payload = this.buildPayload();

    this.profileService.store(payload).subscribe({
      next: (res) => {
        this.profile = res.data;
        if (!this.profile.location) this.profile.location = { city: '' };
        this.saveSuccess = true;
        this.editing = null;
        this.saving = false;
        setTimeout(() => (this.saveSuccess = false), 3000);
      },
      error: (err) => {
        this.saveError = academicError(err, 'Не удалось сохранить профиль. Попробуйте ещё раз.');
        this.saving = false;
      },
    });
  }

  private buildPayload(): Partial<Profile> {
    const p = { ...this.profile };
    // Strip empty strings — backend validates optional pointer fields and rejects ""
    if (!p.birth_date) delete p.birth_date;
    if (!p.gender) delete p.gender;
    if (!p.looking_for) delete p.looking_for;
    // Remove read-only/computed fields the backend doesn't need
    delete (p as any).id;
    delete (p as any).user_id;
    delete (p as any).email;
    delete (p as any).interests;
    delete (p as any).created_at;
    delete (p as any).updated_at;
    return p;
  }
}
