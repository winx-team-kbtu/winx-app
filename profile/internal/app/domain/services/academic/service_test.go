package academic

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	repository "winx-profile/internal/app/domain/repositories/academic"
	"winx-profile/internal/app/models/models"
)

type fakeRepo struct {
	stored map[int64]models.AcademicProfile
	writes int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{stored: map[int64]models.AcademicProfile{}}
}

func (f *fakeRepo) GetByUserID(_ context.Context, userID int64) (models.AcademicProfile, error) {
	p, ok := f.stored[userID]
	if !ok {
		return models.AcademicProfile{}, repository.ErrNotFound
	}

	return p, nil
}

func (f *fakeRepo) Upsert(_ context.Context, p models.AcademicProfile) (models.AcademicProfile, error) {
	f.writes++
	f.stored[p.UserID] = p

	return p, nil
}

func TestGet_ReturnsEmptyProfileWhenNothingSaved(t *testing.T) {
	p, err := NewServiceWithRepository(newFakeRepo()).Get(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}

	if p.Skills == nil || len(p.Skills) != 0 || p.PreferredRole != "" {
		t.Fatalf("expected empty profile, got %+v", p)
	}
}

func TestSave_PersistsAndReloads(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	if _, err := svc.Save(context.Background(), 5, SaveInput{Skills: []string{"Go", "SQL"}, PreferredRole: "Backend"}); err != nil {
		t.Fatal(err)
	}

	got, err := svc.Get(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got.Skills, []string{"Go", "SQL"}) || got.PreferredRole != "Backend" {
		t.Fatalf("unexpected profile: %+v", got)
	}
}

func TestSave_NormalizesInput(t *testing.T) {
	svc := NewServiceWithRepository(newFakeRepo())

	got, err := svc.Save(context.Background(), 5, SaveInput{
		Skills:        []string{"  Go ", "go", "", "   ", "SQL"},
		PreferredRole: "  Designer  ",
	})
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got.Skills, []string{"Go", "SQL"}) {
		t.Fatalf("skills not normalized: %v", got.Skills)
	}

	if got.PreferredRole != "Designer" {
		t.Fatalf("role not trimmed: %q", got.PreferredRole)
	}
}

func TestSave_OptionalValues(t *testing.T) {
	got, err := NewServiceWithRepository(newFakeRepo()).Save(context.Background(), 5, SaveInput{})
	if err != nil {
		t.Fatalf("empty input must be valid: %v", err)
	}

	if len(got.Skills) != 0 || got.PreferredRole != "" {
		t.Fatalf("unexpected: %+v", got)
	}
}

func TestSave_RejectsInvalidInputWithoutWriting(t *testing.T) {
	tooMany := make([]string, MaxSkills+1)
	for i := range tooMany {
		tooMany[i] = strings.Repeat("a", 3) + string(rune('A'+i))
	}

	cases := map[string]SaveInput{
		"skill too long": {Skills: []string{strings.Repeat("x", MaxSkillLength+1)}},
		"role too long":  {PreferredRole: strings.Repeat("r", MaxRoleLength+1)},
		"too many":       {Skills: tooMany},
	}

	for name, input := range cases {
		t.Run(name, func(t *testing.T) {
			repo := newFakeRepo()

			_, err := NewServiceWithRepository(repo).Save(context.Background(), 5, input)
			if !errors.Is(err, ErrInvalidInput) {
				t.Fatalf("expected ErrInvalidInput, got %v", err)
			}

			if repo.writes != 0 {
				t.Fatal("invalid input must not be written")
			}
		})
	}
}

func TestSave_CountsRunesNotBytes(t *testing.T) {
	// 30 Cyrillic runes are 60 bytes but still within the limit.
	skill := strings.Repeat("ж", MaxSkillLength)

	if _, err := NewServiceWithRepository(newFakeRepo()).Save(context.Background(), 5, SaveInput{Skills: []string{skill}}); err != nil {
		t.Fatalf("rune-limited skill must be accepted: %v", err)
	}
}
