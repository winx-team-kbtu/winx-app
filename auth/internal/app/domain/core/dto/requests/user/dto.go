package user

type CreateDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,max=255"`
}

type DeleteDTO struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdateDTO struct {
	Email    string  `json:"email" validate:"required,email,max=255"`
	NewEmail string  `json:"new_email" validate:"omitempty,email,max=255"`
	Password *string `json:"password" validate:"omitempty,min=8,max=50"`
}
