package validate

import (
	dto "auth/internal/app/domain/core/dto/services/password"
	"auth/internal/app/models/models"
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

func Validate(payload dto.ChangePasswordDTO, user *models.User) error {
	if payload.Password == "" {
		return errors.New("current password is required")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return errors.New("invalid password")
	}

	if payload.NewPassword == "" {
		return errors.New("new password is required")
	}

	if payload.NewPassword != payload.ConfirmPassword {
		return errors.New("new password confirmation does not match")
	}

	if payload.NewPassword == payload.Password {
		return errors.New("new password must be different from current password")
	}

	if !isValidPassword(payload.NewPassword) {
		return errors.New("new password must be at least 8 characters and contain upper, lower, number and symbol")
	}

	return nil
}

func isValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSymbol := regexp.MustCompile(`[\W_]`).MatchString(password)

	return hasUpper && hasLower && hasNumber && hasSymbol
}
