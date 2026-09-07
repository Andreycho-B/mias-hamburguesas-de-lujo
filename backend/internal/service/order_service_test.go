package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
)

func TestOrderService_CreateOrder(t *testing.T) {
	t.Run("Rechaza comanda vacía sin productos", func(t *testing.T) {
		mock := &BaseMockQuerier{}
		svc := NewOrderService(nil, mock, nil)

		_, err := svc.CreateOrder(context.Background(), nil, CreateOrderRequest{
			OrderType: "in_store",
			Items:     []OrderItemRequest{},
		})
		if !errors.Is(err, ErrEmptyOrder) {
			t.Fatalf("se esperaba ErrEmptyOrder, obtenido: %v", err)
		}
	})

	t.Run("Rechaza tipo de orden no reconocido", func(t *testing.T) {
		mock := &BaseMockQuerier{}
		svc := NewOrderService(nil, mock, nil)

		_, err := svc.CreateOrder(context.Background(), nil, CreateOrderRequest{
			OrderType: "invalid_type",
			Items:     []OrderItemRequest{{ProductID: 1, Quantity: 1}},
		})
		if !errors.Is(err, ErrInvalidOrderType) {
			t.Fatalf("se esperaba ErrInvalidOrderType, obtenido: %v", err)
		}
	})

	t.Run("Calcula correctamente comanda en punto sin recargo de envío", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetProductByIDFunc: func(ctx context.Context, id int32) (db.Product, error) {
				if id == 1 {
					return db.Product{ID: 1, Name: "Original", Price: 17000}, nil
				}
				return db.Product{}, errors.New("producto no existe")
			},
			CreateOrderFunc: func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
				if arg.DeliveryFee != 0 {
					t.Fatalf("se esperaba deliveryFee 0 en punto, obtenido %d", arg.DeliveryFee)
				}
				if arg.Subtotal != 34000 {
					t.Fatalf("se esperaba subtotal 34000, obtenido %d", arg.Subtotal)
				}
				if arg.TotalAmount != 34000 {
					t.Fatalf("se esperaba total 34000, obtenido %d", arg.TotalAmount)
				}
				return db.Order{
					ID:          pgtype.UUID{Valid: true},
					Code:        arg.Code,
					OrderType:   arg.OrderType,
					TotalAmount: arg.TotalAmount,
				}, nil
			},
			CreateOrderItemFunc: func(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
				return db.OrderItem{}, nil
			},
			GetOrderItemsFunc: func(ctx context.Context, orderID pgtype.UUID) ([]db.GetOrderItemsRow, error) {
				return []db.GetOrderItemsRow{}, nil
			},
		}

		svc := NewOrderService(nil, mock, nil)

		resp, err := svc.CreateOrder(context.Background(), nil, CreateOrderRequest{
			OrderType: "in_store",
			Items: []OrderItemRequest{
				{ProductID: 1, Quantity: 2},
			},
		})
		if err != nil {
			t.Fatalf("error inesperado creando orden en punto: %v", err)
		}
		if resp.Order.TotalAmount != 34000 {
			t.Fatalf("total inesperado: %d", resp.Order.TotalAmount)
		}
	})

	t.Run("Calcula comanda a domicilio sumando $6.000 de tarifa base oficial", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetProductByIDFunc: func(ctx context.Context, id int32) (db.Product, error) {
				return db.Product{ID: 1, Name: "Original", Price: 17000}, nil
			},
			CreateOrderFunc: func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
				if arg.DeliveryFee != 6000 {
					t.Fatalf("se esperaba deliveryFee 6000, obtenido %d", arg.DeliveryFee)
				}
				if arg.TotalAmount != 23000 {
					t.Fatalf("se esperaba total 23000 (17000 + 6000), obtenido %d", arg.TotalAmount)
				}
				return db.Order{
					ID:          pgtype.UUID{Valid: true},
					Code:        arg.Code,
					OrderType:   arg.OrderType,
					TotalAmount: arg.TotalAmount,
				}, nil
			},
			CreateOrderItemFunc: func(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
				return db.OrderItem{}, nil
			},
			GetOrderItemsFunc: func(ctx context.Context, orderID pgtype.UUID) ([]db.GetOrderItemsRow, error) {
				return []db.GetOrderItemsRow{}, nil
			},
		}

		svc := NewOrderService(nil, mock, nil)

		resp, err := svc.CreateOrder(context.Background(), nil, CreateOrderRequest{
			OrderType: "delivery",
			Items: []OrderItemRequest{
				{ProductID: 1, Quantity: 1},
			},
			DeliveryName:    "Cliente Domicilio",
			DeliveryAddress: "Cra 10 # 20-30",
		})
		if err != nil {
			t.Fatalf("error inesperado creando domicilio: %v", err)
		}
		if resp.Order.TotalAmount != 23000 {
			t.Fatalf("total inesperado en domicilio: %d", resp.Order.TotalAmount)
		}
	})

	t.Run("Calcula correctamente comanda con combo y extras", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetProductByIDFunc: func(ctx context.Context, id int32) (db.Product, error) {
				return db.Product{ID: 1, Name: "Original", Price: 17000}, nil
			},
			CreateOrderFunc: func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
				// Base 17000 + Combo 11000 + Tocineta 5000 = 33000
				if arg.Subtotal != 33000 {
					t.Fatalf("se esperaba subtotal 33000, obtenido %d", arg.Subtotal)
				}
				if arg.TotalAmount != 33000 {
					t.Fatalf("se esperaba total 33000, obtenido %d", arg.TotalAmount)
				}
				return db.Order{
					ID:          pgtype.UUID{Valid: true},
					Code:        arg.Code,
					OrderType:   arg.OrderType,
					TotalAmount: arg.TotalAmount,
				}, nil
			},
			CreateOrderItemFunc: func(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
				if arg.UnitPrice != 33000 {
					t.Fatalf("se esperaba unit_price 33000 en item, obtenido %d", arg.UnitPrice)
				}
				return db.OrderItem{}, nil
			},
			GetOrderItemsFunc: func(ctx context.Context, orderID pgtype.UUID) ([]db.GetOrderItemsRow, error) {
				return []db.GetOrderItemsRow{}, nil
			},
		}

		svc := NewOrderService(nil, mock, nil)

		resp, err := svc.CreateOrder(context.Background(), nil, CreateOrderRequest{
			OrderType: "in_store",
			Items: []OrderItemRequest{
				{
					ProductID: 1,
					Quantity:  1,
					Customizations: map[string]interface{}{
						"combo": true,
						"extras": []interface{}{
							map[string]interface{}{"id": "tocineta", "name": "Tocineta", "price": float64(5000)},
						},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("error inesperado creando orden con combo y extras: %v", err)
		}
		if resp.Order.TotalAmount != 33000 {
			t.Fatalf("total inesperado: %d", resp.Order.TotalAmount)
		}
	})
}
