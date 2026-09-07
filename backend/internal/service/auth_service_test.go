package service

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/crypto"
)

func TestAuthService_LoginOAuth_Protection(t *testing.T) {
	jwtSecret := "test-secret-key-32-chars-long-security!!"

	t.Run("Rechaza solicitud sin aceptación de Habeas Data", func(t *testing.T) {
		mock := &BaseMockQuerier{}
		svc := NewAuthService(mock, jwtSecret)

		_, err := svc.LoginOAuth(context.Background(), OAuthLoginRequest{
			Email:      "cliente@example.com",
			HabeasData: false,
		})
		if err == nil || !strings.Contains(err.Error(), "Ley 1581") {
			t.Fatalf("se esperaba error de Ley 1581, se obtuvo: %v", err)
		}
	})

	t.Run("Bloquea intento de escalada de privilegios suplantando cuenta de administrador", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserByEmailFunc: func(ctx context.Context, email string) (db.User, error) {
				return db.User{
					Email: "admin@mias.com",
					Role:  "admin",
				}, nil
			},
		}
		svc := NewAuthService(mock, jwtSecret)

		_, err := svc.LoginOAuth(context.Background(), OAuthLoginRequest{
			Email:      "admin@mias.com",
			FullName:   "Atacante",
			HabeasData: true,
		})
		if err == nil {
			t.Fatal("se esperaba error de bloqueo para cuenta de personal, pero la operación tuvo éxito")
		}
		if !strings.Contains(err.Error(), "cuentas de personal") {
			t.Fatalf("mensaje de error inesperado: %v", err)
		}
	})

	t.Run("Permite registro y login exitoso de cliente con rol customer forzado", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserByEmailFunc: func(ctx context.Context, email string) (db.User, error) {
				return db.User{Role: "customer"}, nil
			},
			CreateOAuthUserFunc: func(ctx context.Context, arg db.CreateOAuthUserParams) (db.User, error) {
				return db.User{
					Email:    arg.Email,
					FullName: arg.FullName,
					Role:     "customer",
				}, nil
			},
		}
		svc := NewAuthService(mock, jwtSecret)

		resp, err := svc.LoginOAuth(context.Background(), OAuthLoginRequest{
			Email:      "comensal@gmail.com",
			FullName:   "Comensal Mías",
			HabeasData: true,
		})
		if err != nil {
			t.Fatalf("error inesperado en login oauth legítimo: %v", err)
		}
		if resp.User.Role != "customer" {
			t.Fatalf("rol esperado 'customer', obtenido: %s", resp.User.Role)
		}
		if resp.Tokens.AccessToken == "" {
			t.Fatal("access token no generado")
		}
	})
}

func TestAuthService_LoginStaff(t *testing.T) {
	jwtSecret := "test-secret-key-32-chars-long-security!!"
	validPass := "CajeroSeguro2026!"
	hashedPass, _ := crypto.HashPassword(validPass)

	t.Run("Autenticación exitosa de personal con Argon2id", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserByEmailFunc: func(ctx context.Context, email string) (db.User, error) {
				return db.User{
					Email:        "cajero@mias.com",
					Role:         "cashier",
					PasswordHash: pgtype.Text{String: hashedPass, Valid: true},
				}, nil
			},
		}
		svc := NewAuthService(mock, jwtSecret)

		resp, err := svc.LoginStaff(context.Background(), StaffLoginRequest{
			Email:    "cajero@mias.com",
			Password: validPass,
		})
		if err != nil {
			t.Fatalf("error inesperado en login staff: %v", err)
		}
		if resp.User.Role != "cashier" {
			t.Fatalf("rol esperado cashier, obtenido %s", resp.User.Role)
		}
	})

	t.Run("Rechazo ante contraseña incorrecta", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserByEmailFunc: func(ctx context.Context, email string) (db.User, error) {
				return db.User{
					Email:        "cajero@mias.com",
					Role:         "cashier",
					PasswordHash: pgtype.Text{String: hashedPass, Valid: true},
				}, nil
			},
		}
		svc := NewAuthService(mock, jwtSecret)

		_, err := svc.LoginStaff(context.Background(), StaffLoginRequest{
			Email:    "cajero@mias.com",
			Password: "PasswordIncorrecto!",
		})
		if err != ErrInvalidPassword {
			t.Fatalf("se esperaba ErrInvalidPassword, obtenido %v", err)
		}
	})
}
