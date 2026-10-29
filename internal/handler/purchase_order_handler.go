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

type PurchaseOrderHandler struct {
	service service.PurchaseOrderService
}

func NewPurchaseOrderHandler(
	service service.PurchaseOrderService,
) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{
		service: service,
	}
}

func (h *PurchaseOrderHandler) Create(c *fiber.Ctx) error {
	var body dto.CreatePurchaseOrderBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreatePurchaseOrder(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	order, err := h.service.CreatePurchaseOrder(c.UserContext(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) CreateFromPR(c *fiber.Ctx) error {
	prID := c.Params("prId")

	var body dto.ConvertPRToPOBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseConvertPRToPO(&body, prID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	order, err := h.service.CreatePurchaseOrderFromPR(c.UserContext(), input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	order, err := h.service.GetPurchaseOrderByID(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) List(c *fiber.Ctx) error {
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

	var orders []*domain.PurchaseOrder
	var err error

	vendorID := c.Query("vendor_id")
	status := c.Query("status")

	if vendorID != "" {
		orders, err = h.service.ListPurchaseOrdersByVendor(c.UserContext(), vendorID, limit, offset)
	} else if status != "" {
		orders, err = h.service.ListPurchaseOrdersByStatus(c.UserContext(), domain.PurchaseOrderStatus(status), limit, offset)
	} else {
		orders, err = h.service.ListPurchaseOrders(c.UserContext(), limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]dto.PurchaseOrderResponse, 0, len(orders))
	for _, order := range orders {
		responses = append(responses, helper.ToPurchaseOrderResponse(order))
	}

	return c.Status(fiber.StatusOK).JSON(dto.ListPurchaseOrdersResponse{
		PurchaseOrders: responses,
		Limit:          limit,
		Offset:         offset,
	})
}

func (h *PurchaseOrderHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.UpdatePurchaseOrderBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	order, err := helper.ParseUpdatePurchaseOrder(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	order.ID = id

	updatedOrder, err := h.service.UpdatePurchaseOrder(c.UserContext(), order)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), updatedOrder.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreatePurchaseOrderResponse(updatedOrder, items))
}

func (h *PurchaseOrderHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")

	order, err := h.service.ApprovePurchaseOrder(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseOrderCannotApprove) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) Issue(c *fiber.Ctx) error {
	id := c.Params("id")

	order, err := h.service.IssuePurchaseOrder(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseOrderCannotIssue) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")

	order, err := h.service.CancelPurchaseOrder(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseOrderCannotCancel) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), order.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get purchase order items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreatePurchaseOrderResponse(order, items))
}

func (h *PurchaseOrderHandler) GetItems(c *fiber.Ctx) error {
	id := c.Params("id")

	items, err := h.service.GetPurchaseOrderItems(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	responses := make([]dto.PurchaseOrderItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, helper.ToPurchaseOrderItemResponse(item))
	}

	return c.Status(fiber.StatusOK).JSON(responses)
}

func (h *PurchaseOrderHandler) AddItem(c *fiber.Ctx) error {
	id := c.Params("id")

	var body dto.CreatePurchaseOrderItemBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	item, err := helper.ParseCreatePurchaseOrderItem(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	createdItem, err := h.service.AddPurchaseOrderItem(c.UserContext(), id, item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToPurchaseOrderItemResponse(createdItem))
}

func (h *PurchaseOrderHandler) UpdateItem(c *fiber.Ctx) error {
	id := c.Params("id")
	itemID := c.Params("itemId")

	var body dto.UpdatePurchaseOrderItemBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	item, err := helper.ParseUpdatePurchaseOrderItem(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	item.ID = itemID
	item.PurchaseOrderID = id

	updatedItem, err := h.service.UpdatePurchaseOrderItem(c.UserContext(), item)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToPurchaseOrderItemResponse(updatedItem))
}

func (h *PurchaseOrderHandler) DeleteItem(c *fiber.Ctx) error {
	itemID := c.Params("itemId")

	if err := h.service.DeletePurchaseOrderItem(c.UserContext(), itemID); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "purchase order item deleted successfully",
	})
}
