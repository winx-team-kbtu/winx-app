package profile

import (
	"time"

	"winx-profile/internal/app/models/models"
)

type LocationResource struct {
	City            string                   `json:"city"`
	Country         string                   `json:"country,omitempty"`
	CurrentLocation *CurrentLocationResource `json:"current_location,omitempty"`
}

type CurrentLocationResource struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type InterestResource struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
}

type Resource struct {
	ID           string             `json:"id"`
	UserID       int64              `json:"user_id"`
	Email        string             `json:"email"`
	FirstName    string             `json:"first_name,omitempty"`
	BirthDate    string             `json:"birth_date,omitempty"`
	Gender       string             `json:"gender,omitempty"`
	ShowGender   *bool              `json:"show_gender,omitempty"`
	ShowAge      *bool              `json:"show_age,omitempty"`
	AboutMe      string             `json:"about_me,omitempty"`
	JobTitle     string             `json:"job_title,omitempty"`
	Company      string             `json:"company,omitempty"`
	School       string             `json:"school,omitempty"`
	InterestIDs  []int64            `json:"interest_ids,omitempty"`
	Interests    []InterestResource `json:"interests,omitempty"`
	InterestedIn []string           `json:"interested_in,omitempty"`
	LookingFor   string             `json:"looking_for,omitempty"`
	Lifestyle    map[string]any     `json:"lifestyle,omitempty"`
	Preferences  map[string]any     `json:"preferences,omitempty"`
	Location     *LocationResource  `json:"location,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

func NewResource(profile models.Profile, interests []models.Interest) *Resource {
	resource := &Resource{
		ID:           profile.ID.Hex(),
		UserID:       profile.UserID,
		Email:        profile.Email,
		FirstName:    profile.FirstName,
		BirthDate:    profile.BirthDate,
		Gender:       profile.Gender,
		ShowGender:   profile.ShowGender,
		ShowAge:      profile.ShowAge,
		AboutMe:      profile.AboutMe,
		JobTitle:     profile.JobTitle,
		Company:      profile.Company,
		School:       profile.School,
		InterestIDs:  profile.InterestIDs,
		InterestedIn: profile.InterestedIn,
		LookingFor:   profile.LookingFor,
		Lifestyle:    profile.Lifestyle,
		Preferences:  profile.Preferences,
		CreatedAt:    profile.CreatedAt,
		UpdatedAt:    profile.UpdatedAt,
	}

	if len(interests) > 0 {
		resource.Interests = make([]InterestResource, 0, len(interests))
		for _, item := range interests {
			resource.Interests = append(resource.Interests, InterestResource{
				ID:       item.ID,
				Name:     item.Name,
				Category: item.Category,
				IconURL:  item.IconURL,
			})
		}
	}

	if profile.Location != nil {
		resource.Location = &LocationResource{
			City:    profile.Location.City,
			Country: profile.Location.Country,
		}
		if profile.Location.CurrentLocation != nil {
			resource.Location.CurrentLocation = &CurrentLocationResource{
				Latitude:  profile.Location.CurrentLocation.Latitude,
				Longitude: profile.Location.CurrentLocation.Longitude,
			}
		}
	}

	return resource
}
