package password

import (
	dto "winx-notification/internal/app/domain/core/dto/repositories/password"
	"winx-notification/internal/app/models/models"
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNotFound        = errors.New("token not found")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidToken    = errors.New("invalid token")
	ErrInvalidPinCode  = errors.New("invalid pincode")
	ErrTokenExpired    = errors.New("token expired")
	ErrUnauthenticated = errors.New("unauthenticated")
)

type Repository interface {
	UpdatePassword(ctx context.Context, userID int64, hashed string) error
	GetUserByResetToken(ctx context.Context, input dto.ResetPasswordDTO) (models.User, error)
	GetUserByPinCode(ctx context.Context, input dto.VerifyPinDTO) (models.PasswordReset, error)
	GetById(ctx context.Context, id int64) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	CreateResetToken(ctx context.Context, email string, pinCode string, token string) error
	DeleteResetToken(ctx context.Context, token string) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateResetToken(ctx context.Context, email string, pinCode string, token string) error {
	var reset models.PasswordReset

	err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.db.WithContext(ctx).Create(&models.PasswordReset{
				Email:   email,
				PinCode: pinCode,
				Token:   token,
			}).Error
		}

		return fmt.Errorf("failed to find reset token: %w", err)
	}

	err = r.db.WithContext(ctx).
		Model(&reset).
		Updates(models.PasswordReset{PinCode: pinCode}).Error
	if err != nil {
		return fmt.Errorf("failed to update reset token: %w", err)
	}

	return nil
}

func (r *repository) GetUserByResetToken(ctx context.Context, input dto.ResetPasswordDTO) (models.User, error) {
	var reset models.PasswordReset

	err := r.db.WithContext(ctx).
		Where("token = ? AND email = ?", input.Token, input.Email).
		First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrInvalidToken
		}
		return models.User{}, fmt.Errorf("failed to get reset password: %w", err)
	}

	if time.Since(reset.CreatedAt) > time.Hour {
		return models.User{}, ErrTokenExpired
	}

	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ?", input.Email).First(&user).Error; err != nil {
		return models.User{}, ErrUserNotFound
	}

	return user, nil
}

func (r *repository) GetUserByPinCode(ctx context.Context, input dto.VerifyPinDTO) (models.PasswordReset, error) {
	var reset models.PasswordReset

	err := r.db.WithContext(ctx).
		Where("pin_code = ? AND email = ?", input.PinCode, input.Email).
		First(&reset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.PasswordReset{}, ErrInvalidPinCode
		}
		return models.PasswordReset{}, fmt.Errorf("failed to get token: %w", err)
	}

	if time.Since(reset.CreatedAt) > time.Hour {
		return models.PasswordReset{}, ErrTokenExpired
	}

	return reset, nil
}

func (r *repository) DeleteResetToken(ctx context.Context, token string) error {
	err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&models.PasswordReset{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete reset token: %w", err)
	}
	return nil
}

func (r *repository) UpdatePassword(ctx context.Context, userID int64, hashed string) error {
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("password", hashed).Error
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (r *repository) GetById(ctx context.Context, id int64) (*models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get User by id: %w", err)
	}

	return &user, nil
}

func (r *repository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get User by email: %w", err)
	}

	return &user, nil
}
