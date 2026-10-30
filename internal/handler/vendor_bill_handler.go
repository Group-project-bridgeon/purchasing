package handler

import (
	"errors"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"strconv"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/handler/helper"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
)

type VendorBillHandler struct {
	service service.VendorBillService
}

func NewVendorBillHandler(
	service service.VendorBillService,
) *VendorBillHandler {
	return &VendorBillHandler{
		service: service,
	}
}

func (h *VendorBillHandler) Create(c *fiber.Ctx) error {
	var body dto.CreateVendorBillBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	if actorErr := middleware.BindActor(c, &body.CreatedBy); actorErr != nil {
		return c.Status(403).JSON(fiber.Map{"error": "actor mismatch"})
	}
	input, err := helper.ParseCreateVendorBill(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	bill, err := h.service.CreateVendorBill(c.UserContext(), input)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateVendorInvoice) ||
			errors.Is(err, domain.ErrInvalidPOStatusForBilling) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": clientError(c, err),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	items, err := h.service.GetVendorBillItems(c.UserContext(), bill.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get vendor bill items",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.ToCreateVendorBillResponse(bill, items))
}

func (h *VendorBillHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")

	bill, err := h.service.GetVendorBillByID(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	items, err := h.service.GetVendorBillItems(c.UserContext(), bill.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get vendor bill items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreateVendorBillResponse(bill, items))
}

func (h *VendorBillHandler) List(c *fiber.Ctx) error {
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

	var bills []*domain.VendorBill
	var err error

	vendorID := c.Query("vendor_id")
	status := c.Query("status")
	matchStatus := c.Query("match_status")

	if vendorID != "" {
		bills, err = h.service.ListVendorBillsByVendor(c.UserContext(), vendorID, limit, offset)
	} else if status != "" {
		bills, err = h.service.ListVendorBillsByStatus(c.UserContext(), domain.BillStatus(status), limit, offset)
	} else if matchStatus != "" {
		bills, err = h.service.ListVendorBillsByMatchStatus(c.UserContext(), domain.BillMatchStatus(matchStatus), limit, offset)
	} else {
		bills, err = h.service.ListVendorBills(c.UserContext(), limit, offset)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	responses := make([]dto.VendorBillResponse, 0, len(bills))
	for _, b := range bills {
		responses = append(responses, helper.ToVendorBillResponse(b))
	}

	return c.Status(fiber.StatusOK).JSON(dto.ListVendorBillsResponse{
		VendorBills: responses,
		Limit:       limit,
		Offset:      offset,
	})
}

func (h *VendorBillHandler) ListByPO(c *fiber.Ctx) error {
	poID := c.Params("id")

	bills, err := h.service.ListVendorBillsByPO(c.UserContext(), poID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	responses := make([]dto.VendorBillResponse, 0, len(bills))
	for _, b := range bills {
		responses = append(responses, helper.ToVendorBillResponse(b))
	}

	return c.Status(fiber.StatusOK).JSON(responses)
}

func (h *VendorBillHandler) Match(c *fiber.Ctx) error {
	id := c.Params("id")

	result, err := h.service.MatchVendorBill(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToMatchBillResponse(result))
}

func (h *VendorBillHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")

	bill, err := h.service.ApproveVendorBill(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVendorBillCannotApprove) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": clientError(c, err),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	items, err := h.service.GetVendorBillItems(c.UserContext(), bill.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get vendor bill items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreateVendorBillResponse(bill, items))
}

func (h *VendorBillHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")

	bill, err := h.service.CancelVendorBill(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrVendorBillCannotCancel) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": clientError(c, err),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": clientError(c, err),
		})
	}

	items, err := h.service.GetVendorBillItems(c.UserContext(), bill.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to get vendor bill items",
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.ToCreateVendorBillResponse(bill, items))
}
