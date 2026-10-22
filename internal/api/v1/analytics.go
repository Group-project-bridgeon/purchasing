package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterAnalyticsRoutes registers all routes for Vendor Scorecards, Rankings, and Spend KPIs.
func RegisterAnalyticsRoutes(
	r fiber.Router,
	h *handler.AnalyticsHandler,
	cfg config.AnalyticsRoutes,
) {
	if h == nil {
		return
	}
	r.Get(cfg.VendorScorecard, h.GetVendorScorecard)
	r.Get(cfg.VendorPerformance, h.GetVendorPerformanceRankings)
	r.Get(cfg.DashboardKPIs, h.GetDashboardKPIs)
	r.Get(cfg.SpendByVendor, h.GetSpendByVendor)
}
