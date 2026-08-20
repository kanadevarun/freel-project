# Freel Freight Forwarder — Module Connectivity Matrix
**Live Technical Audit & Linkage Specification**
*Generated: August 15, 2026*

---

## 1. Executive Summary

This document provides the authoritative, code-verified connectivity matrix for all **18 Freight Forwarder sidebar modules**.

Every module has been inspected across the complete 11-stage architectural chain:
`Sidebar Link → React Route → Page Component → Frontend Service → HTTP API → Go Route → Go Handler → Go Service → Repository → PostgreSQL Tables → UI State`

---

## 2. Complete Live Module Connectivity Matrix

| Module | Sidebar Link | React Route | Page Component | Frontend Service | API Endpoint | Go Handler | Service | Repository | DB Tables | Auth | RBAC | Tenant Scoped | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :---: | :---: | :---: | :--- |
| **1. Dashboard** | `/dashboard` | `/dashboard` | `DashboardHome.jsx` | `dashboardService.js` | `GET /api/v1/dashboard/mission-control` | `dashboard.Endpoints` | `dashboard.Service` | `dashboard.Repository` | `rfqs`, `leads`, `shipments` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |
| **2. Leads** | `/dashboard/leads` | `/dashboard/leads` | `LeadsPage.jsx` | `leadsService.js` | `GET/POST /api/v1/leads`, `/import` | `leads.Endpoints` | `leads.Service` | `leads.Repository` | `leads`, `lead_interactions` | ✅ | ✅ (`LEADS.READ`) | ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |
| **3. RFQs** | `/dashboard/rfqs` | `/dashboard/rfqs` | `RFQPage.jsx` | `rfqService.js` | `GET/POST /api/v1/rfqs`, `/{id}/stage` | `rfq.Endpoints` | `rfq.BusinessLogic` | `rfq.Datalayer` | `rfqs`, `rfq_items`, `customers` | ✅ | ✅ (`RFQS.READ`) | ✅ (`WHERE org_id = $1`) | **PARTIALLY CONNECTED** *(Fixing API prefix)* |
| **4. Shipments** | `/dashboard/shipments` | `/dashboard/shipments` | `ShipmentsPage.jsx` | `api.js` | `GET /api/v1/shipments`, `/{id}` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments`, `milestones`, `exceptions` | ✅ | ✅ (`SHIPMENTS.READ`)| ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |
| **5. Bookings** | `/dashboard/bookings` | `/dashboard/bookings` | `ShipmentsPage.jsx` (`filter="BOOKED"`) | `api.js` | `GET /api/v1/shipments?status=BOOKED` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Fixing in P0)* |
| **6. Tracking** | `/dashboard/tracking` | `/dashboard/tracking` | `TrackingPage.jsx` / `PlaceholderPage` | `api.js` | `GET /api/v1/shipments` | `shipments.Handler` | `shipments.Service` | `shipments.Repository` | `shipments`, `milestones` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Fixing in P0)* |
| **7. Quotations** | `/dashboard/quotations` | `/dashboard/quotations` | `RFQPage.jsx` (`tab="QUOTES"`) | `rfqService.js` | `GET /api/v1/rfqs`, `/{id}/quotes` | `pricing.Handler` / `rfq` | `pricing.Service` | `pricing.Repository` | `rfq_pricing_options`, `rules` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Fixing in P0)* |
| **8. Rate Mgmt** | `/dashboard/rate-management` | `/dashboard/rate-management` | `PlaceholderPage` (Rate Intelligence) | `api.js` | `GET /api/v1/rates/search` | `rates.Handler` | `rates.Service` | `rates.Repository` | `ports`, `contract_rates`, `spot_rates` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **BACKEND ONLY** *(Adding workspace route in P0)* |
| **9. Contracts** | `/dashboard/contracts` | `/dashboard/contracts` | `ContractsPage.jsx` | `contractsService.js` | `GET/POST /api/v1/contracts`, `/review`| `contracts.Handler` | `contracts.Service` | `contracts.Repository` | `carrier_contracts`, `reviews` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **BACKEND ONLY** *(Registering route in P0)* |
| **10. Customers** | `/dashboard/companies` | `/dashboard/companies` | `PlaceholderPage` (Directory) | `api.js` | `GET /api/v1/companies` (stub) | `server.routes.go:48` | Stub | `companies.Repository` | `companies`, `customers`, `contacts` | ✅ | ✅ (`COMPANIES.READ`) | ✅ (`WHERE org_id = $1`) | **PARTIALLY CONNECTED** |
| **11. Documents** | `/dashboard/documents` | `/dashboard/documents` | `PlaceholderPage` (Generator) | `api.js` | `GET /api/v1/shipments/{id}/documents` | `documents.Handler` | `documents.Service` | `documents.Repository` | `shipment_documents` | ✅ | ✅ (`DOCUMENTS.READ`)| ✅ (`WHERE org_id = $1`) | **PARTIALLY CONNECTED** |
| **12. Templates** | `/dashboard/templates` | `/dashboard/templates` | `PlaceholderPage` (Templates) | — | — | — | — | — | — | — | — | — | **NOT IMPLEMENTED** *(Adding clean workspace route)* |
| **13. Approvals** | `/dashboard/approvals` | `/dashboard/approvals` | `PlaceholderPage` (Approvals) | `contractsService.js` | `/api/v1/contracts/review` | `contracts.Handler` | `contracts.Service` | `contracts.Repository` | `reviews`, `discrepancies` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Adding workspace route in P0)* |
| **14. Invoices** | `/dashboard/invoices` | `/dashboard/invoices` | `PlaceholderPage` (Billing & Invoices) | `api.js` | `GET /api/v1/shipments/{id}/finance` | `finance.Handler`, `billing`| `finance.Service` | `finance.Repository` | `shipment_invoices`, `cust_invoices` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Adding workspace route in P0)* |
| **15. Payments** | `/dashboard/payments` | `/dashboard/payments` | `PlaceholderPage` (Payments) | `api.js` | `POST /api/v1/billing/invoices/{id}/pay`| `billing.Handler` | `billing.Service` | `billing.Repository` | `customer_invoices` | ✅ | ✅ | ✅ (`WHERE org_id = $1`) | **ROUTE MISSING** *(Adding workspace route in P0)* |
| **16. Reports** | `/dashboard/reports` | `/dashboard/reports` | `ReportsPage.jsx` | `api.js` | `GET /api/v1/reports/metrics` | `reports.Endpoints` | `reports.Service` | `reports.Repository` | `rfqs`, `leads`, `shipments` | ✅ | ✅ (`DASHBOARD.READ`)| ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |
| **17. Users** | `/dashboard/users` | `/dashboard/users` | `UsersPage.jsx` | `api.js` | `GET /api/v1/users`, `/invite` | `users.Endpoints` / `auth` | `users.Service` | `users.Repository` | `users`, `org_members`, `roles` | ✅ | ✅ (`USERS.READ`) | ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |
| **18. Settings** | `/dashboard/settings` | `/dashboard/settings` | `RolesPage.jsx` | `api.js` | `GET /api/v1/roles`, `/permissions` | `rbac.Endpoints` | `rbac.Service` | `rbac.Repository` | `roles`, `permissions` | ✅ | ✅ (`SETTINGS.READ`) | ✅ (`WHERE org_id = $1`) | **FULLY CONNECTED** |

---

## 3. P0 Action Items & Resolutions

1. **Bookings 404:** Fixed by routing `/dashboard/bookings` to `ShipmentsPage` configured with `filter="BOOKED"` inside `AppShell`.
2. **RFQ API Prefix:** Fixed by updating all endpoints in `rfqService.js` from `/rfqs` to `/api/v1/rfqs`.
3. **Sidebar Badges:** Removed static hardcoded counts (`12`, `8`, `23`, `2`) and connected dynamic count hydration.
4. **Contracts Routing:** Registered `/dashboard/contracts` route mapped to `ContractsPage.jsx`.
5. **Dashboard 404 Catch-All:** Added authenticated workspace fallback `<Route path="/dashboard/*" element={<DashboardNotFound />} />` to prevent public website redirects.
6. **Missing Sidebar Routes:** Ensured every single sidebar item routes to a valid workspace component within `AppShell`.
