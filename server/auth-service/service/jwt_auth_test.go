package service

import (
	"testing"
	"time"
)

func TestJWTAuthProvider(t *testing.T) {
	secret := "test-secret-key-12345"
	duration := 1 * time.Hour
	provider := NewJWTAuthProvider(secret, duration)

	payload := TokenPayload{
		UserID:   "user-123",
		Username: "testuser",
		Nickname: "테스터",
	}

	// 1. 토큰 생성 검증
	token, err := provider.GenerateToken(payload)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// 2. 토큰 서명 및 파싱 검증
	verifiedPayload, err := provider.VerifyToken(token)
	if err != nil {
		t.Fatalf("Failed to verify token: %v", err)
	}

	if verifiedPayload.UserID != payload.UserID {
		t.Errorf("Expected UserID %s, got %s", payload.UserID, verifiedPayload.UserID)
	}

	if verifiedPayload.Username != payload.Username {
		t.Errorf("Expected Username %s, got %s", payload.Username, verifiedPayload.Username)
	}

	if verifiedPayload.Nickname != payload.Nickname {
		t.Errorf("Expected Nickname %s, got %s", payload.Nickname, verifiedPayload.Nickname)
	}

	// 3. 비정상 토큰 예외 처리 검증
	_, err = provider.VerifyToken("invalid.token.signature")
	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}
