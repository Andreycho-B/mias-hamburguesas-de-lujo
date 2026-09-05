package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"mias-backend/internal/database/db"
	"mias-backend/internal/middleware"
	"mias-backend/internal/service"
)

type StampHandler struct {
	stampService *service.StampService
	queries      db.Querier
}

func NewStampHandler(stampService *service.StampService, queries db.Querier) *StampHandler {
	return &StampHandler{
		stampService: stampService,
		queries:      queries,
	}
}

func (h *StampHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	balance, err := h.stampService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"Error al calcular sellos"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(balance)
}

func (h *StampHandler) RequestStamp(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		OrderID       string `json:"order_id"`
		ReceiptNumber string `json:"receipt_number"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		http.Error(w, `{"error":"ID de orden inválido"}`, http.StatusBadRequest)
		return
	}

	stampReq, err := h.stampService.RequestStamp(r.Context(), userID, orderID, req.ReceiptNumber)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stampReq)
}

func (h *StampHandler) ListPendingStaff(w http.ResponseWriter, r *http.Request) {
	requests, err := h.queries.ListPendingStampRequests(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Error al cargar solicitudes de sellos"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(requests)
}

func (h *StampHandler) ReviewStaff(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	reqID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"ID de solicitud inválido"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Approve bool `json:"approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	if err := h.stampService.ReviewStampRequest(r.Context(), reviewerID, reqID, req.Approve); err != nil {
		http.Error(w, `{"error":"Error al procesar solicitud"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Solicitud procesada correctamente"}`))
}

func (h *StampHandler) RequestRedemption(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	redemption, err := h.stampService.RequestRedemption(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redemption)
}

func (h *StampHandler) ListPendingRedemptions(w http.ResponseWriter, r *http.Request) {
	redemptions, err := h.queries.ListPendingRedemptions(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Error al cargar redenciones"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(redemptions)
}

func (h *StampHandler) ApproveRedemption(w http.ResponseWriter, r *http.Request) {
	reviewerID, ok := middleware.GetUserID(r.Context())
	if !ok {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	idStr := chi.URLParam(r, "id")
	redemptionID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"ID de redención inválido"}`, http.StatusBadRequest)
		return
	}

	if err := h.stampService.ApproveRedemption(r.Context(), reviewerID, redemptionID); err != nil {
		http.Error(w, `{"error":"Error al aprobar redención"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"message":"Redención aprobada exitosamente"}`))
}
