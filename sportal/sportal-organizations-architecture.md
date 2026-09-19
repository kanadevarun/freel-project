# SPortal Organizations & Customer 360 Architecture

## 1. Authoritative Organization Entity
The authoritative business entity representing a freight-forwarding customer organization across the LogisticsHQ platform is the canonical MariaDB `organizations` table.
- **Table**: `organizations`
- **Identifier**: `id` (BIGINT, Primary Key)
- **Role**: Serves as the authoritative tenant boundary for all operational domains (shipments, bookings, RFQs, quotes, invoices, compliance, and user memberships).
- **Zero Duplication Policy**: Neither `sportal_organizations`, `sportal_customers`, nor `sportal_companies` were created. SPortal and CPortal directly access this single authoritative record.

---

## 2. Organization / Customer Relationship
In LogisticsHQ:
- An **Organization** is the freight forwarding company subscribing to LogisticsHQ SaaS (the tenant).
- A **Customer** (in table `customers`) is an importer, exporter, shipper, or consignee that does business with that freight forwarding organization.
- **Tenant Context**: All customer entities (`customers.org_id`), forwarder staff users (`org_members.org_id`), shipments (`shipments.org_id`), quotes (`quotes.org_id`), bookings (`bookings.org_id`), and invoices (`invoices.org_id`) carry `org_id` referencing `organizations.id`.
- **Org 1**: Represents the platform administrator/internal LogisticsHQ organization context.

---

## 3. Database Tables Used
The SPortal Organizations management domain directly queries and updates existing MariaDB tables:
1. `organizations`: Core forwarder entity profile (`id`, `name`, `legal_name`, `registration_number`, `tax_number`, `website`, `primary_email`, `phone_number`, `address`, `city`, `state`, `country`, `postal_code`, `industry`, `company_type`, `created_at`, `updated_at`).
2. `org_members`: Resolves staff user membership, forwarder assignments, status, and role bindings (`org_id`, `user_id`, `role_id`, `status`).
3. `roles`: Role definitions for member users (`id`, `name`, `description`).
4. `users`: Identity records for tenant staff (`id`, `email`, `first_name`, `last_name`, `status`).
5. `organization_subscriptions`: Active subscription binding (`org_id`, `plan_id`, `status`, `billing_cycle`, `current_period_end`).
6. `subscription_plans`: Plan catalog (`id`, `name`, `code`, `monthly_price`, `tier`).
7. `audit_logs`: Immutable security and administrative activity trail (`org_id`, `actor_id`, `action`, `module`, `description`, `created_at`).
8. Operational Telemetry Tables:
   - `shipments` (`WHERE org_id = ?`)
   - `rfqs` (`WHERE org_id = ?`)
   - `quotes` (`WHERE org_id = ?`)
   - `bookings` (`WHERE org_id = ?`)
   - `invoices` (`WHERE org_id = ?`)
   - `customers` (`WHERE org_id = ?`)
   - `shipment_exceptions` (joined via shipments for unresolved milestones)

---

## 4. API Endpoints Used & Created
All SPortal organization APIs reside under `/api/v1/sportal/organizations` and require internal staff authentication (`RequireInternalStaff`) and granular RBAC permissions:

| Method | Endpoint | Permission | Description |
|---|---|---|---|
| `GET` | `/api/v1/sportal/organizations` | `organizations:view` | Paginated organization list with search, status/plan filters, and sorting |
| `GET` | `/api/v1/sportal/organizations/{id}` | `organizations:view` | Customer 360 foundation details: profile, subscription, users, stats, activity |
| `POST` | `/api/v1/sportal/organizations` | `organizations:create` | Creates new organization with duplicate validation and default subscription |
| `PATCH` | `/api/v1/sportal/organizations/{id}` | `organizations:update` | Updates organization legal, contact, and address profile |
| `GET` | `/api/v1/sportal/organizations/recent` | `organizations:view` | Top 10 recent organizations for executive dashboard overview |

---

## 5. Backend Services Used
- **Package**: `github.com/freel/backend/internal/sportal`
  - `types.go`: Canonical Go structs for organization requests, responses, filters, and Customer 360 foundation models.
  - `repository.go`: MariaDB sqlx repository executing parameterized queries with proper index usage, pagination, and multi-field duplicate checks.
  - `service.go`: Business logic enforcing internal RBAC (`PermOrganizationsView`, `PermOrganizationsCreate`, `PermOrganizationsUpdate`), field validation, duplicate detection, and audit logging.
  - `handler.go`: HTTP transport handling JSON serialization, query parameter parsing, and RFC-compliant HTTP status codes (200, 201, 400, 403, 409, 500).
  - `rbac.go`: Role-to-permission mapping and internal role enforcement.
  - `middleware.go`: `RequireInternalStaff` and `RequirePermission` guards.
  - `audit`: Integration with `github.com/freel/backend/internal/audit` logging `sportal.organization_created` and `sportal.organization_updated` events.

---

## 6. CPortal Relationship & Shared Truth
- **Single Authoritative MariaDB Database**: Zero data replication or secondary synchronization queues.
- **Consistency**: Any organization created or updated in SPortal immediately reflects in CPortal upon authorized tenant login.
- **Tenant Isolation**: CPortal customer users cannot access `/api/v1/sportal/*` endpoints. Any customer user token attempting SPortal access is rejected with HTTP 403 `FORBIDDEN` and an audit event is registered.

---

## 7. SPortal Authorization & RBAC
Internal roles and their organization permissions:
- `SUPER_ADMIN`, `CEO`, `OWNER`: Full access (`organizations:view`, `organizations:create`, `organizations:update`, `organizations:archive`).
- `ADMIN`: Full access (`organizations:view`, `organizations:create`, `organizations:update`).
- `CUSTOMER_SUCCESS`, `OPERATIONS`, `SUPPORT`: Read-only access (`organizations:view`).
- `FINANCE`: Read-only access (`organizations:view`).

---

## 8. Customer 360 Foundation Structure
The Customer 360 page (`/sportal/organizations/:organizationId`) provides a 360-degree foundation:
```
Organization
├── 1. Overview Header (Name, Legal Name, Status, Plan, Quick Actions)
├── 2. Telemetry Strip (Shipments, RFQs, Bookings, Invoices, Users, Exceptions)
├── 3. Tabbed Sections:
│   ├── Customer 360 Overview
│   ├── Company Profile (Identity, Contacts, Registered Address)
│   ├── Subscription & Plan (Tier, Billing Cycle, Pricing, Renewal)
│   ├── Staff & Users Directory (Active Tenant Members, Roles, Status)
│   └── Audit Ledger (Chronological Operational Trail)
└── 4. Extensibility Framework (Ready for Tasks S4–S11 onboarding, billing, health, integrations)
```

---

## 9. Organization Lifecycle & Status
The system uses the existing organization status model:
- `Active`: Operating forwarder organization with authorized workspace access.
- `Pending`: Newly created organization undergoing initial setup.
- `Inactive` / `Suspended`: Temporarily or permanently disabled organization.
- **Destructive Deletion**: Strictly prohibited. Deletion is replaced with archival and status transitions.

---

## 10. Fields Displayed & Editable
### Displayed:
- Company Trading Name (`name`)
- Legal Name (`legal_name`)
- CIN / Registration Number (`registration_number`)
- GSTIN / Tax ID (`tax_number`)
- Primary Email (`primary_email`)
- Phone Number (`phone_number`)
- Website (`website`)
- Full Address (`address`, `city`, `state`, `country`, `postal_code`)
- Industry & Company Type (`industry`, `company_type`)
- Active Subscription Plan & Status
- Real User Count & Member Profiles
- Operational Counts (Shipments, RFQs, Quotes, Bookings, Invoices, Exceptions)
- Audit History

### Editable via SPortal:
- Trading Name, Legal Name, CIN, GSTIN, Primary Email, Phone Number, Website, Address, City, State, Country, Postal Code, Industry, Company Type.

---

## 11. Duplicate Protection
The backend verifies duplicate signals prior to creation or update:
- Match on `name` or `legal_name`
- Match on `tax_number` (GST)
- Match on `primary_email`
- Self-exclusion on update (`AND id != ?`)
- In case of duplicate match, the API returns HTTP 409 `Conflict` with `DUPLICATE_ORGANIZATION` error code detailing the conflicting existing organization.

---

## 12. Audit Behavior
Administrative organization actions are recorded to `audit_logs`:
- Action: `sportal.organization_created` (Resource: `ORGANIZATION`, ID, Name)
- Action: `sportal.organization_updated` (Resource: `ORGANIZATION`, ID, Name)
- Includes internal actor ID, IP address, user agent, timestamp, and audit correlation ID.

---

## 13. Security Considerations
1. Authentication verification via JWT bearer token or internal session cookie.
2. `RequireInternalStaff` middleware guarantees customer tenant users cannot call SPortal administrative APIs.
3. Parameterized SQL queries prevent SQL injection across all search, filter, and pagination endpoints.
4. Input validation for email format, required fields, and duplicate keys.
5. Cross-Origin Resource Sharing (CORS) configured for internal administrative origins.

---

## 14. Schema Changes
- **Zero Schema Migrations Required**: All required fields exist in the canonical MariaDB schema (`organizations`, `org_members`, `roles`, `organization_subscriptions`, `subscription_plans`).
- Existing real data (33+ forwarder organizations) remained completely intact without alteration or data loss.

---

## 15. Known Limitations & Roadmap
- Commercial payment gateway billing automation will be implemented in Task S5 (Subscriptions & Billing).
- Onboarding workflow orchestration will be populated in Task S4 (Onboarding Engine).
- Live AI Workforce agent assignment per tenant will be connected in Task S11 (SPortal AI Workforce Administration).
