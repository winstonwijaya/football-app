// Package token issues and verifies the JWT access tokens used to
// authenticate every protected route.
package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid or expired token")

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

// Generate signs a token identifying userID, valid for ttl.
func Generate(secret string, userID int64, ttl time.Duration) (tokenString string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(ttl)
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tokenString, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return tokenString, expiresAt, err
}

// Parse verifies the token's signature and expiry and returns its claims.
func Parse(secret string, tokenString string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
