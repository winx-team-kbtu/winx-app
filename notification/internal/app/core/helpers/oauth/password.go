package oauth

import (
	"winx-notification/internal/app/models/models"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PasswordHandler struct {
	db *gorm.DB
}

func NewPasswordHandler(db *gorm.DB) *PasswordHandler {
	return &PasswordHandler{db: db}
}

func (h *PasswordHandler) Password(_ context.Context, _, username, password string) (string, error) {
	var user models.User

	if err := h.db.Where("email = ?", username).First(&user).Error; err != nil {
		return "", errors.New("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("invalid credentials")
	}

	return fmt.Sprintf("%d", user.ID), nil
}
