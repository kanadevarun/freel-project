package rfq

// requirements_engine.go — Task 10: Pure Deterministic Requirements Evaluation
//
// EvaluateRequirements receives a fully-loaded *spec.RFQ (items + quotes already
// joined by GetRFQByID) and returns a structured GetRequirementsResponse.
//
// Design rules:
//   1. ZERO database calls — the engine is a pure function: RFQ in, response out.
//   2. ZERO invented schema fields — every field read here maps to a real DB column.
//   3. AI findings derive only from real fields: rfqs.agent_status, rfqs.lead_id,
//      lead_interactions.ai_confidence (passed in via the RFQ struct's loaded items).
//   4. Document requirements are stage-aware: HBL, MBL, Air Waybill, and customs
//      declarations are NOT_APPLICABLE at the RFQ quotation stage and will NOT
//      appear as blockers.
//   5. Deterministic requirements are the single source of truth. AI findings
//      supplement but NEVER override deterministic rulings.

import (
	"fmt"
	"strings"

	"github.com/freel/backend/internal/rfq/spec"
)

// EvaluateRequirements is the top-level entry point for the requirements engine.
// It accepts the fully-loaded RFQ struct (items embedded) and returns the complete
// requirements evaluation — no I/O, no side effects.
func EvaluateRequirements(rfq *spec.RFQ) *spec.GetRequirementsResponse {
	groups := []spec.RequirementGroup{
		evalShipmentInfo(rfq),
		evalCustomerInfo(rfq),
		evalCargoOperational(rfq),
		evalConditionalCompliance(rfq),
	}

	docReqs := evalDocumentRequirements(rfq)
	docGroup := buildDocumentGroup(docReqs)
	groups = append(groups, docGroup)

	aiFindings := evalAIFindings(rfq)
	if len(aiFindings) > 0 {
		groups = append(groups, buildAIFindingsGroup(aiFindings))
	}

	readiness := calculateReadiness(groups, docReqs)
	readiness.NextBestAction = deriveNextBestAction(readiness)

	return &spec.GetRequirementsResponse{
		OperationalReadiness: readiness,
		Groups:               groups,
		DocumentRequirements: docReqs,
		AIFindings:           aiFindings,
		LeadID:               rfq.LeadID,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Category A — Shipment Information
// Maps to real DB columns: rfqs.origin, destination, incoterms, target_date
// and rfq_items.description, weight_kg, volume_cbm
// ─────────────────────────────────────────────────────────────────────────────

func evalShipmentInfo(rfq *spec.RFQ) spec.RequirementGroup {
	reqs := []spec.Requirement{}

	// rfqs.origin
	reqs = append(reqs, evalField("origin", "Origin Port", spec.SeverityBlocking,
		strPtrVal(rfq.Origin), "Port of loading (POL) is required for route determination and carrier rate lookup."))

	// rfqs.destination
	reqs = append(reqs, evalField("destination", "Destination Port", spec.SeverityBlocking,
		strPtrVal(rfq.Destination), "Port of discharge (POD) is required for route determination and carrier rate lookup."))

	// rfqs.incoterms
	reqs = append(reqs, evalField("incoterms", "Incoterms", spec.SeverityBlocking,
		strPtrVal(rfq.Incoterms), "Incoterms determine cost and risk allocation between buyer and seller."))

	// rfqs.target_date
	targetDateVal := ""
	if rfq.TargetDate != nil {
		targetDateVal = rfq.TargetDate.Format("02 Jan 2006")
	}
	reqs = append(reqs, evalField("target_date", "Cargo Ready Date", spec.SeverityBlocking,
		targetDateVal, "Cargo ready date is required for vessel schedule and space allocation."))

	// rfq_items.description (any item)
	cargoDesc := cargoDescription(rfq.Items)
	reqs = append(reqs, evalField("cargo_description", "Cargo Description", spec.SeverityBlocking,
		cargoDesc, "Cargo description is required for commodity classification and carrier acceptance."))

	// rfq_items.weight_kg (sum > 0)
	totalWeight := totalWeightKG(rfq.Items)
	weightVal := ""
	if totalWeight > 0 {
		weightVal = fmt.Sprintf("%.0f KG", totalWeight)
	}
	reqs = append(reqs, evalField("cargo_weight", "Cargo Weight", spec.SeverityBlocking,
		weightVal, "Gross weight is required for carrier booking and rate calculation."))

	// rfq_items.volume_cbm (sum > 0)
	totalVolume := totalVolumeCBM(rfq.Items)
	volumeVal := ""
	if totalVolume > 0 {
		volumeVal = fmt.Sprintf("%.0f CBM", totalVolume)
	}
	reqs = append(reqs, evalField("cargo_volume", "Cargo Volume", spec.SeverityBlocking,
		volumeVal, "Volume in CBM is required for LCL rating and FCL space confirmation."))

	return buildGroup(spec.CategoryShipmentInfo, "Shipment Information", "🚢", reqs)
}

// ─────────────────────────────────────────────────────────────────────────────
// Category B — Customer Information
// Maps to: customers.name (via rfq.CustomerName), customers.contact_name,
//           customers.contact_email, customers.contact_phone
//           leads.email, leads.phone (via rfq.CustomerEmail/Phone/ContactName)
// ─────────────────────────────────────────────────────────────────────────────

func evalCustomerInfo(rfq *spec.RFQ) spec.RequirementGroup {
	reqs := []spec.RequirementGroup{}
	_ = reqs

	items := []spec.Requirement{}

	// CustomerName: populated via JOIN with customers / leads in GetRFQByID
	items = append(items, evalField("customer_name", "Customer Identified", spec.SeverityRequired,
		rfq.CustomerName, "Customer must be linked before proceeding with quotation workflow."))

	// CustomerContactName: COALESCE(customers.contact_name, leads.contact_name)
	contactName := ""
	if rfq.CustomerContactName != nil {
		contactName = *rfq.CustomerContactName
	}
	items = append(items, evalField("customer_contact", "Primary Contact Person", spec.SeverityRequired,
		contactName, "A named contact person is required for communication and document submission."))

	// CustomerEmail: COALESCE(customers.contact_email, leads.email)
	emailVal := ""
	if rfq.CustomerEmail != nil {
		emailVal = *rfq.CustomerEmail
	}
	items = append(items, evalField("customer_email", "Contact Email", spec.SeverityRequired,
		emailVal, "Contact email is required for quotation delivery and correspondence."))

	// CustomerPhone: COALESCE(customers.contact_phone, leads.phone) — OPTIONAL
	phoneVal := ""
	if rfq.CustomerPhone != nil {
		phoneVal = *rfq.CustomerPhone
	}
	items = append(items, evalFieldSeverity("customer_phone", "Contact Phone", spec.SeverityOptional,
		phoneVal, "Phone is recommended for urgent shipment coordination."))

	// Inject source context for lead-originated RFQs
	if rfq.LeadID != nil {
		for i := range items {
			if items[i].Status == spec.ReqStatusSatisfied && items[i].SourceContext == "" {
				items[i].SourceContext = fmt.Sprintf("From Lead #%d", *rfq.LeadID)
			}
		}
	}

	return buildGroup(spec.CategoryCustomerInfo, "Customer Information", "👤", items)
}

// ─────────────────────────────────────────────────────────────────────────────
// Category C — Cargo & Operational Requirements
// Maps to: rfq_items (count > 0), weight_kg, volume_cbm, quantity
// ─────────────────────────────────────────────────────────────────────────────

func evalCargoOperational(rfq *spec.RFQ) spec.RequirementGroup {
	items := []spec.Requirement{}

	// rfq_items must exist
	itemCountVal := ""
	if len(rfq.Items) > 0 {
		itemCountVal = fmt.Sprintf("%d cargo item(s) defined", len(rfq.Items))
	}
	items = append(items, evalField("cargo_items", "Cargo Items Defined", spec.SeverityRequired,
		itemCountVal, "At least one cargo item must be defined with description, weight, and volume."))

	// Weight confirmed (rfq_items.weight_kg sum > 0)
	totalWeight := totalWeightKG(rfq.Items)
	weightConfirmed := ""
	if totalWeight > 0 {
		weightConfirmed = fmt.Sprintf("%.0f KG (total gross)", totalWeight)
	}
	items = append(items, evalField("weight_confirmed", "Cargo Weight Confirmed", spec.SeverityRequired,
		weightConfirmed, "Gross weight per commodity line must be confirmed."))

	// Volume confirmed (rfq_items.volume_cbm sum > 0)
	totalVolume := totalVolumeCBM(rfq.Items)
	volumeConfirmed := ""
	if totalVolume > 0 {
		volumeConfirmed = fmt.Sprintf("%.0f CBM (total)", totalVolume)
	}
	items = append(items, evalField("volume_confirmed", "Cargo Volume Confirmed", spec.SeverityRequired,
		volumeConfirmed, "Cargo volume in CBM must be confirmed for space planning."))

	// Transport mode: derived from rfqs.incoterms (no separate transport_mode column)
	// The stage and incoterms together imply the mode. Actual mode comes from carrier context.
	// We show it as informational — not blocking since it can be inferred.
	modeHint := deriveTransportModeHint(rfq)
	items = append(items, spec.Requirement{
		ID:          "transport_mode",
		Category:    spec.CategoryCargoOperational,
		Type:        "transport_mode",
		Title:       "Transport Mode",
		Description: "Transport mode determines carrier network, document set, and applicable compliance rules.",
		Status:      spec.ReqStatusSatisfied,
		Severity:    spec.SeverityInformational,
		Value:       modeHint,
	})

	return buildGroup(spec.CategoryCargoOperational, "Cargo & Operational Requirements", "📦", items)
}

// ─────────────────────────────────────────────────────────────────────────────
// Category D — Conditional / Compliance Requirements
// Rules trigger only based on actual cargo description and incoterms values.
// Maps to: rfq_items.description (DG keywords), rfqs.incoterms
// ─────────────────────────────────────────────────────────────────────────────

func evalConditionalCompliance(rfq *spec.RFQ) spec.RequirementGroup {
	items := []spec.Requirement{}

	desc := strings.ToLower(cargoDescription(rfq.Items))

	// DG classification: only triggers if cargo description contains DG-related terms
	// from rfq_items.description — deterministic string scan, no AI required
	dgKeywords := []string{"dangerous", "hazardous", "flammable", "explosive", "corrosive", "toxic", "radioactive", "oxidizer", "dg class", "imdg", "msds", "sds"}
	isDGRisk := false
	for _, kw := range dgKeywords {
		if strings.Contains(desc, kw) {
			isDGRisk = true
			break
		}
	}

	dgReq := spec.Requirement{
		ID:            "dg_declaration",
		Category:      spec.CategoryConditionalCompliance,
		Type:          "dg_declaration",
		Title:         "Dangerous Goods Declaration",
		Description:   "Required when cargo contains or may contain IMDG-classified dangerous goods.",
		IsConditional: true,
	}
	if isDGRisk {
		dgReq.Status = spec.ReqStatusMissing
		dgReq.Severity = spec.SeverityConditional
		dgReq.ConditionReason = "Cargo description contains terminology associated with regulated materials. Please confirm DG classification."
	} else {
		dgReq.Status = spec.ReqStatusNotApplicable
		dgReq.Severity = spec.SeverityConditional
		dgReq.ConditionReason = "No dangerous goods indicators detected in cargo description."
		dgReq.Value = "Not applicable — cargo not flagged as DG"
	}
	items = append(items, dgReq)

	// Temperature-controlled / Reefer: triggers on reefer keywords in cargo description
	reeferKeywords := []string{"reefer", "refrigerat", "frozen", "chilled", "cold chain", "temperature controlled", "perishable", "pharma"}
	isReefer := false
	for _, kw := range reeferKeywords {
		if strings.Contains(desc, kw) {
			isReefer = true
			break
		}
	}

	reeferReq := spec.Requirement{
		ID:            "temperature_requirements",
		Category:      spec.CategoryConditionalCompliance,
		Type:          "temperature_requirements",
		Title:         "Temperature Control Requirements",
		Description:   "Required for reefer/refrigerated shipments — defines setpoint, humidity, ventilation.",
		IsConditional: true,
	}
	if isReefer {
		reeferReq.Status = spec.ReqStatusMissing
		reeferReq.Severity = spec.SeverityConditional
		reeferReq.ConditionReason = "Cargo description indicates temperature-sensitive cargo. Temperature setpoint must be confirmed."
	} else {
		reeferReq.Status = spec.ReqStatusNotApplicable
		reeferReq.Severity = spec.SeverityConditional
		reeferReq.ConditionReason = "No reefer indicators detected."
		reeferReq.Value = "Not applicable — ambient shipment"
	}
	items = append(items, reeferReq)

	// Container type confirmation: CONDITIONAL for FCL incoterms context
	// rfqs.incoterms: if FOB/CIF/CFR/DAP/DDP — typically FCL, container confirmation useful
	incoterms := strings.ToUpper(strPtrVal(rfq.Incoterms))
	fclIncoterms := map[string]bool{"FOB": true, "CIF": true, "CFR": true, "DAP": true, "DDP": true, "FCA": true, "CPT": true, "CIP": true}
	isFCLLikely := fclIncoterms[incoterms]

	containerReq := spec.Requirement{
		ID:            "container_type",
		Category:      spec.CategoryConditionalCompliance,
		Type:          "container_type",
		Title:         "Container Type Confirmation",
		Description:   "FCL container type (20GP, 40GP, 40HC) must be confirmed for ocean freight booking.",
		IsConditional: true,
	}
	if isFCLLikely {
		// We don't have a container_type column in the schema; it would come from rfq_items context.
		// Mark as UNDER_REVIEW — informational prompt, not a blocker.
		containerReq.Status = spec.ReqStatusUnderReview
		containerReq.Severity = spec.SeverityConditional
		containerReq.ConditionReason = fmt.Sprintf("Incoterms %s typically implies FCL shipment. Container type/size should be confirmed.", incoterms)
		containerReq.Value = "Pending confirmation"
	} else {
		containerReq.Status = spec.ReqStatusNotApplicable
		containerReq.Severity = spec.SeverityConditional
		containerReq.ConditionReason = "Container type confirmation not applicable for current Incoterms."
		containerReq.Value = "Not applicable"
	}
	items = append(items, containerReq)

	return buildGroup(spec.CategoryConditionalCompliance, "Conditional / Compliance Requirements", "⚠️", items)
}

// ─────────────────────────────────────────────────────────────────────────────
// Category E — Document Requirements (Stage-Aware)
// Commercial Invoice and Packing List are REQUIRED at RFQ stage.
// Bill of Lading, HBL/MBL, Air Waybill → NOT_APPLICABLE at RFQ stage.
// ─────────────────────────────────────────────────────────────────────────────

func evalDocumentRequirements(rfq *spec.RFQ) []spec.DocumentRequirement {
	docs := []spec.DocumentRequirement{
		// ── Current Stage (Required at RFQ / Quotation stage) ──────────────
		{
			DocType:         spec.ReqDocCommercialInvoice,
			Title:           "Commercial Invoice",
			Status:          spec.ReqStatusMissing, // Evaluated against real rfq.Documents below
			ApplicableStage: spec.DocStageRFQ,
			IsRequired:      true,
			IsConditional:   false,
			Reason:          "Required before shipment processing and customs clearance.",
		},
		{
			DocType:         spec.ReqDocPackingList,
			Title:           "Packing List",
			Status:          spec.ReqStatusMissing, // Evaluated against real rfq.Documents below
			ApplicableStage: spec.DocStageRFQ,
			IsRequired:      true,
			IsConditional:   false,
			Reason:          "Required before shipment processing for customs and carrier acceptance.",
		},
		{
			DocType:         spec.ReqDocProformaInvoice,
			Title:           "Proforma Invoice",
			Status:          spec.ReqStatusNotApplicable,
			ApplicableStage: spec.DocStageQuotation,
			IsRequired:      false,
			IsConditional:   true,
			Reason:          "Required if advance payment or LC terms apply.",
		},

		// ── Future Stage (NOT applicable at current RFQ stage) ─────────────
		// These must NOT appear as blockers during quotation.
		{
			DocType:         spec.ReqDocBillOfLading,
			Title:           "Bill of Lading (OBL)",
			Status:          spec.ReqStatusNotApplicable,
			ApplicableStage: spec.DocStageShipmentExecution,
			IsRequired:      false,
			IsConditional:   false,
			Reason:          "Generated at shipment execution stage after carrier confirmation.",
		},
		{
			DocType:         spec.ReqDocHBL,
			Title:           "House Bill of Lading (HBL)",
			Status:          spec.ReqStatusNotApplicable,
			ApplicableStage: spec.DocStageBookingConfirmed,
			IsRequired:      false,
			IsConditional:   true,
			Reason:          "Applicable based on freight forwarding structure (house BL issued by freight forwarder).",
		},
		{
			DocType:         spec.ReqDocMBL,
			Title:           "Master Bill of Lading (MBL)",
			Status:          spec.ReqStatusNotApplicable,
			ApplicableStage: spec.DocStageBookingConfirmed,
			IsRequired:      false,
			IsConditional:   true,
			Reason:          "Applicable based on carrier and forwarding structure (master BL from shipping line).",
		},
	}

	// Conditional: Certificate of Origin — required for certain trade lanes.
	// We check origin + destination for well-known preferential trade routes.
	// This is a simplified rule — real trade lane rules belong in a future trade_intel service.
	origin := strings.ToUpper(strPtrVal(rfq.Origin))
	dest := strings.ToUpper(strPtrVal(rfq.Destination))

	// Rough heuristic: India-EU routes often require CoO for preferential duties
	needsCoO := (strings.Contains(origin, "INNSA") || strings.Contains(origin, "INDIA") || strings.Contains(origin, "INMUN") || strings.Contains(origin, "INCCU")) &&
		(strings.Contains(dest, "DE") || strings.Contains(dest, "NL") || strings.Contains(dest, "EU") || strings.Contains(dest, "GB"))

	coOStatus := spec.ReqStatusNotApplicable
	coOReason := "Certificate of Origin requirement depends on trade lane and preferential duty application."
	if needsCoO {
		coOStatus = spec.ReqStatusMissing
		coOReason = "India to EU/UK routes may require Certificate of Origin for preferential import duty application. Confirm with customer."
	}
	docs = append(docs, spec.DocumentRequirement{
		DocType:         spec.ReqDocCertificateOfOrigin,
		Title:           "Certificate of Origin",
		Status:          coOStatus,
		ApplicableStage: spec.DocStageShipmentExecution,
		IsRequired:      needsCoO,
		IsConditional:   true,
		Reason:          coOReason,
	})

	// Air Waybill: only relevant if transport mode is air (no air mode column in schema,
	// so we check origin/destination for known air cargo port patterns)
	// Conservative: mark as NOT_APPLICABLE unless specifically needed
	docs = append(docs, spec.DocumentRequirement{
		DocType:         spec.ReqDocAirWaybill,
		Title:           "Air Waybill (HAWB/MAWB)",
		Status:          spec.ReqStatusNotApplicable,
		ApplicableStage: spec.DocStageShipmentExecution,
		IsRequired:      false,
		IsConditional:   true,
		Reason:          "Applicable only for air freight shipments at execution stage.",
	})

	// DG Declaration: conditional on DG cargo keyword detection
	desc := strings.ToLower(cargoDescription(rfq.Items))
	isDGRisk := false
	for _, kw := range []string{"dangerous", "hazardous", "flammable", "explosive", "corrosive", "toxic"} {
		if strings.Contains(desc, kw) {
			isDGRisk = true
			break
		}
	}
	dgStatus := spec.ReqStatusNotApplicable
	if isDGRisk {
		dgStatus = spec.ReqStatusMissing
	}
	docs = append(docs, spec.DocumentRequirement{
		DocType:         spec.ReqDocDGDeclaration,
		Title:           "Dangerous Goods Declaration (MSDS/SDS)",
		Status:          dgStatus,
		ApplicableStage: spec.DocStageRFQ,
		IsRequired:      isDGRisk,
		IsConditional:   true,
		Reason:          "Required for IMDG-classified dangerous goods before carrier acceptance.",
	})


	// Resolve against real rfq.Documents (Task 12)
	if len(rfq.Documents) > 0 {
		docsByType := make(map[string]*spec.RFQDocument)
		for i := range rfq.Documents {
			d := &rfq.Documents[i]
			norm := NormalizeDocType(d.DocumentType)
			existing, ok := docsByType[norm]
			if !ok || isHigherPriorityStatus(d.Status, existing.Status) {
				docsByType[norm] = d
			}
		}

		for i := range docs {
			normReq := NormalizeDocType(docs[i].DocType)
			if matchedDoc, ok := docsByType[normReq]; ok {
				docs[i].DocumentID = &matchedDoc.ID
				docs[i].DocumentStatus = matchedDoc.Status
				docs[i].FileName = matchedDoc.FileName
				docs[i].FileURL = matchedDoc.FileURL
				docs[i].UploadedAt = matchedDoc.UploadedAt
				docs[i].ReviewedAt = matchedDoc.ReviewedAt

				switch matchedDoc.Status {
				case spec.DocStatusApproved:
					docs[i].Status = spec.ReqStatusSatisfied
				case spec.DocStatusUnderReview, spec.DocStatusUploaded:
					docs[i].Status = spec.ReqStatusUnderReview
				case spec.DocStatusRejected:
					docs[i].Status = spec.ReqStatusMissing
					if matchedDoc.RejectionReason != nil && *matchedDoc.RejectionReason != "" {
						docs[i].Reason = "Document rejected: " + *matchedDoc.RejectionReason
					}
				case spec.DocStatusExpired:
					docs[i].Status = spec.ReqStatusMissing
					docs[i].Reason = "Document has expired. Updated document required."
				case spec.DocStatusRequested:
					docs[i].Status = spec.ReqStatusMissing
					docs[i].Reason = "Document has been requested from customer."
				case spec.DocStatusNotRequired:
					docs[i].Status = spec.ReqStatusNotApplicable
				}
			}
		}
	}

	return docs
}


func buildDocumentGroup(docs []spec.DocumentRequirement) spec.RequirementGroup {
	reqs := make([]spec.Requirement, 0, len(docs))
	completeCount := 0
	for _, doc := range docs {
		status := doc.Status
		severity := spec.SeverityRequired
		if !doc.IsRequired {
			severity = spec.SeverityConditional
		}
		if doc.Status == spec.ReqStatusNotApplicable {
			severity = spec.SeverityInformational
		}
		if doc.Status == spec.ReqStatusSatisfied {
			completeCount++
		}
		r := spec.Requirement{
			ID:            doc.DocType,
			Category:      spec.CategoryDocumentRequirements,
			Type:          doc.DocType,
			Title:         doc.Title,
			Description:   doc.Reason,
			Status:        status,
			Severity:      severity,
			IsConditional: doc.IsConditional,
			Value:         doc.ApplicableStage,
		}
		reqs = append(reqs, r)
	}

	// For the group total, only count current-stage required docs
	currentStageDocs := 0
	currentStageComplete := 0
	for _, doc := range docs {
		if doc.ApplicableStage == spec.DocStageRFQ && doc.IsRequired {
			currentStageDocs++
			if doc.Status == spec.ReqStatusSatisfied {
				currentStageComplete++
			}
		}
	}

	groupStatus := "COMPLETE"
	if currentStageComplete < currentStageDocs {
		groupStatus = "INCOMPLETE"
	}

	return spec.RequirementGroup{
		Category:      spec.CategoryDocumentRequirements,
		Title:         "Document Requirements",
		Icon:          "📄",
		CompleteCount: currentStageComplete,
		TotalCount:    currentStageDocs,
		Status:        groupStatus,
		Requirements:  reqs,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Category F — AI Operational Findings
// Derives from real fields: rfqs.agent_status (string), rfqs.lead_id,
// and the ai_confidence scores present in lead_interactions (passed through agent_status).
// Does NOT override any deterministic requirement.
// ─────────────────────────────────────────────────────────────────────────────

func evalAIFindings(rfq *spec.RFQ) []spec.AIFinding {
	findings := []spec.AIFinding{}

	agentStatus := strings.ToUpper(rfq.AgentStatus)

	// If AI has successfully processed this RFQ (agent_status = COMPLETED/ENRICHED/PROCESSED)
	if agentStatus == "COMPLETED" || agentStatus == "ENRICHED" || agentStatus == "PROCESSED" {
		sourceCtx := "AI Agent Processing"
		if rfq.LeadID != nil {
			sourceCtx = fmt.Sprintf("AI Agent Processing — Lead #%d conversation thread", *rfq.LeadID)
		}
		findings = append(findings, spec.AIFinding{
			ID:                  "ai-shipment-extraction",
			Title:               "Shipment Details Extracted",
			Description:         "AI successfully extracted and merged shipment parameters from the customer email thread into this RFQ.",
			Confidence:          "HIGH",
			Recommendation:      "Review extracted values for accuracy and confirm with customer if needed.",
			RequiresHumanReview: false,
			SourceContext:       sourceCtx,
		})
	}

	// If AI is still processing (agent_status = PROCESSING/QUEUED)
	if agentStatus == "PROCESSING" || agentStatus == "QUEUED" {
		findings = append(findings, spec.AIFinding{
			ID:                  "ai-processing",
			Title:               "AI Extraction In Progress",
			Description:         "AI agent is currently processing the customer conversation to extract remaining shipment details.",
			Confidence:          "MEDIUM",
			Recommendation:      "Wait for AI extraction to complete before finalising the RFQ.",
			RequiresHumanReview: false,
		})
	}

	// If agent_status indicates a failure or partial extraction
	if agentStatus == "FAILED" || agentStatus == "PARTIAL" {
		findings = append(findings, spec.AIFinding{
			ID:                  "ai-extraction-incomplete",
			Title:               "Partial AI Extraction",
			Description:         "AI agent was unable to extract all shipment details from the available customer communication.",
			Confidence:          "LOW",
			Recommendation:      "Manually review and complete missing shipment information.",
			RequiresHumanReview: true,
		})
	}

	// If lead exists and rfq has complete items, generate a contextual finding
	if rfq.LeadID != nil {
		cargoDesc := cargoDescription(rfq.Items)
		if cargoDesc != "" {
			// Check for potential DG risk to alert user even without DG confirmation
			desc := strings.ToLower(cargoDesc)
			mightBeDG := false
			ambiguousKeywords := []string{"battery", "paint", "resin", "aerosol", "chemical", "polymer", "solvent", "acid"}
			for _, kw := range ambiguousKeywords {
				if strings.Contains(desc, kw) {
					mightBeDG = true
					break
				}
			}
			if mightBeDG {
				findings = append(findings, spec.AIFinding{
					ID:                  "ai-dg-risk-flag",
					Title:               "Potential Regulated Material Terminology Detected",
					Description:         fmt.Sprintf("Cargo description (\"%s\") contains terminology that may be associated with regulated or restricted materials. This does NOT automatically classify the cargo as dangerous goods.", cargoDesc),
					Confidence:          "MEDIUM",
					Recommendation:      "Request Safety Data Sheet (SDS/MSDS) from customer and verify commodity HS code before carrier booking.",
					RequiresHumanReview: true,
					SourceContext:       fmt.Sprintf("From RFQ cargo item description — Lead #%d", *rfq.LeadID),
				})
			}
		}
	}

	return findings
}

func buildAIFindingsGroup(findings []spec.AIFinding) spec.RequirementGroup {
	reqs := make([]spec.Requirement, 0, len(findings))
	for _, f := range findings {
		status := spec.ReqStatusSatisfied
		severity := spec.SeverityInformational
		if f.RequiresHumanReview {
			status = spec.ReqStatusUnderReview
			severity = spec.SeverityConditional
		}
		reqs = append(reqs, spec.Requirement{
			ID:            f.ID,
			Category:      spec.CategoryAIFindings,
			Type:          f.ID,
			Title:         f.Title,
			Description:   f.Description,
			Status:        status,
			Severity:      severity,
			Value:         fmt.Sprintf("Confidence: %s", f.Confidence),
			IsConditional: false,
			SourceContext: f.SourceContext,
		})
	}

	reviewCount := 0
	for _, f := range findings {
		if f.RequiresHumanReview {
			reviewCount++
		}
	}

	groupStatus := "COMPLETE"
	if reviewCount > 0 {
		groupStatus = "ATTENTION"
	}

	return spec.RequirementGroup{
		Category:      spec.CategoryAIFindings,
		Title:         "AI Operational Findings",
		Icon:          "✦",
		CompleteCount: len(findings) - reviewCount,
		TotalCount:    len(findings),
		Status:        groupStatus,
		Requirements:  reqs,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Readiness Calculation
// ─────────────────────────────────────────────────────────────────────────────

func calculateReadiness(groups []spec.RequirementGroup, docs []spec.DocumentRequirement) spec.OperationalReadiness {
	var (
		blockingCount     int
		missingRequired   int
		conditionalAttn   int
		completeCount     int
		totalCount        int
	)

	for _, g := range groups {
		// AI findings don't affect readiness score
		// Document requirements are tracked separately and don't block READY_FOR_QUOTATION
		// (document upload system is Task 11 — they show in UI but don't gate quotation)
		if g.Category == spec.CategoryAIFindings || g.Category == spec.CategoryDocumentRequirements {
			continue
		}
		for _, r := range g.Requirements {
			if r.Status == spec.ReqStatusNotApplicable {
				continue // Not applicable items don't count in totals
			}
			totalCount++
			switch {
			case r.Status == spec.ReqStatusSatisfied:
				completeCount++
			case r.Severity == spec.SeverityBlocking && r.Status == spec.ReqStatusMissing:
				blockingCount++
			case r.Severity == spec.SeverityRequired && r.Status == spec.ReqStatusMissing:
				missingRequired++
			case r.Severity == spec.SeverityConditional && (r.Status == spec.ReqStatusMissing || r.Status == spec.ReqStatusUnderReview):
				conditionalAttn++
			case r.Status == spec.ReqStatusUnderReview:
				conditionalAttn++
			default:
				if r.Status == spec.ReqStatusSatisfied {
					completeCount++
				}
			}
		}
	}

	// Determine overall status
	overallStatus := spec.ReadinessReadyForQuotation
	if blockingCount > 0 {
		overallStatus = spec.ReadinessInformationRequired
	} else if missingRequired > 0 {
		overallStatus = spec.ReadinessRequirementsIncomplete
	} else if conditionalAttn > 0 {
		overallStatus = spec.ReadinessAttentionRequired
	}

	// Readiness score (0–100)
	score := 0
	if totalCount > 0 {
		score = (completeCount * 100) / totalCount
	}
	if score > 100 {
		score = 100
	}

	return spec.OperationalReadiness{
		OverallStatus:             overallStatus,
		BlockingCount:             blockingCount,
		MissingRequiredCount:      missingRequired,
		ConditionalAttentionCount: conditionalAttn,
		CompleteCount:             completeCount,
		TotalCount:                totalCount,
		ReadinessScore:            score,
	}
}

func deriveNextBestAction(r spec.OperationalReadiness) string {
	switch r.OverallStatus {
	case spec.ReadinessInformationRequired:
		return fmt.Sprintf("Complete the %d blocking requirement(s) before proceeding.", r.BlockingCount)
	case spec.ReadinessRequirementsIncomplete:
		return fmt.Sprintf("%d required item(s) need attention before quotation.", r.MissingRequiredCount)
	case spec.ReadinessAttentionRequired:
		return "All critical requirements satisfied. Review conditional items and proceed to generate quotation."
	case spec.ReadinessReadyForQuotation:
		return "All quotation-stage requirements are complete. Proceed to generate and send quotation."
	default:
		return "Review requirements and complete all outstanding items."
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Helper Functions
// ─────────────────────────────────────────────────────────────────────────────

// evalField creates a BLOCKING or REQUIRED requirement based on whether the value is non-empty.
func evalField(id, title, severity, value, description string) spec.Requirement {
	status := spec.ReqStatusSatisfied
	if value == "" {
		status = spec.ReqStatusMissing
	}
	return spec.Requirement{
		ID:          id,
		Category:    severityToCategory(severity),
		Type:        id,
		Title:       title,
		Description: description,
		Status:      status,
		Severity:    severity,
		Value:       value,
	}
}

func evalFieldSeverity(id, title, severity, value, description string) spec.Requirement {
	r := evalField(id, title, severity, value, description)
	// For OPTIONAL fields, missing is not critical — mark as PENDING instead of MISSING
	if r.Status == spec.ReqStatusMissing && severity == spec.SeverityOptional {
		r.Status = spec.ReqStatusPending
	}
	return r
}

func severityToCategory(severity string) string {
	switch severity {
	case spec.SeverityBlocking:
		return spec.CategoryShipmentInfo
	case spec.SeverityRequired:
		return spec.CategoryCustomerInfo
	default:
		return spec.CategoryCargoOperational
	}
}

func buildGroup(category, title, icon string, reqs []spec.Requirement) spec.RequirementGroup {
	complete := 0
	total := 0
	for _, r := range reqs {
		if r.Status == spec.ReqStatusNotApplicable {
			continue
		}
		total++
		if r.Status == spec.ReqStatusSatisfied {
			complete++
		}
	}
	status := "COMPLETE"
	if complete < total {
		if total-complete >= 1 {
			status = "INCOMPLETE"
		}
	}
	// All items are conditional/informational only
	allConditional := true
	for _, r := range reqs {
		if r.Severity == spec.SeverityBlocking || r.Severity == spec.SeverityRequired {
			allConditional = false
			break
		}
	}
	if allConditional && status == "INCOMPLETE" {
		status = "ATTENTION"
	}

	return spec.RequirementGroup{
		Category:      category,
		Title:         title,
		Icon:          icon,
		CompleteCount: complete,
		TotalCount:    total,
		Status:        status,
		Requirements:  reqs,
	}
}

func strPtrVal(s *string) string {
	if s == nil {
		return ""
	}
	return strings.TrimSpace(*s)
}

func cargoDescription(items []spec.RFQItem) string {
	for _, it := range items {
		if strings.TrimSpace(it.Description) != "" {
			return strings.TrimSpace(it.Description)
		}
	}
	return ""
}

func totalWeightKG(items []spec.RFQItem) float64 {
	total := 0.0
	for _, it := range items {
		if it.WeightKG != nil {
			total += *it.WeightKG
		}
	}
	return total
}

func totalVolumeCBM(items []spec.RFQItem) float64 {
	total := 0.0
	for _, it := range items {
		if it.VolumeCBM != nil {
			total += *it.VolumeCBM
		}
	}
	return total
}

// deriveTransportModeHint infers transport mode from rfqs.incoterms.
// The rfqs table has no transport_mode column — mode is contextual.
func deriveTransportModeHint(rfq *spec.RFQ) string {
	incoterms := strings.ToUpper(strPtrVal(rfq.Incoterms))
	switch incoterms {
	case "FOB", "CIF", "CFR", "FAS":
		return "Ocean Freight (implied by Incoterms)"
	case "CPT", "CIP", "FCA", "DAP", "DDP", "DAT":
		return "Multimodal / Ocean Freight"
	case "EXW":
		return "To be determined — EXW (seller's premises)"
	default:
		if incoterms != "" {
			return fmt.Sprintf("Ocean Freight (%s)", incoterms)
		}
		return "Ocean Freight"
	}
}
