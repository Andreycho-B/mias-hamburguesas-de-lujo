package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/crypto"
)

type AdminHandler struct {
	queries db.Querier
}

func NewAdminHandler(queries db.Querier) *AdminHandler {
	return &AdminHandler{queries: queries}
}

func (h *AdminHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := h.queries.GetSalesMetrics(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al calcular métricas de venta")
		return
	}

	topProducts, err := h.queries.GetTopSellingProducts(r.Context(), 10)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al cargar productos más vendidos")
		return
	}

	ratingStats, _ := h.queries.GetAverageRating(r.Context())

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"sales_metrics":   metrics,
		"top_products":    topProducts,
		"ratings_summary": ratingStats,
	})
}

func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.queries.ListAuditLogs(r.Context(), db.ListAuditLogsParams{
		Limit:  50,
		Offset: 0,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al cargar registros de auditoría")
		return
	}

	writeJSON(w, http.StatusOK, logs)
}

func (h *AdminHandler) CreateStaffUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	if req.Role != "waiter" && req.Role != "cashier" && req.Role != "admin" {
		writeError(w, http.StatusBadRequest, "Rol no permitido")
		return
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al procesar contraseña")
		return
	}

	user, err := h.queries.CreateStaffUser(r.Context(), db.CreateStaffUserParams{
		Email:        req.Email,
		FullName:     req.FullName,
		Phone:        pgtype.Text{String: req.Phone, Valid: req.Phone != ""},
		Role:         req.Role,
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, "El correo ya se encuentra registrado")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}
