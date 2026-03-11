package resources

import "winx-notification/internal/app/models/models"

type Resource struct {
	Email string `json:"email"`
}

func NewResource(user models.User) *Resource {
	return &Resource{
		Email: user.Email,
	}
}
