package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProfileLocation struct {
	City            string           `bson:"city" json:"city"`
	Country         string           `bson:"country,omitempty" json:"country,omitempty"`
	CurrentLocation *ProfileGeoPoint `bson:"current_location,omitempty" json:"current_location,omitempty"`
}

type ProfileGeoPoint struct {
	Latitude  float64 `bson:"latitude" json:"latitude"`
	Longitude float64 `bson:"longitude" json:"longitude"`
}

type Profile struct {
	ID           bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	UserID       int64            `bson:"user_id" json:"user_id"`
	Email        string           `bson:"email" json:"email"`
	FirstName    string           `bson:"first_name,omitempty" json:"first_name,omitempty"`
	BirthDate    string           `bson:"birth_date,omitempty" json:"birth_date,omitempty"`
	Gender       string           `bson:"gender,omitempty" json:"gender,omitempty"`
	ShowGender   *bool            `bson:"show_gender,omitempty" json:"show_gender,omitempty"`
	ShowAge      *bool            `bson:"show_age,omitempty" json:"show_age,omitempty"`
	AboutMe      string           `bson:"about_me,omitempty" json:"about_me,omitempty"`
	JobTitle     string           `bson:"job_title,omitempty" json:"job_title,omitempty"`
	Company      string           `bson:"company,omitempty" json:"company,omitempty"`
	School       string           `bson:"school,omitempty" json:"school,omitempty"`
	InterestIDs  []int64          `bson:"interest_ids,omitempty" json:"interest_ids,omitempty"`
	InterestedIn []string         `bson:"interested_in,omitempty" json:"interested_in,omitempty"`
	LookingFor   string           `bson:"looking_for,omitempty" json:"looking_for,omitempty"`
	Lifestyle    map[string]any   `bson:"lifestyle,omitempty" json:"lifestyle,omitempty"`
	Preferences  map[string]any   `bson:"preferences,omitempty" json:"preferences,omitempty"`
	Location     *ProfileLocation `bson:"location,omitempty" json:"location,omitempty"`
	CreatedAt    time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time        `bson:"updated_at" json:"updated_at"`
}
