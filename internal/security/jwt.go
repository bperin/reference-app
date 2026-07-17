package security

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenIssuer struct {
	secret   []byte
	issuer   string
	audience string
	lifetime time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"sub"`
	Scopes []string  `json:"scopes,omitempty"`
	Roles  []string  `json:"roles,omitempty"`
}

func NewTokenIssuer(secret string, issuer string, audience string, lifetime time.Duration) *TokenIssuer {
	return &TokenIssuer{
		secret:   []byte(secret),
		issuer:   issuer,
		audience: audience,
		lifetime: lifetime,
	}
}

func (t *TokenIssuer) IssueAccessToken(userID uuid.UUID, scopes []string, roles []string) (string, error) {
	now := time.Now()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    t.issuer,
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{t.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(t.lifetime)),
			NotBefore: jwt.NewNumericDate(now),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.New().String(),
		},
		UserID: userID,
		Scopes: scopes,
		Roles:  roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(t.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}

	return signed, nil
}

func (t *TokenIssuer) ValidateAccessToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return t.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	if claims.Issuer != t.issuer {
		return nil, fmt.Errorf("unexpected issuer: %q", claims.Issuer)
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != t.audience {
		return nil, fmt.Errorf("unexpected audience: %v", claims.Audience)
	}

	return claims, nil
}
