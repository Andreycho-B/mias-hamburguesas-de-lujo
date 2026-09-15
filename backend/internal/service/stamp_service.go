package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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
	ErrSelfApproval        = errors.New("un empleado no puede auto-aprobar sus propias solicitudes de sellos o redenciones")
	ErrStampLimitReached   = errors.New("tu cartilla ya tiene 10 sellos completos; redime tu hamburguesa gratis antes de acumular nuevos sellos")
	ErrRequestNotPending   = errors.New("la solicitud ya no se encuentra pendiente de revisión")
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

// RequestStamp registra la solicitud de sellos admitiendo orderID o resolución por código de factura
func (s *StampService) RequestStamp(ctx context.Context, userID, orderID uuid.UUID, receiptNumber string) (*db.StampRequest, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	var order db.Order
	var err error
	if orderID != uuid.Nil {
		order, err = s.queries.GetOrderByID(ctx, uuidutil.ToPG(orderID))
	} else if receiptNumber != "" {
		order, err = s.queries.GetOrderByCode(ctx, receiptNumber)
	} else {
		return nil, errors.New("se requiere ID de orden o código de factura")
	}

	if err != nil {
		return nil, fmt.Errorf("no se encontró una orden válida: %w", err)
	}

	resolvedOrderID := orderID
	if resolvedOrderID == uuid.Nil && order.ID.Valid {
		resolvedOrderID = uuidutil.FromPG(order.ID)
	}

	// Validar estado elegible: ORDEN PAGA (in-store) o ENTREGADA (delivery)
	if order.Status != "PAID" && order.Status != "DELIVERED" {
		return nil, ErrOrderNotEligible
	}

	// Validar ventana de 24 horas
	if order.CreatedAt.Valid && time.Since(order.CreatedAt.Time) > 24*time.Hour {
		return nil, ErrOrderExpired
	}

	// Validar que la orden pertenezca al usuario si ya tiene uno asignado
	if order.UserID.Valid {
		existingUserID := uuidutil.FromPG(order.UserID)
		if existingUserID != uuid.Nil && existingUserID != userID {
			return nil, errors.New("esta orden pertenece a otra cuenta de usuario")
		}
	}

	// Validar regla de congelación: si ya tiene 10 sellos, debe redimir antes de acumular más
	currentBal, err := s.queries.GetUserStampBalance(ctx, uuidutil.ToPG(userID))
	if err == nil && currentBal >= 10 {
		return nil, ErrStampLimitReached
	}

	// Contar hamburguesas calificadas
	burgersCount, err := s.queries.CountBurgersInOrder(ctx, uuidutil.ToPG(resolvedOrderID))
	if err != nil || burgersCount == 0 {
		return nil, ErrNoBurgersInOrder
	}

	request, err := s.queries.CreateStampRequest(ctx, db.CreateStampRequestParams{
		OrderID:       uuidutil.ToPG(resolvedOrderID),
		UserID:        uuidutil.ToPG(userID),
		ReceiptNumber: receiptNumber,
		BurgersCount:  burgersCount,
	})
	if err != nil {
		return nil, fmt.Errorf("la factura #%s ya fue registrada o se encuentra en revisión", receiptNumber)
	}

	if err := s.queries.AssignOrderToUser(ctx, db.AssignOrderToUserParams{
		UserID: uuidutil.ToPG(userID),
		ID:     uuidutil.ToPG(resolvedOrderID),
	}); err != nil {
		slog.Warn("no se pudo asociar orden huérfana al usuario", "order_id", resolvedOrderID, "user_id", userID, "error", err)
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

	// 1. Validar existencia y estado pendiente previo
	existingReq, err := s.queries.GetStampRequestByID(ctx, uuidutil.ToPG(requestID))
	if err != nil {
		return fmt.Errorf("solicitud de sellos no encontrada: %w", err)
	}
	if existingReq.Status != "PENDING" {
		return ErrRequestNotPending
	}

	// 2. Segregación de funciones (SEC-001/SEC-003): Prohibir auto-aprobación
	if reviewerID == uuidutil.FromPG(existingReq.UserID) {
		return ErrSelfApproval
	}

	if s.pool == nil {
		return errors.New("servicio de transacciones no disponible")
	}

	status := "REJECTED"
	if approve {
		status = "APPROVED"
	}

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

		// Validar regla de congelación: tope estricto de 10 sellos (PRD FR-006)
		currentBal, err := qtx.GetUserStampBalance(ctx, request.UserID)
		if err != nil {
			return fmt.Errorf("error consultando balance de sellos: %w", err)
		}

		availableCapacity := int32(10) - currentBal
		if availableCapacity <= 0 {
			return errors.New("el cliente ya alcanzó el tope de 10 sellos; debe redimir su hamburguesa gratis pendiente antes de acumular nuevos sellos")
		}

		stampsToAward := request.BurgersCount
		if stampsToAward > availableCapacity {
			stampsToAward = availableCapacity
		}

		_, err = qtx.AddStampLedgerEntry(ctx, db.AddStampLedgerEntryParams{
			UserID:  request.UserID,
			OrderID: request.OrderID,
			Amount:  stampsToAward,
			Reason:  fmt.Sprintf("Factura #%s aprobada (%d sellos)", request.ReceiptNumber, stampsToAward),
		})
		if err != nil {
			return err
		}
	}

	var orderCode string
	if ord, err := qtx.GetOrderByID(ctx, request.OrderID); err == nil {
		orderCode = ord.Code
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("error confirmando transacción de sellos: %w", err)
	}

	if s.hub != nil {
		var newBal int32
		if b, err := s.queries.GetUserStampBalance(ctx, request.UserID); err == nil {
			newBal = b
		}
		s.hub.Broadcast("stamp_request_reviewed", map[string]interface{}{
			"id":          uuidutil.FromPG(request.ID),
			"order_id":    uuidutil.FromPG(request.OrderID),
			"order_code":  orderCode,
			"status":      request.Status,
			"user_id":     uuidutil.FromPG(request.UserID),
			"new_balance": newBal,
		})
	}

	return nil
}

// RequestRedemption solicita la hamburguesa Original gratis al alcanzar 10 sellos
func (s *StampService) RequestRedemption(ctx context.Context, userID uuid.UUID) (*db.RedemptionRequest, error) {
	if s.queries == nil {
		return nil, errors.New("servicio de base de datos no disponible")
	}

	// 1. Validar si ya tiene una redención PENDING
	existing, err := s.queries.GetPendingRedemptionByUser(ctx, uuidutil.ToPG(userID))
	if err == nil && existing.ID.Valid {
		return nil, ErrPendingRedemption
	}

	// 2. Validar que tenga al menos 10 sellos
	balance, err := s.queries.GetUserStampBalance(ctx, uuidutil.ToPG(userID))
	if err != nil {
		return nil, err
	}
	if balance < 10 {
		return nil, ErrInsufficientStamps
	}

	redemption, err := s.queries.CreateRedemptionRequest(ctx, uuidutil.ToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("error creando solicitud de redención: %w", err)
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
	if s.pool == nil || s.queries == nil {
		return errors.New("servicio de transacciones no disponible")
	}

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

	// Segregación de funciones (SEC-001/SEC-003): Prohibir auto-aprobación
	if reviewerID == uuidutil.FromPG(redemption.UserID) {
		return ErrSelfApproval
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
