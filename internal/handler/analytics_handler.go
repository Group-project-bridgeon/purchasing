package handler

import (
	"strconv"

	"github.com/Group-project-bridgeon/purchasing/internal/handler/helper"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
)

type AnalyticsHandler struct {
	service service.AnalyticsService
}

func NewAnalyticsHandler(
	service service.AnalyticsService,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		service: service,
	}
}

func (h *AnalyticsHandler) GetVendorScorecard(c *fiber.Ctx) error {
	vendorID := c.Params("vendorId")
	if vendorID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "vendorId parameter is required",
		})
	}

	fromDate, toDate, err := helper.ParseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	scorecard, err := h.service.GetVendorScorecard(c.UserContext(), vendorID, fromDate, toDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapScorecardToResponse(scorecard))
}

func (h *AnalyticsHandler) GetVendorPerformanceRankings(c *fiber.Ctx) error {
	fromDate, toDate, err := helper.ParseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	rankings, total, err := h.service.GetVendorPerformanceRankings(c.UserContext(), fromDate, toDate, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapRankingsToResponse(rankings, total))
}

func (h *AnalyticsHandler) GetDashboardKPIs(c *fiber.Ctx) error {
	fromDate, toDate, err := helper.ParseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	kpis, err := h.service.GetDashboardKPIs(c.UserContext(), fromDate, toDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapDashboardKPIsToResponse(kpis))
}

func (h *AnalyticsHandler) GetSpendByVendor(c *fiber.Ctx) error {
	fromDate, toDate, err := helper.ParseDateRange(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	spend, total, err := h.service.GetSpendByVendor(c.UserContext(), fromDate, toDate, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(helper.MapSpendListToResponse(spend, total))
}
