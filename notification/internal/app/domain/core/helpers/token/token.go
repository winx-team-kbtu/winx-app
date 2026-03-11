package token

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateRandomToken(length int) string {
	bytes := make([]byte, length)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
