package handler

import (
	"errors"
	"strconv"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
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

	response := helper.ToPurchaseRequestResponse(
		request,
		items,
	)

	return c.Status(fiber.StatusCreated).JSON(response)
}

func (h *PurchaseRequestHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	request, err := h.service.GetPurchaseRequestByID(
		c.UserContext(),
		id,
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

	response := helper.ToPurchaseRequestResponse(
		request,
		items,
	)

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) List(c *fiber.Ctx) error {
	limit := 20
	offset := 0

	if value := c.Query("limit"); value != "" {
		parsedLimit, err := strconv.Atoi(value)
		if err != nil || parsedLimit <= 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "limit must be a positive integer",
			})
		}

		limit = parsedLimit
	}

	if value := c.Query("offset"); value != "" {
		parsedOffset, err := strconv.Atoi(value)
		if err != nil || parsedOffset < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "offset must be a non-negative integer",
			})
		}

		offset = parsedOffset
	}

	var requests []*domain.PurchaseRequest
	var err error

	status := c.Query("status")

	if status != "" {
		requestStatus := domain.PurchaseRequestStatus(status)

		switch requestStatus {
		case domain.PurchaseRequestStatusDraft,
			domain.PurchaseRequestStatusSubmitted,
			domain.PurchaseRequestStatusApproved,
			domain.PurchaseRequestStatusRejected:

			requests, err = h.service.ListPurchaseRequestsByStatus(
				c.UserContext(),
				requestStatus,
				limit,
				offset,
			)

		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid purchase request status",
			})
		}
	} else {
		requests, err = h.service.ListPurchaseRequests(
			c.UserContext(),
			limit,
			offset,
		)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make(
		[]dto.PurchaseRequestResponse,
		0,
		len(requests),
	)

	for _, request := range requests {
		responses = append(
			responses,
			dto.PurchaseRequestResponse{
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
		)
	}

	response := dto.ListPurchaseRequestsResponse{
		PurchaseRequests: responses,
		Limit:            limit,
		Offset:           offset,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdatePurchaseRequestBody

	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	request, err := helper.ParseUpdatePurchaseRequest(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	request.ID = id

	updatedRequest, err := h.service.UpdatePurchaseRequest(
		c.UserContext(),
		request,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseRequestItems(
		c.UserContext(),
		updatedRequest.ID,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase request items",
		})
	}

	response := helper.ToPurchaseRequestResponse(
		updatedRequest,
		items,
	)

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) Submit(c *fiber.Ctx) error {
	id := c.Params("id")

	request, err := h.service.SubmitPurchaseRequest(
		c.UserContext(),
		id,
	)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseRequestCannotSubmit) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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

	response := helper.ToPurchaseRequestResponse(
		request,
		items,
	)

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")

	request, err := h.service.ApprovePurchaseRequest(
		c.UserContext(),
		id,
	)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseRequestCannotApprove) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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

	response := helper.ToPurchaseRequestResponse(
		request,
		items,
	)

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) Reject(c *fiber.Ctx) error {
	id := c.Params("id")

	request, err := h.service.RejectPurchaseRequest(
		c.UserContext(),
		id,
	)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseRequestCannotReject) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

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

	response := helper.ToPurchaseRequestResponse(
		request,
		items,
	)

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *PurchaseRequestHandler) GetItems(c *fiber.Ctx) error {
	id := c.Params("id")

	items, err := h.service.GetPurchaseRequestItems(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]dto.PurchaseRequestItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, helper.ToPurchaseRequestItemResponse(item))
	}

	return c.Status(fiber.StatusOK).JSON(responses)
}

func (h *PurchaseRequestHandler) AddItem(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.CreatePurchaseRequestItemBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	item, err := helper.ParseCreatePurchaseRequestItem(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	createdItem, err := h.service.AddPurchaseRequestItem(c.UserContext(), id, item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToPurchaseRequestItemResponse(createdItem))
}

func (h *PurchaseRequestHandler) UpdateItem(c *fiber.Ctx) error {
	id := c.Params("id")
	itemID := c.Params("itemId")

	var body dto.UpdatePurchaseRequestItemBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	item, err := helper.ParseUpdatePurchaseRequestItem(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	item.ID = itemID
	item.PurchaseRequestID = id

	updatedItem, err := h.service.UpdatePurchaseRequestItem(c.UserContext(), item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToPurchaseRequestItemResponse(updatedItem))
}

func (h *PurchaseRequestHandler) DeleteItem(c *fiber.Ctx) error {
	itemID := c.Params("itemId")

	if err := h.service.DeletePurchaseRequestItem(c.UserContext(), itemID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "purchase request item deleted successfully",
	})
}
