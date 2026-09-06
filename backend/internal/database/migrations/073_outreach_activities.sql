-- +migrate Up

CREATE TABLE IF NOT EXISTS outreach_activities (
    id BIGSERIAL PRIMARY KEY,
    org_id BIGINT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    lead_id BIGINT NULL REFERENCES leads(id) ON DELETE CASCADE,
    customer_id BIGINT NULL REFERENCES customers(id) ON DELETE SET NULL,
    activity_type VARCHAR(50) NOT NULL, -- EMAIL, CALL, FOLLOW_UP, MEETING, OTHER
    subject VARCHAR(255) NOT NULL,
    description TEXT NULL,
    status VARCHAR(50) DEFAULT 'PENDING', -- PENDING, IN_PROGRESS, COMPLETED, OVERDUE
    priority VARCHAR(50) DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH
    scheduled_at TIMESTAMP WITH TIME ZONE NULL,
    completed_at TIMESTAMP WITH TIME ZONE NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_by BIGINT NULL REFERENCES users(id) ON DELETE SET NULL
);

-- Seed some mock activities matching the UI layout dynamically based on existing leads
INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
SELECT 
    l.org_id,
    l.id,
    'CALL',
    'Follow-up Call',
    'Follow-up Call to John Smith',
    'PENDING',
    'HIGH',
    NOW() + INTERVAL '2 hour',
    (SELECT id FROM users WHERE org_id = l.org_id LIMIT 1)
FROM leads l
WHERE l.company_name LIKE '%Oceanic%' OR l.id = 1 LIMIT 1;

INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
SELECT 
    l.org_id,
    l.id,
    'EMAIL',
    'Email Follow-up',
    'Send follow-up details to Priya',
    'IN_PROGRESS',
    'MEDIUM',
    NOW() + INTERVAL '5 hour',
    (SELECT id FROM users WHERE org_id = l.org_id LIMIT 1)
FROM leads l
WHERE l.company_name LIKE '%Transworld%' OR l.id = 2 LIMIT 1;

INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
SELECT 
    l.org_id,
    l.id,
    'MEETING',
    'Proposal Discussion',
    'Discuss Q3 contract options',
    'PENDING',
    'HIGH',
    NOW() + INTERVAL '1 day',
    (SELECT id FROM users WHERE org_id = l.org_id LIMIT 1)
FROM leads l
WHERE l.company_name LIKE '%Global%' OR l.id = 3 LIMIT 1;

INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
SELECT 
    l.org_id,
    l.id,
    'EMAIL',
    'Follow-up Email',
    'Check in on inactive account',
    'PENDING',
    'LOW',
    NOW() + INTERVAL '1 day' + INTERVAL '6 hour',
    (SELECT id FROM users WHERE org_id = l.org_id LIMIT 1)
FROM leads l
WHERE l.company_name LIKE '%Bluewave%' OR l.id = 4 LIMIT 1;

INSERT INTO outreach_activities (org_id, lead_id, activity_type, subject, description, status, priority, scheduled_at, created_by)
SELECT 
    l.org_id,
    l.id,
    'CALL',
    'Contract Renewal Call',
    'Discuss upcoming renewals',
    'OVERDUE',
    'HIGH',
    NOW() - INTERVAL '1 day',
    (SELECT id FROM users WHERE org_id = l.org_id LIMIT 1)
FROM leads l
WHERE l.company_name LIKE '%Speedex%' OR l.id = 5 LIMIT 1;

-- +migrate Down
DROP TABLE IF EXISTS outreach_activities;
