package bcontext

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"
)

// GetCrossModuleInsights evaluates cross-domain relationships and deterministic rules for a specific entity or organization scope.
func (s *defaultService) GetCrossModuleInsights(ctx context.Context, orgID int64, entityType string, entityID int64, correlationID string, userID int64) (*CrossModuleInsightsResult, error) {
	if orgID <= 0 {
		return nil, fmt.Errorf("invalid organization id: %d", orgID)
	}

	if correlationID == "" {
		correlationID = fmt.Sprintf("corr-insights-%d-%d", orgID, time.Now().UnixNano())
	}

	now := time.Now().UTC()
	var insights []CrossModuleInsight
	var primaryRef *EntityReference
	connectedModulesMap := make(map[string]bool)

	entityUpper := strings.ToUpper(strings.TrimSpace(entityType))

	// ── 1. Specific Entity Traversal ──
	if entityUpper != "" && entityUpper != "ORG" && entityID > 0 {
		switch entityUpper {
		case "CUSTOMER":
			custInsights, ref, err := s.evaluateCustomerCrossModuleInsights(ctx, orgID, entityID, now)
			if err != nil {
				return nil, err
			}
			insights = append(insights, custInsights...)
			primaryRef = ref

		case "SHIPMENT":
			shInsights, ref, err := s.evaluateShipmentCrossModuleInsights(ctx, orgID, entityID, now)
			if err != nil {
				return nil, err
			}
			insights = append(insights, shInsights...)
			primaryRef = ref

		case "INVOICE":
			invInsights, ref, err := s.evaluateInvoiceCrossModuleInsights(ctx, orgID, entityID, now)
			if err != nil {
				return nil, err
			}
			insights = append(insights, invInsights...)
			primaryRef = ref

		case "CONTRACT":
			cntInsights, ref, err := s.evaluateContractCrossModuleInsights(ctx, orgID, entityID, now)
			if err != nil {
				return nil, err
			}
			insights = append(insights, cntInsights...)
			primaryRef = ref

		case "RFQ", "QUOTATION":
			rfqInsights, ref, err := s.evaluateRFQCrossModuleInsights(ctx, orgID, entityID, now)
			if err != nil {
				return nil, err
			}
			insights = append(insights, rfqInsights...)
			primaryRef = ref

		default:
			return nil, fmt.Errorf("unsupported entity type for cross-module insights: %s", entityType)
		}
	} else {
		// ── 2. Organization-Wide Aggregated Scan ──
		orgInsights, err := s.evaluateOrgWideCrossModuleInsights(ctx, orgID, now)
		if err != nil {
			return nil, err
		}
		insights = append(insights, orgInsights...)
		primaryRef = &EntityReference{
			EntityType: "ORGANIZATION",
			EntityID:   orgID,
			Reference:  fmt.Sprintf("ORG-%d", orgID),
			Module:     "organization",
		}
	}

	// Tally severities and collect connected modules
	critCount, highCount, modCount, lowCount := 0, 0, 0, 0
	for _, ins := range insights {
		switch ins.Severity {
		case SeverityCritical:
			critCount++
		case SeverityHigh:
			highCount++
		case SeverityModerate:
			modCount++
		case SeverityLow, SeverityInfo:
			lowCount++
		}

		if ins.PrimaryEntity.Module != "" {
			connectedModulesMap[ins.PrimaryEntity.Module] = true
		}
		for _, rel := range ins.RelatedEntities {
			if rel.Module != "" {
				connectedModulesMap[rel.Module] = true
			}
		}
	}

	// Sort insights by priority score descending
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].PriorityScore > insights[j].PriorityScore
	})

	var connectedModules []string
	for mod := range connectedModulesMap {
		connectedModules = append(connectedModules, mod)
	}
	sort.Strings(connectedModules)

	// ── 3. Grounded AI Synthesis ──
	aiSynthesis := s.buildCrossModuleAISynthesis(primaryRef, insights, now)

	return &CrossModuleInsightsResult{
		PrimaryEntity:      primaryRef,
		OrganizationScope:  orgID,
		CorrelationID:      correlationID,
		CalculatedAt:       now,
		ReadOnly:           true,
		TotalInsightsCount: len(insights),
		CriticalCount:      critCount,
		HighCount:          highCount,
		ModerateCount:      modCount,
		LowCount:           lowCount,
		ConnectedModules:   connectedModules,
		Insights:           insights,
		AISynthesis:        aiSynthesis,
		Warnings:           []string{},
		Limitations: []string{
			"Insights are derived deterministically from existing authorized database records.",
			"Does not constitute legal, tax, credit, or commercial advice.",
			"Always conduct human verification prior to taking commercial enforcement actions.",
		},
	}, nil
}

// GetOrgCrossModuleSummary returns organization-level KPIs and prioritized cross-module items.
func (s *defaultService) GetOrgCrossModuleSummary(ctx context.Context, orgID int64, correlationID string, userID int64) (*OrgCrossModuleSummary, error) {
	result, err := s.GetCrossModuleInsights(ctx, orgID, "ORG", 0, correlationID, userID)
	if err != nil {
		return nil, err
	}

	insightsByType := make(map[string]int)
	affectedEntities := make(map[string]int)

	for _, ins := range result.Insights {
		insightsByType[string(ins.InsightType)]++
		affectedEntities[ins.PrimaryEntity.EntityType]++
	}

	topInsights := result.Insights
	if len(topInsights) > 6 {
		topInsights = topInsights[:6]
	}

	return &OrgCrossModuleSummary{
		OrganizationScope:      orgID,
		CorrelationID:          result.CorrelationID,
		CalculatedAt:           result.CalculatedAt,
		ReadOnly:               true,
		TotalInsights:          result.TotalInsightsCount,
		CriticalInsights:       result.CriticalCount,
		HighInsights:           result.HighCount,
		ModerateInsights:       result.ModerateCount,
		LowInsights:            result.LowCount,
		TopPrioritizedInsights: topInsights,
		InsightsByType:         insightsByType,
		AffectedEntitiesCount:  affectedEntities,
		ConnectedModules:       result.ConnectedModules,
		DataFreshness:          result.CalculatedAt,
		Limitations:            result.Limitations,
	}, nil
}

// ── CUSTOMER CROSS-MODULE EVALUATION ──
func (s *defaultService) evaluateCustomerCrossModuleInsights(ctx context.Context, orgID int64, customerID int64, now time.Time) ([]CrossModuleInsight, *EntityReference, error) {
	var cust struct {
		ID   int64  `db:"id"`
		Name string `db:"name"`
		Code sql.NullString `db:"code"`
	}
	err := s.db.GetContext(ctx, &cust, `SELECT id, name, code FROM customers WHERE id = ? AND org_id = ? LIMIT 1`, customerID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("customer not found or unauthorized: id=%d", customerID)
	}

	custRef := &EntityReference{
		EntityType: "CUSTOMER",
		EntityID:   cust.ID,
		Reference:  cust.Name,
		Module:     "customers",
	}

	var insights []CrossModuleInsight

	// Check 1: Overdue Invoices + Active Shipments (Customer Credit & Operational Exposure)
	var overdueStats struct {
		OverdueCount int     `db:"overdue_count"`
		OverdueSum   float64 `db:"overdue_sum"`
	}
	_ = s.db.GetContext(ctx, &overdueStats, `
		SELECT COUNT(*) as overdue_count, COALESCE(SUM(balance_due), 0) as overdue_sum
		FROM customer_invoices
		WHERE org_id = ? AND customer_id = ? AND status != 'PAID' AND due_date < ?
	`, orgID, customerID, now)

	var activeShipmentsCount int
	_ = s.db.GetContext(ctx, &activeShipmentsCount, `
		SELECT COUNT(*)
		FROM shipments s
		LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
		WHERE s.org_id = ? AND r.customer_id = ? AND s.closure_status = 'ACTIVE'
	`, orgID, customerID)

	if overdueStats.OverdueCount > 0 && activeShipmentsCount > 0 {
		sev := SeverityHigh
		score := 75
		if overdueStats.OverdueSum > 10000 || overdueStats.OverdueCount >= 3 {
			sev = SeverityCritical
			score = 90
		}

		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-cust-credit-ops-%d-%d", customerID, now.Unix()),
			InsightType:       TypeCustomerRisk,
			Category:          "Credit & Operations Exposure",
			Title:             fmt.Sprintf("Customer Has %d Overdue Invoices with Active Shipments", overdueStats.OverdueCount),
			Explanation:       fmt.Sprintf("%s has outstanding overdue receivables of $%.2f across %d invoices while maintaining %d active operational shipments in transit.", cust.Name, overdueStats.OverdueSum, overdueStats.OverdueCount, activeShipmentsCount),
			Severity:          sev,
			PriorityScore:     score,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity:     *custRef,
			RelatedEntities: []EntityReference{
				{EntityType: "INVOICE", EntityID: 0, Reference: fmt.Sprintf("%d Overdue Invoices", overdueStats.OverdueCount), Module: "finance"},
				{EntityType: "SHIPMENT", EntityID: 0, Reference: fmt.Sprintf("%d In-Transit Shipments", activeShipmentsCount), Module: "shipments"},
			},
			Evidence: []InsightEvidenceItem{
				{SourceModule: "finance", FieldName: "overdue_invoices_count", ObservedValue: overdueStats.OverdueCount, Description: "Total unpaid invoices past due date"},
				{SourceModule: "finance", FieldName: "total_overdue_amount", ObservedValue: overdueStats.OverdueSum, Description: "Aggregate balance due past maturity"},
				{SourceModule: "shipments", FieldName: "active_shipments_count", ObservedValue: activeShipmentsCount, Description: "Active shipments currently moving in freight network"},
			},
			RuleApplied:          "RULE_CUST_OVERDUE_ACTIVE_OPS",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Coordinate with finance lead to verify payment commitments before releasing release notes or delivery orders for active consignments.",
			ReadOnly:             true,
		})
	}

	// Check 2: Active Volume Without Active Master Agreement
	var activeContractsCount int
	_ = s.db.GetContext(ctx, &activeContractsCount, `
		SELECT COUNT(*)
		FROM contracts
		WHERE org_id = ? AND party_id = ? AND party_type = 'CUSTOMER' AND status = 'ACTIVE' AND (expiry_date IS NULL OR expiry_date >= ?)
	`, orgID, customerID, now)

	if activeShipmentsCount > 0 && activeContractsCount == 0 {
		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-cust-contract-gap-%d-%d", customerID, now.Unix()),
			InsightType:       TypeCommercialContract,
			Category:          "Contract Coverage Gap",
			Title:             "Customer Generating Operational Freight Without Master Service Contract",
			Explanation:       fmt.Sprintf("%s has %d active shipments running under spot rates or uncommitted agreements with no registered active customer contract.", cust.Name, activeShipmentsCount),
			Severity:          SeverityModerate,
			PriorityScore:     60,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity:     *custRef,
			Evidence: []InsightEvidenceItem{
				{SourceModule: "contracts", FieldName: "active_contracts_count", ObservedValue: 0, Description: "Active contract records for customer"},
				{SourceModule: "shipments", FieldName: "active_shipments_count", ObservedValue: activeShipmentsCount, Description: "Active commercial movements"},
			},
			RuleApplied:          "RULE_CUST_ACTIVE_OPS_NO_CONTRACT",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Engage commercial sales director to negotiate master rate schedule or framework agreement to secure margins.",
			ReadOnly:             true,
		})
	}

	return insights, custRef, nil
}

// ── SHIPMENT CROSS-MODULE EVALUATION ──
func (s *defaultService) evaluateShipmentCrossModuleInsights(ctx context.Context, orgID int64, shipmentID int64, now time.Time) ([]CrossModuleInsight, *EntityReference, error) {
	var sh struct {
		ID            int64          `db:"id"`
		CarrierSCAC   string         `db:"carrier_scac"`
		ClosureStatus string         `db:"closure_status"`
		Status        string         `db:"status"`
		ETA           sql.NullTime   `db:"eta"`
		CreatedAt     time.Time      `db:"created_at"`
		CustomerID    sql.NullInt64  `db:"customer_id"`
		CustomerName  sql.NullString `db:"customer_name"`
	}
	query := `
		SELECT s.id, s.carrier_scac, s.closure_status, s.status, s.eta, s.created_at,
		       r.customer_id, cust.name as customer_name
		FROM shipments s
		LEFT JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
		LEFT JOIN customers cust ON r.customer_id = cust.id AND cust.org_id = s.org_id
		WHERE s.id = ? AND s.org_id = ? LIMIT 1
	`
	err := s.db.GetContext(ctx, &sh, query, shipmentID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("shipment not found or unauthorized: id=%d", shipmentID)
	}

	shRef := &EntityReference{
		EntityType: "SHIPMENT",
		EntityID:   sh.ID,
		Reference:  fmt.Sprintf("SH-%d", sh.ID),
		Module:     "shipments",
	}

	var insights []CrossModuleInsight

	// Check 1: Delayed Shipment with Linked Overdue / Impending Invoice
	var invList []struct {
		ID            int64     `db:"id"`
		InvoiceNumber string    `db:"invoice_number"`
		DueDate       time.Time `db:"due_date"`
		BalanceDue    float64   `db:"balance_due"`
		Status        string    `db:"status"`
	}
	_ = s.db.SelectContext(ctx, &invList, `
		SELECT id, invoice_number, due_date, balance_due, status
		FROM customer_invoices
		WHERE org_id = ? AND shipment_id = ? AND status != 'PAID'
	`, orgID, shipmentID)

	isDelayed := sh.ETA.Valid && sh.ETA.Time.Before(now) && sh.Status != "DELIVERED" && sh.Status != "COMPLETED"

	for _, inv := range invList {
		isOverdue := inv.DueDate.Before(now) && inv.BalanceDue > 0
		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)

		if isDelayed && isOverdue {
			insights = append(insights, CrossModuleInsight{
				InsightID:         fmt.Sprintf("ins-sh-delayed-inv-overdue-%d-%d", shipmentID, inv.ID),
				InsightType:       TypeOperationalFinance,
				Category:          "Delayed Freight with Overdue Receivable",
				Title:             fmt.Sprintf("Delayed Shipment SH-%d Linked to Overdue Invoice %s", shipmentID, inv.InvoiceNumber),
				Explanation:       fmt.Sprintf("Shipment SH-%d has exceeded ETA (%s) but remains uncompleted, while linked invoice %s is %d days overdue for $%.2f.", shipmentID, sh.ETA.Time.Format("2006-01-02"), inv.InvoiceNumber, daysOverdue, inv.BalanceDue),
				Severity:          SeverityHigh,
				PriorityScore:     85,
				Confidence:        "HIGH",
				DataFreshness:     now,
				OrganizationScope: orgID,
				PrimaryEntity:     *shRef,
				RelatedEntities: []EntityReference{
					{EntityType: "INVOICE", EntityID: inv.ID, Reference: inv.InvoiceNumber, Module: "finance"},
				},
				Evidence: []InsightEvidenceItem{
					{SourceModule: "shipments", FieldName: "eta", ObservedValue: sh.ETA.Time.Format(time.RFC3339), Description: "Scheduled vessel ETA"},
					{SourceModule: "shipments", FieldName: "status", ObservedValue: sh.Status, Description: "Current shipment operational milestone"},
					{SourceModule: "finance", FieldName: "invoice_number", ObservedValue: inv.InvoiceNumber, Description: "Linked customer billing record"},
					{SourceModule: "finance", FieldName: "balance_due", ObservedValue: inv.BalanceDue, Description: "Outstanding customer invoice balance"},
				},
				RuleApplied:          "RULE_SHIPMENT_DELAYED_INVOICE_OVERDUE",
				IsDeterministic:      true,
				LifecycleStatus:      "ACTIVE",
				SuggestedHumanAction: "Provide shipment tracking ETA updates to customer while requesting electronic remittance confirmation for overdue billing.",
				ReadOnly:             true,
			})
		}
	}

	// Check 2: Completed Shipment with Zero Invoices (Unbilled Revenue Leakage)
	isDelivered := sh.Status == "DELIVERED" || sh.Status == "COMPLETED"
	daysSinceCreation := int(now.Sub(sh.CreatedAt).Hours() / 24)

	var totalInvoicesForShipment int
	_ = s.db.GetContext(ctx, &totalInvoicesForShipment, `
		SELECT COUNT(*) FROM customer_invoices WHERE org_id = ? AND shipment_id = ?
	`, orgID, shipmentID)

	if isDelivered && totalInvoicesForShipment == 0 && daysSinceCreation >= 2 {
		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-sh-unbilled-%d-%d", shipmentID, now.Unix()),
			InsightType:       TypeOperationalFinance,
			Category:          "Unbilled Revenue Risk",
			Title:             fmt.Sprintf("Completed Shipment SH-%d Has No Customer Invoice", shipmentID),
			Explanation:       fmt.Sprintf("Shipment SH-%d has reached milestone %s but has no linked customer invoice generated in the finance module.", shipmentID, sh.Status),
			Severity:          SeverityHigh,
			PriorityScore:     80,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity:     *shRef,
			Evidence: []InsightEvidenceItem{
				{SourceModule: "shipments", FieldName: "status", ObservedValue: sh.Status, Description: "Execution status delivered"},
				{SourceModule: "finance", FieldName: "invoices_count", ObservedValue: 0, Description: "Zero customer billing records exist"},
			},
			RuleApplied:          "RULE_SHIPMENT_COMPLETED_NO_INVOICE",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Draft and dispatch customer invoice immediately based on confirmed ocean bill of lading and quotation line items.",
			ReadOnly:             true,
		})
	}

	return insights, shRef, nil
}

// ── INVOICE CROSS-MODULE EVALUATION ──
func (s *defaultService) evaluateInvoiceCrossModuleInsights(ctx context.Context, orgID int64, invoiceID int64, now time.Time) ([]CrossModuleInsight, *EntityReference, error) {
	var inv struct {
		ID            int64         `db:"id"`
		InvoiceNumber string        `db:"invoice_number"`
		CustomerID    int64         `db:"customer_id"`
		CustomerName  string        `db:"customer_name"`
		ShipmentID    sql.NullInt64 `db:"shipment_id"`
		TotalAmount   float64       `db:"total_amount"`
		BalanceDue    float64       `db:"balance_due"`
		DueDate       time.Time     `db:"due_date"`
		Status        string        `db:"status"`
	}
	query := `SELECT id, invoice_number, customer_id, customer_name, shipment_id, total_amount, balance_due, due_date, status FROM customer_invoices WHERE id = ? AND org_id = ? LIMIT 1`
	err := s.db.GetContext(ctx, &inv, query, invoiceID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("invoice not found or unauthorized: id=%d", invoiceID)
	}

	invRef := &EntityReference{
		EntityType: "INVOICE",
		EntityID:   inv.ID,
		Reference:  inv.InvoiceNumber,
		Module:     "finance",
	}

	var insights []CrossModuleInsight

	// Check 1: Invoice has no linked shipment record
	if !inv.ShipmentID.Valid || inv.ShipmentID.Int64 <= 0 {
		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-inv-no-shipment-%d-%d", invoiceID, now.Unix()),
			InsightType:       TypeWorkflowAttention,
			Category:          "Unlinked Commercial Transaction",
			Title:             fmt.Sprintf("Invoice %s Has No Linked Operational Shipment", inv.InvoiceNumber),
			Explanation:       fmt.Sprintf("Customer invoice %s ($%.2f) was issued without a direct link to an operational freight shipment.", inv.InvoiceNumber, inv.TotalAmount),
			Severity:          SeverityModerate,
			PriorityScore:     50,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity:     *invRef,
			Evidence: []InsightEvidenceItem{
				{SourceModule: "finance", FieldName: "shipment_id", ObservedValue: nil, Description: "Shipment ID is null"},
				{SourceModule: "finance", FieldName: "total_amount", ObservedValue: inv.TotalAmount, Description: "Billed invoice amount"},
			},
			RuleApplied:          "RULE_INVOICE_UNLINKED_SHIPMENT",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Verify if invoice corresponds to standalone customs brokerage, storage, or associate with corresponding shipment booking.",
			ReadOnly:             true,
		})
	}

	// Check 2: Overdue invoice with repeated late customer history
	if inv.Status != "PAID" && inv.DueDate.Before(now) {
		daysOverdue := int(now.Sub(inv.DueDate).Hours() / 24)
		var otherOverdueCount int
		_ = s.db.GetContext(ctx, &otherOverdueCount, `
			SELECT COUNT(*) FROM customer_invoices
			WHERE org_id = ? AND customer_id = ? AND id != ? AND status != 'PAID' AND due_date < ?
		`, orgID, inv.CustomerID, invoiceID, now)

		if otherOverdueCount >= 2 {
			insights = append(insights, CrossModuleInsight{
				InsightID:         fmt.Sprintf("ins-inv-repeated-late-%d-%d", invoiceID, now.Unix()),
				InsightType:       TypeCustomerRisk,
				Category:          "Chronic Late Payment Pattern",
				Title:             fmt.Sprintf("Invoice %s Overdue with %d Other Unsettled Invoices for %s", inv.InvoiceNumber, otherOverdueCount, inv.CustomerName),
				Explanation:       fmt.Sprintf("Invoice %s is %d days overdue ($%.2f due). Customer %s has %d additional delinquent invoices across the portfolio.", inv.InvoiceNumber, daysOverdue, inv.BalanceDue, inv.CustomerName, otherOverdueCount),
				Severity:          SeverityHigh,
				PriorityScore:     80,
				Confidence:        "HIGH",
				DataFreshness:     now,
				OrganizationScope: orgID,
				PrimaryEntity:     *invRef,
				RelatedEntities: []EntityReference{
					{EntityType: "CUSTOMER", EntityID: inv.CustomerID, Reference: inv.CustomerName, Module: "customers"},
				},
				Evidence: []InsightEvidenceItem{
					{SourceModule: "finance", FieldName: "days_overdue", ObservedValue: daysOverdue, Description: "Days since invoice due date"},
					{SourceModule: "finance", FieldName: "other_overdue_count", ObservedValue: otherOverdueCount, Description: "Co-occurring overdue invoices for party"},
				},
				RuleApplied:          "RULE_INVOICE_CHRONIC_DELINQUENCY",
				IsDeterministic:      true,
				LifecycleStatus:      "ACTIVE",
				SuggestedHumanAction: "Issue unified aging account statement and place credit hold on future quotation approvals until balance settled.",
				ReadOnly:             true,
			})
		}
	}

	return insights, invRef, nil
}

// ── CONTRACT CROSS-MODULE EVALUATION ──
func (s *defaultService) evaluateContractCrossModuleInsights(ctx context.Context, orgID int64, contractID int64, now time.Time) ([]CrossModuleInsight, *EntityReference, error) {
	var cnt struct {
		ID                int64          `db:"id"`
		ContractReference string         `db:"contract_reference"`
		PartyName         string         `db:"party_name"`
		PartyID           sql.NullInt64  `db:"party_id"`
		PartyType         string         `db:"party_type"`
		ExpiryDate        sql.NullTime   `db:"expiry_date"`
		Status            string         `db:"status"`
	}
	query := `SELECT id, contract_reference, party_name, party_id, party_type, expiry_date, status FROM contracts WHERE id = ? AND org_id = ? LIMIT 1`
	err := s.db.GetContext(ctx, &cnt, query, contractID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("contract not found or unauthorized: id=%d", contractID)
	}

	cntRef := &EntityReference{
		EntityType: "CONTRACT",
		EntityID:   cnt.ID,
		Reference:  cnt.ContractReference,
		Module:     "contracts",
	}

	var insights []CrossModuleInsight

	// Check 1: Contract expiring within 30 days while active shipments exist
	if cnt.ExpiryDate.Valid && cnt.ExpiryDate.Time.After(now) {
		daysUntilExpiry := int(cnt.ExpiryDate.Time.Sub(now).Hours() / 24)
		if daysUntilExpiry <= 30 && cnt.PartyID.Valid && cnt.PartyID.Int64 > 0 {
			var activeShipmentsForParty int
			if cnt.PartyType == "CUSTOMER" {
				_ = s.db.GetContext(ctx, &activeShipmentsForParty, `
					SELECT COUNT(*) FROM shipments s
					JOIN rfqs r ON s.rfq_id = r.id AND r.org_id = s.org_id
					WHERE s.org_id = ? AND r.customer_id = ? AND s.closure_status = 'ACTIVE'
				`, orgID, cnt.PartyID.Int64)
			}

			if activeShipmentsForParty > 0 {
				insights = append(insights, CrossModuleInsight{
					InsightID:         fmt.Sprintf("ins-cnt-expiring-active-ops-%d-%d", contractID, now.Unix()),
					InsightType:       TypeCommercialContract,
					Category:          "Expiring Agreement with Active Freight",
					Title:             fmt.Sprintf("Contract %s Expires in %d Days with %d Active Shipments", cnt.ContractReference, daysUntilExpiry, activeShipmentsForParty),
					Explanation:       fmt.Sprintf("Commercial contract %s with %s reaches expiration in %d days (%s) while %d consignments remain actively moving.", cnt.ContractReference, cnt.PartyName, daysUntilExpiry, cnt.ExpiryDate.Time.Format("2006-01-02"), activeShipmentsForParty),
					Severity:          SeverityHigh,
					PriorityScore:     75,
					Confidence:        "HIGH",
					DataFreshness:     now,
					OrganizationScope: orgID,
					PrimaryEntity:     *cntRef,
					Evidence: []InsightEvidenceItem{
						{SourceModule: "contracts", FieldName: "days_until_expiry", ObservedValue: daysUntilExpiry, Description: "Days remaining until agreement termination"},
						{SourceModule: "shipments", FieldName: "active_shipments_count", ObservedValue: activeShipmentsForParty, Description: "Active in-flight cargo movements"},
					},
					RuleApplied:          "RULE_CONTRACT_EXPIRING_ACTIVE_OPS",
					IsDeterministic:      true,
					LifecycleStatus:      "ACTIVE",
					SuggestedHumanAction: "Initiate renewal addendum or extension letter to prevent rate lapse on arriving cargo.",
					ReadOnly:             true,
				})
			}
		}
	}

	return insights, cntRef, nil
}

// ── RFQ / QUOTATION CROSS-MODULE EVALUATION ──
func (s *defaultService) evaluateRFQCrossModuleInsights(ctx context.Context, orgID int64, entityID int64, now time.Time) ([]CrossModuleInsight, *EntityReference, error) {
	var q struct {
		ID              int64         `db:"id"`
		QuotationNumber string        `db:"quotation_number"`
		CustomerID      sql.NullInt64 `db:"customer_id"`
		CustomerName    string        `db:"customer_name"`
		TotalAmount     float64       `db:"total_amount"`
		Status          string        `db:"status"`
	}
	err := s.db.GetContext(ctx, &q, `SELECT id, quotation_number, customer_id, customer_name, total_amount, status FROM quotations WHERE id = ? AND org_id = ? LIMIT 1`, entityID, orgID)
	if err != nil {
		return nil, nil, fmt.Errorf("quotation not found or unauthorized: id=%d", entityID)
	}

	qRef := &EntityReference{
		EntityType: "QUOTATION",
		EntityID:   q.ID,
		Reference:  q.QuotationNumber,
		Module:     "rfq",
	}

	var insights []CrossModuleInsight

	// Check 1: Quotation for customer with severe overdue balance
	if q.CustomerID.Valid && q.CustomerID.Int64 > 0 && q.Status == "DRAFT" {
		var overdueSum float64
		_ = s.db.GetContext(ctx, &overdueSum, `
			SELECT COALESCE(SUM(balance_due), 0) FROM customer_invoices
			WHERE org_id = ? AND customer_id = ? AND status != 'PAID' AND due_date < ?
		`, orgID, q.CustomerID.Int64, now)

		if overdueSum > 5000 {
			insights = append(insights, CrossModuleInsight{
				InsightID:         fmt.Sprintf("ins-quote-overdue-credit-%d-%d", entityID, now.Unix()),
				InsightType:       TypeCommercialContract,
				Category:          "Commercial Credit Warning",
				Title:             fmt.Sprintf("Quotation %s Pending for Customer with $%.2f Overdue", q.QuotationNumber, overdueSum),
				Explanation:       fmt.Sprintf("Quotation %s ($%.2f) is in draft stage, but customer %s has $%.2f in overdue unpaid receivables.", q.QuotationNumber, q.TotalAmount, q.CustomerName, overdueSum),
				Severity:          SeverityHigh,
				PriorityScore:     78,
				Confidence:        "HIGH",
				DataFreshness:     now,
				OrganizationScope: orgID,
				PrimaryEntity:     *qRef,
				Evidence: []InsightEvidenceItem{
					{SourceModule: "rfq", FieldName: "quotation_amount", ObservedValue: q.TotalAmount, Description: "Proposed quotation value"},
					{SourceModule: "finance", FieldName: "customer_overdue_sum", ObservedValue: overdueSum, Description: "Customer delinquent debt"},
				},
				RuleApplied:          "RULE_QUOTE_HIGH_OVERDUE_CUSTOMER",
				IsDeterministic:      true,
				LifecycleStatus:      "ACTIVE",
				SuggestedHumanAction: "Require upfront prepayment terms (PREPAID) on quotation before releasing draft to customer.",
				ReadOnly:             true,
			})
		}
	}

	return insights, qRef, nil
}

// ── ORG-WIDE CROSS-MODULE AGGREGATION ──
func (s *defaultService) evaluateOrgWideCrossModuleInsights(ctx context.Context, orgID int64, now time.Time) ([]CrossModuleInsight, error) {
	var insights []CrossModuleInsight

	// 1. Scan for delayed shipments with overdue invoices across the org
	type DelayedInvRow struct {
		ShipmentID    int64     `db:"shipment_id"`
		InvoiceID     int64     `db:"invoice_id"`
		InvoiceNumber string    `db:"invoice_number"`
		BalanceDue    float64   `db:"balance_due"`
		DueDate       time.Time `db:"due_date"`
		ETA           time.Time `db:"eta"`
		Status        string    `db:"status"`
	}
	var delayedInvs []DelayedInvRow
	delayQuery := `
		SELECT s.id as shipment_id, inv.id as invoice_id, inv.invoice_number,
		       inv.balance_due, inv.due_date, s.eta, s.status
		FROM shipments s
		JOIN customer_invoices inv ON s.id = inv.shipment_id AND inv.org_id = s.org_id
		WHERE s.org_id = ? AND s.status NOT IN ('DELIVERED', 'COMPLETED')
		  AND s.eta < ? AND inv.status != 'PAID' AND inv.due_date < ?
		ORDER BY inv.balance_due DESC LIMIT 10
	`
	_ = s.db.SelectContext(ctx, &delayedInvs, delayQuery, orgID, now, now)
	for _, row := range delayedInvs {
		daysOverdue := int(now.Sub(row.DueDate).Hours() / 24)
		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-org-sh-delay-inv-%d-%d", row.ShipmentID, row.InvoiceID),
			InsightType:       TypeOperationalFinance,
			Category:          "Delayed Freight & Overdue Invoice",
			Title:             fmt.Sprintf("Delayed SH-%d Linked to Overdue %s ($%.2f)", row.ShipmentID, row.InvoiceNumber, row.BalanceDue),
			Explanation:       fmt.Sprintf("Shipment SH-%d has missed ETA while invoice %s is %d days overdue.", row.ShipmentID, row.InvoiceNumber, daysOverdue),
			Severity:          SeverityHigh,
			PriorityScore:     85,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity: EntityReference{
				EntityType: "SHIPMENT",
				EntityID:   row.ShipmentID,
				Reference:  fmt.Sprintf("SH-%d", row.ShipmentID),
				Module:     "shipments",
			},
			RelatedEntities: []EntityReference{
				{EntityType: "INVOICE", EntityID: row.InvoiceID, Reference: row.InvoiceNumber, Module: "finance"},
			},
			Evidence: []InsightEvidenceItem{
				{SourceModule: "shipments", FieldName: "eta", ObservedValue: row.ETA.Format(time.RFC3339), Description: "Scheduled vessel ETA"},
				{SourceModule: "finance", FieldName: "balance_due", ObservedValue: row.BalanceDue, Description: "Outstanding receivable"},
			},
			RuleApplied:          "RULE_SHIPMENT_DELAYED_INVOICE_OVERDUE",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Synchronize shipment delay notifications with finance collection outreach.",
			ReadOnly:             true,
		})
	}

	// 2. Scan for customers with multiple overdue invoices and active shipments
	type DelinquentCustRow struct {
		CustomerID           int64   `db:"customer_id"`
		CustomerName         string  `db:"customer_name"`
		OverdueCount         int     `db:"overdue_count"`
		TotalOverdue         float64 `db:"total_overdue"`
		ActiveShipmentsCount int     `db:"active_shipments_count"`
	}
	var delinquentCusts []DelinquentCustRow
	custRiskQuery := `
		SELECT c.id as customer_id, c.name as customer_name,
		       COUNT(DISTINCT inv.id) as overdue_count,
		       COALESCE(SUM(inv.balance_due), 0) as total_overdue,
		       COUNT(DISTINCT s.id) as active_shipments_count
		FROM customers c
		JOIN customer_invoices inv ON c.id = inv.customer_id AND inv.org_id = c.org_id
		JOIN rfqs r ON c.id = r.customer_id AND r.org_id = c.org_id
		JOIN shipments s ON r.id = s.rfq_id AND s.org_id = c.org_id
		WHERE c.org_id = ? AND inv.status != 'PAID' AND inv.due_date < ? AND s.closure_status = 'ACTIVE'
		GROUP BY c.id, c.name
		HAVING overdue_count >= 1 AND active_shipments_count >= 1
		ORDER BY total_overdue DESC LIMIT 10
	`
	_ = s.db.SelectContext(ctx, &delinquentCusts, custRiskQuery, orgID, now)
	for _, dc := range delinquentCusts {
		sev := SeverityHigh
		score := 80
		if dc.TotalOverdue > 15000 || dc.OverdueCount >= 3 {
			sev = SeverityCritical
			score = 92
		}

		insights = append(insights, CrossModuleInsight{
			InsightID:         fmt.Sprintf("ins-org-cust-risk-%d-%d", dc.CustomerID, now.Unix()),
			InsightType:       TypeCustomerRisk,
			Category:          "Counterparty Exposure",
			Title:             fmt.Sprintf("%s: $%.2f Overdue with %d Cargo Shipments Active", dc.CustomerName, dc.TotalOverdue, dc.ActiveShipmentsCount),
			Explanation:       fmt.Sprintf("%s carries $%.2f in mature delinquent balances across %d invoices while holding %d active cargo consignments in transit.", dc.CustomerName, dc.TotalOverdue, dc.OverdueCount, dc.ActiveShipmentsCount),
			Severity:          sev,
			PriorityScore:     score,
			Confidence:        "HIGH",
			DataFreshness:     now,
			OrganizationScope: orgID,
			PrimaryEntity: EntityReference{
				EntityType: "CUSTOMER",
				EntityID:   dc.CustomerID,
				Reference:  dc.CustomerName,
				Module:     "customers",
			},
			Evidence: []InsightEvidenceItem{
				{SourceModule: "finance", FieldName: "total_overdue", ObservedValue: dc.TotalOverdue, Description: "Total delinquent receivables"},
				{SourceModule: "shipments", FieldName: "active_shipments", ObservedValue: dc.ActiveShipmentsCount, Description: "Active freight movements"},
			},
			RuleApplied:          "RULE_CUST_OVERDUE_ACTIVE_OPS",
			IsDeterministic:      true,
			LifecycleStatus:      "ACTIVE",
			SuggestedHumanAction: "Place credit hold on new quotation bookings until overdue invoices are paid.",
			ReadOnly:             true,
		})
	}

	return insights, nil
}

// ── GROUNDED AI SYNTHESIS BUILDER ──
func (s *defaultService) buildCrossModuleAISynthesis(ref *EntityReference, insights []CrossModuleInsight, now time.Time) *CrossModuleAISynthesis {
	refName := "Organization Portfolio"
	if ref != nil && ref.Reference != "" {
		refName = fmt.Sprintf("%s (%s)", ref.Reference, ref.EntityType)
	}

	if len(insights) == 0 {
		return &CrossModuleAISynthesis{
			ExecutiveSummary:           fmt.Sprintf("Cross-module relationship audit for %s completed with zero high-severity cross-module risks identified.", refName),
			ConnectedSituationAnalysis: "Operational, financial, commercial, and contract records are currently synchronized without conflicting statuses or unmitigated exposures.",
			TradeoffsAndPriorities:     "Maintain routine monitoring of shipping milestones and payment maturities.",
			SuggestedInvestigationQuestions: []string{
				"Are all forward cargo bookings scheduled for the next 14 days linked to verified customer agreements?",
				"Are recurring customers approaching contract renewal within the next quarter?",
			},
			Confidence:  "HIGH",
			GeneratedAt: now,
			Limitations: []string{"Based strictly on current records in MariaDB."},
		}
	}

	var critCount, highCount int
	var topTitles []string
	for i, ins := range insights {
		if ins.Severity == SeverityCritical {
			critCount++
		} else if ins.Severity == SeverityHigh {
			highCount++
		}
		if i < 3 {
			topTitles = append(topTitles, ins.Title)
		}
	}

	execSummary := fmt.Sprintf(
		"Cross-module audit for %s detected %d actionable signals (%d critical, %d high priority) connecting operational milestones with financial receivables and commercial coverage.",
		refName, len(insights), critCount, highCount,
	)

	situationAnalysis := fmt.Sprintf(
		"Top connected situations require attention: %s. Signals reflect operational delays, commercial contract validity, and customer credit exposure.",
		strings.Join(topTitles, "; "),
	)

	tradeoffs := "Balancing customer relationship preservation against operational freight costs requires synchronizing collection communications with transit updates."

	questions := []string{
		"Has the customer confirmed receipt and remittance schedule for overdue invoices on arriving shipments?",
		"Are all active spot shipments protected by valid carrier liability and demurrage terms?",
		"Should upcoming draft quotations be converted to PREPAID status pending balance reconciliation?",
	}

	return &CrossModuleAISynthesis{
		ExecutiveSummary:                execSummary,
		ConnectedSituationAnalysis:      situationAnalysis,
		TradeoffsAndPriorities:          tradeoffs,
		SuggestedInvestigationQuestions: questions,
		Confidence:                      "HIGH",
		GeneratedAt:                     now,
		Limitations: []string{
			"Deterministic analysis based on cross-module database foreign keys.",
			"Does not constitute legal or credit guarantees.",
		},
	}
}
