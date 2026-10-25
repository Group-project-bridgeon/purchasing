package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/gofiber/fiber/v2"
)

// RegisterRFQRoutes registers all routes for RFQs, Vendor Quotations, Comparison Matrix, and Awards.
func RegisterRFQRoutes(
	r fiber.Router,
	h *handler.RFQHandler,
	cfg config.RFQRoutes,
) {
	if h == nil {
		return
	}

	// Sourcing package operations
	r.Post(cfg.Base, h.Create)              // POST /api/v1/rfqs
	r.Post(cfg.FromRequest, h.CreateFromPR) // POST /api/v1/rfqs/from-request/:prId
	r.Get(cfg.Base, h.List)                 // GET  /api/v1/rfqs
	r.Get(cfg.ByID, h.GetByID)              // GET  /api/v1/rfqs/:id
	r.Get(cfg.ByNumber, h.GetByNumber)      // GET  /api/v1/rfqs/number/:rfqNumber
	r.Put(cfg.ByID, h.Update)               // PUT  /api/v1/rfqs/:id
	r.Post(cfg.Publish, h.Publish)          // POST /api/v1/rfqs/:id/publish
	r.Post(cfg.Close, h.Close)              // POST /api/v1/rfqs/:id/close
	r.Post(cfg.Cancel, h.Cancel)            // POST /api/v1/rfqs/:id/cancel

	// Vendor bidding & quotation submissions
	r.Post(cfg.Quotations, h.SubmitQuotation)    // POST /api/v1/rfqs/:id/quotations
	r.Get(cfg.Quotations, h.ListQuotations)      // GET  /api/v1/rfqs/:id/quotations
	r.Get(cfg.QuotationByID, h.GetQuotationByID) // GET  /api/v1/rfqs/:id/quotations/:quotationId

	// Decision matrix & award
	r.Get(cfg.Compare, h.Compare) // GET  /api/v1/rfqs/:id/compare
	r.Post(cfg.Award, h.Award)    // POST /api/v1/rfqs/:id/award/:quotationId
}
