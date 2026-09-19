# Task 3.13.A — Notifications Business Workflow, Notification Lifecycle, Channels, Providers, Delivery Status, AI Intelligence, Escalation Center, Database Mapping, API Traceability, Permissions, Audit, and Technical Architecture

## 1. Executive Architecture & Module Overview

The LogisticsHQ Notifications module serves as the central alerting, escalation, and delivery nerve center for the multi-tenant freight and logistics platform. It manages real-time operator alerts, operational notifications, AI-assisted escalations, and outbound communications across multiple delivery channels.

The architecture strictly adheres to the core system principles:
1. **Go Authoritative Execution Boundary**: All state changes, multi-tenant boundaries, RBAC authorization, rate limiting, database mutations, external provider dispatches (AWS SES and Twilio), and audit logging are owned exclusively by Go (`backend/internal/notifications/` and `backend/internal/integrations/`).
2. **Python AI Sidecar Intelligence Boundary**: AI capabilities (smart priority scoring, escalation reasoning, workload clustering, and operator communication drafts) are hosted exclusively in the Python sidecar (`ai_sidecar/` on port 8090). Python never executes database writes, never calls external notification providers directly, and never bypasses Go authorization.
3. **Multi-Tenant Isolation**: Every database query, cache key, and event payload is scoped strictly to `org_id`. Cross-tenant read, mutation, or webhook routing attempts fail safely with HTTP 404 or 403.
4. **Segregation & Truthful Delivery Reporting**: The module distinguishes between `NOTIFICATION CREATED`, `PENDING`, `DELIVERED`, and `FAILED/BOUNCED`. When external credentials (AWS SES or Twilio) are unconfigured, the system reports truthful `StatusNotConfigured` rather than fabricating success.
5. **Real-Time Operational Evaluation**: The evaluation engine (`engine.go`) continuously synthesizes operational data from shipments, exceptions, approvals, overdue invoices, and expiring contracts into actionable notifications with deduplication hashing (`dedup_hash`).

---

## 2. Notification Data Model & Database Mapping

The module persists its operational state across four primary MariaDB tables in `freel_mysql`:

### 2.1 Table: `notifications`
Stores the central catalog of in-app and cross-channel alerts.
```sql
CREATE TABLE `notifications` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `org_id` bigint(20) NOT NULL,
  `user_id` bigint(20) DEFAULT NULL,
  `role_target` varchar(50) DEFAULT NULL,
  `source_module` varchar(50) NOT NULL,
  `source_record_type` varchar(50) NOT NULL,
  `source_record_id` varchar(100) NOT NULL,
  `recommendation_id` bigint(20) DEFAULT NULL,
  `approval_id` bigint(20) DEFAULT NULL,
  `automation_execution_id` bigint(20) DEFAULT NULL,
  `notification_type` varchar(100) NOT NULL,
  `title` varchar(255) NOT NULL,
  `message` text NOT NULL,
  `severity` varchar(30) NOT NULL DEFAULT 'INFORMATIONAL',
  `priority` varchar(30) NOT NULL DEFAULT 'MEDIUM',
  `status` varchar(50) NOT NULL DEFAULT 'ACTIVE',
  `is_read` tinyint(1) NOT NULL DEFAULT 0,
  `read_at` datetime DEFAULT NULL,
  `is_dismissed` tinyint(1) NOT NULL DEFAULT 0,
  `dismissed_at` datetime DEFAULT NULL,
  `is_escalated` tinyint(1) NOT NULL DEFAULT 0,
  `escalation_level` int(11) NOT NULL DEFAULT 0,
  `escalated_at` datetime DEFAULT NULL,
  `action_required` tinyint(1) NOT NULL DEFAULT 0,
  `action_url` varchar(255) DEFAULT NULL,
  `dedup_hash` varchar(64) NOT NULL,
  `correlation_id` varchar(255) NOT NULL,
  `expires_at` datetime DEFAULT NULL,
  `metadata` longtext DEFAULT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `updated_at` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `delivery_status` varchar(32) NOT NULL DEFAULT 'DELIVERED',
  `is_acknowledged` tinyint(1) NOT NULL DEFAULT 0,
  `acknowledged_at` datetime DEFAULT NULL,
  `acknowledged_by` bigint(20) DEFAULT NULL,
  `is_snoozed` tinyint(1) NOT NULL DEFAULT 0,
  `snoozed_until` datetime DEFAULT NULL,
  `group_key` varchar(128) DEFAULT NULL,
  `ai_summary` text DEFAULT NULL,
  `ai_escalation_reason` text DEFAULT NULL,
  `ai_priority_score` float DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_org_user` (`org_id`,`user_id`),
  KEY `idx_org_read` (`org_id`,`is_read`),
  KEY `idx_org_severity` (`org_id`,`severity`),
  KEY `idx_dedup` (`org_id`,`dedup_hash`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2.2 Table: `notification_escalation_events`
Stores immutable audit history for manager escalations and AI-assisted workflows.
```sql
CREATE TABLE `notification_escalation_events` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT,
  `notification_id` bigint(20) NOT NULL,
  `org_id` bigint(20) NOT NULL,
  `escalation_level` int(11) NOT NULL DEFAULT 1,
  `escalation_reason` varchar(255) NOT NULL,
  `previous_severity` varchar(30) NOT NULL,
  `new_severity` varchar(30) NOT NULL,
  `trigger_type` varchar(50) NOT NULL DEFAULT 'TIME_THRESHOLD',
  `threshold_hours` int(11) NOT NULL DEFAULT 0,
  `acknowledged_at` datetime DEFAULT NULL,
  `acknowledged_by` bigint(20) DEFAULT NULL,
  `resolved_at` datetime DEFAULT NULL,
  `resolved_by` bigint(20) DEFAULT NULL,
  `correlation_id` varchar(255) NOT NULL,
  `created_at` datetime NOT NULL DEFAULT current_timestamp(),
  `ai_escalation_summary` text DEFAULT NULL,
  `recommended_action` varchar(128) DEFAULT NULL,
  `action_proposal_id` varchar(128) DEFAULT NULL,
  `approval_id` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_notif_esc` (`notification_id`),
  KEY `idx_org_esc` (`org_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 2.3 Tables: `user_notification_preferences` & `org_notification_preferences`
Controls per-user thresholds, category toggles, and organizational defaults.
- `min_severity`: Lowest severity that displays or alerts (`INFORMATIONAL`, `LOW`, `MEDIUM`, `HIGH`, `CRITICAL`).
- Category flags: `approvals_enabled`, `automations_enabled`, `recommendations_enabled`, `finance_enabled`, `operations_enabled`, `compliance_enabled`.

---

## 3. Real-Time Operational Evaluation Engine

The engine (`backend/internal/notifications/engine.go`) continuously evaluates live business entities into notifications on demand and via the Event Mesh.

### 3.1 Deduplication Hash Generation
To prevent alert fatigue and redundant notifications:
$$\text{dedup\_hash} = \text{SHA256}(\text{org\_id} : \text{module} : \text{record\_type} : \text{record\_id} : \text{notification\_type})$$
If a notification with this `dedup_hash` already exists in `ACTIVE` or `UNREAD` state for the organization, a duplicate record is skipped.

### 3.2 Evaluation Domains
1. **Approvals (`evaluateApprovals`)**:
   - Source: `approval_requests` where `status = 'Pending'`.
   - Age check: Pending $\ge$ 48 hours escalates severity to `CRITICAL`.
   - Action URL: `/dashboard/approvals?id={id}`.
2. **AI Recommendations (`evaluateRecommendations`)**:
   - Source: `ai_action_proposals` where priority is high or assigned to current user.
   - Action URL: `/dashboard/ai-actions?id={id}`.
3. **Automations (`evaluateAutomations`)**:
   - Source: `automation_executions` where `status = 'FAILED'`.
   - Action URL: `/dashboard/automation?execution_id={id}`.
4. **Shipments & Exceptions (`evaluateShipmentAlerts`)**:
   - Source: Active shipments with overdue ETA, temperature excursion exceptions, or port dwell.
   - Action URL: `/dashboard/shipments/{id}` or `/dashboard/exceptions`.
5. **Finance & Collections (`evaluateInvoices`)**:
   - Source: `customer_invoices` where payment status is overdue past payment terms.
   - Action URL: `/dashboard/finance/invoices?id={id}`.
6. **Contracts & Compliance (`evaluateContracts`)**:
   - Source: Contracts expiring within 30 days or pending document compliance verification.
   - Action URL: `/dashboard/contracts?id={id}`.

---

## 4. In-App Notification Lifecycle & Transitions

```
[Business Trigger / Event]
          │
          ▼
   [Evaluation Sweep] ──(Dedup Hash Match)──> [Skip / Deduped]
          │
          ▼
   [Notification Created (Unread, Active)]
          │
     ┌────┴───────────────────────────────┐
     ▼                                    ▼
[Mark Read]                         [Acknowledge]
 (is_read=1, read_at=NOW)            (is_acknowledged=1, acknowledged_by=UID)
     │                                    │
     ▼                                    ▼
[Mark Unread]                        [Snooze]
 (is_read=0, read_at=NULL)           (is_snoozed=1, snoozed_until=T+Δ)
     │                                    │
     ▼                                    ▼
[Dismiss]                            [Escalate with AI]
 (is_dismissed=1, status='DISMISSED') (is_escalated=1, escalation_level++, audit logged)
```

---

## 5. External Integration Gateway: AWS SES & Twilio

Outbound communications route through the Integration Gateway (`backend/internal/integrations/`):

### 5.1 AWS SES (Email)
- **Adapter**: `backend/internal/integrations/ses_provider.go`.
- **Validation**: Strict RFC 5322 regex validation via `ValidateEmailAddress()`. Email bodies checked against secret leakage.
- **Provider Status**: Returns `StatusNotConfigured` when credentials are absent. Does NOT fake successful delivery.
- **Webhooks**: `POST /api/v1/integrations/webhooks/ses` processes SNS delivery, bounce, and complaint notifications, verifying HMAC signatures and updating `delivery_status`.

### 5.2 Twilio (SMS)
- **Adapter**: `backend/internal/integrations/twilio_provider.go`.
- **Validation**: Strict E.164 phone format validation via `ValidateE164Phone()`. Message bodies validated $\le$ 1600 characters and scanned for credential tokens.
- **Provider Status**: Returns `StatusNotConfigured` when Account SID or Auth Token is absent.
- **Webhooks**: `POST /api/v1/integrations/webhooks/twilio` handles Twilio status callbacks with signature verification and replay protection.

---

## 6. AI Notification Intelligence Boundary

The Python AI sidecar (`ai_sidecar/` port 8090) provides analytical services invoked by Go over authenticated HTTP:

1. **Smart Prioritization (`POST /notifications-escalations/analyze-and-prioritize`)**:
   - Evaluates notification age, customer Tier, financial exposure, and SLA deadlines.
   - Outputs: `ai_priority_score` (0.0 to 1.0), `ai_escalation_reason`, `recommended_action`.
2. **Drafting Communications (`POST /notifications-escalations/generate-escalation-draft`)**:
   - Generates contextual operator communications: `INTERNAL_ESCALATION`, `CUSTOMER_ADVISORY`, or `CARRIER_ESCALATION`.
   - Outputs structured draft subject, body, recipients, and tone.
3. **Security Boundary**:
   - Calls require `X-LogisticsHQ-Service-Key`.
   - Python returns JSON; Go parses, sanitizes, and commits changes to MariaDB.
   - Python is strictly forbidden from directly initiating external SMS/Email or mutating DB records.

---

## 7. API Traceability Matrix

| Method | Path | Auth / Role | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/v1/notifications` | Bearer (All) | Filtered paginated list of notifications |
| `GET` | `/api/v1/notifications/stats` | Bearer (All) | Real-time counts (total, unread, action-required, critical) |
| `GET` | `/api/v1/notifications/unread-count` | Bearer (All) | Unread notification count for badge rendering |
| `GET` | `/api/v1/notifications/{id}` | Bearer (All) | Fetch single notification details |
| `POST` | `/api/v1/notifications/{id}/read` | Bearer (All) | Mark single notification as read |
| `POST` | `/api/v1/notifications/{id}/unread` | Bearer (All) | Mark single notification as unread |
| `POST` | `/api/v1/notifications/read-all` | Bearer (All) | Mark all notifications for org/user as read |
| `POST` | `/api/v1/notifications/{id}/dismiss` | Bearer (All) | Soft-dismiss notification from feed |
| `POST` | `/api/v1/notifications/{id}/acknowledge` | Bearer (Ops/Admin) | Operator explicit acknowledgement |
| `POST` | `/api/v1/notifications/{id}/snooze` | Bearer (Ops/Admin) | Snooze alert for specified minutes |
| `POST` | `/api/v1/notifications/{id}/escalate` | Bearer (Ops/Admin) | Trigger AI-assisted escalation event |
| `POST` | `/api/v1/notifications/{id}/analyze-ai` | Bearer (Ops/Admin) | Invoke AI priority & workload analysis |
| `POST` | `/api/v1/notifications/{id}/generate-draft`| Bearer (Ops/Admin) | Generate AI escalation communication draft |
| `GET` | `/api/v1/notifications/escalations` | Bearer (Admin) | List immutable escalation audit trail |
| `GET` | `/api/v1/notifications/preferences` | Bearer (All) | Get user notification preferences |
| `PUT` | `/api/v1/notifications/preferences` | Bearer (All) | Update user notification preferences |
| `POST` | `/api/v1/notifications/evaluate` | Bearer (Ops/Admin) | Trigger manual evaluation sweep across DB |
| `GET` | `/api/v1/integrations/status` | Bearer (Admin) | Masked health status of external providers |
| `POST` | `/api/v1/integrations/webhooks/twilio`| HMAC Signature | Inbound Twilio delivery status callback |
| `POST` | `/api/v1/integrations/webhooks/ses` | HMAC / Signature | Inbound AWS SES bounce/complaint webhook |

---

## 8. Role-Based Access Control & Multi-Tenant Isolation

- **Tenant Isolation**: Handled via `middleware.GetUserContext(ctx)`. Every repository query joins on `WHERE org_id = ?`. Cross-tenant requests produce HTTP 404.
- **RBAC Matrix**:
  - `Viewer / Read-Only`: Can view notifications and unread badge; cannot acknowledge, snooze, or escalate.
  - `Operator / Specialist`: Can mark read/unread, acknowledge, snooze, and run AI analysis.
  - `Manager / Admin`: Full access including manual evaluation sweeps, escalation reviews, integration health inspections, and organizational preference overrides.
