import { academicError } from '../../core/services/academic.service';
import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { MatchService, Match } from '../../core/services/match.service';
import { IconComponent } from '../../shared/icon/icon.component';
import { SheetComponent } from '../../shared/sheet/sheet.component';

@Component({
  selector: 'app-matches',
  standalone: true,
  imports: [CommonModule, RouterLink, IconComponent, SheetComponent],
  templateUrl: './matches.component.html',
  styleUrl: './matches.component.css',
})
export class MatchesComponent implements OnInit {
  matches: Match[] = [];
  removing: Match | null = null;
  saving = false;
  removalError = '';
  loading = true;
  error = '';

  constructor(private matchService: MatchService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.error = '';
    this.loading = true;
    this.matchService.list().subscribe({
      next: (response) => {
        this.matches = response.data ?? [];
        this.loading = false;
      },
      error: (error) => {
        this.loading = false;
        this.error = academicError(error);
      },
    });
  }

  confirmRemoval(match: Match): void {
    this.removalError = '';
    this.removing = match;
  }

  unmatch(): void {
    const match = this.removing;
    if (!match || this.saving) return;
    this.saving = true;
    this.removalError = '';
    this.matchService.delete(match.id).subscribe({
      next: () => {
        this.saving = false;
        this.removing = null;
        this.matches = this.matches.filter((m) => m.id !== match.id);
      },
      error: (error) => {
        this.saving = false;
        this.removalError = academicError(error, 'Не удалось убрать совпадение.');
      },
    });
  }
}
