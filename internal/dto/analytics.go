package dto

type VendorScorecardResponse struct {
	VendorID              string `json:"vendor_id"`
	TotalOrders           int64  `json:"total_orders"`
	TotalCommittedSpend   string `json:"total_committed_spend"`
	TotalInvoicedSpend    string `json:"total_invoiced_spend"`
	TotalReceipts         int64  `json:"total_receipts"`
	OnTimeReceipts        int64  `json:"on_time_receipts"`
	OnTimeDeliveryRate    string `json:"on_time_delivery_rate"`
	TotalOrderedItems     int64  `json:"total_ordered_items"`
	FulfilledItems        int64  `json:"fulfilled_items"`
	InFullDeliveryRate    string `json:"in_full_delivery_rate"`
	OTIFRate              string `json:"otif_rate"`
	TotalReceivedQuantity string `json:"total_received_quantity"`
	TotalRejectedQuantity string `json:"total_rejected_quantity"`
	QualityAcceptanceRate string `json:"quality_acceptance_rate"`
	TotalReturns          int64  `json:"total_returns"`
	TotalReturnedQuantity string `json:"total_returned_quantity"`
	TotalReturnedAmount   string `json:"total_returned_amount"`
	ReturnRate            string `json:"return_rate"`
	TotalBills            int64  `json:"total_bills"`
	MatchedBills          int64  `json:"matched_bills"`
	DiscrepantBills       int64  `json:"discrepant_bills"`
	BillingAccuracyRate   string `json:"billing_accuracy_rate"`
	CompositeScore        string `json:"composite_score"`
}

type VendorPerformanceRankingResponse struct {
	VendorID              string `json:"vendor_id"`
	TotalOrders           int64  `json:"total_orders"`
	CommittedSpend        string `json:"committed_spend"`
	OTIFRate              string `json:"otif_rate"`
	QualityAcceptanceRate string `json:"quality_acceptance_rate"`
	CompositeScore        string `json:"composite_score"`
	Rank                  int    `json:"rank"`
}

type VendorPerformanceListResponse struct {
	Vendors []VendorPerformanceRankingResponse `json:"vendors"`
	Total   int64                              `json:"total"`
}

type DashboardKPIsResponse struct {
	TotalOrders                int64            `json:"total_orders"`
	TotalCommittedSpend        string           `json:"total_committed_spend"`
	UnbilledReceivingLiability string           `json:"unbilled_receiving_liability"`
	PendingDeliveryExposure    string           `json:"pending_delivery_exposure"`
	OrdersByStatus             map[string]int64 `json:"orders_by_status"`
}

type VendorSpendSummaryResponse struct {
	VendorID        string `json:"vendor_id"`
	TotalOrders     int64  `json:"total_orders"`
	CommittedSpend  string `json:"committed_spend"`
	SpendPercentage string `json:"spend_percentage"`
}

type VendorSpendListResponse struct {
	SpendByVendor []VendorSpendSummaryResponse `json:"spend_by_vendor"`
	TotalSpend    string                       `json:"total_spend"`
}
