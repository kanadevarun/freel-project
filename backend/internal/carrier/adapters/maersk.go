package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/freel/backend/internal/carrier/adapters/httpclient"
	"github.com/freel/backend/internal/carrier/adapters/token"
	"github.com/freel/backend/internal/carrier/domain"
)

// MaerskAdapter integrates A.P. Moller – Maersk Line API (SCAC: MAEU).
type MaerskAdapter struct {
	BaseAdapter
	tokenManager *token.TokenManager
}

// NewMaerskAdapter constructs the official Maersk carrier adapter.
func NewMaerskAdapter() *MaerskAdapter {
	return &MaerskAdapter{
		BaseAdapter: BaseAdapter{
			Code:        "MAERSK",
			CarrierSCAC: "MAEU",
			Capabilities: []domain.Capability{
				domain.CapTracking,
				domain.CapRates,
				domain.CapContractRates,
				domain.CapSpotRates,
				domain.CapBooking,
				domain.CapDocuments,
			},
		},
		tokenManager: token.GetDefaultTokenManager(),
	}
}

func (a *MaerskAdapter) getBaseURL(creds domain.DecryptedCredentials, env domain.Environment) string {
	if creds.BaseURL != "" {
		return creds.BaseURL
	}
	if env == domain.EnvSandbox {
		return "https://api.maersk.com"
	}
	return "https://api.maersk.com"
}

func (a *MaerskAdapter) getHTTPClient(creds domain.DecryptedCredentials, env domain.Environment) *httpclient.CarrierHTTPClient {
	return httpclient.NewCarrierHTTPClient(httpclient.HTTPClientConfig{
		BaseURL:        a.getBaseURL(creds, env),
		Timeout:        15 * time.Second,
		MaxRetries:     2,
		InitialBackoff: 200 * time.Millisecond,
		ProviderCode:   a.Code,
	})
}

// getCredentials normalizes credential field naming conventions for Maersk.
// It maps api_key/client_id to consumerKey, and api_secret/client_secret to apiSecret.
func (a *MaerskAdapter) getCredentials(creds domain.DecryptedCredentials) (consumerKey, apiSecret string) {
	return creds.GetConsumerKey(), creds.GetAPISecret()
}

// buildHeaders sets authentication headers conforming to Maersk API specifications.
// It injects Consumer-Key/X-API-Key and negotiates/injects the OAuth2 Bearer token if secrets are configured.
func (a *MaerskAdapter) buildHeaders(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (map[string]string, error) {
	headers := map[string]string{
		"Accept": "application/json",
	}

	consumerKey, apiSecret := a.getCredentials(creds)
	if consumerKey != "" {
		headers["Consumer-Key"] = consumerKey
		headers["X-API-Key"] = consumerKey
	}

	if consumerKey != "" && apiSecret != "" {
		tokenReq := token.TokenRequest{
			TokenURL:     a.getBaseURL(creds, env) + "/oauth2/access_token",
			ClientID:     consumerKey,
			ClientSecret: apiSecret,
			ProviderCode: a.Code,
			Environment:  env,
		}
		accessToken, err := a.tokenManager.GetToken(ctx, tokenReq)
		if err != nil {
			return nil, err
		}
		headers["Authorization"] = "Bearer " + accessToken
	}

	return headers, nil
}

// executeWithAuth executes an HTTP request against the Maersk API gateway.
// If the initial request fails with HTTP 401 Unauthorized and OAuth credentials are used,
// it automatically invalidates the cached token, requests a fresh OAuth2 bearer token,
// and retries the exact same request once. It will never retry more than once.
func (a *MaerskAdapter) executeWithAuth(
	ctx context.Context,
	creds domain.DecryptedCredentials,
	env domain.Environment,
	method, endpoint string,
	extraHeaders map[string]string,
	body []byte,
) ([]byte, error) {
	client := a.getHTTPClient(creds, env)

	headers, err := a.buildHeaders(ctx, creds, env)
	if err != nil {
		return nil, err
	}
	for k, v := range extraHeaders {
		headers[k] = v
	}

	respBody, statusCode, err := client.Execute(ctx, method, endpoint, headers, body)
	if err == nil {
		return respBody, nil
	}

	consumerKey, apiSecret := a.getCredentials(creds)

	// Check if the failure was an HTTP 401 Unauthorized and OAuth credentials are used
	if statusCode == http.StatusUnauthorized && consumerKey != "" && apiSecret != "" {
		failedToken := strings.TrimPrefix(headers["Authorization"], "Bearer ")
		a.tokenManager.InvalidateIfMatches(a.Code, consumerKey, env, failedToken)

		freshHeaders, tokenErr := a.buildHeaders(ctx, creds, env)
		if tokenErr != nil {
			// Token refresh failed; return sanitized authentication error without making a second operational request
			return nil, tokenErr
		}
		for k, v := range extraHeaders {
			freshHeaders[k] = v
		}

		// Single retry with the newly acquired Bearer token
		retryRespBody, _, retryErr := client.Execute(ctx, method, endpoint, freshHeaders, body)
		if retryErr != nil {
			// If second attempt also fails (e.g. 401 again), stop immediately and return error
			return nil, retryErr
		}
		return retryRespBody, nil
	}

	return nil, err
}

// TestConnection verifies connectivity and authentication with Maersk API gateway.
func (a *MaerskAdapter) TestConnection(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment) (*domain.TestConnectionResult, error) {
	start := time.Now()

	consumerKey, apiSecret := a.getCredentials(creds)
	if consumerKey == "" {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            "Missing Maersk API Key or Client ID. Please provide valid developer credentials.",
			LatencyMs:          time.Since(start).Milliseconds(),
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeInvalidConfig,
			HTTPStatus:         http.StatusBadRequest,
		}, nil
	}

	client := a.getHTTPClient(creds, env)

	// If OAuth secret is provided, verify authentication via OAuth token acquisition
	if apiSecret != "" {
		a.tokenManager.Invalidate(a.Code, consumerKey, env)
		_, err := a.tokenManager.GetToken(ctx, token.TokenRequest{
			TokenURL:     a.getBaseURL(creds, env) + "/oauth2/access_token",
			ClientID:     consumerKey,
			ClientSecret: apiSecret,
			ProviderCode: a.Code,
			Environment:  env,
		})
		latency := time.Since(start).Milliseconds()
		if err != nil {
			return &domain.TestConnectionResult{
				Success:            false,
				Message:            fmt.Sprintf("Maersk OAuth2 authentication failed: %s", domain.SanitizeError(err)),
				LatencyMs:          latency,
				TestedEnvironment:  env,
				TestedCapabilities: a.Capabilities,
				ErrorCode:          domain.ErrCodeAuthFailed,
				HTTPStatus:         http.StatusUnauthorized,
			}, nil
		}
		return &domain.TestConnectionResult{
			Success:            true,
			Message:            "Successfully authenticated with Maersk Developer Portal OAuth2 Gateway.",
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			HTTPStatus:         http.StatusOK,
		}, nil
	}

	// Direct API Key ping (fallback for API-key-only gateway configurations)
	headers, err := a.buildHeaders(ctx, creds, env)
	if err != nil {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            fmt.Sprintf("Failed to prepare headers: %s", domain.SanitizeError(err)),
			LatencyMs:          time.Since(start).Milliseconds(),
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeAuthFailed,
			HTTPStatus:         http.StatusUnauthorized,
		}, nil
	}

	testEndpoint := "/track-and-trace/v2/events?limit=1"
	_, statusCode, err := client.Execute(ctx, "GET", testEndpoint, headers, nil)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		return &domain.TestConnectionResult{
			Success:            false,
			Message:            fmt.Sprintf("Maersk API connection check returned: %s", domain.SanitizeError(err)),
			LatencyMs:          latency,
			TestedEnvironment:  env,
			TestedCapabilities: a.Capabilities,
			ErrorCode:          domain.ErrCodeAuthFailed,
			HTTPStatus:         statusCode,
		}, nil
	}

	return &domain.TestConnectionResult{
		Success:            true,
		Message:            "Successfully connected to Maersk API gateway.",
		LatencyMs:          latency,
		TestedEnvironment:  env,
		TestedCapabilities: a.Capabilities,
		HTTPStatus:         statusCode,
	}, nil
}

// GetTracking queries Maersk DCSA Track & Trace API v2.
func (a *MaerskAdapter) GetTracking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.TrackingRequest) (*domain.NormalizedTrackingResult, error) {
	if err := a.CheckCapability(domain.CapTracking); err != nil {
		return nil, err
	}

	queryParams := url.Values{}
	ref := req.ContainerNumber
	if req.ContainerNumber != "" {
		queryParams.Set("equipmentReference", req.ContainerNumber)
	} else if req.BookingNumber != "" {
		queryParams.Set("carrierBookingReference", req.BookingNumber)
		ref = req.BookingNumber
	} else if req.MBLNumber != "" {
		queryParams.Set("transportDocumentReference", req.MBLNumber)
		ref = req.MBLNumber
	} else {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInvalidRequest,
			"Container number, booking number, or MBL number must be provided for Maersk tracking",
			http.StatusBadRequest,
			false,
			fmt.Errorf("missing tracking identifier"),
		)
	}

	endpoint := fmt.Sprintf("/track-and-trace/v2/events?%s", queryParams.Encode())
	respBody, err := a.executeWithAuth(ctx, creds, env, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	var events []domain.DCSAEvent
	if err := json.Unmarshal(respBody, &events); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_TRACKING",
			domain.ErrCodeInternal,
			"Failed to deserialize Maersk Track & Trace events",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	normalized := NormalizeDCSATrackingEvents(a.CarrierSCAC, ref, events)
	return normalized, nil
}

// GetRates queries Maersk Spot & Tariff pricing engine.
func (a *MaerskAdapter) GetRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.RateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapRates); err != nil {
		return nil, err
	}

	bodyBytes, _ := json.Marshal(req)
	respBody, err := a.executeWithAuth(ctx, creds, env, "POST", "/rates/v1/spot-rates", nil, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// GetContractRates queries Maersk contracted commercial rates.
func (a *MaerskAdapter) GetContractRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.ContractRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapContractRates); err != nil {
		return nil, err
	}

	extraHeaders := make(map[string]string)
	if req.ContractNumber != "" {
		extraHeaders["X-Contract-Number"] = req.ContractNumber
	}

	bodyBytes, _ := json.Marshal(req)
	respBody, err := a.executeWithAuth(ctx, creds, env, "POST", "/rates/v1/contract-rates", extraHeaders, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// GetSpotRates queries Maersk instant spot pricing.
func (a *MaerskAdapter) GetSpotRates(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.SpotRateRequest) ([]domain.NormalizedCarrierRate, error) {
	if err := a.CheckCapability(domain.CapSpotRates); err != nil {
		return nil, err
	}

	bodyBytes, _ := json.Marshal(req)
	respBody, err := a.executeWithAuth(ctx, creds, env, "POST", "/ocean-spot/v1/instant-quotes", nil, bodyBytes)
	if err != nil {
		return nil, err
	}

	return NormalizeMaerskRates(a.CarrierSCAC, respBody)
}

// Maersk / DCSA Booking API Models

type MaerskDCSABookingPayload struct {
	CarrierBookingRequestReference string                     `json:"carrierBookingRequestReference,omitempty"`
	ContractQuotationReference     string                     `json:"contractQuotationReference,omitempty"`
	ReceiptTypeAtOrigin            string                     `json:"receiptTypeAtOrigin"`
	DeliveryTypeAtDestination       string                     `json:"deliveryTypeAtDestination"`
	CargoMovementTypeAtOrigin      string                     `json:"cargoMovementTypeAtOrigin"`
	CargoMovementTypeAtDestination string                     `json:"cargoMovementTypeAtDestination"`
	ServiceContractReference       string                     `json:"serviceContractReference,omitempty"`
	Commodities                    []MaerskDCSACommodity      `json:"commodities"`
	RequestedEquipments            []MaerskDCSARequestedEquip `json:"requestedEquipments"`
	DocumentParties                []MaerskDCSADocumentParty  `json:"documentParties"`
	Transports                     *MaerskDCSATransports      `json:"transports,omitempty"`
	PlaceOfReceipt                 *MaerskDCSALocation        `json:"placeOfReceipt,omitempty"`
	PlaceOfDelivery                *MaerskDCSALocation        `json:"placeOfDelivery,omitempty"`
}

type MaerskDCSACommodity struct {
	CommodityType    string `json:"commodityType"`
	CargoDescription string `json:"cargoDescription,omitempty"`
}

type MaerskDCSARequestedEquip struct {
	RequestedEquipmentTypeCode string `json:"requestedEquipmentTypeCode"`
	Units                      int    `json:"units"`
}

type MaerskDCSAParty struct {
	PartyName string `json:"partyName"`
}

type MaerskDCSADocumentParty struct {
	PartyFunction string          `json:"partyFunction"` // "OS" = Original Shipper, "CN" = Consignee
	Party         MaerskDCSAParty `json:"party"`
}

type MaerskDCSATransports struct {
	PlannedDepartureDate string              `json:"plannedDepartureDate,omitempty"`
	PortOfLoading        *MaerskDCSALocation `json:"portOfLoading,omitempty"`
	PortOfDischarge      *MaerskDCSALocation `json:"portOfDischarge,omitempty"`
}

type MaerskDCSALocation struct {
	UNLocationCode string `json:"unLocationCode,omitempty"`
	LocationName   string `json:"locationName,omitempty"`
}

type rawMaerskBookingResponse struct {
	CarrierBookingReference        string   `json:"carrierBookingReference"`
	CarrierBookingRequestReference string   `json:"carrierBookingRequestReference"`
	BookingStatus                  string   `json:"bookingStatus"`
	BookingReference               string   `json:"bookingReference"`
	BookingNumber                  string   `json:"booking_number"`
	Status                         string   `json:"status"`
	ConfirmationRef                string   `json:"confirmationRef"`
	ConfirmationReference          string   `json:"confirmation_ref"`
	VesselName                     string   `json:"vesselName"`
	VesselNameSnake                string   `json:"vessel_name"`
	VoyageNumber                   string   `json:"voyageNumber"`
	VoyageNumberSnake              string   `json:"voyage_number"`
	ExpectedDepartureDate          string   `json:"expectedDepartureDate"`
	ExpectedArrivalDate            string   `json:"expectedArrivalDate"`
	ETD                            string   `json:"etd"`
	ETA                            string   `json:"eta"`
	CutOffDate                     string   `json:"cutOffDate"`
	CutOffDateSnake                string   `json:"cut_off_date"`
	VGMDeadline                    string   `json:"vgmDeadline"`
	VGMDeadlineSnake               string   `json:"vgm_deadline"`
	EmptyPickupLocation            string   `json:"emptyPickupLocation"`
	EmptyPickupLocationSnake       string   `json:"empty_pickup_location"`
	AllocatedContainers            []string `json:"allocatedContainers"`
	AllocatedContainersSnake       []string `json:"allocated_containers"`
}

// validateBookingRequest validates mandatory fields required by the Maersk booking contract.
func (a *MaerskAdapter) validateBookingRequest(req domain.BookingRequest) error {
	if strings.TrimSpace(req.OriginPort) == "" {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "origin port is required for Maersk booking", http.StatusBadRequest, false, nil)
	}
	if strings.TrimSpace(req.DestinationPort) == "" {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "destination port is required for Maersk booking", http.StatusBadRequest, false, nil)
	}
	if strings.TrimSpace(req.EquipmentType) == "" {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "equipment type is required for Maersk booking", http.StatusBadRequest, false, nil)
	}
	if req.Quantity <= 0 {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "equipment quantity must be at least 1 for Maersk booking", http.StatusBadRequest, false, nil)
	}
	if strings.TrimSpace(req.ShipperName) == "" {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "shipper name is required for Maersk booking", http.StatusBadRequest, false, nil)
	}
	if strings.TrimSpace(req.ConsigneeName) == "" {
		return domain.NewIntegrationError(a.Code, "CREATE_BOOKING", domain.ErrCodeInvalidRequest, "consignee name is required for Maersk booking", http.StatusBadRequest, false, nil)
	}
	return nil
}

// generateIdempotencyKey derives a stable, tenant-safe idempotency key for Maersk booking operations.
func (a *MaerskAdapter) generateIdempotencyKey(req domain.BookingRequest) string {
	if req.IdempotencyKey != "" {
		return req.IdempotencyKey
	}
	if req.CustomerReference != "" {
		return fmt.Sprintf("idemp-maersk-%s", req.CustomerReference)
	}
	// Fallback to deterministic SHA-256 hash of core booking parameters (never random UUID across retries!)
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s|%s|%s|%s|%d|%s|%s",
		req.ContractNumber, req.OriginPort, req.DestinationPort,
		req.EquipmentType, req.Quantity, req.ShipperName, req.ConsigneeName)))
	return fmt.Sprintf("idemp-hash-%x", h.Sum(nil)[:16])
}

// buildMaerskBookingPayload converts internal LogisticsHQ BookingRequest to DCSA BKG payload.
func (a *MaerskAdapter) buildMaerskBookingPayload(req domain.BookingRequest) MaerskDCSABookingPayload {
	payload := MaerskDCSABookingPayload{
		CarrierBookingRequestReference: req.CustomerReference,
		ContractQuotationReference:     req.ContractNumber,
		ReceiptTypeAtOrigin:            "CY",
		DeliveryTypeAtDestination:       "CY",
		CargoMovementTypeAtOrigin:      "FCL",
		CargoMovementTypeAtDestination: "FCL",
		ServiceContractReference:       req.ContractNumber,
	}

	commodityDesc := req.Commodity
	if commodityDesc == "" {
		commodityDesc = "General Cargo"
	}
	payload.Commodities = []MaerskDCSACommodity{
		{
			CommodityType:    commodityDesc,
			CargoDescription: commodityDesc,
		},
	}

	qty := req.Quantity
	if qty <= 0 {
		qty = 1
	}
	eqType := req.EquipmentType
	if eqType == "" {
		eqType = "40HC"
	}
	payload.RequestedEquipments = []MaerskDCSARequestedEquip{
		{
			RequestedEquipmentTypeCode: eqType,
			Units:                      qty,
		},
	}

	var parties []MaerskDCSADocumentParty
	if req.ShipperName != "" {
		parties = append(parties, MaerskDCSADocumentParty{
			PartyFunction: "OS",
			Party: MaerskDCSAParty{
				PartyName: req.ShipperName,
			},
		})
	}
	if req.ConsigneeName != "" {
		parties = append(parties, MaerskDCSADocumentParty{
			PartyFunction: "CN",
			Party: MaerskDCSAParty{
				PartyName: req.ConsigneeName,
			},
		})
	}
	payload.DocumentParties = parties

	transports := &MaerskDCSATransports{}
	if !req.CargoReadyDate.IsZero() {
		transports.PlannedDepartureDate = req.CargoReadyDate.Format("2006-01-02")
	}
	if req.OriginPort != "" {
		transports.PortOfLoading = &MaerskDCSALocation{
			UNLocationCode: strings.ToUpper(strings.TrimSpace(req.OriginPort)),
			LocationName:   strings.ToUpper(strings.TrimSpace(req.OriginPort)),
		}
		payload.PlaceOfReceipt = transports.PortOfLoading
	}
	if req.DestinationPort != "" {
		transports.PortOfDischarge = &MaerskDCSALocation{
			UNLocationCode: strings.ToUpper(strings.TrimSpace(req.DestinationPort)),
			LocationName:   strings.ToUpper(strings.TrimSpace(req.DestinationPort)),
		}
		payload.PlaceOfDelivery = transports.PortOfDischarge
	}
	payload.Transports = transports

	return payload
}

// parseBookingResponse parses Maersk / DCSA booking confirmation JSON into NormalizedBookingResult.
func (a *MaerskAdapter) parseBookingResponse(respBody []byte) (*domain.NormalizedBookingResult, error) {
	var raw rawMaerskBookingResponse
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"CREATE_BOOKING",
			domain.ErrCodeInternal,
			"Failed to parse Maersk booking confirmation",
			http.StatusInternalServerError,
			false,
			err,
		)
	}

	bkgNum := raw.CarrierBookingReference
	if bkgNum == "" {
		bkgNum = raw.BookingReference
	}
	if bkgNum == "" {
		bkgNum = raw.BookingNumber
	}
	if bkgNum == "" {
		bkgNum = raw.CarrierBookingRequestReference
	}

	status := raw.BookingStatus
	if status == "" {
		status = raw.Status
	}
	if status == "" {
		status = "CONFIRMED"
	}

	confRef := raw.ConfirmationRef
	if confRef == "" {
		confRef = raw.ConfirmationReference
	}
	if confRef == "" {
		confRef = bkgNum
	}

	vessel := raw.VesselName
	if vessel == "" {
		vessel = raw.VesselNameSnake
	}
	voyage := raw.VoyageNumber
	if voyage == "" {
		voyage = raw.VoyageNumberSnake
	}
	pickupLoc := raw.EmptyPickupLocation
	if pickupLoc == "" {
		pickupLoc = raw.EmptyPickupLocationSnake
	}
	containers := raw.AllocatedContainers
	if len(containers) == 0 {
		containers = raw.AllocatedContainersSnake
	}

	res := &domain.NormalizedBookingResult{
		BookingNumber:       bkgNum,
		CarrierSCAC:         a.CarrierSCAC,
		Status:              strings.ToUpper(status),
		ConfirmationRef:     confRef,
		VesselName:          vessel,
		VoyageNumber:        voyage,
		EmptyPickupLocation: pickupLoc,
		AllocatedContainers: containers,
	}

	etdStr := raw.ExpectedDepartureDate
	if etdStr == "" {
		etdStr = raw.ETD
	}
	if etdStr != "" {
		t := domain.ParseDCSATime(etdStr)
		if !t.IsZero() {
			res.ETD = &t
		}
	}

	etaStr := raw.ExpectedArrivalDate
	if etaStr == "" {
		etaStr = raw.ETA
	}
	if etaStr != "" {
		t := domain.ParseDCSATime(etaStr)
		if !t.IsZero() {
			res.ETA = &t
		}
	}

	cutOffStr := raw.CutOffDate
	if cutOffStr == "" {
		cutOffStr = raw.CutOffDateSnake
	}
	if cutOffStr != "" {
		t := domain.ParseDCSATime(cutOffStr)
		if !t.IsZero() {
			res.CutOffDate = &t
		}
	}

	vgmStr := raw.VGMDeadline
	if vgmStr == "" {
		vgmStr = raw.VGMDeadlineSnake
	}
	if vgmStr != "" {
		t := domain.ParseDCSATime(vgmStr)
		if !t.IsZero() {
			res.VGMDeadline = &t
		}
	}

	return res, nil
}

// CreateBooking creates a new container space booking with Maersk using DCSA BKG standard payload and stable idempotency key.
func (a *MaerskAdapter) CreateBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.BookingRequest) (*domain.NormalizedBookingResult, error) {
	if err := a.CheckCapability(domain.CapBooking); err != nil {
		return nil, err
	}

	// 1. Mandatory Pre-Flight Validation (zero HTTP requests if invalid)
	if err := a.validateBookingRequest(req); err != nil {
		return nil, err
	}

	// 2. Stable Idempotency-Key Generation
	idempotencyKey := a.generateIdempotencyKey(req)

	// 3. Construct DCSA Booking Payload
	dcsaPayload := a.buildMaerskBookingPayload(req)
	bodyBytes, err := json.Marshal(dcsaPayload)
	if err != nil {
		return nil, domain.NewIntegrationError(
			a.Code,
			"CREATE_BOOKING",
			domain.ErrCodeInvalidRequest,
			"Failed to serialize Maersk booking payload",
			http.StatusBadRequest,
			false,
			err,
		)
	}

	// 4. Inject Idempotency-Key header
	extraHeaders := map[string]string{
		"Idempotency-Key": idempotencyKey,
	}

	// 5. Execute request with Auth + 401 refresh safety + retry idempotency
	respBody, err := a.executeWithAuth(ctx, creds, env, "POST", "/booking/v1/bookings", extraHeaders, bodyBytes)
	if err != nil {
		return nil, err
	}

	return a.parseBookingResponse(respBody)
}

// GetBooking fetches status of an existing Maersk booking.
func (a *MaerskAdapter) GetBooking(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, bookingRef string) (*domain.NormalizedBookingResult, error) {
	if err := a.CheckCapability(domain.CapBooking); err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("/booking/v1/bookings/%s", bookingRef)
	respBody, err := a.executeWithAuth(ctx, creds, env, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	return a.parseBookingResponse(respBody)
}

type rawMaerskDocument struct {
	DocumentID                 string `json:"documentID"`
	DocumentIDSnake            string `json:"document_id"`
	ID                         string `json:"id"`
	DocumentType               string `json:"documentType"`
	DocumentTypeSnake          string `json:"document_type"`
	Type                       string `json:"type"`
	DocumentNumber             string `json:"documentNumber"`
	DocumentNumberSnake        string `json:"document_number"`
	Reference                  string `json:"reference"`
	TransportDocumentReference string `json:"transportDocumentReference"`
	CarrierBookingReference    string `json:"carrierBookingReference"`
	DownloadURL                string `json:"downloadUrl"`
	DownloadURLSnake           string `json:"download_url"`
	DocumentURL                string `json:"documentUrl"`
	DocumentURLSnake           string `json:"document_url"`
	URL                        string `json:"url"`
	MimeType                   string `json:"mimeType"`
	MimeTypeSnake              string `json:"mime_type"`
	ContentType                string `json:"contentType"`
	IssueDate                  string `json:"issueDate"`
	IssueDateSnake             string `json:"issue_date"`
	CreatedAt                  string `json:"createdAt"`
	CreatedAtSnake             string `json:"created_at"`
}

func (a *MaerskAdapter) normalizeMaerskDocument(raw rawMaerskDocument, fallbackRef string) domain.NormalizedDocumentResult {
	docID := raw.DocumentID
	if docID == "" {
		docID = raw.DocumentIDSnake
	}
	if docID == "" {
		docID = raw.ID
	}
	if docID == "" {
		docID = raw.DocumentNumber
	}
	if docID == "" {
		docID = raw.Reference
	}
	if docID == "" {
		docID = fmt.Sprintf("doc-%s", fallbackRef)
	}

	docType := raw.DocumentType
	if docType == "" {
		docType = raw.DocumentTypeSnake
	}
	if docType == "" {
		docType = raw.Type
	}
	if docType == "" {
		docType = "TRANSPORT_DOCUMENT"
	}
	docType = strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(docType), " ", "_"))

	ref := raw.DocumentNumber
	if ref == "" {
		ref = raw.DocumentNumberSnake
	}
	if ref == "" {
		ref = raw.Reference
	}
	if ref == "" {
		ref = raw.TransportDocumentReference
	}
	if ref == "" {
		ref = raw.CarrierBookingReference
	}
	if ref == "" {
		ref = fallbackRef
	}

	dlURL := raw.DownloadURL
	if dlURL == "" {
		dlURL = raw.DownloadURLSnake
	}
	if dlURL == "" {
		dlURL = raw.DocumentURL
	}
	if dlURL == "" {
		dlURL = raw.DocumentURLSnake
	}
	if dlURL == "" {
		dlURL = raw.URL
	}

	mime := raw.MimeType
	if mime == "" {
		mime = raw.MimeTypeSnake
	}
	if mime == "" {
		mime = raw.ContentType
	}
	if mime == "" {
		mime = "application/pdf"
	}

	dateStr := raw.IssueDate
	if dateStr == "" {
		dateStr = raw.IssueDateSnake
	}
	if dateStr == "" {
		dateStr = raw.CreatedAt
	}
	if dateStr == "" {
		dateStr = raw.CreatedAtSnake
	}

	createdAt := domain.ParseDCSATime(dateStr)
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}

	return domain.NormalizedDocumentResult{
		DocumentID:   docID,
		CarrierSCAC:  a.CarrierSCAC,
		DocumentType: docType,
		Reference:    ref,
		DocumentURL:  dlURL,
		MimeType:     mime,
		CreatedAt:    createdAt,
	}
}

func (a *MaerskAdapter) parseDocumentsResponse(respBody []byte, fallbackRef string) ([]domain.NormalizedDocumentResult, error) {
	trimmed := strings.TrimSpace(string(respBody))
	if trimmed == "" || trimmed == "[]" {
		return []domain.NormalizedDocumentResult{}, nil
	}

	var rawList []rawMaerskDocument
	if strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(respBody, &rawList); err != nil {
			return nil, domain.NewIntegrationError(
				a.Code,
				"GET_DOCUMENTS",
				domain.ErrCodeInternal,
				"Failed to parse Maersk documents response",
				http.StatusInternalServerError,
				false,
				err,
			)
		}
	} else if strings.HasPrefix(trimmed, "{") {
		var single rawMaerskDocument
		if err := json.Unmarshal(respBody, &single); err != nil {
			return nil, domain.NewIntegrationError(
				a.Code,
				"GET_DOCUMENTS",
				domain.ErrCodeInternal,
				"Failed to parse Maersk documents response",
				http.StatusInternalServerError,
				false,
				err,
			)
		}
		rawList = append(rawList, single)
	} else {
		return []domain.NormalizedDocumentResult{}, nil
	}

	results := make([]domain.NormalizedDocumentResult, 0, len(rawList))
	for _, item := range rawList {
		results = append(results, a.normalizeMaerskDocument(item, fallbackRef))
	}
	return results, nil
}

// GetDocuments retrieves transport documents from Maersk API.
func (a *MaerskAdapter) GetDocuments(ctx context.Context, creds domain.DecryptedCredentials, env domain.Environment, req domain.DocumentRequest) ([]domain.NormalizedDocumentResult, error) {
	if err := a.CheckCapability(domain.CapDocuments); err != nil {
		return nil, err
	}

	if strings.TrimSpace(req.BookingNumber) == "" {
		return nil, domain.NewIntegrationError(
			a.Code,
			"GET_DOCUMENTS",
			domain.ErrCodeInvalidRequest,
			"booking number or document reference is required to retrieve carrier documents",
			http.StatusBadRequest,
			false,
			nil,
		)
	}

	endpoint := fmt.Sprintf("/transportation-documents/v1/%s", url.PathEscape(strings.TrimSpace(req.BookingNumber)))
	respBody, err := a.executeWithAuth(ctx, creds, env, "GET", endpoint, nil, nil)
	if err != nil {
		return nil, err
	}

	return a.parseDocumentsResponse(respBody, req.BookingNumber)
}
