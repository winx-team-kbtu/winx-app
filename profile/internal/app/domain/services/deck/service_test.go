package deck

import (
	"context"
	"errors"
	"testing"

	"winx-profile/internal/app/models/models"
)

type fakeRepo struct {
	selected   map[[2]int64]bool
	candidates []models.DeckCandidate
	gotLimit   int64
	gotOffset  int64
	listCalls  int
}

func (f *fakeRepo) HasCourse(_ context.Context, userID, courseID int64) (bool, error) {
	return f.selected[[2]int64{userID, courseID}], nil
}

func (f *fakeRepo) ListCandidates(_ context.Context, _, _, limit, offset int64) ([]models.DeckCandidate, error) {
	f.listCalls++
	f.gotLimit, f.gotOffset = limit, offset

	return f.candidates, nil
}

func TestCandidates_RejectsUnselectedCourseWithoutQuerying(t *testing.T) {
	repo := &fakeRepo{selected: map[[2]int64]bool{{1, 10}: true}}

	_, err := NewServiceWithRepository(repo).Candidates(context.Background(), 1, 99, 0, 0)
	if !errors.Is(err, ErrCourseNotSelected) {
		t.Fatalf("expected ErrCourseNotSelected, got %v", err)
	}

	if repo.listCalls != 0 {
		t.Fatal("candidates must not be queried for an unselected course")
	}
}

func TestCandidates_ReturnsCandidatesForSelectedCourse(t *testing.T) {
	repo := &fakeRepo{
		selected:   map[[2]int64]bool{{1, 10}: true},
		candidates: []models.DeckCandidate{{UserID: 2, Name: "Ann"}},
	}

	got, err := NewServiceWithRepository(repo).Candidates(context.Background(), 1, 10, 0, 0)
	if err != nil || len(got) != 1 || got[0].UserID != 2 {
		t.Fatalf("unexpected result %+v err=%v", got, err)
	}
}

func TestCandidates_ClampsPaging(t *testing.T) {
	cases := []struct {
		name                  string
		limit, offset         int64
		wantLimit, wantOffset int64
	}{
		{"defaults", 0, 0, DefaultLimit, 0},
		{"too big", 5000, 3, MaxLimit, 3},
		{"negative offset", 10, -4, 10, 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepo{selected: map[[2]int64]bool{{1, 10}: true}}

			if _, err := NewServiceWithRepository(repo).Candidates(context.Background(), 1, 10, tc.limit, tc.offset); err != nil {
				t.Fatal(err)
			}

			if repo.gotLimit != tc.wantLimit || repo.gotOffset != tc.wantOffset {
				t.Fatalf("got limit=%d offset=%d", repo.gotLimit, repo.gotOffset)
			}
		})
	}
}
