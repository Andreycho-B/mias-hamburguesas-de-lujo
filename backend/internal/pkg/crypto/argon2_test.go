package crypto

import (
	"testing"
)

func TestArgon2PasswordHashing(t *testing.T) {
	password := "MiasSuperSecret2026!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Error hasheando contraseña: %v", err)
	}

	if len(hash) == 0 {
		t.Fatal("El hash generado está vacío")
	}

	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("Error verificando contraseña válida: %v", err)
	}
	if !match {
		t.Fatal("La contraseña válida debería haber coincidido")
	}

	wrongMatch, err := VerifyPassword("WrongPassword123", hash)
	if err != nil {
		t.Fatalf("Error al verificar contraseña incorrecta: %v", err)
	}
	if wrongMatch {
		t.Fatal("La contraseña incorrecta NO debe coincidir")
	}
}
