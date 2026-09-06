-- Migration: 023_workspace_preferences.sql
-- Description: Adds workspace preference fields to the organizations table.

ALTER TABLE organizations
ADD COLUMN default_language VARCHAR(50) DEFAULT 'English',
ADD COLUMN measurement_system VARCHAR(50) DEFAULT 'Metric (kg, cm, km)',
ADD COLUMN weight_unit VARCHAR(50) DEFAULT 'Kilogram (kg)',
ADD COLUMN dimension_unit VARCHAR(50) DEFAULT 'Centimeter (cm)',
ADD COLUMN volume_unit VARCHAR(50) DEFAULT 'Cubic Meter (CBM)',
ADD COLUMN time_format VARCHAR(20) DEFAULT '24 Hours';
