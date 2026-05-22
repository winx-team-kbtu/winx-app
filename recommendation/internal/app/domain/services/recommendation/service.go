package recommendation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	svcDto "winx-recommendation/internal/app/domain/core/dto/services/recommendation"
	candidateRepo "winx-recommendation/internal/app/domain/repositories/candidate"
	preferencesRepo "winx-recommendation/internal/app/domain/repositories/preferences"
	profilemetaRepo "winx-recommendation/internal/app/domain/repositories/profilemeta"
	"winx-recommendation/internal/app/models/models"
	"winx-recommendation/pkg/cache"

	mongodriver "go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

var ErrProfileIncomplete = errors.New("profile incomplete: set your name, birth date, gender, and who you're interested in")

const (
	cacheTTL    = 3 * time.Minute
	cachePrefix = "rec:"
	popPrefix   = "pop:"
)

type Service interface {
	List(ctx context.Context, input svcDto.ListDTO) ([]svcDto.ScoredCandidate, error)
}

type service struct {
	candidateRepo   candidateRepo.Repository
	preferencesRepo preferencesRepo.Repository
	metaRepo        profilemetaRepo.Repository
	cache           cache.Cache
}

func NewService(db *gorm.DB, mongoDB *mongodriver.Database, c cache.Cache) Service {
	return &service{
		candidateRepo:   candidateRepo.NewRepository(db),
		preferencesRepo: preferencesRepo.NewRepository(db),
		metaRepo:        profilemetaRepo.NewRepository(mongoDB),
		cache:           c,
	}
}

// List returns a scored, filtered, paginated list of recommended profiles.
//
// Flow:
//  1. Check Redis cache (key: rec:{userID}) — return on hit
//  2. Load user context: matching_preferences + interest IDs + location (Postgres)
//  3. Load user's own MongoDB profile (gender, interested_in)
//  4. Run scoring SQL: exclude swiped/matched, filter by distance, rank by shared
//     interests and proximity
//  5. Batch-fetch MongoDB profiles for all SQL candidates
//  6. Filter in-memory: age range, gender preference (both directions)
//  7. Compute final score, cache
//  8. Return paginated slice
func (s *service) List(ctx context.Context, input svcDto.ListDTO) ([]svcDto.ScoredCandidate, error) {
	// 1. Cache hit
	if cached, ok := s.fromCache(ctx, input.UserID); ok {
		return paginate(cached, input.Limit, input.Offset), nil
	}

	// 2. User context from Postgres
	uc, err := s.preferencesRepo.GetUserContext(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user context: %w", err)
	}
	log.Printf("[rec] userID=%d prefs: minAge=%d maxAge=%d maxDistKM=%d showGlobal=%v interestedIn=%v lat=%v lng=%v",
		input.UserID,
		uc.Preferences.MinAge, uc.Preferences.MaxAge, uc.Preferences.MaxDistanceKM,
		uc.Preferences.ShowMeGlobal, []string(uc.Preferences.InterestedIn),
		uc.Latitude, uc.Longitude,
	)

	// 3. Own MongoDB profile — required for completeness check
	myMeta, _ := s.metaRepo.GetByUserID(ctx, input.UserID)
	log.Printf("[rec] userID=%d myMeta: gender=%q interestedIn=%v birthDate=%q", input.UserID, myMeta.Gender, myMeta.InterestedIn, myMeta.BirthDate)

	if myMeta.BirthDate == "" || myMeta.Gender == "" || len(myMeta.InterestedIn) == 0 {
		return nil, ErrProfileIncomplete
	}

	// 4. SQL scoring query
	candidates, err := s.candidateRepo.Score(ctx, input.UserID, uc)
	if err != nil {
		return nil, fmt.Errorf("score candidates: %w", err)
	}
	log.Printf("[rec] userID=%d SQL candidates=%d", input.UserID, len(candidates))
	if len(candidates) == 0 {
		return []svcDto.ScoredCandidate{}, nil
	}

	// 5. Batch-fetch MongoDB profiles
	candidateIDs := make([]int64, 0, len(candidates))
	for _, c := range candidates {
		candidateIDs = append(candidateIDs, c.UserID)
	}
	metaMap, err := s.metaRepo.GetByUserIDs(ctx, candidateIDs)
	if err != nil {
		// Non-fatal: proceed without Mongo enrichment
		metaMap = make(map[int64]models.ProfileMeta)
	}
	log.Printf("[rec] userID=%d mongo profiles fetched=%d", input.UserID, len(metaMap))

	// 5b. Batch-read popularity scores from Redis (best-effort — non-fatal).
	popScores := s.fetchPopularityScores(ctx, candidateIDs)

	// 6. In-memory filtering + enrichment
	result := make([]svcDto.ScoredCandidate, 0, len(candidates))
	for _, c := range candidates {
		meta := metaMap[c.UserID] // zero-value ProfileMeta if not found

		// Age filter
		if age := parseAge(meta.BirthDate); age > 0 {
			if age < uc.Preferences.MinAge || age > uc.Preferences.MaxAge {
				log.Printf("[rec]   skip candidateID=%d age=%d out of [%d,%d]", c.UserID, age, uc.Preferences.MinAge, uc.Preferences.MaxAge)
				continue
			}
		}

		// Does the requesting user want to see this candidate's gender?
		// matching_preferences.interested_in (Postgres) is the source of truth;
		// fall back to MongoDB profile when it has not been set yet.
		interestedIn := []string(uc.Preferences.InterestedIn)
		if len(interestedIn) == 0 {
			interestedIn = myMeta.InterestedIn
		}
		_, hasMeta := metaMap[c.UserID]
		if len(interestedIn) > 0 && hasMeta {
			if !containsString(interestedIn, meta.Gender) {
				log.Printf("[rec]   skip candidateID=%d gender=%q not in interestedIn=%v", c.UserID, meta.Gender, interestedIn)
				continue
			}
		}

		// Reciprocal: does the candidate want to see the requesting user's gender?
		if len(meta.InterestedIn) > 0 && myMeta.Gender != "" {
			if !containsStringCI(meta.InterestedIn, myMeta.Gender) {
				log.Printf("[rec]   skip candidateID=%d their interestedIn=%v doesn't include myGender=%q", c.UserID, meta.InterestedIn, myMeta.Gender)
				continue
			}
		}

		var distKM *float64
		if c.DistanceMeters != nil {
			d := *c.DistanceMeters / 1000
			distKM = &d
		}

		result = append(result, svcDto.ScoredCandidate{
			UserID:          c.UserID,
			Score:           computeScore(c.SharedInterests, c.HasPhoto, c.DistanceMeters, popScores[c.UserID]),
			SharedInterests: c.SharedInterests,
			HasPhoto:        c.HasPhoto,
			DistanceKM:      distKM,
			Name:            c.Name,
			City:            c.City,
			Country:         c.Country,
			PhotoURL:        c.PhotoURL,
			Gender:          meta.Gender,
			Age:             parseAge(meta.BirthDate),
			LookingFor:      meta.LookingFor,
			AboutMe:         meta.AboutMe,
		})
	}

	// 7. Sort by final score descending (popularity was added in-memory so SQL
	// order no longer reflects the true ranking).
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Score > result[j].Score
	})

	// 8. Cache full list — skip caching empty results so future candidates
	// (e.g. newly seeded/registered users) are not hidden until TTL expires.
	if len(result) > 0 {
		s.toCache(ctx, input.UserID, result)
	}

	// 9. Paginate
	return paginate(result, input.Limit, input.Offset), nil
}

// computeScore produces a float64 ranking score.
//
//	shared_interests × 3  — primary signal (0–120+ for 40 interests)
//	has_photo × 2         — small quality boost
//	popularity × 0.5      — net right-swipes from other users (can be negative)
//	distance penalty      — −0.2 per km  (50 km ≈ −10 pts)
func computeScore(sharedInterests int, hasPhoto bool, distMeters *float64, popularity int64) float64 {
	score := float64(sharedInterests) * 3
	if hasPhoto {
		score += 2
	}
	score += float64(popularity) * 0.5
	if distMeters != nil {
		score -= (*distMeters / 1000) * 0.2
	}
	return score
}

// fetchPopularityScores reads the net right-swipe counters for all candidates
// in a single Redis round trip (MGet). Missing keys are treated as 0.
// Errors are silently ignored so a Redis hiccup never breaks recommendations.
func (s *service) fetchPopularityScores(ctx context.Context, userIDs []int64) map[int64]int64 {
	if len(userIDs) == 0 {
		return map[int64]int64{}
	}

	keys := make([]string, len(userIDs))
	for i, id := range userIDs {
		keys[i] = fmt.Sprintf("%s%d", popPrefix, id)
	}

	vals, err := s.cache.MGet(ctx, keys...)
	scores := make(map[int64]int64, len(userIDs))
	if err != nil {
		return scores
	}

	for i, id := range userIDs {
		if b, ok := vals[keys[i]]; ok {
			var n int64
			fmt.Sscanf(string(b), "%d", &n)
			scores[id] = n
		}
	}
	return scores
}

// parseAge converts a birth date string to age in years.
// Tries the three most common date formats in order.
func parseAge(birthDate string) int {
	if birthDate == "" {
		return 0
	}
	for _, layout := range []string{"2006-01-02", "02.01.2006", "01/02/2006"} {
		t, err := time.Parse(layout, birthDate)
		if err != nil {
			continue
		}
		now := time.Now()
		age := now.Year() - t.Year()
		if now.YearDay() < t.YearDay() {
			age--
		}
		return age
	}
	return 0
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func containsStringCI(slice []string, s string) bool {
	lower := strings.ToLower(s)
	for _, v := range slice {
		if strings.ToLower(v) == lower {
			return true
		}
	}
	return false
}

func paginate(list []svcDto.ScoredCandidate, limit, offset int) []svcDto.ScoredCandidate {
	if offset >= len(list) {
		return []svcDto.ScoredCandidate{}
	}
	if limit == 0 {
		limit = 20
	}
	end := offset + limit
	if end > len(list) {
		end = len(list)
	}
	return list[offset:end]
}

// ---- cache helpers ----

func cacheKey(userID int64) string {
	return fmt.Sprintf("%s%d", cachePrefix, userID)
}

func (s *service) fromCache(ctx context.Context, userID int64) ([]svcDto.ScoredCandidate, bool) {
	b, err := s.cache.Get(ctx, cacheKey(userID))
	if err != nil {
		return nil, false
	}
	var list []svcDto.ScoredCandidate
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, false
	}
	return list, true
}

func (s *service) toCache(ctx context.Context, userID int64, list []svcDto.ScoredCandidate) {
	b, err := json.Marshal(list)
	if err != nil {
		return
	}
	_ = s.cache.Set(ctx, cacheKey(userID), b, cacheTTL)
}
