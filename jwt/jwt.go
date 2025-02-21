package jwt

import (
	"time"

	"github.com/pkg/errors"

	jwt2 "github.com/golang-jwt/jwt/v5"
)

type JwtStoreValue interface {
	ToMap() map[string]any
	FromMap(m map[string]any) error
}

type Claims struct {
	jwt2.RegisteredClaims
	Value map[string]any `json:"value"`
}

type TokenResponse struct {
	Token     string
	ExpiresAt time.Time
}

func ParseToken(tokenString string, secret string, dest JwtStoreValue) error {
	t, err := jwt2.ParseWithClaims(tokenString, &Claims{},
		func(t *jwt2.Token) (any, error) {
			if _, ok := t.Method.(*jwt2.SigningMethodHMAC); !ok {
				return nil, errors.New("invalid signing method")
			}
			return []byte(secret), nil
		})

	if err != nil {
		return errors.WithMessage(err, "parse with claims")
	}

	claims, ok := t.Claims.(*Claims)
	if !ok {
		return errors.New("token claims are not of type")
	}
	return dest.FromMap(claims.Value)
}

func GenerateToken(secret string, tokenTTL time.Duration, value JwtStoreValue) (*TokenResponse, error) {
	expiresAt := time.Now().UTC().Add(tokenTTL)
	registeredClaims := jwt2.RegisteredClaims{
		ExpiresAt: &jwt2.NumericDate{Time: expiresAt},
		IssuedAt:  &jwt2.NumericDate{Time: time.Now().UTC()}}

	token := jwt2.NewWithClaims(jwt2.SigningMethodHS256, &Claims{registeredClaims, value.ToMap()})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return nil, errors.WithMessage(err, "can't create token")
	}
	return &TokenResponse{
		Token:     tokenString,
		ExpiresAt: expiresAt,
	}, nil
}
