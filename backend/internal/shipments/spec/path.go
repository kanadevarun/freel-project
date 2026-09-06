package spec

// Route paths relative to the /api/v1/shipments mount point
const (
	ListURL          = "/"
	GetURL           = "/{id:[0-9]+}"
	CarrierUpdateURL = "/{id:[0-9]+}/carrier-update"
	ResolveExcURL    = "/exceptions/{id:[0-9]+}/resolve"
)
