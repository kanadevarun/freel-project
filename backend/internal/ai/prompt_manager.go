package ai

import (
	"bytes"
	"fmt"
	"text/template"
)

// PromptManager handles fetching and formatting standard prompts.
type PromptManager interface {
	// GetPrompt fetches a template and fills it with your data.
	// Simple meaning: It grabs a pre-written fill-in-the-blank prompt and inserts the specific details.
	// Example: text, err := promptManager.GetPrompt("lead_score", map[string]string{"company": "Acme"})
	GetPrompt(templateName string, variables map[string]interface{}) (string, error)
}

type promptManager struct {
	templates map[string]string
}

// NewPromptManager creates a new prompt manager.
// Simple meaning: Sets up a library of standard prompts we can reuse.
// Example: pm := NewPromptManager()
func NewPromptManager() PromptManager {
	templates := map[string]string{
		// ── score_lead ──────────────────────────────────────────────────────────────
		// Used by the LeadWorker background job to score a newly created lead.
		// Variables: CompanyName, Industry, EstimatedRevenue, EmployeeCount,
		//            MonthlyShippingVolume, TopSuppliers, IsExporter
		"score_lead": `
You are an expert sales analyst. We have a new lead.
Company Name: {{.CompanyName}}
Industry: {{.Industry}}
Estimated Revenue: {{.EstimatedRevenue}}
Employee Count: {{.EmployeeCount}}
Shipping Volume: {{.MonthlyShippingVolume}} TEUs/month
Top Suppliers: {{.TopSuppliers}}
Is Exporter: {{.IsExporter}}

Please analyze this company and return ONLY a JSON object with two fields:
1. "score" (integer 0-100, where 100 is a perfect fit for a logistics/freight company to sell to)
2. "research_report" (a short 3-sentence summary of why you gave this score and how we should approach them)
`,

		// ── generate_email ──────────────────────────────────────────────────────────
		// Used by the Outreach module to generate personalized cold emails.
		// Simple meaning: We give the AI context about the target company, and it
		// writes a professional, personalized subject + body for a cold outreach email.
		// Variables: CompanyName, Industry, Goal
		// Expected response: JSON with "subject" and "body" fields.
		"generate_email": `
You are an expert B2B sales copywriter for Freel, a modern AI-powered freight forwarding platform.
Write a short, personalized cold outreach email to the following company:

Company Name: {{.CompanyName}}
Industry: {{.Industry}}
Goal of this email: {{.Goal}}

Requirements:
- Keep the subject line under 60 characters.
- Keep the email body to 3 short paragraphs maximum.
- Mention Freel by name and tie it to the company's specific logistics or supply chain needs.
- End with a clear, low-friction call to action (e.g., "Would a 15-minute call work this week?").
- Professional but conversational tone. No buzzwords.

Return ONLY a JSON object with exactly two fields:
1. "subject" - the email subject line (string)
2. "body" - the full email body (string, can include \n for newlines)
`,
		// ── extract_shipment_request ────────────────────────────────────────────────
		// Used by the RFQ module to parse unstructured requests (emails, OCR) into structured RFQ data.
		// Instructs the LLM to extract key fields and explicitly state what it couldn't find.
		// Variables: RawText
		// Expected response: JSON with "data" (origin, destination, weight, dimensions, etc.), "confidence_score", "missing_fields".
		"extract_shipment_request": `
You are an expert freight forwarding AI assistant. A customer has sent us a request for a shipping quotation. 
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
3. "missing_fields" - Array of strings listing the names of fields you could not find but are necessary for a freight quote.
`,
		// ── pricing_analyst ────────────────────────────────────────────────
		// Used by the autonomous Pricing Agent.
		// Variables: Origin, Destination, Incoterms, TargetDate, Items, CarrierRates
		// Expected response: JSON containing structured AIRecommendations and a Draft Quote suggestion.
		"pricing_analyst": `
You are a Senior Pricing Analyst at a freight forwarding company with 15+ years of experience.
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
2. "draft_quote": An object with: "carrier_name" (string), "buy_price" (float), "sell_price" (float), "transit_time_days" (int), "reliability_score" (int), "historical_success_rate" (float)
`,
	}
	return &promptManager{
		templates: templates,
	}
}

// GetPrompt fetches a template and fills it with your data.
// Simple meaning: It grabs a pre-written fill-in-the-blank prompt and inserts the specific details.
// Example: text, err := pm.GetPrompt("lead_score", map[string]interface{}{"CompanyName": "Acme"})
func (p *promptManager) GetPrompt(templateName string, variables map[string]interface{}) (string, error) {
	tmplStr, exists := p.templates[templateName]
	if !exists {
		return "", fmt.Errorf("prompt template %s not found", templateName)
	}

	tmpl, err := template.New(templateName).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}
