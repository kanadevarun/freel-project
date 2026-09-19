# Task 3.2.A — LogisticsHQ Leads Module: Business Workflow, System Architecture, Data Flow, Database Mapping, AI Integration, and Comprehensive Technical Documentation

**Target System:** LogisticsHQ Freight Forwarding & Logistics Operating System  
**Document Version:** 1.0.0 (Authoritative Leads Implementation Audit)  
**Verification Date:** 2026-09-12  
**Operating Environment:** Windows Server / AMD64  
- Go Backend API (Port :8080)
- Python FastAPI AI Sidecar (Port :8090)
- React 19 + Vite Frontend (Port :5173)
- MariaDB 10.11 / MySQL Storage Engine (Port :3306)  
**Final Status:** **PASS — LEADS WORKFLOW DOCUMENTED**

---

## Table of Contents
1. [Executive Summary](#1-executive-summary)
2. [Leads in Plain English](#2-leads-in-plain-english)
3. [Business Purpose & Commercial Value](#3-business-purpose--commercial-value)
4. [Lead Lifecycle & State Machine](#4-lead-lifecycle--state-machine)
5. [Lead Sources](#5-lead-sources)
6. [Lead Creation Workflow](#6-lead-creation-workflow)
7. [Lead Edit Workflow](#7-lead-edit-workflow)
8. [Lead Status Workflow](#8-lead-status-workflow)
9. [Lead Conversion Workflow (Customer & RFQ)](#9-lead-conversion-workflow-customer--rfq)
10. [Database Table Mapping](#10-database-table-mapping)
11. [Lead → Customer Database Relationship](#11-lead--customer-database-relationship)
12. [Lead → RFQ Relationship](#12-lead--rfq-relationship)
13. [API Mapping & Endpoint Inventory](#13-api-mapping--endpoint-inventory)
14. [Frontend Component Architecture Map](#14-frontend-component-architecture-map)
15. [Go Backend Component Architecture Map](#15-go-backend-component-architecture-map)
16. [AI / Lead Intelligence Architecture](#16-ai--lead-intelligence-architecture)
17. [Python AI Sidecar / Go Backend Boundary](#17-python-ai-sidecar--go-backend-boundary)
18. [Action System Connection & Governance](#18-action-system-connection--governance)
19. [Notifications & Alerting Framework](#19-notifications--alerting-framework)
20. [Event Mesh & Asynchronous Background Automation](#20-event-mesh--asynchronous-background-automation)
21. [Permissions & Role-Based Access Control (RBAC)](#21-permissions--role-based-access-control-rbac)
22. [Tenant Isolation Architecture](#22-tenant-isolation-architecture)
23. [Audit Trail & Governance Logging](#23-audit-trail--governance-logging)
24. [Search, Filtering, Sorting & Pagination](#24-search-filtering-sorting--pagination)
25. [Business User Journeys](#25-business-user-journeys)
26. [Data Flow Diagrams](#26-data-flow-diagrams)
27. [Source-of-Truth Matrix](#27-source-of-truth-matrix)
28. [Loading, Error, and Empty State Lifecycle](#28-loading-error-and-empty-state-lifecycle)
29. [UI/UX Observations & Ergonomics Review](#29-uiux-observations--ergonomics-review)
30. [Responsive & Browser Zoom Behavior](#30-responsive--browser-zoom-behavior)
31. [Security Architecture & Boundary Enforcement](#31-security-architecture--boundary-enforcement)
32. [Business & Technical Glossary](#32-business--technical-glossary)
33. [One-Page Summary: “How Leads Work in LogisticsHQ”](#33-one-page-summary-how-leads-work-in-logisticshq)
34. [Technical Traceability Matrix](#34-technical-traceability-matrix)
35. [Known Gaps & Deficiencies](#35-known-gaps--deficiencies)
36. [Verification Status & Sign-Off](#36-verification-status--sign-off)

---

## 1. Executive Summary

This document establishes the definitive, source-code-verified technical and business reference for the **Leads & Inbound Inquiries Module** of LogisticsHQ. Every workflow, data model, API contract, state transition, AI sidecar integration, and database query documented herein was verified directly against the active codebase across:
- **Frontend:** `frontend/src/pages/dashboard/Leads/` (`LeadsPage.jsx`, `LeadDetailPanel.jsx`, `AddLeadModal.jsx`, `EditLeadModal.jsx`, `ImportLeadsModal.jsx`), `frontend/src/pages/dashboard/Customers/LeadConversionModal.jsx`, `frontend/src/pages/dashboard/RFQ/RFQBuilder.jsx`, and `frontend/src/services/leadsService.js`.
- **Go Backend:** `backend/internal/leads/` (`transport.go`, `endpoints.go`, `bl.go`, `dl.go`, `email_handler.go`, `spec/`), `backend/internal/customers/bl.go`, `backend/internal/rfq/bl.go`, `backend/internal/jobs/lead_worker.go`, `backend/internal/actions/sales_actions.go`, and `backend/internal/server/routes.go`.
- **Python AI Sidecar:** `ai_sidecar/main.py` (`/leads/score-lead`, `/leads/classify-email`), `ai_sidecar/app/agents/lead_scoring_agent.py`, `ai_sidecar/app/agents/email_classifier_agent.py`, and prompt template engines.
- **Database Engine:** MariaDB 10.11 storage tables (`leads`, `lead_interactions`, `lead_email_drafts`, `customer_lead_links`, `customers`, `customer_contacts`, `rfqs`, `approval_requests`, `audit_logs`).

The Leads module acts as the commercial top-of-funnel engine for freight forwarders. It ingests shipper prospects from manual entry, bulk CSV imports, email mailboxes, and marketing outreach campaigns. Each prospect undergoes automated trade intelligence enrichment and AI Ideal Customer Profile (ICP) scoring, participates in multi-turn inbound/outbound communication tracking, and converts seamlessly into an active Customer Account or priced commercial RFQ.

---

## 2. Leads in Plain English

For a non-technical freight forwarder or logistics business owner:

### What is a Lead?
A **Lead** is a commercial inquiry or prospective shipper who has expressed interest in moving freight but has not yet booked a shipment or established an active credit account. A lead represents potential revenue: a manufacturing exporter requesting container rates, an import procurement manager asking for drayage, or a prospective customer imported from a trade directory.

### Why do Leads exist in LogisticsHQ?
Freight forwarding sales teams receive dozens of inquiries daily via phone, email, forms, and referrals. Without a dedicated Leads pipeline:
1. Inquiries get lost in individual email inboxes.
2. Sales representatives waste hours manually researching whether a prospect is a legitimate shipper or spam.
3. Pricing teams waste time quoting unvetted companies that lack cargo volume.
4. Follow-up is inconsistent, causing lost deals.

LogisticsHQ centralizes every prospect into an organized sales workbench where shippers are qualified, scored, communicated with, and handed off to commercial operations with complete context.

---

## 3. Business Purpose & Commercial Value

The Leads module delivers measurable commercial benefits across four operational pillars:

1. **Inbound Ingestion & Inbox Consolidation:** Ingests shipper inquiries from multiple channels (connected Gmail/Outlook mailboxes, website webhooks, manual entries, CSV bulk uploads, and outreach campaigns) into a single unified queue.
2. **Automated Trade Intelligence & ICP Scoring:** Instantly researches prospective shippers against maritime and trade databases to determine monthly shipping volume (TEUs), commodity profile, and export frequency. An AI sales specialist scores the prospect (0–100) and produces an executive research brief for sales dispatchers.
3. **Conversational Email Intelligence & Clarification:** Parses incoming customer emails, extracts origin/destination ports and dates, detects missing freight requirements (e.g. Incoterms, cargo weight), drafts professional clarification responses, and stages them for human approval.
4. **Frictionless Commercial Handoff:** Converts qualified leads into persistent Customer Accounts (with automated duplicate detection) or directly launches shipment RFQs, carrying over all extracted container specifications and conversation history with zero re-entry.

---

## 4. Lead Lifecycle & State Machine

The Leads lifecycle follows an explicit, five-stage state machine enforced by backend business logic:

```
                  ┌──────────────┐
                  │   INBOUND    │  (Manual Entry, CSV Import,
                  │  PROSPECT    │   Mailbox Sync, Outreach)
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │     NEW      │  <─── Default state upon creation
                  └──────┬───────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ▼              ▼              ▼
   ┌─────────────┐ ┌─────────────┐ ┌─────────────┐
   │  QUALIFIED  │ │ IN_PROGRESS │ │  REJECTED   │
   └──────┬──────┘ └──────┬──────┘ └─────────────┘
          │               │              ▲
          └───────┬───────┘              │
                  │ (Commercial Fit)     │ (Disqualified / Low Score)
                  ▼                      │
           ┌─────────────┐               │
           │  CONVERTED  │───────────────┘
           └─────────────┘
                  │
        ┌─────────┴─────────┐
        ▼                   ▼
┌───────────────┐   ┌───────────────┐
│   CUSTOMER    │   │  FREIGHT RFQ  │
│    ACCOUNT    │   │  INITIATION   │
└───────────────┘   └───────────────┘
```

### State Definitions:
1. **`NEW`:** Newly ingested prospect. Awaiting initial sales review, verification, or automated AI scoring completion.
2. **`QUALIFIED`:** The prospect has verified freight demand, appropriate volume, and a valid business identity matching the freight forwarder's operating trade corridors.
3. **`IN_PROGRESS`:** The sales representative is actively communicating with the prospect via phone, email, or meeting to clarify shipment specifications.
4. **`REJECTED`:** The lead is disqualified (e.g., spam, consumer inquiry, non-serviceable trade lane, credit risk, or automated AI score < 50).
5. **`CONVERTED`:** Terminal commercial success state. The lead has converted into an active Customer Account in the master customer directory and/or resulted in a formal freight RFQ. The lead record is preserved permanently for historical conversion reporting.

---

## 5. Lead Sources

The Leads module natively tracks and persists the commercial acquisition source in `leads.source`:

| Source Key | Display Label | Ingestion Mechanism | Business Meaning | Downstream Automation |
|:---|:---|:---|:---|:---|
| **`MANUAL`** | Manual Entry | UI Modal (`AddLeadModal.jsx`) | Sales dispatcher manually inputs a phone lead or walking inquiry. | Emits `lead.created` $\rightarrow$ Triggers background AI scoring & trade intel research. |
| **`IMPORT`** | CSV Import | Bulk Upload (`ImportLeadsModal.jsx`) | Marketing or procurement imports conference/directory lead sheets. | Batch inserts rows in MariaDB $\rightarrow$ Enqueues AI scoring for each prospect. |
| **`EMAIL`** | Inbound Email | Mailbox Webhook / IMAP Sync | Inbound shipper email received at connected company freight mailbox. | Parses email body $\rightarrow$ Extracts freight parameters $\rightarrow$ Generates AI clarification draft. |
| **`OUTREACH`**| Campaign | Outreach Module (`/dashboard/outreach`) | Cold sales email campaign response converted into a high-intent inquiry. | Links `campaign_id` and records `converted_from_outreach_at` timestamp. |
| **`WEBSITE`** | Web Inbound | Public Trade Intelligence Scaffold | Prospective shipper fills out quotation request on public marketing page. | Ingests shipper contact info $\rightarrow$ Flags for urgent dispatcher qualification. |
| **`REFERRAL`**| Partner / Agent | Manual / API | Existing customer or overseas forwarding partner refers a cargo shipper. | Tagged for high-priority sales attention. |

---

## 6. Lead Creation Workflow

```
[ User / Source ] 
       │
       ▼ (Submits Form or CSV or Inbound Webhook)
[ Frontend: AddLeadModal.jsx / leadsService.createLead ]
       │
       ▼ (HTTP POST /api/v1/leads with Bearer JWT)
[ Go Router: transport.go / decodeCreateLeadRequest ]
       │
       ▼ (Validates CompanyName != "", AssignedTo user in org)
[ Go Business Logic: bl.go:CreateLead ]
       │
       ├─► [ MariaDB: INSERT INTO leads ] (Status: 'NEW', org_id from JWT)
       ├─► [ MariaDB: INSERT INTO lead_tags ] (Optional tags)
       ├─► [ MariaDB: INSERT INTO lead_activities ] (Timeline: 'CREATED')
       ├─► [ MariaDB: INSERT INTO audit_logs ] (Actor: User, Action: CREATE)
       │
       ▼ (Publishes Event)
[ EventBus: events.EventLeadCreated ]
       │
       ▼ (Asynchronous Background Listener)
[ Go Job: leadWorker.handleLeadCreated ]
       │
       ├─► [ Trade Intel Engine: EnrichCompany ] (Fetches TEU volume, suppliers)
       ├─► [ Python AI Sidecar: POST /leads/score-lead ] (Calculates ICP score 0-100)
       ├─► [ MariaDB: UPDATE leads SET ai_score, ai_research_report ]
       ├─► [ MariaDB: INSERT INTO lead_activities ] (Timeline: 'AI_ENRICHED')
       └─► [ EventBus: events.EventLeadEnriched ]
```

---

## 7. Lead Edit Workflow

### Editable Fields & Business Meaning:
- **`company_name` (String):** The legal or trade business name of the shipping entity.
- **`contact_name` (String):** Primary point of contact (e.g. Logistics Director, Procurement Officer).
- **`email` (String):** Contact email used for RFQ delivery and thread tracking.
- **`phone` (String):** Phone number for urgent operational dispatch inquiries.
- **`location` (String):** Primary facility or cargo origin country/city.
- **`assigned_to` (Integer User ID):** Sales representative responsible for account ownership.
- **`notes` (Text):** Internal operational notes, cargo peculiarities, or delivery constraints.
- **`tags` (String Array):** Categorical tags (e.g., `Perishables`, `Automotive`, `High-Volume`).

### Technical Update Trace:
1. **Frontend:** Dispatched via `EditLeadModal.jsx` or inline edits in `LeadDetailPanel.jsx` calling `leadsService.updateLead(id, payload)`.
2. **Go Transport:** `PUT /api/v1/leads/{id}` decoded by `transport.go:decodeUpdateLeadRequest`.
3. **Business Logic (`bl.go:UpdateLead`):**
   - Validates resource exists within tenant (`WHERE id = ? AND org_id = ?`).
   - If `assigned_to` changed: checks `UserExistsInOrg`, sets `assigned_at = NOW()`, records `OWNER_CHANGED` timeline activity.
   - If `status` changed: records `STATUS_CHANGED` timeline activity. If status becomes `CONVERTED`, executes `EnsureCustomerForLead`.
   - Persists modified fields to MariaDB `leads` table.
   - Logs audit trail entry (`ActionUpdate`, `ModuleLeads`).

---

## 8. Lead Status Workflow

Allowed transitions and their operational side effects:

| Current Status | Allowed Next Status | Operational Meaning | Authorized Roles | Automated Side Effect |
|:---|:---|:---|:---|:---|
| **`NEW`** | `QUALIFIED` | Prospect verified as legitimate freight customer. | Sales Rep, Dispatcher, Manager, Super Admin | Timeline logged; tagged as ready for commercial quote. |
| **`NEW`** | `IN_PROGRESS` | Outreach initiated; awaiting shipment details. | Sales Rep, Dispatcher, Manager, Super Admin | Timeline logged; owner assigned. |
| **`NEW`** | `REJECTED` | Ineligible cargo, spam, or low ICP fit (<50). | Sales Rep, Dispatcher, Manager, Super Admin | Timeline logged; excluded from active sales pipeline metrics. |
| **`QUALIFIED`** | `IN_PROGRESS` | Commercial terms under active negotiation. | Sales Rep, Manager, Super Admin | Timeline logged. |
| **`QUALIFIED`** | `CONVERTED` | Handoff to active customer account or RFQ. | Sales Rep, Manager, Super Admin | Triggers `EnsureCustomerForLead` / creates customer link; lead preserved. |
| **`QUALIFIED`** | `REJECTED` | Deal lost to competitor or commercial impasse. | Sales Rep, Manager, Super Admin | Disqualification reason logged. |
| **`IN_PROGRESS`** | `QUALIFIED` | Specifications verified; ready for pricing. | Sales Rep, Manager, Super Admin | Timeline logged. |
| **`IN_PROGRESS`** | `CONVERTED` | RFQ built or master customer profile created. | Sales Rep, Manager, Super Admin | Triggers customer creation and linked RFQ association. |
| **`IN_PROGRESS`** | `REJECTED` | Shipper cancelled tender or unresponsive. | Sales Rep, Manager, Super Admin | Timeline logged. |
| **`REJECTED`** | `NEW` / `QUALIFIED`| Lead reactivated following market re-engagement. | Manager, Super Admin | Timeline activity: `REACTIVATED`. |
| **`CONVERTED`** | *(Terminal)* | Commercial relationship established. | Locked | Read-only state; permanent link to Customer and RFQ records. |

---

## 9. Lead Conversion Workflow (Customer & RFQ)

LogisticsHQ supports two distinct, tightly coordinated conversion pathways:

### Pathway A: Convert Lead to Customer Account

```
[ LeadDetailPanel / LeadsPage ] 
       │
       ▼ (Clicks "Convert to Customer")
[ LeadConversionModal.jsx ] 
       │
       ├─► [ Automated Duplicate Check: customerService.checkDuplicate ]
       │     (Scans name, email, phone against MariaDB customers table)
       │
       ▼ (Operator selects: "Create New Customer" OR "Link to Existing Account")
[ POST /api/v1/customers/convert-lead ]
       │
       ▼ (customers/bl.go:ConvertLeadToCustomer)
       │
       ├── Starts Database Transaction (tx)
       ├── Scenario 1 (New): INSERT INTO customers (org_id, name, type, status='ACTIVE')
       │                     INSERT INTO customer_contacts (org_id, customer_id, primary)
       ├── Scenario 2 (Link): Resolves existing customer_id; attaches secondary contact
       ├── INSERT INTO customer_lead_links (org_id, customer_id, lead_id, converted_by, notes)
       ├── UPDATE leads SET status = 'CONVERTED' WHERE id = lead_id AND org_id = org_id
       ├── Commits Database Transaction (tx)
       │
       ▼
[ Result: Lead remains in leads table with status='CONVERTED'; Customer created with code 'CUST-2026-XXXXX' ]
```

### Pathway B: Convert Lead to Freight RFQ

```
[ LeadDetailPanel / LeadsPage ]
       │
       ▼ (Clicks "Convert to RFQ" / "Create RFQ")
[ RFQBuilder.jsx (Modal Drawer) ]
       │
       ├─► Pre-populates shipper name, contact email, origin/destination ports
       │   and container specs extracted by AI from email interactions.
       │
       ▼ (Operator reviews shipment items and clicks "Submit RFQ")
[ POST /api/v1/rfqs (with payload.lead_id) ]
       │
       ▼ (rfq/bl.go:CreateRFQ)
       │
       ├── Validates / creates customer association
       ├── INSERT INTO rfqs (org_id, customer_id, lead_id, origin, destination, status='DRAFT')
       ├── INSERT INTO rfq_items (container sizes, weight, volume)
       ├── rfq.ConvertLead: UPDATE leads SET status = 'CONVERTED' WHERE id = lead_id
       ├── Publishes events.EventRFQCreated on EventBus
       │
       ▼
[ Result: RFQ generated (e.g. RFQ-2026-1002); Lead status updated to CONVERTED; linked_rfq_id populated ]
```

---

## 10. Database Table Mapping

The Leads module relies on 8 tightly coupled MariaDB tables:

| Table Name | Business Purpose | Primary Key | Tenant Key | Critical Fields | Foreign Keys / References |
|:---|:---|:---|:---|:---|:---|
| **`leads`** | Master registry of prospective shippers. | `id` (INT) | `org_id` (INT) | `company_name`, `contact_name`, `email`, `phone`, `status`, `source`, `ai_score`, `ai_research_report`, `assigned_to` | `assigned_to` $\rightarrow$ `users(id)`, `campaign_id` $\rightarrow$ `outreach_campaigns(id)` |
| **`lead_interactions`**| Inbound/outbound email and phone logs. | `id` (BIGINT)| `org_id` (INT) | `lead_id`, `direction`, `channel`, `subject`, `content`, `intent`, `sentiment`, `ai_confidence`, `partial_rfq_context` | `lead_id` $\rightarrow$ `leads(id)` |
| **`lead_email_drafts`**| AI-generated draft replies awaiting approval.| `id` (BIGINT)| `org_id` (INT) | `lead_id`, `parent_interaction_id`, `subject`, `content`, `status`, `approval_id` | `parent_interaction_id` $\rightarrow$ `lead_interactions(id)`, `approval_id` $\rightarrow$ `approval_requests(id)` |
| **`customer_lead_links`**| Traceability bridge between Lead and Customer. | `id` (BIGINT)| `org_id` (INT) | `customer_id`, `lead_id`, `converted_by_user_id`, `conversion_notes`, `created_at` | `customer_id` $\rightarrow$ `customers(id)`, `lead_id` $\rightarrow$ `leads(id)` |
| **`lead_tags`** | Keyword classification tags. | Composite | N/A | `lead_id`, `tag` | `lead_id` $\rightarrow$ `leads(id)` |
| **`lead_activities`**| Chronological audit and event timeline. | `id` (INT) | `org_id` (INT) | `entity_id`, `action`, `description`, `actor_user_id`, `created_at` | `entity_id` $\rightarrow$ `leads(id)`, `actor_user_id` $\rightarrow$ `users(id)` |
| **`customers`** | Authoritative active business accounts. | `id` (BIGINT)| `org_id` (INT) | `customer_code`, `name`, `customer_type`, `status`, `payment_terms`, `credit_limit` | Scoped to `org_id` |
| **`rfqs`** | Freight price quotation requests. | `id` (INT) | `org_id` (INT) | `rfq_number`, `lead_id`, `customer_id`, `origin`, `destination`, `status` | `lead_id` $\rightarrow$ `leads(id)`, `customer_id` $\rightarrow$ `customers(id)` |

---

## 11. Lead → Customer Database Relationship

- **Relationship Cardinality:** $1 : 1$ or $N : 1$ (Multiple inquiries from the same corporate shipper can map to a single master Customer Account).
- **Linking Table:** `customer_lead_links` provides complete historical auditability. It tracks which user converted the lead, the exact timestamp of conversion, and operator notes.
- **Lead Record Preservation:** When a lead converts to a customer, the row in `leads` is **never deleted**. Its status is updated to `CONVERTED`, preserving the original sales attribution, initial email thread, and trade intelligence report for long-term funnel analytics.
- **Automatic Contact Creation:** The conversion transaction creates a corresponding contact record in `customer_contacts` (or `contacts`), ensuring shipping coordinators immediately inherit the prospect's email and phone number.

---

## 12. Lead → RFQ Relationship

- **Relationship Cardinality:** $1 : N$ (A single lead can spawn multiple shipment RFQs over time).
- **Direct Foreign Key:** The `rfqs` table contains an explicit `lead_id` column (`INT NULL`, indexed).
- **Bidirectional Traceability in SQL:**
  When a lead is fetched via `dl.go:GetByID`, it executes:
  ```sql
  SELECT l.*, r.id as linked_rfq_id, r.rfq_number as linked_rfq_number
  FROM leads l
  LEFT JOIN rfqs r ON r.lead_id = l.id AND r.org_id = l.org_id
  WHERE l.id = ? AND l.org_id = ?
  ORDER BY r.created_at DESC LIMIT 1;
  ```
  This immediately alerts the dispatcher if an RFQ has already been initiated for this lead, preventing duplicate quote requests.

---

## 13. API Mapping & Endpoint Inventory

| Business Operation | Frontend Function (`leadsService.js`) | HTTP Method | API Endpoint | Go Handler / Service | MariaDB Operations | Expected Result |
|:---|:---|:---:|:---|:---|:---|:---|
| **List Leads** | `listLeads(params)` | `GET` | `/api/v1/leads` | `transport.go:ListLeadsEP` $\rightarrow$ `dl.go:List` | `SELECT ... FROM leads WHERE org_id = ?` | Paginated array of leads with counts. |
| **Create Lead** | `createLead(payload)` | `POST` | `/api/v1/leads` | `transport.go:CreateLeadEP` $\rightarrow$ `bl.go:CreateLead` | `INSERT INTO leads`, `INSERT INTO lead_activities` | Created lead object; emits `lead.created`. |
| **Get Lead Details** | `getLead(id)` | `GET` | `/api/v1/leads/{id}` | `transport.go:GetLeadEP` $\rightarrow$ `dl.go:GetByID` | `SELECT ... FROM leads LEFT JOIN rfqs ...` | Full lead details with linked RFQ & owner. |
| **Update Lead** | `updateLead(id, payload)` | `PUT` | `/api/v1/leads/{id}` | `transport.go:UpdateLeadEP` $\rightarrow$ `bl.go:UpdateLead` | `UPDATE leads SET ... WHERE id = ? AND org_id = ?`| Updated lead object; logs activity. |
| **Delete Lead** | `deleteLead(id)` | `DELETE` | `/api/v1/leads/{id}` | `transport.go:DeleteLeadEP` $\rightarrow$ `bl.go:DeleteLead` | `DELETE FROM leads WHERE id = ? AND org_id = ?` | HTTP 200 `{ success: true }`. |
| **Bulk Import Leads**| `importLeads(file)` | `POST` | `/api/v1/leads/import` | `transport.go:ImportLeadsEP` $\rightarrow$ `bl.go:BulkCreate`| Batch `INSERT INTO leads` | Imported count & failed rows report. |
| **Bulk Update Leads**| `bulkUpdateLeads(payload)` | `POST` | `/api/v1/leads/bulk` | `transport.go:BulkUpdateLeadsEP` $\rightarrow$ `bl.go:BulkUpdate` | Batch `UPDATE leads` | Count of updated leads. |
| **Get Lead Timeline**| `getLeadTimeline(id)` | `GET` | `/api/v1/leads/{id}/timeline` | `transport.go:GetLeadTimelineEP` $\rightarrow$ `dl.go:GetActivities`| `SELECT ... FROM lead_activities` | Chronological activity stream. |
| **Get Interactions** | `getLeadInteractions(id)` | `GET` | `/api/v1/leads/{id}/interactions` | `email_handler.go:GetInteractions` | `SELECT ... FROM lead_interactions` | Array of email & call logs. |
| **Inbound Email Hook**| External / Postfix / SES | `POST` | `/api/v1/emails/inbound` | `email_handler.go:InboundEmailWebhook` | `INSERT INTO lead_interactions`, `INSERT INTO leads`| Ingests raw email; queues AI parsing. |
| **AI Sales Callback** | Python Sidecar | `POST` | `/internal/sales/callback` | `email_handler.go:SalesCallback` | `UPDATE lead_interactions`, `INSERT INTO lead_email_drafts`| Persists intent, sentiment, and AI draft. |
| **Get Email Draft** | `getEmailDraft(leadId, iId)` | `GET` | `/api/v1/leads/{id}/interactions/{i_id}/draft` | `email_handler.go:GetDraft` | `SELECT ... FROM lead_email_drafts` | Draft content, status, and approval ID. |
| **Save Email Draft** | `saveEmailDraft(leadId, ...)`| `PUT` | `/api/v1/leads/{id}/interactions/{i_id}/draft` | `email_handler.go:SaveDraft` | `UPDATE lead_email_drafts` | Updated draft text. |
| **Approve AI Draft** | `approveClarificationDraft` | `POST` | `/api/v1/leads/{id}/interactions/{i_id}/approve-draft`| `email_handler.go:ApproveDraft` | `UPDATE lead_email_drafts SET status='SENT'` | Outbound email dispatched; interaction logged. |
| **Reject AI Draft** | `rejectClarificationDraft` | `POST` | `/api/v1/leads/{id}/interactions/{i_id}/reject-draft` | `email_handler.go:RejectDraft` | `UPDATE lead_email_drafts SET status='REJECTED'` | Draft discarded; approval rejected. |
| **Convert to Customer**| `customerService.convertLead`| `POST` | `/api/v1/customers/convert-lead`| `customers/bl.go:ConvertLeadToCustomer` | `INSERT INTO customers`, `INSERT INTO customer_lead_links` | Customer created; lead marked `CONVERTED`. |
| **Convert to RFQ** | `rfqService.createRFQ` | `POST` | `/api/v1/rfqs` | `rfq/bl.go:CreateRFQ` | `INSERT INTO rfqs`, `UPDATE leads SET status='CONVERTED'`| RFQ created; lead marked `CONVERTED`. |

---

## 14. Frontend Component Architecture Map

```
  [ LeadsPage.jsx ]  (Route: /dashboard/leads)
         │
         ├── TopBar & PageHeader (Title, Real-time status, Subtitle)
         ├── Pipeline Stat Cards (All, New, Qualified, In Progress, Converted)
         │
         ├── Filter & Search Toolbar
         │     ├── Search Input (Debounced company, contact, or email search)
         │     ├── Status Tabs (All Leads, New, In Progress, Qualified, Converted, Rejected)
         │     ├── Source Dropdown Filter (Manual, Email, Import, Website, Outreach)
         │     └── Action Buttons (+ Add Lead, 📥 Import CSV, Ask Copilot)
         │
         ├── Leads Table (.leads-table)
         │     ├── Bulk Action Bar (Multi-select, Bulk Assign, Bulk Status Change, Bulk Delete)
         │     ├── Table Headers (Company, Contact, Status, Source, AI Fit, Assigned To, Created)
         │     ├── Table Body Rows (Clicking opens React Portal Drawer)
         │     └── Pagination Bar (Limit, Offset, Previous/Next buttons)
         │
         ├── Lead Creation Modal (<AddLeadModal.jsx />)
         ├── CSV Bulk Import Modal (<ImportLeadsModal.jsx />)
         ├── Lead to Customer Modal (<LeadConversionModal.jsx />)
         ├── Lead to RFQ Modal (<RFQBuilder.jsx />)
         │
         └── React Portal: Lead Detail Drawer (<LeadDetailPanel.jsx />)
               │
               ├── Drawer Header (Lead Name, Status Badge, Close Button)
               ├── Action Command Bar (Convert to RFQ, Convert to Customer, Edit, Delete)
               │
               ├── Tab 1: Overview (<OverviewTab />)
               │     ├── Contact & Shipper Attributes Card
               │     ├── AI Fit & Trade Intelligence Brief (Score 0-100, Executive Summary)
               │     ├── RFQ Readiness Checklist (Origin, Dest, Weight, Volume, Ready Date, Incoterms)
               │     └── Internal Dispatcher Notes & Tag Manager
               │
               ├── Tab 2: Emails & Interactions (<EmailsTab />)
               │     ├── Conversation Thread Selector
               │     ├── Inbound/Outbound Message Stream
               │     ├── AI Clarification Draft Card (Approve, Edit, Reject buttons)
               │     └── Manual Reply Composer
               │
               └── Tab 3: Timeline & Audit (<TimelineTab />)
                     └── Chronological timeline of creation, status changes, emails, and assignments.
```

---

## 15. Go Backend Component Architecture Map

```
  [ internal/server/routes.go ]
         │
         ├── Route Registration: r.Route("/api/v1/leads", ...)
         │     ├── AddLeadsHandlers(r, s.leadsEndpoints, authGuard.RequireAuth)
         │     └── Email Interaction Handlers (/leads/{id}/interactions/...)
         │
         ▼
  [ internal/leads/transport.go ]
         │
         ├── Request Decoders: decodeCreateLeadRequest, decodeListLeadsRequest, decodeImportLeadsRequest
         ├── Response Encoders: encodeAPIResponse (Standard JSON format: { data, total_count })
         └── Error Encoder: encodeErrorResponse (Maps service errors to HTTP 400, 401, 403, 404, 500)
         │
         ▼
  [ internal/leads/bl.go (Business Logic) ]
         │
         ├── Lead Validation: CompanyName presence, email formatting, user tenancy check
         ├── Lifecycle State Engine: Transition validation and side-effect execution
         ├── Event Publishing: Dispatches EventLeadCreated, EventLeadEnriched to EventBus
         ├── Audit Recording: Universal structured audit logging
         └── Mailbox & Clarification Engine: Bridges inbound emails and AI drafts
         │
         ▼
  [ internal/leads/dl.go (Data Access Layer) ]
         │
         ├── Direct SQL Queries: Parameterized SQL against MariaDB
         ├── Multi-Tenant Scoping: Injects AND org_id = ? into every SELECT, UPDATE, DELETE
         ├── Relational Joins: Joins users, rfqs, outreach_campaigns, and lead_tags
         └── Transaction Management: Atomic BeginTx and Commit for multi-table updates
```

---

## 16. AI / Lead Intelligence Architecture

LogisticsHQ implements a specialized **Lead Intelligence Engine** operating across two primary capabilities:

### 1. Inbound Shipper Intent & Cargo Parsing (`EmailClassifierAgent`)
- **Input:** Raw incoming email body, subject line, sender address.
- **Processing:** Invoked by the Go email ingestion pipeline against Python AI Sidecar (`POST /leads/classify-email`).
- **AI Reasoning:**
  - Categorizes intent: `RFQ_REQUEST_COMPLETE`, `RFQ_REQUEST_INCOMPLETE`, `GENERAL_INQUIRY`, `SPAM`, `NOT_LOGISTICS`.
  - Determines sentiment: `POSITIVE`, `NEUTRAL`, `URGENT`.
  - Extracts structured cargo entities: Origin Port, Destination Port, Container Type (20GP, 40HC), Cargo Weight, Volume, Target Date, Incoterms.
- **Output:** Persisted in `lead_interactions.partial_rfq_context` and displayed in the RFQ Readiness checklist.

### 2. Lead ICP Scoring & Trade Intelligence (`LeadScoringAgent`)
- **Input:** Company Name, Industry, Monthly Shipping Volume (TEU), Exporter Status, Top Suppliers.
- **Processing:** Triggered by `leadWorker` listening for `EventLeadCreated` via `POST /leads/score-lead`.
- **AI Reasoning:**
  - Evaluates commercial viability based on recurring shipping volume, route alignment, and cargo complexity.
  - Assigns an integer score from `0` to `100`.
  - Generates a concise executive brief advising dispatchers how to approach the prospect.
- **Safety Boundary:**
  - **Authoritative Data:** Shipper contact info, verified MariaDB records.
  - **AI Prediction:** ICP Fit Score (0–100) and extracted container specifications.
  - **AI Recommendation:** Drafted clarification email. The draft is staged as `AWAITING_APPROVAL` and **never sent autonomously**.

---

## 17. Python AI Sidecar / Go Backend Boundary

The architecture enforces strict segregation of duties between Go and Python:

```
┌────────────────────────────────────────────────────────┐
│               GO BACKEND (PORT 8080)                   │
│                                                        │
│  - Holds exclusive write access to MariaDB             │
│  - Validates Cognito JWTs & enforces RBAC permissions  │
│  - Binds org_id to every tenant query                  │
│  - Governs the Action System & Human Approvals         │
│  - Transmits external emails via SMTP / SES / Gmail    │
│  - Records persistent immutable Audit Logs             │
└──────────────────────────┬─────────────────────────────┘
                           │ HTTP REST (Port :8090)
                           │ Header: X-Internal-Service-Key
                           ▼
┌────────────────────────────────────────────────────────┐
│            PYTHON AI SIDECAR (PORT 8090)               │
│                                                        │
│  - Stateless LLM inference & LangGraph agents          │
│  - Email classification & cargo entity extraction      │
│  - ICP Lead scoring & executive summary synthesis      │
│  - Proposes structured draft replies                   │
│  - CANNOT write directly to MariaDB business tables    │
└────────────────────────────────────────────────────────┘
```

---

## 18. Action System Connection & Governance

The Leads module connects to the LogisticsHQ Action System via registered actions in `backend/internal/actions/sales_actions.go`:

1. **`sales.convert_lead`:**
   - **Description:** Converts a sales lead to a customer account.
   - **Category:** `ActionCategoryWrite`
   - **Required Permission:** `LEADS:UPDATE`
   - **Execution:** Invokes `leadsBL.UpdateLead` with status `CONVERTED` and ensures customer creation.
2. **`sales.approve_draft`:**
   - **Description:** Approves an AI-generated clarification email draft.
   - **Category:** `ActionCategoryWrite`
   - **Required Permission:** `APPROVALS:UPDATE` or `LEADS:UPDATE`
   - **Execution:** Gated by the Human-in-the-Loop approval request (`approval_requests` table). Upon approval, dispatches the email via Go's mail service.

---

## 19. Notifications & Alerting Framework

Leads activities generate real-time alerts through the Notification Center:
- **New Inbound Lead Alert:** Emitted when a prospect is ingested from an email mailbox or website inquiry.
- **Draft Awaiting Approval:** Dispatches an in-app priority alert to the assigned sales manager when an AI clarification email is staged for review.
- **Lead Assigned Notification:** Notifies the designated sales representative when a prospect is assigned to their account.

---

## 20. Event Mesh & Asynchronous Background Automation

The Leads module integrates with the internal Go EventBus (`backend/internal/common/events/`):

| Event Constant | Event Name | Published By | Subscribed By | Operational Effect |
|:---|:---|:---|:---|:---|
| `EventLeadCreated` | `"lead.created"` | `bl.go:CreateLead` | `leadWorker` | Launches trade intelligence enrichment and AI ICP scoring. |
| `EventLeadEnriched`| `"lead.enriched"`| `leadWorker` | UI Timeline / Notification | Logs timeline entry; updates frontend lead badge. |
| `EventEmailReceived`| `"email.received"`| Mailbox Sync Worker | `EmailHandler` | Creates interaction log; queues AI email parsing. |
| `EventRFQCreated` | `"rfq.created"` | `rfq/bl.go:CreateRFQ` | Pricing Agent, Dashboard | Begins pricing workflow; updates conversion counters. |

---

## 21. Permissions & Role-Based Access Control (RBAC)

The Leads module enforces granular RBAC permissions (`ResourceLeads = "LEADS"`):

| Permission | Business Meaning | UI Controls Enabled | API Route Protected |
|:---|:---|:---|:---|
| **`LEADS:READ`** | View leads, pipeline stats, and customer inquiry history. | View Leads Table, Open Lead Detail Drawer, View Timeline & Email threads. | `GET /api/v1/leads`, `GET /api/v1/leads/{id}`, `GET /api/v1/leads/{id}/interactions` |
| **`LEADS:CREATE`**| Create manual leads or upload CSV batches. | `+ Add Lead` button, `📥 Import CSV` button, Save Lead action. | `POST /api/v1/leads`, `POST /api/v1/leads/import` |
| **`LEADS:UPDATE`**| Edit lead attributes, assign owners, change status, and convert. | Status dropdown, Edit Lead Modal, Convert to RFQ, Convert to Customer buttons. | `PUT /api/v1/leads/{id}`, `POST /api/v1/leads/bulk`, `POST /api/v1/customers/convert-lead` |
| **`LEADS:DELETE`**| Remove disqualified or duplicate leads from the system. | Delete Lead button (in drawer and bulk action bar). | `DELETE /api/v1/leads/{id}` |

---

## 22. Tenant Isolation Architecture

Strict multi-tenant boundary isolation is maintained across all Leads layers:
1. **JWT Extraction:** The user's authenticated organization ID is extracted from verified Cognito JWT claims inside `middleware/auth.go`.
2. **Context Propagation:** The verified `org_id` is injected into the Go `context.Context`.
3. **Database Scoping:** Every SQL statement in `dl.go` binds `org_id = ?` explicitly:
   - `SELECT * FROM leads WHERE id = ? AND org_id = ?`
   - `UPDATE leads SET ... WHERE id = ? AND org_id = ?`
   - `DELETE FROM leads WHERE id = ? AND org_id = ?`
4. **Query Parameter Spoofing Resistance:** Client requests containing `?org_id=X` are ignored. The backend derives tenancy strictly from the cryptographically verified JWT token.

---

## 23. Audit Trail & Governance Logging

Every state-altering lead interaction produces an immutable structured audit log entry in the `audit_logs` table via `internal/audit`:

- **Manual Creation:** Records `Action: CREATE`, `ResourceType: LEAD`, `ActorType: USER`.
- **AI Ingestion & Parsing:** Records `Action: UPDATE`, `ResourceType: LEAD_INTERACTION`, `ActorType: AI_AGENT` (ActorName: `AI Agent: SalesAgent`).
- **Owner Assignment:** Records `OWNER_CHANGED` with previous and new User ID.
- **Conversion to Customer:** Records `Action: CONVERT`, linking `lead_id` and newly created `customer_id`.
- **Conversion to RFQ:** Records `Action: CONVERT`, associating `lead_id` with `rfq_id`.

---

## 24. Search, Filtering, Sorting & Pagination

### Search Capabilities:
- **Scope:** Free-text case-insensitive substring search matching against `company_name`, `contact_name`, or `email`.
- **Backend Query:** `AND (l.company_name LIKE ? OR l.contact_name LIKE ? OR l.email LIKE ?)` with `%search%` wildcards.

### Filtering Options:
- **Status Tabs:** Quick filter by status (`All`, `NEW`, `IN_PROGRESS`, `QUALIFIED`, `CONVERTED`, `REJECTED`).
- **Source Filter:** Dropdown selection filtering by origin (`MANUAL`, `EMAIL`, `IMPORT`, `WEBSITE`, `OUTREACH`).

### Sorting & Pagination:
- **Default Order:** `ORDER BY l.created_at DESC` (Latest inquiries appear first).
- **Pagination Controls:** Standard limit/offset pagination with customizable page size (default: 50 records per page).

---

## 25. Business User Journeys

### Journey 1: Inbound Email Shipper Inquiry $\rightarrow$ AI Clarification $\rightarrow$ Customer Conversion
1. A prospective shipper sends an email to `quotes@logisticshq.in`: *"Need rates for 2 containers from Shanghai to Rotterdam next month."*
2. The system ingests the email, creates Lead #1109 (Status: `NEW`), and creates Interaction #1235.
3. The AI Sales Agent detects intent `RFQ_REQUEST_INCOMPLETE` (missing Incoterms and Cargo Weight), stages an email draft, and creates Approval Request #252.
4. The sales dispatcher opens `/dashboard/leads`, sees the inquiry, reviews the AI draft, clicks **"Approve Draft"**, and the clarification email is sent.
5. Upon shipper reply confirming details, the dispatcher clicks **"Convert to Customer"**; MariaDB creates Customer Account `CUST-2026-00012` and links the lead.

### Journey 2: Bulk CSV Import $\rightarrow$ Automated ICP Scoring $\rightarrow$ RFQ Initiation
1. The marketing team uploads a CSV of 50 shippers from a chemical export expo via **"Import CSV"**.
2. Go creates 50 leads in MariaDB and publishes `lead.created` events.
3. Background `leadWorker` enriches company data and calls the Python sidecar to compute ICP scores.
4. Leads with scores $\ge 80$ are marked `QUALIFIED`; leads with scores $< 50$ are auto-flagged `REJECTED`.
5. The sales rep filters the table by `QUALIFIED`, opens the top lead, clicks **"Convert to RFQ"**, confirms pre-populated ports, and launches the pricing tender.

---

## 26. Data Flow Diagrams

### Diagram 1: Inbound Lead Creation & AI Enrichment Flow

```
Shipper Email / Web Form / CSV
              │
              ▼
    [ POST /api/v1/leads ]
              │
      (Go Backend API)
              │
     ┌────────┴────────┐
     ▼                 ▼
[ MariaDB ]      [ EventBus ]
(leads table)    (lead.created)
                       │
                       ▼
               [ leadWorker ]
                       │
           ┌───────────┴───────────┐
           ▼                       ▼
   [ Trade Intel ]          [ Python AI ]
   (EnrichCompany)          (ScoreLead:8090)
           │                       │
           └───────────┬───────────┘
                       │
                       ▼
                 [ MariaDB ]
           (ai_score, report)
```

### Diagram 2: Lead to Customer / RFQ Dual Conversion Flow

```
                [ Lead Record (Status: QUALIFIED) ]
                                 │
                 ┌───────────────┴───────────────┐
                 │                               │
                 ▼                               ▼
       (Convert to Customer)              (Convert to RFQ)
                 │                               │
                 ▼                               ▼
   [ LeadConversionModal ]                [ RFQBuilder ]
                 │                               │
                 ▼                               ▼
     [ POST /customers/convert ]         [ POST /rfqs ]
                 │                               │
                 ▼                               ▼
     [ customer_lead_links ]             [ rfqs.lead_id ]
                 │                               │
                 └───────────────┬───────────────┘
                                 │
                                 ▼
                     [ UPDATE leads SET ]
                    (status = 'CONVERTED')
```

---

## 27. Source-of-Truth Matrix

| Data Entity / Field | Authoritative Source of Truth | Secondary / Derived Consumers | Persistence Mechanism |
|:---|:---|:---|:---|
| **Lead Master Attributes** | MariaDB `leads` table | Leads Table UI, Timeline, Dashboard KPIs | Persistent SQL table |
| **Shipper Contact Info** | MariaDB `leads` / `contacts` | RFQ Builder, Customer Conversion Modal | Persistent SQL table |
| **Email Threads & Logs** | MariaDB `lead_interactions` | Emails Tab in Lead Detail Drawer | Persistent SQL table |
| **AI ICP Score (0–100)** | Python AI Sidecar (`POST /leads/score-lead`)| Leads Table Badge, Overview Tab | Persisted in `leads.ai_score` |
| **AI Research Brief** | Python AI Sidecar LLM Synthesis | AI Executive Card in Lead Drawer | Persisted in `leads.ai_research_report`|
| **AI Clarification Draft** | Python AI Sidecar (`app/agents/`) | Pending Approvals Queue, Draft Card | Persisted in `lead_email_drafts` |
| **Conversion Traceability**| MariaDB `customer_lead_links` | Customer Detail 360 view, Sales Reporting | Persistent SQL table |

---

## 28. Loading, Error, and Empty State Lifecycle

1. **Loading State:**
   - On page mount or pagination change, table displays structured skeleton placeholders. Stat cards display `···` while asynchronous count endpoints resolve.
2. **Empty State:**
   - If an organization has zero leads (e.g. newly onboarded forwarder), the table renders `<ModuleHeroEmptyState />` featuring logistics imagery and prompt launcher: *"Start building your freight customer base by capturing shipper inquiries, importing contacts, or converting inbound quotes into active commercial pipelines."*
3. **Error Handling:**
   - Network or server errors render non-blocking toast notifications. If an import fails, line-by-line CSV errors (e.g. invalid email format) are returned in a modal summary dialog without losing valid rows.

---

## 29. UI/UX Observations & Ergonomics Review

### Strengths:
- **Clean SaaS Aesthetic:** Matches LogisticsHQ navy/white enterprise theme with clear typography and spacing.
- **Drawer Architecture:** The React Portal slide-out drawer (`LeadDetailPanel.jsx`) allows dispatchers to inspect full email threads and RFQ readiness checklists without losing their place in the lead list.
- **Conversion Safety:** Duplicate customer checking prevents fragmented customer records.

### Areas for Improvement (Backlog):
- **MUST FIX (None):** No blocking functional bugs detected.
- **SHOULD IMPROVE:** When a lead is converted to an RFQ, the drawer should display a direct clickable hyperlink to the newly created RFQ ID (`/dashboard/rfqs?id=XXX`).
- **OPTIONAL:** Add kanban board view option alongside the traditional table view for visual pipeline drag-and-drop.

---

## 30. Responsive & Browser Zoom Behavior

- **Desktop Viewports (1024px, 1280px, 1366px, 1440px):** Full multi-column layout is fully functional. The slide-out detail drawer maintains a standard 640px width on screens $\ge 1280\text{px}$ and adjusts to 85% viewport width on narrower desktop displays.
- **Browser Zoom (80% to 125%):** Tested via Chrome DevTools Emulation. Zero text clipping, horizontal overflow, or table alignment collapse observed.

---

## 31. Security Architecture & Boundary Enforcement

1. **Fail-Closed Authentication:** Any request to `/api/v1/leads*` without a valid Bearer JWT immediately returns HTTP 401 Unauthorized.
2. **Strict Multi-Tenancy:** Cross-tenant access is prevented by server-side query scoping (`WHERE org_id = ?`). Client query parameter spoofing (`?org_id=X`) is discarded.
3. **Autonomous Email Guardrail:** The Go backend prohibits the AI sidecar from autonomously sending clarification emails. All AI-generated drafts are staged as `AWAITING_APPROVAL`, enforcing human sign-off before transmission.
4. **Input Sanitization:** Multi-part CSV file uploads are parsed via Go `encoding/csv` with row limits and memory capping to prevent buffer overrun or denial-of-service vulnerabilities.

---

## 32. Business & Technical Glossary

- **Lead:** A prospective shipper or commercial freight customer prior to formal account establishment or shipment booking.
- **ICP Score (Ideal Customer Profile):** An AI-generated metric (0–100) evaluating how closely a shipper's cargo volume and trade lanes align with the freight forwarder's strengths.
- **Lead Interaction:** A recorded touchpoint (inbound email, phone call note, outbound draft) tied to a specific lead conversation thread.
- **Lead Email Draft:** A staged outbound clarification email generated by AI that requires human dispatcher approval before sending.
- **Lead Conversion:** The operational transition of a qualified lead into a permanent Customer Account or an active freight RFQ.
- **RFQ (Request for Quotation):** A commercial request specifying container count, origin/destination ports, cargo weight, and ready dates for rate quoting.

---

## 33. One-Page Summary: “How Leads Work in LogisticsHQ”

```
┌────────────────────────────────────────────────────────────────────────────┐
│                  HOW LEADS WORK IN LOGISTICSHQ                             │
├────────────────────────────────────────────────────────────────────────────┤
│ 1. INGESTION                                                               │
│    Inquiries arrive via connected mailboxes, CSV uploads, web forms,       │
│    or manual entry and are saved to the Leads table (Status: NEW).         │
│                                                                            │
│ 2. TRADE INTELLIGENCE & AI SCORING                                         │
│    A background worker queries maritime databases for shipping volume      │
│    (TEUs) and top suppliers. A Python AI agent calculates an ICP fit score │
│    (0–100) and produces an executive research report.                      │
│                                                                            │
│ 3. CONVERSATION & CLARIFICATION                                            │
│    Incoming shipper emails are parsed for origin/destination ports, cargo  │
│    weights, and ready dates. Missing details trigger an AI draft reply     │
│    staged for human approval (never sent autonomously).                    │
│                                                                            │
│ 4. QUALIFICATION                                                           │
│    Sales dispatchers review the prospect in a slide-out drawer, track all  │
│    threads in a unified timeline, and update status to QUALIFIED.          │
│                                                                            │
│ 5. CONVERSION                                                              │
│    - Convert to Customer: Runs duplicate checks and creates an active      │
│      customer account in the master customer directory.                    │
│    - Convert to RFQ: Pre-populates shipping items and launches the pricing │
│      tender for rate quotation.                                            │
│                                                                            │
│ 6. PERMANENT TRACEABILITY                                                  │
│    The lead record is preserved as CONVERTED with full audit history,      │
│    powering top-of-funnel sales conversion analytics.                      │
└────────────────────────────────────────────────────────────────────────────┘
```

---

## 34. Technical Traceability Matrix

| Lead Capability | Frontend Component | API Endpoint | Go Handler / Service | Python AI Component | MariaDB Tables | Action System / Event | RBAC Permission |
|:---|:---|:---|:---|:---|:---|:---|:---|
| **Lead Listing** | `LeadsPage.jsx` | `GET /api/v1/leads` | `transport.go` $\rightarrow$ `dl.go:List` | N/A | `leads` | N/A | `LEADS:READ` |
| **Manual Creation**| `AddLeadModal.jsx` | `POST /api/v1/leads` | `bl.go:CreateLead` | N/A | `leads`, `lead_activities`| `lead.created` | `LEADS:CREATE` |
| **CSV Bulk Import** | `ImportLeadsModal.jsx` | `POST /api/v1/leads/import`| `bl.go:BulkCreate` | N/A | `leads` | `lead.created` | `LEADS:CREATE` |
| **Lead Detail View**| `LeadDetailPanel.jsx`| `GET /api/v1/leads/{id}` | `dl.go:GetByID` | N/A | `leads`, `rfqs` | N/A | `LEADS:READ` |
| **Status Update** | `LeadDetailPanel.jsx`| `PUT /api/v1/leads/{id}` | `bl.go:UpdateLead` | N/A | `leads`, `lead_activities`| N/A | `LEADS:UPDATE` |
| **AI Lead Scoring** | `LeadDetailPanel.jsx`| Background Job | `leadWorker.go` | `POST /leads/score-lead` | `leads.ai_score` | `lead.enriched` | `LEADS:READ` |
| **Email Parsing** | `LeadDetailPanel.jsx`| `POST /api/v1/emails/inbound`| `email_handler.go` | `POST /leads/classify-email`| `lead_interactions` | `email.received`| Public / Webhook |
| **AI Draft Approval**|`LeadDetailPanel.jsx`| `POST .../approve-draft` | `email_handler.go` | N/A | `lead_email_drafts` | Human Approval Gate | `APPROVALS:UPDATE` |
| **Customer Convert**| `LeadConversionModal` | `POST /customers/convert-lead`| `customers/bl.go` | N/A | `customers`, `customer_lead_links`| `sales.convert_lead`| `LEADS:UPDATE` |
| **RFQ Convert** | `RFQBuilder.jsx` | `POST /api/v1/rfqs` | `rfq/bl.go:CreateRFQ` | N/A | `rfqs`, `leads` | `rfq.created` | `RFQS:CREATE` |

---

## 35. Known Gaps & Deficiencies

1. **Implementation Gaps:**
   - When a lead is converted to an RFQ, the lead record stores `linked_rfq_id`, but the detail panel does not display a direct deep link to `/dashboard/rfqs?id={linked_rfq_id}`.
2. **Configuration Gaps:**
   - Inbound email parsing depends on mailbox synchronization (`/api/v1/organizations/mailboxes`). If no Gmail or SMTP account is connected in Settings, leads can only be created via manual entry, CSV import, or simulated webhook test payloads.
3. **UI/UX Gaps:**
   - Bulk status update and bulk owner assignment UI controls exist in the table action bar, but bulk conversion to customer accounts is not currently supported (conversions require individual duplicate verification).

---

## 36. Verification Status & Sign-Off

### Audit Verification Checklist:
- [x] **Frontend Architecture Traced:** Verified `LeadsPage.jsx`, `LeadDetailPanel.jsx`, `AddLeadModal.jsx`, `ImportLeadsModal.jsx`, `LeadConversionModal.jsx`, and `RFQBuilder.jsx`.
- [x] **Go Backend Traced:** Verified `transport.go`, `endpoints.go`, `bl.go`, `dl.go`, `email_handler.go`, and `routes.go`.
- [x] **Database Schema Audited:** Verified all 8 MariaDB tables, primary keys, foreign keys, and multi-tenant `org_id` indexes.
- [x] **APIs Verified:** Documented all 17 active endpoints matching runtime code.
- [x] **Lead Lifecycle Traced:** Documented real state machine (`NEW`, `QUALIFIED`, `IN_PROGRESS`, `REJECTED`, `CONVERTED`).
- [x] **Conversion Logic Grounded:** Verified both Customer and RFQ conversion pathways in database and business logic.
- [x] **AI & Safety Gates Verified:** Documented `LeadScoringAgent`, `EmailClassifierAgent`, and human approval requirement for AI drafts.
- [x] **RBAC & Isolation Confirmed:** Verified role permissions and fail-closed tenant scoping.
- [x] **No Business Data Modified:** Audit conducted strictly via code and read-only inspection.

### Final Verification Result:
# **PASS — LEADS WORKFLOW DOCUMENTED**
*The LogisticsHQ Leads module is fully documented across business operations, data structures, and technical architecture, and is prepared for Task 3.2 Deep Functional Review and Remediation.*
