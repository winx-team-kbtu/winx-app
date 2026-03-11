package gorm_token_store

import (
	"context"

	"github.com/go-oauth2/oauth2/v4"
)

type GormTokenStoreMock struct {
	CreateFn          func(ctx context.Context, info oauth2.TokenInfo) error
	RemoveByCodeFn    func(ctx context.Context, code string) error
	RemoveByAccessFn  func(ctx context.Context, access string) error
	RemoveByRefreshFn func(ctx context.Context, refresh string) error
	GetByCodeFn       func(ctx context.Context, code string) (oauth2.TokenInfo, error)
	GetByAccessFn     func(ctx context.Context, access string) (oauth2.TokenInfo, error)
	GetByRefreshFn    func(ctx context.Context, refresh string) (oauth2.TokenInfo, error)
}

func (m *GormTokenStoreMock) Create(ctx context.Context, info oauth2.TokenInfo) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, info)
	}

	return nil
}

func (m *GormTokenStoreMock) RemoveByCode(ctx context.Context, code string) error {
	if m.RemoveByCodeFn != nil {
		return m.RemoveByCodeFn(ctx, code)
	}

	return nil
}

func (m *GormTokenStoreMock) RemoveByAccess(ctx context.Context, access string) error {
	if m.RemoveByAccessFn != nil {
		return m.RemoveByAccessFn(ctx, access)
	}

	return nil
}

func (m *GormTokenStoreMock) RemoveByRefresh(ctx context.Context, refresh string) error {
	if m.RemoveByRefreshFn != nil {
		return m.RemoveByRefreshFn(ctx, refresh)
	}

	return nil
}

func (m *GormTokenStoreMock) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	if m.GetByCodeFn != nil {
		return m.GetByCodeFn(ctx, code)
	}

	return nil, nil
}

func (m *GormTokenStoreMock) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	if m.GetByAccessFn != nil {
		return m.GetByAccessFn(ctx, access)
	}

	return nil, nil
}

func (m *GormTokenStoreMock) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	if m.GetByRefreshFn != nil {
		return m.GetByRefreshFn(ctx, refresh)
	}

	return nil, nil
}
