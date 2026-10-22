package v1

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterRFQRoutes registers all routes for RFQs, Vendor Quotations, Comparison, and Awarding.
func RegisterRFQRoutes(
	r fiber.Router,
	h *handler.RFQHandler,
	cfg config.RFQRoutes,
) {
	if h == nil {
		return
	}
	// Sourcing package operations
	r.Post(cfg.Base, h.Create)
	r.Post(cfg.FromRequest, h.CreateFromPR)
	r.Get(cfg.Base, h.List)
	r.Get(cfg.ByID, h.GetByID)
	r.Get(cfg.ByNumber, h.GetByNumber)
	r.Put(cfg.ByID, h.Update)
	r.Post(cfg.Publish, h.Publish)
	r.Post(cfg.Close, h.Close)
	r.Post(cfg.Cancel, h.Cancel)

	// Vendor bidding & quotation submissions
	r.Post(cfg.Quotations, h.SubmitQuotation)
	r.Get(cfg.Quotations, h.ListQuotations)
	r.Get(cfg.QuotationByID, h.GetQuotationByID)

	// Decision matrix & award
	r.Get(cfg.Compare, h.Compare)
	r.Post(cfg.Award, h.Award)
}
