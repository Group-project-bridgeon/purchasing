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

type RFQHandler struct {
	service service.RFQService
}

func NewRFQHandler(
	service service.RFQService,
) *RFQHandler {
	return &RFQHandler{
		service: service,
	}
}

func (h *RFQHandler) Create(c *fiber.Ctx) error {
	var body dto.CreateRFQBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreateRFQ(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	detail, err := h.service.CreateRFQ(c.UserContext(), input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.MapRFQDetailToResponse(detail))
}

func (h *RFQHandler) CreateFromPR(c *fiber.Ctx) error {
	prID := c.Params("prId")
	if prID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "prId parameter is required",
		})
	}

	var body dto.CreateRFQFromPRBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseCreateRFQFromPR(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	detail, err := h.service.CreateRFQFromPR(c.UserContext(), prID, input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.MapRFQDetailToResponse(detail))
}

func (h *RFQHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	detail, err := h.service.GetRFQByID(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "rfq not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQDetailToResponse(detail))
}

func (h *RFQHandler) GetByNumber(c *fiber.Ctx) error {
	rfqNumber := c.Params("rfqNumber")
	if rfqNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "rfqNumber parameter is required",
		})
	}

	detail, err := h.service.GetRFQByNumber(c.UserContext(), rfqNumber)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "rfq not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQDetailToResponse(detail))
}

func (h *RFQHandler) List(c *fiber.Ctx) error {
	var status *domain.RFQStatus
	if s := c.Query("status"); s != "" {
		st := domain.RFQStatus(s)
		status = &st
	}

	var prID *string
	if pr := c.Query("purchase_request_id"); pr != "" {
		prID = &pr
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	rfqs, total, err := h.service.ListRFQs(c.UserContext(), status, prID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQListToResponse(rfqs, total, limit, offset))
}

func (h *RFQHandler) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	var body dto.UpdateRFQBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	var deadline *string
	if body.SubmissionDeadline != nil {
		deadline = body.SubmissionDeadline
	}

	input := &service.UpdateRFQInput{
		Title: body.Title,
		Notes: body.Notes,
	}

	if deadline != nil {
		t, err := helper.ParseCreateRFQ(&dto.CreateRFQBody{SubmissionDeadline: *deadline})
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		input.SubmissionDeadline = &t.SubmissionDeadline
	}

	updated, err := h.service.UpdateRFQ(c.UserContext(), id, input)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		if errors.Is(err, domain.ErrRFQNotEditable) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQToResponse(updated))
}

func (h *RFQHandler) Publish(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	published, err := h.service.PublishRFQ(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		if errors.Is(err, domain.ErrRFQCannotPublish) || errors.Is(err, domain.ErrRFQDeadlinePassed) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQToResponse(published))
}

func (h *RFQHandler) Close(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	closed, err := h.service.CloseRFQ(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		if errors.Is(err, domain.ErrRFQCannotClose) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQToResponse(closed))
}

func (h *RFQHandler) Cancel(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	cancelled, err := h.service.CancelRFQ(c.UserContext(), id)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		if errors.Is(err, domain.ErrRFQCannotCancel) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRFQToResponse(cancelled))
}

func (h *RFQHandler) SubmitQuotation(c *fiber.Ctx) error {
	rfqID := c.Params("id")
	if rfqID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	var body dto.SubmitVendorQuotationBody
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}

	input, err := helper.ParseSubmitVendorQuotation(&body, rfqID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	detail, err := h.service.SubmitVendorQuotation(c.UserContext(), input)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "rfq not found"})
		}
		if errors.Is(err, domain.ErrRFQNotAcceptingQuotes) ||
			errors.Is(err, domain.ErrRFQDeadlinePassed) ||
			errors.Is(err, domain.ErrDuplicateVendorQuotation) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(helper.MapVendorQuotationDetailToResponse(detail))
}

func (h *RFQHandler) GetQuotationByID(c *fiber.Ctx) error {
	quotationID := c.Params("quotationId")
	if quotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "quotationId parameter is required",
		})
	}

	detail, err := h.service.GetVendorQuotationByID(c.UserContext(), quotationID)
	if err != nil {
		if errors.Is(err, domain.ErrQuotationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "quotation not found"})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapVendorQuotationDetailToResponse(detail))
}

func (h *RFQHandler) ListQuotations(c *fiber.Ctx) error {
	rfqID := c.Params("id")
	if rfqID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	quotations, err := h.service.ListVendorQuotationsByRFQ(c.UserContext(), rfqID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	resp := make([]dto.VendorQuotationResponse, 0, len(quotations))
	for _, q := range quotations {
		resp = append(resp, helper.MapVendorQuotationToResponse(q))
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}

func (h *RFQHandler) Compare(c *fiber.Ctx) error {
	rfqID := c.Params("id")
	if rfqID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}

	matrix, err := h.service.CompareRFQQuotations(c.UserContext(), rfqID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapComparisonMatrixToResponse(matrix))
}

func (h *RFQHandler) Award(c *fiber.Ctx) error {
	rfqID := c.Params("id")
	quotationID := c.Params("quotationId")

	if rfqID == "" || quotationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id and quotationId parameters are required",
		})
	}

	result, err := h.service.AwardVendorQuotation(c.UserContext(), rfqID, quotationID)
	if err != nil {
		if errors.Is(err, domain.ErrRFQNotFound) || errors.Is(err, domain.ErrQuotationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
		}
		if errors.Is(err, domain.ErrQuotationCannotAward) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(dto.AwardQuotationResponse{
		AwardedQuotationID: result.AwardedQuotationID,
		PurchaseOrderID:    result.PurchaseOrderID,
		PONumber:           result.PONumber,
		Status:             result.Status,
	})
}
