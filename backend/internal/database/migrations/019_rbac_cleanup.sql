-- +migrate Up
-- ============================================================
-- 019_rbac_cleanup.sql
-- Idempotent RBAC cleanup migration:
--   1. Ensure full CRUD permission catalog (10 resources × 4 actions = 40)
--   2. Remove WRITE permission references from role_permissions
--   3. Remove WRITE permissions from permissions table
--   4. Seed 6 missing default roles for org 1
--   5. Rebuild SUPER_ADMIN role_permissions (full CRUD, no WRITE)
--   6. Seed least-privilege permissions for each default role
-- ============================================================

-- STEP 1: Insert the canonical 40 permissions (idempotent via INSERT IGNORE)
-- Companies
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('COMPANIES', 'CREATE', 'Create company records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('COMPANIES', 'READ',   'View company records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('COMPANIES', 'UPDATE', 'Edit company records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('COMPANIES', 'DELETE', 'Delete company records');
-- Leads
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('LEADS', 'CREATE', 'Create leads');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('LEADS', 'READ',   'View leads');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('LEADS', 'UPDATE', 'Edit leads');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('LEADS', 'DELETE', 'Delete leads');
-- Opportunities
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OPPORTUNITIES', 'CREATE', 'Create sales opportunities');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OPPORTUNITIES', 'READ',   'View sales opportunities');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OPPORTUNITIES', 'UPDATE', 'Edit sales opportunities');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OPPORTUNITIES', 'DELETE', 'Delete sales opportunities');
-- RFQs
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('RFQS', 'CREATE', 'Create RFQs');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('RFQS', 'READ',   'View RFQs');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('RFQS', 'UPDATE', 'Edit RFQs');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('RFQS', 'DELETE', 'Delete RFQs');
-- Outreach
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OUTREACH', 'CREATE', 'Create outreach campaigns');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OUTREACH', 'READ',   'View outreach campaigns');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OUTREACH', 'UPDATE', 'Edit outreach campaigns');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('OUTREACH', 'DELETE', 'Delete outreach campaigns');
-- Shipments
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SHIPMENTS', 'CREATE', 'Create shipments');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SHIPMENTS', 'READ',   'View shipments');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SHIPMENTS', 'UPDATE', 'Edit shipments');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SHIPMENTS', 'DELETE', 'Delete shipments');
-- Documents
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('DOCUMENTS', 'CREATE', 'Create documents');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('DOCUMENTS', 'READ',   'View documents');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('DOCUMENTS', 'UPDATE', 'Edit documents');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('DOCUMENTS', 'DELETE', 'Delete documents');
-- Finance
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('FINANCE', 'CREATE', 'Create finance records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('FINANCE', 'READ',   'View finance records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('FINANCE', 'UPDATE', 'Edit finance records');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('FINANCE', 'DELETE', 'Delete finance records');
-- Users
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('USERS', 'CREATE', 'Invite/create users');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('USERS', 'READ',   'View users and team');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('USERS', 'UPDATE', 'Edit user profiles and roles');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('USERS', 'DELETE', 'Remove users from organization');
-- Settings
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SETTINGS', 'CREATE', 'Create settings entries');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SETTINGS', 'READ',   'View settings');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SETTINGS', 'UPDATE', 'Edit settings');
INSERT IGNORE INTO permissions (resource, action, description) VALUES ('SETTINGS', 'DELETE', 'Delete settings entries');

-- STEP 2: Remove WRITE permission references from role_permissions (FK constraint first)
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE action = 'WRITE');

-- STEP 3: Remove all WRITE permissions from the catalog
DELETE FROM permissions WHERE action = 'WRITE';

-- STEP 4: Seed 6 missing default roles for org 1 (INSERT IGNORE = idempotent)
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'SALES',         'Manage leads, RFQs and customers');
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'PRICING',        'Manage pricing and quotations');
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'OPERATIONS',     'Manage shipments and tracking');
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'FINANCE',        'Manage invoices and payments');
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'DOCUMENTATION',  'Manage documents and compliance');
INSERT IGNORE INTO roles (org_id, name, description) VALUES (1, 'HR',             'Manage HR and team data');

-- STEP 5: Rebuild SUPER_ADMIN permissions (clear old, re-assign all 40 canonical CRUD permissions)
-- We preserve role id by doing a subquery lookup
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE org_id = 1 AND name = 'SUPER_ADMIN' LIMIT 1);

INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE org_id = 1 AND name = 'SUPER_ADMIN' LIMIT 1),
    p.id
FROM permissions p
WHERE p.resource IN ('COMPANIES','LEADS','OPPORTUNITIES','RFQS','OUTREACH','SHIPMENTS','DOCUMENTS','FINANCE','USERS','SETTINGS')
  AND p.action IN ('CREATE','READ','UPDATE','DELETE');

-- STEP 6: Seed SALES role permissions (C/R/U on sales resources, R-only on ops)
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'SALES' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource IN ('COMPANIES','LEADS','OPPORTUNITIES','RFQS','OUTREACH') AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource IN ('SHIPMENTS','DOCUMENTS','FINANCE') AND p.action = 'READ');

-- STEP 7: Seed PRICING role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'PRICING' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource IN ('RFQS','COMPANIES','OPPORTUNITIES') AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource IN ('SHIPMENTS','DOCUMENTS') AND p.action = 'READ');

-- STEP 8: Seed OPERATIONS role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'OPERATIONS' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource IN ('SHIPMENTS','DOCUMENTS') AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource IN ('COMPANIES','RFQS') AND p.action = 'READ');

-- STEP 9: Seed FINANCE role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'FINANCE' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource = 'FINANCE' AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource IN ('COMPANIES','SHIPMENTS','DOCUMENTS') AND p.action = 'READ');

-- STEP 10: Seed DOCUMENTATION role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'DOCUMENTATION' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource IN ('DOCUMENTS') AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource IN ('SHIPMENTS','COMPANIES') AND p.action = 'READ');

-- STEP 11: Seed HR role permissions
INSERT IGNORE INTO role_permissions (role_id, permission_id)
SELECT (SELECT id FROM roles WHERE org_id = 1 AND name = 'HR' LIMIT 1), p.id
FROM permissions p
WHERE (p.resource = 'USERS' AND p.action IN ('CREATE','READ','UPDATE'))
   OR (p.resource = 'COMPANIES' AND p.action = 'READ');

-- +migrate Down

-- Remove all role_permissions for the 6 seeded roles
DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id FROM roles WHERE org_id = 1 AND name IN ('SALES','PRICING','OPERATIONS','FINANCE','DOCUMENTATION','HR')
);

-- Remove the 6 seeded roles
DELETE FROM roles WHERE org_id = 1 AND name IN ('SALES','PRICING','OPERATIONS','FINANCE','DOCUMENTATION','HR');

-- Re-delete the 40 canonical permissions (will be re-created if migration re-runs)
-- NOTE: This does NOT restore WRITE permissions - that data is intentionally discarded.
