package token

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/go-oauth2/oauth2/v4"
)

type LongTokenGenerate struct {
	LengthBytes        int
	RefreshLengthBytes int
}

func NewLongTokenGenerate(accessBytes, refreshBytes int) *LongTokenGenerate {
	return &LongTokenGenerate{
		LengthBytes:        accessBytes,
		RefreshLengthBytes: refreshBytes,
	}
}

func (g *LongTokenGenerate) Token(_ context.Context, _ *oauth2.GenerateBasic, isGenRefresh bool) (access, refresh string, err error) {
	if g.LengthBytes <= 0 {
		return "", "", errors.New("invalid access token length")
	}

	b := make([]byte, g.LengthBytes)

	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	access = base64.RawURLEncoding.EncodeToString(b)

	if isGenRefresh {
		rbLen := g.RefreshLengthBytes

		if rbLen <= 0 {
			rbLen = g.LengthBytes * 2
		}

		rb := make([]byte, rbLen)

		if _, err = rand.Read(rb); err != nil {
			return "", "", err
		}

		refresh = base64.RawURLEncoding.EncodeToString(rb)
	}

	return access, refresh, nil
}
