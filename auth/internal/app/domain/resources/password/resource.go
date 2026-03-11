package resources

import "auth/internal/app/models/models"

type Resource struct {
	Token string `json:"token"`
}

func NewResource(resetPassword models.PasswordReset) *Resource {
	return &Resource{
		Token: resetPassword.Token,
	}
}
