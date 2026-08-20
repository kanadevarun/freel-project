# LogisticsHQ / Freel — Architecture & Data Model Investigation Report
**Date:** August 15, 2026  
**Investigation Mode:** READ-ONLY Deep Dive (Zero Database or Code Mutations)  
**Target Focus:** Lifecycle analysis of Organization, User, Company, Customer, Lead, and RFQ entities, with complete root cause analysis of the `rfqs_customer_id_fkey` foreign key violation during inbound email parsing.

---

## Executive Summary

An exhaustive audit of the Go backend (`backend/`), database migrations (`backend/internal/database/migrations/`), and Python AI sidecar (`ai_sidecar/`) reveals a **fundamental architectural conflation between `Lead` and `Customer`**.

### The Root Cause of `pq: insert or update on table "rfqs" violates foreign key constraint "rfqs_customer_id_fkey"`:
1. **Schema Definition:** `rfqs.customer_id` has a strict Foreign Key constraint: `REFERENCES customers(id)`.
2. **Entity Mismatch:** Inbound emails create records in the **`leads`** table (`leads.id = 6`).
3. **The Flaw:** In `ai_sidecar/app/agents/email_parser_agent.py` (line 297) and `ai_sidecar/app/tools/leads_tool.py` (line 16), the code assumes `lead_id` is interchangeable with `customer_id` (`customer_id=int(lead_id)`).
4. **The Failure:** When Go attempts `INSERT INTO rfqs (..., customer_id) VALUES (..., 6)`, PostgreSQL rejects the statement because ID `6` exists in `leads`, but **not in `customers`**.

---

## 1. Deep Dive: Entity by Entity Analysis

### 1.1 ORGANIZATION (`organizations`)
*   **Table Schema (`001_rbac.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `name`: `VARCHAR(255) NOT NULL`
    *   `created_at`, `updated_at`: `TIMESTAMPTZ`
*   **What it Represents:**
    *   The **Freight Forwarding Enterprise / Workspace Tenant** (e.g. *"Freel Global Logistics Pvt Ltd"*).
    *   It is the root multi-tenant partition key (`org_id`) across all business data (`users`, `companies`, `customers`, `leads`, `rfqs`, `rates`, `shipments`).
*   **How it is Created:**
    *   Programmatically in `internal/organization/service.go` via `CreateOrganization(ctx, req)`.
    *   Seeded in `backend/cmd/seed/main.go` via `seedOrganisation()`.
*   **How Users are Associated:**
    *   Users are **not** linked via a column on the `users` table.
    *   Users belong to an organization through the join table **`org_members`**:
        `org_members (org_id REFERENCES organizations(id), user_id REFERENCES users(id), role_id REFERENCES roles(id), status)`.

---

### 1.2 USER (`users`)
*   **Table Schema (`001_rbac.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `cognito_sub`: `VARCHAR(255) UNIQUE NOT NULL`
    *   `email`: `VARCHAR(255) UNIQUE NOT NULL`
    *   `first_name`, `last_name`: `VARCHAR(255)`
    *   `created_at`, `updated_at`: `TIMESTAMPTZ`
*   **What it Represents:**
    *   An **internal employee/operator** of the Freight Forwarder (e.g. CEO, Sales Rep, Pricing Manager, Ops Specialist).
*   **How Users are Created & Signup Lifecycle:**
    1. **Signup Initiation (`internal/auth/service.go` line 36):**
       `auth.Service.Signup()` registers the user in AWS Cognito (`s.client.SignUp`) storing `custom:company_name` and `custom:role`.
       *(Note: Signup does not directly insert into the local Postgres `users` table until confirmation/invitation).*
    2. **Invite Acceptance (`internal/auth/service.go` line 235):**
       `auth.Service.AcceptInvite()` registers the user in Cognito, then in a single database transaction inserts into `users` (`INSERT INTO users (cognito_sub, email, first_name)...`) and inserts into `org_members`.
    3. **Cognito Token / Sub Mapping (`internal/middleware/auth.go` line 95–111):**
       On every authenticated API call, JWT middleware validates the Cognito bearer token, extracts `cognitoSub := token.Subject()`, and resolves:
       ```sql
       SELECT u.id AS user_id, om.org_id, r.name AS role
       FROM users u
       JOIN org_members om ON u.id = om.user_id
       JOIN roles r ON om.role_id = r.id
       WHERE u.cognito_sub = $1 AND om.status = 'ACTIVE'
       LIMIT 1
       ```
    4. **Super Admin / Role Assignment:**
       The seed script (`cmd/seed/main.go`) provisions `CEO` (role `ADMIN`), `SALES`, `PRICING`, and `CUSTOMER_CONTACT`.

---

### 1.3 COMPANY (`companies`)
*   **Table Schema (`002_companies_leads.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `org_id`: `BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
    *   `name`: `VARCHAR(255) NOT NULL` (e.g. *"Tata Exports Ltd"*, *"Sun Pharma"*)
    *   `domain`: `VARCHAR(255)` (e.g. *"tata-exports.local"*)
    *   `industry`: `VARCHAR(255)`
    *   `address_id`: `BIGINT REFERENCES addresses(id)`
*   **What it Represents:**
    *   An **external commercial business entity** (shipper, consignee, manufacturer, carrier).
*   **How Companies are Created Today:**
    *   **Only in seed data (`cmd/seed/main.go` lines 202–205).**
    *   There is currently **no runtime service, repository, or REST API** in `backend/internal/` that creates or manages `companies` records. (Only RBAC constants exist: `ResourceCompanies = "COMPANIES"`).

---

### 1.4 CUSTOMER (`customers`)
*   **Table Schema (`002_companies_leads.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `org_id`: `BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
    *   `company_id`: `BIGINT NOT NULL REFERENCES companies(id)`
    *   `status`: `VARCHAR(50) DEFAULT 'ACTIVE'` (`ACTIVE`, `CUSTOMER`, `LOST`)
*   **What a Customer Represents in Freel:**
    *   A **Customer** is an **account relationship record** that establishes that a specific `Company` has been approved/onboarded as an active client of the Freight Forwarder (`Organization`).
    *   `Customer` is **NOT** a user that signed up.
    *   `Customer` is **NOT** a lead.
*   **How Customer Records are Created:**
    *   **Crucial Finding: There is ZERO customer creation code in the entire application runtime.**
    *   `INSERT INTO customers` appears in only **one file in the whole repository**: `backend/cmd/seed/main.go` (line 217).
    *   There is no `internal/customers/` module, no `CreateCustomer` endpoint, and no automated conversion from company/lead to customer.

---

### 1.5 LEAD (`leads`)
*   **Table Schema (`008_leads.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `org_id`: `BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
    *   `company_name`: `VARCHAR(255) NOT NULL`
    *   `contact_name`: `VARCHAR(255)`
    *   `email`: `VARCHAR(255)`
    *   `phone`: `VARCHAR(50)`
    *   `status`: `VARCHAR(50) DEFAULT 'NEW'` (`NEW`, `QUALIFIED`, `ACTIVE`, `CONVERTED`, `REJECTED`)
    *   `source`: `VARCHAR(100)` (`EMAIL`, `MANUAL`, `IMPORT`)
    *   `ai_score`: `INT DEFAULT 0`
    *   `ai_research_report`: `TEXT`
*   **What a Lead Represents:**
    *   A prospective contact or shipper who submitted an inbound email or query, but has **not** yet been vetted, contracted, or converted into an official `Company` + `Customer` account.
*   **How Leads are Created:**
    1. **Inbound Email:** `POST /api/v1/emails/inbound` (`internal/leads/email_handler.go` line 61–77) auto-creates a lead if no lead matches `req.From`.
    2. **Manual Sales CRM:** `POST /api/v1/leads` via `internal/leads/bl.go`.
    3. **CSV Import:** `POST /api/v1/leads/import` via `internal/leads/bl.go`.
*   **Lead Conversion Flow:**
    *   **There is currently NO lead-to-company or lead-to-customer conversion logic anywhere in the Go codebase.**

---

### 1.6 RFQ (`rfqs`)
*   **Table Schema (`004_rfq.sql`):**
    *   `id`: `BIGSERIAL PRIMARY KEY`
    *   `org_id`: `BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE`
    *   `rfq_number`: `VARCHAR(50) NOT NULL UNIQUE (org_id, rfq_number)`
    *   `customer_id`: `BIGINT NOT NULL REFERENCES customers(id)` **<-- STRICT FK**
    *   `stage`: `VARCHAR(50) DEFAULT 'STAGE_RFQ_CREATED'`
    *   `origin`, `destination`, `incoterms`, `target_date`
    *   `sales_assignee_id`, `pricing_assignee_id`
*   **All Code Paths That Create RFQs:**
    1. **Manual / Frontend (`POST /api/v1/rfqs` via `internal/rfq/transport.go`):**
       User enters a `customer_id` integer in the UI (`RFQBuilder.jsx`), which is passed directly to `rfq.BusinessLogic.CreateRFQ`.
    2. **Inbound Email Tool (`POST /internal/rfqs/from-email` via `internal/leads/email_handler.go`):**
       Receives `CreateRFQFromEmailRequest` from Python AI sidecar, which calls `rfq.BusinessLogic.CreateRFQ`.
    3. **Database Seeder (`cmd/seed/main.go`):**
       Inserts seed RFQs linked to seeded `customers[i].ID`.

---

## 2. Inbound Email → RFQ Execution Flow (Current vs. Failure Point)

```
[Inbound Email Arrives]
  │  From: shipper_complete@tata-exports.local
  ▼
[EmailHandler.InboundEmailWebhook] (backend/internal/leads/email_handler.go)
  │  1. GetLeadByEmail() -> Not found
  │  2. CreateLead("Inbound Lead (shipper_complete@tata-exports.local)") -> Returns lead.ID = 6
  │  3. LogInteraction() -> Inserts into lead_interactions (ID=1, lead_id=6)
  │  4. CreateAITask() -> Inserts EMAIL_PARSE task (entity_id="1", payload: {"lead_id": 6, ...})
  ▼
[Queue Worker & Python SalesAgent] (ai_sidecar/app/agents/email_parser_agent.py)
  │  1. classify_and_parse_email_node() -> Extracts origin, destination, weight, incoterms
  │  2. check_completeness_node() -> Validates all mandatory fields are present
  │  3. draft_rfq_node() (Line 297):
  │        Invokes create_rfq_from_email_tool.func(
  │            org_id=5,
  │            customer_id=int(lead_id),  <--- BUG: Passing lead_id (6) as customer_id
  │            ...
  │        )
  ▼
[Go Backend Tool Endpoint] (POST /internal/rfqs/from-email)
  │  EmailHandler.CreateRFQFromEmail() receives {"customer_id": 6, ...}
  │  Calls rfqBL.CreateRFQ(ctx, {CustomerID: 6, ...})
  │  Executes: INSERT INTO rfqs (org_id, customer_id, ...) VALUES (5, 6, ...)
  ▼
[PostgreSQL Database Engine]
  💥 FOREIGN KEY CONSTRAINT VIOLATION:
     "pq: insert or update on table \"rfqs\" violates foreign key constraint \"rfqs_customer_id_fkey\""
     (Because customers table has rows 1..5, but NO row with id=6!)
```

---

## 3. Direct Answers to Core Architectural Questions

### Question 8: "Should `lead.id` ever be directly used as `rfqs.customer_id`?"
**NO. Never.**
*   `leads` and `customers` are completely separate database tables with separate primary key sequences.
*   A `lead` is an unverified prospect from an email, whereas a `customer` is a formal client account linked to a `company`.
*   Passing `lead.id` into `customer_id` is an invalid foreign key reference that will fail on any fresh database.

### Question 9: "What does 'customer' mean in Freel?"
**Answer: C) A commercial client entity (`Company`) with an active account relationship under a Freight Forwarder (`Organization`).**
*   **It is NOT** a signed-up login user (login users are employees of the freight forwarder organization).
*   **It is NOT** a lead (leads are prospective inquiries).
*   **It IS** the corporate client (shipper, importer, exporter) for whom quotes and shipments are executed.

---

## 4. Architectural Entity Relationship Diagram

### 4.1 Actual Current Database Schema (As Implemented in SQL Migrations)

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                               ORGANIZATIONS (Tenant)                            │
│                             id: BIGSERIAL PRIMARY KEY                           │
└────────┬──────────────────────────┬─────────────────────────────┬───────────────┘
         │                          │                             │
         │ (1:N)                    │ (1:N)                       │ (1:N)
         ▼                          ▼                             ▼
┌──────────────────┐       ┌──────────────────┐          ┌────────────────────────┐
│     USERS        │       │    COMPANIES     │          │         LEADS          │
│ id, cognito_sub  │       │ id, name, domain │          │ id, company_name, email│
└────────┬─────────┘       └────────┬─────────┘          └───────────┬────────────┘
         │                          │                                │ (1:N)
         │ (N:M via org_members)    │ (1:1)                          ▼
         ▼                          ▼                         ┌───────────────────┐
┌──────────────────┐       ┌──────────────────┐               │ LEAD_INTERACTIONS │
│   ORG_MEMBERS    │       │    CUSTOMERS     │               │ id, lead_id, raw  │
│ user_id, org_id  │       │ id, company_id   │               └───────────────────┘
└──────────────────┘       └────────┬─────────┘
                                    │
                                    │ (1:N)
                                    ▼
                           ┌──────────────────┐
                           │       RFQS       │
                           │ id, customer_id  │ ◄─── (Strict FK: customers.id)
                           └──────────────────┘
```

---

## 5. Summary Tables

### Table A: Entity Metadata Table

| Entity | Table Name | Primary Key | Parent / Foreign Keys | Created By | Current Purpose in Codebase |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Organization** | `organizations` | `id` (BIGINT) | None | `organization.Service` / Seeder | Tenant partition root (Freight Forwarder workspace). |
| **User** | `users` | `id` (BIGINT) | Linked via `org_members` | `auth.Service.AcceptInvite` / Seeder | Freight forwarder staff with login credentials and RBAC roles. |
| **Role & Member**| `roles`, `org_members` | `id` (BIGINT) | `org_id`, `user_id`, `role_id` | `auth.Service`, `users.Repo` | RBAC permissions mapping users to org capabilities. |
| **Company** | `companies` | `id` (BIGINT) | `org_id` | **Seeder Only** (No runtime API) | Master corporate entity representing shipper/client companies. |
| **Customer** | `customers` | `id` (BIGINT) | `org_id`, `company_id` | **Seeder Only** (No runtime API) | Formal billing account status for a Company. Required by RFQs. |
| **Lead** | `leads` | `id` (BIGINT) | `org_id` | `leads.EmailHandler`, `leads.BL` | Prospective inquiry or sender of inbound emails. |
| **Interaction** | `lead_interactions` | `id` (BIGINT) | `org_id`, `lead_id`, `linked_rfq_id`| `leads.EmailHandler` | Thread history and raw communication records. |
| **RFQ** | `rfqs` | `id` (BIGINT) | `org_id`, `customer_id`, `sales_assignee_id` | `rfq.BL`, `EmailHandler` | Request for Quotation workflow record. |

---

### Table B: Relationship & Discrepancy Table

| Relationship | Current Implementation | Expected / Intended Domain Meaning | Problem Identified? |
| :--- | :--- | :--- | :--- |
| **Lead ↔ Customer** | Python passes `lead_id` as `customer_id` | A Lead is an unverified prospect; a Customer is an active commercial account | 🔴 **CRITICAL BUG**: Foreign key constraint failure (`rfqs_customer_id_fkey`). |
| **Inbound Email ↔ Customer** | Webhook creates a `Lead`, but RFQ requires a `Customer` | Inbound RFQ should either link to existing Customer by domain, auto-create a Customer account, or allow `rfqs.lead_id` | 🔴 **MISSING LINK**: No customer resolution or auto-provisioning logic exists on email ingestion. |
| **Company ↔ Customer** | `customers.company_id REFERENCES companies(id)` | 1 Company has 1 Customer account status in the org | ⚠️ **ORPHANED SCHEMA**: No runtime service exists to create `companies` or `customers`. |
| **RFQ ↔ Customer** | `rfqs.customer_id BIGINT NOT NULL REFERENCES customers(id)` | RFQ belongs to an existing client customer account | ⚠️ **RESTRICTIVE FK**: Blocks guest/inbound quote requests from new leads who are not yet seeded customers. |

---

## 6. Exact Files and Functions Directory

| Operation | Implementation File | Exact Function / Method |
| :--- | :--- | :--- |
| **Signup (Cognito)** | `backend/internal/auth/service.go` | `func (s *Service) Signup(ctx, req)` (Line 36) |
| **User Creation (DB)** | `backend/internal/auth/service.go`<br>`backend/internal/users/repository.go` | `func (s *Service) AcceptInvite(ctx, req)` (Line 235)<br>`func (r *repositoryImpl) CreateUser(ctx, user)` (Line 196) |
| **Company Creation** | `backend/cmd/seed/main.go` | `func (s *seeder) seedCustomers(ctx, orgID)` (Line 183) |
| **Customer Creation** | `backend/cmd/seed/main.go` | `func (s *seeder) seedCustomers(ctx, orgID)` (Line 217) |
| **Lead Creation** | `backend/internal/leads/bl.go`<br>`backend/internal/leads/email_handler.go` | `func (b *businessLogic) CreateLead(ctx, req)` (Line 41)<br>`func (h *EmailHandler) InboundEmailWebhook(w, r)` (Line 66) |
| **Lead Conversion** | *None* | *No implementation exists in current codebase.* |
| **Manual RFQ Creation** | `backend/internal/rfq/bl.go` | `func (b *businessLogic) CreateRFQ(ctx, req)` (Line 57) |
| **Email Ingestion Webhook**| `backend/internal/leads/email_handler.go` | `func (h *EmailHandler) InboundEmailWebhook(w, r)` (Line 46) |
| **AI Sidecar RFQ Creation**| `ai_sidecar/app/agents/email_parser_agent.py`<br>`ai_sidecar/app/tools/leads_tool.py` | `def draft_rfq_node(state)` (Line 271)<br>`def create_rfq_from_email_tool(...)` (Line 10) |
| **Internal RFQ from Email**| `backend/internal/leads/email_handler.go` | `func (h *EmailHandler) CreateRFQFromEmail(w, r)` (Line 182) |

---

## 7. Recommended Architectural Solutions for Inbound Email Flow

To allow inbound email RFQs to succeed cleanly without making arbitrary assumptions or fabricating invalid data, there are two viable architectural patterns:

### Recommended Option A: Dynamic Lead-to-Customer Resolution (Enterprise CRM Pattern)
When an inbound email is parsed and ready to create an RFQ:
1. Extract the sender's domain (e.g. `tata-exports.local`) and company name.
2. In Go backend (`CreateRFQFromEmail`):
   * Check if a `company` exists with this domain under the `org_id`. If not, insert into `companies (org_id, name, domain)`.
   * Check if a `customer` record exists for this `company_id`. If not, insert into `customers (org_id, company_id, status='ACTIVE')`.
   * Use this valid `customer.id` for `rfqs.customer_id`.
   * Update the `lead.status = 'CONVERTED'` and link `lead_interactions.linked_rfq_id = rfq.id`.

### Recommended Option B: Direct Lead Support on RFQs (Flexible Logistics Pattern)
Make `customer_id` nullable on `rfqs` and add `lead_id`:
1. Modify schema:
   * `ALTER TABLE rfqs ALTER COLUMN customer_id DROP NOT NULL;`
   * `ALTER TABLE rfqs ADD COLUMN lead_id BIGINT REFERENCES leads(id) ON DELETE SET NULL;`
2. If RFQ comes from an established customer, populate `customer_id`.
3. If RFQ comes from an unverified inbound email, populate `lead_id`. When the quote is won/booked, convert the lead to a customer account.

---
*Report generated in READ-ONLY mode. No codebase or database alterations were made during this audit.*
