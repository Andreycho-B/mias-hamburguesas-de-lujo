package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
	"mias-backend/internal/pkg/uuidutil"
)

var (
	ErrOrderNotEligible     = errors.New("la orden debe estar en estado ORDEN PAGA para solicitar sellos")
	ErrOrderExpired         = errors.New("el plazo de 24 horas para registrar la factura ha expirado")
	ErrNoBurgersInOrder     = errors.New("la orden no contiene hamburguesas calificadas para sellos")
	ErrInsufficientStamps   = errors.New("se requieren 10 sellos activos para redimir una hamburguesa gratis")
	ErrPendingRedemption    = errors.New("ya tienes una solicitud de redención pendiente de aprobación")
	ErrPendingStampRequest  = errors.New("ya tienes una solicitud pendiente para esta factura")
)

type StampService struct {
	queries db.Querier
	hub     *SSEHub
}

func NewStampService(queries db.Querier, hub *SSEHub) *StampService {
	return &StampService{
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
	order, err := s.queries.GetOrderByID(ctx, uuidutil.ToPG(orderID))
	if err != nil {
		return nil, ErrOrderNotFound
	}

	if order.Status != "PAID" {
		return nil, ErrOrderNotEligible
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
	s.hub.Broadcast("new_stamp_request", map[string]interface{}{
		"id":             uuidutil.FromPG(request.ID),
		"order_code":     order.Code,
		"receipt_number": request.ReceiptNumber,
		"burgers_count":  request.BurgersCount,
		"user_id":        uuidutil.FromPG(request.UserID),
		"created_at":     request.CreatedAt,
	})

	return &request, nil
}

// ReviewStampRequest permite a un Mesero o Cajero aprobar o rechazar la solicitud
func (s *StampService) ReviewStampRequest(ctx context.Context, reviewerID, requestID uuid.UUID, approve bool) error {
	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}

	request, err := s.queries.UpdateStampRequestStatus(ctx, db.UpdateStampRequestStatusParams{
		ID:         uuidutil.ToPG(requestID),
		Status:     status,
		ReviewedBy: uuidutil.ToPG(reviewerID),
	})
	if err != nil {
		return err
	}

	// Si se aprueba, registrar los sellos en el libro mayor inmutable (stamp_ledger)
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

	// Notificar actualización al cliente y personal
	s.hub.Broadcast("stamp_request_reviewed", map[string]interface{}{
		"id":      uuidutil.FromPG(request.ID),
		"status":  request.Status,
		"user_id": uuidutil.FromPG(request.UserID),
	})

	return nil
}

// RequestRedemption solicita la hamburguesa Original gratis al alcanzar 10 sellos
func (s *StampService) RequestRedemption(ctx context.Context, userID uuid.UUID) (*db.RedemptionRequest, error) {
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

	s.hub.Broadcast("new_redemption_request", map[string]interface{}{
		"id":         uuidutil.FromPG(redemption.ID),
		"user_id":    uuidutil.FromPG(redemption.UserID),
		"created_at": redemption.CreatedAt,
	})

	return &redemption, nil
}

// ApproveRedemption permite al cajero validar la redención presencial en 1 clic
func (s *StampService) ApproveRedemption(ctx context.Context, reviewerID, redemptionID uuid.UUID) error {
	redemption, err := s.queries.UpdateRedemptionStatus(ctx, db.UpdateRedemptionStatusParams{
		ID:         uuidutil.ToPG(redemptionID),
		Status:     "APPROVED",
		ReviewedBy: uuidutil.ToPG(reviewerID),
	})
	if err != nil {
		return err
	}

	// Descontar 10 sellos del libro mayor inmutable
	_, err = s.queries.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
		UserID:  redemption.UserID,
		OrderID: pgtype.UUID{Valid: false},
		Amount:  -10,
		Reason:  "Redención presencial de Hamburguesa Original",
	})
	if err != nil {
		return err
	}

	// Disparar evento de celebración al cliente
	s.hub.Broadcast("redemption_approved", map[string]interface{}{
		"id":      uuidutil.FromPG(redemption.ID),
		"user_id": uuidutil.FromPG(redemption.UserID),
		"status":  "APPROVED",
	})

	return nil
}
