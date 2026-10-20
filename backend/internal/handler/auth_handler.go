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
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	resp, err := h.authService.LoginStaff(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Establecer cookie HttpOnly con el Refresh Token
	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	resp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	resp, err := h.authService.Register(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})

	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) OAuthLogin(w http.ResponseWriter, r *http.Request) {
	var req service.OAuthLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	resp, err := h.authService.LoginOAuth(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    resp.Tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})

	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) RefreshSession(w http.ResponseWriter, r *http.Request) {
	tokenVal := ""
	if cookie, err := r.Cookie("mias_refresh_token"); err == nil && cookie.Value != "" {
		tokenVal = cookie.Value
	}
	if tokenVal == "" && r.Body != nil {
		var bodyReq struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&bodyReq); err == nil {
			tokenVal = bodyReq.RefreshToken
		}
	}

	if tokenVal == "" {
		writeError(w, http.StatusUnauthorized, "Refresh token no encontrado")
		return
	}

	tokens, err := h.authService.RefreshSession(r.Context(), tokenVal)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    tokens.RefreshToken,
		Path:     "/api/v1/auth/refresh",
		MaxAge:   7 * 24 * 3600,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})

	writeJSON(w, http.StatusOK, tokens)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	isTLS := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteLaxMode
	if isTLS {
		sameSite = http.SameSiteNoneMode
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "mias_refresh_token",
		Value:    "",
		Path:     "/api/v1/auth/refresh",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   isTLS,
		SameSite: sameSite,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Sesión cerrada correctamente"})
}

func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	role, _ := middleware.GetUserRole(r.Context())

	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id": userID,
		"role":    role,
	})
}
