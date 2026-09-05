package handler

import (
	"encoding/json"
	"net/http"

	"mias-backend/internal/middleware"
	"mias-backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) StaffLogin(w http.ResponseWriter, r *http.Request) {
	var req service.StaffLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authService.LoginStaff(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	// Establecer cookie HttpOnly con el Refresh Token
	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   false, // Cambiar a true en producción HTTPS
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req service.OAuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	resp, err := h.authService.LoginOAuth(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("mias_refresh_token")
	if err != nil {
		http.Error(w, `{"error":"Refresh token no encontrado"}`, http.StatusUnauthorized)
		return
	}

	tokens, err := h.authService.RefreshSession(r.Context(), cookie.Value)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"Sesión cerrada correctamente"}`))
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	role, _ := middleware.GetUserRole(r.Context())

	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id": userID,
		"role":    role,
	})
}
