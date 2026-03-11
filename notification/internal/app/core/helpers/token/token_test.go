package token

import (
	"context"
	"encoding/base64"
	"testing"

	"github.com/go-oauth2/oauth2/v4"
)

func TestLongTokenGenerate_Token(t *testing.T) {
	tests := map[string]struct {
		gen            LongTokenGenerate
		isGenRefresh   bool
		wantAccessLen  int
		wantRefreshLen int
		wantErr        bool
	}{
		"error: invalid access token length (zero)": {
			gen: LongTokenGenerate{
				LengthBytes: 0,
			},
			isGenRefresh:   false,
			wantAccessLen:  0,
			wantRefreshLen: 0,
			wantErr:        true,
		},
		"error: invalid access token length (negative)": {
			gen: LongTokenGenerate{
				LengthBytes: -10,
			},
			isGenRefresh:   true,
			wantAccessLen:  0,
			wantRefreshLen: 0,
			wantErr:        true,
		},
		"success: access only, no refresh": {
			gen: LongTokenGenerate{
				LengthBytes: 32,
			},
			isGenRefresh:   false,
			wantAccessLen:  32,
			wantRefreshLen: 0,
			wantErr:        false,
		},
		"success: access + refresh with custom refresh length": {
			gen: LongTokenGenerate{
				LengthBytes:        16,
				RefreshLengthBytes: 24,
			},
			isGenRefresh:   true,
			wantAccessLen:  16,
			wantRefreshLen: 24,
			wantErr:        false,
		},
		"success: access + refresh with default refresh length (len*2)": {
			gen: LongTokenGenerate{
				LengthBytes:        20,
				RefreshLengthBytes: 0,
			},
			isGenRefresh:   true,
			wantAccessLen:  20,
			wantRefreshLen: 40,
			wantErr:        false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			access, refresh, err := test.gen.Token(
				context.Background(),
				&oauth2.GenerateBasic{},
				test.isGenRefresh,
			)

			if test.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				if access != "" {
					t.Fatalf("expected empty access token on error, got %q", access)
				}
				if refresh != "" {
					t.Fatalf("expected empty refresh token on error, got %q", refresh)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if test.wantAccessLen == 0 {
				if access != "" {
					t.Fatalf("expected empty access token, got %q", access)
				}
			} else {
				if access == "" {
					t.Fatalf("expected non-empty access token")
				}

				decodedAccess, err := base64.RawURLEncoding.DecodeString(access)
				if err != nil {
					t.Fatalf("failed to decode access token as base64url: %v", err)
				}
				if len(decodedAccess) != test.wantAccessLen {
					t.Fatalf("invalid access token length: want %d bytes, got %d", test.wantAccessLen, len(decodedAccess))
				}
			}

			if test.wantRefreshLen == 0 {
				if refresh != "" {
					t.Fatalf("expected empty refresh token, got %q", refresh)
				}
			} else {
				if refresh == "" {
					t.Fatalf("expected non-empty refresh token")
				}

				decodedRefresh, err := base64.RawURLEncoding.DecodeString(refresh)
				if err != nil {
					t.Fatalf("failed to decode refresh token as base64url: %v", err)
				}
				if len(decodedRefresh) != test.wantRefreshLen {
					t.Fatalf("invalid refresh token length: want %d bytes, got %d", test.wantRefreshLen, len(decodedRefresh))
				}
			}
		})
	}
}
