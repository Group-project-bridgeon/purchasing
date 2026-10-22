package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/db/generated"
	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresPurchaseOrderRepository struct {
	queries *generated.Queries
}

var _ PurchaseOrderRepository = (*postgresPurchaseOrderRepository)(nil)

func NewPostgresPurchaseOrderRepository(
	db *pgxpool.Pool,
) PurchaseOrderRepository {
	return &postgresPurchaseOrderRepository{
		queries: generated.New(db),
	}
}

func NewPostgresPurchaseOrderRepositoryWithTx(
	tx pgx.Tx,
) PurchaseOrderRepository {
	return &postgresPurchaseOrderRepository{
		queries: generated.New(tx),
	}
}

func mapPurchaseOrderToDomain(
	po generated.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	subtotal, err := numericToDecimal(po.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("convert subtotal: %w", err)
	}

	taxAmount, err := numericToDecimal(po.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("convert tax amount: %w", err)
	}

	totalAmount, err := numericToDecimal(po.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("convert total amount: %w", err)
	}

	result := &domain.PurchaseOrder{
		ID:          po.ID.String(),
		PONumber:    po.PoNumber,
		VendorID:    po.VendorID.String(),
		Status:      domain.PurchaseOrderStatus(po.Status),
		OrderDate:   po.OrderDate.Time,
		Currency:    po.Currency,
		Subtotal:    subtotal,
		TaxAmount:   taxAmount,
		TotalAmount: totalAmount,
		CreatedAt:   po.CreatedAt.Time,
		UpdatedAt:   po.UpdatedAt.Time,
	}

	if po.PurchaseRequestID.Valid {
		prID := po.PurchaseRequestID.String()
		result.PurchaseRequestID = &prID
	}

	if po.ExpectedDeliveryDate.Valid {
		d := po.ExpectedDeliveryDate.Time
		result.ExpectedDeliveryDate = &d
	}

	if po.PaymentTerms.Valid {
		pt := po.PaymentTerms.String
		result.PaymentTerms = &pt
	}

	if po.ShippingAddress.Valid {
		sa := po.ShippingAddress.String
		result.ShippingAddress = &sa
	}

	if po.Notes.Valid {
		notes := po.Notes.String
		result.Notes = &notes
	}

	return result, nil
}

func (r *postgresPurchaseOrderRepository) CreatePurchaseOrder(
	ctx context.Context,
	order *domain.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	vendorID := pgtype.UUID{}
	if err := vendorID.Scan(order.VendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	prID := pgtype.UUID{}
	if order.PurchaseRequestID != nil && *order.PurchaseRequestID != "" {
		if err := prID.Scan(*order.PurchaseRequestID); err != nil {
			return nil, fmt.Errorf("invalid purchase request id: %w", err)
		}
	}

	var expectedDeliveryDate pgtype.Date
	if order.ExpectedDeliveryDate != nil {
		expectedDeliveryDate = pgtype.Date{
			Time:  *order.ExpectedDeliveryDate,
			Valid: true,
		}
	}

	subtotal, err := decimalToNumeric(order.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("invalid subtotal: %w", err)
	}

	taxAmount, err := decimalToNumeric(order.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid tax amount: %w", err)
	}

	totalAmount, err := decimalToNumeric(order.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	var paymentTerms pgtype.Text
	if order.PaymentTerms != nil {
		paymentTerms = pgtype.Text{String: *order.PaymentTerms, Valid: true}
	}

	var shippingAddress pgtype.Text
	if order.ShippingAddress != nil {
		shippingAddress = pgtype.Text{String: *order.ShippingAddress, Valid: true}
	}

	var notes pgtype.Text
	if order.Notes != nil {
		notes = pgtype.Text{String: *order.Notes, Valid: true}
	}

	currency := order.Currency
	if currency == "" {
		currency = "USD"
	}

	created, err := r.queries.CreatePurchaseOrder(
		ctx,
		generated.CreatePurchaseOrderParams{
			PoNumber:             order.PONumber,
			PurchaseRequestID:    prID,
			VendorID:             vendorID,
			Column4:              pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ExpectedDeliveryDate: expectedDeliveryDate,
			Currency:             currency,
			Subtotal:             subtotal,
			TaxAmount:            taxAmount,
			TotalAmount:          totalAmount,
			PaymentTerms:         paymentTerms,
			ShippingAddress:      shippingAddress,
			Notes:                notes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create purchase order: %w", err)
	}

	return mapPurchaseOrderToDomain(created)
}

func (r *postgresPurchaseOrderRepository) GetPurchaseOrderByID(
	ctx context.Context,
	id string,
) (*domain.PurchaseOrder, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	po, err := r.queries.GetPurchaseOrderByID(ctx, poID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found: %w", err)
		}
		return nil, fmt.Errorf("get purchase order by id: %w", err)
	}

	return mapPurchaseOrderToDomain(po)
}

func (r *postgresPurchaseOrderRepository) GetPurchaseOrderByNumber(
	ctx context.Context,
	poNumber string,
) (*domain.PurchaseOrder, error) {
	po, err := r.queries.GetPurchaseOrderByNumber(ctx, poNumber)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found: %w", err)
		}
		return nil, fmt.Errorf("get purchase order by number: %w", err)
	}

	return mapPurchaseOrderToDomain(po)
}

func (r *postgresPurchaseOrderRepository) ListPurchaseOrders(
	ctx context.Context,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	orders, err := r.queries.ListPurchaseOrders(
		ctx,
		generated.ListPurchaseOrdersParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase orders: %w", err)
	}

	result := make([]*domain.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		mapped, err := mapPurchaseOrderToDomain(po)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseOrderRepository) ListPurchaseOrdersByStatus(
	ctx context.Context,
	status domain.PurchaseOrderStatus,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	orders, err := r.queries.ListPurchaseOrdersByStatus(
		ctx,
		generated.ListPurchaseOrdersByStatusParams{
			Status: string(status),
			Limit:  int32(limit),
			Offset: int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase orders by status: %w", err)
	}

	result := make([]*domain.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		mapped, err := mapPurchaseOrderToDomain(po)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseOrderRepository) ListPurchaseOrdersByVendor(
	ctx context.Context,
	vendorID string,
	limit int,
	offset int,
) ([]*domain.PurchaseOrder, error) {
	vID := pgtype.UUID{}
	if err := vID.Scan(vendorID); err != nil {
		return nil, fmt.Errorf("invalid vendor id: %w", err)
	}

	orders, err := r.queries.ListPurchaseOrdersByVendor(
		ctx,
		generated.ListPurchaseOrdersByVendorParams{
			VendorID: vID,
			Limit:    int32(limit),
			Offset:   int32(offset),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list purchase orders by vendor: %w", err)
	}

	result := make([]*domain.PurchaseOrder, 0, len(orders))
	for _, po := range orders {
		mapped, err := mapPurchaseOrderToDomain(po)
		if err != nil {
			return nil, err
		}
		result = append(result, mapped)
	}

	return result, nil
}

func (r *postgresPurchaseOrderRepository) UpdatePurchaseOrder(
	ctx context.Context,
	order *domain.PurchaseOrder,
) (*domain.PurchaseOrder, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(order.ID); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	var expectedDeliveryDate pgtype.Date
	if order.ExpectedDeliveryDate != nil {
		expectedDeliveryDate = pgtype.Date{
			Time:  *order.ExpectedDeliveryDate,
			Valid: true,
		}
	}

	subtotal, err := decimalToNumeric(order.Subtotal)
	if err != nil {
		return nil, fmt.Errorf("invalid subtotal: %w", err)
	}

	taxAmount, err := decimalToNumeric(order.TaxAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid tax amount: %w", err)
	}

	totalAmount, err := decimalToNumeric(order.TotalAmount)
	if err != nil {
		return nil, fmt.Errorf("invalid total amount: %w", err)
	}

	var paymentTerms pgtype.Text
	if order.PaymentTerms != nil {
		paymentTerms = pgtype.Text{String: *order.PaymentTerms, Valid: true}
	}

	var shippingAddress pgtype.Text
	if order.ShippingAddress != nil {
		shippingAddress = pgtype.Text{String: *order.ShippingAddress, Valid: true}
	}

	var notes pgtype.Text
	if order.Notes != nil {
		notes = pgtype.Text{String: *order.Notes, Valid: true}
	}

	updated, err := r.queries.UpdatePurchaseOrder(
		ctx,
		generated.UpdatePurchaseOrderParams{
			ID:                   poID,
			ExpectedDeliveryDate: expectedDeliveryDate,
			Currency:             order.Currency,
			Subtotal:             subtotal,
			TaxAmount:            taxAmount,
			TotalAmount:          totalAmount,
			PaymentTerms:         paymentTerms,
			ShippingAddress:      shippingAddress,
			Notes:                notes,
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found or not editable: %w", err)
		}
		return nil, fmt.Errorf("update purchase order: %w", err)
	}

	return mapPurchaseOrderToDomain(updated)
}

func (r *postgresPurchaseOrderRepository) UpdatePurchaseOrderStatus(
	ctx context.Context,
	id string,
	status domain.PurchaseOrderStatus,
) (*domain.PurchaseOrder, error) {
	poID := pgtype.UUID{}
	if err := poID.Scan(id); err != nil {
		return nil, fmt.Errorf("invalid purchase order id: %w", err)
	}

	updated, err := r.queries.UpdatePurchaseOrderStatus(
		ctx,
		generated.UpdatePurchaseOrderStatusParams{
			ID:     poID,
			Status: string(status),
		},
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("purchase order not found: %w", err)
		}
		return nil, fmt.Errorf("update purchase order status: %w", err)
	}

	return mapPurchaseOrderToDomain(updated)
}
