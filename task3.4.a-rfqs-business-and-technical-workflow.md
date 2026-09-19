# Task 3.4.A — RFQ Business Workflow, End-to-End Process, Database Mapping, API Traceability, AI Integration, and Complete Technical Documentation

> **Status:** PASS — RFQ WORKFLOW DOCUMENTED  
> **Repository:** LogisticsHQ (`freel-project`)  
> **Modules Evaluated:** React Frontend (`frontend/src/pages/dashboard/RFQ`), Go Backend (`backend/internal/rfq`, `backend/internal/rfq/pricing_workflow`, `backend/internal/pricing`, `backend/internal/rates`), MariaDB Schema (`rfqs`, `rfq_items`, `rfq_quotes`, `rfq_pricing_optimizations`, `rfq_pricing_versions`, `ai_rfq_requirements_extractions`, `quotations`, `bookings`, `shipments`), Python AI Sidecar (`ai_sidecar/app/agents/rfq_parser_agent.py`, `ai_sidecar/app/rfq_pricing`).  
> **Date:** September 13, 2026  
> **Evaluator:** Deep Functional Audit Agent (Antigravity)  

---

## 1. Executive Summary

The **Request for Quotation (RFQ)** module is the commercial and operational switchboard of LogisticsHQ. In global freight forwarding, every revenue-generating cargo movement begins as an RFQ. It captures the shipper's demand (origin, destination, cargo specifications, equipment, target departure dates, and trade terms) and orchestrates the complex multi-step procurement and pricing workflow required to return an authoritative, profitable **Quotation** to the customer.

In LogisticsHQ, the RFQ module is implemented as a **Full-Lifecycle Commercial Cockpit**. Rather than a passive intake form, it actively drives:
1. **Unstructured Inquiry Parsing:** Inbound customer emails and free-text freight inquiries are ingested and converted into structured parameters by the Python AI Sidecar (`RFQParserAgent`) via Go boundary validation.
2. **Operational Requirements & Readiness Engine:** Deterministic rule evaluation (`requirements_engine.go`) assesses missing mandatory fields, stage-aware compliance obligations (dangerous goods, reefer temperature controls), and required documents (Commercial Invoice, Packing List, Certificate of Origin).
3. **Multi-Carrier Quote Comparison:** Procurement managers record, compare, benchmark, and approve carrier tariffs (`rfq_quotes`), contrasting ocean freight buy prices against origin and destination terminal handling charges.
4. **Predictive Pricing & Margin Optimization:** AI pricing workflows (`backend/internal/rfq/pricing_workflow`) analyze historical lane rates, market indices, and customer tiers to predict landed freight costs, recommend optimal sell margins, and gate below-margin proposals through the enterprise **Approval System**.
5. **Downstream Operational Handoff:** Upon commercial victory, the RFQ provides structured handoffs directly into the **Bookings** workspace and **Shipment Execution** tracking engine.

All operations strictly enforce multi-tenant isolation (`org_id`), Go-governed RBAC policies, universal audit logging (`audit_logs`), and Event Mesh asynchronous event distribution.

---

## 2. RFQ in Plain English

### What does "RFQ" mean?
**RFQ** stands for **Request for Quotation**. In everyday terms, it is a formal freight inquiry from a customer asking:  
*"How much will it cost, and how long will it take, to ship my goods from Port A to Port B under specific commercial conditions?"*

### Why does a customer create an RFQ?
When a manufacturing company, exporter, or retail importer needs to move freight—whether 20 metric tons of machinery from Mumbai to Hamburg or 5 pallets of electronics from Bangalore to Frankfurt—they do not simply buy a shipping ticket online like an airline passenger. Freight rates fluctuate weekly based on bunker fuel prices, equipment availability, carrier capacity, and seasonal surcharges. The customer submits an RFQ detailing their cargo to obtain firm, binding pricing before booking space on a vessel or aircraft.

### What information does an RFQ contain?
- **Who:** The customer (shipper, consignee, or trader) and primary operational contact.
- **Where:** Origin (Port of Loading / POL) and Destination (Port of Discharge / POD).
- **What:** Cargo description, container equipment type (e.g., 20ft Standard, 40ft High Cube, Reefer), piece count, gross weight (kg), and volume (CBM).
- **When:** Target departure or cargo-ready date.
- **Commercial Terms (Incoterms):** Clarifies who pays for freight, insurance, and customs (e.g., FOB, CIF, EXW, DDP).
- **Special Handling:** Dangerous Goods (DG/IMDG class), temperature settings, or customs brokerage requirements.

### Who works with an RFQ?
1. **Sales Executives:** Qualify the customer inquiry and log initial requirements.
2. **Pricing & Procurement Specialists:** Query carrier tariffs (Maersk, MSC, Hapag-Lloyd), evaluate buy prices, add commercial markups, and structure customer quotes.
3. **Operations Coordinators:** Verify cargo feasibility, customs documentation, and booking handoff readiness.
4. **Commercial Directors / Approvers:** Review low-margin or high-risk quotes before they reach the customer.

### What happens after an RFQ is created?
Once an RFQ is opened, the system checks whether the information is complete. Carriers are matched to the lane, buy rates are secured, and pricing margins are computed. When approved, a formal **Quotation** is generated and delivered to the customer. If the customer accepts, the RFQ is marked **WON** and immediately converts into an operational **Booking** and live **Shipment**.

---

## 3. Business Purpose

The RFQ module fulfills six critical commercial objectives:

- **Demand Intake Centralization:** Unifies inquiries originating from customer portals, sales representatives, and inbound customer emails into a single, standardized pipeline.
- **Revenue & Margin Protection:** Enforces minimum gross margin floors (e.g., 8% hard floor, 15% target margin) and requires managerial approvals before discount pricing can be released.
- **Accelerated Turnaround Time (TAT):** Reduces quote turnaround time from days to minutes through AI-assisted parsing, automated tariff lookup, and dynamic pricing previews.
- **Auditability & Governance:** Maintains an immutable chronological audit trail (`audit_logs`) tracking who changed stage, approved carrier rates, or adjusted margins.
- **Pre-Booking Risk Prevention:** Identifies missing customs paperwork (e.g., MSDS certificates, packing lists) before cargo dispatch, preventing expensive port demurrage and customs seizures.
- **Commercial-to-Operations Continuity:** Eliminates double-entry by transmitting validated cargo weights, container counts, and agreed rates seamlessly from quotation into carrier booking confirmations.

---

## 4. RFQ Lifecycle

The RFQ lifecycle is governed by canonical stage states defined in `backend/internal/rfq/spec/const.go` and enforced via `UpdateStage` in `bl.go`.

```
                    ┌─────────────────┐
                    │   RFQ_CREATED   │ ◄── New inquiry created (Manual/AI)
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ PRICING_ASSIGNED│ ◄── Assigned to pricing specialist / tariff matching
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │ QUOTE_GENERATED │ ◄── Carrier tariffs calculated, margin preview set
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   QUOTE_SENT    │ ◄── Formal quotation dispatched to customer
                    └────────┬────────┘
                             │
            ┌────────────────┴────────────────┐
            ▼                                 ▼
   ┌─────────────────┐               ┌─────────────────┐
   │   NEGOTIATION   │               │      LOST       │ ◄── Customer declined or expired
   └────────┬────────┘               └─────────────────┘
            │
            ▼
   ┌─────────────────┐
   │       WON       │ ◄── Customer accepted quotation
   └────────┬────────┘
            │
            ▼
   ┌─────────────────┐
   │ SHIPMENT_CREATED│ ◄── Booking confirmed & operational shipment generated
   └─────────────────┘
```

### State Transition Matrix

| Current Stage | Allowed Next Stage | Business Meaning | Who Can Change It | Side Effect / Trigger |
| :--- | :--- | :--- | :--- | :--- |
| **`RFQ_CREATED`** | `PRICING_ASSIGNED` | Inbound requirements recorded and validated; ready for carrier rate procurement. | Sales Agent / Pricing Lead | Emits `EventRFQAssigned`; assigns `pricing_assignee_id`. |
| **`PRICING_ASSIGNED`** | `QUOTE_GENERATED` | Carrier buy rates secured; commercial margin applied; draft quote formed. | Pricing Specialist / AI Workflow | Emits `EventQuoteGenerated`; persists `quotations` draft. |
| **`QUOTE_GENERATED`** | `QUOTE_SENT` | Quotation verified, approved if required, and transmitted to customer. | Pricing Lead / Commercial Mgr | Emits `EventQuoteSent`; dispatches customer notification; sets quote `valid_until`. |
| **`QUOTE_SENT`** | `NEGOTIATION` | Customer requested revisions on price, free days, or transit time. | Commercial Sales Rep | Emits `EventRFQUpdated`; logs negotiation activity. |
| **`QUOTE_SENT`** / `NEGOTIATION` | `WON` | Customer accepted commercial quotation. | Sales Rep / Customer Portal | Emits `EventRFQWon`; marks quotation `ACCEPTED`; unlocks Booking Handoff. |
| **`QUOTE_SENT`** / `NEGOTIATION` | `LOST` | Customer selected rival forwarder, canceled shipment, or quote expired. | Sales Rep / Auto-Expiry | Emits `EventRFQLost`; records loss reason in audit trail. |
| **`WON`** | `SHIPMENT_CREATED` | Booking confirmed with carrier and handed off to freight operations. | Operations Coordinator | Emits `EventRFQUpdated`; links `shipments.rfq_id` to operational tracking. |

---

## 5. RFQ Creation Paths

LogisticsHQ supports three distinct, verified creation paths:

```
Path A: Direct UI Creation
User -> RFQBuilder.jsx -> POST /api/v1/rfqs -> Go Validation -> MariaDB (rfqs, rfq_items) -> Event Mesh -> UI

Path B: AI Inbound Text / Email Extraction
Customer Email / Inquiry -> POST /api/v1/rfqs/parse-shipment-request -> Python RFQParserAgent -> Structured JSON -> Review in UI -> POST /api/v1/rfqs

Path C: Customer 360 / CRM Inception
CustomerDetailsPage.jsx -> "+ Create RFQ" -> Pre-populates Customer ID, Address, Contact -> POST /api/v1/rfqs
```

### Detailed Execution Trace (Path A - Direct Creation):
1. **Frontend Form Submission:** User enters origin, destination, incoterms, target date, equipment, and cargo line items in `RFQBuilder.jsx`.
2. **Transport & Validation:** React client dispatches `POST /api/v1/rfqs` with Bearer token.
3. **Go Boundary & Security:** Transport decoder (`transport.go`) extracts `org_id` from JWT claims, rejecting unauthenticated requests. Business logic (`bl.go:CreateRFQ`) validates mandatory fields (`customer_id`, `origin`, `destination`).
4. **Data Persistence:** Data layer (`dl.go:CreateRFQ`) executes a transactional insert into `rfqs`, auto-generating `rfq_number` (`RFQ-YYYY-#####`), and bulk-inserts line items into `rfq_items`.
5. **Audit & Event Mesh:** Emits `ActionCreate` to `audit_logs` and publishes `EventRFQCreated` onto the in-process Event Mesh.
6. **UI Response:** Returns HTTP 201 with created RFQ payload; UI redirects to `/dashboard/rfqs/:id`.

---

## 6. Customer → RFQ Relationship

### Business View
Every RFQ must belong to an identifiable commercial counterparty. When a freight forwarder quotes rates, credit terms, payment liability, and customs declarations apply directly to the legal customer account.

### Technical Implementation
- **Database Link:** `rfqs.customer_id` stores the foreign key referencing `customers.id`.
- **Tenant Verification:** Both `rfqs.org_id` and `customers.org_id` must match.
- **Hydration at Read Time:**
  ```sql
  SELECT r.id, r.org_id, r.rfq_number, r.customer_id,
         COALESCE(c.name, '') AS customer_name,
         c.contact_email AS customer_email,
         c.contact_phone AS customer_phone,
         c.contact_name AS customer_contact_name,
         r.stage, r.origin, r.destination, r.incoterms, r.target_date
  FROM rfqs r
  LEFT JOIN customers c ON r.customer_id = c.id AND r.org_id = c.org_id
  WHERE r.org_id = ? AND r.id = ?
  ```
- **UI Presentation:** The top header card of `RFQDetailPage.jsx` displays the Customer legal name, contact email, telephone, and commercial tier directly from the joined relationship.

---

## 7. Inbound Email / Text → RFQ Workflow

LogisticsHQ implements an automated AI pipeline for extracting freight parameters from unstructured emails, WhatsApp messages, or quotation requests.

```
Inbound Email / Text
       │
       ▼
POST /api/v1/rfqs/parse-shipment-request (Go Backend)
       │
       ▼
ai.NewSidecarClient().ParseShipmentRequest()
       │
       ▼
POST /rfq/parse-shipment-request (Python AI Sidecar:8090)
       │
       ▼
RFQParserAgent.parse_request()
(Executes LLM with prompt: rfq.extract_shipment_request)
       │
       ▼
Extracts JSON: { origin, destination, incoterms, weight, volume }
+ confidence_score (0-100) + missing_fields
       │
       ▼
Go Business Logic Validation (bl.go)
       │
       ▼
Pre-populates RFQ Creation Modal for Human Operator Verification
```

- **Python/Go Boundary:** Python performs reasoning and parameter extraction only. It never inserts directly into the `rfqs` table.
- **Human-in-the-Loop Safety:** The extracted parameters are presented in `RFQBuilder.jsx` with confidence badges. The operator reviews, corrects, and confirms the details before creating the business record.

---

## 8. RFQ Data Fields

| Field Name | Technical Column | Data Type | Business Meaning & Usage |
| :--- | :--- | :--- | :--- |
| **RFQ Number** | `rfq_number` | `VARCHAR(100)` | Unique commercial identifier (e.g. `RFQ-2026-1001`). Generated sequentially. |
| **Organization ID** | `org_id` | `BIGINT(20)` | Multi-tenant partition key ensuring forwarder data isolation. |
| **Customer ID** | `customer_id` | `BIGINT(20)` | Foreign key to `customers.id`. Identifies billing counterparty. |
| **Stage** | `stage` | `VARCHAR(50)` | Commercial progression (`RFQ_CREATED`, `PRICING_ASSIGNED`, `QUOTE_SENT`, `WON`, etc.). |
| **Origin** | `origin` | `VARCHAR(255)` | Port of Loading (POL) or pickup facility (e.g. `INNSA (Nhava Sheva)`). |
| **Destination** | `destination` | `VARCHAR(255)` | Port of Discharge (POD) or delivery site (e.g. `DEHAM (Hamburg)`). |
| **Incoterms** | `incoterms` | `VARCHAR(50)` | Commercial trade terms (`FOB`, `CIF`, `EXW`, `CIP`, `DDP`). |
| **Target Date** | `target_date` | `DATETIME` | Expected cargo-ready date or vessel sailing deadline. |
| **Cargo Details** | `cargo_details` | `TEXT` | Free-form narrative of commodity (e.g. *"Automotive spare parts in wooden crates"*). |
| **Container Type** | `container_type` | `VARCHAR(50)` | Equipment category (`20FT_STANDARD`, `40FT_HIGH_CUBE`, `REEFER`, `LCL`). |
| **Quantity** | `quantity` | `INT(11)` | Number of containers or packages. |
| **Weight** | `weight` | `DECIMAL(10,2)` | Gross cargo weight in kilograms (kg). |
| **Volume** | `volume` | `DECIMAL(10,2)` | Cubic volume in cubic meters (CBM). |
| **Lead ID** | `lead_id` | `BIGINT(20)` | Optional linkage to `leads.id` if originating from CRM outreach. |
| **Sales Assignee** | `sales_assignee_id`| `BIGINT(20)` | User ID of commercial sales account executive. |
| **Pricing Assignee**| `pricing_assignee_id`| `BIGINT(20)`| User ID of procurement/pricing specialist evaluating carrier rates. |
| **Health Score** | `health_score` | `INT(11)` | Operational readiness metric (0-100) computed by requirements engine. |

---

## 9. RFQ → Quotation Workflow

The transition from an RFQ to a customer quotation is the central commercial transaction of freight forwarding:

```
1. Procurement (RFQ Quotes)
   Carrier Tariffs matched -> Recorded in rfq_quotes (buy_price, ocean_freight, surcharges)
   Pricing specialist marks one quote as "RECOMMENDED" and clicks "APPROVE".

2. AI Quotation Workflow (Phase 3 Task 3.4)
   User opens tab "⚡ AI Quotation Workflow" (RFQDetailPage.jsx)
   -> GET /api/v1/rfqs/:id/pricing-workflow/overview
   -> POST /api/v1/rfqs/:id/pricing-workflow/pricing-preview
      (Computes base cost: $2,480.00, target margin: 15%, recommended price: $2,917.65)
   -> If margin < 8% (Policy Floor): Flagged for Manager Approval (approval_status: REQUIRED)

3. Draft Quotation Generation
   User clicks "Generate Quotation Draft"
   -> POST /api/v1/rfqs/:id/pricing-workflow/drafts
   -> Inserts record into quotations table (status: DRAFT, quotation_number: QT-2026-XXXX)
   -> Advances RFQ stage to QUOTE_GENERATED.

4. Client Dispatch
   Pricing Manager reviews and transmits quotation
   -> RFQ stage advances to QUOTE_SENT
   -> Email / in-app notification delivered to customer.
```

### Distinction of Rates and Prices:
- **Authoritative Carrier Cost:** Persisted in `rfq_quotes.buy_price` and `quotations.total_cost`. Sourced directly from carrier contracts, tariffs, or broker quotes.
- **AI Pricing Recommendation:** Generated statelessly by `pricing_workflow/service.go` and Python AI Sidecar (`rfq_pricing/agent.py`). Evaluates lane seasonality, fuel volatility, and customer sensitivity.
- **Authoritative Customer Price:** Stored in `quotations.total_amount`. Represents the final binding legal charge approved by human commercial management.

---

## 10. RFQ → Booking Workflow

Once the customer accepts a quotation, the RFQ transitions to commercial victory (`WON`) and initiates operational booking execution:

1. **Eligibility Check:** The frontend invokes `GET /api/v1/rfqs/:id/bookings`. Go backend checks:
   - RFQ Stage is `WON` or quotation is `ACCEPTED`.
   - All blocking documentation requirements (e.g. commercial invoice, DG form) are `SATISFIED`.
2. **Booking Generation:** User clicks "Create Booking" on the `Booking` subtab in `RFQDetailPage.jsx`.
   - Sends `POST /api/v1/rfqs/:id/bookings`.
   - Creates a booking record in `bookings` table with `status = 'CONFIRMED'`.
   - Links `bookings.rfq_id = rfqs.id`.
3. **Carrier Dispatch Integration:** Triggers `carrier_booking_engine.go` to transmit shipping instructions to the selected shipping line (e.g. Maersk, MSC, Ocean Network Express).
4. **Shipment Execution Handoff:** Once booking confirmation (BC) is received, the shipment operations engine initializes a live tracking entity in `shipments` table (`status = 'BOOKED'`).

---

## 11. Pricing / Rate Intelligence

The pricing engine operates on three distinct analytical tiers:

1. **Carrier Tariff Matching (`backend/internal/rates`):**
   - Normalizes origin and destination UN/LOCODEs (`port_normalizer.go`).
   - Evaluates active contracted rates and spot market sheets in `rates` and `spot_rates`.
   - Returns ranked carrier options with reliability scores and historical success rates.
2. **Predictive Pricing Engine (`ai_sidecar/app/rfq_pricing`):**
   - Assesses spot price volatility on the trade corridor.
   - Evaluates cargo density, demurrage risks, and bunker fuel adjustment factors (BAF).
   - Generates three candidate commercial strategies:
     - **Conservative (Market Parity):** Low margin (8-10%), high conversion probability.
     - **Balanced (Recommended):** Standard target margin (12-16%), optimal risk-return.
     - **Aggressive (Premium Service):** High margin (18-22%), for priority space allocation.
3. **Margin Risk Governance:**
   - Minimum Margin Policy: Any quote yielding `< 8.00%` gross margin triggers `margin_risk_level = 'HIGH'` and blocks dispatch until a Commercial Director approves.

---

## 12. Carrier / Rate Data

| Rate Category | Technical Source | Persistence Table | Usage in RFQ |
| :--- | :--- | :--- | :--- |
| **Contracted Tariff** | Rate contracts negotiated with ocean/air carriers. | `rates`, `rate_contracts` | Default baseline for standard lane inquiries. |
| **Spot Market Rate** | Dynamic spot rates updated weekly or via carrier APIs. | `spot_rates` | Applied for ad-hoc or peak season container freight. |
| **Carrier RFQ Quote** | Quotation submitted by carrier agent for specific shipment. | `rfq_quotes` | Authoritative operational buy cost for an active RFQ. |
| **AI Predicted Cost** | Landed cost projection synthesized by Python AI Sidecar. | `rfq_pricing_optimizations` | Baseline for commercial margin recommendation. |

---

## 13. Database Table Mapping

```
                               ┌───────────────┐
                               │ organizations │
                               └───────┬───────┘
                                       │ 1:N
                                       ▼
    ┌──────────────────────────────────┬─────────────────────────────────┐
    │                                  │                                 │
    ▼                                  ▼                                 ▼
┌───────────┐                  ┌───────────────┐                 ┌───────────────┐
│ customers │ ◄── 1:N ──────── │     rfqs      │ ──────── 1:N ──►│   rfq_items   │
└───────────┘                  └───────┬───────┘                 └───────────────┘
                                       │
        ┌──────────────────────────────┼──────────────────────────────┐
        ▼                              ▼                              ▼
┌───────────────┐              ┌───────────────┐              ┌───────────────┐
│  rfq_quotes   │              │ rfq_documents │              │  quotations   │
└───────────────┘              └───────────────┘              └───────┬───────┘
        │                                                             │
        ▼                                                             ▼
┌──────────────────────────────┐                              ┌───────────────┐
│ rfq_pricing_optimizations    │                              │   bookings    │
└──────────────────────────────┘                              └───────┬───────┘
                                                                      │
                                                                      ▼
                                                              ┌───────────────┐
                                                              │   shipments   │
                                                              └───────────────┘
```

### Canonical Table Specifications

| Table | Business Purpose | Key Operation | Important Fields | Related Tables |
| :--- | :--- | :--- | :--- | :--- |
| **`rfqs`** | Core system of record for customer freight inquiries. | Query, insert, stage advancement. | `id`, `org_id`, `rfq_number`, `customer_id`, `stage`, `origin`, `destination`, `incoterms`, `target_date`, `health_score`. | `customers`, `rfq_items`, `rfq_quotes`, `quotations`. |
| **`rfq_items`** | Line-item cargo packing specifications. | Cargo volume and weight breakdown. | `id`, `rfq_id`, `description`, `quantity`, `weight_kg`, `volume_cbm`. | `rfqs`. |
| **`rfq_quotes`** | Procured carrier quotations and tariffs. | Buy price comparison, margin derivation. | `id`, `rfq_id`, `carrier_id`, `carrier_name`, `buy_price`, `sell_price`, `ocean_freight`, `origin_charges`, `status`, `is_recommended`. | `rfqs`, `carriers`. |
| **`rfq_documents`**| Shipping documents attached to the inquiry. | Pre-clearance compliance verification. | `id`, `org_id`, `rfq_id`, `document_type`, `status`, `file_name`, `s3_key`, `verified_at`. | `rfqs`, `documents`. |
| **`rfq_pricing_optimizations`** | AI pricing calculations and margin analysis records. | Pricing preview, margin governance. | `id`, `org_id`, `rfq_id`, `quotation_id`, `base_cost`, `predicted_cost`, `recommended_price`, `recommended_margin_pct`, `margin_risk_level`, `requires_approval`. | `rfqs`, `quotations`, `approvals`. |
| **`rfq_pricing_versions`** | Revision history of pricing calculations. | Historical audit of price changes. | `id`, `optimization_id`, `version_number`, `recommended_price`, `margin_pct`, `created_at`. | `rfq_pricing_optimizations`. |
| **`ai_rfq_requirements_extractions`** | AI analysis of missing cargo requirements. | Operational readiness evaluation. | `id`, `org_id`, `rfq_id`, `status`, `extracted_data`, `missing_fields`, `clarification_needed`, `confidence_score`. | `rfqs`. |
| **`quotations`** | Commercial pricing proposal dispatched to customer. | Proposal generation and acceptance. | `id`, `org_id`, `quotation_number`, `rfq_id`, `customer_id`, `status`, `total_amount`, `total_cost`, `gross_margin_pct`, `valid_until`. | `rfqs`, `customers`, `bookings`. |
| **`bookings`** | Confirmed carrier space reservation. | Operational booking handoff. | `id`, `org_id`, `booking_number`, `rfq_id`, `carrier_name`, `origin_port`, `destination_port`, `status`. | `rfqs`, `quotations`, `shipments`. |

---

## 14. RFQ Relationship Map

```
RFQ (rfqs)
├── Customer (customers) [1:1 via customer_id]
│   └── Primary Contact (contacts)
├── Cargo Items (rfq_items) [1:N via rfq_id]
├── Documents (rfq_documents) [1:N via rfq_id]
│   └── Commercial Invoice, Packing List, Certificate of Origin, MSDS
├── Requirements Engine (requirements_engine.go)
│   ├── Operational Readiness Rules (Blocking vs Conditional)
│   └── AI Extractions (ai_rfq_requirements_extractions)
├── Carrier Quotes (rfq_quotes) [1:N via rfq_id]
│   ├── Maersk, MSC, CMA CGM, Hapag-Lloyd Tariffs
│   └── Recommendation & Approval Engine (quote_engine.go)
├── Pricing Optimization (rfq_pricing_optimizations) [1:1 via rfq_id]
│   ├── Cost Modeling & Predicted BAF
│   └── Revision Versions (rfq_pricing_versions)
├── Quotation (quotations) [1:N via rfq_id]
│   └── Customer Approval / Rejection Lifecycle
├── Operational Handoffs
│   ├── Bookings (bookings) [1:N via rfq_id]
│   └── Shipments (shipments) [1:N via rfq_id]
├── Enterprise Audit Trail (audit_logs) [ModuleRFQs]
└── Event Mesh (eventBus) [EventRFQCreated, EventQuoteSent, EventRFQWon]
```

---

## 15. API Mapping

| Business Operation | Frontend Function | Method | Endpoint | Go Handler / Service | Database Target | Result |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **List RFQs** | `rfqService.listRFQs` | `GET` | `/api/v1/rfqs` | `endpoints.ListRFQsEP` $\rightarrow$ `dl.ListRFQs` | `rfqs` | Returns paginated array of RFQ summaries with joined customer names. |
| **Get RFQ Workspace** | `rfqService.getRFQ` | `GET` | `/api/v1/rfqs/{id}` | `endpoints.GetRFQEP` $\rightarrow$ `dl.GetRFQByID` | `rfqs`, `rfq_items`, `customers` | Full RFQ detail with cargo items and contacts. |
| **Create RFQ** | `rfqService.createRFQ` | `POST` | `/api/v1/rfqs` | `endpoints.CreateRFQEP` $\rightarrow$ `bl.CreateRFQ` | `rfqs`, `rfq_items`, `audit_logs` | Creates RFQ record; emits `EventRFQCreated`. |
| **AI Text Ingestion** | `rfqService.parseShipmentRequest` | `POST` | `/api/v1/rfqs/parse-shipment-request` | `endpoints.ParseShipmentRequestEP` $\rightarrow$ `bl.ParseShipmentRequest` | None (Calls Python AI Sidecar) | Returns structured `{origin, dest, weight, volume, confidence}`. |
| **Update Stage** | `rfqService.updateStage` | `PUT` | `/api/v1/rfqs/{id}/stage` | `endpoints.UpdateStageEP` $\rightarrow$ `bl.UpdateStage` | `rfqs`, `audit_logs` | Advances lifecycle stage; publishes Event Mesh event. |
| **Evaluate Requirements**| `rfqService.getRequirements`| `GET` | `/api/v1/rfqs/{id}/requirements` | `endpoints.GetRequirementsEP` $\rightarrow$ `requirements_engine` | `rfqs`, `rfq_items`, `rfq_documents` | Returns blocking items, readiness grade, and document checklist. |
| **Carrier Tariff Lookup**| `rfqService.getCarrierRates`| `GET` | `/api/v1/rfqs/{id}/carrier-rates` | `endpoints.GetCarrierRatesEP` $\rightarrow$ `rates.CarrierRatesEngine` | `rates`, `spot_rates` | Returns ranked carrier options for trade lane. |
| **Add Carrier Quote** | `rfqService.createQuote` | `POST` | `/api/v1/rfqs/{id}/quotes` | `endpoints.CreateQuoteEP` $\rightarrow$ `quote_engine` | `rfq_quotes`, `audit_logs` | Inserts carrier tariff quote. |
| **Approve Quote** | `rfqService.approveQuote` | `POST` | `/api/v1/rfqs/{id}/quotes/{qid}/approve`| `endpoints.ApproveRFQQuoteEP` | `rfq_quotes`, `rfqs` | Approves buy rate; advances RFQ to `QUOTE_GENERATED`. |
| **Pricing Overview** | `rfqService.getPricingOverview`| `GET` | `/api/v1/rfqs/{id}/pricing-workflow/overview`| `pricing_workflow.Handler.HandleGetOverview`| `rfq_pricing_optimizations` | Returns pricing analysis, facts, and approval status. |
| **Pricing Preview** | `rfqService.calculatePricingPreview`| `POST` | `/api/v1/rfqs/{id}/pricing-workflow/pricing-preview`| `pricing_workflow.Handler.HandlePricingPreview`| Stateless computation | Returns margin strategies, recommended price, and cost breakdown. |
| **Generate Quote Draft**| `rfqService.generateDraft`| `POST` | `/api/v1/rfqs/{id}/pricing-workflow/drafts`| `pricing_workflow.Handler.HandleGenerateDraft`| `quotations`, `rfq_pricing_optimizations` | Persists quotation draft; advances RFQ stage. |
| **Booking Handoff** | `rfqService.getRFQBookings` | `GET` | `/api/v1/rfqs/{id}/bookings` | `endpoints.GetBookingHandoffEP` $\rightarrow$ `booking_engine` | `bookings`, `rfqs` | Returns booking eligibility and linked bookings. |
| **Create Booking** | `rfqService.createBooking` | `POST` | `/api/v1/rfqs/{id}/bookings` | `endpoints.CreateBookingEP` $\rightarrow$ `booking_engine` | `bookings`, `audit_logs` | Creates confirmed booking; triggers carrier EDI integration. |
| **Shipment Handoff** | `rfqService.getRFQShipments`| `GET` | `/api/v1/rfqs/{id}/shipments` | `endpoints.GetShipmentHandoffEP` | `shipments` | Returns linked operational tracking records. |

---

## 16. Frontend Component Map

- **`RFQPage.jsx`:** Top-level directory page. Manages search filtering (`searchQuery`), stage tabs (`DRAFT`, `AWAITING_QUOTE`, `WON`, `LOST`), pagination, and modal toggles (`RFQBuilder`).
- **`RFQList.jsx`:** Renders the main table view with responsive column layout, progress completeness meters, Incoterm badges, and deep-link buttons.
- **`RFQBuilder.jsx`:** Modal form for creating a new RFQ. Integrates free-text AI paste parsing (`parseShipmentRequest`) and dynamic cargo line-item additions.
- **`RFQDetailPage.jsx`:** Master workspace container. Controls URL search params (`?tab=...`), header metadata, and lazy tab hydration.
- **`RFQHeader.jsx`:** Displays RFQ number, status badge, customer card, trade lane route tags, completeness indicator, and primary action dropdowns (`Download PDF`, `More Actions`).
- **`RFQOverview.jsx`:** Summarizes key dates, commodity descriptions, Incoterms, and commercial account assignments.
- **`RFQCargoItems.jsx`:** Manages container equipment specifications, piece counts, gross weight, and cubic volume tables.
- **`RFQRequirements.jsx`:** Operational readiness scorecard. Highlights blocking requirements (missing commercial invoice, unconfirmed weight) and conditional compliance flags.
- **`RFQDocuments.jsx`:** Document repository for attaching, viewing, and verifying required shipping paperwork.
- **`RFQQuotes.jsx`:** Carrier quote comparison matrix. Visualizes ocean freight, origin/destination terminal charges, transit times, and recommendation toggles.
- **`RFQPricingIntelligenceSection.jsx`:** Displays historical lane price benchmarks, spot market trends, and predictive margin risk cards.
- **`RFQIntelligentPricingWorkflowSection.jsx`:** AI Quotation Workflow engine (Phase 3 Task 3.4). Allows pricing specialists to configure target margins, run pricing simulations, and generate quotations.
- **`RFQBookingHandoff.jsx`:** Post-victory operational module. Bridges accepted quotes into confirmed carrier bookings.
- **`RFQShipmentHandoff.jsx`:** Displays active shipment milestones, container tracking IDs, and delivery progress.
- **`RFQActivityTimeline.jsx`:** Chronological event log displaying all actions taken on the RFQ.

---

## 17. Go Backend Component Map

- **`transport.go`:** Chi router registration and Go-Kit HTTP decoders/encoders. Enforces JSON parsing, URL parameter extraction, and HTTP status code mappings.
- **`endpoints.go`:** Go-Kit RPC endpoints. Decouples transport protocol from core domain services.
- **`bl.go`:** Central business logic service (`businessLogic`). Orchestrates stage transitions, calls the Event Mesh, interfaces with the AI sidecar client, and records audit logs.
- **`dl.go`:** Data access layer (`dataLayer`). Executes prepared SQL queries against MariaDB with strict `org_id` parameters.
- **`requirements_engine.go`:** Rule evaluation engine. Deterministically computes RFQ operational readiness (`INFORMATION_REQUIRED`, `REQUIREMENTS_INCOMPLETE`, `READY_FOR_QUOTATION`).
- **`quote_engine.go`:** Evaluates carrier quotes, enforces quote approval criteria, and checks carrier reliability scores.
- **`booking_engine.go`:** Governs booking eligibility rules, validates approved quotations, and creates operational booking records.
- **`carrier_booking_engine.go`:** Interfaces with external carrier booking systems (e.g. Inttra, direct shipping line APIs).
- **`pricing_workflow/service.go`:** High-level AI pricing workflow engine. Connects Go with the Python AI sidecar, applies margin policies, and handles quotation draft generation.

---

## 18. Python / Go Boundary Verification

LogisticsHQ maintains a strict architectural division between Python AI processing and Go business execution:

| Responsibility | Python AI Sidecar (`ai_sidecar`) | Go Backend (`backend/internal`) |
| :--- | :--- | :--- |
| **Reasoning & Natural Language** | Executes LLMs for unstructured shipment extraction (`RFQParserAgent`). | None. Calls Python via HTTP client. |
| **Statistical Pricing Modeling** | Evaluates lane volatility, predicts fuel adjustments. | Validates pricing math, enforces margin bounds. |
| **Authentication & Authorization** | None. Completely stateless service. | Verifies JWT tokens, enforces RBAC roles. |
| **Tenant Isolation** | Receives sanitized `org_id` in payload. | Enforces `WHERE org_id = ?` on all database operations. |
| **Database Mutations** | **STRICTLY FORBIDDEN.** Zero business SQL writes. | Authoritative system of record. Executes all transactional writes. |
| **External Integrations & Actions** | None. | Dispatches emails, calls carrier booking APIs, writes audit records. |

---

## 19. RFQ AI Features

1. **Unstructured Shipment Request Parsing (`RFQParserAgent`):**
   - **Input:** Raw email text or broker message.
   - **Processing:** LLM extracts POL, POD, Incoterm, gross weight, and volume.
   - **Output:** Structured JSON with confidence score (0-100) and list of missing fields.
   - **Gating:** Go validates extracted data before presenting it in the UI.
2. **Operational Requirements & Gap Extraction (`ExtractRequirements`):**
   - **Input:** RFQ route, cargo commodity, container type.
   - **Processing:** Identifies mandatory vs optional regulatory requirements (e.g. IMO declaration for chemicals).
   - **Output:** Populates `ai_rfq_requirements_extractions` with status `CLARIFICATION_REQUIRED` or `READY`.
3. **Predictive Landed Cost & Margin Optimization (`CalculatePricingPreview`):**
   - **Input:** Approved carrier buy price + trade lane benchmark.
   - **Processing:** Analyzes lane risk and suggests optimal sell price to maximize win probability.
   - **Output:** Three strategy options (Conservative, Balanced, Aggressive) with margin risk flags.

---

## 20. Action System Integration

When an RFQ operation produces an external side effect or alters financial commitments, it is routed through the Go Action System:

- **Generate Quotation Action:** Validates user authority $\rightarrow$ checks margin floor $\rightarrow$ registers action $\rightarrow$ writes quotation record $\rightarrow$ logs audit.
- **Carrier Booking Action:** Verifies RFQ is in `WON` stage $\rightarrow$ checks carrier integration credentials $\rightarrow$ dispatches booking request $\rightarrow$ updates status to `BOOKING_REQUESTED`.
- **Direct Bypass Attempt:** Direct API calls attempting to skip required stages or bypass approval thresholds return HTTP 400 or HTTP 403.

---

## 21. Approvals

The RFQ module integrates with the centralized `approvals` service (`backend/internal/approvals`):

- **Trigger:** Quotation margin is calculated below the organization's minimum policy threshold (`gross_margin_pct < 8.00%`).
- **Approval Workflow:**
  1. `pricing_workflow.HandleSubmitApproval` creates an approval request record in `approvals` table (`type = 'QUOTE_MARGIN_EXCEPTION'`).
  2. The RFQ draft quotation status transitions to `SUBMITTED_FOR_REVIEW`.
  3. Commercial Director receives an in-app notification and reviews the margin justification.
  4. Upon approval (`status = 'APPROVED'`), the quotation is unlocked for customer dispatch. If rejected, the draft returns to `CHANGES_REQUESTED`.

---

## 22. Notifications

- **In-App Alerts:** Published to `notifications` table for assigned sales and pricing agents when an RFQ is assigned, quoted, or won.
- **Unread Counter:** Badge updates via `GET /api/v1/notifications/unread-count`.
- **Customer Email Dispatch:** Triggered via SES email integration when a quotation transitions to `SENT`.

---

## 23. Event Mesh / Automation

The RFQ module publishes domain events to the internal Go `eventBus`:

- `events.EventRFQCreated`: Triggered upon new RFQ creation.
- `events.EventRFQAssigned`: Triggered when pricing specialist is assigned.
- `events.EventQuoteGenerated`: Triggered when quotation draft is compiled.
- `events.EventQuoteSent`: Triggered upon quotation dispatch.
- `events.EventRFQWon`: Triggered upon quotation acceptance. Unlocks booking automation.
- `events.EventRFQLost`: Triggered upon inquiry cancellation or expiration.

---

## 24. Permissions / RBAC

Enforced at the Go transport and business logic layers:

| Permission / Role | View RFQ | Create RFQ | Edit Cargo | Procure Rates | Approve Quote | Book Carrier |
| :--- | :---: | :---: | :---: | :---: | :---: | :---: |
| **`SUPER_ADMIN` / `ORG_ADMIN`** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **`COMMERCIAL_MANAGER`** | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **`PRICING_SPECIALIST`** | ✅ | ❌ | ❌ | ✅ | ✅ (Within margin) | ❌ |
| **`SALES_REPRESENTATIVE`** | ✅ | ✅ | ✅ | ❌ | ❌ | ❌ |
| **`OPERATIONS_DISPATCHER`**| ✅ | ❌ | ❌ | ❌ | ❌ | ✅ |

---

## 25. Tenant Isolation

Multi-tenancy is enforced at every layer of the architecture:

1. **Authentication Token:** JWT claims contain verified `org_id`.
2. **Go Context Injection:** Extracted by `middleware.RequireAuth` and placed into request `context.Context`.
3. **Database Enforcement:** Every single SQL query incorporates `org_id = ?`:
   ```sql
   SELECT * FROM rfqs WHERE org_id = ? AND id = ?
   UPDATE rfqs SET stage = ? WHERE org_id = ? AND id = ?
   ```
4. **Sub-Resource Containment:** Line items (`rfq_items`), quotes (`rfq_quotes`), and optimizations (`rfq_pricing_optimizations`) are scoped to the organization. Requests targeting records belonging to another tenant fail immediately with resource-not-found errors.

---

## 26. Audit Trail

Every state-altering RFQ event records an entry in `audit_logs`:
- **Module:** `ModuleRFQs`
- **Resource Type:** `RFQ`
- **Resource ID:** String representation of `rfq.ID`
- **Actions Recorded:** `CREATE`, `UPDATE`, `ActionAssign`, `ActionApprove`, `ActionDisable`
- **Data Payload:** Captures `Before` and `After` JSON snapshots (e.g., stage change from `PRICING_ASSIGNED` to `QUOTE_GENERATED`).

---

## 27. Search / Filter / Sort / Pagination

- **Search:** `searchQuery` checks `rfq_number`, `customer_name`, `origin`, `destination`, and `incoterms`.
- **Mode Filter:** Filters by transport mode (`ALL`, `OCEAN`, `AIR`, `ROAD`).
- **Status Filter:** Filters by stage (`ALL`, `RFQ_CREATED`, `PRICING_ASSIGNED`, `QUOTE_SENT`, `WON`, `LOST`).
- **Incoterm Filter:** Filters by commercial term (`FOB`, `CIF`, `EXW`, `CIP`, `DDP`).
- **Pagination:** Uses client-side slicing and backend SQL `LIMIT ? OFFSET ?` (default 8 items per page).

---

## 28. Business User Journeys

### Journey 1: Standard Customer Inquiry to Quote
1. Sales rep receives email from *Acme Corp* requesting rates for 2x40HC from Nhava Sheva to Hamburg.
2. Rep opens `/dashboard/rfqs` and clicks `+ New RFQ`.
3. Rep inputs route, equipment, and target date, then clicks `Create RFQ`.
4. RFQ enters `PRICING_ASSIGNED`. Pricing team queries carrier rates, selects Maersk Line tariff ($3,200 buy price), and applies 15% margin.
5. Quotation `QT-2026-0012` is generated ($3,764.71) and emailed to customer.
6. Customer accepts. RFQ transitions to `WON`.

### Journey 2: AI Email Extraction to Quotation Draft
1. Forwarder receives complex multi-paragraph freight inquiry via email.
2. Operator opens `RFQBuilder`, pastes raw email body into "AI Shipment Extractor", and clicks "Parse".
3. Python `RFQParserAgent` extracts POL (Nhava Sheva), POD (Hamburg), weight (15,000 kg), and volume (30 CBM) with 95% confidence.
4. Operator verifies extracted data and clicks "Confirm & Create".
5. Operator navigates to `⚡ AI Quotation Workflow` tab, reviews cost forecast, and generates formal quotation draft in one click.

---

## 29. Data Flow Diagrams

### Manual RFQ Inception & Quotation
```
[User Form] ──► [RFQBuilder.jsx] ──► POST /api/v1/rfqs ──► [bl.CreateRFQ]
                                                                  │
       ┌──────────────────────────────────────────────────────────┴────────────────────────┐
       ▼                                                                                   ▼
[MariaDB rfqs / rfq_items]                                                         [audit_logs]
       │                                                                                   │
       ▼                                                                                   ▼
[GET /api/v1/rfqs/:id] ◄── [RFQDetailPage.jsx] ◄── [Event Mesh: EventRFQCreated] ──────────┘
```

### AI Inbound Email Processing
```
[Inbound Email Text] ──► POST /parse-shipment-request ──► [Go Backend]
                                                                │
                                                                ▼
                                                [Python AI Sidecar:8090]
                                                                │
                                                                ▼
                                                      [RFQParserAgent]
                                                                │
                                                                ▼
[Pre-populated Form] ◄── Response: {origin, dest, weight, vol} ─┘
```

---

## 30. Source-of-Truth Matrix

| Data Domain | Classification | Authoritative System | Persistence Location |
| :--- | :--- | :--- | :--- |
| **RFQ Header & Stage** | Authoritative Data | Go Backend / MariaDB | `rfqs` table |
| **Cargo Line Items** | Authoritative Data | Go Backend / MariaDB | `rfq_items` table |
| **Carrier Buy Rates** | External / Verified Data | Carrier Contracts / Tariffs | `rfq_quotes`, `rates` tables |
| **Operational Readiness** | Derived Data | Go Requirements Engine | Evaluated on demand |
| **Extracted Requirements**| AI Prediction | Python AI Sidecar | `ai_rfq_requirements_extractions` |
| **Recommended Sell Price**| AI Recommendation | Python / Go Workflow | `rfq_pricing_optimizations` table |
| **Customer Quotation** | Authoritative Data | Go Quotation Service | `quotations` table |
| **Confirmed Booking** | Authoritative Operational | Carrier Booking Engine | `bookings` table |

---

## 31. Error, Loading, and Empty States

- **Loading State:** Displays structured skeleton cards and loading banner (`"Loading RFQ Workspace & syncing operational records..."`) to prevent false zeros or flickering.
- **Empty Directory:** Displays friendly logistics illustration and empty state callout: *"No RFQs found matching your criteria. Create your first RFQ to get started."*
- **Empty Subtabs:** Inactive subtabs (e.g. Quotes, Bookings) render clean guidance cards explaining how to advance the RFQ to unlock downstream features.
- **Validation Errors:** Required field omissions return structured JSON errors (`400 Bad Request`) and render toast notifications without crashing the UI.

---

## 32. UI / UX Observations

### What Works Exceptionally Well:
- **Comprehensive 360 Workspace:** Having Overview, Cargo, Requirements, Documents, Activity, Quotes, Pricing Intel, and Booking Handoff tabs in a single screen eliminates fragmented navigation.
- **Completeness Meter:** Clear visual progress bar (`4/7 Parameters 57%`) guides sales reps on required fields.
- **Stage Badging:** Consistent, color-coded stage indicators (`Created`, `Pricing Assigned`, `Won`, `Lost`).

### Areas for Improvement (Prioritized for Task 3.4 Deep Review):
- **SHOULD IMPROVE:** In the main directory table, the Customer column currently displays generic fallback labels (`Customer #1`) for legacy records where `customers.name` was not joined. Ensure all active customer records hydrate legal names cleanly.
- **SHOULD IMPROVE:** The AI Quotation Workflow tab provides deep pricing capabilities; adding inline rate-refresh buttons will enhance operator efficiency.
- **OPTIONAL:** Add batch-selection checkboxes in the RFQ directory to allow bulk status updates.

---

## 33. Responsive & Zoom Observations

- **1440 × 900 (Standard Desktop):** Flawless layout. Side navigation docked, table columns spacious, tab bar fits cleanly.
- **1366 × 768 (Small Laptop):** Header badges wrap naturally without overlapping; action buttons remain accessible.
- **1280 × 720 (Compact Display):** Workspace tabs scroll horizontally; main content retains full legibility with zero clipping.
- **Zoom Levels (80% - 125%):** Tested scaling across 80%, 90%, 100%, 110%, and 125%. Text remains crisp, card padding adjusts proportionally, and modals render centered.

---

## 34. Security Architecture

- **Authentication:** Verified through AWS Cognito JWT Bearer tokens.
- **RBAC Enforcement:** Fine-grained role permissions prevent unauthorized users from viewing confidential carrier buy prices or releasing sub-margin quotations.
- **Tenant Isolation:** Guaranteed via `WHERE org_id = ?` filters on all MariaDB queries.
- **Stateless AI Boundary:** Python AI services operate without direct database credentials, communicating solely through sanitized JSON contracts.
- **Fail-Closed Design:** Unauthenticated or invalid token requests fail immediately with HTTP 401.

---

## 35. Business & Technical Glossary

- **POL (Port of Loading):** The port or terminal where cargo is loaded onto the operating vessel.
- **POD (Port of Discharge):** The destination port or terminal where cargo is unloaded.
- **Incoterms:** Standardized three-letter trade terms (e.g. FOB = Free on Board, CIF = Cost Insurance Freight) defining buyer/seller risk and cost allocation.
- **BAF (Bunker Adjustment Factor):** Surcharge applied by ocean carriers to compensate for fuel price fluctuations.
- **THC (Terminal Handling Charges):** Charges levied by port terminals for loading, unloading, and staging containers.
- **Demurrage:** Penalty fees charged when containers remain inside the port terminal beyond agreed free days.
- **Detention:** Penalty fees charged when containers are held outside the terminal beyond agreed free days.

---

## 36. One-Page “How RFQs Work”

```
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                        HOW RFQs WORK IN LOGISTICSHQ                                   │
└────────────────────────────────────────────────────────────────────────────────────────┘

1. INQUIRY INTAKE
   A freight inquiry arrives via email, customer portal, or sales input.
   AI extracts key shipment data (origin, destination, weight, volume).
   An RFQ is opened with a unique code (e.g. RFQ-2026-1001).

2. OPERATIONAL READINESS
   The Requirements Engine checks cargo feasibility, mandatory documents
   (commercial invoices, packing lists), and dangerous goods compliance.

3. CARRIER PROCUREMENT
   The pricing team queries ocean/air carrier tariffs, secures buy rates,
   and compares transit times and free days across carriers.

4. PRICING & MARGIN MODELING
   AI evaluates spot market trends and recommends sell prices.
   Target margins (e.g. 15%) are applied. If margin is < 8%, an approval request
   is automatically submitted to commercial management.

5. QUOTATION DISPATCH
   An authoritative quotation (e.g. QT-2026-0012) is compiled and dispatched
   to the customer. RFQ stage advances to QUOTE_SENT.

6. CLOSURE & OPERATIONAL HANDOFF
   - If customer accepts: RFQ is marked WON.
   - Converts seamlessly into a confirmed BOOKING and active SHIPMENT tracking record.
```

---

## 37. Technical Traceability Matrix

| Capability | Frontend Component | API Endpoint | Go Service | Python AI | Database Table | Event Mesh | Permission |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **RFQ List** | `RFQList.jsx` | `GET /api/v1/rfqs` | `dl.ListRFQs` | N/A | `rfqs`, `customers` | N/A | `rfq:read` |
| **RFQ Detail** | `RFQDetailPage.jsx` | `GET /api/v1/rfqs/:id` | `dl.GetRFQByID` | N/A | `rfqs`, `rfq_items` | N/A | `rfq:read` |
| **Create RFQ** | `RFQBuilder.jsx` | `POST /api/v1/rfqs` | `bl.CreateRFQ` | N/A | `rfqs`, `rfq_items` | `EventRFQCreated` | `rfq:write` |
| **AI Parsing** | `RFQBuilder.jsx` | `POST /parse-shipment-request` | `bl.ParseShipmentRequest` | `RFQParserAgent` | N/A | N/A | `rfq:write` |
| **Stage Change** | `RFQHeader.jsx` | `PUT /api/v1/rfqs/:id/stage` | `bl.UpdateStage` | N/A | `rfqs`, `audit_logs` | `EventRFQUpdated` | `rfq:manage` |
| **Carrier Quotes** | `RFQQuotes.jsx` | `GET /api/v1/rfqs/:id/quotes` | `dl.ListQuotes` | N/A | `rfq_quotes` | N/A | `pricing:read` |
| **AI Pricing** | `RFQIntelligentPricingWorkflowSection.jsx` | `POST /pricing-workflow/pricing-preview` | `pricing_workflow.Service` | `rfq_pricing/agent.py` | `rfq_pricing_optimizations` | N/A | `pricing:manage` |
| **Draft Quote** | `RFQIntelligentPricingWorkflowSection.jsx` | `POST /pricing-workflow/drafts` | `pricing_workflow.Service` | N/A | `quotations` | `EventQuoteGenerated` | `pricing:manage` |
| **Booking Handoff**| `RFQBookingHandoff.jsx` | `POST /api/v1/rfqs/:id/bookings` | `booking_engine.Service` | N/A | `bookings` | `EventRFQWon` | `booking:write` |

---

## 38. Known Gaps

1. **DOCUMENTATION / DATA SEEDING GAP:** Certain legacy seed records in `rfqs` reference customer IDs (`customer_id = 1, 2, 3`) that exist in older database fixtures without matching legal name records in the current tenant table. This causes the UI to fall back to `Customer #1` instead of a named corporation. (Addressed in Task 3.4 functional verification).
2. **EXTERNAL INTEGRATION GAP:** Live carrier EDI booking transmission (`carrier_booking_engine.go`) requires live ocean carrier API credentials (e.g. Inttra / Maersk API). In the local development environment, carrier booking operations execute against verified internal state machines and mock adapters.
3. **UI/UX GAP:** Adding inline quote regeneration buttons on the `Quotes` tab will allow operators to refresh expired carrier tariffs directly without navigating between sub-workspaces.

---

## 39. Verification Status

- **Running Go Backend:** Verified live responses on port 8080 (`/api/v1/rfqs`, `/api/v1/rfqs/1/requirements`, `/api/v1/rfqs/1/quotes`, `/api/v1/rfqs/1/pricing-workflow/overview`, `/api/v1/rfqs/1/bookings`, `/api/v1/rfqs/1/shipments` all returned HTTP 200).
- **Running Python AI Sidecar:** Verified live responses on port 8090 (`/health` returned HTTP 200, stateless extraction agents loaded).
- **MariaDB Schema:** Verified physical tables (`rfqs`, `rfq_items`, `rfq_quotes`, `rfq_pricing_optimizations`, `rfq_pricing_versions`, `quotations`, `bookings`, `shipments`).
- **Chrome Browser Validation:** Captured and inspected live screenshots (`rfq_list_live.png`, `rfq_detail_live.png`).

### Final Acceptance Verdict:
**PASS — RFQ WORKFLOW DOCUMENTED**
