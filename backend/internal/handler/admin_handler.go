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
		http.Error(w, `{"error":"Error al calcular métricas de venta"}`, http.StatusInternalServerError)
		return
	}

	topProducts, err := h.queries.GetTopSellingProducts(r.Context(), 10)
	if err != nil {
		http.Error(w, `{"error":"Error al cargar productos más vendidos"}`, http.StatusInternalServerError)
		return
	}

	ratingStats, _ := h.queries.GetAverageRating(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
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
		http.Error(w, `{"error":"Error al cargar registros de auditoría"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func (h *AdminHandler) CreateStaffUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Role     string `json:"role"` // "waiter", "cashier", "admin"
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	if req.Role != "waiter" && req.Role != "cashier" && req.Role != "admin" {
		http.Error(w, `{"error":"Rol no permitido"}`, http.StatusBadRequest)
		return
	}

	hash, err := crypto.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error":"Error al hashear contraseña"}`, http.StatusInternalServerError)
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
		http.Error(w, `{"error":"El correo ya se encuentra registrado"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        user.ID,
		"email":     user.Email,
		"full_name": user.FullName,
		"role":      user.Role,
	})
}
