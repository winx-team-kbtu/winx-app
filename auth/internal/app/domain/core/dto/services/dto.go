package requests

type LoginDTO struct {
	Email    string
	Password string
}

type RegisterDTO struct {
	Email           string
	Password        string
	ConsentAccepted bool
}

type RefreshTokenDTO struct {
	RefreshToken string
}
