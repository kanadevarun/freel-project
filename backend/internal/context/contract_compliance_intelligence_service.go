package bcontext

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
)

// GetContract360ComplianceIntelligence generates a deterministic, read-only 360 contract and compliance profile.
func (s *defaultService) GetContract360ComplianceIntelligence(ctx context.Context, orgID, contractID int64, correlationID string, userID int64) (*Contract360ComplianceIntelligence, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	if contractID <= 0 {
		return nil, fmt.Errorf("invalid contract id: %d", contractID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-cnt-%d-%d", contractID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	// 1. Query Contract Header Record
	cntQuery := `
		SELECT 
			id, org_id, contract_reference, contract_name, contract_type,
			party_id, party_name, COALESCE(transport_mode, 'Multimodal') as transport_mode,
			status, COALESCE(currency, 'USD') as currency, COALESCE(contract_value, 0.0) as contract_value,
			effective_date, expiry_date, COALESCE(owner, '') as owner,
			created_at, updated_at
		FROM contracts
		WHERE id = ? AND org_id = ?`

	var c struct {
		ID                int64        `db:"id"`
		OrgID             int64        `db:"org_id"`
		ContractReference string       `db:"contract_reference"`
		ContractName      string       `db:"contract_name"`
		ContractType      string       `db:"contract_type"`
		PartyID           int64        `db:"party_id"`
		PartyName         string       `db:"party_name"`
		TransportMode     string       `db:"transport_mode"`
		Status            string       `db:"status"`
		Currency          string       `db:"currency"`
		ContractValue     float64      `db:"contract_value"`
		EffectiveDate     sql.NullTime `db:"effective_date"`
		ExpiryDate        sql.NullTime `db:"expiry_date"`
		Owner             string       `db:"owner"`
		CreatedAt         time.Time    `db:"created_at"`
		UpdatedAt         time.Time    `db:"updated_at"`
	}

	if err := s.db.GetContext(ctx, &c, cntQuery, contractID, orgID); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("contract not found or unauthorized: id=%d", contractID)
		}
		return nil, fmt.Errorf("failed to query contract: %w", err)
	}

	// 2. Query Renewal Tracking (if present)
	var renewalDate *time.Time
	var renewalRow struct {
		TargetCompletionDate sql.NullTime `db:"target_completion_date"`
	}
	if err := s.db.GetContext(ctx, &renewalRow, `SELECT target_completion_date FROM contract_renewal_tracking WHERE contract_id = ? AND org_id = ? LIMIT 1`, contractID, orgID); err == nil && renewalRow.TargetCompletionDate.Valid {
		renewalDate = &renewalRow.TargetCompletionDate.Time
	}

	// 3. Query Document Count
	var docCount int
	_ = s.db.GetContext(ctx, &docCount, `SELECT COUNT(*) FROM contract_documents WHERE contract_id = ? AND org_id = ?`, contractID, orgID)

	// 4. Compute Date Metrics
	var effTime, expTime *time.Time
	var daysUntilExpiry, daysExpired int
	var isExpired, isExpiringSoon, isCriticalExpiry, hasMissingEffective, hasMissingExpiry, hasConflictingDates bool

	if c.EffectiveDate.Valid {
		t := c.EffectiveDate.Time
		effTime = &t
	} else {
		hasMissingEffective = true
	}

	if c.ExpiryDate.Valid {
		t := c.ExpiryDate.Time
		expTime = &t
		diffDays := int(math.Ceil(t.Sub(now).Hours() / 24.0))
		if diffDays < 0 || strings.ToUpper(c.Status) == "EXPIRED" {
			isExpired = true
			daysExpired = int(math.Abs(float64(diffDays)))
			daysUntilExpiry = 0
		} else {
			daysUntilExpiry = diffDays
			daysExpired = 0
			if diffDays <= 7 {
				isCriticalExpiry = true
				isExpiringSoon = true
			} else if diffDays <= 30 {
				isExpiringSoon = true
			}
		}
	} else {
		hasMissingExpiry = true
	}

	if effTime != nil && expTime != nil && expTime.Before(*effTime) {
		hasConflictingDates = true
	}

	hasMissingDocument := (docCount == 0)

	// 5. Completeness Score Calculation (0 - 100)
	completeness := 0
	if c.ContractReference != "" {
		completeness += 15
	}
	if !hasMissingEffective {
		completeness += 15
	}
	if !hasMissingExpiry {
		completeness += 15
	}
	if c.PartyName != "" && c.PartyID > 0 {
		completeness += 15
	}
	if c.Owner != "" {
		completeness += 10
	}
	if docCount > 0 {
		completeness += 15
	}

	// Determine party type based on contract type
	partyType := "CUSTOMER"
	upperType := strings.ToUpper(c.ContractType)
	if strings.Contains(upperType, "CARRIER") || strings.Contains(upperType, "VENDOR") || strings.Contains(upperType, "SERVICE") {
		partyType = "CARRIER"
	}

	identity := ContractIdentitySummary{
		ContractID:          c.ID,
		ContractReference:   c.ContractReference,
		ContractName:        c.ContractName,
		ContractType:        c.ContractType,
		PartyID:             c.PartyID,
		PartyName:           c.PartyName,
		PartyType:           partyType,
		TransportMode:       c.TransportMode,
		Status:              c.Status,
		Currency:            c.Currency,
		ContractValue:       c.ContractValue,
		EffectiveDate:       effTime,
		ExpiryDate:          expTime,
		RenewalDate:         renewalDate,
		DaysUntilExpiry:     daysUntilExpiry,
		DaysExpired:         daysExpired,
		Owner:               c.Owner,
		DocumentCount:       docCount,
		CompletenessScore:   completeness,
		IsActive:            strings.ToUpper(c.Status) == "ACTIVE" && !isExpired && !hasConflictingDates,
		IsExpired:           isExpired,
		IsExpiringSoon:      isExpiringSoon,
		IsCriticalExpiry:    isCriticalExpiry,
		HasMissingDocument:  hasMissingDocument,
		HasMissingEffective: hasMissingEffective,
		HasMissingExpiry:    hasMissingExpiry,
		HasConflictingDates: hasConflictingDates,
		CreatedAt:           c.CreatedAt,
		UpdatedAt:           c.UpdatedAt,
	}

	// 6. Query Commercial Terms
	termsQuery := `
		SELECT id, term_category, term_key, term_title, term_value, value_type, currency, is_critical
		FROM contract_terms
		WHERE contract_id = ? AND org_id = ?
		ORDER BY display_order ASC, id ASC`

	var rawTerms []struct {
		ID           int64          `db:"id"`
		TermCategory string         `db:"term_category"`
		TermKey      string         `db:"term_key"`
		TermTitle    string         `db:"term_title"`
		TermValue    string         `db:"term_value"`
		ValueType    string         `db:"value_type"`
		Currency     sql.NullString `db:"currency"`
		IsCritical   bool           `db:"is_critical"`
	}
	_ = s.db.SelectContext(ctx, &rawTerms, termsQuery, contractID, orgID)

	var termsList []ContractTermSummary
	var freeTimeTerms, paymentTerms, liabilityTerms, slaTerms, cancelTerms []string
	var commCount, opCount, payCount, liabCount, critCount int

	for _, t := range rawTerms {
		var cur *string
		if t.Currency.Valid {
			cStr := t.Currency.String
			cur = &cStr
		}
		termsList = append(termsList, ContractTermSummary{
			ID:           t.ID,
			TermCategory: t.TermCategory,
			TermKey:      t.TermKey,
			TermTitle:    t.TermTitle,
			TermValue:    t.TermValue,
			ValueType:    t.ValueType,
			Currency:     cur,
			IsCritical:   t.IsCritical,
		})

		catUpper := strings.ToUpper(t.TermCategory)
		titleUpper := strings.ToUpper(t.TermTitle)
		keyUpper := strings.ToUpper(t.TermKey)

		if t.IsCritical {
			critCount++
		}

		switch catUpper {
		case "COMMERCIAL":
			commCount++
		case "OPERATIONAL":
			opCount++
		case "PAYMENT":
			payCount++
		case "LIABILITY":
			liabCount++
		}

		termDisplay := fmt.Sprintf("%s: %s", t.TermTitle, t.TermValue)
		if strings.Contains(titleUpper, "FREE") || strings.Contains(titleUpper, "DEMURRAGE") || strings.Contains(titleUpper, "DETENTION") || strings.Contains(keyUpper, "FREE_TIME") {
			freeTimeTerms = append(freeTimeTerms, termDisplay)
		} else if strings.Contains(titleUpper, "PAYMENT") || strings.Contains(titleUpper, "CREDIT") || strings.Contains(keyUpper, "PAYMENT") {
			paymentTerms = append(paymentTerms, termDisplay)
		} else if strings.Contains(titleUpper, "SLA") || strings.Contains(titleUpper, "TRANSIT") || strings.Contains(keyUpper, "SERVICE_LEVEL") {
			slaTerms = append(slaTerms, termDisplay)
		} else if strings.Contains(titleUpper, "LIABILITY") || strings.Contains(titleUpper, "CLAIM") || strings.Contains(titleUpper, "DAMAGE") {
			liabilityTerms = append(liabilityTerms, termDisplay)
		} else if strings.Contains(titleUpper, "CANCEL") || strings.Contains(titleUpper, "TERMINAT") {
			cancelTerms = append(cancelTerms, termDisplay)
		}
	}

	if len(rawTerms) > 0 {
		completeness += 15
		if completeness > 100 {
			completeness = 100
		}
		identity.CompletenessScore = completeness
	}

	commercial := ContractCommercialTermsSummary{
		TotalTermsCount:        len(termsList),
		CommercialTermsCount:   commCount,
		OperationalTermsCount:  opCount,
		PaymentTermsCount:      payCount,
		LiabilityTermsCount:    liabCount,
		CriticalTermsCount:     critCount,
		FreeTimeDemurrageTerms: freeTimeTerms,
		PaymentCreditTerms:     paymentTerms,
		ServiceLevelTerms:      slaTerms,
		LiabilityTerms:         liabilityTerms,
		CancellationTerms:      cancelTerms,
		Terms:                  termsList,
	}

	// 7. Query Contract Obligations & SLAs
	obQuery := `
		SELECT id, obligation_reference, title, obligation_type, responsible_party, COALESCE(owner, '') as owner,
		       priority, status, due_date, is_recurring
		FROM contract_obligations
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id ASC`

	var rawObligations []struct {
		ID                  int64        `db:"id"`
		ObligationReference string       `db:"obligation_reference"`
		Title               string       `db:"title"`
		ObligationType      string       `db:"obligation_type"`
		ResponsibleParty    string       `db:"responsible_party"`
		Owner               string       `db:"owner"`
		Priority            string       `db:"priority"`
		Status              string       `db:"status"`
		DueDate             sql.NullTime `db:"due_date"`
		IsRecurring         bool         `db:"is_recurring"`
	}
	_ = s.db.SelectContext(ctx, &rawObligations, obQuery, contractID, orgID)

	var obList []ContractObligationSummary
	var activeOb, fulfilledOb, overdueOb, breachedOb int
	var hasUnownedOb bool

	for _, ob := range rawObligations {
		var due *time.Time
		isOver := false
		if ob.DueDate.Valid {
			d := ob.DueDate.Time
			due = &d
			if d.Before(now) && strings.ToUpper(ob.Status) != "FULFILLED" && strings.ToUpper(ob.Status) != "WAIVED" {
				isOver = true
				overdueOb++
			}
		}

		stUpper := strings.ToUpper(ob.Status)
		switch stUpper {
		case "ACTIVE":
			activeOb++
		case "FULFILLED":
			fulfilledOb++
		case "BREACHED":
			breachedOb++
		}

		if ob.Owner == "" && stUpper == "ACTIVE" {
			hasUnownedOb = true
		}

		obList = append(obList, ContractObligationSummary{
			ID:                  ob.ID,
			ObligationReference: ob.ObligationReference,
			Title:               ob.Title,
			ObligationType:      ob.ObligationType,
			ResponsibleParty:    ob.ResponsibleParty,
			Owner:               ob.Owner,
			Priority:            ob.Priority,
			Status:              ob.Status,
			DueDate:             due,
			IsOverdue:           isOver,
			IsRecurring:         ob.IsRecurring,
		})
	}

	obligations := ContractObligationsSummary{
		TotalObligations:     len(obList),
		ActiveObligations:    activeOb,
		FulfilledObligations: fulfilledOb,
		OverdueObligations:   overdueOb,
		BreachedObligations:  breachedOb,
		HasUnownedObligation: hasUnownedOb,
		Obligations:          obList,
	}

	// 8. Query Compliance Requirements & Events
	reqQuery := `
		SELECT id, requirement_type, title, responsible_party, status, risk_severity,
		       valid_from, valid_until, evidence_document_id
		FROM contract_compliance_requirements
		WHERE contract_id = ? AND org_id = ?
		ORDER BY id ASC`

	var rawReqs []struct {
		ID                 int64          `db:"id"`
		RequirementType    string         `db:"requirement_type"`
		Title              string         `db:"title"`
		ResponsibleParty   string         `db:"responsible_party"`
		Status             string         `db:"status"`
		RiskSeverity       string         `db:"risk_severity"`
		ValidFrom          sql.NullTime   `db:"valid_from"`
		ValidUntil         sql.NullTime   `db:"valid_until"`
		EvidenceDocumentID sql.NullString `db:"evidence_document_id"`
	}
	_ = s.db.SelectContext(ctx, &rawReqs, reqQuery, contractID, orgID)

	var reqList []ComplianceRequirementSummary
	var verReq, pendReq, expReq, expSoonReq, missingEvid int

	for _, r := range rawReqs {
		var vf, vu *time.Time
		var isExp, isExpSoon bool
		if r.ValidFrom.Valid {
			t := r.ValidFrom.Time
			vf = &t
		}
		if r.ValidUntil.Valid {
			t := r.ValidUntil.Time
			vu = &t
			diff := int(math.Ceil(t.Sub(now).Hours() / 24.0))
			if diff < 0 {
				isExp = true
				expReq++
			} else if diff <= 30 {
				isExpSoon = true
				expSoonReq++
			}
		}

		stUpper := strings.ToUpper(r.Status)
		if stUpper == "VERIFIED" {
			verReq++
		} else if stUpper == "PENDING" {
			pendReq++
		}

		var evid *string
		if r.EvidenceDocumentID.Valid && r.EvidenceDocumentID.String != "" {
			e := r.EvidenceDocumentID.String
			evid = &e
		} else {
			missingEvid++
		}

		reqList = append(reqList, ComplianceRequirementSummary{
			ID:               r.ID,
			RequirementType:  r.RequirementType,
			Title:            r.Title,
			ResponsibleParty: r.ResponsibleParty,
			Status:           r.Status,
			RiskSeverity:     r.RiskSeverity,
			ValidFrom:        vf,
			ValidUntil:       vu,
			IsExpired:        isExp,
			IsExpiringSoon:   isExpSoon,
			EvidenceDocID:    evid,
		})
	}

	// Compliance Events
	evQuery := `
		SELECT id, event_type, severity, status, title, COALESCE(description, '') as description, detected_at
		FROM contract_compliance_events
		WHERE contract_id = ? AND org_id = ?
		ORDER BY detected_at DESC LIMIT 15`

	var rawEvents []struct {
		ID          int64     `db:"id"`
		EventType   string    `db:"event_type"`
		Severity    string    `db:"severity"`
		Status      string    `db:"status"`
		Title       string    `db:"title"`
		Description string    `db:"description"`
		DetectedAt  time.Time `db:"detected_at"`
	}
	_ = s.db.SelectContext(ctx, &rawEvents, evQuery, contractID, orgID)

	var evList []ComplianceEventSummary
	var openEvents, highEvents int
	for _, e := range rawEvents {
		isRes := strings.ToUpper(e.Status) == "RESOLVED"
		if !isRes {
			openEvents++
			if strings.ToUpper(e.Severity) == "HIGH" || strings.ToUpper(e.Severity) == "CRITICAL" {
				highEvents++
			}
		}
		evList = append(evList, ComplianceEventSummary{
			ID:          e.ID,
			EventType:   e.EventType,
			Severity:    e.Severity,
			Status:      e.Status,
			Title:       e.Title,
			Description: e.Description,
			DetectedAt:  e.DetectedAt,
			IsResolved:  isRes,
		})
	}

	complianceStatus := "COMPLIANT"
	if len(reqList) == 0 && len(evList) == 0 {
		complianceStatus = "UNKNOWN"
	} else if expReq > 0 || highEvents > 0 {
		complianceStatus = "ACTION_REQUIRED"
	} else if pendReq > 0 || expSoonReq > 0 || missingEvid > 0 {
		complianceStatus = "PENDING_VERIFICATION"
	}

	compliance := ContractComplianceSummary{
		TotalRequirements:        len(reqList),
		VerifiedRequirements:     verReq,
		PendingRequirements:      pendReq,
		ExpiredRequirements:      expReq,
		ExpiringSoonRequirements: expSoonReq,
		MissingEvidenceCount:     missingEvid,
		OpenEventsCount:          openEvents,
		HighSeverityEventsCount:  highEvents,
		ComplianceStatus:         complianceStatus,
		Requirements:             reqList,
		Events:                   evList,
	}

	// 9. Query Connected Links (Quotations, Shipments, Invoices)
	var linkedQuotes, linkedShipments, linkedInvoices int
	_ = s.db.GetContext(ctx, &linkedQuotes, `SELECT COUNT(*) FROM contract_links WHERE contract_id = ? AND org_id = ? AND linked_entity_type = 'QUOTATION'`, contractID, orgID)
	_ = s.db.GetContext(ctx, &linkedShipments, `SELECT COUNT(*) FROM contract_links WHERE contract_id = ? AND org_id = ? AND linked_entity_type = 'SHIPMENT'`, contractID, orgID)
	_ = s.db.GetContext(ctx, &linkedInvoices, `SELECT COUNT(*) FROM contract_links WHERE contract_id = ? AND org_id = ? AND linked_entity_type = 'INVOICE'`, contractID, orgID)

	coverage := ContractCoverageScope{
		CoveredParties:   []string{c.PartyName},
		CoveredModes:     []string{c.TransportMode},
		LinkedQuotations: linkedQuotes,
		LinkedShipments:  linkedShipments,
		LinkedInvoices:   linkedInvoices,
	}

	// 10. Risk Indicators & Score Calculation
	riskScore := 0
	var riskFactors []string

	if isExpired {
		riskScore += 40
		riskFactors = append(riskFactors, fmt.Sprintf("Agreement expired %d days ago", daysExpired))
	} else if isCriticalExpiry {
		riskScore += 30
		riskFactors = append(riskFactors, fmt.Sprintf("Critical expiration: %d days remaining", daysUntilExpiry))
	} else if isExpiringSoon {
		riskScore += 15
		riskFactors = append(riskFactors, fmt.Sprintf("Expiring soon: %d days remaining", daysUntilExpiry))
	}

	if hasMissingDocument {
		riskScore += 20
		riskFactors = append(riskFactors, "No executed agreement document attached")
	}

	if hasMissingExpiry {
		riskScore += 15
		riskFactors = append(riskFactors, "Agreement is missing a specified expiry date")
	}

	if hasMissingEffective {
		riskScore += 15
		riskFactors = append(riskFactors, "Agreement is missing an effective date")
	}

	if hasConflictingDates {
		riskScore += 25
		riskFactors = append(riskFactors, "Conflicting schedule: expiry date precedes effective date")
	}

	if highEvents > 0 {
		riskScore += 25
		riskFactors = append(riskFactors, fmt.Sprintf("%d unresolved high-severity compliance issues", highEvents))
	}

	if overdueOb > 0 {
		riskScore += 15
		riskFactors = append(riskFactors, fmt.Sprintf("%d contractual obligations/SLAs are currently overdue", overdueOb))
	}

	if expReq > 0 {
		riskScore += 20
		riskFactors = append(riskFactors, fmt.Sprintf("%d compliance requirement certificates/permits have expired", expReq))
	}

	if riskScore > 100 {
		riskScore = 100
	}

	riskRating := "LOW"
	if riskScore >= 70 {
		riskRating = "CRITICAL"
	} else if riskScore >= 45 {
		riskRating = "ELEVATED"
	} else if riskScore >= 20 {
		riskRating = "MODERATE"
	}

	riskIndicators := ContractRiskIndicators{
		OverallRiskRating:           riskRating,
		OverallRiskScore:            riskScore,
		IsExpired:                   isExpired,
		IsExpiringSoon:              isExpiringSoon,
		HasMissingDocument:          hasMissingDocument,
		HasMissingEffective:         hasMissingEffective,
		HasMissingExpiry:            hasMissingExpiry,
		HasConflictingDates:         hasConflictingDates,
		HasUnresolvedComplianceRisk: highEvents > 0,
		HasOverdueObligation:        overdueOb > 0,
		RiskFactors:                 riskFactors,
	}

	// 11. Deterministic Rule-Based AI Summary with Grounded Citations
	var citations []string
	citations = append(citations, fmt.Sprintf("[Contract: #%d (%s)]", c.ID, c.ContractReference))
	if c.PartyName != "" {
		citations = append(citations, fmt.Sprintf("[Party: #%d (%s)]", c.PartyID, c.PartyName))
	}
	if len(termsList) > 0 {
		citations = append(citations, fmt.Sprintf("[Terms: %d clauses]", len(termsList)))
	}
	if len(reqList) > 0 {
		citations = append(citations, fmt.Sprintf("[Compliance: %d requirements]", len(reqList)))
	}

	var reviewAreas []string
	if isExpired || isExpiringSoon {
		reviewAreas = append(reviewAreas, "Initiate commercial renewal review with account owner and legal counsel.")
	}
	if hasMissingDocument {
		reviewAreas = append(reviewAreas, "Upload signed master agreement or rate annexure PDF to establish authoritative documentation.")
	}
	if highEvents > 0 || expReq > 0 {
		reviewAreas = append(reviewAreas, "Verify updated insurance certificates or carrier compliance documentation.")
	}
	if overdueOb > 0 {
		reviewAreas = append(reviewAreas, "Audit overdue SLAs and contact responsible party for remediation.")
	}
	if len(reviewAreas) == 0 {
		reviewAreas = append(reviewAreas, "Standard quarterly rate and volume review recommended.")
	}

	execSummary := fmt.Sprintf(
		"Contract %s (%s) with %s is %s with completeness score of %d%%. Agreement is operating under %s rating (score: %d/100).",
		c.ContractReference, c.ContractName, c.PartyName, c.Status, completeness, riskRating, riskScore,
	)

	coverageAssessment := fmt.Sprintf(
		"Covers transport mode '%s' in currency %s with %d linked commercial records. Rate table has %d defined clauses.",
		c.TransportMode, c.Currency, linkedQuotes+linkedShipments+linkedInvoices, len(termsList),
	)

	complianceAssessment := fmt.Sprintf(
		"Compliance status is %s across %d requirements and %d events (%d high-severity open issues).",
		complianceStatus, len(reqList), len(evList), highEvents,
	)

	aiSummary := ContractAIComplianceSummary{
		ExecutiveSummary:     execSummary,
		CoverageAssessment:   coverageAssessment,
		ComplianceAssessment: complianceAssessment,
		SuggestedReviewAreas: reviewAreas,
		Confidence:           "HIGH",
		Citations:            citations,
		Limitations: []string{
			"Deterministic analysis based strictly on records in MariaDB.",
			"Does not provide definitive legal advice or regulatory guarantees.",
		},
	}

	return &Contract360ComplianceIntelligence{
		ContractID:     c.ID,
		Identity:       identity,
		Commercial:     commercial,
		Obligations:    obligations,
		Compliance:     compliance,
		Coverage:       coverage,
		RiskIndicators: riskIndicators,
		AISummary:      aiSummary,
		ReadOnly:       true,
		CorrelationID:  correlationID,
		DataFreshness:  now,
	}, nil
}

// GetContractCoverageForEntity evaluates whether an operational shipment, invoice, or quote has valid contract coverage.
func (s *defaultService) GetContractCoverageForEntity(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, userID int64) (*ContractCoverageEvaluation, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	if entityID <= 0 {
		return nil, fmt.Errorf("invalid entity id: %d", entityID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-cov-%s-%d", strings.ToLower(entityType), entityID)
	}

	now := time.Now().UTC()
	entityUpper := strings.ToUpper(entityType)

	var targetRef, customerOrCarrierName, targetMode, targetCurrency string
	var customerOrCarrierID int64
	var targetDate *time.Time

	switch entityUpper {
	case "SHIPMENT":
		var sh struct {
			ID           int64          `db:"id"`
			CarrierSCAC  string         `db:"carrier_scac"`
			CarrierName  sql.NullString `db:"carrier_name"`
			CustomerID   sql.NullInt64  `db:"customer_id"`
			CustomerName sql.NullString `db:"customer_name"`
			CreatedAt    time.Time      `db:"created_at"`
		}
		shQuery := `
			SELECT s.id, s.carrier_scac, c.name as carrier_name,
			       r.customer_id, cust.name as customer_name, s.created_at
			FROM shipments s
			LEFT JOIN carriers c ON s.carrier_scac = c.scac
			LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
			LEFT JOIN customers cust ON r.customer_id = cust.id AND cust.org_id = s.org_id
			WHERE s.id = ? AND s.org_id = ? LIMIT 1`
		if err := s.db.GetContext(ctx, &sh, shQuery, entityID, orgID); err != nil {
			return nil, fmt.Errorf("shipment not found or unauthorized: id=%d", entityID)
		}
		targetRef = fmt.Sprintf("SH-%d", sh.ID)
		if sh.CustomerID.Valid && sh.CustomerID.Int64 > 0 {
			customerOrCarrierID = sh.CustomerID.Int64
			customerOrCarrierName = sh.CustomerName.String
		} else if sh.CarrierName.Valid {
			customerOrCarrierName = sh.CarrierName.String
		} else {
			customerOrCarrierName = sh.CarrierSCAC
		}
		targetMode = "OCEAN"
		t := sh.CreatedAt
		targetDate = &t

	case "INVOICE":
		var inv struct {
			InvoiceNumber string    `db:"invoice_number"`
			CustomerID    int64     `db:"customer_id"`
			CustomerName  string    `db:"customer_name"`
			Currency      string    `db:"currency"`
			InvoiceDate   time.Time `db:"invoice_date"`
		}
		invQuery := `SELECT invoice_number, customer_id, customer_name, currency, invoice_date FROM customer_invoices WHERE id = ? AND org_id = ? LIMIT 1`
		if err := s.db.GetContext(ctx, &inv, invQuery, entityID, orgID); err != nil {
			return nil, fmt.Errorf("invoice not found or unauthorized: id=%d", entityID)
		}
		targetRef = inv.InvoiceNumber
		customerOrCarrierID = inv.CustomerID
		customerOrCarrierName = inv.CustomerName
		targetCurrency = inv.Currency
		t := inv.InvoiceDate
		targetDate = &t

	case "QUOTATION":
		var q struct {
			QuotationNumber string         `db:"quotation_number"`
			CustomerID      sql.NullInt64  `db:"customer_id"`
			CustomerName    sql.NullString `db:"customer_name"`
			Currency        string         `db:"currency"`
			TransportMode   sql.NullString `db:"transport_mode"`
			CreatedAt       time.Time      `db:"created_at"`
		}
		qQuery := `SELECT quotation_number, customer_id, customer_name, currency, transport_mode, created_at FROM quotations WHERE id = ? AND org_id = ? LIMIT 1`
		if err := s.db.GetContext(ctx, &q, qQuery, entityID, orgID); err != nil {
			return nil, fmt.Errorf("quotation not found or unauthorized: id=%d", entityID)
		}
		targetRef = q.QuotationNumber
		if q.CustomerID.Valid {
			customerOrCarrierID = q.CustomerID.Int64
		}
		if q.CustomerName.Valid {
			customerOrCarrierName = q.CustomerName.String
		}
		targetCurrency = q.Currency
		if q.TransportMode.Valid {
			targetMode = q.TransportMode.String
		}
		t := q.CreatedAt
		targetDate = &t

	default:
		return nil, fmt.Errorf("unsupported entity type for contract coverage: %s", entityType)
	}

	// Look for active contracts for this party (customer or carrier)
	var matchedContract struct {
		ID                int64        `db:"id"`
		ContractReference string       `db:"contract_reference"`
		ContractName      string       `db:"contract_name"`
		Currency          string       `db:"currency"`
		TransportMode     string       `db:"transport_mode"`
		Status            string       `db:"status"`
		EffectiveDate     sql.NullTime `db:"effective_date"`
		ExpiryDate        sql.NullTime `db:"expiry_date"`
	}

	matchQuery := `
		SELECT id, contract_reference, contract_name, COALESCE(currency, 'USD') as currency,
		       COALESCE(transport_mode, 'Multimodal') as transport_mode, status,
		       effective_date, expiry_date
		FROM contracts
		WHERE org_id = ? AND party_id = ? AND status = 'ACTIVE'
		ORDER BY id DESC LIMIT 1`

	hasMatch := false
	var gaps []string
	partyMatches := false
	modeMatches := false
	currencyMatches := false
	dateWithinValidity := false

	if err := s.db.GetContext(ctx, &matchedContract, matchQuery, orgID, customerOrCarrierID); err == nil {
		hasMatch = true
		partyMatches = true

		tUpper := strings.ToUpper(targetMode)
		cUpper := strings.ToUpper(matchedContract.TransportMode)
		if targetMode == "" || strings.EqualFold(matchedContract.TransportMode, "Multimodal") || strings.EqualFold(matchedContract.TransportMode, targetMode) || (tUpper == "OCEAN" && strings.Contains(cUpper, "OCEAN")) || (tUpper == "AIR" && strings.Contains(cUpper, "AIR")) || (tUpper == "ROAD" && strings.Contains(cUpper, "ROAD")) {
			modeMatches = true
		} else {
			gaps = append(gaps, fmt.Sprintf("Transport mode mismatch: entity is %s, contract covers %s", targetMode, matchedContract.TransportMode))
		}

		if targetCurrency == "" || strings.EqualFold(matchedContract.Currency, targetCurrency) {
			currencyMatches = true
		} else {
			gaps = append(gaps, fmt.Sprintf("Currency mismatch: entity is %s, contract is %s", targetCurrency, matchedContract.Currency))
		}

		if targetDate != nil {
			isAfterEff := !matchedContract.EffectiveDate.Valid || targetDate.After(matchedContract.EffectiveDate.Time) || targetDate.Equal(matchedContract.EffectiveDate.Time)
			isBeforeExp := !matchedContract.ExpiryDate.Valid || targetDate.Before(matchedContract.ExpiryDate.Time) || targetDate.Equal(matchedContract.ExpiryDate.Time)
			if isAfterEff && isBeforeExp {
				dateWithinValidity = true
			} else {
				gaps = append(gaps, "Transaction date falls outside the active contract validity window")
			}
		} else {
			dateWithinValidity = true
		}
	} else {
		gaps = append(gaps, fmt.Sprintf("No active master service contract found for party %s (ID: %d)", customerOrCarrierName, customerOrCarrierID))
	}

	coverageStatus := "OUTSIDE_COVERAGE"
	if hasMatch {
		if partyMatches && modeMatches && currencyMatches && dateWithinValidity {
			coverageStatus = "FULLY_COVERED"
		} else if partyMatches {
			coverageStatus = "PARTIALLY_COVERED"
		}
	}

	var matchedID *int64
	var matchedRef, matchedName string
	var citations []string
	citations = append(citations, fmt.Sprintf("[%s: #%d (%s)]", entityUpper, entityID, targetRef))

	if hasMatch {
		mID := matchedContract.ID
		matchedID = &mID
		matchedRef = matchedContract.ContractReference
		matchedName = matchedContract.ContractName
		citations = append(citations, fmt.Sprintf("[Contract: #%d (%s)]", mID, matchedRef))
	}

	aiSummary := ContractAIComplianceSummary{
		ExecutiveSummary: fmt.Sprintf(
			"Coverage assessment for %s %s: status is %s.",
			entityUpper, targetRef, coverageStatus,
		),
		CoverageAssessment: fmt.Sprintf(
			"Party Match: %t, Mode Match: %t, Currency Match: %t, Date Validity: %t.",
			partyMatches, modeMatches, currencyMatches, dateWithinValidity,
		),
		ComplianceAssessment: fmt.Sprintf("%d coverage gaps identified.", len(gaps)),
		SuggestedReviewAreas: gaps,
		Confidence:           "HIGH",
		Citations:            citations,
		Limitations: []string{
			"Deterministic evaluation based on real contracts matching party and organization ID in MariaDB.",
			"Does not constitute legal confirmation of contract applicability.",
		},
	}

	return &ContractCoverageEvaluation{
		TargetEntityType:      entityUpper,
		TargetEntityID:        entityID,
		TargetReference:       targetRef,
		CustomerOrCarrierID:   customerOrCarrierID,
		CustomerOrCarrierName: customerOrCarrierName,
		TargetMode:            targetMode,
		TargetCurrency:        targetCurrency,
		TargetDate:            targetDate,
		HasApplicableContract: hasMatch,
		MatchedContractID:     matchedID,
		MatchedContractRef:    matchedRef,
		MatchedContractName:   matchedName,
		CoverageStatus:        coverageStatus,
		PartyMatches:          partyMatches,
		ModeMatches:           modeMatches,
		CurrencyMatches:       currencyMatches,
		DateWithinValidity:    dateWithinValidity,
		CoverageGaps:          gaps,
		AISummary:             aiSummary,
		ReadOnly:              true,
		CorrelationID:         correlationID,
		DataFreshness:         now,
	}, nil
}

// GetOrgContractComplianceSummary aggregates organization-wide contract health and compliance posture.
func (s *defaultService) GetOrgContractComplianceSummary(ctx context.Context, orgID int64, correlationID string, userID int64) (*OrgContractComplianceSummary, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}
	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-org-cnt-%d-%d", orgID, time.Now().UnixNano())
	}

	now := time.Now().UTC()

	var totalContracts, activeContracts, draftContracts, expiredContracts int
	_ = s.db.GetContext(ctx, &totalContracts, `SELECT COUNT(*) FROM contracts WHERE org_id = ?`, orgID)
	_ = s.db.GetContext(ctx, &activeContracts, `SELECT COUNT(*) FROM contracts WHERE org_id = ? AND status = 'ACTIVE'`, orgID)
	_ = s.db.GetContext(ctx, &draftContracts, `SELECT COUNT(*) FROM contracts WHERE org_id = ? AND status = 'DRAFT'`, orgID)
	_ = s.db.GetContext(ctx, &expiredContracts, `SELECT COUNT(*) FROM contracts WHERE org_id = ? AND status = 'EXPIRED'`, orgID)

	var missingDocs, missingExpiry int
	_ = s.db.GetContext(ctx, &missingDocs, `SELECT COUNT(*) FROM contracts c WHERE c.org_id = ? AND NOT EXISTS (SELECT 1 FROM contract_documents cd WHERE cd.contract_id = c.id)`, orgID)
	_ = s.db.GetContext(ctx, &missingExpiry, `SELECT COUNT(*) FROM contracts WHERE org_id = ? AND expiry_date IS NULL`, orgID)

	// Expiring contracts within 30 days and 7 days
	var rawExpiring []struct {
		ID                int64        `db:"id"`
		ContractReference string       `db:"contract_reference"`
		ContractName      string       `db:"contract_name"`
		ContractType      string       `db:"contract_type"`
		PartyID           int64        `db:"party_id"`
		PartyName         string       `db:"party_name"`
		TransportMode     string       `db:"transport_mode"`
		Status            string       `db:"status"`
		Currency          string       `db:"currency"`
		ContractValue     float64      `db:"contract_value"`
		EffectiveDate     sql.NullTime `db:"effective_date"`
		ExpiryDate        sql.NullTime `db:"expiry_date"`
		Owner             string       `db:"owner"`
		CreatedAt         time.Time    `db:"created_at"`
		UpdatedAt         time.Time    `db:"updated_at"`
	}

	expiringQuery := `
		SELECT id, contract_reference, contract_name, contract_type, party_id, party_name,
		       COALESCE(transport_mode, 'Multimodal') as transport_mode, status,
		       COALESCE(currency, 'USD') as currency, COALESCE(contract_value, 0.0) as contract_value,
		       effective_date, expiry_date, COALESCE(owner, '') as owner, created_at, updated_at
		FROM contracts
		WHERE org_id = ? AND status = 'ACTIVE' AND expiry_date IS NOT NULL
		  AND expiry_date >= CURDATE() AND expiry_date <= DATE_ADD(CURDATE(), INTERVAL 30 DAY)
		ORDER BY expiry_date ASC`

	_ = s.db.SelectContext(ctx, &rawExpiring, expiringQuery, orgID)

	var expiringList []ContractIdentitySummary
	var exp30d, crit7d int
	for _, c := range rawExpiring {
		diff := int(math.Ceil(c.ExpiryDate.Time.Sub(now).Hours() / 24.0))
		exp30d++
		isCrit := false
		if diff <= 7 {
			crit7d++
			isCrit = true
		}
		var eff *time.Time
		if c.EffectiveDate.Valid {
			t := c.EffectiveDate.Time
			eff = &t
		}
		exp := c.ExpiryDate.Time

		expiringList = append(expiringList, ContractIdentitySummary{
			ContractID:        c.ID,
			ContractReference: c.ContractReference,
			ContractName:      c.ContractName,
			ContractType:      c.ContractType,
			PartyID:           c.PartyID,
			PartyName:         c.PartyName,
			TransportMode:     c.TransportMode,
			Status:            c.Status,
			Currency:          c.Currency,
			ContractValue:     c.ContractValue,
			EffectiveDate:     eff,
			ExpiryDate:        &exp,
			DaysUntilExpiry:   diff,
			Owner:             c.Owner,
			IsActive:          true,
			IsExpiringSoon:    true,
			IsCriticalExpiry:  isCrit,
		})
	}

	// Compliance Aggregates
	var totalReqs, verReqs, pendReqs, overdueReqs int
	_ = s.db.GetContext(ctx, &totalReqs, `SELECT COUNT(*) FROM contract_compliance_requirements WHERE org_id = ?`, orgID)
	_ = s.db.GetContext(ctx, &verReqs, `SELECT COUNT(*) FROM contract_compliance_requirements WHERE org_id = ? AND status = 'VERIFIED'`, orgID)
	_ = s.db.GetContext(ctx, &pendReqs, `SELECT COUNT(*) FROM contract_compliance_requirements WHERE org_id = ? AND status = 'PENDING'`, orgID)
	_ = s.db.GetContext(ctx, &overdueReqs, `SELECT COUNT(*) FROM contract_compliance_requirements WHERE org_id = ? AND valid_until IS NOT NULL AND valid_until < CURDATE()`, orgID)

	var openEvents, highEvents int
	_ = s.db.GetContext(ctx, &openEvents, `SELECT COUNT(*) FROM contract_compliance_events WHERE org_id = ? AND status != 'RESOLVED'`, orgID)
	_ = s.db.GetContext(ctx, &highEvents, `SELECT COUNT(*) FROM contract_compliance_events WHERE org_id = ? AND status != 'RESOLVED' AND severity IN ('HIGH', 'CRITICAL')`, orgID)

	// High risk compliance issues
	var rawHighIssues []struct {
		ID          int64     `db:"id"`
		EventType   string    `db:"event_type"`
		Severity    string    `db:"severity"`
		Status      string    `db:"status"`
		Title       string    `db:"title"`
		Description string    `db:"description"`
		DetectedAt  time.Time `db:"detected_at"`
	}
	highIssuesQuery := `
		SELECT id, event_type, severity, status, title, COALESCE(description, '') as description, detected_at
		FROM contract_compliance_events
		WHERE org_id = ? AND status != 'RESOLVED' AND severity IN ('HIGH', 'CRITICAL')
		ORDER BY detected_at DESC LIMIT 5`
	_ = s.db.SelectContext(ctx, &rawHighIssues, highIssuesQuery, orgID)

	var highIssuesList []ComplianceEventSummary
	for _, h := range rawHighIssues {
		highIssuesList = append(highIssuesList, ComplianceEventSummary{
			ID:          h.ID,
			EventType:   h.EventType,
			Severity:    h.Severity,
			Status:      h.Status,
			Title:       h.Title,
			Description: h.Description,
			DetectedAt:  h.DetectedAt,
			IsResolved:  false,
		})
	}

	health := "GOOD"
	if highEvents > 0 || overdueReqs > 0 || crit7d > 0 {
		health = "CRITICAL"
	} else if openEvents > 2 || exp30d > 2 || missingDocs > 2 {
		health = "HIGH_RISK"
	} else if totalContracts > 0 && activeContracts == 0 {
		health = "MODERATE"
	} else if totalContracts == 0 {
		health = "MODERATE"
	} else {
		health = "EXCELLENT"
	}

	return &OrgContractComplianceSummary{
		TotalContracts:              totalContracts,
		ActiveContracts:             activeContracts,
		DraftContracts:              draftContracts,
		ExpiringContracts30d:        exp30d,
		CriticalExpiring7d:          crit7d,
		ExpiredContracts:            expiredContracts,
		ContractsMissingDocuments:   missingDocs,
		ContractsMissingExpiryDate:  missingExpiry,
		TotalComplianceRequirements: totalReqs,
		VerifiedRequirements:        verReqs,
		PendingRequirements:         pendReqs,
		OverdueRequirements:         overdueReqs,
		OpenComplianceEventsCount:   openEvents,
		HighSeverityEventsCount:     highEvents,
		OverallComplianceHealth:     health,
		ExpiringContracts:           expiringList,
		HighRiskComplianceIssues:    highIssuesList,
		ReadOnly:                    true,
		CorrelationID:               correlationID,
		DataFreshness:               now,
	}, nil
}
