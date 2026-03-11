package requests

type LoginDTO struct {
	Email    string
	Password string
}

type RegisterDTO struct {
	Email    string
	Password string
}

type RefreshTokenDTO struct {
	RefreshToken string
}
