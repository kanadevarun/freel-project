package ai

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"
)

// PromptDefinition represents an authoritative versioned prompt.
type PromptDefinition struct {
	PromptKey       string   `json:"prompt_key"`
	Version         string   `json:"version"`
	WorkflowName    string   `json:"workflow_name"`
	Purpose         string   `json:"purpose"`
	TemplateContent string   `json:"template_content"`
	InputVariables  []string `json:"input_variables"`
	OutputFormat    string   `json:"output_format"`
	IsActive        bool     `json:"is_active"`
}

// PromptManager handles fetching, formatting, and versioning prompts across AI workflows.
type PromptManager interface {
	GetPrompt(promptKey string, variables map[string]interface{}) (string, error)
	GetPromptWithVersion(promptKey string, version string, variables map[string]interface{}) (string, error)
	GetPromptDefinition(promptKey string, version string) (*PromptDefinition, error)
	ListPrompts() []*PromptDefinition
}

type promptManager struct {
	mu        sync.RWMutex
	prompts   map[string]map[string]*PromptDefinition // key -> version -> definition
	aliases   map[string]string                       // legacy_key -> canonical_key
}

// NewPromptManager initializes the authoritative prompt manager preloaded with v1.0.0 templates.
func NewPromptManager() PromptManager {
	pm := &promptManager{
		prompts: make(map[string]map[string]*PromptDefinition),
		aliases: make(map[string]string),
	}

	// Register aliases for legacy keys
	pm.aliases["score_lead"] = "leads.lead_scoring"
	pm.aliases["generate_email"] = "outreach.cold_email"
	pm.aliases["extract_shipment_request"] = "rfq.extract_shipment_request"
	pm.aliases["pricing_analyst"] = "pricing.analyst"

	// 1. Pricing Analyst
	pm.register(&PromptDefinition{
		PromptKey:    "pricing.analyst",
		Version:      "1.0.0",
		WorkflowName: "pricing",
		Purpose:      "Autonomous Senior Pricing Analyst Agent system prompt to evaluate RFQ carrier rates, apply markup rules, and recommend optimal quotation options.",
		TemplateContent: `You are a Senior Pricing Analyst at a freight forwarding company with 15+ years of experience.
Your job is to analyze trade lanes, apply margin rules, and recommend the best carrier options.
You NEVER invent or hallucinate market rates. You ONLY use the provided Carrier Rates.

Shipment Request:
Origin: {{.Origin}}
Destination: {{.Destination}}
Incoterms: {{.Incoterms}}
Target Date: {{.TargetDate}}
Cargo Items: {{.Items}}

Available Carrier Rates:
{{.CarrierRates}}

Task:
1. Review the available carrier rates.
2. Select the optimal carrier based on a balance of reliability, transit time, and cost.
3. Suggest a target margin (e.g. 15-20%) based on the trade lane and reliability.
4. Prepare a short reasoning explaining your choice to the human Sales rep.
5. Provide a confidence score (0-100). If you are missing data, lower the score.

Return ONLY a JSON object with exactly two fields:
1. "recommendation": An object with: "type" (string), "priority" (string), "confidence" (int), "reason" (string), "suggested_action" (string)
2. "draft_quote": An object with: "carrier_name" (string), "buy_price" (float), "sell_price" (float), "transit_time_days" (int), "reliability_score" (int), "historical_success_rate" (float)`,
		InputVariables: []string{"Origin", "Destination", "Incoterms", "TargetDate", "Items", "CarrierRates"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 2. Sales Email Classification & Extraction
	pm.register(&PromptDefinition{
		PromptKey:    "sales.email_classification",
		Version:      "1.0.0",
		WorkflowName: "sales",
		Purpose:      "Analyzes inbound shipper emails, classifies intent, and extracts structured quote requirements.",
		TemplateContent: `Analyze the following inbound email from a shipper and extract key metadata:

Email From: {{.Sender}}
Subject: {{.Subject}}
Body:
{{.Body}}
{{.PriorContext}}

Extract the following fields and return ONLY a JSON object:
- intent: "RFQ_REQUEST", "QUESTION", "MEETING", "UNSUBSCRIBE", "FOLLOW_UP"
- sentiment: "POSITIVE", "NEUTRAL", "NEGATIVE"
- confidence: integer 0-100
- lead_name: string or null
- company_domain: string or null
- origin_port: string or null
- destination_port: string or null
- incoterms: string or null
- cargo_description: string or null
- cargo_weight: float or null
- cargo_volume: float or null
- target_date: string YYYY-MM-DD or null
- ai_summary: string

Response JSON:`,
		InputVariables: []string{"Sender", "Subject", "Body", "PriorContext", "Today"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 3. Sales Reply Draft
	pm.register(&PromptDefinition{
		PromptKey:    "sales.reply_draft",
		Version:      "1.0.0",
		WorkflowName: "sales",
		Purpose:      "Drafts polite response asking customer for missing quote parameters.",
		TemplateContent: `Draft a polite, professional, and concise email reply to a customer who sent a quote request but missed some mandatory details.

Original Email Subject: {{.Subject}}
Original Email Body:
{{.Body}}

Today's Date: {{.Today}}
Missing Mandatory Fields: {{.MissingFields}}
{{.ReplyContext}}
Write only the email body starting with "{{.Greeting}}" and sign off as "LogisticsHQ Sales Team".`,
		InputVariables: []string{"Subject", "Body", "Today", "MissingFields", "ReplyContext", "Greeting"},
		OutputFormat:   "TEXT",
		IsActive:       true,
	})

	// 4. Operations Tracking Parse
	pm.register(&PromptDefinition{
		PromptKey:    "operations.tracking_parse",
		Version:      "1.0.0",
		WorkflowName: "operations",
		Purpose:      "Parses carrier milestone updates and exception triggers.",
		TemplateContent: `You are an expert freight logistics operations analyst.
Parse this raw carrier update and extract:
1. milestones: list of events with code, date, location, notes
2. exception_signals: anomalies with type, detected, details, delay_hours
3. needs_human_review: boolean
4. ai_summary: single sentence summary

Raw Update:
{{.RawDescription}}

Return ONLY a JSON object.`,
		InputVariables: []string{"RawDescription"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 5. Contracts Rate Extraction
	pm.register(&PromptDefinition{
		PromptKey:    "contracts.rate_extraction",
		Version:      "1.0.0",
		WorkflowName: "contracts",
		Purpose:      "Extracts ocean and air rate cards and surcharges from contract documents.",
		TemplateContent: `You are an expert freight forwarding operations executive. Read the following contract text and extract all port-pair rates.

Document Text:
{{.RawText}}

Carrier SCAC: {{.CarrierSCAC}}
Carrier Name: {{.CarrierName}}

Return a JSON object with key "rates" containing list of rate cards.`,
		InputVariables: []string{"RawText", "CarrierSCAC", "CarrierName"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 6. Compliance Document Verification
	pm.register(&PromptDefinition{
		PromptKey:    "compliance.doc_verification",
		Version:      "1.0.0",
		WorkflowName: "compliance",
		Purpose:      "Verifies shipping documentation against master booking data.",
		TemplateContent: `Compare extracted shipping document metadata against master shipment records.
Identify discrepancies in gross weight, container numbers, seal numbers, and party names.
Return ONLY JSON with discrepancies, compliance_status, and summary.`,
		InputVariables: []string{"DocumentText", "ExpectedData"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 7. Finance Invoice Audit
	pm.register(&PromptDefinition{
		PromptKey:    "finance.invoice_audit",
		Version:      "1.0.0",
		WorkflowName: "finance",
		Purpose:      "Audits vendor freight invoices against agreed quotation charges.",
		TemplateContent: `Audit vendor freight invoice line items against approved quotation rates.
Flag unauthorized charges, rate variance, or currency discrepancies.
Return ONLY JSON with discrepancies, audit_verdict, and total_discrepancy_amount.`,
		InputVariables: []string{"InvoiceText", "QuotationRates"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 8. Leads Lead Scoring
	pm.register(&PromptDefinition{
		PromptKey:    "leads.lead_scoring",
		Version:      "1.0.0",
		WorkflowName: "leads",
		Purpose:      "Evaluates commercial fit and shipping volume potential for prospective leads.",
		TemplateContent: `You are an expert sales analyst. We have a new lead.
Company Name: {{.CompanyName}}
Industry: {{.Industry}}
Estimated Revenue: {{.EstimatedRevenue}}
Employee Count: {{.EmployeeCount}}
Shipping Volume: {{.MonthlyShippingVolume}} TEUs/month
Top Suppliers: {{.TopSuppliers}}
Is Exporter: {{.IsExporter}}

Please analyze this company and return ONLY a JSON object with two fields:
1. "score" (integer 0-100, where 100 is a perfect fit for a logistics/freight company to sell to)
2. "research_report" (a short 3-sentence summary of why you gave this score and how we should approach them)`,
		InputVariables: []string{"CompanyName", "Industry", "EstimatedRevenue", "EmployeeCount", "MonthlyShippingVolume", "TopSuppliers", "IsExporter"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 9. Outreach Cold Email
	pm.register(&PromptDefinition{
		PromptKey:    "outreach.cold_email",
		Version:      "1.0.0",
		WorkflowName: "outreach",
		Purpose:      "Generates personalized B2B cold outreach email subject lines and bodies.",
		TemplateContent: `You are an expert B2B sales copywriter for Freel, a modern AI-powered freight forwarding platform.
Write a short, personalized cold outreach email to the following company:

Company Name: {{.CompanyName}}
Industry: {{.Industry}}
Goal of this email: {{.Goal}}

Requirements:
- Keep the subject line under 60 characters.
- Keep the email body to 3 short paragraphs maximum.
- Mention Freel by name and tie it to the company's specific logistics or supply chain needs.
- End with a clear, low-friction call to action.

Return ONLY a JSON object with exactly two fields:
1. "subject" - the email subject line (string)
2. "body" - the full email body (string, can include \n for newlines)`,
		InputVariables: []string{"CompanyName", "Industry", "Goal"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	// 10. RFQ Extract Shipment Request
	pm.register(&PromptDefinition{
		PromptKey:    "rfq.extract_shipment_request",
		Version:      "1.0.0",
		WorkflowName: "rfq",
		Purpose:      "Parses unstructured shipment requests into structured RFQ data.",
		TemplateContent: `You are an expert freight forwarding AI assistant. A customer has sent us a request for a shipping quotation. 
Extract the shipment details from the following raw text (which might be an email, a PDF OCR dump, or a WhatsApp message).

Raw Text:
"""
{{.RawText}}
"""

Requirements:
- Find the Origin and Destination (city, port, or country).
- Find the Cargo Weight and Volume (if mentioned).
- Find the Incoterms (e.g., FOB, EXW, CIF) if mentioned.
- Assess how confident you are in your extraction from 0 to 100.
- List any critical missing fields that a human would need to ask the customer for (e.g., "Cargo Weight", "Incoterms", "Target Date").

Return ONLY a JSON object with exactly three fields:
1. "data" - A JSON object containing: "origin" (string), "destination" (string), "weight" (string), "volume" (string), "incoterms" (string). Use null for missing values.
2. "confidence_score" - Integer 0-100.
3. "missing_fields" - Array of strings listing the names of fields you could not find but are necessary for a freight quote.`,
		InputVariables: []string{"RawText"},
		OutputFormat:   "JSON",
		IsActive:       true,
	})

	return pm
}

func (p *promptManager) register(defn *PromptDefinition) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.prompts[defn.PromptKey] == nil {
		p.prompts[defn.PromptKey] = make(map[string]*PromptDefinition)
	}
	p.prompts[defn.PromptKey][defn.Version] = defn
}

func (p *promptManager) resolveKey(key string) string {
	if canonical, exists := p.aliases[key]; exists {
		return canonical
	}
	return key
}

// GetPrompt formats the prompt using the default version "1.0.0".
func (p *promptManager) GetPrompt(promptKey string, variables map[string]interface{}) (string, error) {
	return p.GetPromptWithVersion(promptKey, "1.0.0", variables)
}

// GetPromptWithVersion formats the prompt for a specific version.
func (p *promptManager) GetPromptWithVersion(promptKey string, version string, variables map[string]interface{}) (string, error) {
	defn, err := p.GetPromptDefinition(promptKey, version)
	if err != nil {
		return "", err
	}

	tmpl, err := template.New(defn.PromptKey).Parse(defn.TemplateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse prompt template '%s': %w", defn.PromptKey, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to execute prompt template '%s': %w", defn.PromptKey, err)
	}

	return buf.String(), nil
}

// GetPromptDefinition retrieves the prompt definition by key and version.
func (p *promptManager) GetPromptDefinition(promptKey string, version string) (*PromptDefinition, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	resolvedKey := p.resolveKey(promptKey)
	versions, exists := p.prompts[resolvedKey]
	if !exists {
		return nil, fmt.Errorf("prompt key '%s' not found in prompt registry", promptKey)
	}

	if version == "" {
		version = "1.0.0"
	}

	defn, exists := versions[version]
	if !exists {
		// Return latest version available if specific version missing
		for _, d := range versions {
			return d, nil
		}
		return nil, fmt.Errorf("version '%s' for prompt '%s' not found", version, promptKey)
	}

	return defn, nil
}

// ListPrompts lists all registered prompt definitions across all workflows.
func (p *promptManager) ListPrompts() []*PromptDefinition {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*PromptDefinition
	for _, versions := range p.prompts {
		for _, d := range versions {
			result = append(result, d)
		}
	}
	return result
}
