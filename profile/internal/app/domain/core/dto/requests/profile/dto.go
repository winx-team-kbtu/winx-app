package profile

type LocationDTO struct {
	City            string              `json:"city" validate:"required,max=255"`
	Country         *string             `json:"country" validate:"omitempty,max=255"`
	CurrentLocation *CurrentLocationDTO `json:"current_location" validate:"omitempty"`
}

type CurrentLocationDTO struct {
	Latitude  float64 `json:"latitude" validate:"required,gte=-90,lte=90"`
	Longitude float64 `json:"longitude" validate:"required,gte=-180,lte=180"`
}

type StoreDTO struct {
	FirstName    *string        `json:"first_name" validate:"omitempty,max=255"`
	BirthDate    *string        `json:"birth_date" validate:"omitempty,datetime=2006-01-02"`
	Gender       *string        `json:"gender" validate:"omitempty,oneof=man woman non_binary other"`
	ShowGender   *bool          `json:"show_gender"`
	ShowAge      *bool          `json:"show_age"`
	AboutMe      *string        `json:"about_me" validate:"omitempty,max=2000"`
	JobTitle     *string        `json:"job_title" validate:"omitempty,max=255"`
	Company      *string        `json:"company" validate:"omitempty,max=255"`
	School       *string        `json:"school" validate:"omitempty,max=255"`
	InterestIDs  []int64        `json:"interest_ids" validate:"omitempty,max=20,dive,gt=0"`
	InterestedIn []string       `json:"interested_in" validate:"omitempty,max=4,dive,oneof=man woman non_binary everyone"`
	LookingFor   *string        `json:"looking_for" validate:"omitempty,oneof=long_term_partner long_term_open_to_short short_term_open_to_long_term short_term new_friends still_figuring_it_out"`
	Lifestyle    map[string]any `json:"lifestyle" validate:"omitempty"`
	Preferences  map[string]any `json:"preferences" validate:"omitempty"`
	Location     *LocationDTO   `json:"location" validate:"required"`
}
