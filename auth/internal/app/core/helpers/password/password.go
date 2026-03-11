package password

import (
	"os"

	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = 12
const TestCost = 4

func Hash(plain string) (string, error) {
	cost := DefaultCost
	if os.Getenv("APP_ENV") == "test" {
		cost = TestCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plain), cost)
	return string(hash), err
}

func Check(hash string, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
