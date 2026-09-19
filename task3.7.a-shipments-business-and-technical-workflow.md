# Task 3.7.A — Shipments Business Workflow, Shipment Lifecycle, Tracking/Milestones, Exception Flow, Database Mapping, API Traceability, AI Integration, and Complete Technical Documentation

> **Status**: PASS — SHIPMENT WORKFLOW DOCUMENTED  
> **System**: LogisticsHQ (Freight Forwarding & Autonomous Logistics OS)  
> **Module**: Shipments Operations, Carrier Tracking, Milestones & Disruption Management  
> **Target Document**: `task3.7.a-shipments-business-and-technical-workflow.md`  
> **Verification Date**: September 13, 2026  
> **Review Scope**: React Frontend, Go Backend, MariaDB Database, Python AI Sidecar, RBAC & Multi-Tenant Isolation  

---

## 1. Executive Summary

The **Shipments Module** in LogisticsHQ represents the operational execution phase of international freight forwarding. While an **RFQ** handles commercial customer requests, a **Quotation** models buy/sell pricing and carrier cost margins, and a **Booking** reserves physical vessel space and equipment allocation with a carrier line, a **Shipment** is the live, physical execution contract that moves cargo from origin to destination across maritime, inland, and customs corridors.

This document presents the verified, end-to-end reality of the LogisticsHQ Shipments module as it currently functions across the React frontend, Go backend micro-framework, MariaDB database, carrier adapter engine, and Python AI Sidecar. It is specifically designed to be immediately understandable by non-technical logistics coordinators, business owners, and customer service teams, while providing exhaustive technical traceability, SQL queries, HTTP endpoint specifications, and component mappings for software developers, QA engineers, and platform maintainers.

### Key Architectural & Operational Highlights
1. **Lineage-Grounded Creation**: Shipments are instantiated directly from confirmed upstream Bookings via [CreateShipmentFromBookingTx](file:///c:/Users/Sai/go/src/freel-project/backend/internal/rfq/dl.go#L1544) or auto-created upon Won RFQ confirmation via [CreateFromRFQ](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/bl.go#L114). Direct manual creation with container allocation is supported via the operations UI.
2. **Canonical Five-Milestone Lifecycle**: Every shipment is anchored to a standard progressive milestone chain (`BOOKED` &rarr; `DEPARTED` &rarr; `IN_TRANSIT` &rarr; `ARRIVED` &rarr; `DELIVERED`). Milestone completions cascade to the master shipment status with monotonic progression checks that strictly prevent out-of-order date regression.
3. **Multi-Channel Carrier Telemetry & Normalization**: Carrier updates arrive via three distinct channels:
   - **Carrier REST/DCSA APIs**: Handled on-demand or via the [CarrierTrackingEngine](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/carrier_tracking_engine.go#L59) adapter registry.
   - **Real-Time Carrier Webhooks**: Ingested at `/webhooks/carriers/{carrier}` with HMAC signature validation and automated multi-tenant correlation.
   - **Background Polling Scheduler**: Executed by the status-aware [TrackingRefreshScheduler](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/tracking_scheduler.go#L16) and distributed [CarrierPoller](file:///c:/Users/Sai/go/src/freel-project/backend/internal/jobs/carrier_poller.go#L14) with database row leasing (`FOR UPDATE`).
4. **Deterministic Exception & Disruption Engine**: Operational exceptions are categorized across 11 standard types (e.g., `SCHEDULE_DELAY`, `VESSEL_ROLLOVER`, `PORT_CONGESTION`, `CUSTOMS_HOLD`) and four severities (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`). Factual carrier exceptions are strictly isolated from predictive AI delay risk projections.
5. **Strict Go / Python AI Boundary**: All business logic, database transactions, tenant isolation, carrier API communications, and audit records are executed strictly within the Go backend. The Python AI Sidecar serves solely as a stateless prediction, natural language synthesis, and reasoning provider. It cannot directly mutate shipment records or execute SQL queries.

---

## 2. Shipments in Plain English

### What is a Shipment?
In freight forwarding, a **Shipment** is the actual journey of physical cargo being moved from a seller (shipper) to a buyer (consignee). If a *Booking* is like buying an airline ticket, the *Shipment* is boarding the plane, flying through the sky, clearing customs, and arriving at the destination baggage claim.

### Why Does it Exist?
A booking simply reserves space on a ship or airplane. Once the carrier confirms the space, logistics operators need a dedicated workspace to:
- Track where the cargo and containers physically are in the world.
- Ensure the containers get loaded onto the vessel before the cutoff time.
- Monitor whether the vessel departs and arrives on schedule.
- Catch delays, customs holds, or missed connections as early as possible.
- Store mandatory shipping documents (Bills of Lading, Customs Declarations, Commercial Invoices).
- Track actual operational costs against customer billing to protect profit margins.
- Keep the cargo owner (the customer) updated with trustworthy arrival times.

### Who Works with Shipments?
- **Operations Coordinators**: Monitor active voyages, log container movements, update milestones, and coordinate container pickups.
- **Tracking & Tracing Teams**: Ingest carrier updates, monitor vessel positions, and flag route deviations.
- **Customer Service Teams**: Answer customer questions about arrival dates and send proactive delay notices.
- **Customs Brokers & Documentation Specialists**: Verify that Bills of Lading, packing lists, and clearance certificates match before the ship berths.
- **Finance & Billing Specialists**: Audit ocean freight invoices, detention/demurrage fees, and issue final customer invoices.

### How Tracking & Milestones Work
A shipment's journey is broken into checkpoints called **Milestones**. As the ship moves across oceans, ocean carriers send electronic messages (via EDI, webhooks, or satellite AIS trackers). When an event occurs—such as a container passing through port gates or a vessel leaving port—LogisticsHQ records the exact timestamp and location, advancing the shipment from `BOOKED` to `DEPARTED` to `IN_TRANSIT` to `ARRIVED` to `DELIVERED`.

### What Happens When a Shipment is Delayed?
If bad weather slows down a ship, or if a port is congested, an **Exception** is created in the system. Operations staff can acknowledge the exception, contact the carrier for an updated schedule, and notify the customer. Once the issue is resolved (e.g., the container is released from customs hold), the exception is marked `RESOLVED`, and the shipment resumes its standard operational flow.

---

## 3. Business Purpose

The Shipments module exists to fulfill four critical commercial and operational mandates:

```
┌───────────────────────────────────────────────────────────────────────────────┐
│                        CORE SHIPMENT MANDATES                                 │
├───────────────────────┬────────────────────────┬──────────────────────────────┤
│ 1. VISIBILITY         │ 2. CONTROL             │ 3. FINANCIAL INTEGRITY       │
│ Real-time container   │ Immediate detection of │ Real-time comparison of buy  │
│ tracking, vessel AIS, │ customs holds, rolled  │ rates, carrier invoices, and │
│ and milestone proof.  │ cargo, and ETA drift.  │ final customer billing.      │
├───────────────────────┴────────────────────────┴──────────────────────────────┤
│ 4. COLLABORATIVE EXCELLENCE                                                   │
│ Single source of truth connecting forwarders, ocean carriers, customs        │
│ authorities, and cargo owners.                                               │
└───────────────────────────────────────────────────────────────────────────────┘
```

Without the Shipments module, logistics coordinators would be forced to manually log into dozens of carrier tracking portals, copy-paste container numbers into spreadsheets, and discover costly delivery delays days after they occurred. LogisticsHQ unifies this entire operational lifecycle into a single, automated, multi-tenant workspace.

---

## 4. Booking &rarr; Shipment Business Flow

The handoff from a commercial Booking to operational Shipment execution is a governed milestone in LogisticsHQ:

```
[Customer Accepts Quote]
          │
          ▼
[Booking Created (DRAFT)]
          │
          ▼
[Carrier EDI Space Request] ─── (Carrier Booking Confirmed)
          │
          ▼
[Booking Confirmed (CONFIRMED)]
          │
          ├──> Trigger: Operations Coordinator clicks "Convert to Shipment"
          │    OR System Auto-converts on EDI 301 confirmation
          ▼
[CreateShipmentFromBookingTx]
          │  1. Verifies Booking status == 'CONFIRMED'
          │  2. Checks organizational ownership & deep RFQ lineage
          │  3. Performs idempotency check (returns existing if already active)
          │  4. Copies Ports, SCAC, Vessel, Voyage, ETD, ETA, & Containers
          │  5. Inserts record into `shipments` table with status 'BOOKED'
          │  6. Seeds 5 standard milestones in `shipment_milestones`
          │  7. Emits universal activity log for BOOKING and SHIPMENT
          │  8. Publishes 'shipment.created' to Event Bus
          ▼
[Active Shipment Workspace]
          │
          ├──> Carrier Tracking & Polling begins
          ├──> Document compliance checklist generated
          └──> Real-time margin tracking activated
```

### Actual Step-by-Step Implementation Trace:
1. **Trigger**: An operations coordinator views a confirmed booking at `/dashboard/bookings/{id}` and clicks **Convert to Shipment**, or an inbound EDI 301 electronic confirmation completes the booking.
2. **Authorization & Tenant Guard**: The request is validated by [RequireAuth](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/routes.go#L59) and [RBAC](file:///c:/Users/Sai/go/src/freel-project/backend/internal/rbac/service.go#L136) (`SHIPMENTS:CREATE`). The caller's `org_id` is extracted from the JWT token.
3. **Lineage Validation**: The backend executes a relational check against `rfqs` and `bookings` to ensure the booking belongs to the caller's organization.
4. **Idempotent Record Creation**: The Go backend runs [CreateShipmentFromBookingTx](file:///c:/Users/Sai/go/src/freel-project/backend/internal/rfq/dl.go#L1544) inside a database transaction (`BEGIN...COMMIT`). If an active shipment already exists for this booking, the existing record is returned without duplicate insertion.
5. **Milestone Seeding**: The system automatically inserts initial milestone records into `shipment_milestones`:
   - `BOOKED`: Status `COMPLETED` (planned and actual timestamps match booking confirmation).
   - `DEPARTED`: Status `PLANNED` (target date: ETD).
   - `IN_TRANSIT`: Status `PLANNED` (target date: ETD + voyage mid-point).
   - `ARRIVED`: Status `PLANNED` (target date: ETA).
   - `DELIVERED`: Status `PLANNED` (target date: ETA + inland transit buffer).
6. **Audit & Event Generation**:
   - Inserts audit record into `activities` (`entity_type = 'SHIPMENT'`, `action = 'SHIPMENT_CREATED'`).
   - Inserts linked audit record into `activities` (`entity_type = 'BOOKING'`, `action = 'SHIPMENT_CREATED'`).
   - Publishes `shipment.created` event on the in-process event bus.

---

## 5. Shipment Lifecycle

LogisticsHQ implements a canonical seven-state shipment lifecycle defined in [spec/const.go](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/spec/const.go#L3).

```
 ┌─────────────────┐
 │ BOOKING_PENDING │  (Shipment shell created from RFQ prior to carrier booking confirmation)
 └────────┬────────┘
          │ Carrier booking confirmation received
          ▼
     ┌────────┐
     │ BOOKED │  (Carrier confirmed; container allocated; awaiting origin terminal gate-in)
     └────┬───┘
          │ Origin departure verified (ATD recorded)
          ▼
    ┌──────────┐
    │ DEPARTED │  (Vessel departed origin port; maritime transit initiated)
    └─────┬────┘
          │ Deep-sea passage confirmed via AIS / Carrier EDI 315
          ▼
   ┌────────────┐
   │ IN_TRANSIT │  (Vessel at sea; position tracked along navigational corridor)
   └──────┬─────┘
          │ Destination berthing / container discharge verified (ATA recorded)
          ▼
    ┌─────────┐
    │ ARRIVED │  (Vessel berthed; containers discharged; awaiting customs clearance)
    └────┬────┘
          │ Final consignee delivery order executed & empty container returned
          ▼
   ┌───────────┐
   │ DELIVERED │  (Cargo delivered to consignee; operational journey complete)
   └───────────┘
```

### Shipment Lifecycle Transition Matrix

| Current Status | Allowed Next Status | Business Meaning | Who Can Change It | Validation & Conditions | Side Effect |
| :--- | :--- | :--- | :--- | :--- | :--- |
| `BOOKING_PENDING` | `BOOKED`, `EXCEPTION` | Shipment initiated but carrier space not yet confirmed | System / Operations Coordinator | Carrier booking reference assigned | Seeding of standard milestone sequence; emit `shipment.status_updated` |
| `BOOKED` | `DEPARTED`, `EXCEPTION` | Carrier confirmed space; cargo gated-in at origin port | Carrier API / Webhook / Operations | Actual departure date (ATD) recorded | Marks `DEPARTED` milestone `COMPLETED`; recalculates transit variance |
| `DEPARTED` | `IN_TRANSIT`, `EXCEPTION` | Vessel left origin berth; entering deep-sea corridor | Carrier API / AIS / Operations | Position coordinates update outside origin port | Activates waypoint passage tracking; marks `IN_TRANSIT` milestone `COMPLETED` |
| `IN_TRANSIT` | `ARRIVED`, `EXCEPTION` | Cargo traveling on maritime navigational corridor | Carrier API / AIS / Operations | Destination port arrival event (ATA) recorded | Marks `ARRIVED` milestone `COMPLETED`; triggers customs document check |
| `ARRIVED` | `DELIVERED`, `EXCEPTION` | Cargo discharged at destination port terminal | Carrier API / Operations / Trucking | Proof of Delivery (POD) uploaded or Gate-Out event | Marks `DELIVERED` milestone `COMPLETED`; flags shipment as `READY_FOR_CLOSURE` |
| `DELIVERED` | `CLOSED` | Operational and commercial journey finished | Operations Manager / Finance | Zero open critical exceptions; all billing charges approved | Sets `closure_status = 'CLOSED'`; locks operational modifications |
| `EXCEPTION` | Previous Active Status (`BOOKED`, `DEPARTED`, etc.) | An operational blocker has temporarily flagged the movement | Operations Coordinator / System | All open critical exceptions marked `RESOLVED` or `DISMISSED` | Restores status to highest completed milestone; logs restoration in audit trail |

> [!IMPORTANT]
> **Status Monotonicity Guard**: The Go backend enforces a strict rank check ([ShouldAdvanceStatus](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/carrier_tracking_engine.go#L40)). Stale or out-of-order carrier webhooks cannot regress a shipment's lifecycle status. If an `IN_TRANSIT` shipment receives a delayed `DEPARTED` webhook from a regional carrier server, the event is saved to `carrier_tracking_events`, but the master shipment status remains `IN_TRANSIT`.

---

## 6. Shipment Creation Paths

LogisticsHQ supports three distinct, verified shipment creation paths:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       SHIPMENT CREATION PATHWAYS                            │
├─────────────────────────┬─────────────────────────┬─────────────────────────┤
│ PATH A: BOOKING HANDOFF │ PATH B: RFQ DIRECT WIN  │ PATH C: MANUAL CREATION │
│ (Standard Workflow)     │ (Instant Contract)      │ (Spot / External Move)  │
│ Confirmed Booking       │ Won RFQ auto-generates  │ Operations creates spot │
│ converted via UI button │ shipment with status    │ shipment directly with  │
│ or carrier confirmation │ `BOOKING_PENDING`       │ container & SCAC inputs │
└─────────────────────────┴─────────────────────────┴─────────────────────────┘
```

### Path A: Booking &rarr; Shipment Handoff (Primary Path)
- **Business Trigger**: A customer accepts a quotation, carrier booking space is confirmed, and the logistics coordinator clicks **Convert to Shipment** on the Booking detail page.
- **Frontend Handler**: `handleCreateShipment()` in [BookingsPage.jsx](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Bookings/BookingsPage.jsx).
- **Backend API**: `POST /api/v1/bookings/{id}/shipments`.
- **Go Execution**: Calls [CreateShipmentFromBookingTx](file:///c:/Users/Sai/go/src/freel-project/backend/internal/rfq/dl.go#L1544).
- **Database Result**: Inserts into `shipments` with status `BOOKED`, copies container numbers, links `booking_id`, seeds milestones, and emits `activities`.

### Path B: Won RFQ Direct Auto-Creation
- **Business Trigger**: An RFQ is marked `WON` with an `APPROVED` carrier quote.
- **Go Execution**: [CreateFromRFQ](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/bl.go#L114) in `shipments.Service`.
- **Database Result**: Inserts into `shipments` with status `BOOKING_PENDING`, links `rfq_id` and `quote_id`, and resolves carrier SCAC via `carriers` table. Seeding creates 5 standard milestones with planned transit offsets.

### Path C: Manual Direct Shipment Creation
- **Business Trigger**: Operations coordinator handles a spot move or pre-existing carrier booking directly via the Shipments Workspace.
- **Frontend Handler**: `+ New Shipment` button on [ShipmentsPage.jsx](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Shipments/ShipmentsPage.jsx).
- **Go Handler**: [CreateShipmentTx](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/dl.go#L321).
- **Database Result**: Inserts record into `shipments` table with user-supplied container numbers, SCAC, origin, and destination.

---

## 7. Booking / RFQ / Quotation Relationships

LogisticsHQ maintains complete end-to-end commercial-to-operational traceability:

```
┌─────────────┐       1:N       ┌───────────────┐       1:1       ┌─────────────┐       1:1       ┌──────────────┐
│    RFQs     ├────────────────►│  QUOTATIONS   ├────────────────►│  BOOKINGS   ├────────────────►│  SHIPMENTS   │
│ (Commercial │                 │ (Buy / Sell   │                 │ (Carrier    │                 │ (Live Cargo  │
│  Inquiry)   │                 │  Pricing)     │                 │  Allocation)│                 │  Execution)  │
└──────┬──────┘                 └───────┬───────┘                 └──────┬──────┘                 └──────┬───────┘
       │                                │                                │                               │
       │ id                             │ id                             │ id                            │ id
       │ customer_id                    │ rfq_id                         │ quote_id                      │ booking_id
       │ rfq_number                     │ carrier_name                   │ booking_number                │ rfq_id
       └────────────────────────────────┴────────────────────────────────┴───────────────────────────────┘
```

### Relational Mapping:
- **Direct Relational Link**: `shipments.booking_id` &rarr; `bookings.id` (Foreign Key / Relational Join).
- **Upstream Inquiry Link**: `shipments.rfq_id` &rarr; `rfqs.id` (Enables one-click navigation from Shipment directly to original customer RFQ).
- **Commercial Pricing Link**: `shipments.quote_id` &rarr; `rfq_quotes.id` (Provides baseline buy cost and sell price for real-time margin calculations).
- **Customer Association**: Derived via `LEFT JOIN rfqs r ON s.rfq_id = r.id LEFT JOIN customers c ON r.customer_id = c.id`.

---

## 8. Shipment Data Fields

The canonical shipment data fields in LogisticsHQ are mapped between MariaDB storage, Go domain structs, and React UI presentation:

| Field Name | Database Column | Go Type | UI Representation | Source of Truth |
| :--- | :--- | :--- | :--- | :--- |
| **Shipment ID** | `shipments.id` | `int64` | `SH-280` | MariaDB Auto-Increment |
| **Organization ID** | `shipments.org_id` | `int64` | Tenant Scope (Internal) | Authenticated JWT Claims |
| **Booking Reference** | `shipments.booking_number` | `*string` | `BKG-1789237642119` | Upstream Booking / Carrier Confirmation |
| **Carrier SCAC** | `shipments.carrier_scac` | `string` | `MAEU` (Maersk Line) | Carrier Master Table (`carriers`) |
| **Master B/L (MBL)** | `shipments.mbl_number` | `*string` | `MAEU123456789` | Ocean Carrier Sea Waybill / Bill of Lading |
| **House B/L (HBL)** | `shipments.hbl_number` | `*string` | `LH-HBL-2601` | Freight Forwarder House Bill of Lading |
| **Container Numbers** | `shipments.container_numbers` | `JSONStringSlice` | `["MSKU9012345", "MSKU9012346"]` | Carrier Allocation / Equipment Manifest |
| **Operational Status**| `shipments.status` | `string` | Badge: `BOOKED`, `IN_TRANSIT` | Operational Lifecycle Engine |
| **Origin Port** | `shipments.origin_port` | `string` | `INNSA (Nhava Sheva)` | Port Master Table / Booking Spec |
| **Destination Port** | `shipments.destination_port` | `string` | `DEHAM (Hamburg)` | Port Master Table / Booking Spec |
| **Vessel Name** | `shipments.vessel_name` | `*string` | `MAERSK MC-KINNEY MOLLER` | Carrier Schedule / AIS Telemetry |
| **Voyage Number** | `shipments.voyage_number` | `*string` | `2604E` | Ocean Carrier Voyage Schedule |
| **Estimated Departure**| `shipments.etd` | `*time.Time` | `Aug 22, 2026, 11:00 PM` | Carrier Schedule Baseline |
| **Estimated Arrival** | `shipments.eta` | `*time.Time` | `Aug 28, 2026, 11:00 PM` | Carrier Schedule Baseline |
| **Closure Status** | `shipments.closure_status` | `string` | `ACTIVE`, `READY_FOR_CLOSURE` | Operational Closure Engine |
| **Customer Name** | Derived (`customers.name`)| `*string` | `Apex Global Logistics Corp` | Upstream Customer Record |
| **Active Exceptions** | Calculated Count | `int64` | `0 None` or `1 Active` (Red) | Active Unresolved Exceptions in DB |

---

## 9. Carrier Tracking

Carrier tracking in LogisticsHQ is architected through a multi-provider gateway that decouples carrier-specific formats from internal operational state:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       CARRIER TRACKING PIPELINE                             │
└─────────────────────────────────────┬───────────────────────────────────────┘
                                      │
               ┌──────────────────────┼──────────────────────┐
               ▼                      ▼                      ▼
      [Carrier REST API]      [Carrier Webhook]      [Background Poller]
       (DCSA Standards)       (HMAC Validated)       (Status-Aware Loop)
               │                      │                      │
               └──────────────────────┼──────────────────────┘
                                      ▼
                        [Carrier Adapter Registry]
                     (Maersk, MSC, CMA CGM, Hapag)
                                      ▼
                      [Event Normalization Engine]
                    (Standardized DCSA Milestone)
                                      ▼
                     [Deduplication & Transaction]
                    (carrier_tracking_events table)
                                      ▼
              ┌───────────────────────┴───────────────────────┐
              ▼                                               ▼
   [Shipment Master Update]                       [Milestone Update Engine]
   (Status, ETA, ETD, Vessel)                     (shipment_milestones table)
              │                                               │
              └───────────────────────┬───────────────────────┘
                                      ▼
                         [Universal Event Mesh Bus]
                        ('shipment.milestone_updated')
                                      ▼
                            [React Operations UI]
```

### Distinction of Tracking Data Layers:
1. **Live Carrier Telemetry**: Factual events transmitted directly by the ocean shipping line (e.g., container gate-in at terminal, vessel departure confirmed by terminal operating system).
2. **Locally Stored Telemetry**: Authoritative MariaDB records stored in `carrier_tracking_events` and `shipment_tracking_positions` providing historical auditability.
3. **AI Predictions**: Statistically modeled delay risks and projected arrival windows produced by the Python AI Sidecar. **AI predictions are never written to the master carrier ETA fields**; they are displayed in dedicated intelligence sections.

---

## 10. Tracking References

LogisticsHQ utilizes four industry-standard tracking references to correlate inbound tracking events with active shipments:

```
Priority 1: Booking Reference   (e.g., BKG-1789237642119)
Priority 2: Container Number    (e.g., MSKU9012345 - ISO 6346 Standard)
Priority 3: Master B/L (MBL)    (e.g., MAEU123456789 - Ocean Carrier Bill)
Priority 4: House B/L (HBL)     (e.g., LH-HBL-2601 - Forwarder House Bill)
```

### Correlation & Matching Logic:
In [HandleInboundCarrierEvent](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/bl.go#L783):
- The system first attempts to match the incoming event by **Booking Number**.
- If unmatched, it evaluates the **Container Number** against the JSON array `shipments.container_numbers`.
- If unmatched, it searches the **MBL Number**.
- If unmatched, it searches the **HBL Number**.
- If exactly one shipment matches, `matching_status = 'MATCHED'` and the event proceeds to processing.
- If multiple shipments match (e.g., a shared booking reference), `matching_status = 'AMBIGUOUS'` and the event is held for human review.
- If no shipments match, `matching_status = 'UNMATCHED'` and the event is safely held without data loss.

---

## 11. Milestones

Shipment milestones in LogisticsHQ represent chronological checkpoints along the international transport corridor:

| Milestone Code | Standard Description | Typical Source | Status Progression | Authoritative vs Predicted |
| :--- | :--- | :--- | :--- | :--- |
| `BOOKED` | Booking confirmed by shipping line | Carrier EDI 301 / API | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Booking Confirmation) |
| `GATE_IN` | Container received at origin terminal gate | Terminal TOS / EDI 315 | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Terminal Gate-In) |
| `LOADED` | Container loaded onto ocean vessel | Terminal Operating System | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Stowage Verification) |
| `DEPARTED` | Vessel departed origin port berth | Harbor Master / Carrier API | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Vessel Outward Pilot) |
| `IN_TRANSIT` | Vessel underway on maritime passage | AIS Satellite Telemetry | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Real-Time GPS / AIS) |
| `ARRIVED` | Vessel berthed at destination port | Port Authority / Carrier API | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (First Line Secured) |
| `DISCHARGED` | Container discharged from vessel to yard | Terminal Crane Telemetry | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Terminal Stevedore Log) |
| `CUSTOMS` | Import customs clearance completed | Customs Authority / Broker | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Customs Entry Status) |
| `DELIVERED` | Container delivered to consignee facility | Consignee Delivery Order / POD | `PLANNED` &rarr; `COMPLETED` | **Authoritative Fact** (Signed Proof of Delivery) |

---

## 12. ETA / ETD / ATD / ATA

LogisticsHQ strictly distinguishes between **Operational Facts** and **AI Predictions**:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         DATE & SCHEDULE ARCHITECTURE                        │
├──────────────────────────────────────┬──────────────────────────────────────┤
│ OPERATIONAL FACTS (Authoritative)    │ PREDICTIVE INTELLIGENCE (AI Models)  │
├──────────────────────────────────────┼──────────────────────────────────────┤
│ • Planned ETD: Contractual departure │ • Predicted Arrival Window: Machine  │
│ • Actual ATD: Verified vessel wheel  │   learning forecast (e.g. +48h drift)│
│ • Planned ETA: Carrier schedule date │ • Delay Risk Classification: Low,    │
│ • Actual ATA: Verified port berthing │   Medium, High, Critical             │
│ • Stored in: `shipments` master table│ • Stored in: `ai_predictions` table  │
└──────────────────────────────────────┴──────────────────────────────────────┘
```

> [!CAUTION]
> **Strict Semantic Boundary**: Under no circumstances does the Python AI Sidecar update `shipments.eta` or `shipments.etd`. The master fields represent legal, contractual baselines established by the carrier. AI predictions are presented in the UI under the **Predictive ETA & Schedule Variance** section with explicit confidence scores (e.g., `90% (HIGH)`) and source citations.

---

## 13. Tracking Webhooks

LogisticsHQ provides an enterprise webhook receiver for ocean carrier tracking push notifications:

```
[Carrier Push Request]
          │
          ▼
[POST /webhooks/carriers/{carrier}/{integration_id}]
          │
          ├──> 1. Authentication: Validates HMAC-SHA256 signature against carrier secret
          │
          ├──> 2. Tenant Resolution: Extracts `org_id` from path or `X-Integration-ID` header
          │
          ├──> 3. Payload Parsing: Unmarshals carrier-specific JSON/XML (DCSA compliant)
          │
          ├──> 4. Idempotency Check: Verifies `event_id` uniqueness in `carrier_tracking_events`
          │
          ├──> 5. Shipment Correlation: Matches Booking/Container/MBL/HBL
          │
          ├──> 6. Master Sync: Cascades milestone completion and updates `shipments` table
          │
          └──> 7. Audit & Event: Logs activity and publishes `shipment.milestone_updated`
```

### Technical Webhook Specifications:
- **Endpoints**:
  - `POST /webhooks/carriers/{carrier}` (General carrier webhook).
  - `POST /webhooks/carriers/{carrier}/{integration_id}` (Tenant-qualified webhook).
- **Security Validation**: Verified via `adapter.VerifyWebhookSignature(req.Body, req.Headers)` using carrier-configured pre-shared secret keys.
- **Replay Protection**: Database unique constraint on `carrier_tracking_events.event_id` prevents duplicate processing.

---

## 14. Tracking Polling

For carriers that do not support webhooks, LogisticsHQ runs two status-aware polling mechanisms:

1. **TrackingRefreshScheduler** ([tracking_scheduler.go](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/tracking_scheduler.go#L16)):
   - **Trigger**: Automatic background ticker running every 60 seconds.
   - **Frequency by Status**:
     - `BOOKED`: Polls every 360 minutes (6 hours).
     - `DEPARTED`: Polls every 120 minutes (2 hours).
     - `IN_TRANSIT`: Polls every 60 minutes (1 hour).
     - `ARRIVED`: Polls every 180 minutes (3 hours).
   - **Concurrency Control**: Thread-safe `sync.Map` prevents overlapping queries for the same shipment.

2. **Distributed CarrierPoller** ([carrier_poller.go](file:///c:/Users/Sai/go/src/freel-project/backend/internal/jobs/carrier_poller.go#L14)):
   - **Database Row Leasing**: Uses transactional `SELECT ... FOR UPDATE` to lock eligible shipments and prevent multi-pod worker collisions in clustered environments.

---

## 15. Exceptions

An **Exception** is any unplanned event, operational delay, or documentation blocker that threatens the timely delivery of cargo:

```
[Carrier Event / Metric Drift / Manual Detection]
                        │
                        ▼
            [Exception Evaluation]
                        │
            ┌───────────┴───────────┐
            ▼                       ▼
     [FACTUAL EXCEPTION]     [PREDICTED DISRUPTION]
     • Customs hold issued   • Weather front detected
     • Vessel rolled         • Congestion trending
     • Schedule slipped >24h • High dwell risk
            │                       │
            ▼                       ▼
   [Create Exception]       [AI Intelligence Alert]
 (shipment_exceptions)      (Predictive Banner)
            │                       │
            ▼                       ▼
   [Operations Triage]     [Proactive Review]
```

### Verified Exception Categories:
- `SCHEDULE_DELAY`: General operational schedule delay exceeding allowable threshold.
- `ETD_DELAY`: Delayed departure from origin port.
- `ETA_DELAY`: Delayed arrival at destination port.
- `VESSEL_ROLLOVER`: Container bumped from booked vessel to subsequent voyage.
- `PORT_CONGESTION`: Vessel waiting at anchor due to terminal berth unavailability.
- `CUSTOMS_HOLD`: Regulatory or customs clearance hold on cargo or documentation.
- `DOCUMENT_ISSUE`: Missing, rejected, or discrepant shipping documents.
- `CARRIER_DELAY`: Carrier-initiated equipment, mechanical, or routing hold.
- `ROUTE_DEVIATION`: Vessel departed authorized maritime corridor.
- `CONTAINER_ISSUE`: Damaged container, seal failure, or temperature deviation.
- `OTHER`: Unclassified operational exception.

---

## 16. AI Shipment Intelligence

The LogisticsHQ Shipments module integrates specialized predictive AI models via the Python AI Sidecar:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      AI SHIPMENT INTELLIGENCE SUITE                         │
├────────────────────────┬─────────────────────────┬──────────────────────────┤
│ MODEL                  │ INPUT SIGNALS           │ OUTPUT ARTIFACT          │
├────────────────────────┼─────────────────────────┼──────────────────────────┤
│ 1. Predictive ETA &    │ Vessel speed, heading,  │ Projected arrival window │
│    Delay Forecasting   │ waypoints, port dwell   │ (+48h variance, 90% conf)│
├────────────────────────┼─────────────────────────┼──────────────────────────┤
│ 2. Disruption &        │ Active exceptions, AIS  │ Early warning signal,    │
│    Rollover Prediction │ congestion, weather     │ proactive mitigation rec │
├────────────────────────┼─────────────────────────┼──────────────────────────┤
│ 3. Document Readiness  │ Mandatory doc checklist,│ Readiness score, storage │
│    & Compliance Risk   │ cutoffs, OCR extraction │ charge risk classification│
├────────────────────────┼─────────────────────────┼──────────────────────────┤
│ 4. Carrier Reliability │ Historical on-time rate,│ Carrier reliability score│
│    Scorecard           │ delay variance, SCAC    │ (Tier: EXCELLENT / FAIR) │
└────────────────────────┴─────────────────────────┴──────────────────────────┘
```

---

## 17. Python / Go Boundary

LogisticsHQ enforces an uncompromised architectural separation between Go and Python:

```
┌──────────────────────────────────────┐        ┌──────────────────────────────────────┐
│           GO BACKEND ENGINE          │        │          PYTHON AI SIDECAR           │
├──────────────────────────────────────┤        ├──────────────────────────────────────┤
│ • Authentication & JWT validation    │        │ • Natural Language Processing        │
│ • Role-Based Access Control (RBAC)   │        │ • Delay & disruption predictions     │
│ • Multi-tenant isolation enforcement │  HTTP  │ • Document OCR text reasoning        │
│ • MariaDB transactional mutations    │◄──────►│ • Communication drafting             │
│ • Carrier API & webhook integrations │ Internal│ • Zero database access              │
│ • Action System execution engine     │ Token  │ • Zero carrier API access            │
│ • Universal audit logging            │        │ • Zero direct state mutation         │
└──────────────────────────────────────┘        └──────────────────────────────────────┘
```

---

## 18. Action System

Shipment actions execute through the unified, permission-governed Action System:

```
User / AI Copilot
       │
       ▼
[Action Registry] ─── (Validates input against Schema)
       │
       ▼
[RBAC Enforcement] ── (Verifies SHIPMENTS:UPDATE permission)
       │
       ▼
[Action Execution]
  ├── 'shipments.get': Retrieves shipment detail
  ├── 'shipments.update_milestone': Updates tracking milestone code & timestamp
  ├── 'shipments.create_exception': Logs operational exception
  ├── 'shipments.update_eta': Updates estimated arrival schedule
  └── 'shipments.refresh_tracking': Forces live carrier sync
       │
       ▼
[Audit Logging] ──── (Records actor, action, timestamp in `activities`)
```

---

## 19. Approvals

Certain high-impact operational and financial actions require explicit Human-in-the-Loop (HITL) approval:

```
[Operational Event]
       │
       ▼
[AI Disruption Recommendation / Financial Review]
       │
       ▼
[Action Requires Approval?] ──(Threshold Check: e.g. Customer Notice, Billing Credit)
       │
       ├──► NO  ──► Direct Execution
       │
       └──► YES ──► [Create Approval Request] (`approvals` table)
                          │
                          ▼
                    [Operations Manager Review]
                          │
                    ┌─────┴─────┐
                    ▼           ▼
               [APPROVED]   [REJECTED]
                    │           │
                    ▼           ▼
               [Execution]  [Dismissed]
```

---

## 20. Notifications

LogisticsHQ triggers auditable notifications upon significant shipment events:

```
Shipment Event (Milestone / Exception / Delay)
       │
       ▼
[Notification Engine]
       │
       ├──► In-App Notification Center: Bell alert in TopBar navigation
       │
       ├──► Email Notification: Sent via SES email integration (if configured)
       │
       └──► Customer Outreach: Supervised communication draft via Copilot
```

---

## 21. Event Mesh / Automation

The internal Event Bus distributes shipment state changes across platform subsystems:

```
┌────────────────────────────┬─────────────────────────────┬──────────────────────────┐
│ EVENT TYPE                 │ PAYLOAD                     │ SUBSCRIBERS              │
├────────────────────────────┼─────────────────────────────┼──────────────────────────┤
│ `shipment.created`         │ `shipment_id`, `rfq_id`,    │ Document Compliance,     │
│                            │ `org_id`                    │ Tracking Scheduler       │
├────────────────────────────┼─────────────────────────────┼──────────────────────────┤
│ `shipment.milestone_updated│ `shipment_id`, `org_id`,    │ Operations Copilot,      │
│                            │ `milestone_code`, `status`  │ Notification Service     │
├────────────────────────────┼─────────────────────────────┼──────────────────────────┤
│ `shipment.exception_raised`│ `shipment_id`, `org_id`,    │ Command Center,          │
│                            │ `exception_type`, `severity`│ Alert Engine             │
└────────────────────────────┴─────────────────────────────┴──────────────────────────┘
```

---

## 22. Database Table Mapping

The Shipments module is backed by 10 specialized MariaDB tables:

| Table Name | Business Purpose | Key Columns | Related Tables |
| :--- | :--- | :--- | :--- |
| `shipments` | Master shipment execution record | `id`, `org_id`, `rfq_id`, `quote_id`, `booking_id`, `carrier_scac`, `booking_number`, `mbl_number`, `status`, `origin_port`, `destination_port`, `container_numbers`, `etd`, `eta`, `closure_status` | `bookings`, `rfqs`, `rfq_quotes`, `customers` |
| `shipment_milestones` | Milestone checkpoints and completion history | `id`, `shipment_id`, `milestone_code`, `description`, `planned_date`, `actual_date`, `status`, `location`, `source_event_id` | `shipments` |
| `shipment_exceptions` | Operational blockers and disruption records | `id`, `org_id`, `shipment_id`, `exception_type`, `severity`, `status`, `title`, `resolved`, `resolved_at`, `resolved_by`, `resolution_notes` | `shipments`, `users` |
| `shipment_documents` | Attached shipping documents & compliance metadata | `id`, `org_id`, `shipment_id`, `doc_type`, `category`, `s3_key`, `file_name`, `status`, `uploaded_by`, `reviewed_by` | `shipments` |
| `shipment_financial_charges`| Operational line item costs and revenues | `id`, `org_id`, `shipment_id`, `booking_id`, `category`, `charge_type`, `estimated_amount`, `actual_amount`, `currency`, `status` | `shipments`, `bookings` |
| `shipment_financial_summaries`| Aggregated profitability and margin metrics | `id`, `org_id`, `shipment_id`, `estimated_revenue`, `actual_revenue`, `estimated_cost`, `actual_cost`, `actual_margin`, `financial_status` | `shipments` |
| `shipment_tracking_positions`| Real-time vessel telemetry and coordinates | `id`, `org_id`, `shipment_id`, `vessel_name`, `latitude`, `longitude`, `speed_knots`, `heading_degrees`, `recorded_at` | `shipments` |
| `shipment_tracking_alerts` | Persistent operational alerts and monitoring | `id`, `org_id`, `shipment_id`, `alert_key`, `alert_type`, `severity`, `title`, `status`, `first_detected_at` | `shipments` |
| `shipment_tracking_refresh_runs`| Audit log of carrier telemetry sync executions | `id`, `org_id`, `shipment_id`, `provider_name`, `trigger_type`, `status`, `started_at`, `completed_at`, `new_events` | `shipments` |
| `carrier_tracking_events` | Ingested raw carrier tracking events | `id`, `org_id`, `event_id`, `carrier_scac`, `booking_number`, `container_number`, `milestone_code`, `event_time`, `matching_status` | `shipments` |

---

## 23. Shipment Relationship Map

```
Shipment (SH-280)
├── Customer: Derived via linked RFQ (Apex Global Logistics Corp)
├── Booking: Upstream space reservation (BKG-1789237642119 / ID #9236)
├── RFQ: Commercial customer inquiry (RFQ-20260912-235722-019 / ID #9138)
├── Quotation: Quoted baseline pricing (Quote #9153)
├── Carrier: Ocean shipping line (MAEU - Maersk Line)
├── Vessel & Voyage: MAERSK MC-KINNEY MOLLER / Voyage 2604E
├── Cargo & Containers: Equipment manifest (["MSKU9012345", "MSKU9012346"])
├── Milestones: Progressive journey checkpoints (`shipment_milestones`)
├── Tracking Telemetry: Vessel position, speed, and heading (`shipment_tracking_positions`)
├── Exceptions: Active operational holds (`shipment_exceptions`)
├── Documents: Bills of Lading, Customs Declarations (`shipment_documents`)
├── Financials: Line-item costs, revenues, and gross margin (`shipment_financial_charges`)
├── AI Intelligence: Predictive arrival windows and disruption forecasts
└── Audit Trail: Immutable historical activity log (`activities`)
```

---

## 24. API Mapping

| Business Operation | Frontend Function | HTTP Method | Endpoint | Go Handler / Service | Database Tables | Result |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **List Shipments** | `shipmentService.getShipments` | `GET` | `/api/v1/shipments?workspace=true` | `endpoints.ListShipmentsEP` | `shipments` | Returns paginated shipments & KPIs |
| **Get Shipment Detail** | `shipmentService.getShipmentDetail`| `GET` | `/api/v1/shipments/{id}` | `endpoints.GetShipmentEP` | `shipments`, `shipment_milestones` | Returns full shipment detail |
| **Update Milestone** | `shipmentService.updateMilestone` | `PUT` | `/api/v1/shipments/{id}/milestones` | `endpoints.UpdateMilestoneEP` | `shipment_milestones`, `shipments` | Advances milestone & updates status |
| **Get Exceptions** | `shipmentService.getShipmentExceptions`| `GET` | `/api/v1/shipments/{id}/exceptions` | `endpoints.GetShipmentExceptionsEP`| `shipment_exceptions` | Returns operational exceptions |
| **Create Exception** | `shipmentService.createShipmentException`| `POST`| `/api/v1/shipments/{id}/exceptions` | `endpoints.CreateShipmentExceptionEP`| `shipment_exceptions` | Logs new exception |
| **Resolve Exception** | `shipmentService.resolveException` | `POST` | `/api/v1/shipments/{id}/exceptions/{exId}/resolve` | `endpoints.ResolveShipmentExceptionEP`| `shipment_exceptions`, `shipments` | Marks resolved & restores status |
| **Sync Tracking** | `shipmentService.refreshShipmentTracking`| `POST` | `/api/v1/shipments/{id}/tracking/refresh` | `endpoints.RefreshShipmentTrackingEP`| `carrier_tracking_events`, `shipments` | Polls carrier API & synchronizes events |
| **Get Tracking Intel** | `shipmentService.getTrackingIntelligence`| `GET` | `/api/v1/shipments/{id}/tracking/intelligence` | `endpoints.GetShipmentTrackingIntelligenceEP`| Computed Telemetry | Returns journey progress & alerts |
| **Get Documents** | `shipmentService.getShipmentDocuments` | `GET` | `/api/v1/shipments/{id}/documents` | `endpoints.GetShipmentDocumentsEP` | `shipment_documents` | Returns docs & compliance summary |
| **Get Financials** | `shipmentService.getShipmentFinancials`| `GET` | `/api/v1/shipments/{id}/financials` | `endpoints.GetShipmentFinancialsEP`| `shipment_financial_summaries` | Returns revenue, cost, & gross margin |
| **Inbound Carrier Webhook**| External Carrier System | `POST` | `/webhooks/carriers/{carrier}/{integration_id}` | `endpoints.InboundWebhookEP` | `carrier_tracking_events` | Ingests carrier push updates |

---

## 25. Frontend Component Map

The Shipments frontend is built with modular React components located in [frontend/src/pages/dashboard/Shipments/](file:///c:/Users/Sai/go/src/freel-project/frontend/src/pages/dashboard/Shipments/):

```
Shipments Frontend Structure:
├── ShipmentsPage.jsx          (Main Workspace: KPI cards, search, filters, data table)
├── ShipmentDetail.jsx         (Comprehensive 360-degree shipment execution view)
├── DocumentWorkspace.jsx      (Document compliance checklists, OCR extraction, discrepancy resolution)
├── FinanceWorkspace.jsx       (Carrier cost line items, revenue tracking, margin analysis)
├── BillingWorkspace.jsx       (Customer invoice generation and operational shipment closure)
├── CustomDropdown.jsx         (Specialized dropdown selector for status and filters)
├── components/
│   ├── AutonomousExceptionLifecycleCard.jsx (Interactive exception triage & resolution)
│   ├── AutonomousShipmentLifecycleCard.jsx  (Milestone progression tracker)
│   ├── ShipmentOperationsAutomationSection.jsx (Operations Copilot actions & risk analysis)
│   └── ShipmentOperationsIntelligenceSection.jsx (Phase 4 predictive delay & schedule variance)
└── Shipments.css              (Vanilla CSS styling system)
```

---

## 26. Go Backend Component Map

The backend is architected in [backend/internal/shipments/](file:///c:/Users/Sai/go/src/freel-project/backend/internal/shipments/):

- **`transport.go`**: Go-Kit HTTP server transport, decoding, and route registration.
- **`endpoints.go`**: Go-Kit endpoint handlers connecting HTTP requests to domain business logic.
- **`bl.go`**: Core business logic service implementing status cascades, milestone rules, and exception triage.
- **`dl.go`**: Data access layer executing parameterized MariaDB SQL queries with tenant filtering.
- **`carrier_tracking_engine.go`**: Carrier adapter coordinator implementing DCSA event normalization.
- **`tracking_scheduler.go`**: Status-aware background polling engine.
- **`spec/`**: Canonical domain types, requests, responses, and lifecycle constants.

---

## 27. Permissions / RBAC

Shipment operations are governed by the platform RBAC service using `ResourceShipments` (`SHIPMENTS`):

| Role | Allowed Actions | Capabilities |
| :--- | :--- | :--- |
| **SUPER_ADMIN** | `create`, `read`, `update`, `delete`, `manage` | Full administrative control across all shipments |
| **ADMIN** | `create`, `read`, `update`, `delete`, `manage` | Full organizational shipment management |
| **OPERATIONS / COORDINATOR** | `create`, `read`, `update` | Convert bookings, update milestones, resolve exceptions, upload docs |
| **FINANCE** | `read`, `update` (financial charges) | Review financials, recalculate margins, generate invoices |
| **SALES_REP / ACCOUNT_MGR**| `read` | View shipment status and ETA for customer inquiries |
| **CUSTOMER / VIEWER** | `read` (restricted scope) | View tracking progress for their organization's cargo |

---

## 28. Tenant Isolation

Multi-tenant isolation is strictly enforced at every architectural layer:
1. **JWT Authentication**: User context is verified by [RequireAuth](file:///c:/Users/Sai/go/src/freel-project/backend/internal/server/routes.go#L59), extracting `OrgID`.
2. **SQL Query Scoping**: Every shipment query includes `WHERE s.org_id = ?`.
3. **Cross-Tenant Verification**: Cross-tenant requests (e.g., Organization 2 requesting Shipment #1 belonging to Organization 1) return `HTTP 404 / Resource not found` (Code 1003).

---

## 29. Audit Trail

Every state change produces an immutable audit record in the `activities` table:
- `SHIPMENT_CREATED`: Initial creation from Booking or RFQ.
- `SHIPMENT_STATUS_UPDATED`: Progression from one lifecycle state to another.
- `SHIPMENT_TRACKING_REFRESHED`: Carrier API synchronization run.
- `SHIPMENT_EXCEPTION_CREATED`: Operational exception logged.
- `SHIPMENT_EXCEPTION_ACKNOWLEDGED`: Exception reviewed by coordinator.
- `SHIPMENT_EXCEPTION_RESOLVED`: Exception resolved with documented notes.
- `SHIPMENT_DOCUMENT_APPROVED`: Compliance officer approved shipping document.
- `SHIPMENT_COMPLETED`: Shipment closed operationally.

---

## 30. Search / Filter / Sort / Pagination

The Shipments list page provides comprehensive operational filtering:
- **Search Bar**: Debounced text search matching Shipment ID, Customer Name, Booking Number, MBL Reference, or Vessel Name.
- **Status Filter**: Dropdown filtering by `BOOKED`, `DEPARTED`, `IN_TRANSIT`, `ARRIVED`, `DELIVERED`, `EXCEPTION`.
- **Tracking State Filter**: Filter by `ON_TRACK`, `DELAYED`, `AT_RISK`.
- **Carrier Filter**: Filter by ocean line (e.g., `MAEU`, `MSCU`, `CMAU`).
- **Pagination**: Server-side pagination (`page`, `limit`, defaulting to 10 records per page).

---

## 31. Business User Journeys

### Journey 1: Standard Freight Movement
`Booking Confirmed` &rarr; `Convert to Shipment` &rarr; `Container Gate-In` &rarr; `Vessel Departure` &rarr; `Sea Transit` &rarr; `Port Arrival` &rarr; `Consignee Delivery` &rarr; `Close Shipment`.

### Journey 2: Carrier Webhook Ingestion & Milestone Advance
`Vessel Departs Port` &rarr; `Carrier Sends Webhook` &rarr; `HMAC Verified` &rarr; `Container Matched` &rarr; `DEPARTED Milestone Completed` &rarr; `Status Advances to DEPARTED` &rarr; `TopBar Notification Sent`.

### Journey 3: Disruption Detection & Exception Resolution
`Port Congestion Occurs` &rarr; `Delay Exceeds 24h` &rarr; `Exception Created (PORT_CONGESTION)` &rarr; `Status Changes to EXCEPTION` &rarr; `Coordinator Reschedules Feeder` &rarr; `Exception Resolved` &rarr; `Status Restored to IN_TRANSIT`.

### Journey 4: Predictive AI Disruption & Human-in-the-Loop Action
`AIS Detects Severe Storm on Route` &rarr; `AI Sidecar Forecasts +48h Drift` &rarr; `Copilot Displays Predictive Banner` &rarr; `Coordinator Reviews Recommendation` &rarr; `Coordinator Clicks 'Request Action'` &rarr; `Proactive Consignee Notice Drafted`.

---

## 32. Data Flow Diagrams

```
[Carrier Network / AIS Satellite]
               │
               ▼
[LogisticsHQ Gateway / Poller]
               │
               ▼
    [Go Backend Service] ◄───► [MariaDB (Authoritative Storage)]
               │
               ├───► [Python AI Sidecar] (Stateless Predictions)
               │
               ▼
   [React Operations Workspace]
```

---

## 33. Source-of-Truth Matrix

| Information Element | Authoritative Source | Secondary / Derived Source | AI Role |
| :--- | :--- | :--- | :--- |
| **Master Shipment Status** | MariaDB `shipments.status` | Milestone Cascade | None (Deterministic) |
| **Carrier Milestone Event** | Carrier API / Webhook | `shipment_milestones` table | None (Factual) |
| **Vessel Position (Lat/Lon)**| AIS Feed / Carrier API | `shipment_tracking_positions` | None (Factual) |
| **Master Contractual ETA** | Carrier Schedule | `shipments.eta` table field | None (Authoritative) |
| **Predicted Arrival Window** | Python AI Sidecar | `ai_predictions` table | **Full Machine Learning** |
| **Exception Record** | MariaDB `shipment_exceptions` | Operations Triage UI | Predictive Disruption Signal |
| **Financial Margin** | `shipment_financial_summaries` | Calculated Line Items | None (Mathematical) |

---

## 34. Error / Loading / Empty States

The module implements resilient UI states:
- **Empty State**: When no shipments exist, a clean empty state card prompts the user to create a new shipment or import from active bookings.
- **Loading State**: Skeleton loaders shimmer across KPI metrics and table rows during network requests.
- **Search Not Found**: When search filters return zero results, a dedicated "No shipments matching criteria" alert appears with a "Clear Filters" button.
- **Missing Upstream Booking**: If a shipment was created manually without a booking, the UI gracefully displays `Unassigned` instead of crashing.
- **Carrier Disconnected**: When a carrier API is unconfigured, the UI displays a helpful notice: *"Carrier integration is not configured. Showing latest persisted operational tracking data."*

---

## 35. UI / UX Observations

### What Works Well:
- **Comprehensive 360° Detail View**: Consolidates overview metrics, route corridors, equipment manifests, milestones, exceptions, documents, and financials into a single unified screen.
- **Visual Schedule Variance**: Color-coded red variance tags (`+171.8d late`) clearly indicate delayed movements.
- **Direct Upstream Navigation**: One-click links directly navigate back to the associated Booking (`BKG-1789237642119`) or RFQ (`RFQ-20260912-235722-019`).

### Opportunities for Improvement (Classified for Task 3.7):
- **MUST FIX (Functional Remediation)**:
  - Customer name appears blank on some converted shipments where RFQ customer relationship was indirect.
  - Exception resolution notes should support multi-line rich text for complex insurance claims.
- **SHOULD IMPROVE (UI Polish)**:
  - Add quick filter chips for `Delayed (>24h)` on the main list view.
  - Provide an interactive route corridor map using Mapbox/Leaflet coordinates.
- **OPTIONAL (Enhancements)**:
  - Export shipment tracking reports directly to branded PDF for cargo owners.

---

## 36. Responsive / Zoom Observations

Verified across five browser zoom levels and three desktop resolutions:

| Viewport / Zoom | Visual Layout Stability | Observations & Findings |
| :--- | :--- | :--- |
| **100% Zoom (1440x900)** | Excellent | Full 6-card KPI row, complete table columns, no horizontal scroll. |
| **80% Zoom** | Excellent | Compact information density; all sections cleanly visible. |
| **90% Zoom** | Excellent | Standard high-density display; fonts and badges crisp. |
| **110% Zoom** | Good | Table maintains proper padding; KPI cards wrap cleanly. |
| **125% Zoom** | Good | Fluid flexbox wrapping; sidebar collapses gracefully without clipping. |
| **1280x720 (Laptop)** | Excellent | KPI grid neatly wraps into 2 rows of 3 cards; table remains fully legible. |
| **1366x768 (Standard)** | Excellent | Balanced layout with full table visibility. |

---

## 37. Security Architecture

- **Authentication**: Stateless JWT bearer tokens with standard expiration and refresh flow.
- **Tenant Isolation**: Mandatory `org_id` foreign keys enforced on all SELECT, INSERT, UPDATE queries. Cross-tenant access attempts return HTTP 404.
- **Webhook Security**: Pre-shared HMAC-SHA256 signatures validated before webhook payload deserialization.
- **AI Prompt Hardening**: Dedicated regex filters in Python AI Sidecar sanitize inputs against prompt injection attacks attempting to bypass manager approvals.

---

## 38. Business + Technical Glossary

- **SCAC**: Standard Carrier Alpha Code (e.g., `MAEU` for Maersk, `MSCU` for MSC).
- **MBL**: Master Bill of Lading issued by the ocean shipping line to the freight forwarder.
- **HBL**: House Bill of Lading issued by the freight forwarder to the actual cargo owner.
- **ETD / ATD**: Estimated Time of Departure / Actual Time of Departure.
- **ETA / ATA**: Estimated Time of Arrival / Actual Time of Arrival.
- **POD**: Proof of Delivery signed by the cargo consignee.
- **DCSA**: Digital Container Shipping Association (international standards for container tracking events).
- **AIS**: Automatic Identification System (satellite and terrestrial vessel transponder telemetry).
- **Demurrage**: Fee charged by shipping lines when containers remain inside port terminals beyond agreed free days.
- **Detention**: Fee charged when containers are held outside port terminals beyond agreed free days.

---

## 39. One-Page “How Shipments Work”

```
                       HOW SHIPMENTS WORK IN LOGISTICSHQ
                       
1. INITIALIZATION:
   Booking Confirmed ──► Convert to Shipment ──► Shipment Created with status 'BOOKED'
                                             ──► 5 Standard Milestones Seeded
                                             
2. CARRIER TRACKING:
   Ocean Vessel Moves ──► Carrier API / Webhook / Poller ──► DCSA Normalization
                                                         ──► Position & Milestone Update
                                                         ──► Status Advances (DEPARTED ➔ IN_TRANSIT)
                                                         
3. EXCEPTION HANDLING:
   Port Congestion / Customs Hold ──► Exception Created ──► Status Flagged (EXCEPTION)
                                  ──► Coordinator Mitigates ──► Exception RESOLVED
                                  ──► Status Restored to Latest Milestone
                                  
4. PREDICTIVE AI COPILOT:
   Satellite AIS Feed ──► Python AI Model ──► Delay Window Forecast (+48h)
                      ──► Copilot Displays Recommendation
                      ──► Human Approves ──► Proactive Customer Update Sent
                      
5. DELIVERY & CLOSURE:
   Vessel Berths ──► Container Discharged ──► Customs Cleared ──► Consignee Delivery (POD)
                 ──► Status: DELIVERED ──► Financial Audit ──► Shipment CLOSED
```

---

## 40. Technical Traceability Matrix

| Shipment Capability | Frontend Component | API Endpoint | Go Service Function | Python AI Component | Database Table | RBAC Permission |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Workspace Listing** | `ShipmentsPage.jsx` | `GET /api/v1/shipments` | `svc.GetShipmentsWorkspace` | N/A | `shipments` | `SHIPMENTS:READ` |
| **Shipment Detail** | `ShipmentDetail.jsx` | `GET /api/v1/shipments/{id}` | `svc.GetShipmentByID` | N/A | `shipments`, `shipment_milestones` | `SHIPMENTS:READ` |
| **Milestone Update** | `ManualUpdatePanel.jsx` | `PUT /api/v1/shipments/{id}/milestones` | `svc.UpdateMilestone` | N/A | `shipment_milestones`, `shipments` | `SHIPMENTS:UPDATE` |
| **Exception Logging** | `ShipmentDetail.jsx` | `POST /api/v1/shipments/{id}/exceptions` | `svc.CreateShipmentException` | N/A | `shipment_exceptions` | `SHIPMENTS:CREATE` |
| **Exception Triage** | `AutonomousExceptionLifecycleCard.jsx` | `POST /.../resolve` | `svc.ResolveShipmentException` | N/A | `shipment_exceptions` | `SHIPMENTS:UPDATE` |
| **Telemetry Sync** | `ShipmentDetail.jsx` | `POST /api/v1/shipments/{id}/tracking/refresh`| `engine.SyncShipmentTracking` | N/A | `carrier_tracking_events` | `SHIPMENTS:UPDATE` |
| **Tracking Intelligence** | `ShipmentOperationsIntelligenceSection.jsx` | `GET /.../tracking/intelligence` | `svc.GetShipmentTrackingIntelligence`| N/A | Computed Telemetry | `SHIPMENTS:READ` |
| **Predictive Delay** | `ShipmentOperationsIntelligenceSection.jsx` | Internal Sidecar Proxy | Proxy to Sidecar | `predictions.predict_shipment_delay` | `ai_predictions` | `SHIPMENTS:READ` |
| **Document Compliance** | `DocumentWorkspace.jsx`| `GET /api/v1/shipments/{id}/documents` | `svc.GetShipmentDocuments` | `contract_compliance.agent` | `shipment_documents` | `SHIPMENTS:READ` |
| **Financial Costing** | `FinanceWorkspace.jsx` | `GET /api/v1/shipments/{id}/financials`| `svc.GetShipmentFinancials` | N/A | `shipment_financial_charges`| `SHIPMENTS:READ` |
| **Inbound Webhook** | N/A (External Push) | `POST /webhooks/carriers/...` | `endpoints.InboundWebhookEP` | N/A | `carrier_tracking_events` | Public HMAC |

---

## 41. Known Gaps

The following architectural and operational gaps were identified during inspection:

1. **Carrier API Integration Status (External Integration Gap)**:
   - Live external tracking calls fall back to cached telemetry when external carrier production credentials are not populated in the environment. This is an environment configuration state, not a software architecture defect.
2. **Interactive Route Corridor Map (UI/UX Gap)**:
   - Navigational waypoints and vessel coordinates are calculated in the backend, but the frontend currently displays them in tabular telemetry format rather than an interactive geospatial map canvas.
3. **Multi-Carrier Booking Consolidation (Workflow Enhancement)**:
   - Shipments currently track container equipment associated with a single primary ocean carrier SCAC. Multi-modal leg splitting (e.g., ocean carrier &rarr; rail ramp &rarr; drayage trucker) requires separate milestone notes.

---

## 42. Verification Status

| Subsystem | Inspection Method | Verified Behavior | Status |
| :--- | :--- | :--- | :--- |
| **Frontend Route** | React Router / Browser CDP | `/dashboard/shipments` and `/dashboard/shipments/{id}` load cleanly | **VERIFIED** |
| **Backend API** | Live HTTP Curl (`test-token`) | `GET /api/v1/shipments` returns paginated records and live KPIs | **VERIFIED** |
| **Tenant Isolation** | Cross-Tenant HTTP Curl | Org 2 cannot access Org 1 shipments (`HTTP 404 / Code 1003`) | **VERIFIED** |
| **Database Schema** | DDL and SQL Queries | 10 MariaDB tables verified with indexes and constraints | **VERIFIED** |
| **Milestone Engine** | Go BL Code Inspection | 5 canonical milestones seeded; monotonic progression enforced | **VERIFIED** |
| **Tracking Pipeline** | Go Adapter Inspection | DCSA event normalization, webhook HMAC validation, poller loop | **VERIFIED** |
| **Exception Flow** | Go Service & MariaDB | 11 exception types, 4 severities, acknowledge/resolve workflows | **VERIFIED** |
| **AI Boundary** | Python Sidecar Inspection | Python is strictly read-only / predictive; zero direct SQL mutations | **VERIFIED** |
| **Action System** | Go Action Registry | `shipments.*` actions defined with permissions and audit logging | **VERIFIED** |
| **Responsive UI** | Chrome CDP Automation | Verified across 80%-125% zoom and 1280x720 to 1440x900 viewports | **VERIFIED** |

---

## Final Status

**PASS — SHIPMENT WORKFLOW DOCUMENTED**
