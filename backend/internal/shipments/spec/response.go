package spec

// APIResponse represents the standard JSON API envelope
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// ShipmentDetailData represents the payload returned by GetShipment details endpoint
type ShipmentDetailData struct {
	Shipment   *Shipment            `json:"shipment"`
	Milestones []*ShipmentMilestone `json:"milestones"`
	Exceptions []*ShipmentException `json:"exceptions"`
}

// InternalShipmentDetailData represents the payload returned by GetShipmentInternal endpoint
type InternalShipmentDetailData struct {
	Shipment   *Shipment            `json:"shipment"`
	Milestones []*ShipmentMilestone `json:"milestones"`
}

// ShipmentWorkspaceData represents the payload returned by ListShipments in workspace mode
type ShipmentWorkspaceData struct {
	Shipments  []*Shipment  `json:"shipments"`
	KPIs       ShipmentKPIs `json:"kpis"`
	Pagination Pagination   `json:"pagination"`
}

// Pagination contains the workspace query pagination stats
type Pagination struct {
	CurrentPage int `json:"current_page"`
	PageSize    int `json:"page_size"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
}
