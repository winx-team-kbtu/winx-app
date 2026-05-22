import { Component, OnInit } from '@angular/core';
import { CommonModule, DecimalPipe } from '@angular/common';
import { RecommendationService, RecommendationProfile } from '../../core/services/recommendation.service';
import { MatchService } from '../../core/services/match.service';

type SwipeState = 'idle' | 'left' | 'right';

@Component({
  selector: 'app-discover',
  standalone: true,
  imports: [CommonModule, DecimalPipe],
  templateUrl: './discover.component.html',
  styleUrl: './discover.component.css',
})
export class DiscoverComponent implements OnInit {
  profiles: RecommendationProfile[] = [];
  currentIndex = 0;
  loading = true;
  swipeState: SwipeState = 'idle';
  matchMessage = '';

  constructor(
    private recommendations: RecommendationService,
    private matchService: MatchService
  ) {}

  ngOnInit(): void {
    this.load();
  }

  private load(): void {
    this.loading = true;
    this.recommendations.list(20, 0).subscribe({
      next: (res) => {
        this.profiles = res.data ?? [];
        this.loading = false;
      },
      error: () => {
        this.profiles = [];
        this.loading = false;
      },
    });
  }

  get current(): RecommendationProfile | null {
    return this.profiles[this.currentIndex] ?? null;
  }

  get hasMore(): boolean {
    return this.currentIndex < this.profiles.length;
  }

  swipeLeft(): void {
    const profile = this.current;
    if (!profile?.user_id || this.swipeState !== 'idle') return;
    this.swipeState = 'left';
    this.matchService.swipeLeft(profile.user_id).subscribe({ error: () => {} });
    setTimeout(() => this.next(), 420);
  }

  swipeRight(): void {
    const profile = this.current;
    if (!profile?.user_id || this.swipeState !== 'idle') return;
    this.swipeState = 'right';
    const name = profile.name || 'them';
    this.matchService.swipeRight(profile.user_id).subscribe({
      next: (res) => {
        if (res.message?.toLowerCase().includes('match')) {
          this.matchMessage = `It's a match with ${name}! 💕`;
          setTimeout(() => (this.matchMessage = ''), 3000);
        }
      },
      error: () => {},
    });
    setTimeout(() => this.next(), 420);
  }

  private next(): void {
    this.currentIndex++;
    this.swipeState = 'idle';
    if (this.currentIndex >= this.profiles.length) {
      this.load();
      this.currentIndex = 0;
    }
  }

  getPhotoUrl(profile: RecommendationProfile | null): string {
    return profile?.photo_url || '';
  }
}
