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
