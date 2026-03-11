package password

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHashAndCheck(t *testing.T) {
	tests := map[string]struct {
		plain      string
		wrongPlain string
		shouldErr  bool
	}{
		"success: valid password": {
			plain:      "rightpass123",
			wrongPlain: "wrongpass123",
			shouldErr:  false,
		},
		"error: empty password": {
			plain:      "",
			wrongPlain: "something",
			shouldErr:  false,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			hash, err := Hash(test.plain)

			if (err != nil) != test.shouldErr {
				t.Fatalf("error expectancy mismatch: expected shouldErr=%v, got err=%v", test.shouldErr, err)
			}

			if err != nil {
				return
			}

			if hash == "" {
				t.Fatalf("hash should not be empty")
			}

			if ok := Check(hash, test.plain); !ok {
				t.Errorf("Check(hash, plain) = false, want true")
			}

			if ok := Check(hash, test.wrongPlain); ok {
				t.Errorf("Check(hash, wrongPlain) = true, want false")
			}

			if diff := cmp.Diff(true, Check(hash, test.plain)); diff != "" {
				t.Errorf("unexpected mismatch in Check: %v", diff)
			}
		})
	}
}
