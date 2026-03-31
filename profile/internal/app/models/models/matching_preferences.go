package models

import (
	"time"

	"github.com/lib/pq"
)

type MatchingPreferences struct {
	UserID           int64          `json:"user_id"`
	MinAge           int            `json:"min_age"`
	MaxAge           int            `json:"max_age"`
	MaxDistanceKM    int            `json:"max_distance_km"`
	InterestedIn     pq.StringArray `json:"interested_in" gorm:"type:varchar(50)[]"`
	ShowMeGlobal     bool           `json:"show_me_global"`
	OnlyShowVerified bool           `json:"only_show_verified"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (MatchingPreferences) TableName() string {
	return "matching_preferences"
}
