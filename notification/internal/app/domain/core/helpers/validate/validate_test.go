package validate

import (
	dto "winx-notification/internal/app/domain/core/dto/services/password"
	"winx-notification/internal/app/models"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/crypto/bcrypt"
)

func TestValidateChangePassword(t *testing.T) {
	hashed, _ := bcrypt.GenerateFromPassword([]byte("CorrectPassword"), bcrypt.DefaultCost)

	user := &models.User{
		Password: string(hashed),
	}

	tests := map[string]struct {
		payload   dto.ChangePasswordDTO
		shouldErr string
	}{
		"error: empty current password": {
			payload: dto.ChangePasswordDTO{
				Password: "",
			},
			shouldErr: "current password is required",
		},

		"error: invalid current password": {
			payload: dto.ChangePasswordDTO{
				Password:        "WrongPassword",
				NewPassword:     "NewPass123!",
				ConfirmPassword: "NewPass123!",
			},
			shouldErr: "invalid password",
		},

		"error: empty new password": {
			payload: dto.ChangePasswordDTO{
				Password:        "CorrectPassword",
				NewPassword:     "",
				ConfirmPassword: "",
			},
			shouldErr: "new password is required",
		},

		"error: confirmation does not match": {
			payload: dto.ChangePasswordDTO{
				Password:        "CorrectPassword",
				NewPassword:     "NewPass123!",
				ConfirmPassword: "WrongConfirm",
			},
			shouldErr: "new password confirmation does not match",
		},

		"error: weak password": {
			payload: dto.ChangePasswordDTO{
				Password:        "CorrectPassword",
				NewPassword:     "weak",
				ConfirmPassword: "weak",
			},
			shouldErr: "new password must be at least 8 characters and contain upper, lower, number and symbol",
		},

		"error: new password equals old password": {
			payload: dto.ChangePasswordDTO{
				Password:        "CorrectPassword",
				NewPassword:     "CorrectPassword",
				ConfirmPassword: "CorrectPassword",
			},
			shouldErr: "new password must be different from current password",
		},

		"success": {
			payload: dto.ChangePasswordDTO{
				Password:        "CorrectPassword",
				NewPassword:     "StrongPass123!",
				ConfirmPassword: "StrongPass123!",
			},
			shouldErr: "",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			err := Validate(test.payload, user)

			if test.shouldErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected error %q, got nil", test.shouldErr)
			}

			if diff := cmp.Diff(test.shouldErr, err.Error()); diff != "" {
				t.Errorf("result mismatch (-expected +got): %v", diff)
			}
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := map[string]struct {
		password string
		expected bool
	}{
		"invalid: too short": {
			password: "Aa1!",
			expected: false,
		},
		"invalid: no uppercase": {
			password: "password1!",
			expected: false,
		},
		"invalid: no lowercase": {
			password: "PASSWORD1!",
			expected: false,
		},
		"invalid: no number": {
			password: "Password!!",
			expected: false,
		},
		"invalid: no symbol": {
			password: "Password1",
			expected: false,
		},
		"invalid: only lowercase": {
			password: "password",
			expected: false,
		},
		"invalid: only uppercase": {
			password: "PASSWORD",
			expected: false,
		},
		"invalid: only numbers": {
			password: "12345678",
			expected: false,
		},
		"invalid: only letesters": {
			password: "PasswordPassword",
			expected: false,
		},
		"valid: contains upper, lower, number, symbol": {
			password: "Aa1!aaaa",
			expected: true,
		},
		"valid: long and complex": {
			password: "Very$StrongPass123!",
			expected: true,
		},
		"valid: accepts underscore as symbol": {
			password: "Abcdef1_",
			expected: true,
		},
		"valid: non-latin symbols also count as symbol": {
			password: "Aa1@пароль",
			expected: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			got := isValidPassword(test.password)

			if diff := cmp.Diff(test.expected, got); diff != "" {
				t.Errorf("result mismatch (-expected +got): %v", diff)
			}
		})
	}
}
