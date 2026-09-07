package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/uuidutil"
)

var (
	ErrOrderNotEligible    = errors.New("la orden debe estar en estado ORDEN PAGA o ENTREGADA para solicitar sellos")
	ErrOrderExpired        = errors.New("el plazo de 24 horas para registrar la factura ha expirado")
	ErrNoBurgersInOrder    = errors.New("la orden no contiene hamburguesas calificadas para sellos")
	ErrInsufficientStamps  = errors.New("se requieren 10 sellos activos para redimir una hamburguesa gratis")
	ErrPendingRedemption   = errors.New("ya tienes una solicitud de redención pendiente de aprobación")
	ErrPendingStampRequest = errors.New("ya tienes una solicitud pendiente para esta factura")
)

type StampService struct {
	pool    *pgxpool.Pool
	queries db.Querier
	hub     *SSEHub
}

func NewStampService(pool *pgxpool.Pool, queries db.Querier, hub *SSEHub) *StampService {
	return &StampService{
		pool:    pool,
		queries: queries,
		hub:     hub,
	}
}

type StampBalanceResponse struct {
	ActiveStamps     int32 `json:"active_stamps"`
	MaxGoal          int32 `json:"max_goal"` // 10
	CanRedeem        bool  `json:"can_redeem"`
	HasPendingRedeem bool  `json:"has_pending_redemption"`
}

func (s *StampService) GetBalance(ctx context.Context, userID uuid.UUID) (*StampBalanceResponse, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	balance, err := s.queries.GetUserStampBalance(ctx, uuidutil.ToPG(userID))
	if err != nil {
		return nil, err
	}

	pending, err := s.queries.GetPendingRedemptionByUser(ctx, uuidutil.ToPG(userID))
	hasPending := err == nil && pending.Status == "PENDING"

	return &StampBalanceResponse{
		ActiveStamps:     balance,
		MaxGoal:          10,
		CanRedeem:        balance >= 10 && !hasPending,
		HasPendingRedeem: hasPending,
	}, nil
}

// RequestStamp registra la solicitud de sellos usando el número de la factura física
func (s *StampService) RequestStamp(ctx context.Context, userID, orderID uuid.UUID, receiptNumber string) (*db.StampRequest, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	order, err := s.queries.GetOrderByID(ctx, uuidutil.ToPG(orderID))
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.Status != "PAID" && order.Status != "DELIVERED" {
		return nil, ErrOrderNotEligible
	}

	// Blindaje: Si la orden fue creada por un usuario específico, solo él puede solicitar los sellos
	if order.UserID.Valid && uuidutil.FromPG(order.UserID) != userID {
		return nil, errors.New("la orden no pertenece a la cuenta de usuario autenticada")
	}

	// Comprobar ventana de 24 horas
	if time.Since(order.CreatedAt.Time) > 24*time.Hour {
		return nil, ErrOrderExpired
	}

	// Contar hamburguesas calificadas
	burgersCount, err := s.queries.CountBurgersInOrder(ctx, uuidutil.ToPG(orderID))
	if err != nil || burgersCount == 0 {
		return nil, ErrNoBurgersInOrder
	}

	request, err := s.queries.CreateStampRequest(ctx, db.CreateStampRequestParams{
		OrderID:       uuidutil.ToPG(orderID),
		UserID:        uuidutil.ToPG(userID),
		ReceiptNumber: receiptNumber,
		BurgersCount:  burgersCount,
	})
	if err != nil {
		return nil, fmt.Errorf("la factura #%s ya fue registrada o se encuentra en revisión", receiptNumber)
	}

	// Notificar en tiempo real a Meseros y Cajero
	if s.hub != nil {
		s.hub.Broadcast("new_stamp_request", map[string]interface{}{
			"id":             uuidutil.FromPG(request.ID),
			"order_code":     order.Code,
			"receipt_number": request.ReceiptNumber,
			"burgers_count":  request.BurgersCount,
			"user_id":        uuidutil.FromPG(request.UserID),
			"created_at":     request.CreatedAt,
		})
	}

	return &request, nil
}

// ReviewStampRequest permite a un Mesero o Cajero aprobar o rechazar la solicitud de forma atómica
func (s *StampService) ReviewStampRequest(ctx context.Context, reviewerID, requestID uuid.UUID, approve bool) error {
	if s.queries == nil {
		return errors.New("servicio de base de datos no disponible")
	}

	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}

	// Ejecución atómica en PostgreSQL con bloqueo pesimista
	if s.pool != nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("error iniciando transacción: %w", err)
		}
		defer tx.Rollback(ctx)

		qtx := db.New(tx)

		request, err := qtx.UpdateStampRequestStatus(ctx, db.UpdateStampRequestStatusParams{
			ID:         uuidutil.ToPG(requestID),
			Status:     status,
			ReviewedBy: uuidutil.ToPG(reviewerID),
		})
		if err != nil {
			return err
		}

		if approve {
			// Bloqueo pesimista del usuario para garantizar consistencia atómica del libro mayor
			_, err = qtx.GetUserByIDForUpdate(ctx, request.UserID)
			if err != nil {
				return fmt.Errorf("error adquiriendo bloqueo de usuario: %w", err)
			}

			_, err = qtx.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
				UserID:  request.UserID,
				OrderID: request.OrderID,
				Amount:  request.BurgersCount,
				Reason:  fmt.Sprintf("Factura #%s aprobada", request.ReceiptNumber),
			})
			if err != nil {
				return err
			}
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("error confirmando transacción de sellos: %w", err)
		}

		if s.hub != nil {
			s.hub.Broadcast("stamp_request_reviewed", map[string]interface{}{
				"id":      uuidutil.FromPG(request.ID),
				"status":  request.Status,
				"user_id": uuidutil.FromPG(request.UserID),
			})
		}

		return nil
	}

	// Modo degradado / tests unitarios con Querier mock
	request, err := s.queries.UpdateStampRequestStatus(ctx, db.UpdateStampRequestStatusParams{
		ID:         uuidutil.ToPG(requestID),
		Status:     status,
		ReviewedBy: uuidutil.ToPG(reviewerID),
	})
	if err != nil {
		return err
	}

	if approve {
		_, err = s.queries.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
			UserID:  request.UserID,
			OrderID: request.OrderID,
			Amount:  request.BurgersCount,
			Reason:  fmt.Sprintf("Factura #%s aprobada", request.ReceiptNumber),
		})
		if err != nil {
			return err
		}
	}

	if s.hub != nil {
		s.hub.Broadcast("stamp_request_reviewed", map[string]interface{}{
			"id":      uuidutil.FromPG(request.ID),
			"status":  request.Status,
			"user_id": uuidutil.FromPG(request.UserID),
		})
	}

	return nil
}

// RequestRedemption solicita la hamburguesa Original gratis al alcanzar 10 sellos
func (s *StampService) RequestRedemption(ctx context.Context, userID uuid.UUID) (*db.RedemptionRequest, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	balance, err := s.queries.GetUserStampBalance(ctx, uuidutil.ToPG(userID))
	if err != nil || balance < 10 {
		return nil, ErrInsufficientStamps
	}

	existing, err := s.queries.GetPendingRedemptionByUser(ctx, uuidutil.ToPG(userID))
	if err == nil && existing.Status == "PENDING" {
		return nil, ErrPendingRedemption
	}

	redemption, err := s.queries.CreateRedemptionRequest(ctx, uuidutil.ToPG(userID))
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Broadcast("new_redemption_request", map[string]interface{}{
			"id":         uuidutil.FromPG(redemption.ID),
			"user_id":    uuidutil.FromPG(redemption.UserID),
			"created_at": redemption.CreatedAt,
		})
	}

	return &redemption, nil
}

// ApproveRedemption permite al cajero validar la redención presencial de forma atómica y pesimista
func (s *StampService) ApproveRedemption(ctx context.Context, reviewerID, redemptionID uuid.UUID) error {
	if s.queries == nil {
		return errors.New("servicio de base de datos no disponible")
	}

	// Ejecución atómica en PostgreSQL con bloqueo pesimista
	if s.pool != nil {
		tx, err := s.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("error iniciando transacción: %w", err)
		}
		defer tx.Rollback(ctx)

		qtx := db.New(tx)

		redemption, err := qtx.UpdateRedemptionStatus(ctx, db.UpdateRedemptionStatusParams{
			ID:         uuidutil.ToPG(redemptionID),
			Status:     "APPROVED",
			ReviewedBy: uuidutil.ToPG(reviewerID),
		})
		if err != nil {
			return err
		}

		// Bloqueo pesimista del usuario para serializar balance
		_, err = qtx.GetUserByIDForUpdate(ctx, redemption.UserID)
		if err != nil {
			return fmt.Errorf("error adquiriendo bloqueo de usuario: %w", err)
		}

		// Re-verificar balance bajo el bloqueo pesimista
		balance, err := qtx.GetUserStampBalance(ctx, redemption.UserID)
		if err != nil {
			return err
		}
		if balance < 10 {
			return ErrInsufficientStamps
		}

		// Descontar 10 sellos del libro mayor inmutable
		_, err = qtx.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
			UserID:  redemption.UserID,
			OrderID: pgtype.UUID{Valid: false},
			Amount:  -10,
			Reason:  "Redención presencial de Hamburguesa Original",
		})
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("error confirmando transacción de redención: %w", err)
		}

		if s.hub != nil {
			s.hub.Broadcast("redemption_approved", map[string]interface{}{
				"id":      uuidutil.FromPG(redemption.ID),
				"user_id": uuidutil.FromPG(redemption.UserID),
				"status":  "APPROVED",
			})
		}

		return nil
	}

	// Modo degradado / tests unitarios con Querier mock
	redemption, err := s.queries.UpdateRedemptionStatus(ctx, db.UpdateRedemptionStatusParams{
		ID:         uuidutil.ToPG(redemptionID),
		Status:     "APPROVED",
		ReviewedBy: uuidutil.ToPG(reviewerID),
	})
	if err != nil {
		return err
	}

	_, err = s.queries.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
		UserID:  redemption.UserID,
		OrderID: pgtype.UUID{Valid: false},
		Amount:  -10,
		Reason:  "Redención presencial de Hamburguesa Original",
	})
	if err != nil {
		return err
	}

	if s.hub != nil {
		s.hub.Broadcast("redemption_approved", map[string]interface{}{
			"id":      uuidutil.FromPG(redemption.ID),
			"user_id": uuidutil.FromPG(redemption.UserID),
			"status":  "APPROVED",
		})
	}

	return nil
}
