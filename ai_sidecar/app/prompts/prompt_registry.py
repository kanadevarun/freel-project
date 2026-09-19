"""
prompt_registry.py — Authoritative Prompt Management & Versioning System for LogisticsHQ AI Workflows

Supports:
- Canonical prompt keys (e.g. pricing.analyst, sales.email_classification, contracts.rate_extraction)
- Semantic versioning (e.g. 1.0.0)
- Scoped workflow isolation (pricing, sales, operations, contracts, compliance, finance, leads, outreach)
- Variable interpolation and validation
- Clean integration with MariaDB ai_prompt_templates table with robust in-memory fallback
"""

import re
from typing import Dict, Any, Optional, List

class PromptDefinition:
    def __init__(
        self,
        prompt_key: str,
        version: str,
        workflow_name: str,
        purpose: str,
        template: str,
        input_variables: Optional[List[str]] = None,
        output_format: str = "JSON",
        is_active: bool = True,
    ):
        self.prompt_key = prompt_key
        self.version = version
        self.workflow_name = workflow_name
        self.purpose = purpose
        self.template = template
        self.input_variables = input_variables or []
        self.output_format = output_format
        self.is_active = is_active

    def format(self, variables: Optional[Dict[str, Any]] = None) -> str:
        """Formats template replacing both {{.Var}} and {{Var}} style variables."""
        vars_dict = variables or {}
        text = self.template
        for k, v in vars_dict.items():
            val_str = str(v) if v is not None else ""
            text = text.replace(f"{{{{.{k}}}}}", val_str)
            text = text.replace(f"{{{{{k}}}}}", val_str)
        return text


# Authoritative Prompt Registry mapping (prompt_key, version) -> PromptDefinition
CANONICAL_PROMPTS: Dict[str, Dict[str, PromptDefinition]] = {}


def register_prompt(definition: PromptDefinition) -> None:
    if definition.prompt_key not in CANONICAL_PROMPTS:
        CANONICAL_PROMPTS[definition.prompt_key] = {}
    CANONICAL_PROMPTS[definition.prompt_key][definition.version] = definition


# ─────────────────────────────────────────────────────────────────────────────
# Pre-registered Standard V1.0.0 Prompts
# ─────────────────────────────────────────────────────────────────────────────

# 1. Pricing Analyst
register_prompt(PromptDefinition(
    prompt_key="pricing.analyst",
    version="1.0.0",
    workflow_name="pricing",
    purpose="Autonomous Senior Pricing Analyst Agent system prompt to evaluate RFQ carrier rates, apply markup rules, and recommend optimal quotation options.",
    template="""You are a Senior Pricing Analyst Agent for a Freight Forwarder.
Your goal is to recommend the best quotation options (up to 3 carrier quote options) for an RFQ based on:
1. Candidate rates (contract and spot rates returned by search_rates_tool)
2. Active markup and minimum margin rules (returned by get_pricing_rules_tool)
3. Live port congestion, carrier General Rate Increases (GRIs), or seasonal surcharges (using the search tool to verify if needed).

To compute the sell price:
- Apply the appropriate markup percentage (from pricing rules) to the buy price. E.g. if markup is 20%, sell = buy * 1.20.
- Verify that the resulting margin (Sell - Buy) / Sell meets or exceeds the min_margin_pct defined in the rules.
- If multiple rules match, prioritize LANE rules over DEFAULT rules, and CUSTOMER_TIER rules. Higher priority values always apply first.

When you are ready to conclude and return the recommended options:
Your final AIMessage MUST contain a JSON block representing the suggested quotes list in the following format:
```json
[
  {
    "carrier_name": "Maersk",
    "transit_time_days": 14,
    "buy_price": 2800.00,
    "sell_price": 3360.00,
    "is_recommended": true,
    "reliability_score": 92,
    "historical_success_rate": 0.95,
    "ai_reasoning": "Standard contract rate. Safe transit duration."
  }
]
```
Also, summarize your overall reasoning outside the JSON block.""",
    input_variables=[],
    output_format="JSON"
))

# 2. Sales Email Classification & Parsing
register_prompt(PromptDefinition(
    prompt_key="sales.email_classification",
    version="1.0.0",
    workflow_name="sales",
    purpose="Analyzes inbound customer emails, classifies intent and sentiment, and extracts structured shipping requirements.",
    template="""Analyze the following inbound email from a shipper and extract key metadata:

Email From: {{.Sender}}
Subject: {{.Subject}}
Body:
{{.Body}}
{{.PriorContext}}

Extract the following fields and return ONLY a JSON object:
- intent: "RFQ_REQUEST" (if requesting a shipping quote), "QUESTION" (general inquiry), "MEETING" (scheduling), "UNSUBSCRIBE" (opt-out), "FOLLOW_UP" (reply to existing thread).
- sentiment: "POSITIVE", "NEUTRAL", "NEGATIVE".
- confidence: integer from 0 to 100 representing your confidence in intent classification.
- lead_name: The sender's name if signed or mentioned, or null.
- company_domain: The domain of the sender (e.g. extract from from_email like 'tataexports.com', but exclude generic domains like 'gmail.com', 'yahoo.com', 'outlook.com', etc.).
- origin_port: The name of the origin port or city mentioned (e.g., "Nhava Sheva", "Mumbai", "INNSA"), or null.
- destination_port: The name of the destination port or city mentioned (e.g., "Hamburg", "DEHAM"), or null.
- incoterms: The Incoterms code mentioned (e.g. FOB, CIF, EXW, FCA), or null.
- cargo_description: Text description of the commodity/goods, or null.
- cargo_weight: Weight in KG (as float), or null.
- cargo_volume: Volume in CBM (as float), or null.
- target_date: Cargo ready date if mentioned (formatted as YYYY-MM-DD). If relative date (e.g. "next month"), calculate approximate date from today ({{.Today}}), or null.
- ai_summary: A short, concise summary (1-2 sentences) of the email's request.

Response JSON:""",
    input_variables=["Sender", "Subject", "Body", "PriorContext", "Today"],
    output_format="JSON"
))

# 3. Sales Reply Draft
register_prompt(PromptDefinition(
    prompt_key="sales.reply_draft",
    version="1.0.0",
    workflow_name="sales",
    purpose="Drafts a polite, professional, and concise email reply to request missing mandatory quote details from an incomplete customer inquiry.",
    template="""Draft a polite, professional, and concise email reply to a customer who sent a quote request but missed some mandatory details.

Original Email Subject: {{.Subject}}
Original Email Body:
{{.Body}}

Today's Date: {{.Today}}
Missing Mandatory Fields that you MUST request: {{.MissingFields}}
{{.ReplyContext}}
Write only the email body. Do not include subject line or header fields. Start with the greeting "{{.Greeting}}" and sign off professionally as "LogisticsHQ Sales Team".""",
    input_variables=["Subject", "Body", "Today", "MissingFields", "ReplyContext", "Greeting"],
    output_format="TEXT"
))

# 4. Operations Tracking Parse
register_prompt(PromptDefinition(
    prompt_key="operations.tracking_parse",
    version="1.0.0",
    workflow_name="operations",
    purpose="Parses raw unstructured carrier tracking descriptions and extracts milestones, exceptions, and delay estimates.",
    template="""You are an expert freight logistics operations analyst.
Your job is to parse a raw carrier tracking update and extract structured information.

Given a raw description from a carrier update, extract:

1. milestones: A list of shipping milestones detected. Each entry must have:
   - code: one of BOOKED, DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED (map appropriately)
   - date: ISO8601 date string (e.g. "2026-08-15T00:00:00Z"), null if not mentioned
   - location: port or place name if mentioned, null otherwise
   - notes: brief note about this event

2. exception_signals: A list of anomalies detected. Each entry must have:
   - type: one of DELAY, ROLLOVER, CUSTOMS_HOLD, PORT_CONGESTION, WEATHER
   - detected: true/false
   - details: brief text explaining what triggered this signal
   - delay_hours: estimated hours of delay if applicable, 0 otherwise

3. needs_human_review: true if the update is ambiguous or contains a serious unrecognized event

4. ai_summary: A single plain English sentence summarizing what happened to this shipment.

Return ONLY a JSON object with these 4 keys. No markdown, no explanation.""",
    input_variables=[],
    output_format="JSON"
))

# 5. Contracts Rate Extraction
register_prompt(PromptDefinition(
    prompt_key="contracts.rate_extraction",
    version="1.0.0",
    workflow_name="contracts",
    purpose="Extracts structured ocean and air rate cards, surcharges, free days, and routing options from carrier service contracts and circulars.",
    template="""You are an expert freight forwarding operations executive. Read the following shipping contract text and extract all port-pair rates.

Document Text:
{{.RawText}}

Carrier SCAC: {{.CarrierSCAC}}
Carrier Name: {{.CarrierName}}

For each rate found, construct a JSON object matching this structure:
{
  "origin_port": "UN/LOCODE or port name (e.g. INNSA)",
  "destination_port": "UN/LOCODE or port name (e.g. DEHAM)",
  "via_port": "transshipment port LOCODE or null",
  "service_code": "route code or service name or null",
  "carrier_scac": "{{.CarrierSCAC}}",
  "carrier_name": "{{.CarrierName}}",
  "vessel_name": "vessel name if specified or null",
  "equipment_type": "40GP",
  "ocean_freight": <ocean freight base amount as float>,
  "origin_charges": <origin terminal handling fee as float>,
  "destination_charges": <destination terminal handling fee as float>,
  "surcharges": [
     {
       "code": "BAF|CAF|PSS|etc",
       "description": "fuel adjust factor or currency surcharge description",
       "amount": <surcharge amount as float>,
       "unit": "PER_TEU|PER_CONTAINER|PER_SHIPMENT",
       "included": true/false
     }
  ],
  "total_buy_price": <ocean_freight + origin_charges + destination_charges + all non-included surcharges>,
  "free_days_origin": <free detention days POL or 0>,
  "free_days_destination": <free demurrage days POD or 14>,
  "transit_days": <transit time in days as integer or null>,
  "incoterms": "FOB|CIF|DDP|EXW|null",
  "valid_from": "YYYY-MM-DD",
  "valid_until": "YYYY-MM-DD"
}

Return a JSON object with a single key "rates" containing a list of these objects:
{"rates": [...]}""",
    input_variables=["RawText", "CarrierSCAC", "CarrierName"],
    output_format="JSON"
))

# 6. Compliance Document Verification
register_prompt(PromptDefinition(
    prompt_key="compliance.doc_verification",
    version="1.0.0",
    workflow_name="compliance",
    purpose="Verifies shipping documents against master booking data to identify discrepancies in weights, container numbers, and seal codes.",
    template="""You are a trade compliance auditor. Compare the extracted shipping document metadata against master shipment records.
Identify discrepancies in:
1. Gross weight
2. Container numbers
3. Seal numbers
4. Shipper and consignee names

Document Content:
{{.document_text}}

Expected Data:
{{.expected_data}}

Return ONLY a JSON object:
{
  "discrepancies": [
    {
      "field": "field_name",
      "expected_value": "expected",
      "actual_value": "actual",
      "severity": "HIGH|MEDIUM|LOW",
      "recommendation": "suggested action"
    }
  ],
  "compliance_status": "COMPLIANT|DISCREPANCIES_FOUND",
  "summary": "Auditor summary"
}""",
    input_variables=["document_text", "expected_data"],
    output_format="JSON"
))

# 7. Finance Invoice Audit
register_prompt(PromptDefinition(
    prompt_key="finance.invoice_audit",
    version="1.0.0",
    workflow_name="finance",
    purpose="Audits vendor freight invoices against agreed quotation line items and checks for unauthorized surcharges or overbilling.",
    template="""You are an automated freight invoice reconciliation specialist.
Audit the following invoice line items against approved quotation rates and contract charges.
Flag any unauthorized charges, rate discrepancies, or incorrect currency conversions.

Carrier Invoice:
{{.invoice_text}}

Agreed Contract Rates:
{{.quotation_rates}}

Return ONLY a JSON object:
{
  "invoice_number": "...",
  "vendor_name": "...",
  "discrepancies": [
    {
      "charge_code": "...",
      "invoiced_amount": 0.0,
      "expected_amount": 0.0,
      "variance": 0.0,
      "status": "OPEN|RESOLVED"
    }
  ],
  "audit_verdict": "APPROVED|REQUIRES_HUMAN_REVIEW|REJECTED",
  "total_discrepancy_amount": 0.0
}""",
    input_variables=["invoice_text", "quotation_rates"],
    output_format="JSON"
))

# 8. Leads Lead Scoring
register_prompt(PromptDefinition(
    prompt_key="leads.lead_scoring",
    version="1.0.0",
    workflow_name="leads",
    purpose="Evaluates commercial fit and shipping potential for new inbound prospective shipper leads.",
    template="""You are an expert sales analyst. We have a new lead.
Company Name: {{.CompanyName}}
Industry: {{.Industry}}
Estimated Revenue: {{.EstimatedRevenue}}
Employee Count: {{.EmployeeCount}}
Shipping Volume: {{.MonthlyShippingVolume}} TEUs/month
Top Suppliers: {{.TopSuppliers}}
Is Exporter: {{.IsExporter}}

Please analyze this company and return ONLY a JSON object with two fields:
1. "score" (integer 0-100, where 100 is a perfect fit for a logistics/freight company to sell to)
2. "research_report" (a short 3-sentence summary of why you gave this score and how we should approach them)""",
    input_variables=["CompanyName", "Industry", "EstimatedRevenue", "EmployeeCount", "MonthlyShippingVolume", "TopSuppliers", "IsExporter"],
    output_format="JSON"
))

# 9. Outreach Cold Email
register_prompt(PromptDefinition(
    prompt_key="outreach.cold_email",
    version="1.0.0",
    workflow_name="outreach",
    purpose="Generates personalized B2B cold outreach email subject lines and bodies tailored to shipper supply chain profile.",
    template="""You are an expert B2B sales copywriter for Freel, a modern AI-powered freight forwarding platform.
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
2. "body" - the full email body (string, can include \\n for newlines)""",
    input_variables=["CompanyName", "Industry", "Goal"],
    output_format="JSON"
))

# 10. RFQ Extract Shipment Request
register_prompt(PromptDefinition(
    prompt_key="rfq.extract_shipment_request",
    version="1.0.0",
    workflow_name="rfq",
    purpose="Extracts structured shipment parameters from unstructured text messages, email bodies, and RFQ document snippets.",
    template="""You are an expert freight forwarding AI assistant. A customer has sent us a request for a shipping quotation. 
Extract the shipment details from the following raw text (which might be an email, a PDF OCR dump, or a WhatsApp message).

Raw Text:
\"\"\"
{{.RawText}}
\"\"\"

Requirements:
- Find the Origin and Destination (city, port, or country).
- Find the Cargo Weight and Volume (if mentioned).
- Find the Incoterms (e.g., FOB, EXW, CIF) if mentioned.
- Assess how confident you are in your extraction from 0 to 100.
- List any critical missing fields that a human would need to ask the customer for (e.g., "Cargo Weight", "Incoterms", "Target Date").

Return ONLY a JSON object with exactly three fields:
1. "data" - A JSON object containing: "origin" (string), "destination" (string), "weight" (string), "volume" (string), "incoterms" (string). Use null for missing values.
2. "confidence_score" - Integer 0-100.
3. "missing_fields" - Array of strings listing the names of fields you could not find but are necessary for a freight quote.""",
    input_variables=["RawText"],
    output_format="JSON"
))


def get_prompt_definition(prompt_key: str, version: Optional[str] = None) -> PromptDefinition:
    """Retrieves a PromptDefinition by key and version, failing loudly if not found."""
    if prompt_key not in CANONICAL_PROMPTS:
        raise KeyError(f"Prompt '{prompt_key}' not found in prompt registry.")
    versions = CANONICAL_PROMPTS[prompt_key]
    ver = version or "1.0.0"
    if ver not in versions:
        # Fall back to latest available version if requested version missing
        latest = sorted(versions.keys())[-1]
        print(f"[Prompt Registry] Notice: Requested version '{ver}' for '{prompt_key}' not found; using '{latest}'")
        ver = latest
    return versions[ver]


def get_prompt(prompt_key: str, version: Optional[str] = None, variables: Optional[Dict[str, Any]] = None) -> str:
    """Retrieves and formats a registered prompt template."""
    defn = get_prompt_definition(prompt_key, version)
    return defn.format(variables)


def list_prompts_for_workflow(workflow_name: str) -> List[PromptDefinition]:
    """Returns all prompt definitions belonging to a specific workflow."""
    matches = []
    for versions in CANONICAL_PROMPTS.values():
        for defn in versions.values():
            if defn.workflow_name.lower() == workflow_name.lower():
                matches.append(defn)
    return matches


def list_registered_prompts() -> List[PromptDefinition]:
    """Returns all registered prompt definitions across all workflows."""
    matches = []
    for versions in CANONICAL_PROMPTS.values():
        for defn in versions.values():
            matches.append(defn)
    return matches


class PromptRegistry:
    """Convenience class providing static access to prompt registry."""
    @staticmethod
    def get_prompt(prompt_key: str, version: Optional[str] = None, variables: Optional[Dict[str, Any]] = None) -> str:
        return get_prompt(prompt_key, version, variables)

    @staticmethod
    def get_definition(prompt_key: str, version: Optional[str] = None) -> PromptDefinition:
        return get_prompt_definition(prompt_key, version)

    @staticmethod
    def get_prompt_metadata(prompt_key: str, version: Optional[str] = None) -> Dict[str, Any]:
        defn = get_prompt_definition(prompt_key, version)
        return {
            "prompt_key": defn.prompt_key,
            "version": defn.version,
            "workflow": defn.workflow_name,
            "purpose": defn.purpose,
            "input_variables": defn.input_variables,
            "output_format": defn.output_format,
            "is_active": defn.is_active,
        }

    @staticmethod
    def list_by_workflow(workflow_name: str) -> List[PromptDefinition]:
        return list_prompts_for_workflow(workflow_name)

