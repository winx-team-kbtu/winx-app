package profile

type LocationDTO struct {
	City            string
	Country         *string
	CurrentLocation *CurrentLocationDTO
}

type CurrentLocationDTO struct {
	Latitude  float64
	Longitude float64
}

type StoreDTO struct {
	UserID       int64
	Email        string
	FirstName    *string
	BirthDate    *string
	Gender       *string
	ShowGender   *bool
	ShowAge      *bool
	AboutMe      *string
	JobTitle     *string
	Company      *string
	School       *string
	InterestIDs  []int64
	InterestedIn []string
	LookingFor   *string
	Lifestyle    map[string]any
	Preferences  map[string]any
	Location     *LocationDTO
}
