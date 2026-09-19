# Task 3.3.A — Customers Business Workflow, Data Flow, Database Mapping, AI Integration, and Complete Technical Documentation

> **Status:** PASS — CUSTOMERS WORKFLOW DOCUMENTED  
> **Repository:** LogisticsHQ (`freel-project`)  
> **Modules Evaluated:** React Frontend (`CustomersPage`, `CustomerDetailsPage`, `CustomerModal`, `LeadConversionModal`), Go Backend (`backend/internal/customers`), MariaDB Schema (`customers`, `contacts`, `customer_addresses`, `customer_lead_links`, `customer_health_evaluations`, `customer_risk_events`, `customer_opportunity_events`, `customer_ownership_history`, `customer_communication_preferences`, `customer_followup_records`), Python AI Sidecar (`ai_sidecar/app/customer_relationship`, `workforce/agents.py`).  
> **Date:** September 13, 2026  
> **Evaluator:** Deep Functional Audit Agent (Antigravity)

---

## 1. Executive Summary

The **Customers** module in LogisticsHQ is the commercial core and system of record for accounts receivable counterparties, shippers, consignees, traders, manufacturers, and freight forwarders. It serves as the bridge between pre-sales pipeline acquisition (**Leads**) and operational freight fulfillment (**RFQs**, **Quotations**, **Bookings**, **Shipments**, and **Invoices**).

Rather than functioning as a static address book, the module acts as a **Commercial Intelligence & Risk Workspace**. It couples MariaDB transactional ground-truth with deterministic operational heuristics and Python-based AI intelligence to govern credit limits, detect churn risks, highlight account inactivity, propose contract renewals, and orchestrate human-governed customer follow-up actions.

All records are strictly isolated by multi-tenant organization boundaries (`org_id`), guarded by Go-level RBAC (`ResourceCompanies`), audited through universal audit logs, and integrated bi-directionally with cross-module workflows.

---

## 2. Customers in Plain English

To anyone working in logistics, a **Customer** represents a company or organization that buys freight forwarding and logistics services from your business.

In simple terms:
1. **How they get here:** A Customer usually begins as a sales inquiry or **Lead**. When a salesperson qualifies the prospect and secures their first business engagement, the Lead is converted into a Customer account. Customers can also be added manually or integrated from external ERPs.
2. **What you track:** You maintain legal trading identities, tax identification (GSTIN/Tax ID/EORI), credit limits, payment terms (e.g., NET30, NET60), physical operational facilities (warehouses, billing addresses, discharge ports), and direct human contacts (operations managers, commercial directors, billing specialists).
3. **How your team uses it:** 
   - **Sales & Commercial:** Monitors open quotations, win/loss conversion rates, and contract expirations.
   - **Operations:** Attaches active bookings and shipments to the customer profile, ensuring cargo is billed to the right counterparty.
   - **Finance & Credit Control:** Tracks credit utilization, flags accounts on credit hold, and prevents dispatching cargo for delinquent accounts.
   - **Customer Success:** Reviews automated health scores and alert badges (Healthy, Watch, At Risk, Critical) to address issues before an account churns.

---

## 3. Business Purpose

The Customers module fulfills five primary commercial objectives:
- **Master Data Authority:** Eliminates duplicated company profiles and ensures unified company IDs across shipping lines, airlines, customs brokers, and warehouses.
- **Credit & Financial Risk Governance:** Enforces structured credit terms, payment limits, and hold policies to safeguard company cash flow.
- **Cross-Lifecycle Continuity (360° Visibility):** Provides an operational cockpit where freight inquiries (RFQs), pricing estimates (Quotes), booked freight (Bookings), live transit operations (Shipments), and master service agreements (Contracts) converge under a single account.
- **Account Accountability & Governance:** Records commercial ownership (Primary and Secondary Account Managers) with historical audit tracking whenever accounts are reassigned.
- **Proactive Relationship Intelligence:** Automatically scans operational inactivity, unassigned accounts, unresolved cargo delays, and open proposals to trigger proactive follow-ups before commercial relationships deteriorate.

---

## 4. Customer Lifecycle

The lifecycle of a Customer in LogisticsHQ is tracked through two distinct dimensions: **Operational Account Status** (`status`) and **Evaluated Health State** (`health_status`).

### Operational Status Transitions

| Current State | Next State | Business Meaning | Who Can Change It | Side Effect |
| :--- | :--- | :--- | :--- | :--- |
| **ONBOARDING** | `ACTIVE` | Account profile created; tax ID and payment terms verified. | Sales / Commercial Admin | Allows creating RFQs, quotes, and linking shipments. |
| **ACTIVE** | `INACTIVE` | Account temporarily suspended or voluntarily decommissioned. | Commercial Manager / Super Admin | Prevents issuing new quotations or accepting bookings. |
| **ACTIVE** | `AT_RISK` | Account health deteriorated (health score < 60) or on credit hold. | Automated Intelligence / System | Triggers Critical/Warning badges on Attention Panel. |
| **AT_RISK** | `ACTIVE` | Operational blockers resolved, overdue balances cleared. | System / Commercial Admin | Resolves risk events, updates health score ≥ 80. |
| **ACTIVE** / **INACTIVE** | `CHURNED` | Long-term operational inactivity (>180 days with zero inquiries/shipments). | System / Commercial Lead | Triggers re-engagement opportunity recommendations. |
| *Any State* | `ARCHIVED` (`archived_at != NULL`) | Account hidden from active directory (`status = 'INACTIVE'`). | Super Admin / Org Admin | Soft-deletes account; hides from standard list queries. |
| `ARCHIVED` | `ACTIVE` | Restores archived account to active status (`archived_at = NULL`). | Super Admin / Org Admin | Re-enables account in list; writes audit log. |

---

## 5. Customer Creation Paths

LogisticsHQ supports two primary, fully verified creation mechanisms:

```
Path 1: Manual Customer Creation
User (UI) → POST /api/v1/customers → Go Handler → BL Validation → DL Transaction (customers + contacts + customer_addresses) → MariaDB → Audit Log → UI

Path 2: Deterministic Lead Conversion
Sales User → POST /api/v1/customers/convert-lead → Duplicate Engine → Link Existing OR Create New → DL Transaction (customers + customer_contacts + customer_lead_links + update leads.status='CONVERTED') → Audit Log → Event Mesh → UI
```

### Path 1: Manual Customer Creation
1. **Business Trigger:** Commercial manager manually registers a known shipper, trader, or corporate partner.
2. **UI/API:** User clicks `+ New Customer` in `CustomersPage.jsx`, opening `CustomerModal.jsx`. Submits `POST /api/v1/customers`.
3. **Validation:** Legal company name is mandatory. Currency defaults to `USD`, payment terms default to `NET30`, health score initialized to `80`.
4. **Authorization & Tenant:** Enforced via `rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)` and `userCtx.OrgID`.
5. **Database Execution:** Generates autoincrement ID, formats human-readable code (`CUST-YYYY-#####`), and optionally creates initial contact (`contacts`) and billing address (`customer_addresses`).
6. **Audit Trail:** Writes `CUSTOMER` creation record to `audit_logs`.

---

## 6. Lead → Customer Workflow

Converting an acquired Sales Lead into a verified Customer Account is a critical business handover.

```
                  ┌───────────────────────────────┐
                  │    Sales Lead (Leads Page)    │
                  └──────────────┬────────────────┘
                                 │ Click "Convert to Customer"
                                 ▼
                  ┌───────────────────────────────┐
                  │   LeadConversionModal.jsx     │
                  │   (Duplicate Scan Invoked)    │
                  └──────────────┬────────────────┘
                                 │
                 ┌───────────────┴───────────────┐
                 ▼                               ▼
       [Matches Found ≥ 50%]           [No Matches Found]
                 │                               │
        User selects:                   User selects:
  Link to Existing Customer       Create Brand New Customer
                 │                               │
                 └───────────────┬───────────────┘
                                 │ Submits POST /api/v1/customers/convert-lead
                                 ▼
                  ┌───────────────────────────────┐
                  │ Go Backend: ConvertLeadHandler│
                  └──────────────┬────────────────┘
                                 │
           ┌─────────────────────┼─────────────────────┐
           ▼                     ▼                     ▼
 ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
 │  Customer Entity │  │ Lead Audit Link  │  │   Lead Status    │
 │ (Create or Link) │  │customer_lead_link│  │ UPDATE leads SET │
 │  code: CUST-...  │  │lead_id=..., notes│  │status='CONVERTED'│
 └──────────────────┘  └──────────────────┘  └──────────────────┘
```

### Business Meaning
When a sales representative establishes viable commercial intent, they click "Convert" on the Lead record. The system scans existing accounts to protect against duplicate records. The user may either link the lead to an existing customer account (adding the lead contact to that customer) or spawn a brand new Customer Account entity.

### Technical Step-by-Step
1. **Frontend Trigger:** User invokes `LeadConversionModal.jsx`.
2. **Duplicate Detection API:** Calls `POST /api/v1/customers/check-duplicate` with `{ name, email, phone, lead_id }`.
3. **Scoring Engine (`duplicate_engine.go`):**
   - Exact Tax ID / GSTIN match $\rightarrow$ **100% confidence**
   - Exact corporate domain match $\rightarrow$ **85% confidence**
   - Legal name exact/normalized match $\rightarrow$ **80% confidence**
   - Primary contact email match $\rightarrow$ **75% confidence**
   - Primary contact phone match $\rightarrow$ **70% confidence**
   - Combined domain + name partial match $\rightarrow$ **90% confidence**
4. **Execution (`POST /api/v1/customers/convert-lead`):**
   - Verifies lead existence, tenant isolation, and prior conversion status. If `lead.Status == 'CONVERTED'` and neither `force_create_new=true` nor `link_to_existing_customer_id` is supplied, rejects duplicate conversion to prevent accidental customer sprawl.
   - Begins MariaDB transaction (`tx`).
   - If linking: Verifies existing customer, inserts contact from lead into `contacts` if new.
   - If creating new: Inserts into `customers` with status `ACTIVE`, assigns customer code `CUST-YYYY-#####`, creates primary contact in `contacts`.
   - Inserts record into `customer_lead_links` (`lead_id`, `customer_id`, `converted_by_user_id`, `conversion_notes`).
   - Updates `leads` table: `status = 'CONVERTED'`, `converted_from_outreach_at = NOW()`.
   - Commits transaction and records audit log.

---

## 7. Customer Edit Workflow

Updating an existing customer is initiated via `Edit Customer` on the Customer List kebab menu or from the top action bar of the Customer 360 page.

- **Editable Fields:** Company Name, Trading Name, Customer Type (Shipper, Consignee, Forwarder, Broker, etc.), Domain, Industry, Tax ID / PAN / EORI, Currency, Payment Terms, Credit Limit, Health Score, Account Owner ID, Website, Country, City, Contact Person, Contact Email, Contact Phone, Notes, and Operational Status.
- **Restricted/Immutable Fields:** `id` (Primary Key), `org_id` (Tenant Isolation Key), and `customer_code` (Unique commercial identifier).
- **Audit Logging:** Go backend records old and new values (`Before` vs `After` JSON payload) in `audit_logs` under action `UPDATE`.

---

## 8. Customer Contacts

Every Customer account can have multiple associated human contacts (commercial directors, dispatch coordinators, invoicing accountants).

- **Table:** `contacts` (or `customer_contacts`) with foreign key `customer_id` and tenant filter `org_id`.
- **Role Tagging:** Structured via `contact_role`: `COMMERCIAL`, `OPERATIONS`, `LOGISTICS`, `FINANCE`, `MANAGEMENT`, `DECISION_MAKER`, `BILLING`, `OTHER`.
- **Primary Contact Flag:** `is_primary = TRUE` designates the default point of contact displayed on the main directory table, invoices, and communication drafts.
- **Endpoints:**
  - `GET /api/v1/customers/{id}/contacts` $\rightarrow$ List contacts sorted by `is_primary DESC`.
  - `POST /api/v1/customers/{id}/contacts` $\rightarrow$ Add new contact.
  - `PUT /api/v1/customers/{id}/contacts/{contact_id}` $\rightarrow$ Update contact details.
  - `DELETE /api/v1/customers/{id}/contacts/{contact_id}` $\rightarrow$ Delete contact.

---

## 9. Customer → RFQ Relationship

- **Relationship Type:** One-to-Many (`customers.id` $\rightarrow$ `rfqs.customer_id`).
- **Operational Purpose:** Every freight rate inquiry is tied directly to the commercial customer requesting it.
- **UI Integration:** The Customer 360 page includes a dedicated `+ New RFQ` action in the header and displays the inquiry history in the **Commercial** tab.
- **Backend Query:** `SELECT id, rfq_number, status, stage, origin, destination, mode_of_transport, created_at FROM rfqs WHERE org_id = ? AND customer_id = ? ORDER BY created_at DESC`.

---

## 10. Customer → Quotation Relationship

- **Relationship Type:** One-to-Many (`customers.id` $\rightarrow$ `quotations.customer_id`).
- **Commercial Meaning:** Price proposals generated by sales engineers referencing specific routing, ocean/air base freights, fuel surcharges, and detention free days.
- **Metrics Computed:** Total quotation value, open quote pipeline, accepted quote value, and quote conversion rate percentage (`accepted_quotations / total_quotations * 100`).
- **Direct Navigation:** Quotations table in the Customer 360 **Commercial** tab links directly to `/dashboard/quotations/{id}`.

---

## 11. Customer → Booking / Shipment Relationships

- **Booking Link:** `bookings.rfq_id = rfqs.id` where `rfqs.customer_id = customers.id`. Provides carrier confirmations, container allocation, and loading port schedules.
- **Shipment Link:** Direct foreign key `shipments.customer_id = customers.id`. Tracks live transit milestones, customs status, and operational cargo exceptions.
- **UI Tab:** **Operational** tab in Customer 360 displays all active and historical freight movements with direct links to `/dashboard/shipments/{id}` and `/dashboard/bookings/{id}`.

---

## 12. Customer → Finance & Invoicing

- **Relationship:** Direct linkage to `shipment_customer_invoices` (via `shipment_id` and RFQ customer reference).
- **Financial Profile Fields:** `currency`, `payment_terms`, `credit_limit`, `credit_status` (`GOOD_STANDING`, `REVIEW_REQUIRED`, `ON_HOLD`, `NO_CREDIT_LIMIT`), and `commercial_notes`.
- **YTD Revenue Calculation:** Evaluated dynamically via `SUM(total_amount)` from `shipment_customer_invoices` where status in (`APPROVED`, `PAID`).
- **Credit Enforcement:** If `credit_status` is `ON_HOLD`, the system flags the account in the Attention Panel and AI follow-up drawer, blocking standard autonomous dispatch without credit officer sign-off.

---

## 13. Customer → Contracts & Compliance

- **Contract Linkage:** Linked via `contract_parties.customer_id = customers.id` or `contract_links.linked_entity_type = 'CUSTOMER'` and `linked_entity_id = customers.id`.
- **Documents Repository:** Integrated with the Universal Document Management system via `GET /api/v1/documents?customer_id={id}`. Allows uploading and previewing KYC filings, tax registrations, certificates of incorporation, and signed master service agreements (MSAs).

---

## 14. Database Table Mapping

The table below lists the MariaDB tables responsible for the Customers module:

| Table | Business Purpose | Customer Operation | Important Fields | Related Tables |
| :--- | :--- | :--- | :--- | :--- |
| `customers` | Master account entity for shippers, consignees, forwarders | Directory listing, CRUD, credit limits, account governance | `id`, `org_id`, `customer_code`, `name`, `customer_type`, `tax_id`, `currency`, `payment_terms`, `credit_limit`, `credit_status`, `health_score`, `account_owner_id`, `status`, `archived_at` | `organizations`, `users`, `rfqs`, `quotations`, `shipments` |
| `contacts` | Individual employee contacts under customer accounts | Directory, primary contact tag, communication recipients | `id`, `org_id`, `customer_id`, `first_name`, `last_name`, `email`, `phone`, `mobile`, `job_title`, `department`, `contact_role`, `is_primary` | `customers`, `organizations` |
| `customer_addresses` | Physical facilities (warehouses, billing, ports) | Tax invoicing, dispatch coordination | `id`, `org_id`, `customer_id`, `address_type`, `address_line_1`, `city`, `state`, `postal_code`, `country`, `is_primary_billing`, `is_primary_shipping` | `customers`, `organizations` |
| `customer_lead_links` | Audit trail connecting original Lead to Customer account | Lead conversion history | `id`, `org_id`, `customer_id`, `lead_id`, `converted_by_user_id`, `conversion_notes`, `created_at` | `customers`, `leads`, `users` |
| `customer_health_evaluations` | Historical log of customer health scores and factor breakdowns | Retention scoring, attention alerts | `id`, `org_id`, `customer_id`, `health_status`, `health_score`, `contributing_factors_json`, `evaluated_at` | `customers`, `organizations` |
| `customer_risk_events` | Granular operational & financial risk flags | Attention panel, risk resolution workflow | `id`, `org_id`, `customer_id`, `risk_type`, `severity`, `title`, `description`, `is_resolved`, `resolved_at`, `resolved_by`, `resolution_note` | `customers`, `users` |
| `customer_opportunity_events` | Proactive commercial opportunities (renewals, follow-ups) | Sales revenue recommendations | `id`, `org_id`, `customer_id`, `opportunity_type`, `priority`, `title`, `reason`, `suggested_action`, `detected_at` | `customers`, `organizations` |
| `customer_ownership_history` | Audit log of commercial account ownership reassignments | Account governance | `id`, `org_id`, `customer_id`, `previous_owner_id`, `new_owner_id`, `ownership_type`, `changed_by_user_id`, `change_reason`, `created_at` | `customers`, `users` |
| `customer_communication_preferences` | Governed channel and contact rules (opt-out, quiet hours) | Autonomous outreach guardrails | `id`, `org_id`, `customer_id`, `preferred_channel`, `opt_out`, `business_hours_only`, `max_followups_per_incident`, `min_followup_interval_hours` | `customers`, `organizations` |
| `customer_followup_records` | History of generated follow-up messages, drafts, and responses | Action System / Autonomous follow-ups | `id`, `org_id`, `customer_id`, `recipient_email`, `subject`, `full_body`, `status`, `approval_status`, `idempotency_key`, `sent_at`, `customer_response` | `customers`, `organizations` |
| `ai_recommendations` | Centralized recommendations table | Cross-module action sync | `id`, `org_id`, `source_type`, `source_id`, `category`, `priority`, `risk_level`, `customer_id`, `customer_name`, `followup_type`, `draft_status`, `dedup_hash` | `customers`, `organizations` |

---

## 15. Customer Relationship Map

Verified relational schema of a Customer account in LogisticsHQ:

```
Customer (customers)
├── Contacts (contacts / customer_contacts)
├── Operational Addresses (customer_addresses)
├── Converted Leads Link (customer_lead_links → leads)
├── Rate Inquiries (rfqs)
├── Price Proposals (quotations)
├── Bookings (bookings via rfqs)
├── Freight Movements (shipments)
├── Customer Invoices (shipment_customer_invoices)
├── Service Contracts (contracts via contract_parties & contract_links)
├── Master Documents (documents via customer_id metadata)
├── Health Evaluations (customer_health_evaluations)
├── Risk Events (customer_risk_events)
├── Opportunity Events (customer_opportunity_events)
├── Ownership History (customer_ownership_history)
├── Communication Preferences (customer_communication_preferences)
├── Governed Follow-Up Records (customer_followup_records)
└── AI Recommendations (ai_recommendations)
```

---

## 16. API Mapping

| Business Operation | Frontend Function | HTTP Method | Endpoint | Go Handler / Service | Database Target | Result |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **List Customers** | `getCustomers(params)` | `GET` | `/api/v1/customers` | `ListCustomersEndpoint` | `customers` | Paginated customer list + total count |
| **Get Directory KPIs** | `getCustomerKPIs()` | `GET` | `/api/v1/customers/kpis` | `GetKPIsEndpoint` | `customers`, `contracts`, `shipment_customer_invoices` | Aggregate counts and revenue metrics |
| **Get Customer 360** | `getCustomerDashboard(id)`| `GET` | `/api/v1/customers/{id}/dashboard`| `Get360DashboardEndpoint` | `customers`, `rfqs`, `quotations`, `bookings`, `shipments` | Full 360° cockpit payload + timeline |
| **Duplicate Scan** | `checkDuplicate(payload)`| `POST` | `/api/v1/customers/check-duplicate` | `CheckDuplicateEndpoint` | `customers` (in-memory candidate scoring) | Array of matched candidates with confidence scores |
| **Create Customer** | `createCustomer(payload)` | `POST` | `/api/v1/customers` | `CreateCustomerEndpoint` | `customers`, `contacts`, `customer_addresses` | New Customer entity + initial contact/address |
| **Convert Lead** | `convertLead(payload)` | `POST` | `/api/v1/customers/convert-lead` | `ConvertLeadEndpoint` | `customers`, `customer_lead_links`, `leads` | Converted/linked Customer entity + status update |
| **Update Customer** | `updateCustomer(id, payload)`| `PUT`| `/api/v1/customers/{id}` | `UpdateCustomerEndpoint` | `customers` | Updated Customer entity + audit record |
| **Archive Customer** | `archiveCustomer(id)` | `POST` | `/api/v1/customers/{id}/archive` | `ArchiveCustomerEndpoint` | `customers` (`archived_at=NOW()`) | Account marked inactive/archived |
| **Reactivate Customer**| `reactivateCustomer(id)`| `POST`| `/api/v1/customers/{id}/reactivate` | `ReactivateCustomerEndpoint`| `customers` (`archived_at=NULL`) | Account restored to active |
| **List Contacts** | `listContacts(id)` | `GET` | `/api/v1/customers/{id}/contacts` | `ListContactsEndpoint` | `contacts` | Array of customer contacts |
| **Add Contact** | `addContact(id, payload)` | `POST` | `/api/v1/customers/{id}/contacts` | `AddContactEndpoint` | `contacts` | Created contact entity |
| **Get Financial Profile**| `getFinancialProfile(id)` | `GET` | `/api/v1/customers/{id}/financial-profile`| `GetFinancialProfileEndpoint`| `customers` | Currency, terms, credit limit, credit status |
| **Update Financial Profile**|`updateFinancialProfile(id, p)`| `PUT`| `/api/v1/customers/{id}/financial-profile`| `UpdateFinancialProfileEndpoint`| `customers` | Updated financial terms |
| **Evaluate Intelligence**| `evaluateCustomerIntelligence(id)`| `POST`| `/api/v1/customers/{id}/intelligence/evaluate`| `EvaluateCustomerIntelligenceEndpoint`| `customer_health_evaluations`, `customer_risk_events`, Python sidecar | Intelligence profile with risks and opportunities |
| **Get Attention Items** | `getAttentionItems()` | `GET` | `/api/v1/customers/intelligence/attention`| `GetAttentionItemsEndpoint` | `customer_risk_events`, `customers` | Array of unaddressed customer risk events |
| **Resolve Risk Event** | `resolveCustomerRisk(id, rId, p)`| `POST`| `/api/v1/customers/{id}/risks/{risk_id}/resolve`| `ResolveCustomerRiskEndpoint`| `customer_risk_events` | Marks risk resolved with notes |

---

## 17. Frontend Component Map

- **`CustomersPage.jsx`**: Main directory landing page (`/dashboard/customers`).
  - *Responsibilities:* Renders header action bar (`Refresh Intelligence`, `+ New Customer`), 5 KPI metric cards (Healthy, Watch, At Risk, Critical, Opportunities), the Proactive Attention Required banner, the filter toolbar (search, status, type, country), and the paginated accounts table.
- **`CustomerDetailsPage.jsx`**: Comprehensive account cockpit (`/dashboard/customers/:id`).
  - *Responsibilities:* Orchestrates 9 sub-tabs (`Overview`, `Contacts`, `Addresses`, `Commercial`, `Operational`, `Financial`, `Documents`, `Timeline`, `Intelligence`). Integrates cross-module cards, document upload modals, and direct action triggers.
- **`CustomerModal.jsx`**: Modal dialog for creating and editing customer accounts.
  - *Responsibilities:* Form input for legal name, trading name, customer type, tax identifiers (GSTIN, PAN, EORI), currency, credit limits, payment terms, and initial contact details.
- **`LeadConversionModal.jsx`**: Dedicated dialog invoked when converting a Lead into a Customer.
  - *Responsibilities:* Runs automated duplicate detection, displays matching candidates with confidence percentages, and allows the user to choose between linking to an existing account or creating a new entity.
- **`CustomerIntelligence360Section.jsx`**: Embedded intelligence section within the Customer Detail page.
  - *Responsibilities:* Visualizes health score telemetry, contributing factors, active risk events, detected opportunities, and provides the risk resolution workflow.
- **`AutonomousCustomerRelationshipCard.jsx`**: Autonomous follow-up card.
  - *Responsibilities:* Displays governed follow-up status, channel preferences, draft generation, and human-in-the-loop review controls.

---

## 18. Go Backend Component Map

Located in `backend/internal/customers/`:
- **`transport.go`**: Configures HTTP routing using `chi.Router` and Go-kit HTTP transport. Decodes request parameters (pagination, filters, JSON payloads) and encodes responses. Enforces RBAC permissions (`ResourceCompanies`).
- **`endpoints.go`**: Defines Go-kit endpoint wrappers, connecting incoming HTTP requests to business logic methods.
- **`bl.go` (`BusinessLogic`)**: Implements business rules:
  - Input validation (non-empty legal name, uppercase currency codes, uppercase payment terms).
  - Database transactions spanning multiple tables during creation and lead conversion.
  - Recording structured audit entries via `audit.Record`.
- **`dl.go` (`DataLayer`)**: Executes SQL queries against MariaDB via `sqlx.DB`:
  - Parameterized filters preventing SQL injection.
  - Dynamic KPI aggregation and YTD revenue summation.
  - Sidecar communication with the Python AI service.
  - Synchronization of actionable risks into `ai_recommendations`.
- **`duplicate_engine.go`**: In-memory candidate evaluation calculating match confidence scores (0–100%) across Tax ID, corporate domain, company name, email, and phone number.
- **`customer_intelligence_engine.go`**: Deterministic heuristic rules computing customer health score, risk detection, opportunity identification, and activity trends.

---

## 19. AI & Customer Intelligence

The Customers module employs a hybrid intelligence model combining **deterministic Go heuristics** with **probabilistic Python AI reasoning**:

```
                          ┌────────────────────────────────┐
                          │   Authoritative MariaDB Data   │
                          │   (Bookings, RFQs, Invoices)   │
                          └───────────────┬────────────────┘
                                          │
                                          ▼
                          ┌────────────────────────────────┐
                          │ Deterministic Heuristic Engine │
                          │ (customer_intelligence_engine) │
                          │ • Health Score: 0-100          │
                          │ • Activity Trend: Stable/Dec   │
                          └───────────────┬────────────────┘
                                          │
                        Enriched Structured Payload
                                          │
                                          ▼
                          ┌────────────────────────────────┐
                          │     Python AI Sidecar (:8090)  │
                          │ /customer-relationship/evaluate│
                          │ • Strategic Risk Reasoning     │
                          │ • Follow-up Action Scoring     │
                          │ • Grounded Draft Generation    │
                          └───────────────┬────────────────┘
                                          │
                          Grounded Predictions / Drafts
                                          │
                                          ▼
                          ┌────────────────────────────────┐
                          │ Go Backend Validation & Sync   │
                          │ • Prevents direct DB mutation  │
                          │ • Inserts ai_recommendations   │
                          │ • Requires Human Approval      │
                          └────────────────────────────────┘
```

### Distinction of Information Classes:
1. **Authoritative Customer Data:** MariaDB persistent fields (`name`, `tax_id`, `credit_limit`, `currency`, `payment_terms`).
2. **Deterministic Derived Metrics:** Health Score (`0–100`), Health Status (`HEALTHY`, `WATCH`, `AT_RISK`, `CRITICAL`), Activity Trend (`INCREASING`, `STABLE`, `DECLINING`, `INACTIVE`).
3. **AI Predictions:** Churn risk probability, reorder momentum, lane loyalty scores.
4. **AI Recommendations:** Suggested follow-up drafts, contract renewal proposals, credit review alerts.

---

## 20. Python / Go Boundary

To maintain system integrity and security:
- **Go Backend Responsibility:** Handles authentication (`JWT`), authorization (`RBAC`), tenant isolation (`org_id`), persistence (`sqlx`), transaction rollbacks, external action dispatch, and audit logging.
- **Python AI Sidecar Responsibility:** Receives sanitized JSON contexts from Go via internal HTTP (`http://localhost:8090/customer-relationship/evaluate` and `/draft`). Executes sentiment analysis, prompt engineering, and risk reasoning.
- **Boundary Rule:** Python **never** connects directly to MariaDB and **never** mutates Customer records. All AI outputs are returned to Go as transient proposals requiring validation.

---

## 21. Action System

Customer-related actions flow through the unified LogisticsHQ Action System:
1. **Trigger:** User or AI identifies an action (e.g., "Send Account Follow-Up Note", "Resolve Commercial Credit Hold", "Initiate Contract Renewal").
2. **Policy Evaluation:** Checks user permissions (`ResourceCompanies:Update` or `ResourceCompanies:Delete`).
3. **Human-in-the-Loop Approval:** If an autonomous follow-up draft is generated, it remains in status `DRAFT` / `PENDING_APPROVAL` until approved by a commercial manager.
4. **Execution:** Dispatches message via internal notification channels or configured email providers.
5. **Persistence & Audit:** Updates `customer_followup_records` with timestamp, correlation ID, and audit trail.

---

## 22. Notifications

- **Attention Panel Alerts:** When a customer triggers a `CRITICAL` or `WARNING` risk event, an in-app badge appears on the Customers directory.
- **Follow-Up Reminders:** Synchronized with the centralized Notifications & Escalations Center (`notifications` table) when payment terms expire or quotations remain unanswered for >7 days.
- **External Email Delivery:** Draft messages prepared in `customer_followup_records` can be dispatched through live SMTP/SendGrid providers once configured. In development, drafts are logged locally without external side-effects.

---

## 23. Event Mesh & Automation

Customer state changes publish events to the internal Event Mesh:
- `customer.created`: Emitted when a new account is registered; initializes default communication preferences.
- `customer.updated`: Emitted on company name, owner, or credit limit modifications; invalidates cached metrics.
- `customer.lead_converted`: Emitted upon successful lead conversion; triggers outreach campaign attribution closure.
- `customer.risk_detected`: Emitted when an account enters `AT_RISK` or `CRITICAL` status; notifies account owner.

---

## 24. Permissions & RBAC

Customer operations are governed by the unified RBAC policy engine:

| Operation | Required Resource | Required Action | Minimum Role |
| :--- | :--- | :--- | :--- |
| **View Customers & Directory** | `companies` | `read` | Sales Representative / Operations Specialist |
| **View Customer 360 & Timeline** | `companies` | `read` | Sales Representative / Operations Specialist |
| **Create New Customer** | `companies` | `create` | Commercial Manager / Account Executive |
| **Convert Sales Lead** | `companies` | `create` | Commercial Manager / Account Executive |
| **Edit Customer Details** | `companies` | `update` | Commercial Manager / Super Admin |
| **Update Financial Profile** | `companies` | `update` | Finance Manager / Commercial Director |
| **Reassign Account Owner** | `companies` | `update` | Sales Director / Super Admin |
| **Resolve Risk Event** | `companies` | `update` | Account Manager / Operations Supervisor |
| **Archive Customer** | `companies` | `delete` | Super Admin / Org Admin |
| **Delete Customer Contact** | `companies` | `delete` | Commercial Manager / Super Admin |

---

## 25. Tenant Isolation

LogisticsHQ enforces strict multi-tenant isolation across every database operation:
- **Organization Boundary:** Every query includes `org_id = ?` injected from the authenticated user's session token (`userCtx.OrgID`).
- **Foreign Key Cascade:** All child tables (`contacts`, `customer_addresses`, `customer_lead_links`, `customer_health_evaluations`, `customer_risk_events`) enforce foreign key constraints with `ON DELETE CASCADE` referencing `organizations(id)`.
- **Cross-Tenant Leakage Prevention:** Even if an attacker guesses another tenant's Customer ID (`GET /api/v1/customers/1234`), the query `WHERE org_id = ? AND id = ?` returns `sql.ErrNoRows` (HTTP 404/403).

---

## 26. Audit Trail

Every state-altering operation in the Customers module writes an immutable log to `audit_logs`:
- **Customer Creation:** `Action = CREATE`, `ResourceType = 'CUSTOMER'`, includes customer name and initial parameters.
- **Customer Modification:** `Action = UPDATE`, captures full `Before` and `After` state snapshots.
- **Lead Conversion:** Records `Action = CREATE`, resource ID, and conversion link referencing the source Lead ID.
- **Account Ownership Change:** In addition to `audit_logs`, writes a permanent entry to `customer_ownership_history` recording previous owner, new owner, actor user ID, and reason.
- **Customer Archival / Reactivation:** Records `Action = DELETE` (archive) and `Action = ENABLE` (reactivate).

---

## 27. Search, Filter, Sort, and Pagination

The Customers directory provides comprehensive query controls:
- **Search (`search`):** Case-insensitive wildcard match (`LIKE %...%`) across `c.name`, `c.trading_name`, `c.customer_code`, `c.contact_name`, `c.contact_email`, `c.contact_phone`, `c.city`, and `c.country`.
- **Status Filter (`status`):** Filters by `ACTIVE`, `INACTIVE`, `PROSPECT`, `AT_RISK`, `CHURNED`, or `ALL`.
- **Customer Type Filter (`customer_type`):** Filters by `SHIPPER`, `CONSIGNEE`, `FORWARDER`, `3PL`, or `ALL`.
- **Country Filter (`country`):** Filters by specific country locations (e.g., India, UAE, Sweden, Singapore, USA).
- **Sorting (`sort_by`, `sort_order`):** Supports sorting by `name`, `code`, `status`, `type`, `health_score`, and `created_at` (`asc` / `desc`).
- **Pagination (`page`, `limit`):** Defaults to 10 records per page; computes `LIMIT %d OFFSET %d` in SQL and returns exact `total` count.

---

## 28. Business User Journeys

### Journey 1: Lead → Customer Relationship Establishment
1. A sales rep qualifies Lead `#1185` and clicks **Convert Lead**.
2. The system runs an automated duplicate scan; no duplicates found.
3. Rep selects Account Type **Shipper**, enters tax ID, and clicks **Create & Convert**.
4. System creates Customer Account `CUST-2026-09398`, establishes primary contact, sets lead status to `CONVERTED`, and redirects to the Customer 360 cockpit.

### Journey 2: Customer → RFQ → Quote → Booking Handover
1. An active shipper calls requesting a rate for 2x40HC containers from Rotterdam to Singapore.
2. Inside Customer 360, operations clicks **+ New RFQ**.
3. Commercial team generates Quotation `#QTN-2026-0042` with 14 free detention days.
4. Customer accepts the rate proposal; quote converts into confirmed Booking `#BK-2026-0019`.
5. All milestones appear immediately in the Customer's **Activity Timeline**.

### Journey 3: Credit Limit Governance & Risk Intervention
1. Finance notes an overdue balance for Customer `CUST-2026-0012`.
2. Finance manager updates credit status to `ON_HOLD` with commercial notes.
3. Intelligence engine immediately drops account health score from 85 to 55 (`AT_RISK`).
4. An attention badge appears on the directory: `Commercial Account On Credit Hold`.
5. Operations is barred from accepting unhedged bookings until finance clears the balance.

---

## 29. Data Flow Diagrams

### Customer Creation Data Flow
```
User (Browser)
      │ Submits Create Form
      ▼
CustomersPage / CustomerModal
      │ POST /api/v1/customers
      ▼
Go HTTP Transport (authMiddleware + RBAC Guard)
      │ Validates OrgID & Permissions
      ▼
Business Logic (bl.go)
      │ Validates name, normalizes currency & terms
      ▼
Data Layer (dl.go)
      │ BEGIN TRANSACTION
      │ INSERT INTO customers ...
      │ INSERT INTO contacts ...
      │ INSERT INTO customer_addresses ...
      │ COMMIT
      ▼
MariaDB Database
      │
      ├──> Universal Audit Trail (audit_logs)
      └──> Returns Customer Object
      ▼
Frontend UI updates with new Customer Account
```

### Intelligence & Health Evaluation Data Flow
```
Timer / User Trigger
      │ POST /api/v1/customers/:id/intelligence/evaluate
      ▼
Go Data Layer (dl.go)
      │ Queries MariaDB: RFQ count, Quote conversion, Shipments, Overdue balances
      ▼
Deterministic Evaluation (customer_intelligence_engine.go)
      │ Computes Health Score (0-100) & Contributing Factors
      ▼
HTTP POST to Python Sidecar (:8090/customer-relationship/evaluate)
      │ Structured context passed
      ▼
Python CustomerRelationshipAgent
      │ Evaluates complex risk patterns & draft communication
      ▼
JSON Response returned to Go Backend
      │ Validates response structure
      ▼
Go DL inserts customer_risk_events & ai_recommendations
      │
      ▼
Frontend Customer 360 displays updated Telemetry & Action Prompts
```

---

## 30. Source-of-Truth Matrix

| Information Item | Authoritative Source of Truth | Secondary / Derived Consumers | Persistence Mechanism |
| :--- | :--- | :--- | :--- |
| **Customer Legal Name** | `customers.name` | RFQs, Quotes, Invoices, Timeline | MariaDB (`VARCHAR(255)`) |
| **Customer Code** | `customers.customer_code` | Search index, UI chips, Documents | MariaDB (`VARCHAR(50)`) |
| **Operational Status** | `customers.status` | Navigation guards, Directory table | MariaDB (`VARCHAR(50)`) |
| **Credit Status & Limits** | `customers.credit_status` & `credit_limit`| Invoicing guards, Attention panel | MariaDB (`DECIMAL(14,2)`) |
| **Primary Contact** | `contacts` (`is_primary = TRUE`) | Email reply drafts, Table rows | MariaDB (`contacts` table) |
| **Health Score & Status** | `customer_health_evaluations` | Header stat badges, Directory pills | Deterministic Engine $\rightarrow$ MariaDB |
| **Risk Flags & Attention Items**| `customer_risk_events` | Attention Panel, Intelligence Tab | Heuristics + AI $\rightarrow$ MariaDB |
| **Commercial Opportunities** | `customer_opportunity_events`| Opportunities badge, Sales alerts | Heuristics + AI $\rightarrow$ MariaDB |
| **AI Follow-Up Message Draft** | `customer_followup_records` | Follow-up Drawer, Approval Center | Python AI Sidecar $\rightarrow$ MariaDB |
| **YTD Revenue** | `shipment_customer_invoices` | Metric cards, Financial profile | Dynamic SQL Aggregate |

---

## 31. Error, Loading, and Empty States

Verified directly in `CustomersPage.jsx` and `CustomerDetailsPage.jsx`:
- **Loading State:** Displays animated spinner with status text: `"Evaluating Customer Intelligence & Activity Signals..."`.
- **Zero Customers (Global Empty State):** Displays `ModuleHeroEmptyState` featuring an enterprise shield icon, explanatory value propositions (360° Health, Credit Governance, Document Master), and two action buttons: `Create First Customer` and `Import from Leads`.
- **Search / Filter Zero Results:** Displays targeted empty state card: `"No Customer Accounts Found. There are no customer profiles matching your active search query or filter criteria"` with `Clear Filters` and `Create First Customer` buttons.
- **API Failure State:** Displays red alert icon, error message `"Unable to load customer directory. Please check network connection."`, and an interactive `Retry Connection` button.
- **Unauthorized Access (401/403):** Intercepted by `api.js` redirecting to login or displaying access denied warning.

---

## 32. UI / UX Observations

Visual review from live Chrome rendering (`customers_page_live.png` and `customer_detail_live.png`):

### What Works Well:
- The top 5 KPI cards (`Healthy Accounts`, `Under Watch`, `At Risk Accounts`, `Critical Accounts`, `Detected Opportunities`) give commercial executives an instant pulse of customer portfolio health.
- The `Attention Required Panel` provides clean severity tabs (Critical, Warnings, Attention) and highlights accounts requiring immediate intervention.
- The Customer 360 cockpit cleanly separates commercial, operational, financial, and compliance data into intuitive tabs with zero clutter.
- Inline avatar pills with dynamic color hashing create distinct visual separation between companies.

### Observations for Future Improvement:
- **SHOULD IMPROVE:** In `CustomersPage.jsx`, the country filter is currently hardcoded to 5 countries (`India`, `UAE`, `Sweden`, `Singapore`, `USA`). It should dynamically load distinct countries from the database.
- **SHOULD IMPROVE:** The Customer 360 header breadcrumb contains an icon without an explicit "Back" tooltip. Adding an explicit back button improves navigation speed.
- **OPTIONAL:** Provide bulk CSV export capability on the main Customer Directory table.

---

## 33. Responsive and Zoom Observations

Verified across 4 viewports using Chrome DevTools Protocol device metric overrides:
- **100% Zoom (1440x900):** Layout renders with complete sidebar, 5-column KPI grid, and full table width without horizontal body overflow.
- **110% Zoom Equivalent (1280x800):** KPI grid cleanly adjusts; attention panel cards stack vertically. No text clipping observed.
- **125% Zoom Equivalent (1024x768):** Navigation remains responsive; customer table maintains horizontal scrollability within its parent container (`.cust-table-card`), preventing viewport clipping.
- **Tablet Portrait (768x1024):** Sidebar collapses into mobile trigger; action buttons in table wrap cleanly into action menus.

---

## 34. Security Architecture

1. **Authentication:** All customer endpoints require verified Bearer JWT tokens in the `Authorization` header.
2. **RBAC Guardrails:** Enforced at Go HTTP routing layer (`r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, ...))`). Unauthorized tokens receive HTTP 403 Forbidden.
3. **Multi-Tenant Isolation:** `org_id` is extracted strictly from server-side JWT context and never accepted from untrusted client query parameters.
4. **Parameterized SQL Queries:** All database interactions use `sqlx` parameterized arguments (`?`), preventing SQL injection attacks.
5. **AI Sidecar Isolation:** Python sidecar communicates over a private localhost network with mandatory machine-to-machine key authentication (`X-LogisticsHQ-Service-Key`). Python never connects to the database directly.

---

## 35. Business + Technical Glossary

- **Customer:** A legally verified trading entity or counterparty purchasing freight logistics services.
- **Shipper:** The merchant, exporter, or manufacturer consigning freight for shipment.
- **Consignee:** The receiver, importer, or buyer to whom cargo is delivered.
- **Customer 360:** The comprehensive single-pane view synthesizing commercial, operational, financial, and intelligence telemetry for a customer.
- **Health Score:** A normalized integer rating (0–100) reflecting an account's operational momentum, inquiry frequency, and financial standing.
- **Credit Status:** Commercial status governing credit risk (`GOOD_STANDING`, `REVIEW_REQUIRED`, `ON_HOLD`, `NO_CREDIT_LIMIT`).
- **Lead Conversion:** The deterministic transition of a sales inquiry into an authorized Customer entity.
- **Duplicate Confidence Score:** A 0–100% rating computed across tax identifiers, domains, names, and contact parameters to prevent duplicated account creation.
- **Customer Lead Link:** An immutable audit record linking an original lead ID to the resulting customer ID.

---

## 36. One-Page “How Customers Work” Summary

```
                      POTENTIAL BUSINESS PROSPECT
                                   │
                                   ▼
                             SALES LEAD
                      (Acquisition & Qualification)
                                   │
                     Convert Lead to Customer Action
                                   ▼
                   CUSTOMER COMMERCIAL ACCOUNT (360°)
            ┌──────────────────────┴──────────────────────┐
            ▼                                             ▼
   Commercial Workflow                           Operational Fulfillment
  • RFQ Intake                                  • Carrier Bookings
  • Rate Quotations                             • Live Shipment Tracking
  • Service Contracts                           • Document Compliance
            │                                             │
            └──────────────────────┬──────────────────────┘
                                   │
                                   ▼
                          FINANCE & BILLING
                      • Payment Terms (NET30)
                      • Credit Limit Controls
                      • Customer Invoicing
                                   │
                                   ▼
                 ONGOING RELATIONSHIP GOVERNANCE (AI)
                  • Automated Health Scoring (0-100)
                  • Proactive Risk & Attention Alerts
                  • Governed Follow-Up Actions
```

### Technical Summary
1. Client issues `POST /api/v1/customers` or `POST /api/v1/customers/convert-lead`.
2. Go HTTP transport validates JWT and checks RBAC permission `ResourceCompanies`.
3. Business logic validates inputs and executes multi-table transaction in MariaDB (`customers`, `contacts`, `customer_addresses`, `customer_lead_links`).
4. Universal audit trail logs the event.
5. Customer 360 queries synthesize RFQs, quotes, shipments, and billing data in real-time.
6. Deterministic heuristics and Python AI sidecar evaluate health and risk signals without direct database mutation.
7. Frontend renders unified, responsive commercial cockpit.

---

## 37. Technical Traceability Matrix

| Customer Capability | Frontend Component | API Endpoint | Go Service / BL | Python / AI Sidecar | MariaDB Table | Event Mesh / Workflow | Action System | RBAC Permission |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Directory Listing** | `CustomersPage.jsx` | `GET /api/v1/customers` | `bl.ListCustomers` | N/A | `customers` | N/A | N/A | `companies:read` |
| **KPI Metrics** | `CustomersPage.jsx` | `GET /api/v1/customers/kpis` | `bl.GetCustomerKPIs` | N/A | `customers`, `contracts` | N/A | N/A | `companies:read` |
| **Customer 360 View** | `CustomerDetailsPage.jsx` | `GET /api/v1/customers/{id}/dashboard`| `bl.GetCustomer360Dashboard` | N/A | `customers`, `rfqs`, `quotations`, `bookings`, `shipments` | N/A | N/A | `companies:read` |
| **Duplicate Scan** | `LeadConversionModal.jsx` | `POST /api/v1/customers/check-duplicate`| `bl.CheckDuplicate` | N/A | `customers` | N/A | N/A | `companies:read` |
| **Customer Creation** | `CustomerModal.jsx` | `POST /api/v1/customers` | `bl.CreateCustomer` | N/A | `customers`, `contacts`, `customer_addresses` | `customer.created` | N/A | `companies:create` |
| **Lead Conversion** | `LeadConversionModal.jsx` | `POST /api/v1/customers/convert-lead` | `bl.ConvertLeadToCustomer` | N/A | `customers`, `customer_lead_links`, `leads` | `customer.lead_converted`| N/A | `companies:create` |
| **Update Customer** | `CustomerModal.jsx` | `PUT /api/v1/customers/{id}` | `bl.UpdateCustomer` | N/A | `customers` | `customer.updated` | N/A | `companies:update` |
| **Manage Contacts** | `CustomerDetailsPage.jsx` | `POST/PUT/DEL /customers/{id}/contacts`| `bl.Add/Update/DeleteContact`| N/A | `contacts` | N/A | N/A | `companies:create/update/delete` |
| **Financial Profile** | `CustomerDetailsPage.jsx` | `GET/PUT /customers/{id}/financial-profile`| `bl.Get/UpdateFinancialProfile`| N/A | `customers` | N/A | N/A | `companies:read/update` |
| **Intelligence Eval** | `CustomerIntelligence360Section.jsx`| `POST /customers/{id}/intelligence/evaluate`| `bl.EvaluateAndPersistCustomerIntelligence`| `/customer-relationship/evaluate` | `customer_health_evaluations`, `customer_risk_events` | `customer.risk_detected` | `ai_recommendations` | `companies:read` |
| **Resolve Risk Event**| `CustomerIntelligence360Section.jsx`| `POST /customers/{id}/risks/{rId}/resolve`| `bl.ResolveCustomerRisk` | N/A | `customer_risk_events` | N/A | N/A | `companies:update` |
| **Autonomous Follow-Up**| `AutonomousCustomerRelationshipCard.jsx`| `POST /autonomy/customers/{id}/records/{recId}/send`| `autonomy.Service` | `/customer-relationship/draft` | `customer_followup_records` | `customer.followup_sent` | Action Approval Engine | `companies:update` |

---

## 38. Known Gaps

1. **UI/UX Gap:** Hardcoded country filter in `CustomersPage.jsx` (currently lists India, UAE, Sweden, Singapore, USA). Should be dynamically populated from distinct `country` values in `customers`.
2. **UI/UX Gap:** Customer 360 sub-tabs do not persist their active state in URL query params (e.g., `?tab=commercial`), causing the view to reset to `overview` upon browser reload.
3. **External Integration Gap:** Live outbound email/SMS delivery for follow-up message drafts depends on configured SMTP/SendGrid credentials. In development environments without live SMTP, drafts remain persisted in `customer_followup_records` without live dispatch.

---

## 39. Verification Status

- **Codebase & Schema Inspection:** Confirmed across Go files (`bl.go`, `dl.go`, `types.go`, `transport.go`, `duplicate_engine.go`, `customer_intelligence_engine.go`), MariaDB migrations (`069`, `070`, `071`, `090`, `114`), and Python sidecar (`customer_relationship/agent.py`).
- **Live Backend Verification:** Verified running Go server on `:8080` (PID 26640) returned HTTP 200 with 9 registered customer records, valid KPIs, and zero unhandled exceptions.
- **Live MariaDB Verification:** Verified running database on `:3306` (PID 12988) with active schema tables (`customers`, `contacts`, `customer_addresses`, `customer_lead_links`, `customer_health_evaluations`, `customer_risk_events`).
- **Live Frontend & Browser Verification:** Executed headless Google Chrome session via CDP at 1440x900; captured live screenshots (`customers_page_live.png` and `customer_detail_live.png`) confirming full visual rendering of KPI cards, attention panel, customer directory table, 360 overview, predictive intelligence card, and activity timeline.
- **Multi-Tenant & Security Verification:** Confirmed strict `org_id` filtering on all queries and RBAC permission checks (`ResourceCompanies`).

**FINAL STATUS:**  
`PASS — CUSTOMERS WORKFLOW DOCUMENTED`
