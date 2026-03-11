package gorm_token_store

import (
	"winx-notification/internal/app/models/models"
	"winx-notification/pkg/graylog/logger"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/go-oauth2/oauth2/v4"
	oauthErrs "github.com/go-oauth2/oauth2/v4/errors"
	oauthModels "github.com/go-oauth2/oauth2/v4/models"
	"gorm.io/gorm"
)

type GormTokenStore struct {
	db *gorm.DB
}

func NewGormTokenStore(db *gorm.DB) *GormTokenStore { return &GormTokenStore{db: db} }

func tokenToRow(info oauth2.TokenInfo) (*models.OAuthToken, error) {
	payload, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}
	row := &models.OAuthToken{
		ClientID:  info.GetClientID(),
		Payload:   payload,
		CreatedAt: time.Now(),
	}
	if v := info.GetUserID(); v != "" {
		row.UserID = &v
	}
	if v := info.GetRedirectURI(); v != "" {
		row.RedirectURI = &v
	}
	if v := info.GetScope(); v != "" {
		row.Scope = &v
	}

	if v := info.GetCode(); v != "" {
		row.Code = &v
		exp := info.GetCodeCreateAt().Add(info.GetCodeExpiresIn())
		row.CodeExpiresAt = &exp
	}
	if v := info.GetAccess(); v != "" {
		row.Access = &v
		exp := info.GetAccessCreateAt().Add(info.GetAccessExpiresIn())
		row.AccessExpiresAt = &exp
	}
	if v := info.GetRefresh(); v != "" {
		row.Refresh = &v
		exp := info.GetRefreshCreateAt().Add(info.GetRefreshExpiresIn())
		row.RefreshExpiresAt = &exp
	}

	return row, nil
}

func rowToToken(row *models.OAuthToken) (oauth2.TokenInfo, error) {
	var t oauthModels.Token
	if err := json.Unmarshal(row.Payload, &t); err != nil {
		return nil, err
	}

	return &t, nil
}

func (s *GormTokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	row, err := tokenToRow(info)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method Create when tokenToRow: %s", err.Error()))

		return err
	}

	if err = s.db.WithContext(ctx).Create(row).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method Create when create DB: %s", err.Error()))

		return err
	}

	return nil
}

func (s *GormTokenStore) RemoveByCode(ctx context.Context, code string) error {
	if err := s.db.WithContext(ctx).Where("code = ?", code).Delete(&models.OAuthToken{}).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method RemoveByCode: %s", err.Error()))

		return err
	}

	return nil
}
func (s *GormTokenStore) RemoveByAccess(ctx context.Context, access string) error {
	if err := s.db.WithContext(ctx).Where("access = ?", access).Delete(&models.OAuthToken{}).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method RemoveByAccess: %s", err.Error()))

		return err
	}

	return nil
}
func (s *GormTokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	if err := s.db.WithContext(ctx).
		Where("refresh = ?", refresh).
		Delete(&models.OAuthToken{}).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method RemoveByRefresh: %s", err.Error()))

		return err
	}

	return nil
}

func (s *GormTokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	var row models.OAuthToken
	if err := s.db.WithContext(ctx).
		Where("code = ? AND (code_expires_at IS NULL OR code_expires_at > now())", code).
		First(&row).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method GetByCode: %s", err.Error()))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauthErrs.ErrInvalidAuthorizeCode
		}

		return nil, err
	}

	tokenInfo, err := rowToToken(&row)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method GetByCode when rowToToken: %s", err.Error()))

		return tokenInfo, err
	}

	return tokenInfo, nil
}
func (s *GormTokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	var row models.OAuthToken
	if err := s.db.WithContext(ctx).
		Where("access = ? AND (access_expires_at IS NULL OR access_expires_at > now())", access).
		First(&row).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method GetByAccess: %s", err.Error()))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauthErrs.ErrInvalidAccessToken
		}

		return nil, err
	}

	tokenInfo, err := rowToToken(&row)
	if err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method GetByAccess when rowToToken: %s", err.Error()))

		return tokenInfo, err
	}

	return tokenInfo, nil
}
func (s *GormTokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	var row models.OAuthToken
	if err := s.db.WithContext(ctx).
		Where("refresh = ? AND (refresh_expires_at IS NULL OR refresh_expires_at > now())", refresh).
		First(&row).Error; err != nil {
		logger.Log.Error(fmt.Sprintf("failed GormTokenStore method GetByRefresh: %s", err.Error()))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, oauthErrs.ErrInvalidRefreshToken
		}

		return nil, err
	}

	tokenInfo, err := rowToToken(&row)
	if err != nil {
		logger.Log.Error(
			fmt.Sprintf("failed GormTokenStore method GetByRefresh when rowToToken: %s", err.Error()),
		)

		return tokenInfo, err
	}

	return tokenInfo, nil
}
