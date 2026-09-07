package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
)

func TestStampService_GetBalance(t *testing.T) {
	testUserID := uuid.New()

	t.Run("Indica can_redeem falso cuando saldo es inferior a 10 sellos", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserStampBalanceFunc: func(ctx context.Context, userID pgtype.UUID) (int32, error) {
				return 7, nil
			},
			GetPendingRedemptionByUserFunc: func(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error) {
				return db.RedemptionRequest{}, errors.New("ninguna pendiente")
			},
		}

		svc := NewStampService(nil, mock, nil)
		resp, err := svc.GetBalance(context.Background(), testUserID)
		if err != nil {
			t.Fatalf("error inesperado en balance: %v", err)
		}
		if resp.ActiveStamps != 7 {
			t.Fatalf("se esperaban 7 sellos, obtenido %d", resp.ActiveStamps)
		}
		if resp.CanRedeem {
			t.Fatal("no debería poder redimir con 7 sellos")
		}
	})

	t.Run("Indica can_redeem verdadero cuando saldo alcanza 10 sellos sin redención pendiente", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserStampBalanceFunc: func(ctx context.Context, userID pgtype.UUID) (int32, error) {
				return 10, nil
			},
			GetPendingRedemptionByUserFunc: func(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error) {
				return db.RedemptionRequest{}, errors.New("ninguna")
			},
		}

		svc := NewStampService(nil, mock, nil)
		resp, err := svc.GetBalance(context.Background(), testUserID)
		if err != nil {
			t.Fatalf("error inesperado: %v", err)
		}
		if !resp.CanRedeem {
			t.Fatal("debería poder redimir con 10 sellos")
		}
	})
}

func TestStampService_RequestStamp_Rules(t *testing.T) {
	testUserID := uuid.New()
	testOrderID := uuid.New()

	t.Run("Rechaza solicitud si la orden no está en estado PAID (Orden Paga)", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetOrderByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.Order, error) {
				return db.Order{
					Status: "PENDING_PAYMENT",
				}, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		_, err := svc.RequestStamp(context.Background(), testUserID, testOrderID, "00124")
		if !errors.Is(err, ErrOrderNotEligible) {
			t.Fatalf("se esperaba ErrOrderNotEligible, obtenido: %v", err)
		}
	})

	t.Run("Rechaza solicitud si transcurrieron más de 24 horas desde la compra", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetOrderByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.Order, error) {
				return db.Order{
					Status: "PAID",
					CreatedAt: pgtype.Timestamptz{
						Time:  time.Now().Add(-25 * time.Hour),
						Valid: true,
					},
				}, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		_, err := svc.RequestStamp(context.Background(), testUserID, testOrderID, "00124")
		if !errors.Is(err, ErrOrderExpired) {
			t.Fatalf("se esperaba ErrOrderExpired, obtenido: %v", err)
		}
	})
}

func TestStampService_RequestRedemption(t *testing.T) {
	testUserID := uuid.New()

	t.Run("Rechaza redención si el saldo es menor a 10", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetUserStampBalanceFunc: func(ctx context.Context, userID pgtype.UUID) (int32, error) {
				return 9, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		_, err := svc.RequestRedemption(context.Background(), testUserID)
		if !errors.Is(err, ErrInsufficientStamps) {
			t.Fatalf("se esperaba ErrInsufficientStamps, obtenido: %v", err)
		}
	})
}
