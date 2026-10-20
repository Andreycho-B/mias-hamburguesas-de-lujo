package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/uuidutil"
)

var (
	ErrInvalidOrderType = errors.New("tipo de orden inválido (debe ser in_store, delivery o takeout)")
	ErrOrderNotFound    = errors.New("orden no encontrada")
	ErrEmptyOrder       = errors.New("la orden no puede estar vacía")
)

type OrderService struct {
	pool    *pgxpool.Pool
	queries db.Querier
	hub     *SSEHub
}

func NewOrderService(pool *pgxpool.Pool, queries db.Querier, hub *SSEHub) *OrderService {
	return &OrderService{
		pool:    pool,
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
	OrderType        string             `json:"order_type"` // "in_store", "delivery" o "takeout"
	Items            []OrderItemRequest `json:"items"`
	DeliveryName     string             `json:"delivery_name,omitempty"`
	DeliveryPhone    string             `json:"delivery_phone,omitempty"`
	DeliveryAddress  string             `json:"delivery_address,omitempty"`
	DeliveryNotes    string             `json:"delivery_notes,omitempty"`
	DeliveryFee      *int32             `json:"delivery_fee,omitempty"`
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
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}
	if req.OrderType != "in_store" && req.OrderType != "delivery" && req.OrderType != "takeout" {
		return nil, ErrInvalidOrderType
	}

	var deliveryFee int32 = 0
	initialStatus := "PENDING_PAYMENT"

	if req.OrderType == "delivery" {
		if req.DeliveryFee != nil {
			deliveryFee = *req.DeliveryFee
		} else {
			deliveryFee = 6000 // Tarifa base oficial de domicilio
		}
		initialStatus = "IN_PREPARATION"
	}

	prefix := "ORD"
	if req.OrderType == "delivery" {
		prefix = "DOM"
	} else if req.OrderType == "takeout" {
		prefix = "TKA"
	}
	code := fmt.Sprintf("%s-%04d", prefix, rand.Intn(9000)+1000)

	var dbUserID pgtype.UUID
	if userID != nil {
		dbUserID = uuidutil.ToPG(*userID)
	}

	// Ejecución atómica con transacción si pool está disponible
	if s.pool != nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return nil, fmt.Errorf("error iniciando transacción: %w", err)
		}
		defer tx.Rollback(ctx)

		qtx := db.New(tx)

		var subtotal int32 = 0
		var calculatedItems []CalculatedItem

		for _, item := range req.Items {
			product, err := qtx.GetProductByID(ctx, item.ProductID)
			if err != nil {
				return nil, fmt.Errorf("producto %d no encontrado", item.ProductID)
			}
			unitPrice := calculateItemUnitPrice(product.Price, item.Customizations)
			itemTotal := unitPrice * item.Quantity
			subtotal += itemTotal
			calculatedItems = append(calculatedItems, CalculatedItem{
				Req:       item,
				UnitPrice: unitPrice,
				Name:      product.Name,
			})
		}

		totalAmount := subtotal + deliveryFee

		order, err := qtx.CreateOrder(ctx, db.CreateOrderParams{
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

		for _, ci := range calculatedItems {
			customJSON, _ := json.Marshal(ci.Req.Customizations)
			_, err := qtx.CreateOrderItem(ctx, db.CreateOrderItemParams{
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

		items, err := qtx.GetOrderItems(ctx, order.ID)
		if err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, fmt.Errorf("error confirmando transacción de orden: %w", err)
		}

		if s.hub != nil {
			s.hub.Broadcast("new_order", map[string]interface{}{
				"id":         uuidutil.FromPG(order.ID),
				"code":       order.Code,
				"order_type": order.OrderType,
				"status":     order.Status,
				"total":      order.TotalAmount,
				"created_at": order.CreatedAt,
			})
		}

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

	// Modo degradado / tests unitarios con Querier mock
	var subtotal int32 = 0
	var calculatedItems []CalculatedItem

	for _, item := range req.Items {
		product, err := s.queries.GetProductByID(ctx, item.ProductID)
		if err != nil {
			return nil, fmt.Errorf("producto %d no encontrado", item.ProductID)
		}
		unitPrice := calculateItemUnitPrice(product.Price, item.Customizations)
		itemTotal := unitPrice * item.Quantity
		subtotal += itemTotal
		calculatedItems = append(calculatedItems, CalculatedItem{
			Req:       item,
			UnitPrice: unitPrice,
			Name:      product.Name,
		})
	}

	totalAmount := subtotal + deliveryFee

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

	items, err := s.queries.GetOrderItems(ctx, order.ID)
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast("new_order", map[string]interface{}{
			"id":         uuidutil.FromPG(order.ID),
			"code":       order.Code,
			"order_type": order.OrderType,
			"status":     order.Status,
			"total":      order.TotalAmount,
			"created_at": order.CreatedAt,
		})
	}

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
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	order, err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     uuidutil.ToPG(orderID),
		Status: "PAID",
	})
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast("order_paid", map[string]interface{}{
			"id":     uuidutil.FromPG(order.ID),
			"code":   order.Code,
			"status": order.Status,
		})
	}

	return &order, nil
}

func (s *OrderService) UpdateDeliveryStatus(ctx context.Context, orderID uuid.UUID, newStatus string) (*db.Order, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	order, err := s.queries.UpdateOrderStatus(ctx, db.UpdateOrderStatusParams{
		ID:     uuidutil.ToPG(orderID),
		Status: newStatus,
	})
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast("delivery_status_updated", map[string]interface{}{
			"id":     uuidutil.FromPG(order.ID),
			"code":   order.Code,
			"status": order.Status,
		})
	}

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

func calculateItemUnitPrice(basePrice int32, customizations map[string]interface{}) int32 {
	unitPrice := basePrice
	if customizations == nil {
		return unitPrice
	}
	// Combo papa + gaseosa (+$11.000)
	if combo, ok := customizations["combo"].(bool); ok && combo {
		unitPrice += 11000
	}
	// Adiciones extra (tocineta, queso, etc.)
	if extras, ok := customizations["extras"].([]interface{}); ok {
		for _, ex := range extras {
			if exMap, ok := ex.(map[string]interface{}); ok {
				if priceFloat, ok := exMap["price"].(float64); ok {
					unitPrice += int32(priceFloat)
				} else if priceInt, ok := exMap["price"].(int); ok {
					unitPrice += int32(priceInt)
				}
			}
		}
	}
	return unitPrice
}

