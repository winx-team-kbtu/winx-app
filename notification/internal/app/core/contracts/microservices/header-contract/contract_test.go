package headercontract

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGetAuthUser(t *testing.T) {
	tests := map[string]struct {
		ctx       context.Context
		expected  AuthUser
		shouldErr bool
	}{
		"success: auth user exists in context": {
			ctx: context.WithValue(
				context.Background(),
				AuthUserKey{},
				AuthUser{
					ID:    1,
					Email: "user@example.com",
				},
			),
			expected: AuthUser{
				ID:    1,
				Email: "user@example.com",
			},
			shouldErr: false,
		},
		"error: auth user missing in context": {
			ctx:       context.Background(),
			expected:  AuthUser{},
			shouldErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			got, err := GetAuthUser(test.ctx)

			if test.shouldErr != (err != nil) {
				t.Fatalf("error expectancy mismatch: shouldErr=%v, err!=nil=%v", test.shouldErr, err != nil)
			}

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("result mismatch (-expected +got): %v", diff)
			}
		})
	}
}
