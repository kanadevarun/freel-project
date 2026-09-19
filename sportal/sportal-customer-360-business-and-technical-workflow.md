# SPortal Customer 360: Cross-Module Business View & Relationship Intelligence

## 1. Executive Summary & Purpose

The **SPortal Customer 360 Workspace** is the core operational, commercial, and technical relationship cockpit built for LogisticsHQ internal SaaS leadership, account executives, and operations engineers.

When an authorized internal user navigates to **Organizations** and selects a freight-forwarder customer (or browses directly to `/organizations/:organizationId` / `/customer-360/:organizationId`), SPortal loads a holistic, 360-degree single-pane view of that customer's complete lifecycle and interaction with the platform.

Crucially, **Customer 360 introduces zero duplicate schemas, databases, or shadow state engines**. All telemetry, shipments, invoices, contracts, exceptions, documents, users, automations, and audit entries are sourced directly from the authoritative live MariaDB engine.

---

## 2. Information Architecture & Workspace Topology

The Customer 360 experience is architected in accordance with the enterprise SaaS aesthetic reference (`frontend/public/images/sportal/sporatlDashboard.png` and `frontend/public/images/sportal/sportalCustomerView.png`).

```
+---------------------------------------------------------------------------------------------------+
|  Header: Org Icon | Company Name | Active Badge | ORG-0001 • Freight Forwarder • HQ • Joined Date |
|  Contact Links: Website | Email | Phone           Action Buttons: [Edit Org] [Actions ▾]         |
+---------------------------------------------------------------------------------------------------+
|  Horizontal Nav Tabs: Overview | Company | Onboarding | Subscription | Billing | Users (5) |       |
|                       Operations | Shipments (4) | Exceptions (2) | Contracts (1 Due) |           |
|                       Compliance | AI & Automation | Integrations | Documents | Activity          |
+---------------------------------------------------------------------------------------------------+
| [OVERVIEW TAB]                                                                                    |
|  6 KPI Cards:                                                                                     |
|   1. Active Shipments (4)    2. Open Exceptions (2)      3. Total Users (5)                      |
|   4. Outstanding Invoices ($) 5. Commercial Tier (Pro)   6. Customer Health (Score: 75/100)       |
+---------------------------------------------------------------------------------------------------+
|  4 Middle Panels:                                                                                 |
|   1. Organization Details   2. Subscription & Billing    3. Recent Audit Activity 4. Customer Team |
+---------------------------------------------------------------------------------------------------+
|  3 Bottom Panels:                                                                                 |
|   1. Operational Overview   2. Shipment Trends Dual Bar  3. Open Items & Alerts Deep Links        |
+---------------------------------------------------------------------------------------------------+
```

---

## 3. Cross-Module Data Mapping & Architecture

| Customer 360 Domain | Authoritative MariaDB Tables / Columns | SPortal Endpoint | Deep Link Target |
| :--- | :--- | :--- | :--- |
| **Organization Identity** | `organizations` (`name`, `legal_name`, `registration_number`, `tax_number`, `website`, `primary_email`, `phone_number`, `address`, `city`, `state`, `country`, `status`) | `GET /api/v1/sportal/organizations/:id` | `/organizations` |
| **Subscription & Quotas** | `organization_subscriptions` (`plan_id`, `status`, `billing_cycle`, `current_period_end`, `auto_renew`) + `subscription_plans` | `GET /api/v1/sportal/organizations/:id` & `GET /api/v1/sportal/organizations/:id/subscription` | `/subscriptions` |
| **Users & Workforce** | `users` (`first_name`, `last_name`, `email`, `role`, `status`) + `user_roles` (`name`) | `GET /api/v1/sportal/organizations/:id/users` | `/users?orgId=:id` |
| **Shipments** | `shipments` (`booking_number`, `carrier_scac`, `origin_port`, `destination_port`, `vessel_name`, `voyage_number`, `status`, `created_at`) | `GET /api/v1/sportal/organizations/:id/shipments` | `/shipments?orgId=:id` |
| **Invoices & Billing** | `customer_invoices` / `invoices` (`invoice_number`, `customer_name`, `total_amount`, `balance_due`, `status`, `due_date`, `created_at`) | `GET /api/v1/sportal/organizations/:id/invoices` | `/finance?orgId=:id` |
| **Contracts** | `contracts` (`contract_reference`, `contract_name`, `contract_type`, `party_name`, `contract_value`, `status`, `effective_date`, `expiry_date`) | `GET /api/v1/sportal/organizations/:id/contracts` | `/contracts?orgId=:id` |
| **Exceptions** | `shipment_exceptions` (`shipment_id`, `exception_type`, `severity`, `title`, `description`, `status`, `created_at`) | `GET /api/v1/sportal/organizations/:id/exceptions` | `/exceptions?orgId=:id` |
| **AI Workforce** | `ai_automations` (`is_enabled`) + `ai_processing_tasks` (`status`, `created_at`) | `GET /api/v1/sportal/organizations/:id/ai-summary` | `/ai?orgId=:id` |
| **Integrations** | `carrier_integrations` / `system_integrations` / fallback gateways | `GET /api/v1/sportal/organizations/:id/integrations` | `/integrations` |
| **Documents** | `shipment_documents` (`file_name`, `doc_type`, `file_size`, `status`, `created_at`) | `GET /api/v1/sportal/organizations/:id/documents` | `/documents?orgId=:id` |
| **Immutable Audit** | `audit_logs` (`action`, `module`, `description`, `result`, `created_at`) | `GET /api/v1/sportal/organizations/:id` (`recent_activity`) | `/audit-logs` |

---

## 4. End-to-End Operational Lifecycle & Workflows

### 4.1. Account Exploration Workflow
1. Internal staff member opens SPortal (`/organizations`).
2. Staff clicks on customer row (e.g. `Freel Global Logistics Pvt Ltd`).
3. SPortal fetches `GET /api/v1/sportal/organizations/1`, retrieving identity, subscription, health score, 6-month trends, open alerts, and team summary in parallel with `userSummary` and `richSubscription`.
4. Header presents quick operational telemetry, organization badges, contact metadata, and commercial subscription terms.
5. 6 KPI cards provide instant situation awareness: 4 Active Shipments, 2 Open Exceptions, 5 Users, 6 Outstanding Invoices ($121,330 due), Professional Subscription term, and Customer Health (Needs Attention — 75/100).

### 4.2. Exception & Risk Mitigation Workflow
1. Under **Open Items & Alerts**, staff clicks on `[high] Open Exceptions (2)`.
2. The user is taken directly to the **Exceptions** tab (or deep-linked to the core Exceptions module with tenant filter `?orgId=1`).
3. Staff investigates unresolved shipment deviation `SHP-1` (`ETA_DELAY - CRITICAL`) and `SHP-280` (`DOCUMENT_ISSUE - HIGH`).
4. Staff can update or resolve the deviation without leaving the platform context.

### 4.3. Commercial & Subscription Governance Workflow
1. Staff opens the **Subscription** or **Billing** tab.
2. SPortal lazy-loads `GET /api/v1/sportal/organizations/1/invoices` and displays outstanding balances, invoice numbers, and due dates.
3. From the header **Actions ▾** dropdown, staff can:
   - **Change Subscription Plan**: Upgrade/downgrade between Starter, Professional, Enterprise tiers with proration calculations.
   - **Manual Renewal**: Extend the subscription period by +1, +3, +6, or +12 months.
   - **Toggle Auto-Renew**: Enable or disable automatic recurring renewal with reason recording.
   - **Cancel Subscription**: Terminate access immediately or at period end.
   - **Export Customer Dossier**: Generate a clean PDF/print dossier.

### 4.4. Security, Isolation & Multi-Tenancy Controls
- **Internal Staff Only**: All Customer 360 endpoints require internal staff authentication (`RequireInternalStaff`). External customer forwarder users are strictly forbidden (`403 Forbidden`).
- **Granular RBAC**: Staff role must possess `PermOrganizationsView` permission.
- **Tenant Isolation**: Cross-module queries require explicit tenant context (`org_id = ?`). Queries against invalid or non-existent IDs return explicit `404 Not Found`.
- **Sensitive Data Redaction**: AI task raw prompts, chain-of-thought tokens, carrier API secret credentials, and private customer tokens are excluded from all API serialization.
