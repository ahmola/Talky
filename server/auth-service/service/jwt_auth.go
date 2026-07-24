package service

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthProvider struct {
	secretKey     []byte
	tokenDuration time.Duration
}

type CustomClaims struct {
	jwt.RegisteredClaims
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
}

func NewJWTAuthProvider(secret string, duration time.Duration) *JWTAuthProvider {
	return &JWTAuthProvider{
		secretKey:     []byte(secret),
		tokenDuration: duration,
	}
}

func (j *JWTAuthProvider) GenerateToken(payload TokenPayload) (string, error) {
	claims := CustomClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   payload.Username,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID:   payload.UserID,
		Nickname: payload.Nickname,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

func (j *JWTAuthProvider) VerifyToken(tokenString string) (*TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &TokenPayload{
		UserID:   claims.UserID,
		Username: claims.Subject,
		Nickname: claims.Nickname,
	}, nil
}
