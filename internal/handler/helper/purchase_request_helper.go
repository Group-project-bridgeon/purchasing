package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/shopspring/decimal"
)

func ParseCreatePurchaseRequest(
	body *dto.CreatePurchaseRequestBody,
) (*service.CreatePurchaseRequestInput, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var requiredBy *time.Time

	if body.RequiredBy != nil && *body.RequiredBy != "" {
		parsedDate, err := time.Parse(
			"2006-01-02",
			*body.RequiredBy,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid required_by date: %w",
				err,
			)
		}

		requiredBy = &parsedDate
	}

	request := &domain.PurchaseRequest{
		RequestedBy:  body.RequestedBy,
		DepartmentID: body.DepartmentID,
		RequiredBy:   requiredBy,
		Purpose:      body.Purpose,
		Notes:        body.Notes,
	}

	items := make(
		[]*domain.PurchaseRequestItem,
		0,
		len(body.Items),
	)

	for i, itemBody := range body.Items {
		quantity, err := decimal.NewFromString(itemBody.Quantity)
		if err != nil {
			return nil, fmt.Errorf(
				"invalid quantity for item %d: %w",
				i+1,
				err,
			)
		}

		var estimatedUnitPrice *decimal.Decimal

		if itemBody.EstimatedUnitPrice != nil &&
			*itemBody.EstimatedUnitPrice != "" {

			price, err := decimal.NewFromString(
				*itemBody.EstimatedUnitPrice,
			)
			if err != nil {
				return nil, fmt.Errorf(
					"invalid estimated_unit_price for item %d: %w",
					i+1,
					err,
				)
			}

			estimatedUnitPrice = &price
		}

		items = append(
			items,
			&domain.PurchaseRequestItem{
				ItemID:             itemBody.ItemID,
				Description:        itemBody.Description,
				Quantity:           quantity,
				EstimatedUnitPrice: estimatedUnitPrice,
			},
		)
	}

	return &service.CreatePurchaseRequestInput{
		Request: request,
		Items:   items,
	}, nil
}
