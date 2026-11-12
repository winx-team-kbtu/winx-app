package models

import "time"

type AcademicProfile struct {
	UserID        int64
	Skills        []string
	PreferredRole string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
