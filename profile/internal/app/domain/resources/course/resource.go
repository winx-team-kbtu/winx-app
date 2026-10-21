package course

import "winx-profile/internal/app/models/models"

type Resource struct {
	ID    int64  `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

func NewResource(item models.Course) *Resource {
	return &Resource{
		ID:    item.ID,
		Code:  item.Code,
		Title: item.Title,
	}
}

func NewCollection(items []models.Course) []*Resource {
	out := make([]*Resource, 0, len(items))
	for _, item := range items {
		out = append(out, NewResource(item))
	}

	return out
}
