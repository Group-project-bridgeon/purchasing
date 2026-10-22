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

type PurchaseReturnHandler struct {
	service service.PurchaseReturnService
}

func NewPurchaseReturnHandler(
	service service.PurchaseReturnService,
) *PurchaseReturnHandler {
	return &PurchaseReturnHandler{
		service: service,
	}
}

func (h *PurchaseReturnHandler) Create(c *fiber.Ctx) error {
	var body dto.CreatePurchaseReturnBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreatePurchaseReturn(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	detail, err := h.service.CreatePurchaseReturn(c.UserContext(), input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidPOStatusForReturn) ||
			errors.Is(err, domain.ErrExceedsReturnableQuantity) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.MapPurchaseReturnDetailToResponse(detail))
}

func (h *PurchaseReturnHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	detail, err := h.service.GetPurchaseReturnByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnDetailToResponse(detail))
}

func (h *PurchaseReturnHandler) GetByNumber(c *fiber.Ctx) error {
	returnNumber := c.Params("returnNumber")
	if returnNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "returnNumber parameter is required",
		})
	}

	detail, err := h.service.GetPurchaseReturnByNumber(c.UserContext(), returnNumber)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnDetailToResponse(detail))
}

func (h *PurchaseReturnHandler) List(c *fiber.Ctx) error {
	var poID *string
	if po := c.Query("po_id"); po != "" {
		poID = &po
	}

	var vendorID *string
	if v := c.Query("vendor_id"); v != "" {
		vendorID = &v
	}

	var status *domain.PurchaseReturnStatus
	if s := c.Query("status"); s != "" {
		st := domain.PurchaseReturnStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	returns, total, err := h.service.ListPurchaseReturns(c.UserContext(), poID, vendorID, status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnListToResponse(returns, total, limit, offset))
}

func (h *PurchaseReturnHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	var body dto.UpdatePurchaseReturnBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	updated, err := h.service.UpdatePurchaseReturn(c.UserContext(), id, &service.UpdatePurchaseReturnInput{
		Reason:         body.Reason,
		Carrier:        body.Carrier,
		TrackingNumber: body.TrackingNumber,
		Notes:          body.Notes,
	})
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		if errors.Is(err, domain.ErrPurchaseReturnNotEditable) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnToResponse(updated))
}

func (h *PurchaseReturnHandler) Approve(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	approved, err := h.service.ApprovePurchaseReturn(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		if errors.Is(err, domain.ErrPurchaseReturnCannotApprove) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnToResponse(approved))
}

func (h *PurchaseReturnHandler) Ship(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	var body dto.ShipPurchaseReturnBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	shipped, err := h.service.ShipPurchaseReturn(c.UserContext(), id, body.Carrier, body.TrackingNumber)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		if errors.Is(err, domain.ErrPurchaseReturnCannotShip) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnToResponse(shipped))
}

func (h *PurchaseReturnHandler) Complete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	detail, err := h.service.CompletePurchaseReturn(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		if errors.Is(err, domain.ErrPurchaseReturnCannotComplete) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnDetailToResponse(detail))
}

func (h *PurchaseReturnHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	cancelled, err := h.service.CancelPurchaseReturn(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrPurchaseReturnNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "purchase return not found",
			})
		}
		if errors.Is(err, domain.ErrPurchaseReturnCannotCancel) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapPurchaseReturnToResponse(cancelled))
}
