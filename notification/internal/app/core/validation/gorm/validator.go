package gormvalidator

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type GormValidator struct {
	db *gorm.DB
}

func New(db *gorm.DB) *GormValidator {
	return &GormValidator{db: db}
}

func (v *GormValidator) Unique(ctx context.Context, table string, scope, ignore map[string]any) (bool, error) {
	tx := v.db.WithContext(ctx).Table(table)

	for k, v := range scope {
		tx = tx.Where(fmt.Sprintf("%s = ?", k), v)
	}
	for k, v := range ignore {
		tx = tx.Where(fmt.Sprintf("%s <> ?", k), v)
	}

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, fmt.Errorf("gorm validator unique: %w", err)
	}

	return count == 0, nil
}

func (v *GormValidator) Exists(ctx context.Context, table string, scope, ignore map[string]any) (bool, error) {
	tx := v.db.WithContext(ctx).Table(table)

	for k, v := range scope {
		tx = tx.Where(fmt.Sprintf("%s = ?", k), v)
	}
	for k, v := range ignore {
		tx = tx.Where(fmt.Sprintf("%s <> ?", k), v)
	}

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, fmt.Errorf("gorm validator exists: %w", err)
	}

	return count > 0, nil
}
