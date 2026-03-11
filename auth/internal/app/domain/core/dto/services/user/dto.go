package user

type CreateDTO struct {
	Email    string
	Password string
}

type DeleteDTO struct {
	Email string
}

type UpdateDTO struct {
	Email    string
	NewEmail string
	Password *string
}
