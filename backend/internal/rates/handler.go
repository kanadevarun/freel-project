package rates

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/utils"
)

// Handler exposes the Rate Intelligence API endpoints.
type Handler struct {
	svc Service
}

// NewHandler creates a new rates Handler.
func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// SearchRates handles GET /api/v1/rates/search
//
// Query parameters:
//   - origin      (required) — UN/LOCODE or port alias, e.g., "INNSA" or "Nhava Sheva"
//   - destination (required) — UN/LOCODE or port alias, e.g., "DEHAM"
//   - equipment   (optional) — "20GP"|"40GP"|"40HC"|"45HC"|"REEFER"; default "40GP"
//   - date        (optional) — "YYYY-MM-DD"; filter rates valid on this date
//   - carriers    (optional) — comma-separated SCACs, e.g., "MAEU,MSCU"
//   - sources     (optional) — comma-separated sources, e.g., "SPOT_API,CONTRACT_PDF"
//
// Response: { success: true, data: RateSearchResult }
func (h *Handler) SearchRates(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	origin := strings.TrimSpace(r.URL.Query().Get("origin"))
	destination := strings.TrimSpace(r.URL.Query().Get("destination"))
	if origin == "" || destination == "" {
		utils.Error(w, http.StatusBadRequest, "origin and destination query parameters are required", "MISSING_PARAMS")
		return
	}

	equipment := strings.TrimSpace(r.URL.Query().Get("equipment"))
	if equipment == "" {
		equipment = "40GP"
	}
	incoterms := strings.TrimSpace(r.URL.Query().Get("incoterms"))

	q := RateQuery{
		OrgID:           userCtx.OrgID,
		OriginPort:      origin,
		DestinationPort: destination,
		EquipmentType:   equipment,
		MaxResults:      20,
		Incoterms:       incoterms,
	}

	// Optional carrier filter
	if carriersParam := r.URL.Query().Get("carriers"); carriersParam != "" {
		for _, scac := range strings.Split(carriersParam, ",") {
			if s := strings.TrimSpace(scac); s != "" {
				q.CarrierSCACs = append(q.CarrierSCACs, strings.ToUpper(s))
			}
		}
	}

	// Optional source filter
	if sourcesParam := r.URL.Query().Get("sources"); sourcesParam != "" {
		for _, src := range strings.Split(sourcesParam, ",") {
			if s := strings.TrimSpace(src); s != "" {
				q.Sources = append(q.Sources, RateSource(strings.ToUpper(s)))
			}
		}
	}

	result, err := h.svc.SearchRates(r.Context(), q)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "RATE_SEARCH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Rates retrieved successfully", result)
}

// GetRate handles GET /api/v1/rates/{id}
// Returns full detail for a single canonical rate including surcharge breakdown.
func (h *Handler) GetRate(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	// Extract ID from path — works with both chi and gorilla/mux
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	id := ""
	if len(parts) > 0 {
		id = parts[len(parts)-1]
	}
	if id == "" {
		utils.Error(w, http.StatusBadRequest, "rate id is required", "MISSING_ID")
		return
	}

	rate, err := h.svc.GetRateByID(r.Context(), userCtx.OrgID, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			utils.Error(w, http.StatusNotFound, "Rate not found", "NOT_FOUND")
			return
		}
		utils.Error(w, http.StatusInternalServerError, err.Error(), "GET_RATE_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Rate retrieved successfully", rate)
}

// RefreshSpotRates handles POST /api/v1/rates/spot/refresh
//
// Body: { "origin": "INNSA", "destination": "DEHAM", "equipment_type": "40GP" }
//
// Forces a fresh carrier API fetch for the given lane, normalizes and stores
// the results, and returns the updated rate set. Use when the user wants
// live prices instead of cached results.
func (h *Handler) RefreshSpotRates(w http.ResponseWriter, r *http.Request) {
	userCtx, ok := r.Context().Value(middleware.UserContextKey).(middleware.UserContext)
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}

	var body struct {
		Origin      string `json:"origin"`
		Destination string `json:"destination"`
		Equipment   string `json:"equipment_type"`
		Incoterms   string `json:"incoterms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.Error(w, http.StatusBadRequest, "Invalid request body", "INVALID_PAYLOAD")
		return
	}
	if body.Origin == "" || body.Destination == "" {
		utils.Error(w, http.StatusBadRequest, "origin and destination are required", "MISSING_PARAMS")
		return
	}
	if body.Equipment == "" {
		body.Equipment = "40GP"
	}

	q := RateQuery{
		OrgID:           userCtx.OrgID,
		OriginPort:      body.Origin,
		DestinationPort: body.Destination,
		EquipmentType:   body.Equipment,
		MaxResults:      20,
		Incoterms:       body.Incoterms,
	}

	result, err := h.svc.RefreshSpotRates(r.Context(), userCtx.OrgID, q, nil)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, err.Error(), "SPOT_REFRESH_FAILED")
		return
	}

	utils.Success(w, http.StatusOK, "Spot rates refreshed successfully", result)
}

// NormalizePort handles GET /internal/ports/normalize
//
// Simple meaning:
//   This is a secure internal helper endpoint. The Python AI agents call this endpoint
//   to clean up human typos in port names (like "Nhava Sheva" or "Jnpt") and translate them
//   to standard 5-character shipping codes called UN/LOCODEs (like "INNSA").
//
// Access Security:
//   Requires the caller to pass a secret key header: 'X-LogisticsHQ-Service-Key'.
//   If missing or incorrect, it returns a 401 Unauthorized error.
//
// Example Request:
//   GET /internal/ports/normalize?query=Nhava+Sheva
//   Header: X-LogisticsHQ-Service-Key: internal-service-key-logisticshq
//
// Example JSON Response:
//   {
//     "query": "Nhava Sheva",
//     "normalized": "INNSA",
//     "is_known": true
//   }
func (h *Handler) NormalizePort(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the caller. We look for the service key in the request header.
	token := r.Header.Get("X-LogisticsHQ-Service-Key")
	if token == "" {
		// Fallback to checking the query parameter for development testing convenience.
		token = r.URL.Query().Get("service_key")
	}
	
	// Fetch the expected token from the system's environment variables.
	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expectedToken == "" {
		if os.Getenv("APP_ENV") == "production" {
			http.Error(w, "Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments", http.StatusInternalServerError)
			return
		}
		expectedToken = "internal-service-key-logisticshq"
	}
	
	// If the token is invalid, reject the request immediately.
	if token != expectedToken {
		http.Error(w, "Unauthorized access: Invalid service key token", http.StatusUnauthorized)
		return
	}

	// 2. Read and sanitize the 'query' URL parameter.
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		http.Error(w, "Missing parameter: 'query' is required", http.StatusBadRequest)
		return
	}

	// 3. Perform the actual lookup in the alias map dictionary.
	// Example: "Navi Mumbai" will map to "INNSA".
	normalized := NormalizePort(query)
	
	// Check if this is a known port listed in the database/alias dictionary.
	isKnown := IsKnownPort(query)

	// 4. Wrap result in a map and serialize to JSON response.
	resp := map[string]interface{}{
		"query":      query,
		"normalized": normalized,
		"is_known":   isKnown,
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// SearchPorts handles GET /internal/ports/search
//
// Simple meaning:
//   Validates X-LogisticsHQ-Service-Key internal token and queries
//   static portAliasMap for aliases or LOCODEs containing query.
func (h *Handler) SearchPorts(w http.ResponseWriter, r *http.Request) {
	// 1. Authenticate the caller. We look for the service key in the request header.
	token := r.Header.Get("X-LogisticsHQ-Service-Key")
	if token == "" {
		token = r.URL.Query().Get("service_key")
	}
	
	expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
	if expectedToken == "" {
		if os.Getenv("APP_ENV") == "production" {
			http.Error(w, "Configuration error: INTERNAL_SERVICE_TOKEN must be specified in production environments", http.StatusInternalServerError)
			return
		}
		expectedToken = "internal-service-key-logisticshq"
	}
	
	if token != expectedToken {
		http.Error(w, "Unauthorized access: Invalid service key token", http.StatusUnauthorized)
		return
	}

	// 2. Read and sanitize the 'query' URL parameter.
	query := strings.TrimSpace(r.URL.Query().Get("query"))
	if query == "" {
		http.Error(w, "Missing parameter: 'query' is required", http.StatusBadRequest)
		return
	}

	// 3. Search port alias map
	results := SearchPorts(query)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(results)
}



