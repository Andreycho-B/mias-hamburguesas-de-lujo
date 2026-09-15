package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"mias-backend/internal/database/db"
	"mias-backend/internal/middleware"
	"mias-backend/internal/pkg/uuidutil"
	"mias-backend/internal/service"
)

type OrderHandler struct {
	orderService *service.OrderService
	queries      db.Querier
	pool         *pgxpool.Pool
}

func NewOrderHandler(orderService *service.OrderService, queries db.Querier, pool *pgxpool.Pool) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		queries:      queries,
		pool:         pool,
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

	if userID == nil {
		http.Error(w, `{"error":"Debes iniciar sesión o registrarte para realizar un pedido"}`, http.StatusUnauthorized)
		return
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

	var stampStatus string
	var stampReceipt string
	if h.pool != nil {
		var st, rc string
		err := h.pool.QueryRow(r.Context(),
			"SELECT status, receipt_number FROM stamp_requests WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1",
			order.ID,
		).Scan(&st, &rc)
		if err == nil {
			stampStatus = st
			stampReceipt = rc
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order":          order,
		"items":          items,
		"stamp_status":   stampStatus,
		"receipt_number": stampReceipt,
	})
}

func (h *OrderHandler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	codesParam := strings.TrimSpace(r.URL.Query().Get("codes"))
	phoneParam := strings.TrimSpace(r.URL.Query().Get("phone"))

	if !ok && codesParam == "" && phoneParam == "" {
		http.Error(w, `{"error":"No autenticado"}`, http.StatusUnauthorized)
		return
	}

	var userPhone string
	if ok {
		user, err := h.queries.GetUserByID(r.Context(), uuidutil.ToPG(userID))
		if err == nil && user.Phone.Valid {
			userPhone = strings.TrimSpace(user.Phone.String)
		}
	}
	if userPhone == "" && phoneParam != "" {
		userPhone = phoneParam
	}

	var codeList []string
	if codesParam != "" {
		for _, c := range strings.Split(codesParam, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				codeList = append(codeList, c)
			}
		}
	}

	var orders []db.Order
	if h.pool != nil {
		var rows pgx.Rows
		var err error

		if ok {
			dbUID := uuidutil.ToPG(userID)
			if userPhone != "" {
				rows, err = h.pool.Query(r.Context(), `
					SELECT id, code, user_id, order_type, status, subtotal, delivery_fee, discount_amount, total_amount,
					       delivery_name, delivery_phone, delivery_address, delivery_notes, payment_method, cash_change_amount, created_at, updated_at
					FROM orders
					WHERE user_id = $1 OR (delivery_phone = $2 AND $2 != '')
					ORDER BY created_at DESC
					LIMIT 50
				`, dbUID, userPhone)
			} else {
				rows, err = h.pool.Query(r.Context(), `
					SELECT id, code, user_id, order_type, status, subtotal, delivery_fee, discount_amount, total_amount,
					       delivery_name, delivery_phone, delivery_address, delivery_notes, payment_method, cash_change_amount, created_at, updated_at
					FROM orders
					WHERE user_id = $1
					ORDER BY created_at DESC
					LIMIT 50
				`, dbUID)
			}
		} else if len(codeList) > 0 {
			rows, err = h.pool.Query(r.Context(), `
				SELECT id, code, user_id, order_type, status, subtotal, delivery_fee, discount_amount, total_amount,
				       delivery_name, delivery_phone, delivery_address, delivery_notes, payment_method, cash_change_amount, created_at, updated_at
				FROM orders
				WHERE code = ANY($1)
				ORDER BY created_at DESC
				LIMIT 50
			`, codeList)
		}

		if err == nil && rows != nil {
			defer rows.Close()
			for rows.Next() {
				var o db.Order
				err := rows.Scan(
					&o.ID, &o.Code, &o.UserID, &o.OrderType, &o.Status, &o.Subtotal, &o.DeliveryFee, &o.DiscountAmount, &o.TotalAmount,
					&o.DeliveryName, &o.DeliveryPhone, &o.DeliveryAddress, &o.DeliveryNotes, &o.PaymentMethod, &o.CashChangeAmount, &o.CreatedAt, &o.UpdatedAt,
				)
				if err == nil {
					orders = append(orders, o)
				}
			}
		}
	}

	type OrderWithDetails struct {
		Order       db.Order              `json:"order"`
		Items       []db.GetOrderItemsRow `json:"items"`
		StampStatus string                `json:"stamp_status,omitempty"`
		ReceiptNum  string                `json:"receipt_number,omitempty"`
	}

	resp := make([]OrderWithDetails, 0, len(orders))
	for _, o := range orders {
		items, _ := h.queries.GetOrderItems(r.Context(), o.ID)
		var st, rc string
		if sRow, err := h.queries.GetStampStatusByOrderID(r.Context(), o.ID); err == nil {
			st = sRow.Status
			rc = sRow.ReceiptNumber
		}
		resp = append(resp, OrderWithDetails{
			Order:       o,
			Items:       items,
			StampStatus: st,
			ReceiptNum:  rc,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
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
