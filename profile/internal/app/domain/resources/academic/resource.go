package academic

import "winx-profile/internal/app/models/models"

type Resource struct {
	Skills        []string `json:"skills"`
	PreferredRole string   `json:"preferred_role"`
}

func NewResource(item models.AcademicProfile) *Resource {
	skills := item.Skills
	if skills == nil {
		skills = []string{}
	}

	return &Resource{
		Skills:        skills,
		PreferredRole: item.PreferredRole,
	}
}
