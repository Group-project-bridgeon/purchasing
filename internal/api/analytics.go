package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
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

	r.Get(cfg.VendorScorecard, middleware.Require("purchasing.analytics:read"), h.GetVendorScorecard)             // GET /api/v1/analytics/vendors/:vendorId/scorecard
	r.Get(cfg.VendorPerformance, middleware.Require("purchasing.analytics:read"), h.GetVendorPerformanceRankings) // GET /api/v1/analytics/vendors/performance
	r.Get(cfg.DashboardKPIs, middleware.Require("purchasing.analytics:read"), h.GetDashboardKPIs)                 // GET /api/v1/analytics/dashboard/kpis
	r.Get(cfg.SpendByVendor, middleware.Require("purchasing.analytics:read"), h.GetSpendByVendor)                 // GET /api/v1/analytics/spend/by-vendor
}
