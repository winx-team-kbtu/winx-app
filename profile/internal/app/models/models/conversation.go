package models

import "time"

type Conversation struct {
	ID            int64      `json:"id"`
	CourseID      int64      `json:"course_id"`
	CourseCode    string     `json:"course_code"`
	CourseTitle   string     `json:"course_title"`
	PeerID        int64      `json:"peer_id"`
	PeerName      string     `json:"peer_name"`
	PreferredRole string     `json:"preferred_role"`
	Skills        []string   `json:"skills" gorm:"-"`
	SkillsJSON    []byte     `json:"-" gorm:"column:skills"`
	Status        string     `json:"status"`
	InitiatedBy   *int64     `json:"initiated_by"`
	ClosedBy      *int64     `json:"closed_by"`
	TeamID        *int64     `json:"team_id"`
	MyTeamID      *int64     `json:"my_team_id"`
	MyTeamName    string     `json:"my_team_name"`
	MyTeamSize    int64      `json:"my_team_size"`
	PeerTeamID    *int64     `json:"peer_team_id"`
	LastMessage   string     `json:"last_message"`
	LastMessageAt *time.Time `json:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at"`
	CanWrite      bool       `json:"can_write"`

	InvitationID       *int64  `json:"invitation_id"`
	InviterID          *int64  `json:"inviter_id"`
	InvitationStatus   *string `json:"invitation_status"`
	InvitationTeamName *string `json:"invitation_team_name"`
}
type ConversationMessage struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	SenderID  int64     `json:"sender_id"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
