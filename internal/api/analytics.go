package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterAnalyticsRoutes registers all routes for Vendor Scorecards, Performance Rankings, and Spend KPIs.
func RegisterAnalyticsRoutes(
	r fiber.Router,
	h *handler.AnalyticsHandler,
	cfg config.AnalyticsRoutes,
) {
	if h == nil {
		return
	}

	r.Get(cfg.VendorScorecard, h.GetVendorScorecard)             // GET /api/v1/analytics/vendors/:vendorId/scorecard
	r.Get(cfg.VendorPerformance, h.GetVendorPerformanceRankings) // GET /api/v1/analytics/vendors/performance
	r.Get(cfg.DashboardKPIs, h.GetDashboardKPIs)                 // GET /api/v1/analytics/dashboard/kpis
	r.Get(cfg.SpendByVendor, h.GetSpendByVendor)                 // GET /api/v1/analytics/spend/by-vendor
}
