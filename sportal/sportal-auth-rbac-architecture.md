# SPortal Authentication & RBAC Architecture Specification
## LogisticsHQ Internal SaaS Administration Platform — Security & Authorization Boundary

---

## 1. Executive Summary

**SPortal** is the internal control center of the **LogisticsHQ** platform. While **CPortal** serves external freight-forwarding customer organizations and their employees, **SPortal** is strictly reserved for the LogisticsHQ internal team (Founders, Customer Success, Platform Engineering, Finance, and Operations).

Task S2 establishes the complete, production-oriented authentication, internal role-based access control (RBAC), and server-side security boundary for SPortal, anchored by the following principles:

1. **Strict Identity & Application Boundary**: A customer organization user—regardless of whether they possess `SUPER_ADMIN`, `ADMIN`, or broad logistics operational permissions within their tenant—cannot log into or access SPortal.
2. **Shared Data Source of Truth**: SPortal uses the authoritative MariaDB tables (`users`, `organizations`, `org_members`, `roles`, `permissions`) via the shared Go backend. Zero duplicate tables (`sportal_users`, `sportal_roles`, `sportal_passwords`) were created.
3. **Server-Side Enforcement**: Security is enforced at the HTTP middleware and service layer. Client-side route guards and UI state are purely decorative and supplementary.
4. **Sensitive Data Segregation**: Access to platform secrets, bank escrow accounts, tax PAN/GST credentials, and provider API keys is guarded by explicit sensitive permissions (`billing:sensitive_view`) and audited to an immutable ledger.

---

## 2. Product & Identity Distinction: Internal vs. Customer Users

```
                            LogisticsHQ Platform (MariaDB)
                                          │
                  ┌───────────────────────┴───────────────────────┐
                  ▼                                               ▼
     LogisticsHQ Internal Tenant                      Customer Tenant Organizations
             (Org ID = 1)                               (Org ID 2, 3, 4, ... 1809)
                  │                                               │
        ┌─────────┴─────────┐                           ┌─────────┴─────────┐
        ▼                   ▼                           ▼                   ▼
    Internal Roles      Internal Staff               Customer Roles     Tenant Users
   - SUPER_ADMIN       - Varun (CEO)                - SUPER_ADMIN      - Forwarder Admin
   - CEO               - Platform Admin             - SALES            - Sales Rep
   - OWNER             - Finance Lead               - PRICING          - Pricing Analyst
   - CUSTOMER_SUCCESS  - Support Engineer           - OPERATIONS       - Dispatcher
   - FINANCE                                        - CUSTOMER_CONTACT - Cargo Shipper
   - SUPPORT                                                      │
   - OPERATIONS                                                   ▼
                  │                                         CPortal Only
                  ▼                                     (app.logisticshq.in)
             SPortal Only                               [BLOCKED FROM SPORTAL]
       (sportal.logisticshq.in)
```

### Identity Boundary Matrix

| User Concept | Organization Scope | Role Authority | CPortal Access | SPortal Access |
| :--- | :--- | :--- | :--- | :--- |
| **LogisticsHQ Founder / CEO** | Org ID = 1 | `CEO` / `SUPER_ADMIN` | Permitted (Authoritative) | **Allowed (Full SPortal Access)** |
| **LogisticsHQ Finance Lead** | Org ID = 1 | `FINANCE` | Permitted (Financials) | **Allowed (Billing, Subscriptions, Orgs)** |
| **LogisticsHQ Customer Success** | Org ID = 1 | `CUSTOMER_SUCCESS` | Permitted (Supervisory) | **Allowed (Onboarding, Health, Usage)** |
| **Customer Freight Forwarder Admin** | Org ID > 1 (e.g. Org 2) | `SUPER_ADMIN` | **Allowed (Tenant Ops)** | **DENIED (HTTP 403 Forbidden)** |
| **Customer Operations / Sales Rep** | Org ID > 1 (e.g. Org 2) | `SALES`, `OPERATIONS` | **Allowed (Tenant Ops)** | **DENIED (HTTP 403 Forbidden)** |
| **External Shipper / Cargo Client** | Org ID > 1 | `CUSTOMER_CONTACT` | Restricted Tracking Portal | **DENIED (HTTP 403 Forbidden)** |

---

## 3. SPortal RBAC Model & Canonical Permissions

SPortal defines 26 canonical internal permissions organized into 14 functional categories:

```go
// Canonical SPortal Permissions
const (
    // Organizations
    PermOrgsView    = "organizations:view"
    PermOrgsCreate  = "organizations:create"
    PermOrgsUpdate  = "organizations:update"
    PermOrgsArchive = "organizations:archive"

    // Onboarding
    PermOnboardingView     = "onboarding:view"
    PermOnboardingCreate   = "onboarding:create"
    PermOnboardingUpdate   = "onboarding:update"
    PermOnboardingComplete = "onboarding:complete"

    // Subscriptions
    PermSubscriptionsView   = "subscriptions:view"
    PermSubscriptionsCreate = "subscriptions:create"
    PermSubscriptionsUpdate = "subscriptions:update"
    PermSubscriptionsCancel = "subscriptions:cancel"

    // Billing & Finance
    PermBillingView          = "billing:view"
    PermBillingManage        = "billing:manage"
    PermBillingSensitiveView = "billing:sensitive_view" // Guard for bank, PAN, GST secrets

    // Users & Roles
    PermUsersView    = "users:view"
    PermUsersCreate  = "users:create"
    PermUsersUpdate  = "users:update"
    PermUsersDisable = "users:disable"

    // Customer Operations & Telemetry
    PermCustomerOpsView    = "customer_operations:view"
    PermUsageView          = "usage:view"
    PermCustomerHealthView = "customer_health:view"

    // Integrations & Documents
    PermIntegrationsView   = "integrations:view"
    PermIntegrationsManage = "integrations:manage"
    PermDocumentsView      = "documents:view"
    PermDocumentsManage    = "documents:manage"

    // Support & Audit
    PermSupportView   = "support:view"
    PermSupportManage = "support:manage"
    PermAuditView     = "audit:view"

    // Settings, AI & Platform
    PermSettingsView       = "settings:view"
    PermSettingsManage     = "settings:manage"
    PermAIView             = "ai:view"
    PermAIUse              = "ai:use"
    PermAIManage           = "ai:manage"
    PermPlatformHealthView = "platform_health:view"
    PermPlatformAdmin      = "platform_admin"
)
```

### Role-to-Permission Mapping Matrix

| Permission Category | Super Admin / CEO | Platform Admin | Finance | Customer Success | Support | Operations | Technical | Customer Roles |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| `organizations:view` | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `organizations:create/update` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `onboarding:view` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `subscriptions:view/manage` | ✅ | ✅ | ✅ | ✅ (View) | ❌ | ❌ | ❌ | ❌ |
| `billing:view/manage` | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ |
| **`billing:sensitive_view`** | **✅** | **❌** | **✅** | **❌** | **❌** | **❌** | **❌** | **❌** |
| `users:view/manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ❌ | ❌ |
| `customer_operations:view` | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `usage:view` | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ | ❌ |
| `customer_health:view` | ✅ | ✅ | ❌ | ✅ | ❌ | ❌ | ❌ | ❌ |
| `integrations:view/manage` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| `support:view/manage` | ✅ | ✅ | ❌ | ✅ (View) | ✅ | ❌ | ❌ | ❌ |
| `audit:view` | ✅ | ✅ | ✅ | ❌ | ✅ | ✅ | ✅ | ❌ |
| `platform_health:view` | ✅ | ✅ | ❌ | ❌ | ❌ | ❌ | ✅ | ❌ |
| `ai:view/use/manage` | ✅ | ✅ | ❌ | ✅ (Use) | ✅ (Use) | ❌ | ✅ | ❌ |

---

## 4. Server-Side Enforcement Architecture

All SPortal backend endpoints are mounted under `/api/v1/sportal/` and protected by a two-stage middleware pipeline:

```
Incoming Request
    │
    ▼
[stage 1: AuthGuard.RequireAuth]
    - Extracts Bearer token from Authorization header
    - Parses AWS Cognito JWT or validates test-token in development/test
    - Resolves UserContext { UserID, OrgID, Role, CognitoID } from database
    - Fails with HTTP 401 Unauthorized if missing/invalid
    │
    ▼
[stage 2: sportal.RequireInternalStaff]
    - Verifies userCtx.OrgID == 1 (LogisticsHQ Internal Tenant)
    - Verifies sportal.IsInternalStaffRole(userCtx.Role) == true
    - If userCtx.OrgID != 1 (Customer Tenant):
        * Records security audit log: Action="sportal.forbidden_access_attempt"
        * Returns HTTP 403 Forbidden: "Forbidden: Only authorized LogisticsHQ internal personnel may access SPortal administration"
    │
    ▼
[stage 3: sportal.RequirePermission(perm)] (optional per-endpoint guard)
    - Checks sportal.RoleHasPermission(userCtx.Role, requiredPerm)
    - If missing permission:
        * Records security audit log: Action="sportal.permission_denied"
        * Returns HTTP 403 Forbidden: "Forbidden: Role '...' lacks required permission '...'"
    │
    ▼
[Handler Execution]
    - Serves authorized business logic
```

---

## 5. Authentication Flow & Session Lifecycle

1. **Login Request (`POST /api/v1/sportal/auth/login`)**:
   - User submits `email` and `password`.
   - Server resolves database record:
     ```sql
     SELECT u.id, u.email, u.first_name, u.last_name, o.id as org_id, o.name as org_name,
            r.name as role_name, om.status as membership_status
     FROM users u
     JOIN org_members om ON u.id = om.user_id
     JOIN organizations o ON om.org_id = o.id
     LEFT JOIN roles r ON om.role_id = r.id
     WHERE u.email = ?
     ```
   - **Boundary Enforcement**: If `org_id != 1` or role is not internal staff:
     Returns HTTP 403 Forbidden with `code: ACCESS_DENIED`.
   - **Status Check**: If `membership_status != "ACTIVE"`:
     Returns HTTP 403 Forbidden with `Account suspended or inactive`.
   - **Success**: Returns `access_token`, user profile, organization metadata, and the resolved permissions array.
   - **Audit**: Emits `ActionLogin` to the unified audit ledger.

2. **Session Hydration (`GET /api/v1/sportal/auth/me`)**:
   - Validates Bearer token via `AuthGuard`.
   - Checks internal staff boundary via `RequireInternalStaff`.
   - Returns fresh user profile, role, and permission array.
   - If token is expired or revoked: returns HTTP 401.

3. **Logout (`POST /api/v1/sportal/auth/logout`)**:
   - Emits `ActionLogout` to the audit ledger.
   - Frontend clears `sportal_access_token` and `sportal_session_user` from localStorage and redirects to `/login`.

---

## 6. Sensitive Data Protection

Protected financial endpoints (e.g. `GET /api/v1/sportal/finance/sensitive`) contain gateway account IDs, escrow settlement nodes, and tax credentials.

- **Mandatory Permission**: Must have `billing:sensitive_view`.
- **Restricted Access**: Available only to `SUPER_ADMIN`, `CEO`, `OWNER`, and `FINANCE`.
- **Blocked Roles**: `ADMIN`, `CUSTOMER_SUCCESS`, `SUPPORT`, `OPERATIONS`, `TECHNICAL`, and all customer tenants receive HTTP 403 Forbidden.
- **Audit Requirement**: Every access to sensitive financial records generates an immutable audit record (`Action: sportal.sensitive_financial_access`, `Module: PAYMENTS`).

---

## 7. Audit Logging Integration

SPortal security events integrate directly into the existing universal audit system (`backend/internal/audit`):

| SPortal Event | Audit Action | Audit Module | Result |
| :--- | :--- | :--- | :--- |
| **Internal Staff Login** | `domain.ActionLogin` | `AUTHENTICATION` | `SUCCESS` |
| **Failed Credentials** | `domain.ActionLoginFailed` | `AUTHENTICATION` | `FAILED` |
| **Customer Tenant Block** | `sportal.access_denied` | `AUTHENTICATION` | `FAILED` |
| **Staff Logout** | `domain.ActionLogout` | `AUTHENTICATION` | `SUCCESS` |
| **API Boundary Violation** | `sportal.forbidden_access_attempt` | `AUTHENTICATION` | `FAILED` |
| **Permission Denied** | `sportal.permission_denied` | `ROLES_PERMISSIONS` | `FAILED` |
| **Sensitive Data Access** | `sportal.sensitive_financial_access` | `PAYMENTS` | `SUCCESS` |
