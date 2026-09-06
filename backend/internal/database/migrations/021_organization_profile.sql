-- Migration: 021_organization_profile.sql
-- Description: Adds company profile fields to the organizations table.

ALTER TABLE organizations
ADD COLUMN legal_name VARCHAR(255) NULL,
ADD COLUMN registration_number VARCHAR(100) NULL,
ADD COLUMN tax_number VARCHAR(100) NULL,
ADD COLUMN website VARCHAR(255) NULL,
ADD COLUMN primary_email VARCHAR(255) NULL,
ADD COLUMN phone_number VARCHAR(50) NULL,
ADD COLUMN support_email VARCHAR(255) NULL,
ADD COLUMN address TEXT NULL,
ADD COLUMN city VARCHAR(100) NULL,
ADD COLUMN state VARCHAR(100) NULL,
ADD COLUMN country VARCHAR(100) NULL,
ADD COLUMN postal_code VARCHAR(50) NULL,
ADD COLUMN industry VARCHAR(100) NULL,
ADD COLUMN company_type VARCHAR(100) NULL,
ADD COLUMN default_currency VARCHAR(10) DEFAULT 'USD',
ADD COLUMN default_timezone VARCHAR(100) DEFAULT 'UTC',
ADD COLUMN date_format VARCHAR(50) DEFAULT 'MM/DD/YYYY',
ADD COLUMN logo_url VARCHAR(512) NULL;
