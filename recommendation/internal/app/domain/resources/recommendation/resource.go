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
	Name            string   `json:"name,omitempty"`
	City            string   `json:"city,omitempty"`
	Country         string   `json:"country,omitempty"`
	PhotoURL        string   `json:"photo_url,omitempty"`
	Gender          string   `json:"gender,omitempty"`
	Age             int      `json:"age,omitempty"`
	LookingFor      string   `json:"looking_for,omitempty"`
	AboutMe         string   `json:"about_me,omitempty"`
}

func NewResource(c svcDto.ScoredCandidate) Resource {
	return Resource{
		UserID:          c.UserID,
		Score:           c.Score,
		SharedInterests: c.SharedInterests,
		HasPhoto:        c.HasPhoto,
		DistanceKM:      c.DistanceKM,
		Name:            c.Name,
		City:            c.City,
		Country:         c.Country,
		PhotoURL:        c.PhotoURL,
		Gender:          c.Gender,
		Age:             c.Age,
		LookingFor:      c.LookingFor,
		AboutMe:         c.AboutMe,
	}
}

func NewCollection(candidates []svcDto.ScoredCandidate) []Resource {
	out := make([]Resource, 0, len(candidates))
	for _, c := range candidates {
		out = append(out, NewResource(c))
	}
	return out
}
