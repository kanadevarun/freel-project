-- Add agent_status to track the state of the AI Agent working on this RFQ
ALTER TABLE rfqs 
ADD COLUMN IF NOT EXISTS agent_status VARCHAR(50) DEFAULT 'IDLE';

-- Add ai_reasoning to quotes to store the justification for margin/carrier selection
ALTER TABLE rfq_quotes
ADD COLUMN IF NOT EXISTS ai_reasoning TEXT;
