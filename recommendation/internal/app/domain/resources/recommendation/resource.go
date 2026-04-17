package recommendation

import (
	svcDto "winx-recommendation/internal/app/domain/core/dto/services/recommendation"
)

type Resource struct {
	UserID          int64    `json:"user_id"`
	Score           float64  `json:"score"`
	SharedInterests int      `json:"shared_interests"`
	HasPhoto        bool     `json:"has_photo"`
	DistanceKM      *float64 `json:"distance_km,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Age             int      `json:"age,omitempty"`
	LookingFor      string   `json:"looking_for,omitempty"`
}

func NewResource(c svcDto.ScoredCandidate) Resource {
	return Resource{
		UserID:          c.UserID,
		Score:           c.Score,
		SharedInterests: c.SharedInterests,
		HasPhoto:        c.HasPhoto,
		DistanceKM:      c.DistanceKM,
		Gender:          c.Gender,
		Age:             c.Age,
		LookingFor:      c.LookingFor,
	}
}

func NewCollection(candidates []svcDto.ScoredCandidate) []Resource {
	out := make([]Resource, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, NewResource(c))
	}
	return out
}
