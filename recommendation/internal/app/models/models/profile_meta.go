package models

import "time"

// ProfileMeta mirrors the MongoDB profiles collection (winx-profile).
// Only the fields the recommendation algorithm needs are mapped.
type ProfileMeta struct {
	UserID       int64     `bson:"user_id"`
	Gender       string    `bson:"gender"`
	BirthDate    string    `bson:"birth_date"` // "YYYY-MM-DD"
	InterestedIn []string  `bson:"interested_in"`
	LookingFor   string    `bson:"looking_for"`
	AboutMe      string    `bson:"about_me"`
	UpdatedAt    time.Time `bson:"updated_at"`
}
