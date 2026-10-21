# LogisticsHQ — TASK 3: Maersk Integration Audit & Verification Report

**Document:** `MAERSK_INTEGRATION_AUDIT.md`  
**Date:** October 5, 2026  
**Auditor:** LogisticsHQ Core Engineering  
**Scope:** Maersk API Integration Audit & Readiness Verification  
**Status:** AUDIT COMPLETE — ZERO CODE MODIFICATIONS APPLIED  

---

## 1. Executive Summary

This audit evaluates the current state of the Maersk shipping line integration in LogisticsHQ (`freel-project`). The audit covers the adapter layer (`MaerskAdapter`), authentication and token lifecycle (`TokenManager`), transport resilience (`CarrierHTTPClient`), event/rate normalization (`normalizer.go`), database persistence, and end-to-end user journeys across Tracking, Rates, Bookings, and Documents.

### Key Takeaways:
1. **Solid Foundational Architecture:** The modular adapter pattern, DCSA v2 track-and-trace event normalization, AES-256-GCM credential encryption (Task 1), and capability enforcement (Task 2.1) are well designed, clean, and backed by robust unit and mock integration tests.
2. **Critical Authentication Disconnect (Action Required):** While `token.TokenManager` implements valid OAuth2 client credentials flow with thread-safe caching, `MaerskAdapter.buildHeaders()` **never invokes `tokenManager.GetToken()`**. As a result, operational requests (`GetTracking`, `GetRates`, `CreateBooking`, etc.) only attach `Consumer-Key` and `X-API-Key` headers without an `Authorization: Bearer <token>` header. Furthermore, the UI form stores credentials as `api_key` and `api_secret`, whereas `MaerskAdapter.TestConnection` checks for `client_id` and `client_secret`.
3. **Tracking & Rates are Functionally Built:** Tracking correctly normalizes DCSA v2 events and updates shipment status. Rates parses instant ocean spot quotes and contract rates. However, live API calls fail without the bearer token injection fix.
4. **Booking & Documents Require Wiring:** Booking creation issues requests to `/booking/v1/bookings`, but lacks idempotency keys and DCSA BKG payload translation. `GetDocuments` is implemented in the adapter but has no exposed HTTP transport route or UI trigger in LogisticsHQ.
5. **Sandbox Testing Verdict:** **NOT YET READY for live sandbox calls** until the OAuth token header injection and credential field mapping are corrected (estimated 1-2 hours in Task 3.1).

---

## 2. Maersk Architecture

### 2.1 File Location & Component Layout
- **Adapter Implementation:** [`backend/internal/carrier/adapters/maersk.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/maersk.go)
- **Token Management:** [`backend/internal/carrier/adapters/token/manager.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/token/manager.go)
- **HTTP Transport:** [`backend/internal/carrier/adapters/httpclient/client.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/httpclient/client.go)
- **Data Normalizers:** [`backend/internal/carrier/adapters/normalizer.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/normalizer.go)
- **Adapter Registry:** [`backend/internal/carrier/adapters/registry.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/registry.go)
- **Domain Models & Metadata:** [`backend/internal/carrier/domain/provider.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/domain/provider.go)
- **Service Orchestrator:** [`backend/internal/carrier/service/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/service/service.go)

### 2.2 Constructor & Identity
```go
func NewMaerskAdapter() *MaerskAdapter {
    return &MaerskAdapter{
        BaseAdapter: BaseAdapter{
            providerCode: "MAERSK",
            scac:         "MAEU",
            supportedCapabilities: []domain.CarrierCapability{
                domain.CapTracking,
                domain.CapRates,
                domain.CapSpotRates,
                domain.CapContractRates,
                domain.CapBooking,
                domain.CapDocuments,
            },
        },
        tokenManager: token.GetDefaultTokenManager(),
    }
}
```
- **Provider Code:** `MAERSK`
- **SCAC:** `MAEU` (also recognizes `MSKU` for container prefixes)
- **Capabilities Registered:** `TRACKING`, `RATES`, `SPOT_RATES`, `CONTRACT_RATES`, `BOOKING`, `DOCUMENTS`

---

## 3. Maersk Capability Matrix

| Capability | Implemented? | Real API Endpoint? | Tests? | Status | Audit Findings |
|---|---|---|---|---|---|
| **Connection Test** | YES | `GET /track-and-trace/v2/events` (ping) or OAuth token request | YES (`httptest`) | **PARTIAL** | Authenticates against token endpoint or gateway ping. Fails if credentials use `api_key`/`api_secret` instead of `client_id`/`client_secret`. |
| **Tracking** | YES | `GET /track-and-trace/v2/events` (DCSA v2) | YES (`httptest`) | **PARTIAL** | DCSA v2 parsing and event normalization complete. Blocked by missing `Authorization: Bearer` header. |
| **Spot Rates** | YES | `POST /ocean-spot/v1/instant-quotes` | NO (Mock only) | **PARTIAL** | Request payload, container equipment, and surcharge parsing implemented. Blocked by missing Bearer token. |
| **Contract Rates**| YES | `GET /rates/v1/contract-rates` | NO (Mock only) | **PARTIAL** | Contract number header and query construction implemented. Blocked by missing Bearer token. |
| **Booking** | YES | `POST /booking/v1/bookings` | NO (Mock only) | **PARTIAL** | Generic booking struct posted. Missing DCSA BKG v1 payload mapping and idempotency header. |
| **Booking Status**| YES | `GET /booking/v1/bookings/{ref}` | NO (Mock only) | **PARTIAL** | Wired in `CarrierBookingEngine.SyncCarrierBooking`, but blocked on live API by missing Bearer token. |
| **Documents** | YES | `GET /transportation-documents/v1/{ref}` | NO | **NOT WIRED** | Method exists in `MaerskAdapter`, but no backend HTTP route or frontend UI triggers it. |

---

## 4. Authentication Flow

### 4.1 Data & Credentials Pipeline
```text
FF Admin enters credentials in Carrier Settings UI
   ↓ (api_key, api_secret, base_url, client_id, client_secret)
Encrypted in MariaDB via AES-256-GCM (carrier_integrations.encrypted_credentials)
   ↓
Decrypted in backend memory (service.decryptCredentials)
   ↓
Passed to MaerskAdapter methods as domain.CarrierCredentials
   ↓
[CRITICAL GAP] buildHeaders() sets 'Consumer-Key' & 'X-API-Key' but DOES NOT call tokenManager.GetToken()
   ↓
Carrier API request sent without Bearer token (causes 401 Unauthorized against Maersk OAuth Gateway)
```

### 4.2 Credentials Required by Maersk Developer Portal
Maersk Developer Portal requires:
- **API Key / Consumer Key:** Passed in header `Consumer-Key: <consumer_key>`.
- **API Secret / Consumer Secret:** Used in OAuth2 client credentials flow:
  - `POST https://api.maersk.com/oauth2/access_token`
  - Headers: `Content-Type: application/x-www-form-urlencoded`
  - Body: `grant_type=client_credentials&client_id=<key>&client_secret=<secret>`
  - Response: `{"access_token": "...", "token_type": "Bearer", "expires_in": 3600}`
- **Subsequent API Calls:** Must send **BOTH**:
  - `Consumer-Key: <consumer_key>`
  - `Authorization: Bearer <access_token>`

### 4.3 Credential Field Mismatch
In [`domain/provider.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/domain/provider.go), the Maersk definition specifies form fields `api_key` and `api_secret`. In [`service/service.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/service/service.go), these map to `creds.APIKey` and `creds.APISecret`. But [`adapters/maersk.go:TestConnection`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/maersk.go#L58-L68) checks:
```go
if creds.ClientID != "" && creds.ClientSecret != "" {
    // requests token
}
```
If the user fills in the UI fields `api_key` and `api_secret`, `creds.ClientID` is empty, causing OAuth verification to be skipped.

---

## 5. Token Lifecycle & Cache / 401 Behavior

### 5.1 TokenManager Inspection (`adapters/token/manager.go`)
- **Cache Key:** `fmt.Sprintf("%s:%s:%s", req.ProviderCode, req.Environment, req.ClientID)`
- **Thread Safety:** Uses `sync.RWMutex` with double-checked locking around cache reads and writes.
- **Expiry Calculation & Safety Buffer:**
  - Token response `expires_in` (seconds) is parsed.
  - Expiry is stored as `time.Now().Add(time.Duration(resp.ExpiresIn) * time.Second)`.
  - **Safety Buffer:** When checking validity: `time.Until(entry.ExpiresAt) > 60*time.Second`. This correctly refreshes the token 60 seconds before expiration.
- **Cache Invalidation:** Exposes `Invalidate(providerCode, env, clientID)` which acquires `mu.Lock()` and deletes the cache entry.

### 5.2 401 Response Behavior (Defect Analysis)
1. **HTTP 401 Handling:** When an API request receives HTTP 401 Unauthorized, `CarrierHTTPClient` classifies it as `domain.ErrCodeAuthFailed` (non-retryable) and returns immediately.
2. **Missing Invalidation:** If a token is remotely revoked or expires prematurely, the cached token is **NOT invalidated**.
3. **No Automatic Retry on 401:** The client does not invalidate the cache, re-fetch a fresh token, and retry the operational request. The request simply fails.

---

## 6. Maersk Tracking

### 6.1 Flow & Endpoint Details
- **Endpoint:** `GET /track-and-trace/v2/events`
- **Query Parameters:** `carrierBookingReference={bkg}`, `equipmentReference={container}`
- **Headers:** `Accept: application/json`, `Consumer-Key: {apiKey}` (*Bearer token currently missing*)
- **Normalization:** Uses [`adapters.NormalizeDCSATrackingEvents`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/normalizer.go#L18-L113)
- **Supported DCSA Milestones:**
  - `RECE` / `GTIN` -> `GATE_IN`
  - `LOAD` -> `LOADED`
  - `DEPA` -> `DEPARTED`
  - `ARRI` -> `ARRIVED`
  - `DISC` -> `DISCHARGED`
  - `GTOT` -> `GATE_OUT`
  - `DELV` -> `DELIVERED`
- **CarrierTrackingEngine Integration:**
  - Handled in [`internal/shipments/carrier_tracking_engine.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/carrier_tracking_engine.go).
  - Events are saved into `shipment_tracking_events` with duplicate protection:
    `ON DUPLICATE KEY UPDATE raw_payload = VALUES(raw_payload)`.
  - Updates shipment status via non-regression status logic (`ShouldAdvanceStatus`).

### 6.2 Missing / Incorrect Handling
1. **Missing Bearer Token:** Live requests fail with 401 because `buildHeaders()` does not attach `Authorization: Bearer <token>`.
2. **Missing Pagination:** If tracking events exceed 100 entries, Maersk uses pagination headers (`Next-Page`), which are not currently traversed.

---

## 7. Maersk Spot Rates

### 7.1 Flow & Endpoint Details
- **Endpoint:** `POST /ocean-spot/v1/instant-quotes`
- **Headers:** `Content-Type: application/json`, `Consumer-Key: {apiKey}`
- **Payload Generated:**
```json
{
  "origin": {"unLocationCode": "INNSA"},
  "destination": {"unLocationCode": "NLRTM"},
  "cargoDetails": {
    "commodityDescription": "FAK",
    "equipmentType": "40HC"
  },
  "departureDate": "2026-10-15T00:00:00Z"
}
```
- **Normalization:** Uses [`NormalizeMaerskRates`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/normalizer.go#L119-L162).
- **Normalized `CarrierRate` Mapping:**
  - `RateID`: `quoteId`
  - `CarrierSCAC`: `MAEU`
  - `CarrierName`: `Maersk`
  - `BaseRate`: `totalPrice.amount`
  - `Currency`: `totalPrice.currency`
  - `TransitDays`: Parsed from `transitTimeDays`
  - `ValidUntil`: Parsed from `priceValidityEndDate`
  - `FreeDays`: Parsed from `demurrageAndDetention.freeDays`
  - `RateType`: `"SPOT"`

### 7.2 Missing / Gaps
- Commodities are hardcoded to `"FAK"` (Freight All Kinds) if not specified in `RateRequest`.
- Commodity code validation (HS code) is not passed to the API.

---

## 8. Maersk Contract Rates

### 8.1 Flow & Endpoint Details
- **Endpoint:** `GET /rates/v1/contract-rates`
- **Query Parameters:**
  - `origin={originPort}`
  - `destination={destPort}`
  - `equipmentType={equipmentType}`
- **Headers:**
  - `X-Contract-Number: {contractNumber}`
  - `Consumer-Key: {apiKey}`
- **Normalization:** Same `NormalizeMaerskRates` parser with `RateType: "CONTRACT"`.

### 8.2 Distinction from Spot Rates
- Spot rates use `POST` to `/ocean-spot/v1/instant-quotes` with real-time price validity windows.
- Contract rates use `GET` to `/rates/v1/contract-rates` and require a pre-negotiated customer contract number in `creds.ContractNumber` or the request metadata.

---

## 9. Maersk Booking

### 9.1 Flow & Endpoint Details
- **Endpoint:** `POST /booking/v1/bookings`
- **Payload Sent:** Directly marshals `domain.BookingRequest`:
```json
{
  "carrier_scac": "MAEU",
  "quote_id": "SPOT-12345",
  "contract_number": "MAEU-CT-991",
  "origin_port": "INNSA",
  "destination_port": "NLRTM",
  "equipment_type": "40HC",
  "cargo_description": "Electronics",
  "shipper_name": "Acme Forwarding",
  "consignee_name": "Global Retailers"
}
```
- **Response Parsing:**
  - Expects `bookingReference` / `bookingConfirmationNumber`.
  - Normalizes into `domain.BookingResult` with status `"PENDING"` or `"CONFIRMED"`.

### 9.2 Critical Deficiencies
1. **Schema Mismatch:** Maersk’s DCSA BKG v1 API does not accept flat snake_case JSON. It requires structured JSON with `declaredValue`, `cargoMovementTypeAtOrigin`, `commodities`, `requestedEquipments`, and `documentParties`.
2. **Missing Idempotency Key:** No `Idempotency-Key` or `X-Request-ID` is passed.

---

## 10. Booking Status

### 10.1 Flow & Endpoint Details
- **Endpoint:** `GET /booking/v1/bookings/{bookingReference}`
- **Response Normalized:** Maps `bookingStatus` (`CONFIRMED`, `PENDING`, `REJECTED`, `CANCELLED`) to `domain.BookingResult`.
- **System Wiring:** Fully wired in [`CarrierBookingEngine.SyncCarrierBooking`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/rfq/carrier_booking_engine.go#L112-L140). Updates `carrier_bookings.booking_status`, `vessel_name`, `voyage_number`, and `synced_at`.

---

## 11. Documents

### 11.1 Adapter Implementation
- **Endpoint:** `GET /transportation-documents/v1/{bookingReference}`
- **Method:** [`MaerskAdapter.GetDocuments`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/maersk.go#L292-L343)
- **Parsing:** Parses `documentType` (Bill of Lading, Sea Waybill), `documentNumber`, `downloadUrl`, `issueDate`.

### 11.2 Architectural Disconnect
- While `MaerskAdapter.GetDocuments` and `CarrierService.GetDocuments` exist, there is **NO HTTP endpoint in `carrier/transport/handler.go`** exposing it to the frontend, and **NO UI component** invokes it.
- **Verdict:** Backend adapter code exists, but document retrieval is disconnected from the product workflow.

---

## 12. Environment: Sandbox vs. Production

### 12.1 URL Resolution Logic (`adapters/maersk.go:getBaseURL`)
```go
func (a *MaerskAdapter) getBaseURL(creds domain.CarrierCredentials) string {
    if creds.BaseURL != "" {
        return creds.BaseURL
    }
    if creds.Environment == domain.EnvProduction {
        return "https://api.maersk.com"
    }
    return "https://api.maersk.com" // Sandbox default
}
```
### 12.2 Critical Vulnerability / Findings
1. Both `EnvProduction` and `EnvSandbox` default to `https://api.maersk.com`.
2. Maersk does not have a separate sandbox host like `api.sandbox.maersk.com`. Instead, Maersk uses different mock routes or tenant sandbox credentials on `https://api.maersk.com`.
3. If an FF enters sandbox credentials, but the endpoint does not differentiate paths, requests hit the live gateway. Without explicit sandbox indicators or credential validation, test calls could inadvertently interact with production services.

---

## 13. Error Handling & Secret Redaction

### 13.1 HTTP Status Code Handling (`CarrierHTTPClient`)
- **400 Bad Request:** Returns `ErrCodeInvalidRequest` with sanitized response message.
- **401 Unauthorized:** Returns `ErrCodeAuthFailed`. Non-retryable.
- **403 Forbidden:** Returns `ErrCodeForbidden`. Non-retryable.
- **404 Not Found:** Returns `ErrCodeNotFound`. Non-retryable.
- **429 Too Many Requests:** Returns `ErrCodeRateLimited`. **Retryable** (retried with exponential backoff).
- **500, 502, 503, 504:** Returns `ErrCodeServerError`. **Retryable**.
- **Context Timeout:** Returns `ErrCodeTimeout`. Non-retryable.

### 13.2 Credential Protection & Redaction
- In [`httpclient/client.go`](file:///c:/Users/Sai/go/src/freel-project/backend/internal/carrier/adapters/httpclient/client.go#L112-L135), error messages are wrapped in `domain.IntegrationError`.
- No `Authorization`, `Consumer-Key`, `ClientSecret`, or raw tokens are written to user-facing error strings or logs.
- Request correlation IDs (`X-Correlation-ID`) are logged instead of sensitive payloads.

---

## 14. Retry Behavior

### 14.1 Configuration
- **Max Retries:** 2 (Total attempts = 3).
- **Initial Backoff:** 100 milliseconds, doubling each retry (`100ms`, `200ms`).
- **Retry Condition:** Retried on network errors, HTTP 429, or HTTP status >= 500.

### 14.2 High-Risk Defect: POST Request Retries
- `CarrierHTTPClient.Execute` executes retries identically for **ALL HTTP methods**, including `POST`.
- If `POST /booking/v1/bookings` experiences a gateway timeout (`HTTP 504`) or connection drop *after* Maersk has already created the booking, `CarrierHTTPClient` automatically retries the identical POST.
- Because no idempotency key is included, **this can result in duplicate bookings and duplicate freight charges**.

---

## 15. Booking Idempotency & Safety

| Check | Status | Verification Detail |
|---|---|---|
| **Local Duplicate Prevention** | **PROTECTED** | `CarrierBookingEngine.SubmitCarrierBooking` checks if an active booking already exists for the shipment. If found and not cancelled, it returns an error: `"active carrier booking already exists"`. |
| **LogisticsHQ Idempotency Key** | **MISSING** | No unique UUID idempotency key is generated or passed to the adapter. |
| **Maersk Idempotency Header** | **MISSING** | `MaerskAdapter.CreateBooking` does not attach `Idempotency-Key` or `X-Request-ID`. |
| **HTTP Client POST Retry Guard**| **UNPROTECTED**| `CarrierHTTPClient` does not disable retries on `POST` booking operations. |

---

## 16. Existing Test Suite Analysis

### 16.1 Test Inventory
1. **`carrier_adapter_test.go:TestDCSATrackingNormalization`**
   - Validates DCSA v2 JSON parsing, equipment events (Gate-in, Loaded, Departed, Arrived), chronological sorting, and status assignment.
2. **`carrier_adapter_test.go:TestCarrierHTTPClientResilienceAndErrorMapping`**
   - Tests `CarrierHTTPClient` against `httptest.Server`: verifies 429 retries, immediate 401 failure, 403 forbidden translation, and backoff.
3. **`carrier_adapter_test.go:TestOAuth2TokenManagerCaching`**
   - Tests `TokenManager` against `httptest.Server`: validates token retrieval, in-memory caching, and TTL avoidance of duplicate calls.
4. **`carrier_adapter_test.go:TestMaerskAdapterOperationsAndCapabilityEnforcement`**
   - Comprehensive end-to-end integration test with real database and `httptest.Server`:
     - Connects Maersk with `TRACKING` capability.
     - Runs `TestConnection` against mock gateway.
     - Executes `GetTracking` with DCSA payload.
     - Verifies capability guard rejects `GetRates` when capability is disabled.
     - Tests disabled integration toggle.

### 16.2 Test Gaps
- No test verifying `MaerskAdapter.GetRates` JSON payload parsing.
- No test verifying `MaerskAdapter.CreateBooking` payload construction.
- No test verifying OAuth token attachment to operational calls (`GetTracking`, `GetRates`).
- No test for token expiration refresh or 401 recovery.

---

## 17. End-to-End User Journeys

### Journey A — Connect Maersk
`Settings → Add Maersk → Enter Credentials → Save → Test Connection → Connected`
- **Classification:** **PARTIAL**
- **Assessment:** Works when passing `base_url` or when credentials match internal expectations. However, fails if the user provides `api_key`/`api_secret` in the UI because `TestConnection` expects `client_id`/`client_secret` to trigger token validation.

### Journey B — Track Shipment
`Shipment → Tracking → Maersk API → Events → LogisticsHQ Timeline`
- **Classification:** **PARTIAL**
- **Assessment:** Data normalization, database persistence, and timeline rendering are 100% complete. However, real external calls to Maersk will fail with HTTP 401 because the adapter does not attach the OAuth Bearer token to the request headers.

### Journey C — Get Rates
`RFQ → Rate Search → Maersk → Rates → Normalized Results → Comparison`
- **Classification:** **PARTIAL**
- **Assessment:** Search engine queries all connected carriers with `CapRates` / `CapSpotRates`, ranks results by cheapest/fastest, and presents them in the UI. Blocked on live calls due to the Bearer token omission.

### Journey D — Book Shipment
`Shipment/RFQ → Book → Maersk → Booking Reference → LogisticsHQ Booking`
- **Classification:** **PARTIAL**
- **Assessment:** `CarrierBookingEngine` handles RFQ conversion and prevents local duplicate submissions. But payload mapping to DCSA BKG v1 is incomplete, and lack of idempotency keys makes retrying hazardous.

### Journey E — Documents
`Booking → Documents → Maersk → LogisticsHQ Document Result/UI`
- **Classification:** **NOT IMPLEMENTED**
- **Assessment:** While `MaerskAdapter.GetDocuments()` exists, there is no HTTP route, controller handler, or UI view to view or download documents from Maersk.

---

## 18. Production Readiness Scorecard

| Area | Status | Critical Remediation Needed |
|---|---|---|
| **Authentication Flow** | **NEEDS FIX** | Attach Bearer token to all operational requests; align UI form fields (`api_key`/`api_secret` vs `client_id`/`client_secret`). |
| **Credential Security** | **READY** | AES-256-GCM encryption with envelope security implemented in Task 1. |
| **Sandbox Support** | **NEEDS FIX** | Sandbox host defaults to production URL; requires clear sandbox endpoint/mock routing. |
| **Production Support** | **NEEDS FIX** | Production endpoints valid, but operational calls fail without Bearer tokens. |
| **Tracking Pipeline** | **READY (Code)** / **NEEDS FIX (Auth)** | DCSA v2 parsing and DB ingestion ready; requires auth header fix. |
| **Rates Pipeline** | **READY (Code)** / **NEEDS FIX (Auth)** | Spot and contract rate logic ready; requires auth header fix. |
| **Booking Pipeline** | **NEEDS FIX** | Requires DCSA BKG payload formatting and idempotency keys. |
| **Documents Pipeline**| **NOT IMPLEMENTED**| Backend method orphaned; requires transport route and UI. |
| **Error Handling** | **READY** | Standardized `IntegrationError`, status code mappings, and credential redaction verified. |
| **Retry & Idempotency**| **NEEDS FIX** | POST requests retried without idempotency keys; risks duplicate bookings. |
| **Token Lifecycle** | **NEEDS FIX** | Cache does not invalidate on 401; no 401 auto-refresh retry. |
| **Multi-Tenancy** | **READY** | Strict `org_id` scoping enforced across all database queries and token cache keys. |

---

## 19. Gaps Found

1. **Gap 1 (Critical): OAuth Bearer Token Not Attached to Operational Requests**  
   `MaerskAdapter.buildHeaders()` only sets `Consumer-Key` and `X-API-Key`. It never calls `tokenManager.GetToken()`, so all operational calls (`GetTracking`, `GetRates`, `CreateBooking`, etc.) are missing `Authorization: Bearer <token>`.
2. **Gap 2 (Critical): Credential Field Key Mismatch**  
   The UI configuration specifies `api_key` and `api_secret`, which deserialize into `creds.APIKey` and `creds.APISecret`. However, `MaerskAdapter.TestConnection` specifically inspects `creds.ClientID` and `creds.ClientSecret` to decide whether to fetch an OAuth token.
3. **Gap 3 (High): Remote Token Invalidation & 401 Recovery**  
   When a 401 Unauthorized is returned, `CarrierHTTPClient` treats it as a terminal error. The cached token is not cleared via `TokenManager.Invalidate()`, and the request is not retried with a newly acquired token.
4. **Gap 4 (High): Non-Idempotent Booking POST Retries**  
   `CarrierHTTPClient` retries 5xx and timeout errors on `POST /booking/v1/bookings` without an `Idempotency-Key` or `X-Request-ID` header, creating duplicate booking exposure.
5. **Gap 5 (Medium): Default Base URL Identical for Sandbox and Production**  
   `MaerskAdapter.getBaseURL` returns `https://api.maersk.com` regardless of whether `Environment` is `SANDBOX` or `PRODUCTION`.
6. **Gap 6 (Medium): Incomplete DCSA BKG v1 Schema Mapping**  
   `MaerskAdapter.CreateBooking` directly serializes the internal Go struct rather than transforming it into the official DCSA BKG schema expected by Maersk.
7. **Gap 7 (Low): Documents Disconnected from Transport and Frontend**  
   `GetDocuments` exists on the adapter and service, but lacks an HTTP handler route and frontend UI integration.

---

## 20. Recommended Fixes

1. **Fix Authentication Pipeline:**
   - Update `MaerskAdapter.buildHeaders(ctx, creds)` to take a context, retrieve an OAuth token via `a.tokenManager.GetToken(...)` if `creds.APISecret` or `creds.ClientSecret` is present, and add `Authorization: Bearer <token>`.
   - Update credential mapping in `service.decryptCredentials` or `MaerskAdapter` to treat `api_key`/`api_secret` interchangeably with `client_id`/`client_secret`.
2. **Implement 401 Invalidation & Refresh:**
   - In `CarrierHTTPClient` or `MaerskAdapter`, if an operational call receives HTTP 401, invalidate the token in `TokenManager`, request a fresh token, and retry the request once.
3. **Protect Booking Operations:**
   - Add an `Idempotency-Key: <uuid>` header to `MaerskAdapter.CreateBooking`.
   - Disable automatic network retries for `POST /booking/v1/bookings` inside `CarrierHTTPClient` unless an idempotency key is explicitly provided and honored by the carrier.
4. **Environment Differentiation:**
   - Clearly delineate sandbox vs. production endpoints or document that Maersk Sandbox requires mock gateway paths or sandbox tenant credentials.

---

## 21. Recommended Task 3.1 Scope

Task 3.1 should focus exclusively on **Maersk Authentication, Header Injection, and Sandbox Readiness**:
1. Fix credential field mapping so `api_key` and `api_secret` properly feed OAuth2 token acquisition.
2. Update `MaerskAdapter.buildHeaders` to inject `Authorization: Bearer <token>` into all operational API calls.
3. Add token cache invalidation and a single-retry mechanism when an HTTP 401 is encountered.
4. Add `Idempotency-Key` generation for `CreateBooking` and configure `CarrierHTTPClient` to safeguard `POST` operations.
5. Add unit and httptest integration tests verifying that bearer tokens, consumer keys, and idempotency headers are accurately constructed on all requests.

---
*Report prepared and verified against LogisticsHQ codebase. No code modifications were made during this audit.*
