import { RouterLink } from '@angular/router';
import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule, DecimalPipe } from '@angular/common';
import {
  RecommendationService,
  RecommendationProfile,
} from '../../core/services/recommendation.service';
import { MatchService } from '../../core/services/match.service';
import { ProfileService } from '../../core/services/profile.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';
import { optionName } from '../../shared/text';
import { SwipeCardDirective, SwipeDirection } from '../../shared/swipe-card/swipe-card.directive';

import { Subject, forkJoin, takeUntil, timer } from 'rxjs';
import { academicError } from '../../core/services/academic.service';

type SwipeState = 'idle' | 'left' | 'right';

@Component({
  selector: 'app-discover',
  standalone: true,
  imports: [CommonModule, DecimalPipe, IconComponent, SwipeCardDirective, RouterLink, SheetComponent],
  templateUrl: './discover.component.html',
  styleUrl: './discover.component.css',
})
export class DiscoverComponent implements OnInit, OnDestroy {
  profiles: RecommendationProfile[] = [];
  currentIndex = 0;
  loading = true;
  error = '';
  needsSetup = false;
  private readonly destroyed = new Subject<void>();
  swipeState: SwipeState = 'idle';
  match: { profile: RecommendationProfile; chatID: number } | null = null;
  myName = '';
  failedPhotos = new Set<number>();
  detailsOpen = false;
  dragX = 0;
  get dragTransform(): string | null {
    return this.swipeState === 'idle' && this.dragX
      ? 'translateX(' + this.dragX + 'px) rotate(' + this.dragX / 22 + 'deg)'
      : null;
  }
  swipe(direction: SwipeDirection): void {
    direction === 'right' ? this.swipeRight() : this.swipeLeft();
  }
  optionName = optionName;

  constructor(
    private recommendations: RecommendationService,
    private matchService: MatchService,
    private profile: ProfileService,
  ) {}

  ngOnInit(): void {
    this.load();
    this.profile.getMe().pipe(takeUntil(this.destroyed)).subscribe({
      next: (result) => (this.myName = result.data?.first_name ?? ''),
      error: () => {},
    });
  }

  load(): void {
    this.loading = true;
    this.error = '';
    this.needsSetup = false;
    this.currentIndex = 0;
    this.recommendations
      .list(20, 0)
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: (res) => {
          this.profiles = res.data ?? [];
          this.loading = false;
        },
        error: (error) => {
          this.profiles = [];
          this.loading = false;
          if (error instanceof HttpErrorResponse && error.status === 422) this.needsSetup = true;
          else this.error = academicError(error);
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
    this.saveSwipe('left');
  }
  swipeRight(): void {
    this.saveSwipe('right');
  }

  dismissMatch(): void {
    this.match = null;
    this.next();
  }

  private saveSwipe(direction: 'left' | 'right'): void {
    const profile = this.current;
    if (!profile?.user_id || this.swipeState !== 'idle' || this.match) return;
    this.error = '';
    this.swipeState = direction;
    this.dragX = 0;
    const request =
      direction === 'left'
        ? this.matchService.swipeLeft(profile.user_id)
        : this.matchService.swipeRight(profile.user_id);
    const delay = window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 0 : 280;
    forkJoin({ response: request, animation: timer(delay) })
      .pipe(takeUntil(this.destroyed))
      .subscribe({
        next: ({ response }) => {
          if (direction === 'right' && response.data) {
            this.match = { profile, chatID: response.data.chat_id };
            this.swipeState = 'idle';
            return;
          }
          this.next();
        },
        error: (error) => {
          this.swipeState = 'idle';
          this.error = academicError(error, 'Не удалось сохранить выбор. Попробуйте ещё раз.');
        },
      });
  }

  ngOnDestroy(): void {
    this.destroyed.next();
    this.destroyed.complete();
  }

  private next(): void {
    this.detailsOpen = false;
    this.currentIndex++;
    this.swipeState = 'idle';
    if (this.currentIndex >= this.profiles.length) {
      this.load();
      this.currentIndex = 0;
    }
  }
}
