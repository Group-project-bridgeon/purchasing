package helper

import (
	"fmt"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/shopspring/decimal"
)

func ParseUpdatePurchaseRequestItem(
	body *dto.UpdatePurchaseRequestItemBody,
) (*domain.PurchaseRequestItem, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	quantity, err := decimal.NewFromString(body.Quantity)
	if err != nil {
		return nil, fmt.Errorf("invalid quantity: %w", err)
	}

	var estimatedUnitPrice *decimal.Decimal

	if body.EstimatedUnitPrice != nil && *body.EstimatedUnitPrice != "" {
		price, err := decimal.NewFromString(*body.EstimatedUnitPrice)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid estimated_unit_price: %w",
				err,
			)
		}

		estimatedUnitPrice = &price
	}

	return &domain.PurchaseRequestItem{
		ItemID:             body.ItemID,
		Description:        body.Description,
		Quantity:           quantity,
		EstimatedUnitPrice: estimatedUnitPrice,
	}, nil
}