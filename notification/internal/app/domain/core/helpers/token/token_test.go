package token

import (
	"encoding/hex"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGenerateRandomToken(t *testing.T) {
	tests := map[string]struct {
		length int
	}{
		"len 0":  {length: 0},
		"len 1":  {length: 1},
		"len 16": {length: 16},
		"len 32": {length: 32},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			token := GenerateRandomToken(test.length)

			expectedLen := test.length * 2

			if diff := cmp.Diff(expectedLen, len(token)); diff != "" {
				t.Errorf("result mismatch (-expected +got): %v", diff)
			}

			decoded, err := hex.DecodeString(token)
			if err != nil {
				t.Fatalf("token is not valid hex: %v", err)
			}

			if diff := cmp.Diff(test.length, len(decoded)); diff != "" {
				t.Errorf("result mismatch (-expected +got): %v", diff)
			}
		})
	}
}
