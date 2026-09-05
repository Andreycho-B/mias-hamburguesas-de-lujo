package service

import (
	"context"
	"errors"
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
	queries   db.Querier
	jwtSecret string
}

func NewAuthService(queries db.Querier, jwtSecret string) *AuthService {
	return &AuthService{
		queries:   queries,
		jwtSecret: jwtSecret,
	}
}

type StaffLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OAuthLoginRequest struct {
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	AvatarURL  string `json:"avatar_url"`
	Phone      string `json:"phone"`
	HabeasData bool   `json:"habeas_data_accepted"`
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

// LoginOAuth autentica o registra a clientes mediante Google OAuth con Habeas Data
func (s *AuthService) LoginOAuth(ctx context.Context, req OAuthLoginRequest) (*AuthResponse, error) {
	if !req.HabeasData {
		return nil, errors.New("debe aceptar la política de tratamiento de datos (Ley 1581) para continuar")
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

// RefreshSession rota el Refresh Token y emite un nuevo Access Token
func (s *AuthService) RefreshSession(ctx context.Context, refreshToken string) (*pkgjwt.TokenPair, error) {
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
