package service

import "errors"

var (
	ErrInvalidToken          = errors.New("invalid token")
	ErrAuthenticationFailed = errors.New("authentication failed")
)

// TokenPayload는 인증 토큰에 포함될 사용자 메타데이터 정보입니다.
type TokenPayload struct {
	UserID   string
	Username string
	Nickname string
}

// AuthProvider는 로그인 토큰 생성 및 검증을 담당하는 추상화 인터페이스입니다.
type AuthProvider interface {
	GenerateToken(payload TokenPayload) (string, error)
	VerifyToken(token string) (*TokenPayload, error)
}
