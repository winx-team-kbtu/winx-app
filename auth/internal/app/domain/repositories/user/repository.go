package user

import (
	dto "auth/internal/app/domain/core/dto/repositories/user"
	"auth/internal/app/models/models"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("user not found")
)

type Repository interface {
	Create(ctx context.Context, input dto.CreateDTO) (models.User, error)
	Delete(ctx context.Context, input dto.DeleteDTO) (bool, error)
	Update(ctx context.Context, input dto.UpdateDTO) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, input dto.CreateDTO) (models.User, error) {
	user := models.User{
		Email:    input.Email,
		Password: input.Password,
	}

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return models.User{}, fmt.Errorf("failed to create User: %w", err)
	}

	return user, nil
}

func (r *repository) Delete(ctx context.Context, input dto.DeleteDTO) (bool, error) {
	res := r.db.WithContext(ctx).Where("email = ?", input.Email).Delete(&models.User{})
	if res.Error != nil {
		return false, fmt.Errorf("failed to delete User: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return false, ErrNotFound
	}

	return true, nil
}

func (r *repository) Update(ctx context.Context, input dto.UpdateDTO) (models.User, error) {
	updates := map[string]interface{}{
		"email": input.NewEmail,
	}

	if input.Password != nil {
		updates["password"] = *input.Password
	}

	res := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ?", input.Email).
		Updates(updates)

	if res.Error != nil {
		return models.User{}, fmt.Errorf("failed to update user: %w", res.Error)
	}

	if res.RowsAffected == 0 {
		return models.User{}, ErrNotFound
	}

	return r.GetByEmail(ctx, input.NewEmail)
}

func (r *repository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	var user models.User

	if err := r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.User{}, ErrNotFound
		}
	}

	return user, nil
}
