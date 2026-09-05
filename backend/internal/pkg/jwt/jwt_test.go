package jwt

import (
	"testing"

	"github.com/google/uuid"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-12345678901234567890"
	userID := uuid.New()
	role := "admin"

	tokens, err := GenerateTokenPair(userID, role, secret)
	if err != nil {
		t.Fatalf("Error generando token pair: %v", err)
	}

	if tokens.AccessToken == "" || tokens.RefreshToken == "" {
		t.Fatal("Los tokens generados no pueden ser vacíos")
	}

	// Validar access token
	claims, err := ValidateToken(tokens.AccessToken, secret, "access")
	if err != nil {
		t.Fatalf("Error validando access token: %v", err)
	}
	if claims.UserID != userID || claims.Role != role {
		t.Fatalf("Claims no coinciden: esperado user=%s, role=%s; obtenido user=%s, role=%s",
			userID, role, claims.UserID, claims.Role)
	}

	// Validar refresh token
	refreshClaims, err := ValidateToken(tokens.RefreshToken, secret, "refresh")
	if err != nil {
		t.Fatalf("Error validando refresh token: %v", err)
	}
	if refreshClaims.UserID != userID || refreshClaims.Role != role {
		t.Fatalf("Claims de refresh no coinciden")
	}

	// Intentar validar access token como refresh token (debe fallar)
	_, err = ValidateToken(tokens.AccessToken, secret, "refresh")
	if err != ErrWrongType {
		t.Fatalf("Esperaba ErrWrongType al cruzar tipos de token, obtuve: %v", err)
	}

	// Intentar validar con clave incorrecta
	_, err = ValidateToken(tokens.AccessToken, "clave-falsa", "access")
	if err != ErrInvalidToken {
		t.Fatalf("Esperaba ErrInvalidToken con clave incorrecta, obtuve: %v", err)
	}
}
