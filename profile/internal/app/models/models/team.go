package models

import "time"

type Team struct {
	ID          int64
	CourseID    int64
	Name        string
	JoinCode    string
	MemberCount int64
	CreatedAt   time.Time
}
