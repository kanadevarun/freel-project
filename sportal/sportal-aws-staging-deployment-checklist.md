# LogisticsHQ SPortal & Platform — AWS Staging Deployment Checklist

> **Target Environment:** AWS Staging  
> **Release Target:** SPortal & Platform Baseline Candidate  
> **Classification:** Internal DevOps & Cloud Engineering Execution Checklist  
> **Execution Rule:** Zero hardcoded secrets. Non-destructive database migrations only. External integrations run in sandbox/mock mode. Real customer emails/SMS must not be triggered.

---

## 1. AWS Staging Infrastructure Prerequisites

- [ ] **AWS Account & Region Verification**
  - Staging VPC provisioned in target region (e.g., `ap-south-1` or `us-east-1`).
  - Dedicated staging subnets:
    - 2x Public Subnets (for Application Load Balancer).
    - 2x Private Application Subnets (for ECS Fargate tasks / EC2 instances).
    - 2x Private Data Subnets (for Amazon RDS MariaDB).
  - Internet Gateway (IGW) attached to VPC.
  - NAT Gateway configured in public subnet for outbound private subnet traffic (package updates, external API webhooks).

- [ ] **Security Group Architecture**
  - `sg-staging-alb`: Inbound `80`, `443` from internet (`0.0.0.0/0`). Outbound to `sg-staging-backend` (port `8080`).
  - `sg-staging-backend`: Inbound `8080` only from `sg-staging-alb`. Inbound from staging worker/sidecar. Outbound to RDS (`3306`), S3 endpoints, SES/Twilio/Carrier endpoints.
  - `sg-staging-sidecar`: Inbound `8090` only from `sg-staging-backend`. Outbound to LLM providers via HTTPS (`443`). No direct inbound from internet.
  - `sg-staging-db`: Inbound `3306` only from `sg-staging-backend` and staging migration task runner.

---

## 2. DNS & SSL / TLS Certificates (AWS Route 53 & ACM)

- [ ] **AWS Certificate Manager (ACM)**
  - Request / validate wildcard or SAN SSL certificate covering:
    - `staging.logisticshq.in`
    - `staging-app.logisticshq.in` (CPortal Staging)
    - `staging-sportal.logisticshq.in` (SPortal Staging)
    - `staging-api.logisticshq.in` (Go Backend API Staging)
  - Verify certificate status is `ISSUED`.

- [ ] **Route 53 DNS Records**
  - `staging-sportal.logisticshq.in` -> CloudFront distribution or ALB CNAME.
  - `staging-app.logisticshq.in` -> CloudFront distribution or ALB CNAME.
  - `staging-api.logisticshq.in` -> ALB DNS name (Alias A / AAAA record).
  - Public website (`staging.logisticshq.in`) -> CloudFront / static origin.

---

## 3. Secret Management (AWS Secrets Manager / SSM Parameter Store)

- [ ] **Database Secrets (`/staging/freel/db`)**
  - `DB_USER`: `freel_staging_user`
  - `DB_PASSWORD`: Securely generated 32+ character random string.
  - `DB_HOST`: RDS MariaDB cluster/instance endpoint.
  - `DB_PORT`: `3306`
  - `DB_NAME`: `freel_staging`

- [ ] **Authentication & Token Secrets (`/staging/freel/auth`)**
  - `JWT_SECRET`: 64-byte high-entropy string.
  - `INTERNAL_SERVICE_TOKEN`: 64-byte high-entropy token for Go <-> Sidecar / Worker M2M communication.

- [ ] **AWS Service Integrations (`/staging/freel/aws`)**
  - IAM Roles for Tasks (IRSA / ECS Task Roles) preferred over static IAM credentials.
  - S3 Bucket Name: `freel-documents-staging-ap-south-1`.
  - AWS Region: e.g., `ap-south-1`.

- [ ] **External Provider Secrets (Sandbox Credentials) (`/staging/freel/integrations`)**
  - `STRIPE_SECRET_KEY`: `sk_test_...` (Stripe test mode).
  - `STRIPE_WEBHOOK_SECRET`: `whsec_...`
  - `TWILIO_ACCOUNT_SID`: Test subaccount SID or sandbox.
  - `TWILIO_AUTH_TOKEN`: Test auth token.
  - `TWILIO_PHONE_NUMBER`: Test Twilio phone number.
  - `CARRIER_API_KEY`: Sandbox carrier testing key.

- [ ] **AI Provider Secrets (`/staging/freel/ai`)**
  - `OPENAI_API_KEY`: API key for staging AI inference.
  - `TAVILY_API_KEY`: API key for search grounding (if enabled).

---

## 4. Database Provisioning & Canonical Migrations

- [ ] **RDS MariaDB Provisioning**
  - Engine: MariaDB 10.6+ or 10.11 LTS.
  - Multi-AZ: Enabled for high availability.
  - Storage Encryption: AWS KMS default or custom key enabled.
  - Automated Backups: 7 days retention.
  - Connection parameter group: UTF-8mb4 character set.

- [ ] **Canonical Migration Execution**
  - Execute automated schema migration runner using Go binary or dedicated migration task:
    - Path: `backend/internal/database/migrations` (Embedded `001_...` through `127_sportal_platform_settings.sql`).
  - Verify zero migration failures:
    - Table `schema_migrations` contains all 138 migration versions intact.
  - Verify essential tables present:
    - `organizations`, `users`, `roles`, `permissions`, `subscriptions`, `shipments`, `invoices`, `sportal_platform_settings`, `audit_logs`, `action_requests`.

- [ ] **Persistent Data Protection Guard**
  - Staging database is completely isolated (`freel_staging`).
  - Production database host (`freel-prod.rds...`) is blocked from staging VPC routing tables.

---

## 5. Go Shared Backend Deployment

- [ ] **Container Image Build & Push (AWS ECR)**
  - Dockerfile: Multi-stage Go build (`golang:1.24-alpine` -> `alpine:3.20` or `distroless`).
  - Binary: `server` with all embedded assets and migrations.
  - Build tag: `v1.0.0-staging-YYYYMMDD`.
  - Image vulnerability scan: Zero high/critical vulnerabilities.

- [ ] **ECS Task Definition / Environment Injection**
  - Inject environment variables from Secrets Manager / SSM:
    - `PORT=8080`
    - `ENV=staging`
    - `APP_URL=https://staging-app.logisticshq.in`
    - `SPORTAL_URL=https://staging-sportal.logisticshq.in`
    - `CORS_ALLOWED_ORIGINS=https://staging-app.logisticshq.in,https://staging-sportal.logisticshq.in`
    - `AI_SIDECAR_URL=http://staging-sidecar.internal:8090`
  - Health check configured:
    - `HTTP GET :8080/health` -> Interval `15s`, Timeout `5s`, Retries `3`.

- [ ] **Deploy & Verify Health**
  - Start ECS service with desired count >= 2 across 2 AZs.
  - ALB Target Group health check status: `healthy`.
  - Curl check: `curl -f -i https://staging-api.logisticshq.in/health` returns `200 OK`.

---

## 6. Python AI Sidecar Deployment

- [ ] **Container Image Build & Push (AWS ECR)**
  - Dockerfile: Multi-stage Python build (`python:3.11-slim`).
  - Dependencies: `pip install -r requirements.txt`.
  - Non-root user: `appuser`.

- [ ] **Sidecar Task Definition & Security Boundary**
  - Network: Private VPC / Service Connect only (`staging-sidecar.internal`).
  - Direct ingress from internet: DENIED.
  - Environment:
    - `PORT=8090`
    - `ENVIRONMENT=staging`
    - `GO_BACKEND_URL=http://staging-api.internal:8080`
    - `INTERNAL_SERVICE_TOKEN=...`
    - `AUTONOMY_LEVEL=advisory` (Strictly non-autonomous in staging).
    - `FEATURE_FLAG_AI_ACTIONS_ENABLED=false` (Execution through Go Action System only).

- [ ] **Deploy & Verify Health**
  - ECS Service started in private subnet.
  - Health check: `curl -f http://staging-sidecar.internal:8090/health` returns `200 OK`.

---

## 7. AI Task Worker Deployment

- [ ] **Worker Process Verification**
  - Worker configured as background queue consumer.
  - Connects to shared staging MariaDB and internal event queue.
  - Retries, exponential backoff, and dead-letter handling configured.
  - Health endpoint / heartbeat liveness check verified.

---

## 8. SPortal Frontend Staging Deployment

- [ ] **Production Build with Staging Environment**
  - Inject build-time environment variables:
    - `VITE_API_URL=https://staging-api.logisticshq.in`
    - `VITE_APP_URL=https://staging-sportal.logisticshq.in`
    - `VITE_APP_ENV=staging`
    - `VITE_PORTAL_NAME=LogisticsHQ SPortal`
  - Run build: `npm run build` in `sportal/`.
  - Verify static output in `dist/`:
    - Zero `localhost:8080` references.
    - Zero server secrets or private keys in JS chunks.
    - Chunk split: `vendor`, `react-vendor`, and route chunks present.

- [ ] **S3 + CloudFront Static Hosting Setup**
  - Staging S3 Bucket: `sportal-frontend-staging`.
  - Block all public S3 access (Origin Access Control / OAC enabled).
  - CloudFront Distribution configured:
    - Viewer Protocol Policy: `Redirect HTTP to HTTPS`.
    - Custom error response: `404` -> `/index.html` (HTTP `200`) for SPA routing.
    - Security headers: HSTS, X-Content-Type-Options, Frame-Options, Content-Security-Policy.
  - Invalidate CloudFront cache: `/*`.

---

## 9. CPortal Frontend Staging Deployment (Coordinated Release)

- [ ] **CPortal Production Build**
  - Inject build-time environment:
    - `VITE_API_URL=https://staging-api.logisticshq.in`
    - `VITE_APP_URL=https://staging-app.logisticshq.in`
  - Run build: `npm run build` in `frontend/`.
  - Sync to CPortal staging S3 bucket and invalidate CloudFront cache.

---

## 10. External Integrations Staging Safety

- [ ] **AWS SES Safe Staging Configuration**
  - Region configured (e.g., `ap-south-1`).
  - Staging sandbox mode active.
  - Verified sender email domain: `staging.logisticshq.in` or test mailbox.
  - Customer suppression list enabled.
  - **Live production emails: FORBIDDEN.**
  - Real email delivery state: `REAL EMAIL DELIVERY — NOT EXECUTED`.

- [ ] **Twilio Safe Staging Configuration**
  - Test credentials configured.
  - Webhook callback endpoints pointed to `https://staging-api.logisticshq.in/webhooks/twilio`.
  - **Live customer SMS: FORBIDDEN.**
  - Real SMS delivery state: `REAL SMS DELIVERY — NOT EXECUTED`.

- [ ] **Carrier Tracking Integrations**
  - Staging / mock webhook endpoints configured.
  - Webhook signature validation enabled.
  - Replay attack defenses active.

---

## 11. Staging Smoke Test Plan

- [ ] **SPortal Core Smoke Checklist**
  - [ ] Login with authorized LogisticsHQ internal user (`admin@logisticshq.in`).
  - [ ] Dashboard renders KPI cards, charts, and activity feeds cleanly without console errors.
  - [ ] Visual verification against `sporatlDashboard.png` (Navy sidebar `#0B192C`, clean cards, light theme).
  - [ ] Organizations module: Lists staging tenant organizations.
  - [ ] Customer 360: View organization detail, subscriptions, usage, and health.
  - [ ] Visual verification against `sportalCustomerView.png`.
  - [ ] Subscriptions & Billing: View active plan, invoices, and billing cycles.
  - [ ] Users & Roles: Inspect internal users and customer users.
  - [ ] Integrations: View status of ERP, Carrier, Payment, and Notification gateways.
  - [ ] Documents: View compliance certificates and contracts with authorized signed URLs.
  - [ ] SPortal AI: Query executive intelligence; verify prompt injection resistance.
  - [ ] Settings & Admin: Verify feature flags, audit logs, and operational controls.

- [ ] **CPortal Cross-Portal Consistency Smoke Checklist**
  - [ ] Customer login on `staging-app.logisticshq.in`.
  - [ ] Access customer dashboard and active shipments.
  - [ ] Verify data updates in CPortal instantly reflect in SPortal Customer 360.

---

## 12. Security Acceptance Gate

- [ ] **Authentication & RBAC Boundary**
  - Customer JWT token rejected with `403 Forbidden` on SPortal administrative routes.
  - Unauthenticated requests rejected with `401 Unauthorized`.
  - Legacy test tokens (`test-token`, `test-token-org<N>`) strictly rejected (`401 Unauthorized`).
- [ ] **Tenant Isolation & IDOR**
  - Cross-tenant queries return `404 Not Found` or `403 Forbidden`.
- [ ] **CORS Strict Origin Enforcement**
  - Request with origin `https://staging-sportal.logisticshq.in` receives `Access-Control-Allow-Origin`.
  - Request with origin `https://attacker-domain.com` receives no CORS headers.
- [ ] **Secret Hygiene**
  - Inspect browser network panel and console: Zero DB passwords, AWS secret keys, or JWT secrets exposed.

---

## 13. Rollback Procedure

- [ ] **SPortal Frontend Rollback**
  - Re-point CloudFront origin to previous S3 version or restore previous build artifact.
  - Invalidate CloudFront cache (`/*`). Recovery time: < 2 minutes.

- [ ] **Go Backend Rollback**
  - In ECS console or CLI, update service task definition revision to previous known-good tag `v0.9.x`.
  - ECS conducts rolling deployment back to previous container. Recovery time: < 3 minutes.

- [ ] **Database Migration Rollback Consideration**
  - All migrations are designed to be non-destructive (additive schema changes).
  - In case of critical schema issue: restore staging database snapshot taken immediately before deployment.
