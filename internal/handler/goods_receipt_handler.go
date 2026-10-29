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

type GoodsReceiptHandler struct {
	service service.GoodsReceiptService
}

func NewGoodsReceiptHandler(
	service service.GoodsReceiptService,
) *GoodsReceiptHandler {
	return &GoodsReceiptHandler{
		service: service,
	}
}

func (h *GoodsReceiptHandler) Create(c *fiber.Ctx) error {
	var body dto.CreateGoodsReceiptBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreateGoodsReceipt(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	grn, err := h.service.CreateGoodsReceipt(c.UserContext(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPOStatusForReceiving) ||
			errors.Is(err, domain.ErrExceedsOrderedQuantity) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetGoodsReceiptItems(c.UserContext(), grn.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get goods receipt items",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToCreateGoodsReceiptResponse(grn, items))
}

func (h *GoodsReceiptHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	grn, err := h.service.GetGoodsReceiptByID(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetGoodsReceiptItems(c.UserContext(), grn.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get goods receipt items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreateGoodsReceiptResponse(grn, items))
}

func (h *GoodsReceiptHandler) List(c *fiber.Ctx) error {
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

	var receipts []*domain.GoodsReceipt
	var err error

	vendorID := c.Query("vendor_id")
	if vendorID != "" {
		receipts, err = h.service.ListGoodsReceiptsByVendor(c.UserContext(), vendorID, limit, offset)
	} else {
		receipts, err = h.service.ListGoodsReceipts(c.UserContext(), limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]dto.GoodsReceiptResponse, 0, len(receipts))
	for _, gr := range receipts {
		responses = append(responses, helper.ToGoodsReceiptResponse(gr))
	}

	return c.Status(fiber.StatusOK).JSON(dto.ListGoodsReceiptsResponse{
		GoodsReceipts: responses,
		Limit:         limit,
		Offset:        offset,
	})
}

func (h *GoodsReceiptHandler) ListByPO(c *fiber.Ctx) error {
	poID := c.Params("id")

	receipts, err := h.service.ListGoodsReceiptsByPO(c.UserContext(), poID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]dto.GoodsReceiptResponse, 0, len(receipts))
	for _, gr := range receipts {
		responses = append(responses, helper.ToGoodsReceiptResponse(gr))
	}

	return c.Status(fiber.StatusOK).JSON(responses)
}

func (h *GoodsReceiptHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")

	grn, err := h.service.CancelGoodsReceipt(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrGoodsReceiptCannotCancel) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetGoodsReceiptItems(c.UserContext(), grn.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get goods receipt items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreateGoodsReceiptResponse(grn, items))
}
