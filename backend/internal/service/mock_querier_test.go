package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"mias-backend/internal/database/db"
)

// BaseMockQuerier implementa db.Querier con valores por defecto para pruebas unitarias sin Postgres
type BaseMockQuerier struct {
	AddStampLedgerEntryFunc        func(ctx context.Context, arg db.AddStampLedgerEntryParams) (db.StampLedger, error)
	AssignOrderToUserFunc          func(ctx context.Context, arg db.AssignOrderToUserParams) error
	CountBurgersInOrderFunc        func(ctx context.Context, orderID pgtype.UUID) (int32, error)
	CountUsersFunc                 func(ctx context.Context) (int64, error)
	CreateAuditLogFunc             func(ctx context.Context, arg db.CreateAuditLogParams) (db.AdminAuditLog, error)
	CreateDeliveryReviewFunc       func(ctx context.Context, arg db.CreateDeliveryReviewParams) (db.DeliveryReview, error)
	CreateOAuthUserFunc            func(ctx context.Context, arg db.CreateOAuthUserParams) (db.User, error)
	CreateOrderFunc                func(ctx context.Context, arg db.CreateOrderParams) (db.Order, error)
	CreateOrderItemFunc            func(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error)
	CreateProductFunc              func(ctx context.Context, arg db.CreateProductParams) (db.Product, error)
	CreateRedemptionRequestFunc    func(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error)
	CreateStaffUserFunc            func(ctx context.Context, arg db.CreateStaffUserParams) (db.User, error)
	CreateStampRequestFunc         func(ctx context.Context, arg db.CreateStampRequestParams) (db.StampRequest, error)
	ExpirePendingStampRequestsFunc func(ctx context.Context) error
	GetAverageRatingFunc           func(ctx context.Context) (db.GetAverageRatingRow, error)
	GetOrderByCodeFunc             func(ctx context.Context, code string) (db.Order, error)
	GetOrderByIDFunc               func(ctx context.Context, id pgtype.UUID) (db.Order, error)
	GetOrderItemsFunc              func(ctx context.Context, orderID pgtype.UUID) ([]db.GetOrderItemsRow, error)
	GetOrderReviewFunc             func(ctx context.Context, orderID pgtype.UUID) (db.DeliveryReview, error)
	GetPendingRedemptionByUserFunc func(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error)
	GetProductByIDFunc             func(ctx context.Context, id int32) (db.Product, error)
	GetProductBySlugFunc           func(ctx context.Context, slug string) (db.GetProductBySlugRow, error)
	GetSalesMetricsFunc            func(ctx context.Context) (db.GetSalesMetricsRow, error)
	GetStampRequestByIDFunc        func(ctx context.Context, id pgtype.UUID) (db.GetStampRequestByIDRow, error)
	GetStampStatusByOrderIDFunc    func(ctx context.Context, orderID pgtype.UUID) (db.GetStampStatusByOrderIDRow, error)
	GetTopSellingProductsFunc      func(ctx context.Context, limit int32) ([]db.GetTopSellingProductsRow, error)
	GetUserByEmailFunc             func(ctx context.Context, email string) (db.User, error)
	GetUserByIDFunc                func(ctx context.Context, id pgtype.UUID) (db.User, error)
	GetUserByIDForUpdateFunc       func(ctx context.Context, id pgtype.UUID) (db.User, error)
	GetUserStampBalanceFunc        func(ctx context.Context, userID pgtype.UUID) (int32, error)
	ListActiveCategoriesFunc       func(ctx context.Context) ([]db.Category, error)
	ListActiveDeliveryOrdersFunc   func(ctx context.Context) ([]db.Order, error)
	ListActiveProductsFunc         func(ctx context.Context) ([]db.ListActiveProductsRow, error)
	ListAllCategoriesFunc          func(ctx context.Context) ([]db.Category, error)
	ListAllStampRequestsForStaffFunc func(ctx context.Context) ([]db.ListAllStampRequestsForStaffRow, error)
	ListAuditLogsFunc              func(ctx context.Context, arg db.ListAuditLogsParams) ([]db.ListAuditLogsRow, error)
	ListOrdersByUserIDFunc         func(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error)
	ListPendingInStoreOrdersFunc   func(ctx context.Context) ([]db.Order, error)
	ListPendingRedemptionsFunc     func(ctx context.Context) ([]db.ListPendingRedemptionsRow, error)
	ListPendingStampRequestsFunc   func(ctx context.Context) ([]db.ListPendingStampRequestsRow, error)
	ListProductsByCategoryFunc     func(ctx context.Context, categoryID int32) ([]db.Product, error)
	ListRecentOrdersForStaffFunc   func(ctx context.Context, arg db.ListRecentOrdersForStaffParams) ([]db.Order, error)
	ListRecentReviewsFunc          func(ctx context.Context, arg db.ListRecentReviewsParams) ([]db.ListRecentReviewsRow, error)
	ListUserStampHistoryFunc       func(ctx context.Context, arg db.ListUserStampHistoryParams) ([]db.StampLedger, error)
	ListUsersFunc                  func(ctx context.Context, arg db.ListUsersParams) ([]db.User, error)
	ToggleProductStockFunc         func(ctx context.Context, arg db.ToggleProductStockParams) (db.Product, error)
	UpdateOrderStatusFunc          func(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error)
	UpdateProductFunc              func(ctx context.Context, arg db.UpdateProductParams) (db.Product, error)
	UpdateRedemptionStatusFunc     func(ctx context.Context, arg db.UpdateRedemptionStatusParams) (db.RedemptionRequest, error)
	UpdateStampRequestStatusFunc   func(ctx context.Context, arg db.UpdateStampRequestStatusParams) (db.StampRequest, error)
	UpdateUserLastOrderFunc        func(ctx context.Context, id pgtype.UUID) error
}

func (m *BaseMockQuerier) AddStampLedgerEntry(ctx context.Context, arg db.AddStampLedgerEntryParams) (db.StampLedger, error) {
	if m.AddStampLedgerEntryFunc != nil {
		return m.AddStampLedgerEntryFunc(ctx, arg)
	}
	return db.StampLedger{}, nil
}

func (m *BaseMockQuerier) CountBurgersInOrder(ctx context.Context, orderID pgtype.UUID) (int32, error) {
	if m.CountBurgersInOrderFunc != nil {
		return m.CountBurgersInOrderFunc(ctx, orderID)
	}
	return 0, nil
}

func (m *BaseMockQuerier) CountUsers(ctx context.Context) (int64, error) {
	if m.CountUsersFunc != nil {
		return m.CountUsersFunc(ctx)
	}
	return 0, nil
}

func (m *BaseMockQuerier) CreateAuditLog(ctx context.Context, arg db.CreateAuditLogParams) (db.AdminAuditLog, error) {
	if m.CreateAuditLogFunc != nil {
		return m.CreateAuditLogFunc(ctx, arg)
	}
	return db.AdminAuditLog{}, nil
}

func (m *BaseMockQuerier) CreateDeliveryReview(ctx context.Context, arg db.CreateDeliveryReviewParams) (db.DeliveryReview, error) {
	if m.CreateDeliveryReviewFunc != nil {
		return m.CreateDeliveryReviewFunc(ctx, arg)
	}
	return db.DeliveryReview{}, nil
}

func (m *BaseMockQuerier) CreateOAuthUser(ctx context.Context, arg db.CreateOAuthUserParams) (db.User, error) {
	if m.CreateOAuthUserFunc != nil {
		return m.CreateOAuthUserFunc(ctx, arg)
	}
	return db.User{}, nil
}

func (m *BaseMockQuerier) CreateOrder(ctx context.Context, arg db.CreateOrderParams) (db.Order, error) {
	if m.CreateOrderFunc != nil {
		return m.CreateOrderFunc(ctx, arg)
	}
	return db.Order{}, nil
}

func (m *BaseMockQuerier) CreateOrderItem(ctx context.Context, arg db.CreateOrderItemParams) (db.OrderItem, error) {
	if m.CreateOrderItemFunc != nil {
		return m.CreateOrderItemFunc(ctx, arg)
	}
	return db.OrderItem{}, nil
}

func (m *BaseMockQuerier) CreateProduct(ctx context.Context, arg db.CreateProductParams) (db.Product, error) {
	if m.CreateProductFunc != nil {
		return m.CreateProductFunc(ctx, arg)
	}
	return db.Product{}, nil
}

func (m *BaseMockQuerier) CreateRedemptionRequest(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error) {
	if m.CreateRedemptionRequestFunc != nil {
		return m.CreateRedemptionRequestFunc(ctx, userID)
	}
	return db.RedemptionRequest{}, nil
}

func (m *BaseMockQuerier) CreateStaffUser(ctx context.Context, arg db.CreateStaffUserParams) (db.User, error) {
	if m.CreateStaffUserFunc != nil {
		return m.CreateStaffUserFunc(ctx, arg)
	}
	return db.User{}, nil
}

func (m *BaseMockQuerier) CreateStampRequest(ctx context.Context, arg db.CreateStampRequestParams) (db.StampRequest, error) {
	if m.CreateStampRequestFunc != nil {
		return m.CreateStampRequestFunc(ctx, arg)
	}
	return db.StampRequest{}, nil
}

func (m *BaseMockQuerier) ExpirePendingStampRequests(ctx context.Context) error {
	if m.ExpirePendingStampRequestsFunc != nil {
		return m.ExpirePendingStampRequestsFunc(ctx)
	}
	return nil
}

func (m *BaseMockQuerier) GetAverageRating(ctx context.Context) (db.GetAverageRatingRow, error) {
	if m.GetAverageRatingFunc != nil {
		return m.GetAverageRatingFunc(ctx)
	}
	return db.GetAverageRatingRow{}, nil
}

func (m *BaseMockQuerier) GetOrderByCode(ctx context.Context, code string) (db.Order, error) {
	if m.GetOrderByCodeFunc != nil {
		return m.GetOrderByCodeFunc(ctx, code)
	}
	return db.Order{}, nil
}

func (m *BaseMockQuerier) GetOrderByID(ctx context.Context, id pgtype.UUID) (db.Order, error) {
	if m.GetOrderByIDFunc != nil {
		return m.GetOrderByIDFunc(ctx, id)
	}
	return db.Order{}, nil
}

func (m *BaseMockQuerier) GetOrderItems(ctx context.Context, orderID pgtype.UUID) ([]db.GetOrderItemsRow, error) {
	if m.GetOrderItemsFunc != nil {
		return m.GetOrderItemsFunc(ctx, orderID)
	}
	return nil, nil
}

func (m *BaseMockQuerier) GetOrderReview(ctx context.Context, orderID pgtype.UUID) (db.DeliveryReview, error) {
	if m.GetOrderReviewFunc != nil {
		return m.GetOrderReviewFunc(ctx, orderID)
	}
	return db.DeliveryReview{}, nil
}

func (m *BaseMockQuerier) GetPendingRedemptionByUser(ctx context.Context, userID pgtype.UUID) (db.RedemptionRequest, error) {
	if m.GetPendingRedemptionByUserFunc != nil {
		return m.GetPendingRedemptionByUserFunc(ctx, userID)
	}
	return db.RedemptionRequest{}, nil
}

func (m *BaseMockQuerier) GetProductByID(ctx context.Context, id int32) (db.Product, error) {
	if m.GetProductByIDFunc != nil {
		return m.GetProductByIDFunc(ctx, id)
	}
	return db.Product{}, nil
}

func (m *BaseMockQuerier) GetProductBySlug(ctx context.Context, slug string) (db.GetProductBySlugRow, error) {
	if m.GetProductBySlugFunc != nil {
		return m.GetProductBySlugFunc(ctx, slug)
	}
	return db.GetProductBySlugRow{}, nil
}

func (m *BaseMockQuerier) GetSalesMetrics(ctx context.Context) (db.GetSalesMetricsRow, error) {
	if m.GetSalesMetricsFunc != nil {
		return m.GetSalesMetricsFunc(ctx)
	}
	return db.GetSalesMetricsRow{}, nil
}

func (m *BaseMockQuerier) GetStampRequestByID(ctx context.Context, id pgtype.UUID) (db.GetStampRequestByIDRow, error) {
	if m.GetStampRequestByIDFunc != nil {
		return m.GetStampRequestByIDFunc(ctx, id)
	}
	return db.GetStampRequestByIDRow{}, nil
}

func (m *BaseMockQuerier) GetTopSellingProducts(ctx context.Context, limit int32) ([]db.GetTopSellingProductsRow, error) {
	if m.GetTopSellingProductsFunc != nil {
		return m.GetTopSellingProductsFunc(ctx, limit)
	}
	return nil, nil
}

func (m *BaseMockQuerier) GetUserByEmail(ctx context.Context, email string) (db.User, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return db.User{}, nil
}

func (m *BaseMockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (db.User, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, id)
	}
	return db.User{}, nil
}

func (m *BaseMockQuerier) GetUserByIDForUpdate(ctx context.Context, id pgtype.UUID) (db.User, error) {
	if m.GetUserByIDForUpdateFunc != nil {
		return m.GetUserByIDForUpdateFunc(ctx, id)
	}
	return db.User{}, nil
}

func (m *BaseMockQuerier) GetUserStampBalance(ctx context.Context, userID pgtype.UUID) (int32, error) {
	if m.GetUserStampBalanceFunc != nil {
		return m.GetUserStampBalanceFunc(ctx, userID)
	}
	return 0, nil
}

func (m *BaseMockQuerier) ListActiveCategories(ctx context.Context) ([]db.Category, error) {
	if m.ListActiveCategoriesFunc != nil {
		return m.ListActiveCategoriesFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListActiveDeliveryOrders(ctx context.Context) ([]db.Order, error) {
	if m.ListActiveDeliveryOrdersFunc != nil {
		return m.ListActiveDeliveryOrdersFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListActiveProducts(ctx context.Context) ([]db.ListActiveProductsRow, error) {
	if m.ListActiveProductsFunc != nil {
		return m.ListActiveProductsFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListAllCategories(ctx context.Context) ([]db.Category, error) {
	if m.ListAllCategoriesFunc != nil {
		return m.ListAllCategoriesFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListAuditLogs(ctx context.Context, arg db.ListAuditLogsParams) ([]db.ListAuditLogsRow, error) {
	if m.ListAuditLogsFunc != nil {
		return m.ListAuditLogsFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListOrdersByUserID(ctx context.Context, arg db.ListOrdersByUserIDParams) ([]db.Order, error) {
	if m.ListOrdersByUserIDFunc != nil {
		return m.ListOrdersByUserIDFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListPendingInStoreOrders(ctx context.Context) ([]db.Order, error) {
	if m.ListPendingInStoreOrdersFunc != nil {
		return m.ListPendingInStoreOrdersFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListPendingRedemptions(ctx context.Context) ([]db.ListPendingRedemptionsRow, error) {
	if m.ListPendingRedemptionsFunc != nil {
		return m.ListPendingRedemptionsFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListPendingStampRequests(ctx context.Context) ([]db.ListPendingStampRequestsRow, error) {
	if m.ListPendingStampRequestsFunc != nil {
		return m.ListPendingStampRequestsFunc(ctx)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListProductsByCategory(ctx context.Context, categoryID int32) ([]db.Product, error) {
	if m.ListProductsByCategoryFunc != nil {
		return m.ListProductsByCategoryFunc(ctx, categoryID)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListRecentOrdersForStaff(ctx context.Context, arg db.ListRecentOrdersForStaffParams) ([]db.Order, error) {
	if m.ListRecentOrdersForStaffFunc != nil {
		return m.ListRecentOrdersForStaffFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListRecentReviews(ctx context.Context, arg db.ListRecentReviewsParams) ([]db.ListRecentReviewsRow, error) {
	if m.ListRecentReviewsFunc != nil {
		return m.ListRecentReviewsFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListUserStampHistory(ctx context.Context, arg db.ListUserStampHistoryParams) ([]db.StampLedger, error) {
	if m.ListUserStampHistoryFunc != nil {
		return m.ListUserStampHistoryFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ListUsers(ctx context.Context, arg db.ListUsersParams) ([]db.User, error) {
	if m.ListUsersFunc != nil {
		return m.ListUsersFunc(ctx, arg)
	}
	return nil, nil
}

func (m *BaseMockQuerier) ToggleProductStock(ctx context.Context, arg db.ToggleProductStockParams) (db.Product, error) {
	if m.ToggleProductStockFunc != nil {
		return m.ToggleProductStockFunc(ctx, arg)
	}
	return db.Product{}, nil
}

func (m *BaseMockQuerier) UpdateOrderStatus(ctx context.Context, arg db.UpdateOrderStatusParams) (db.Order, error) {
	if m.UpdateOrderStatusFunc != nil {
		return m.UpdateOrderStatusFunc(ctx, arg)
	}
	return db.Order{}, nil
}

func (m *BaseMockQuerier) UpdateProduct(ctx context.Context, arg db.UpdateProductParams) (db.Product, error) {
	if m.UpdateProductFunc != nil {
		return m.UpdateProductFunc(ctx, arg)
	}
	return db.Product{}, nil
}

func (m *BaseMockQuerier) UpdateRedemptionStatus(ctx context.Context, arg db.UpdateRedemptionStatusParams) (db.RedemptionRequest, error) {
	if m.UpdateRedemptionStatusFunc != nil {
		return m.UpdateRedemptionStatusFunc(ctx, arg)
	}
	return db.RedemptionRequest{}, nil
}

func (m *BaseMockQuerier) UpdateStampRequestStatus(ctx context.Context, arg db.UpdateStampRequestStatusParams) (db.StampRequest, error) {
	if m.UpdateStampRequestStatusFunc != nil {
		return m.UpdateStampRequestStatusFunc(ctx, arg)
	}
	return db.StampRequest{}, nil
}

func (m *BaseMockQuerier) UpdateUserLastOrder(ctx context.Context, id pgtype.UUID) error {
	if m.UpdateUserLastOrderFunc != nil {
		return m.UpdateUserLastOrderFunc(ctx, id)
	}
	return nil
}

func (m *BaseMockQuerier) AssignOrderToUser(ctx context.Context, arg db.AssignOrderToUserParams) error {
	if m.AssignOrderToUserFunc != nil {
		return m.AssignOrderToUserFunc(ctx, arg)
	}
	return nil
}

func (m *BaseMockQuerier) GetStampStatusByOrderID(ctx context.Context, orderID pgtype.UUID) (db.GetStampStatusByOrderIDRow, error) {
	if m.GetStampStatusByOrderIDFunc != nil {
		return m.GetStampStatusByOrderIDFunc(ctx, orderID)
	}
	return db.GetStampStatusByOrderIDRow{}, nil
}

func (m *BaseMockQuerier) ListAllStampRequestsForStaff(ctx context.Context) ([]db.ListAllStampRequestsForStaffRow, error) {
	if m.ListAllStampRequestsForStaffFunc != nil {
		return m.ListAllStampRequestsForStaffFunc(ctx)
	}
	return nil, nil
}

var _ db.Querier = (*BaseMockQuerier)(nil)
