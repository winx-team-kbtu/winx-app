package social

type SocialAuthDTO struct {
	Provider       string
	ProviderUserID string
	Email          string
}

type SocialUserDTO struct {
	ID    string
	Email string
}
