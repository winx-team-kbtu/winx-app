package rules

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

type GormChecker struct{ DB *gorm.DB }

func (c GormChecker) IsUnique(ctx context.Context, table, column string, value any, scope, ignore map[string]any) (bool, error) {
	tx := c.DB.WithContext(ctx).Table(table).Where(fmt.Sprintf("%s = ?", column), value)

	for k, v := range scope {
		tx = tx.Where(fmt.Sprintf("%s = ?", k), v)
	}
	for k, v := range ignore {
		tx = tx.Where(fmt.Sprintf("%s <> ?", k), v)
	}

	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}
