package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

type RefreshTokenGenerator struct{}

func NewRefreshTokenGenerator() *RefreshTokenGenerator {
	return &RefreshTokenGenerator{}
}

func (g *RefreshTokenGenerator) Generate() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func (g *RefreshTokenGenerator) Hash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
