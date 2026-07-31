package spec

type GetMetricsResponse struct {
	LeadConversion float64 `json:"lead_conversion"`
	RFQConversion  float64 `json:"rfq_conversion"`
	WinRate        float64 `json:"win_rate"`
	Revenue        float64 `json:"revenue"`
}
