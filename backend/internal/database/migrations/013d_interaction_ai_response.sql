-- Migration: 013d_interaction_ai_response.sql
-- Add ai_summary and drafted_reply columns to lead_interactions to store AI summaries and generated emails.

ALTER TABLE lead_interactions ADD COLUMN IF NOT EXISTS ai_summary TEXT;
ALTER TABLE lead_interactions ADD COLUMN IF NOT EXISTS drafted_reply TEXT;
