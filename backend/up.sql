internal/database/migrations/001_rbac.sql:-- +migrate Up
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE organizations (
internal/database/migrations/001_rbac.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/001_rbac.sql-    name VARCHAR(255) NOT NULL,
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE users (
internal/database/migrations/001_rbac.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/001_rbac.sql-    cognito_sub VARCHAR(255) UNIQUE NOT NULL,
internal/database/migrations/001_rbac.sql-    email VARCHAR(255) UNIQUE NOT NULL,
internal/database/migrations/001_rbac.sql-    first_name VARCHAR(255),
internal/database/migrations/001_rbac.sql-    last_name VARCHAR(255),
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE roles (
internal/database/migrations/001_rbac.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/001_rbac.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/001_rbac.sql-    name VARCHAR(255) NOT NULL,
internal/database/migrations/001_rbac.sql-    description TEXT,
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    UNIQUE(org_id, name)
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE org_members (
internal/database/migrations/001_rbac.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/001_rbac.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/001_rbac.sql-    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
internal/database/migrations/001_rbac.sql-    role_id BIGINT NOT NULL REFERENCES roles(id),
internal/database/migrations/001_rbac.sql-    status VARCHAR(50) DEFAULT 'ACTIVE',
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    UNIQUE(org_id, user_id)
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE permissions (
internal/database/migrations/001_rbac.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/001_rbac.sql-    resource VARCHAR(255) NOT NULL,
internal/database/migrations/001_rbac.sql-    action VARCHAR(255) NOT NULL,
internal/database/migrations/001_rbac.sql-    description TEXT,
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    UNIQUE(resource, action)
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-CREATE TABLE role_permissions (
internal/database/migrations/001_rbac.sql-    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
internal/database/migrations/001_rbac.sql-    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
internal/database/migrations/001_rbac.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/001_rbac.sql-    PRIMARY KEY (role_id, permission_id)
internal/database/migrations/001_rbac.sql-);
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-
internal/database/migrations/001_rbac.sql-DROP TABLE role_permissions;
internal/database/migrations/001_rbac.sql-DROP TABLE permissions;
internal/database/migrations/001_rbac.sql-DROP TABLE org_members;
internal/database/migrations/001_rbac.sql-DROP TABLE roles;
internal/database/migrations/001_rbac.sql-DROP TABLE users;
internal/database/migrations/001_rbac.sql-DROP TABLE organizations;
--
internal/database/migrations/002_companies_leads.sql:-- +migrate Up
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-CREATE TABLE addresses (
internal/database/migrations/002_companies_leads.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/002_companies_leads.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/002_companies_leads.sql-    address_line_1 VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    address_line_2 VARCHAR(255),
internal/database/migrations/002_companies_leads.sql-    city VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    state VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    postal_code VARCHAR(20) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    country_code VARCHAR(10) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/002_companies_leads.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/002_companies_leads.sql-);
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-CREATE TABLE companies (
internal/database/migrations/002_companies_leads.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/002_companies_leads.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/002_companies_leads.sql-    name VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    domain VARCHAR(255),
internal/database/migrations/002_companies_leads.sql-    industry VARCHAR(255),
internal/database/migrations/002_companies_leads.sql-    address_id BIGINT REFERENCES addresses(id),
internal/database/migrations/002_companies_leads.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/002_companies_leads.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/002_companies_leads.sql-);
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-CREATE TABLE contacts (
internal/database/migrations/002_companies_leads.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/002_companies_leads.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/002_companies_leads.sql-    company_id BIGINT REFERENCES companies(id) ON DELETE SET NULL,
internal/database/migrations/002_companies_leads.sql-    first_name VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    last_name VARCHAR(255) NOT NULL,
internal/database/migrations/002_companies_leads.sql-    email VARCHAR(255),
internal/database/migrations/002_companies_leads.sql-    phone VARCHAR(50),
internal/database/migrations/002_companies_leads.sql-    job_title VARCHAR(255),
internal/database/migrations/002_companies_leads.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/002_companies_leads.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/002_companies_leads.sql-);
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-CREATE TABLE customers (
internal/database/migrations/002_companies_leads.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/002_companies_leads.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/002_companies_leads.sql-    company_id BIGINT NOT NULL REFERENCES companies(id),
internal/database/migrations/002_companies_leads.sql-    status VARCHAR(50) DEFAULT 'ACTIVE',
internal/database/migrations/002_companies_leads.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/002_companies_leads.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/002_companies_leads.sql-);
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-
internal/database/migrations/002_companies_leads.sql-DROP TABLE customers;
internal/database/migrations/002_companies_leads.sql-DROP TABLE contacts;
internal/database/migrations/002_companies_leads.sql-DROP TABLE companies;
internal/database/migrations/002_companies_leads.sql-DROP TABLE addresses;
--
internal/database/migrations/003_outreach.sql:-- +migrate Up
internal/database/migrations/003_outreach.sql-
internal/database/migrations/003_outreach.sql-CREATE TABLE outreach_campaigns (
internal/database/migrations/003_outreach.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/003_outreach.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/003_outreach.sql-    name VARCHAR(255) NOT NULL,
internal/database/migrations/003_outreach.sql-    status VARCHAR(50) DEFAULT 'DRAFT',
internal/database/migrations/003_outreach.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/003_outreach.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/003_outreach.sql-);
internal/database/migrations/003_outreach.sql-
internal/database/migrations/003_outreach.sql-CREATE TABLE outreach_sequences (
internal/database/migrations/003_outreach.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/003_outreach.sql-    campaign_id BIGINT NOT NULL REFERENCES outreach_campaigns(id) ON DELETE CASCADE,
internal/database/migrations/003_outreach.sql-    step_number INT NOT NULL,
internal/database/migrations/003_outreach.sql-    channel VARCHAR(50) NOT NULL, -- EMAIL, LINKEDIN
internal/database/migrations/003_outreach.sql-    template TEXT NOT NULL,
internal/database/migrations/003_outreach.sql-    delay_days INT DEFAULT 0,
internal/database/migrations/003_outreach.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/003_outreach.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/003_outreach.sql-);
internal/database/migrations/003_outreach.sql-
internal/database/migrations/003_outreach.sql-
internal/database/migrations/003_outreach.sql-DROP TABLE outreach_sequences;
internal/database/migrations/003_outreach.sql-DROP TABLE outreach_campaigns;
--
internal/database/migrations/004_rfq.sql:-- +migrate Up
internal/database/migrations/004_rfq.sql-
internal/database/migrations/004_rfq.sql-CREATE TABLE rfqs (
internal/database/migrations/004_rfq.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/004_rfq.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/004_rfq.sql-    customer_id BIGINT NOT NULL REFERENCES customers(id),
internal/database/migrations/004_rfq.sql-    status VARCHAR(50) DEFAULT 'DRAFT',
internal/database/migrations/004_rfq.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/004_rfq.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/004_rfq.sql-);
internal/database/migrations/004_rfq.sql-
internal/database/migrations/004_rfq.sql-CREATE TABLE rfq_items (
internal/database/migrations/004_rfq.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/004_rfq.sql-    rfq_id BIGINT NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
internal/database/migrations/004_rfq.sql-    description TEXT,
internal/database/migrations/004_rfq.sql-    quantity INT NOT NULL,
internal/database/migrations/004_rfq.sql-    weight_kg DECIMAL(10,2),
internal/database/migrations/004_rfq.sql-    volume_cbm DECIMAL(10,2),
internal/database/migrations/004_rfq.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/004_rfq.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/004_rfq.sql-);
internal/database/migrations/004_rfq.sql-
internal/database/migrations/004_rfq.sql-
internal/database/migrations/004_rfq.sql-DROP TABLE rfq_items;
internal/database/migrations/004_rfq.sql-DROP TABLE rfqs;
--
internal/database/migrations/005_opportunities.sql:-- +migrate Up
internal/database/migrations/005_opportunities.sql-
internal/database/migrations/005_opportunities.sql-CREATE TABLE opportunities (
internal/database/migrations/005_opportunities.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/005_opportunities.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/005_opportunities.sql-    customer_id BIGINT NOT NULL REFERENCES customers(id),
internal/database/migrations/005_opportunities.sql-    title VARCHAR(255) NOT NULL,
internal/database/migrations/005_opportunities.sql-    amount DECIMAL(15,2),
internal/database/migrations/005_opportunities.sql-    stage VARCHAR(50) DEFAULT 'PROSPECTING',
internal/database/migrations/005_opportunities.sql-    expected_close_date DATE,
internal/database/migrations/005_opportunities.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/005_opportunities.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/005_opportunities.sql-);
internal/database/migrations/005_opportunities.sql-
internal/database/migrations/005_opportunities.sql-
internal/database/migrations/005_opportunities.sql-DROP TABLE opportunities;
--
internal/database/migrations/006_audit_activity.sql:-- +migrate Up
internal/database/migrations/006_audit_activity.sql-
internal/database/migrations/006_audit_activity.sql-CREATE TABLE audit_logs (
internal/database/migrations/006_audit_activity.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/006_audit_activity.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/006_audit_activity.sql-    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
internal/database/migrations/006_audit_activity.sql-    action VARCHAR(255) NOT NULL,
internal/database/migrations/006_audit_activity.sql-    resource_type VARCHAR(255) NOT NULL,
internal/database/migrations/006_audit_activity.sql-    resource_id BIGINT NOT NULL,
internal/database/migrations/006_audit_activity.sql-    details JSONB,
internal/database/migrations/006_audit_activity.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/006_audit_activity.sql-);
internal/database/migrations/006_audit_activity.sql-
internal/database/migrations/006_audit_activity.sql-CREATE TABLE activities (
internal/database/migrations/006_audit_activity.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/006_audit_activity.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/006_audit_activity.sql-    entity_type VARCHAR(255) NOT NULL,
internal/database/migrations/006_audit_activity.sql-    entity_id BIGINT NOT NULL,
internal/database/migrations/006_audit_activity.sql-    action VARCHAR(255) NOT NULL,
internal/database/migrations/006_audit_activity.sql-    description TEXT,
internal/database/migrations/006_audit_activity.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
internal/database/migrations/006_audit_activity.sql-);
internal/database/migrations/006_audit_activity.sql-
internal/database/migrations/006_audit_activity.sql-
internal/database/migrations/006_audit_activity.sql-DROP TABLE activities;
internal/database/migrations/006_audit_activity.sql-DROP TABLE audit_logs;
--
internal/database/migrations/007_invitations.sql:-- +migrate Up
internal/database/migrations/007_invitations.sql-
internal/database/migrations/007_invitations.sql-CREATE TABLE invitations (
internal/database/migrations/007_invitations.sql-    id BIGSERIAL PRIMARY KEY,
internal/database/migrations/007_invitations.sql-    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
internal/database/migrations/007_invitations.sql-    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
internal/database/migrations/007_invitations.sql-    email VARCHAR(255) NOT NULL,
internal/database/migrations/007_invitations.sql-    token VARCHAR(255) UNIQUE NOT NULL,
internal/database/migrations/007_invitations.sql-    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
internal/database/migrations/007_invitations.sql-    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/007_invitations.sql-    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
internal/database/migrations/007_invitations.sql-    UNIQUE(org_id, email)
internal/database/migrations/007_invitations.sql-);
internal/database/migrations/007_invitations.sql-
internal/database/migrations/007_invitations.sql-
internal/database/migrations/007_invitations.sql-DROP TABLE invitations;
