# Task 1.3 — Production Service-Key Configuration Hardening Report

## 1. Original Vulnerability Summary
During the Post-Phase-7 Production Readiness Audit, an unsafe hardcoded development credential fallback was discovered in `backend/internal/notifications/service.go`:
```go
const DefaultServiceKey = "lhq_sec_dev_9a7d83f1c4e2b0a95e6f1837d28c4091a3b5c7e8f0123456789abcdef0123456"
```
When `INTERNAL_SERVICE_TOKEN` was absent or unset from the environment, `NewService()` silently assigned this hardcoded token as `s.serviceKey` for all sidecar requests via the `X-LogisticsHQ-Service-Key` header.
In addition, equivalent fallback behaviors were identified in:
- `backend/internal/middleware/internal_auth.go` (`devLocalOnlyFallbackKey`)
- `backend/internal/ai/sidecar_client.go`
- `backend/internal/reports/service.go`
- `backend/internal/copilot/service.go`
- `backend/internal/event_workflows/service.go`

This fallback was not gated to local development or testing, creating a serious security hazard where staging or production instances could inadvertently transmit or accept hardcoded development secrets.

---

## 2. Affected Files
- `backend/internal/notifications/service.go`
- `backend/internal/middleware/internal_auth.go`
- `backend/internal/config/config.go`
- `backend/cmd/server/main.go`
- `backend/internal/ai/sidecar_client.go`
- `backend/internal/reports/service.go`
- `backend/internal/copilot/service.go`
- `backend/internal/event_workflows/service.go`
- `backend/internal/notifications/service_key_test.go` (new comprehensive environment matrix test suite)
- `backend/internal/middleware/internal_auth_test.go` (new fail-closed verification test suite)

---

## 3. Root Cause
1. **Unconditioned Credential Fallbacks**: `NewService()` unconditionally assigned `key = DefaultServiceKey` if `os.Getenv("INTERNAL_SERVICE_TOKEN")` was empty, without verifying the target runtime environment (`APP_ENV`, `ENV`, or `ENVIRONMENT`).
2. **Asymmetric Inbound Verification**: In `internal_auth.go`, `ValidateInternalServiceToken()` checked `isProd := os.Getenv("APP_ENV") == "production"`. In any environment other than literal `"production"` (e.g. `staging`, `preview`, `sandbox`, or unset), it silently permitted `devLocalOnlyFallbackKey`.
3. **Configuration Struct Gap**: `backend/internal/config/config.go` did not expose a centralized, typed `InternalServiceToken` field on the `Config` struct.

---

## 4. Configuration Flow & Remediation

### 4.1. Centralized Config Exposure
In `backend/internal/config/config.go`:
- Added `InternalServiceToken string` to the `Config` struct.
- In `LoadConfig()`, resolved `INTERNAL_SERVICE_TOKEN` (falling back to `INTERNAL_SERVICE_KEY` and `AI_SIDECAR_SERVICE_KEY`), defaulting strictly to empty string `""` without hardcoded secrets.

### 4.2. Fail-Closed Service Key Resolution in `notifications/service.go`
- Implemented `ResolveServiceKey(explicitKey, env string) string`:
  - If an explicit key is supplied, use it.
  - If environment variables `INTERNAL_SERVICE_TOKEN`, `INTERNAL_SERVICE_KEY`, or `AI_SIDECAR_SERVICE_KEY` are populated, use the configured value.
  - **Environment Gating**: Inspect normalized environment (`APP_ENV`, `ENV`, `ENVIRONMENT`).
  - **Permitted Development Mode**: If and only if the environment is explicitly `"development"` or `"test"`, `DefaultServiceKey` is permitted for local developer convenience and offline unit tests.
  - **Fail-Closed Staging/Production**: For staging, production, unknown, or missing environments, `ResolveServiceKey` returns empty string `""`.
- Added options `WithEnvironment(env string)` and `WithServiceKey(key string)`.
- In `AnalyzeWithAI` and `GenerateDraftWithAI`:
  - Added immediate entry guards:
    ```go
    if s.serviceKey == "" {
        return nil, fmt.Errorf("sidecar communication disabled: internal service key is not configured")
    }
    ```
  - Outbound sidecar calls fail safely without leaking credentials or performing unauthenticated requests.

### 4.3. Hardened Machine-to-Machine Inbound Authentication (`internal_auth.go`)
- Updated `getExpectedInternalServiceToken()`:
  - Replaced unsafe `isProd := os.Getenv("APP_ENV") == "production"` with `isDevelopmentOrTestEnv()`.
  - In staging, production, unknown, or missing environments with no configured token, the middleware immediately halts request processing and returns HTTP 500 (`Internal service configuration error`).
  - `devLocalOnlyFallbackKey` cannot be activated outside development/test.

### 4.4. Harmonized Other Service Callers
- Updated `ai/sidecar_client.go`, `reports/service.go`, `copilot/service.go`, and `event_workflows/service.go` to enforce the identical environment gating so `DefaultServiceKey` is never substituted in staging or production.

### 4.5. Server Bootstrap
In `backend/cmd/server/main.go:289-293`:
- Pass `notifications.WithEnvironment(cfg.AppEnv)` and `notifications.WithServiceKey(cfg.InternalServiceToken)` to `notifications.NewService()`.

---

## 5. Environment Behavior Matrix

| Environment | Service Key Configured | Resolved Service Key | Outbound Sidecar Call | Inbound M2M Auth Status | Fallback Key Used |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Development** | Present (`secret-xyz`) | `secret-xyz` | `200 OK` | `200 OK` | No |
| **Development** | Absent (`""`) | `DefaultServiceKey` | Functional (Dev mode) | Functional (Dev fallback) | Yes (Permitted dev only) |
| **Test** | Present (`secret-xyz`) | `secret-xyz` | `200 OK` | `200 OK` | No |
| **Test** | Absent (`""`) | `DefaultServiceKey` | Functional (Test mode) | Functional (Test fallback) | Yes (Permitted test only) |
| **Staging** | Present (`secret-xyz`) | `secret-xyz` | `200 OK` | `200 OK` | No |
| **Staging** | Absent (`""`) | `""` (Empty) | Fails closed (Error) | `500 CONFIG_ERROR` | **BLOCKED** |
| **Production** | Present (`secret-xyz`) | `secret-xyz` | `200 OK` | `200 OK` | No |
| **Production** | Absent (`""`) | `""` (Empty) | Fails closed (Error) | `500 CONFIG_ERROR` | **BLOCKED** |
| **Unknown** (e.g. `sandbox`) | Absent (`""`) | `""` (Empty) | Fails closed (Error) | `500 CONFIG_ERROR` | **BLOCKED** |
| **Missing** (`""`) | Absent (`""`) | `""` (Empty) | Fails closed (Error) | `500 CONFIG_ERROR` | **BLOCKED** |

---

## 6. Secret Handling & Non-Disclosure Validation
- **No Hardcoded Production Secrets**: No new static secrets were added to any source file or committed configuration.
- **Log & Error Sanitization**: Error messages from `AnalyzeWithAI`, `GenerateDraftWithAI`, and `InternalServiceAuthMiddleware` use generic, secure descriptions (`"sidecar communication disabled: internal service key is not configured"` and `"Internal service configuration error"`).
- **Test-Verified Zero Leakage**: Automated tests confirmed that response bodies and error strings never contain secret substrings, development keys, or tokens.

---

## 7. Verification & Tests Executed

### 7.1. Notifications Service Key Tests (`service_key_test.go`)
Command: `go test -v ./internal/notifications/...`
```
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/development_+_key_present_->_works_with_configured_key
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/development_+_key_absent_->_follows_existing_safe_development_behavior
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/test_+_key_present_->_works_with_configured_key
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/test_+_key_absent_->_follows_existing_safe_test_behavior
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/staging_+_key_present_->_works_with_configured_key
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/staging_+_key_absent_->_fails_safely;_no_fallback
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/production_+_key_present_->_works_with_configured_key
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/production_+_key_absent_->_fails_safely;_no_fallback
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/unknown_environment_(sandbox)_+_key_absent_->_no_development_fallback
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/unknown_environment_(preview)_+_key_absent_->_no_development_fallback
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/missing_environment_+_key_absent_->_no_development_fallback
=== RUN   TestResolveServiceKey_FullEnvironmentMatrix/production_+_explicit_option_key_->_uses_explicit_key
--- PASS: TestResolveServiceKey_FullEnvironmentMatrix (0.00s)
=== RUN   TestNotifications_SidecarCallsFailClosedWithoutServiceKey
--- PASS: TestNotifications_SidecarCallsFailClosedWithoutServiceKey (0.00s)
=== RUN   TestDeduplicationHash
--- PASS: TestDeduplicationHash (0.00s)
=== RUN   TestNotificationCRUDAndIsolation
--- PASS: TestNotificationCRUDAndIsolation (0.01s)
=== RUN   TestEscalationTransitionAndAudit
--- PASS: TestEscalationTransitionAndAudit (0.01s)
=== RUN   TestNotificationPreferences
--- PASS: TestNotificationPreferences (0.00s)
=== RUN   TestEvaluationEngineDedup
--- PASS: TestEvaluationEngineDedup (0.04s)
=== RUN   TestAdvancedLifecycleAcknowledgeSnoozeAndEscalate
--- PASS: TestAdvancedLifecycleAcknowledgeSnoozeAndEscalate (0.01s)
PASS - github.com/freel/backend/internal/notifications
```

### 7.2. Middleware Internal Auth Tests (`internal_auth_test.go`)
Command: `go test -v ./internal/middleware/...`
```
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Production_with_missing_token_fails_closed_with_500_config_error
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Staging_with_missing_token_fails_closed_with_500_config_error
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Unknown_environment_with_missing_token_fails_closed
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Missing_environment_with_missing_token_fails_closed
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Production_with_valid_configured_token_succeeds
=== RUN   TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging/Production_with_invalid_token_rejected_with_401_Unauthorized
--- PASS: TestInternalServiceAuthMiddleware_FailClosedInProductionAndStaging (0.00s)
PASS - github.com/freel/backend/internal/middleware
```

### 7.3. Related Sidecar Clients & Workflows
Command: `go test -v ./internal/ai/... ./internal/copilot/... ./internal/reports/... ./internal/event_workflows/...`
```
PASS - github.com/freel/backend/internal/ai
PASS - github.com/freel/backend/internal/copilot
PASS - github.com/freel/backend/internal/reports
PASS - github.com/freel/backend/internal/event_workflows
```

### 7.4. Repository-Wide Compilation
Command: `go build ./...`
```
Exit Code: 0 (Zero errors)
```

---

## 8. Database and Business Behavior Integrity
- **Database**: Zero migrations modified; zero tables altered; no business data created, modified, or deleted.
- **Production Code**: No notification business logic, routing, or delivery semantics were changed.
- **Tests**: Zero tests were skipped, weakened, or removed.

---

## 9. Remaining Concerns
None. The service key configuration path is completely fail-closed outside explicitly designated development and test environments across all internal and outbound machine-to-machine communications.

---

## Final Status
**PASS — TASK 1.3 COMPLETE**
