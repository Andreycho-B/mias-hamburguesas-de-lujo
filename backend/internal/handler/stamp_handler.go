package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"mias-backend/internal/database/db"
	"mias-backend/internal/middleware"
	"mias-backend/internal/service"
)

type StampHandler struct {
	stampService *service.StampService
	queries      db.Querier
	pool         *pgxpool.Pool
}

func NewStampHandler(stampService *service.StampService, queries db.Querier, pool *pgxpool.Pool) *StampHandler {
	return &StampHandler{
		stampService: stampService,
		queries:      queries,
		pool:         pool,
	}
}

func (h *StampHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	balance, err := h.stampService.GetBalance(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al calcular sellos")
		return
	}

	writeJSON(w, http.StatusOK, balance)
}

func (h *StampHandler) RequestStamp(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	var req struct {
		OrderID       string `json:"order_id"`
		ReceiptNumber string `json:"receipt_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	var orderID uuid.UUID
	if req.OrderID != "" {
		parsed, err := uuid.Parse(req.OrderID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "ID de orden inválido")
			return
		}
		orderID = parsed
	}

	stampReq, err := h.stampService.RequestStamp(r.Context(), userID, orderID, req.ReceiptNumber)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, stampReq)
}

func (h *StampHandler) ListPendingStaff(w http.ResponseWriter, r *http.Request) {
	requests, err := h.queries.ListAllStampRequestsForStaff(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al cargar solicitudes de sellos")
		return
	}

	writeJSON(w, http.StatusOK, requests)
}

func (h *StampHandler) ReviewStaff(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	idStr := chi.URLParam(r, "id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de solicitud inválido")
		return
	}

	var req struct {
		Approve bool `json:"approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Cuerpo de petición inválido")
		return
	}

	if err := h.stampService.ReviewStampRequest(r.Context(), reviewerID, reqID, req.Approve); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Solicitud procesada correctamente"})
}

func (h *StampHandler) RequestRedemption(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	redemption, err := h.stampService.RequestRedemption(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, redemption)
}

func (h *StampHandler) ListPendingRedemptions(w http.ResponseWriter, r *http.Request) {
	redemptions, err := h.queries.ListPendingRedemptions(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Error al cargar redenciones")
		return
	}

	writeJSON(w, http.StatusOK, redemptions)
}

func (h *StampHandler) ApproveRedemption(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "No autenticado")
		return
	}

	idStr := chi.URLParam(r, "id")
	redemptionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID de redención inválido")
		return
	}

	if err := h.stampService.ApproveRedemption(r.Context(), reviewerID, redemptionID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "Redención aprobada exitosamente"})
}
