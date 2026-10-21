package course

import (
	"context"
	"errors"
	"sort"
	"testing"

	"winx-profile/internal/app/models/models"
)

type fakeRepo struct {
	catalogue map[int64]models.Course
	selected  map[int64][]int64
	replaced  int
	failWrite bool
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		catalogue: map[int64]models.Course{
			1: {ID: 1, Code: "ALG201", Title: "Algorithms"},
			2: {ID: 2, Code: "DBS202", Title: "Databases"},
			3: {ID: 3, Code: "WEB203", Title: "Web"},
		},
		selected: map[int64][]int64{},
	}
}

func (f *fakeRepo) List(context.Context) ([]models.Course, error) {
	return f.all(), nil
}

func (f *fakeRepo) GetByIDs(_ context.Context, ids []int64) ([]models.Course, error) {
	return f.byIDs(ids), nil
}

func (f *fakeRepo) ListByUserID(_ context.Context, userID int64) ([]models.Course, error) {
	return f.byIDs(f.selected[userID]), nil
}

func (f *fakeRepo) ReplaceUserCourses(_ context.Context, userID int64, ids []int64) error {
	if f.failWrite {
		return errors.New("db down")
	}

	f.replaced++
	f.selected[userID] = ids

	return nil
}

func (f *fakeRepo) all() []models.Course {
	out := make([]models.Course, 0, len(f.catalogue))
	for _, c := range f.catalogue {
		out = append(out, c)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })

	return out
}

func (f *fakeRepo) byIDs(ids []int64) []models.Course {
	var out []models.Course

	for _, id := range ids {
		if c, ok := f.catalogue[id]; ok {
			out = append(out, c)
		}
	}

	sort.Slice(out, func(i, j int) bool { return out[i].Code < out[j].Code })

	return out
}

func TestSaveUserCourses_SaveAndReload(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	saved, err := svc.SaveUserCourses(context.Background(), 7, []int64{2})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if len(saved) != 1 || saved[0].ID != 2 {
		t.Fatalf("unexpected saved courses: %+v", saved)
	}

	reloaded, err := svc.GetUserCourses(context.Background(), 7)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}

	if len(reloaded) != 1 || reloaded[0].ID != 2 {
		t.Fatalf("selection did not persist: %+v", reloaded)
	}
}

func TestSaveUserCourses_DeduplicatesIDs(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	saved, err := svc.SaveUserCourses(context.Background(), 7, []int64{1, 1, 3, 1})
	if err != nil {
		t.Fatalf("save: %v", err)
	}

	if len(saved) != 2 {
		t.Fatalf("expected 2 unique courses, got %+v", saved)
	}
}

func TestSaveUserCourses_ReplacesPreviousSelection(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	if _, err := svc.SaveUserCourses(context.Background(), 7, []int64{1, 2}); err != nil {
		t.Fatalf("first save: %v", err)
	}

	saved, err := svc.SaveUserCourses(context.Background(), 7, []int64{3})
	if err != nil {
		t.Fatalf("second save: %v", err)
	}

	if len(saved) != 1 || saved[0].ID != 3 {
		t.Fatalf("selection was not replaced: %+v", saved)
	}
}

func TestSaveUserCourses_EmptyClearsSelection(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	if _, err := svc.SaveUserCourses(context.Background(), 7, []int64{1}); err != nil {
		t.Fatalf("save: %v", err)
	}

	saved, err := svc.SaveUserCourses(context.Background(), 7, nil)
	if err != nil {
		t.Fatalf("clear: %v", err)
	}

	if len(saved) != 0 {
		t.Fatalf("expected empty selection, got %+v", saved)
	}
}

func TestSaveUserCourses_RejectsUnknownCourseWithoutMutation(t *testing.T) {
	repo := newFakeRepo()
	svc := NewServiceWithRepository(repo)

	if _, err := svc.SaveUserCourses(context.Background(), 7, []int64{1}); err != nil {
		t.Fatalf("seed selection: %v", err)
	}

	_, err := svc.SaveUserCourses(context.Background(), 7, []int64{2, 999})
	if !errors.Is(err, ErrInvalidCourseIDs) {
		t.Fatalf("expected ErrInvalidCourseIDs, got %v", err)
	}

	if repo.replaced != 1 {
		t.Fatalf("rejected request must not write, replaced=%d", repo.replaced)
	}

	if got := repo.selected[7]; len(got) != 1 || got[0] != 1 {
		t.Fatalf("previous selection mutated: %v", got)
	}
}

func TestSaveUserCourses_RejectsTooMany(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	ids := make([]int64, MaxUserCourses+1)
	for i := range ids {
		ids[i] = int64(i + 1)
	}

	if _, err := svc.SaveUserCourses(context.Background(), 7, ids); !errors.Is(err, ErrInvalidCourseIDs) {
		t.Fatalf("expected ErrInvalidCourseIDs, got %v", err)
	}
}

func TestSaveUserCourses_PropagatesRepositoryError(t *testing.T) {
	repo := newFakeRepo()
	repo.failWrite = true

	_, err := NewServiceWithRepository(repo).SaveUserCourses(context.Background(), 7, []int64{1})
	if err == nil || errors.Is(err, ErrInvalidCourseIDs) {
		t.Fatalf("expected repository error, got %v", err)
	}
}

func TestSelectionsAreIsolatedBetweenUsers(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	if _, err := svc.SaveUserCourses(context.Background(), 1, []int64{1}); err != nil {
		t.Fatalf("save: %v", err)
	}

	other, err := svc.GetUserCourses(context.Background(), 2)
	if err != nil {
		t.Fatalf("get: %v", err)
	}

	if len(other) != 0 {
		t.Fatalf("user 2 must not see user 1 selection: %+v", other)
	}
}
