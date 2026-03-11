package password

type ForgotPasswordDTO struct {
	Email string
}

type ResetPasswordDTO struct {
	Email                   string
	Token                   string
	NewPassword             string
	NewPasswordConfirmation string
}

type ChangePasswordDTO struct {
	NewPassword     string
	Password        string
	ConfirmPassword string
	Token           string
}

type VerifyPinDTO struct {
	Email   string
	PinCode string
}
