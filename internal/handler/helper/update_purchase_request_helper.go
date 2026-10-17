package helper

import (
	"fmt"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
)

func ParseUpdatePurchaseRequest(
	body *dto.UpdatePurchaseRequestBody,
) (*domain.PurchaseRequest, error) {
	if body == nil {
		return nil, fmt.Errorf("request body cannot be nil")
	}

	var requiredBy *time.Time

	if body.RequiredBy != nil && *body.RequiredBy != "" {
		parsedDate, err := time.Parse("2006-01-02", *body.RequiredBy)
		if err != nil {
			return nil, fmt.Errorf("invalid required_by date: %w", err)
		}

		requiredBy = &parsedDate
	}

	return &domain.PurchaseRequest{
		RequiredBy: requiredBy,
		Purpose:    body.Purpose,
		Notes:      body.Notes,
	}, nil
}
