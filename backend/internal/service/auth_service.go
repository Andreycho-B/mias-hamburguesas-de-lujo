package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/crypto"
	pkgjwt "mias-backend/internal/pkg/jwt"
	"mias-backend/internal/pkg/uuidutil"
)

var (
	ErrUserNotFound      = errors.New("usuario no encontrado")
	ErrInvalidPassword   = errors.New("credenciales incorrectas")
	ErrUnauthorizedStaff = errors.New("acceso exclusivo para personal del restaurante")
)

type AuthService struct {
	queries        db.Querier
	jwtSecret      string
	googleClientID string
}

func NewAuthService(queries db.Querier, jwtSecret string, googleClientID ...string) *AuthService {
	clientID := ""
	if len(googleClientID) > 0 {
		clientID = googleClientID[0]
	}
	return &AuthService{
		queries:        queries,
		jwtSecret:      jwtSecret,
		googleClientID: clientID,
	}
}

type StaffLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	FullName   string `json:"full_name"`
	Phone      string `json:"phone,omitempty"`
	HabeasData bool   `json:"habeas_data_accepted"`
}

type OAuthLoginRequest struct {
	Email       string `json:"email"`
	FullName    string `json:"full_name"`
	AvatarURL   string `json:"avatar_url"`
	Phone       string `json:"phone"`
	HabeasData  bool   `json:"habeas_data_accepted"`
	GoogleToken string `json:"google_token,omitempty"`
}

type AuthResponse struct {
	Tokens *pkgjwt.TokenPair `json:"tokens"`
	User   UserDTO           `json:"user"`
}

type UserDTO struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	FullName  string     `json:"full_name"`
	AvatarURL string     `json:"avatar_url,omitempty"`
	Phone     string     `json:"phone,omitempty"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
}

// LoginStaff autentica a Meseros, Cajeros y Administrador con Argon2id
func (s *AuthService) LoginStaff(ctx context.Context, req StaffLoginRequest) (*AuthResponse, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	if user.Role != "waiter" && user.Role != "cashier" && user.Role != "admin" {
		return nil, ErrUnauthorizedStaff
	}

	if !user.PasswordHash.Valid {
		return nil, ErrInvalidPassword
	}

	match, err := crypto.VerifyPassword(req.Password, user.PasswordHash.String)
	if err != nil || !match {
		return nil, ErrInvalidPassword
	}

	userID := uuidutil.FromPG(user.ID)

	tokens, err := pkgjwt.GenerateTokenPair(userID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Tokens: tokens,
		User: UserDTO{
			ID:        userID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarUrl.String,
			Phone:     user.Phone.String,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// Login autentica a usuarios (clientes o personal) con correo y contraseña cifrada Argon2id
func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	cleanEmail := strings.TrimSpace(strings.ToLower(req.Email))
	cleanPassword := strings.TrimSpace(req.Password)
	if cleanEmail == "" || cleanPassword == "" {
		return nil, ErrInvalidPassword
	}

	user, err := s.queries.GetUserByEmail(ctx, cleanEmail)
	if err != nil {
		return nil, ErrInvalidPassword
	}

	if !user.PasswordHash.Valid || user.PasswordHash.String == "" {
		return nil, errors.New("esta cuenta no tiene contraseña configurada. Ingresa con Google")
	}

	match, err := crypto.VerifyPassword(cleanPassword, user.PasswordHash.String)
	if err != nil || !match {
		return nil, ErrInvalidPassword
	}

	userID := uuidutil.FromPG(user.ID)
	tokens, err := pkgjwt.GenerateTokenPair(userID, user.Role, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Tokens: tokens,
		User: UserDTO{
			ID:        userID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarUrl.String,
			Phone:     user.Phone.String,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// Register crea una nueva cuenta de cliente con validación y cifrado de contraseña Argon2id
func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	cleanEmail := strings.TrimSpace(strings.ToLower(req.Email))
	cleanName := strings.TrimSpace(req.FullName)
	cleanPassword := strings.TrimSpace(req.Password)

	if cleanEmail == "" || cleanPassword == "" || cleanName == "" {
		return nil, errors.New("todos los campos son obligatorios")
	}

	if len(cleanPassword) < 6 {
		return nil, errors.New("la contraseña debe tener al menos 6 caracteres")
	}

	if _, err := s.queries.GetUserByEmail(ctx, cleanEmail); err == nil {
		return nil, errors.New("este correo ya se encuentra registrado. Inicia sesión")
	}

	hash, err := crypto.HashPassword(cleanPassword)
	if err != nil {
		return nil, fmt.Errorf("error procesando contraseña: %w", err)
	}

	phoneVal := pgtype.Text{String: strings.TrimSpace(req.Phone), Valid: strings.TrimSpace(req.Phone) != ""}
	user, err := s.queries.CreateStaffUser(ctx, db.CreateStaffUserParams{
		Email:        cleanEmail,
		FullName:     cleanName,
		Phone:        phoneVal,
		Role:         "customer",
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("error al registrar usuario: %w", err)
	}

	userID := uuidutil.FromPG(user.ID)
	tokens, err := pkgjwt.GenerateTokenPair(userID, "customer", s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Tokens: tokens,
		User: UserDTO{
			ID:        userID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarUrl.String,
			Phone:     user.Phone.String,
			Role:      user.Role,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// LoginOAuth autentica o registra a clientes mediante Google OAuth con Habeas Data
func (s *AuthService) LoginOAuth(ctx context.Context, req OAuthLoginRequest) (*AuthResponse, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	if !req.HabeasData {
		return nil, errors.New("debe aceptar la política de tratamiento de datos (Ley 1581) para continuar")
	}

	// Verificación server-side si se envía token de Google (Zero-Trust)
	if req.GoogleToken != "" {
		vEmail, vName, vPicture, err := s.verifyGoogleToken(ctx, req.GoogleToken)
		if err == nil && vEmail != "" {
			req.Email = vEmail
			if req.FullName == "" && vName != "" {
				req.FullName = vName
			}
			if req.AvatarURL == "" && vPicture != "" {
				req.AvatarURL = vPicture
			}
		}
	}

	if req.Email == "" {
		return nil, errors.New("el correo electrónico es obligatorio")
	}

	// Blindaje Zero-Trust: Impedir suplantación de cuentas de personal a través del flujo OAuth de clientes
	existingUser, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		if existingUser.Role != "customer" {
			return nil, errors.New("acceso no permitido: las cuentas de personal deben ingresar exclusivamente mediante la terminal de operaciones")
		}
	}

	avatar := pgtype.Text{String: req.AvatarURL, Valid: req.AvatarURL != ""}
	phone := pgtype.Text{String: req.Phone, Valid: req.Phone != ""}

	user, err := s.queries.CreateOAuthUser(ctx, db.CreateOAuthUserParams{
		Email:     req.Email,
		FullName:  req.FullName,
		AvatarUrl: avatar,
		Phone:     phone,
	})
	if err != nil {
		return nil, err
	}

	if user.Role != "customer" {
		return nil, errors.New("rol de usuario no autorizado para acceso cliente")
	}

	userID := uuidutil.FromPG(user.ID)

	tokens, err := pkgjwt.GenerateTokenPair(userID, "customer", s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		Tokens: tokens,
		User: UserDTO{
			ID:        userID,
			Email:     user.Email,
			FullName:  user.FullName,
			AvatarURL: user.AvatarUrl.String,
			Phone:     user.Phone.String,
			Role:      "customer",
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

// RefreshSession rota el Refresh Token y emite un nuevo Access Token
func (s *AuthService) RefreshSession(ctx context.Context, refreshToken string) (*pkgjwt.TokenPair, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	claims, err := pkgjwt.ValidateToken(refreshToken, s.jwtSecret, "refresh")
	if err != nil {
		return nil, err
	}

	user, err := s.queries.GetUserByID(ctx, uuidutil.ToPG(claims.UserID))
	if err != nil {
		return nil, ErrUserNotFound
	}

	userID := uuidutil.FromPG(user.ID)

	return pkgjwt.GenerateTokenPair(userID, user.Role, s.jwtSecret)
}

type googleTokenInfo struct {
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Aud           string `json:"aud"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

type googleUserInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) verifyGoogleToken(ctx context.Context, token string) (string, string, string, error) {
	client := &http.Client{Timeout: 5 * time.Second}

	// 1. Si parece un JWT (ID token firmado por Google)
	if strings.Contains(token, ".") && strings.HasPrefix(token, "eyJ") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://oauth2.googleapis.com/tokeninfo?id_token="+token, nil)
		if err == nil {
			resp, err := client.Do(req)
			if err == nil && resp.StatusCode == http.StatusOK {
				defer resp.Body.Close()
				var info googleTokenInfo
				if err := json.NewDecoder(resp.Body).Decode(&info); err == nil && info.Email != "" {
					return info.Email, info.Name, info.Picture, nil
				}
			}
		}
	}

	// 2. Si es un Access Token de OAuth 2.0 (Google Identity Services)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/oauth2/v3/userinfo", nil)
	if err != nil {
		return "", "", "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", "", errors.New("token de Google no validado")
	}

	var uInfo googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&uInfo); err != nil {
		return "", "", "", err
	}

	return uInfo.Email, uInfo.Name, uInfo.Picture, nil
}

