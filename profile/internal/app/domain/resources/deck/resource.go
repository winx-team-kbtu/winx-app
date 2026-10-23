package deck

import "winx-profile/internal/app/models/models"

type Resource struct {
	UserID        int64    `json:"user_id"`
	Name          string   `json:"name"`
	Skills        []string `json:"skills"`
	PreferredRole string   `json:"preferred_role"`
}

func NewResource(item models.DeckCandidate) *Resource {
	skills := item.Skills
	if skills == nil {
		skills = []string{}
	}

	return &Resource{
		UserID:        item.UserID,
		Name:          item.Name,
		Skills:        skills,
		PreferredRole: item.PreferredRole,
	}
}

func NewCollection(items []models.DeckCandidate) []*Resource {
	out := make([]*Resource, 0, len(items))
	for _, item := range items {
		out = append(out, NewResource(item))
	}

	return out
}
