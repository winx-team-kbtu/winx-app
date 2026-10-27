package choice

import (
	"context"
	"errors"
	"testing"

	"winx-profile/internal/app/models/models"
)

type fakeRepo struct {
	selected  map[[2]int64]bool
	inCourse  map[[2]int64]bool // candidate membership
	choices   map[[3]int64]string
	upserts   int
	shortlist []models.DeckCandidate
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		selected: map[[2]int64]bool{{1, 10}: true},
		inCourse: map[[2]int64]bool{{2, 10}: true},
		choices:  map[[3]int64]string{},
	}
}

func (f *fakeRepo) HasCourse(_ context.Context, userID, courseID int64) (bool, error) {
	return f.selected[[2]int64{userID, courseID}], nil
}

func (f *fakeRepo) Upsert(_ context.Context, userID, candidateID, courseID int64, choice string) (bool, error) {
	f.upserts++

	if !f.inCourse[[2]int64{candidateID, courseID}] {
		return false, nil
	}

	f.choices[[3]int64{userID, candidateID, courseID}] = choice

	return true, nil
}

func (f *fakeRepo) ListShortlist(context.Context, int64, int64, int64, int64) ([]models.DeckCandidate, error) {
	return f.shortlist, nil
}

func TestChoose_LikeAndPassAreStored(t *testing.T) {
	repo := newFakeRepo()
	svc := NewServiceWithRepository(repo)

	if err := svc.Choose(context.Background(), 1, 2, 10, Like); err != nil {
		t.Fatal(err)
	}

	if repo.choices[[3]int64{1, 2, 10}] != Like {
		t.Fatalf("like not stored: %v", repo.choices)
	}

	if err := svc.Choose(context.Background(), 1, 2, 10, Pass); err != nil {
		t.Fatal(err)
	}

	if len(repo.choices) != 1 || repo.choices[[3]int64{1, 2, 10}] != Pass {
		t.Fatalf("choice must be replaced, not duplicated: %v", repo.choices)
	}
}

func TestChoose_IsIdempotentOnRetry(t *testing.T) {
	repo := newFakeRepo()
	svc := NewServiceWithRepository(repo)

	for i := 0; i < 3; i++ {
		if err := svc.Choose(context.Background(), 1, 2, 10, Like); err != nil {
			t.Fatalf("retry %d: %v", i, err)
		}
	}

	if len(repo.choices) != 1 {
		t.Fatalf("exactly one persisted choice expected: %v", repo.choices)
	}
}

func TestChoose_Rejections(t *testing.T) {
	cases := []struct {
		name                    string
		user, candidate, course int64
		choice                  string
		want                    error
		wantWrites              int
	}{
		{"unknown choice value", 1, 2, 10, "superlike", ErrInvalidChoice, 0},
		{"self choice", 1, 1, 10, Like, ErrInvalidCandidate, 0},
		{"course not selected by user", 1, 2, 99, Like, ErrCourseNotSelected, 0},
		{"candidate not in course", 1, 3, 10, Like, ErrInvalidCandidate, 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := newFakeRepo()

			err := NewServiceWithRepository(repo).Choose(context.Background(), tc.user, tc.candidate, tc.course, tc.choice)
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}

			if len(repo.choices) != 0 {
				t.Fatalf("rejected choice must not persist: %v", repo.choices)
			}

			if repo.upserts != tc.wantWrites {
				t.Fatalf("upsert attempts=%d want %d", repo.upserts, tc.wantWrites)
			}
		})
	}
}

func TestShortlist_RequiresSelectedCourse(t *testing.T) {
	if _, err := NewServiceWithRepository(newFakeRepo()).Shortlist(context.Background(), 1, 99, 0, 0); !errors.Is(err, ErrCourseNotSelected) {
		t.Fatalf("expected ErrCourseNotSelected, got %v", err)
	}
}

func TestShortlist_ReturnsEntries(t *testing.T) {
	repo := newFakeRepo()
	repo.shortlist = []models.DeckCandidate{{UserID: 2}}

	got, err := NewServiceWithRepository(repo).Shortlist(context.Background(), 1, 10, 0, 0)
	if err != nil || len(got) != 1 {
		t.Fatalf("got %+v err=%v", got, err)
	}
}
