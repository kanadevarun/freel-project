-- Add health score to RFQs
ALTER TABLE rfqs 
ADD COLUMN IF NOT EXISTS health_score INT DEFAULT 0;

-- Add AI recommendation fields to quotes for carrier comparison
ALTER TABLE rfq_quotes
ADD COLUMN IF NOT EXISTS reliability_score INT DEFAULT 0,
ADD COLUMN IF NOT EXISTS historical_success_rate FLOAT DEFAULT 0.0;
