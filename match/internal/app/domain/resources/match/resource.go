package match

import (
	"time"

	svcDto "winx-match/internal/app/domain/core/dto/services/match"
	"winx-match/internal/app/models/models"
)

// Resource is the minimal match representation used when a match is first created.
type Resource struct {
	ID        int64     `json:"id"`
	UserOneID int64     `json:"user_one_id"`
	UserTwoID int64     `json:"user_two_id"`
	CreatedAt time.Time `json:"created_at"`
}

func NewResource(m models.Match) *Resource {
	return &Resource{
		ID:        m.ID,
		UserOneID: m.UserOneID,
		UserTwoID: m.UserTwoID,
		CreatedAt: m.CreatedAt,
	}
}

// EnrichedResource is returned when listing matches — includes the matched user's profile.
type EnrichedResource struct {
	ID            int64     `json:"id"`
	MatchedUserID int64     `json:"matched_user_id"`
	Name          string    `json:"name"`
	City          string    `json:"city"`
	Country       string    `json:"country"`
	PhotoURL      string    `json:"photo_url"`
	CreatedAt     time.Time `json:"created_at"`
}

func NewCollection(matches []svcDto.MatchWithProfile) []*EnrichedResource {
	result := make([]*EnrichedResource, 0, len(matches))
	for _, m := range matches {
		result = append(result, &EnrichedResource{
			ID:            m.ID,
			MatchedUserID: m.MatchedUserID,
			Name:          m.Name,
			City:          m.City,
			Country:       m.Country,
			PhotoURL:      m.PhotoURL,
			CreatedAt:     m.CreatedAt,
		})
	}
	return result
}
