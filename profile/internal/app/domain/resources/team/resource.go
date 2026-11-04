package team

import (
	"time"

	"winx-profile/internal/app/models/models"
)

type Resource struct {
	ID          int64     `json:"id"`
	CourseID    int64     `json:"course_id"`
	Name        string    `json:"name"`
	JoinCode    string    `json:"join_code"`
	MemberCount int64     `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewResource(item models.Team) *Resource {
	return &Resource{
		ID:          item.ID,
		CourseID:    item.CourseID,
		Name:        item.Name,
		JoinCode:    item.JoinCode,
		MemberCount: item.MemberCount,
		CreatedAt:   item.CreatedAt,
	}
}
