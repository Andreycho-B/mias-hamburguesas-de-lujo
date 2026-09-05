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

type OrderHandler struct {
	orderService *service.OrderService
	queries      db.Querier
}

func NewOrderHandler(orderService *service.OrderService, queries db.Querier) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		queries:      queries,
	}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de comanda inválido"}`, http.StatusBadRequest)
		return
	}

	var userID *uuid.UUID
	if id, ok := middleware.GetUserID(r.Context()); ok {
		userID = &id
	}

	resp, err := h.orderService.CreateOrder(r.Context(), userID, req)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *OrderHandler) GetOrderByCode(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	order, err := h.queries.GetOrderByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"Orden no encontrada"}`, http.StatusNotFound)
		return
	}

	items, err := h.queries.GetOrderItems(r.Context(), order.ID)
	if err != nil {
		http.Error(w, `{"error":"Error al cargar items"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order": order,
		"items": items,
	})
}

func (h *OrderHandler) ListPendingInStore(w http.ResponseWriter, r *http.Request) {
	orders, err := h.queries.ListPendingInStoreOrders(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Error al cargar órdenes"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) ListActiveDelivery(w http.ResponseWriter, r *http.Request) {
	orders, err := h.queries.ListActiveDeliveryOrders(r.Context())
	if err != nil {
		http.Error(w, `{"error":"Error al cargar domicilios"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orders)
}

func (h *OrderHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"ID de orden inválido"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderService.MarkAsPaid(r.Context(), orderID)
	if err != nil {
		http.Error(w, `{"error":"Error actualizando estado"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) UpdateDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, `{"error":"ID de orden inválido"}`, http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"` // "ON_THE_WAY", "DELIVERED"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Cuerpo de petición inválido"}`, http.StatusBadRequest)
		return
	}

	order, err := h.orderService.UpdateDeliveryStatus(r.Context(), orderID, req.Status)
	if err != nil {
		http.Error(w, `{"error":"Error actualizando entrega"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(order)
}
