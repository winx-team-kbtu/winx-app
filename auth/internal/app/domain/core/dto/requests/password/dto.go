package password

type ForgotPasswordDTO struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordDTO struct {
	Email                   string `json:"email" binding:"required,email"`
	Token                   string `json:"token" binding:"required"`
	NewPassword             string `json:"new_password" binding:"required,min=6"`
	NewPasswordConfirmation string `json:"new_password_confirmation" binding:"required,min=6"`
}

type ChangePasswordDTO struct {
	Password        string `json:"password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"new_password_confirmation"`
}

type VerifyPinDTO struct {
	Email   string `json:"email" binding:"required,email"`
	PinCode string `json:"pin_code" binding:"required"`
}
