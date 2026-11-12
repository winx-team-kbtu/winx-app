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
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewResource(item models.Team) *Resource {
	return &Resource{
		ID:          item.ID,
		CourseID:    item.CourseID,
		Name:        item.Name,
		JoinCode:    item.JoinCode,
		MemberCount: item.MemberCount,
		Status:      string(item.Status()),
		CreatedAt:   item.CreatedAt,
	}
}

type RosterCourse struct {
	ID    int64  `json:"id"`
	Code  string `json:"code"`
	Title string `json:"title"`
}

type RosterMember struct {
	UserID        int64    `json:"user_id"`
	Name          string   `json:"name"`
	Skills        []string `json:"skills"`
	PreferredRole string   `json:"preferred_role"`
}

type RosterResource struct {
	TeamID   int64          `json:"team_id"`
	TeamName string         `json:"team_name"`
	Size     int            `json:"size"`
	Status   string         `json:"status"`
	Course   RosterCourse   `json:"course"`
	Members  []RosterMember `json:"members"`
}

func NewRosterResource(item models.TeamRoster) *RosterResource {
	members := make([]RosterMember, 0, len(item.Members))

	for _, m := range item.Members {
		skills := m.Skills
		if skills == nil {
			skills = []string{}
		}

		members = append(members, RosterMember{
			UserID:        m.UserID,
			Name:          m.Name,
			Skills:        skills,
			PreferredRole: m.PreferredRole,
		})
	}

	return &RosterResource{
		TeamID:   item.TeamID,
		TeamName: item.TeamName,
		Size:     len(members),
		Status:   string(item.Status()),
		Course:   RosterCourse{ID: item.CourseID, Code: item.CourseCode, Title: item.CourseTitle},
		Members:  members,
	}
}
