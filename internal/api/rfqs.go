package api

import (
	"github.com/Group-project-bridgeon/purchasing/internal/config"
	"github.com/Group-project-bridgeon/purchasing/internal/handler"
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
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
	r.Post(cfg.Base, middleware.Require("purchasing.rfqs:create"), h.Create)                                                   // POST /api/v1/rfqs
	r.Post(cfg.FromRequest, middleware.Require("purchasing.rfqs:create", "purchasing.purchase-requests:read"), h.CreateFromPR) // POST /api/v1/rfqs/from-request/:prId
	r.Get(cfg.Base, middleware.Require("purchasing.rfqs:read"), h.List)                                                        // GET  /api/v1/rfqs
	r.Get(cfg.ByID, middleware.Require("purchasing.rfqs:read"), h.GetByID)                                                     // GET  /api/v1/rfqs/:id
	r.Get(cfg.ByNumber, middleware.Require("purchasing.rfqs:read"), h.GetByNumber)                                             // GET  /api/v1/rfqs/number/:rfqNumber
	r.Put(cfg.ByID, middleware.Require("purchasing.rfqs:update"), h.Update)                                                    // PUT  /api/v1/rfqs/:id
	r.Post(cfg.Publish, middleware.Require("purchasing.rfqs:publish"), h.Publish)                                              // POST /api/v1/rfqs/:id/publish
	r.Post(cfg.Close, middleware.Require("purchasing.rfqs:close"), h.Close)                                                    // POST /api/v1/rfqs/:id/close
	r.Post(cfg.Cancel, middleware.Require("purchasing.rfqs:cancel"), h.Cancel)                                                 // POST /api/v1/rfqs/:id/cancel

	// Vendor bidding & quotation submissions
	r.Post(cfg.Quotations, middleware.Require("purchasing.vendor-quotations:create"), h.SubmitQuotation)  // POST /api/v1/rfqs/:id/quotations
	r.Get(cfg.Quotations, middleware.Require("purchasing.vendor-quotations:read"), h.ListQuotations)      // GET  /api/v1/rfqs/:id/quotations
	r.Get(cfg.QuotationByID, middleware.Require("purchasing.vendor-quotations:read"), h.GetQuotationByID) // GET  /api/v1/rfqs/:id/quotations/:quotationId

	// Decision matrix & award
	r.Get(cfg.Compare, middleware.Require("purchasing.vendor-quotations:read"), h.Compare)                       // GET  /api/v1/rfqs/:id/compare
	r.Post(cfg.Award, middleware.Require("purchasing.rfqs:award", "purchasing.purchase-orders:create"), h.Award) // POST /api/v1/rfqs/:id/award/:quotationId
}
