package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// RFQPricingData holds deterministic pricing and quote telemetry for an RFQ
type RFQPricingData struct {
	RFQID            int64
	OrgID            int64
	RFQNumber        string
	CustomerID       int64
	CustomerName     string
	Origin           string
	Destination      string
	Incoterms        string
	Stage            string
	Status           string
	QuotesCount      int
	SelectedQuoteID  int64
	CarrierName      string
	BuyPrice         float64
	SellPrice        float64
	MarginAmount     float64
	MarginPercentage float64
	HasMissingCosts  bool
	ContextMap       map[string]interface{}
}

// ContractRateData holds deterministic rate and expiry telemetry for a contract
type ContractRateData struct {
	ContractID        int64
	OrgID             int64
	ContractReference string
	ContractName      string
	ContractType      string
	PartyName         string
	Status            string
	Currency          string
	ContractValue     float64
	EffectiveDate     string
	ExpiryDate        string
	DaysUntilExpiry   int
	IsExpired         bool
	ContextMap        map[string]interface{}
}

// PricingMarginDataProvider fetches real persistent telemetry for RFQs, quotes, and contracts
type PricingMarginDataProvider struct {
	db *sql.DB
}

// NewPricingMarginDataProvider instantiates the provider
func NewPricingMarginDataProvider(db *sql.DB) *PricingMarginDataProvider {
	return &PricingMarginDataProvider{db: db}
}

// FetchRFQData queries rfqs and rfq_quotes with strict tenant isolation
func (p *PricingMarginDataProvider) FetchRFQData(ctx context.Context, orgID int64, rfqID int64) (*RFQPricingData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := `
		SELECT 
			id, org_id, rfq_number, COALESCE(customer_id, 0),
			COALESCE(origin, ''), COALESCE(destination, ''),
			COALESCE(incoterms, ''), COALESCE(stage, ''), COALESCE(status, '')
		FROM rfqs
		WHERE id = ? AND org_id = ?
	`
	var (
		rfqNumber, origin, destination, incoterms, stage, status string
		customerID                                               int64
	)

	err := p.db.QueryRowContext(ctx, query, rfqID, orgID).Scan(
		&rfqID, &orgID, &rfqNumber, &customerID,
		&origin, &destination, &incoterms, &stage, &status,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("rfq %d not found for organization %d", rfqID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying rfq: %w", err)
	}

	// Fetch customer name if linked
	var customerName string
	if customerID > 0 {
		_ = p.db.QueryRowContext(ctx, "SELECT COALESCE(name, '') FROM customers WHERE id = ? AND org_id = ?", customerID, orgID).Scan(&customerName)
	}

	// Fetch quotes associated with RFQ
	quoteQuery := `
		SELECT 
			id, COALESCE(carrier_name, 'Carrier'),
			COALESCE(buy_price, 0.0), COALESCE(sell_price, 0.0),
			COALESCE(ocean_freight, 0.0), COALESCE(origin_charges, 0.0), COALESCE(destination_charges, 0.0),
			COALESCE(is_recommended, 0)
		FROM rfq_quotes
		WHERE rfq_id = ?
		ORDER BY is_recommended DESC, id ASC
	`
	rows, err := p.db.QueryContext(ctx, quoteQuery, rfqID)
	if err != nil {
		return nil, fmt.Errorf("failed querying rfq quotes: %w", err)
	}
	defer rows.Close()

	quotesCount := 0
	var (
		selectedQuoteID                                           int64
		carrierName                                               string
		buyPrice, sellPrice, oceanFreight, originChg, destChg     float64
		isRecommended                                             int
		hasMissingCosts                                           bool
	)

	for rows.Next() {
		var qID int64
		var cName string
		var bPrice, sPrice, oFreight, oChg, dChg float64
		var isRec int

		if err := rows.Scan(&qID, &cName, &bPrice, &sPrice, &oFreight, &oChg, &dChg, &isRec); err == nil {
			quotesCount++
			if quotesCount == 1 || isRec == 1 {
				selectedQuoteID = qID
				carrierName = cName
				buyPrice = bPrice
				sellPrice = sPrice
				oceanFreight = oFreight
				originChg = oChg
				destChg = dChg
				isRecommended = isRec
			}
			if bPrice <= 0 || sPrice <= 0 {
				hasMissingCosts = true
			}
		}
	}

	marginAmount := sellPrice - buyPrice
	marginPercentage := 0.0
	if sellPrice > 0 {
		marginPercentage = (marginAmount / sellPrice) * 100.0
	}

	contextMap := map[string]interface{}{
		"rfq_id":             rfqID,
		"rfq_number":         rfqNumber,
		"customer_id":        customerID,
		"customer_name":      customerName,
		"origin":             origin,
		"destination":        destination,
		"incoterms":          incoterms,
		"stage":              stage,
		"status":             status,
		"quotes_count":       quotesCount,
		"selected_quote_id":  selectedQuoteID,
		"carrier_name":       carrierName,
		"buy_price":          buyPrice,
		"sell_price":         sellPrice,
		"margin_amount":      marginAmount,
		"margin_percentage":  marginPercentage,
		"ocean_freight":      oceanFreight,
		"origin_charges":     originChg,
		"dest_charges":       destChg,
		"has_missing_costs":  hasMissingCosts,
		"is_recommended":     isRecommended == 1,
	}

	return &RFQPricingData{
		RFQID:            rfqID,
		OrgID:            orgID,
		RFQNumber:        rfqNumber,
		CustomerID:       customerID,
		CustomerName:     customerName,
		Origin:           origin,
		Destination:      destination,
		Incoterms:        incoterms,
		Stage:            stage,
		Status:           status,
		QuotesCount:      quotesCount,
		SelectedQuoteID:  selectedQuoteID,
		CarrierName:      carrierName,
		BuyPrice:         buyPrice,
		SellPrice:        sellPrice,
		MarginAmount:     marginAmount,
		MarginPercentage: marginPercentage,
		HasMissingCosts:  hasMissingCosts,
		ContextMap:       contextMap,
	}, nil
}

// FetchContractData queries contracts with strict tenant isolation
func (p *PricingMarginDataProvider) FetchContractData(ctx context.Context, orgID int64, contractID int64) (*ContractRateData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	query := `
		SELECT 
			id, org_id, COALESCE(contract_reference, ''), COALESCE(contract_name, ''),
			COALESCE(contract_type, ''), COALESCE(party_name, ''), COALESCE(status, ''),
			COALESCE(currency, 'USD'), COALESCE(contract_value, 0.0),
			effective_date, expiry_date
		FROM contracts
		WHERE id = ? AND org_id = ?
	`
	var (
		ref, name, cType, party, status, currency string
		val                                       float64
		effDate, expDate                          sql.NullTime
	)

	err := p.db.QueryRowContext(ctx, query, contractID, orgID).Scan(
		&contractID, &orgID, &ref, &name, &cType, &party, &status,
		&currency, &val, &effDate, &expDate,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("contract %d not found for organization %d", contractID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying contract: %w", err)
	}

	now := time.Now().UTC()
	daysUntilExpiry := 999
	isExpired := false
	expDateStr := ""
	effDateStr := ""

	if effDate.Valid {
		effDateStr = effDate.Time.Format("2006-01-02")
	}
	if expDate.Valid {
		expDateStr = expDate.Time.Format("2006-01-02")
		daysUntilExpiry = int(expDate.Time.Sub(now).Hours() / 24)
		if daysUntilExpiry < 0 {
			isExpired = true
		}
	}

	contextMap := map[string]interface{}{
		"contract_id":        contractID,
		"contract_reference": ref,
		"contract_name":      name,
		"contract_type":      cType,
		"party_name":         party,
		"status":             status,
		"currency":           currency,
		"contract_value":     val,
		"effective_date":     effDateStr,
		"expiry_date":        expDateStr,
		"days_until_expiry":  daysUntilExpiry,
		"is_expired":         isExpired,
	}

	return &ContractRateData{
		ContractID:        contractID,
		OrgID:             orgID,
		ContractReference: ref,
		ContractName:      name,
		ContractType:      cType,
		PartyName:         party,
		Status:            status,
		Currency:          currency,
		ContractValue:     val,
		EffectiveDate:     effDateStr,
		ExpiryDate:        expDateStr,
		DaysUntilExpiry:   daysUntilExpiry,
		IsExpired:         isExpired,
		ContextMap:        contextMap,
	}, nil
}
