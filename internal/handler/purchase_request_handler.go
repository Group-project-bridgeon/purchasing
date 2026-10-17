package handler

import (
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/handler/helper"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
)

type PurchaseRequestHandler struct {
	service service.PurchaseRequestService
}

func NewPurchaseRequestHandler(
	service service.PurchaseRequestService,
) *PurchaseRequestHandler {
	return &PurchaseRequestHandler{
		service: service,
	}
}
func (h *PurchaseRequestHandler) Create(c *fiber.Ctx) error {
	var body dto.CreatePurchaseRequestBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreatePurchaseRequest(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	request, err := h.service.CreatePurchaseRequest(
		c.UserContext(),
		input,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseRequestItems(
		c.UserContext(),
		request.ID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase request items",
		})
	}

	itemResponses := make(
		[]dto.PurchaseRequestItemResponse,
		0,
		len(items),
	)

	for _, item := range items {
		var estimatedUnitPrice *string

		if item.EstimatedUnitPrice != nil {
			value := item.EstimatedUnitPrice.String()
			estimatedUnitPrice = &value
		}

		itemResponses = append(
			itemResponses,
			dto.PurchaseRequestItemResponse{
				ID:                 item.ID,
				PurchaseRequestID:  item.PurchaseRequestID,
				ItemID:             item.ItemID,
				Description:        item.Description,
				Quantity:           item.Quantity.String(),
				EstimatedUnitPrice: estimatedUnitPrice,
			},
		)
	}

	response := dto.CreatePurchaseRequestResponse{
		PurchaseRequest: dto.PurchaseRequestResponse{
			ID:            request.ID,
			RequestNumber: request.RequestNumber,
			RequestedBy:   request.RequestedBy,
			DepartmentID:  request.DepartmentID,
			Status:        request.Status,
			RequestedAt:   request.RequestedAt,
			RequiredBy:    request.RequiredBy,
			Purpose:       request.Purpose,
			Notes:         request.Notes,
			CreatedAt:     request.CreatedAt,
			UpdatedAt:     request.UpdatedAt,
		},
		Items: itemResponses,
	}

	return c.Status(fiber.StatusCreated).JSON(response)
}
