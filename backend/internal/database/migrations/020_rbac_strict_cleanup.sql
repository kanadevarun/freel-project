-- +migrate Up
-- ============================================================
-- 020_rbac_strict_cleanup.sql
-- Strict cleanup of permissions and role_permissions to guarantee
-- exactly 40 canonical permissions with sequential IDs (1 to 40)
-- without breaking org_members or existing roles.
-- ============================================================

-- 1. Remove all existing permission mappings
TRUNCATE TABLE role_permissions;

-- 2. Remove all existing permissions
DELETE FROM permissions;

-- 3. Reset auto-increment so the new permissions start at ID 1
ALTER TABLE permissions AUTO_INCREMENT = 1;

-- 4. Insert exactly 40 canonical permissions
-- This ensures sequential IDs 1 to 40.
INSERT INTO permissions (resource, action, description) VALUES 
('COMPANIES', 'CREATE', 'Create company records'),
('COMPANIES', 'READ', 'View company records'),
('COMPANIES', 'UPDATE', 'Edit company records'),
('COMPANIES', 'DELETE', 'Delete company records'),
('LEADS', 'CREATE', 'Create leads'),
('LEADS', 'READ', 'View leads'),
('LEADS', 'UPDATE', 'Edit leads'),
('LEADS', 'DELETE', 'Delete leads'),
('OPPORTUNITIES', 'CREATE', 'Create sales opportunities'),
('OPPORTUNITIES', 'READ', 'View sales opportunities'),
('OPPORTUNITIES', 'UPDATE', 'Edit sales opportunities'),
('OPPORTUNITIES', 'DELETE', 'Delete sales opportunities'),
('RFQS', 'CREATE', 'Create RFQs'),
('RFQS', 'READ', 'View RFQs'),
('RFQS', 'UPDATE', 'Edit RFQs'),
('RFQS', 'DELETE', 'Delete RFQs'),
('OUTREACH', 'CREATE', 'Create outreach campaigns'),
('OUTREACH', 'READ', 'View outreach campaigns'),
('OUTREACH', 'UPDATE', 'Edit outreach campaigns'),
('OUTREACH', 'DELETE', 'Delete outreach campaigns'),
('SHIPMENTS', 'CREATE', 'Create shipments'),
('SHIPMENTS', 'READ', 'View shipments'),
('SHIPMENTS', 'UPDATE', 'Edit shipments'),
('SHIPMENTS', 'DELETE', 'Delete shipments'),
('DOCUMENTS', 'CREATE', 'Create documents'),
('DOCUMENTS', 'READ', 'View documents'),
('DOCUMENTS', 'UPDATE', 'Edit documents'),
('DOCUMENTS', 'DELETE', 'Delete documents'),
('FINANCE', 'CREATE', 'Create finance records'),
('FINANCE', 'READ', 'View finance records'),
('FINANCE', 'UPDATE', 'Edit finance records'),
('FINANCE', 'DELETE', 'Delete finance records'),
('USERS', 'CREATE', 'Invite/create users'),
('USERS', 'READ', 'View users and team'),
('USERS', 'UPDATE', 'Edit user profiles and roles'),
('USERS', 'DELETE', 'Remove users from organization'),
('SETTINGS', 'CREATE', 'Create settings entries'),
('SETTINGS', 'READ', 'View settings'),
('SETTINGS', 'UPDATE', 'Edit settings'),
('SETTINGS', 'DELETE', 'Delete settings entries');

-- 5. Map permissions to roles
-- SUPER_ADMIN gets all 40 permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'SUPER_ADMIN';

-- SALES
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'SALES'
AND (
  (p.resource IN ('COMPANIES','LEADS','OPPORTUNITIES','RFQS','OUTREACH') AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource IN ('SHIPMENTS','DOCUMENTS','FINANCE') AND p.action = 'READ')
);

-- PRICING
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'PRICING'
AND (
  (p.resource IN ('RFQS','COMPANIES','OPPORTUNITIES') AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource IN ('SHIPMENTS','DOCUMENTS') AND p.action = 'READ')
);

-- OPERATIONS
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'OPERATIONS'
AND (
  (p.resource IN ('SHIPMENTS','DOCUMENTS') AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource IN ('COMPANIES','RFQS') AND p.action = 'READ')
);

-- FINANCE
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'FINANCE'
AND (
  (p.resource = 'FINANCE' AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource IN ('COMPANIES','SHIPMENTS','DOCUMENTS') AND p.action = 'READ')
);

-- DOCUMENTATION
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'DOCUMENTATION'
AND (
  (p.resource = 'DOCUMENTS' AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource IN ('SHIPMENTS','COMPANIES') AND p.action = 'READ')
);

-- HR
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
WHERE r.org_id = 1 AND r.name = 'HR'
AND (
  (p.resource = 'USERS' AND p.action IN ('CREATE','READ','UPDATE'))
  OR (p.resource = 'COMPANIES' AND p.action = 'READ')
);

-- +migrate Down
-- Cannot safely rollback to unstructured IDs.
