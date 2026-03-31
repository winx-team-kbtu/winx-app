package interest

import "winx-profile/internal/app/models/models"

type Resource struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
}

func NewResource(item models.Interest) *Resource {
	return &Resource{
		ID:       item.ID,
		Name:     item.Name,
		Category: item.Category,
		IconURL:  item.IconURL,
	}
}

func NewCollection(items []models.Interest) []*Resource {
	out := make([]*Resource, 0, len(items))
	for _, item := range items {
		out = append(out, NewResource(item))
	}

	return out
}
