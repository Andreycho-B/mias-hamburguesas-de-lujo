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

	t.Run("Permite solicitud si la orden está en estado DELIVERED (Domicilio entregado)", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetOrderByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.Order, error) {
				return db.Order{
					Status:    "DELIVERED",
					CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				}, nil
			},
			CountBurgersInOrderFunc: func(ctx context.Context, orderID pgtype.UUID) (int32, error) {
				return 2, nil
			},
			CreateStampRequestFunc: func(ctx context.Context, arg db.CreateStampRequestParams) (db.StampRequest, error) {
				return db.StampRequest{
					ID:            pgtype.UUID{Valid: true},
					ReceiptNumber: arg.ReceiptNumber,
					BurgersCount:  arg.BurgersCount,
					Status:        "PENDING",
				}, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		resp, err := svc.RequestStamp(context.Background(), testUserID, testOrderID, "00125")
		if err != nil {
			t.Fatalf("error inesperado en orden DELIVERED: %v", err)
		}
		if resp.BurgersCount != 2 {
			t.Fatalf("se esperaban 2 hamburguesas, obtenido %d", resp.BurgersCount)
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

	t.Run("Rechaza solicitud si el usuario ya alcanzó el tope de 10 sellos (regla de congelación)", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetOrderByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.Order, error) {
				return db.Order{
					Status:    "PAID",
					CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
				}, nil
			},
			GetUserStampBalanceFunc: func(ctx context.Context, userID pgtype.UUID) (int32, error) {
				return 10, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		_, err := svc.RequestStamp(context.Background(), testUserID, testOrderID, "00129")
		if !errors.Is(err, ErrStampLimitReached) {
			t.Fatalf("se esperaba ErrStampLimitReached, obtenido: %v", err)
		}
	})
}

func TestStampService_ReviewStampRequest_Security(t *testing.T) {
	staffUserID := uuid.New()
	customerUserID := uuid.New()
	reqID := uuid.New()

	t.Run("Prohíbe auto-aprobación de solicitudes por parte del mismo empleado", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetStampRequestByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.GetStampRequestByIDRow, error) {
				return db.GetStampRequestByIDRow{
					ID:     id,
					UserID: pgtype.UUID{Bytes: staffUserID, Valid: true},
					Status: "PENDING",
				}, nil
			},
		}

		// En este caso staffUserID intenta aprobar su propia solicitud
		svc := NewStampService(nil, mock, nil)
		err := svc.ReviewStampRequest(context.Background(), staffUserID, reqID, true)
		if !errors.Is(err, ErrSelfApproval) {
			t.Fatalf("se esperaba ErrSelfApproval, obtenido: %v", err)
		}
	})

	t.Run("Rechaza revisión si la solicitud ya no está PENDING", func(t *testing.T) {
		mock := &BaseMockQuerier{
			GetStampRequestByIDFunc: func(ctx context.Context, id pgtype.UUID) (db.GetStampRequestByIDRow, error) {
				return db.GetStampRequestByIDRow{
					ID:     id,
					UserID: pgtype.UUID{Bytes: customerUserID, Valid: true},
					Status: "APPROVED",
				}, nil
			},
		}

		svc := NewStampService(nil, mock, nil)
		err := svc.ReviewStampRequest(context.Background(), staffUserID, reqID, true)
		if !errors.Is(err, ErrRequestNotPending) {
			t.Fatalf("se esperaba ErrRequestNotPending, obtenido: %v", err)
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
