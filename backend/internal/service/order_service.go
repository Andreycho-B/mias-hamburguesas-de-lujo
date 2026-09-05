package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/uuidutil"
)

var (
	ErrInvalidOrderType = errors.New("tipo de orden inválido (debe ser in_store o delivery)")
	ErrOrderNotFound    = errors.New("orden no encontrada")
	ErrEmptyOrder       = errors.New("la orden no puede estar vacía")
)

type OrderService struct {
	queries db.Querier
	hub     *SSEHub
}

func NewOrderService(queries db.Querier, hub *SSEHub) *OrderService {
	return &OrderService{
		queries: queries,
		hub:     hub,
	}
}

type OrderItemRequest struct {
	ProductID      int32                  `json:"product_id"`
	Quantity       int32                  `json:"quantity"`
	Customizations map[string]interface{} `json:"customizations"`
}

type CreateOrderRequest struct {
	OrderType        string             `json:"order_type"` // "in_store" o "delivery"
	Items            []OrderItemRequest `json:"items"`
	DeliveryName     string             `json:"delivery_name,omitempty"`
	DeliveryPhone    string             `json:"delivery_phone,omitempty"`
	DeliveryAddress  string             `json:"delivery_address,omitempty"`
	DeliveryNotes    string             `json:"delivery_notes,omitempty"`
	PaymentMethod    string             `json:"payment_method,omitempty"`
	CashChangeAmount int32              `json:"cash_change_amount,omitempty"`
}

type OrderDetailResponse struct {
	Order           db.Order              `json:"order"`
	Items           []db.GetOrderItemsRow `json:"items"`
	WhatsAppMessage string                `json:"whatsapp_message,omitempty"`
}

type CalculatedItem struct {
	Req       OrderItemRequest
	UnitPrice int32
	Name      string
}

func (s *OrderService) CreateOrder(ctx context.Context, userID *uuid.UUID, req CreateOrderRequest) (*OrderDetailResponse, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}
	if req.OrderType != "in_store" && req.OrderType != "delivery" {
		return nil, ErrInvalidOrderType
	}

	// 1. Calcular subtotal consultando los precios reales en la base de datos
	var subtotal int32 = 0
	var calculatedItems []CalculatedItem

	for _, item := range req.Items {
		product, err := s.queries.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("producto %d no encontrado", item.ProductID)
		}
		itemTotal := product.Price * item.Quantity
		subtotal += itemTotal
		calculatedItems = append(calculatedItems, CalculatedItem{
			Req:       item,
			UnitPrice: product.Price,
			Name:      product.Name,
		})
	}

	var deliveryFee int32 = 0
	initialStatus := "PENDING_PAYMENT"

	if req.OrderType == "delivery" {
		deliveryFee = 6000 // Tarifa base de domicilio acordada en la especificación
		initialStatus = "IN_PREPARATION"
	}

	totalAmount := subtotal + deliveryFee

	// 2. Generar código alfanumérico secuencial
	prefix := "ORD"
	if req.OrderType == "delivery" {
		prefix = "DOM"
	}
	code := fmt.Sprintf("%s-%04d", prefix, rand.Intn(9000)+1000)

	var dbUserID pgtype.UUID
	if userID != nil {
		dbUserID = uuidutil.ToPG(*userID)
	}

	// 3. Crear orden principal
	order, err := s.queries.CreateOrder(ctx, db.CreateOrderParams{
		Code:             code,
		UserID:           dbUserID,
		OrderType:        req.OrderType,
		Status:           initialStatus,
		Subtotal:         subtotal,
		DeliveryFee:      deliveryFee,
		DiscountAmount:   0,
		TotalAmount:      totalAmount,
		DeliveryName:     pgtype.Text{String: req.DeliveryName, Valid: req.DeliveryName != ""},
		DeliveryPhone:    pgtype.Text{String: req.DeliveryPhone, Valid: req.DeliveryPhone != ""},
		DeliveryAddress:  pgtype.Text{String: req.DeliveryAddress, Valid: req.DeliveryAddress != ""},
		DeliveryNotes:    pgtype.Text{String: req.DeliveryNotes, Valid: req.DeliveryNotes != ""},
		PaymentMethod:    pgtype.Text{String: req.PaymentMethod, Valid: req.PaymentMethod != ""},
		CashChangeAmount: pgtype.Int4{Int32: req.CashChangeAmount, Valid: req.CashChangeAmount > 0},
	})
	if err != nil {
		return nil, err
	}

	// 4. Crear items de la orden
	for _, ci := range calculatedItems {
		customJSON, _ := json.Marshal(ci.Req.Customizations)
		_, err := s.queries.CreateOrderItem(ctx, db.CreateOrderItemParams{
			OrderID:        order.ID,
			ProductID:      ci.Req.ProductID,
			Quantity:       ci.Req.Quantity,
			UnitPrice:      ci.UnitPrice,
			Customizations: customJSON,
		})
		if err != nil {
			return nil, err
		}
	}

	// Obtener items completos
	items, err := s.queries.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	// 5. Notificar a través de SSE al panel de caja
	s.hub.Broadcast("new_order", map[string]interface{}{
		"id":         uuidutil.FromPG(order.ID),
		"code":       order.Code,
		"order_type": order.OrderType,
		"status":     order.Status,
		"total":      order.TotalAmount,
		"created_at": order.CreatedAt,
	})

	// 6. Generar mensaje para WhatsApp si es domicilio
	var whatsappMsg string
	if req.OrderType == "delivery" {
		whatsappMsg = s.formatWhatsAppMessage(order, calculatedItems)
	}

	return &OrderDetailResponse{
		Order:           order,
		Items:           items,
		WhatsAppMessage: whatsappMsg,
	}, nil
}

func (s *OrderService) MarkAsPaid(ctx context.Context, orderID uuid.UUID) (*db.Order, error) {
	order, err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     uuidutil.ToPG(orderID),
		Status: "PAID",
	})
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast("order_paid", map[string]interface{}{
		"id":     uuidutil.FromPG(order.ID),
		"code":   order.Code,
		"status": order.Status,
	})

	return &order, nil
}

func (s *OrderService) UpdateDeliveryStatus(ctx context.Context, orderID uuid.UUID, newStatus string) (*db.Order, error) {
	order, err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     uuidutil.ToPG(orderID),
		Status: newStatus,
	})
	if err != nil {
		return nil, err
	}

	s.hub.Broadcast("delivery_status_updated", map[string]interface{}{
		"id":     uuidutil.FromPG(order.ID),
		"code":   order.Code,
		"status": order.Status,
	})

	return &order, nil
}

func (s *OrderService) formatWhatsAppMessage(order db.Order, items []CalculatedItem) string {
	msg := fmt.Sprintf("🍔 *NUEVO PEDIDO A DOMICILIO — MÍAS HAMBURGUESAS* 🍔\n"+
		"*Orden:* #%s\n"+
		"*Cliente:* %s (%s)\n"+
		"*Dirección:* %s\n",
		order.Code, order.DeliveryName.String, order.DeliveryPhone.String, order.DeliveryAddress.String)

	if order.DeliveryNotes.Valid && order.DeliveryNotes.String != "" {
		msg += fmt.Sprintf("*Indicaciones:* %s\n", order.DeliveryNotes.String)
	}

	msg += "\n*Detalle del Pedido:*\n"
	for _, it := range items {
		msg += fmt.Sprintf("• %dx %s ($%d)\n", it.Req.Quantity, it.Name, it.UnitPrice*it.Req.Quantity)
	}

	msg += fmt.Sprintf("\n*Subtotal:* $%d\n"+
		"*Domicilio:* $%d\n"+
		"*Total a Pagar:* $%d\n"+
		"*Método de Pago:* %s\n",
		order.Subtotal, order.DeliveryFee, order.TotalAmount, order.PaymentMethod.String)

	if order.CashChangeAmount.Valid && order.CashChangeAmount.Int32 > 0 {
		cambio := order.CashChangeAmount.Int32 - order.TotalAmount
		msg += fmt.Sprintf("*Paga con:* $%d (Cambio: $%d)\n", order.CashChangeAmount.Int32, cambio)
	}

	return msg
}
