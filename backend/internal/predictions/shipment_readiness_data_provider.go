package predictions

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ShipmentReadinessData encapsulates verified operational facts for shipment documentation and readiness
type ShipmentReadinessData struct {
	ShipmentID          int64
	OrgID               int64
	BookingNumber       string
	TrackingNumber      string
	CarrierSCAC         string
	OriginPort          string
	DestinationPort     string
	Status              string
	ComplianceStatus    string
	ETD                 *time.Time
	ETA                 *time.Time
	ActualDeparture     *time.Time
	ActualArrival       *time.Time
	PresentDocuments    []string
	MissingDocuments    []string
	HasPendingDocument  bool
	PendingDocumentName string
	DocumentSummaries   []map[string]interface{}
	OpenDiscrepancies   []map[string]interface{}
	OpenExceptions      []map[string]interface{}
	LatestMilestone     string
	NextMilestone       string
	MilestoneReference  string
	CutoffReference     string
	DaysToETA           int
	ContextMap          map[string]interface{}
}

// ShipmentReadinessDataProvider retrieves persistent MariaDB facts for shipment readiness
type ShipmentReadinessDataProvider struct {
	db *sql.DB
}

// NewShipmentReadinessDataProvider initializes the data provider
func NewShipmentReadinessDataProvider(db *sql.DB) *ShipmentReadinessDataProvider {
	return &ShipmentReadinessDataProvider{db: db}
}

// FetchShipmentReadinessData gathers authoritative shipment records with tenant isolation
func (p *ShipmentReadinessDataProvider) FetchShipmentReadinessData(ctx context.Context, orgID int64, shipmentID int64) (*ShipmentReadinessData, error) {
	if p.db == nil {
		return nil, fmt.Errorf("database connection not available")
	}

	queryShipment := `
		SELECT 
			id, org_id, COALESCE(booking_number, ''), COALESCE(mbl_number, ''),
			COALESCE(carrier_scac, ''), COALESCE(origin_port, ''), COALESCE(destination_port, ''),
			COALESCE(status, 'DRAFT'),
			etd, eta
		FROM shipments
		WHERE id = ? AND org_id = ?
	`

	data := &ShipmentReadinessData{
		ShipmentID:        shipmentID,
		OrgID:             orgID,
		PresentDocuments:  make([]string, 0),
		MissingDocuments:  make([]string, 0),
		DocumentSummaries: make([]map[string]interface{}, 0),
		OpenDiscrepancies: make([]map[string]interface{}, 0),
		OpenExceptions:    make([]map[string]interface{}, 0),
	}

	var etd, eta sql.NullTime
	err := p.db.QueryRowContext(ctx, queryShipment, shipmentID, orgID).Scan(
		&data.ShipmentID, &data.OrgID, &data.BookingNumber, &data.TrackingNumber,
		&data.CarrierSCAC, &data.OriginPort, &data.DestinationPort,
		&data.Status,
		&etd, &eta,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("shipment %d not found for organization %d", shipmentID, orgID)
	} else if err != nil {
		return nil, fmt.Errorf("failed querying shipment: %w", err)
	}

	statusUpper := strings.ToUpper(data.Status)
	if statusUpper == "CUSTOMS_HOLD" {
		data.ComplianceStatus = "HOLD"
	} else {
		data.ComplianceStatus = "COMPLIANT"
	}

	now := time.Now().UTC()
	if etd.Valid {
		data.ETD = &etd.Time
	}
	if eta.Valid {
		data.ETA = &eta.Time
		data.DaysToETA = int(eta.Time.Sub(now).Hours() / 24)
	}

	// 2. Query Shipment Documents
	queryDocs := `
		SELECT id, doc_type, COALESCE(file_name, ''), COALESCE(status, 'PENDING'), created_at
		FROM shipment_documents
		WHERE shipment_id = ? AND org_id = ?
		ORDER BY created_at ASC
	`
	docRows, err := p.db.QueryContext(ctx, queryDocs, shipmentID, orgID)
	if err == nil {
		defer docRows.Close()
		for docRows.Next() {
			var docID int64
			var docType, fileName, docStatus string
			var createdAt time.Time
			if err := docRows.Scan(&docID, &docType, &fileName, &docStatus, &createdAt); err == nil {
				data.PresentDocuments = append(data.PresentDocuments, docType)
				docStatusUpper := strings.ToUpper(docStatus)
				if docStatusUpper == "PENDING" || docStatusUpper == "PENDING_REVIEW" || docStatusUpper == "PROCESSING" {
					data.HasPendingDocument = true
					if data.PendingDocumentName == "" {
						data.PendingDocumentName = fileName
					}
				}
				data.DocumentSummaries = append(data.DocumentSummaries, map[string]interface{}{
					"id":         docID,
					"doc_type":   docType,
					"file_name":  fileName,
					"status":     docStatus,
					"created_at": createdAt.Format(time.RFC3339),
				})
			}
		}
	}

	// Calculate deterministic missing required docs
	// Standard mandatory docs for ocean/air freight: COMMERCIAL_INVOICE, PACKING_LIST, and BILL_OF_LADING (or MBL/HBL)
	hasBL := false
	hasInvoice := false
	hasPackingList := false
	for _, dt := range data.PresentDocuments {
		dtUpper := strings.ToUpper(dt)
		if strings.Contains(dtUpper, "BILL_OF_LADING") || strings.Contains(dtUpper, "MBL") || strings.Contains(dtUpper, "HBL") || strings.Contains(dtUpper, "WAYBILL") {
			hasBL = true
		}
		if strings.Contains(dtUpper, "INVOICE") {
			hasInvoice = true
		}
		if strings.Contains(dtUpper, "PACKING") {
			hasPackingList = true
		}
	}
	if !hasBL {
		data.MissingDocuments = append(data.MissingDocuments, "BILL_OF_LADING")
	}
	if !hasInvoice {
		data.MissingDocuments = append(data.MissingDocuments, "COMMERCIAL_INVOICE")
	}
	if !hasPackingList {
		data.MissingDocuments = append(data.MissingDocuments, "PACKING_LIST")
	}

	// 3. Query Shipment Document Discrepancies
	queryDisc := `
		SELECT id, field_name, COALESCE(expected_value, ''), COALESCE(actual_value, ''),
		       COALESCE(source_document, ''), COALESCE(target_document, ''), COALESCE(status, 'OPEN')
		FROM shipment_document_discrepancies
		WHERE shipment_id = ? AND org_id = ? AND status = 'OPEN'
		ORDER BY created_at DESC
	`
	discRows, err := p.db.QueryContext(ctx, queryDisc, shipmentID, orgID)
	if err == nil {
		defer discRows.Close()
		for discRows.Next() {
			var dID int64
			var fieldName, expVal, actVal, srcDoc, tgtDoc, dStatus string
			if err := discRows.Scan(&dID, &fieldName, &expVal, &actVal, &srcDoc, &tgtDoc, &dStatus); err == nil {
				data.OpenDiscrepancies = append(data.OpenDiscrepancies, map[string]interface{}{
					"id":              dID,
					"field_name":      fieldName,
					"expected_value":  expVal,
					"actual_value":    actVal,
					"source_document": srcDoc,
					"target_document": tgtDoc,
					"status":          dStatus,
				})
			}
		}
	}

	// 4. Query Open Shipment Exceptions
	queryEx := `
		SELECT id, exception_type, severity, COALESCE(title, ''), COALESCE(description, ''), COALESCE(status, 'OPEN')
		FROM shipment_exceptions
		WHERE shipment_id = ? AND org_id = ? AND status != 'RESOLVED'
		ORDER BY created_at DESC
	`
	exRows, err := p.db.QueryContext(ctx, queryEx, shipmentID, orgID)
	if err == nil {
		defer exRows.Close()
		for exRows.Next() {
			var eID int64
			var eType, eSev, eTitle, eDesc, eStatus string
			if err := exRows.Scan(&eID, &eType, &eSev, &eTitle, &eDesc, &eStatus); err == nil {
				data.OpenExceptions = append(data.OpenExceptions, map[string]interface{}{
					"id":             eID,
					"exception_type": eType,
					"severity":       eSev,
					"title":          eTitle,
					"description":    eDesc,
					"status":         eStatus,
				})
			}
		}
	}

	// 5. Query Milestones
	queryMilestones := `
		SELECT milestone_code, COALESCE(description, ''), planned_date, actual_date, COALESCE(status, 'PENDING')
		FROM shipment_milestones
		WHERE shipment_id = ?
		ORDER BY COALESCE(actual_date, planned_date, id) ASC
	`
	mRows, err := p.db.QueryContext(ctx, queryMilestones, shipmentID)
	if err == nil {
		defer mRows.Close()
		for mRows.Next() {
			var mCode, mDesc, mStatus string
			var pDate, aDate sql.NullTime
			if err := mRows.Scan(&mCode, &mDesc, &pDate, &aDate, &mStatus); err == nil {
				mStatusUpper := strings.ToUpper(mStatus)
				if mStatusUpper == "COMPLETED" || aDate.Valid {
					data.LatestMilestone = mCode
				} else if data.NextMilestone == "" {
					data.NextMilestone = mCode
				}
			}
		}
	}

	// Determine deterministic milestone reference and cutoff reference based on operational state
	statusUpper = strings.ToUpper(data.Status)
	if statusUpper == "CUSTOMS_HOLD" {
		data.MilestoneReference = "CUSTOMS_CLEARANCE"
		data.CutoffReference = "CUSTOMS_SUBMISSION_DEADLINE"
	} else if len(data.OpenDiscrepancies) > 0 {
		if data.NextMilestone != "" {
			data.MilestoneReference = data.NextMilestone
		} else {
			data.MilestoneReference = "ARRIVAL"
		}
		data.CutoffReference = "IMPORT_MANIFEST_DEADLINE"
	} else if statusUpper == "IN_TRANSIT" || statusUpper == "DEPARTED" {
		if data.NextMilestone != "" {
			data.MilestoneReference = data.NextMilestone
		} else {
			data.MilestoneReference = "DELIVERY"
		}
		data.CutoffReference = "TERMINAL_STORAGE_CUTOFF"
	} else if statusUpper == "DELIVERED" || statusUpper == "COMPLETED" {
		data.MilestoneReference = "FINAL_SETTLEMENT"
		data.CutoffReference = "AUDIT_COMPLIANCE_CLOSURE"
	} else {
		data.MilestoneReference = "DEPARTURE"
		data.CutoffReference = "CARRIER_DOCUMENTATION_CUTOFF"
	}

	// 6. Build ContextMap for Python Sidecar
	ctxMap := map[string]interface{}{
		"shipment_id":           data.ShipmentID,
		"org_id":                data.OrgID,
		"booking_number":        data.BookingNumber,
		"tracking_number":       data.TrackingNumber,
		"carrier_scac":          data.CarrierSCAC,
		"origin_port":           data.OriginPort,
		"destination_port":      data.DestinationPort,
		"status":                data.Status,
		"compliance_status":     data.ComplianceStatus,
		"present_documents":     data.PresentDocuments,
		"missing_documents":     data.MissingDocuments,
		"has_pending_document":  data.HasPendingDocument,
		"pending_document_name": data.PendingDocumentName,
		"documents":             data.DocumentSummaries,
		"open_discrepancies":    data.OpenDiscrepancies,
		"open_exceptions":       data.OpenExceptions,
		"latest_milestone":      data.LatestMilestone,
		"next_milestone":        data.NextMilestone,
		"milestone_reference":   data.MilestoneReference,
		"cutoff_reference":      data.CutoffReference,
		"days_to_eta":           data.DaysToETA,
	}
	if data.ETD != nil {
		ctxMap["etd"] = data.ETD.Format(time.RFC3339)
	}
	if data.ETA != nil {
		ctxMap["eta"] = data.ETA.Format(time.RFC3339)
	}
	if data.ActualDeparture != nil {
		ctxMap["actual_departure"] = data.ActualDeparture.Format(time.RFC3339)
	}
	if data.ActualArrival != nil {
		ctxMap["actual_arrival"] = data.ActualArrival.Format(time.RFC3339)
	}

	data.ContextMap = ctxMap
	return data, nil
}
