package shipments

import (
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/shipments/spec"
	"github.com/freel/backend/internal/svcerror"
)

// NormalizeDocType standardizes different variations, aliases, and lowercase strings into a canonical uppercase DocumentType.
// For example: "master_bl", "mbl", "BL" all become "MBL".
func NormalizeDocType(raw string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(raw))
	cleaned = strings.ReplaceAll(cleaned, "-", "_")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")

	switch cleaned {
	case spec.DocTypeMasterBillOfLading, "MASTER_BL", "MASTER_BILL_OF_LADING", spec.DocTypeBillOfLading, "BL":
		return spec.DocTypeMasterBillOfLading
	case spec.DocTypeHouseBillOfLading, "HOUSE_BL", "HOUSE_BILL_OF_LADING":
		return spec.DocTypeHouseBillOfLading
	case spec.DocTypeSeaWaybill, "WAYBILL":
		return spec.DocTypeSeaWaybill
	case spec.DocTypeAirWaybill, "AWB", "AIR_WAY_BILL":
		return spec.DocTypeAirWaybill
	case spec.DocTypeArrivalNotice, "AN":
		return spec.DocTypeArrivalNotice
	case spec.DocTypeDeliveryOrder, "DO":
		return spec.DocTypeDeliveryOrder
	case spec.DocTypeBookingConfirmation, "BOOKING_CONF":
		return spec.DocTypeBookingConfirmation
	case spec.DocTypeCarrierRelease:
		return spec.DocTypeCarrierRelease
	case spec.DocTypeCommercialInvoice, "CI", "INVOICE":
		return spec.DocTypeCommercialInvoice
	case spec.DocTypeProformaInvoice, "PI":
		return spec.DocTypeProformaInvoice
	case spec.DocTypePackingList, "PL":
		return spec.DocTypePackingList
	case spec.DocTypePurchaseOrder, "PO":
		return spec.DocTypePurchaseOrder
	case spec.DocTypeCertificateOfOrigin, "COO":
		return spec.DocTypeCertificateOfOrigin
	case spec.DocTypeCustomsDeclaration, "CUSTOMS", "CUSTOMS_ENTRY":
		return spec.DocTypeCustomsDeclaration
	case spec.DocTypeImportPermit:
		return spec.DocTypeImportPermit
	case spec.DocTypeExportPermit:
		return spec.DocTypeExportPermit
	case spec.DocTypeDutyDocumentation:
		return spec.DocTypeDutyDocumentation
	case spec.DocTypeRegulatoryCertificate:
		return spec.DocTypeRegulatoryCertificate
	case spec.DocTypeCustomsClearance, "CLEARANCE":
		return spec.DocTypeCustomsClearance
	case spec.DocTypeImportDocumentation:
		return spec.DocTypeImportDocumentation
	case spec.DocTypeExportDocumentation:
		return spec.DocTypeExportDocumentation
	case spec.DocTypeInspectionCert, "INSPECTION":
		return spec.DocTypeInspectionCert
	case spec.DocTypeInsuranceCert, "INSURANCE":
		return spec.DocTypeInsuranceCert
	case spec.DocTypeInsurancePolicy:
		return spec.DocTypeInsurancePolicy
	case spec.DocTypeDangerousGoods, "MSDS", "DANGEROUS_GOODS", "DG_DECLARATION":
		return spec.DocTypeDangerousGoods
	case spec.DocTypeCargoCertificate:
		return spec.DocTypeCargoCertificate
	case spec.DocTypeShippingInstructions, "SI":
		return spec.DocTypeShippingInstructions
	case spec.DocTypeCargoManifest, "MANIFEST":
		return spec.DocTypeCargoManifest
	case spec.DocTypeWeightCertificate, "VGM", "WEIGHT_CERT":
		return spec.DocTypeWeightCertificate
	case spec.DocTypeDeliveryReceipt, "POD", "PROOF_OF_DELIVERY":
		return spec.DocTypeDeliveryReceipt
	case spec.DocTypeCarrierDocument:
		return spec.DocTypeCarrierDocument
	case spec.DocTypeCustomerDocument:
		return spec.DocTypeCustomerDocument
	case spec.DocTypeOther:
		return spec.DocTypeOther
	default:
		return cleaned
	}
}

// InferCategory maps a document type to its primary operational category.
// Categories include: TRANSPORT, COMMERCIAL, CUSTOMS, INSURANCE, OPERATIONAL, CARGO, OTHER.
func InferCategory(docType string) string {
	norm := NormalizeDocType(docType)
	switch norm {
	case spec.DocTypeBillOfLading, spec.DocTypeMasterBillOfLading, spec.DocTypeHouseBillOfLading,
		spec.DocTypeSeaWaybill, spec.DocTypeAirWaybill, spec.DocTypeArrivalNotice,
		spec.DocTypeDeliveryOrder, spec.DocTypeBookingConfirmation, spec.DocTypeCarrierRelease:
		return spec.DocCategoryTransport

	case spec.DocTypeCommercialInvoice, spec.DocTypeProformaInvoice, spec.DocTypePackingList,
		spec.DocTypePurchaseOrder, spec.DocTypeCertificateOfOrigin:
		return spec.DocCategoryCommercial

	case spec.DocTypeCustomsDeclaration, spec.DocTypeCustomsClearance, spec.DocTypeImportDocumentation,
		spec.DocTypeExportDocumentation, spec.DocTypeImportPermit, spec.DocTypeExportPermit,
		spec.DocTypeDutyDocumentation, spec.DocTypeRegulatoryCertificate, spec.DocTypeInspectionCert:
		return spec.DocCategoryCustoms

	case spec.DocTypeInsuranceCert, spec.DocTypeInsurancePolicy:
		return spec.DocCategoryInsurance

	case spec.DocTypeShippingInstructions, spec.DocTypeCargoManifest, spec.DocTypeWeightCertificate,
		spec.DocTypeDeliveryReceipt:
		return spec.DocCategoryOperational

	case spec.DocTypeDangerousGoods, spec.DocTypeCargoCertificate:
		return spec.DocCategoryCargo

	default:
		return spec.DocCategoryOther
	}
}

// ComputeDocumentValidity determines if a document is currently valid, expiring soon (within 14 days), or expired.
func ComputeDocumentValidity(expiresAt *time.Time) string {
	return ComputeDocumentValidityAt(expiresAt, time.Now())
}

// ComputeDocumentValidityAt evaluates document validity relative to a reference timestamp.
func ComputeDocumentValidityAt(expiresAt *time.Time, now time.Time) string {
	if expiresAt == nil {
		return spec.DocValidityValid
	}
	if expiresAt.Before(now) {
		return spec.DocValidityExpired
	}
	// Expiring within 14 days (14 * 24 hours)
	if expiresAt.Before(now.Add(14 * 24 * time.Hour)) {
		return spec.DocValidityExpiringSoon
	}
	return spec.DocValidityValid
}

// ValidateDocumentTransition enforces permitted lifecycle steps for a document file.
// For example: A document cannot jump straight from MISSING to APPROVED without being uploaded first.
func ValidateDocumentTransition(currentStatus, targetStatus string) error {
	curr := strings.ToUpper(strings.TrimSpace(currentStatus))
	target := strings.ToUpper(strings.TrimSpace(targetStatus))

	if curr == target {
		return nil
	}

	valid := false
	switch curr {
	case spec.DocStatusMissing, "":
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusUnderReview || target == spec.DocStatusRequested)
	case spec.DocStatusRequested:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusUnderReview || target == spec.DocStatusMissing)
	case spec.DocStatusUploaded:
		valid = (target == spec.DocStatusUnderReview || target == spec.DocStatusApproved || target == spec.DocStatusRejected || target == spec.DocStatusSuperseded)
	case spec.DocStatusUnderReview:
		valid = (target == spec.DocStatusApproved || target == spec.DocStatusRejected || target == spec.DocStatusUploaded || target == spec.DocStatusSuperseded)
	case spec.DocStatusApproved:
		valid = (target == spec.DocStatusExpired || target == spec.DocStatusUnderReview || target == spec.DocStatusSuperseded)
	case spec.DocStatusRejected:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusUnderReview || target == spec.DocStatusSuperseded)
	case spec.DocStatusExpired:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusUnderReview || target == spec.DocStatusSuperseded)
	case spec.DocStatusSuperseded:
		valid = (target == spec.DocStatusUploaded)
	default:
		valid = true
	}

	if !valid {
		return svcerror.WrapServiceError(
			svcerror.ErrInvalidArgument,
			fmt.Errorf("invalid document status transition from %s to %s", curr, target),
		)
	}
	return nil
}

// docRequirementDef defines a standard operational document rule for checking compliance.
type docRequirementDef struct {
	DocType          string
	Name             string
	Category         string
	RequirementLevel string
	Reason           string
}

// getStandardDocumentRequirements returns the deterministic document rule set required for a shipment.
func getStandardDocumentRequirements(sh *spec.Shipment) []docRequirementDef {
	return []docRequirementDef{
		{
			DocType:          spec.DocTypeMasterBillOfLading,
			Name:             "Master Bill of Lading (MBL) / Waybill",
			Category:         spec.DocCategoryTransport,
			RequirementLevel: spec.RequirementLevelCritical,
			Reason:           "Authoritative carrier title and release document required for transit execution",
		},
		{
			DocType:          spec.DocTypeCommercialInvoice,
			Name:             "Commercial Invoice",
			Category:         spec.DocCategoryCommercial,
			RequirementLevel: spec.RequirementLevelCritical,
			Reason:           "Mandatory for customs valuation, import duties, and financial reconciliation",
		},
		{
			DocType:          spec.DocTypePackingList,
			Name:             "Packing List",
			Category:         spec.DocCategoryCommercial,
			RequirementLevel: spec.RequirementLevelRequired,
			Reason:           "Required for physical container inspection, destination tally, and delivery",
		},
		{
			DocType:          spec.DocTypeCustomsDeclaration,
			Name:             "Customs Declaration / Entry",
			Category:         spec.DocCategoryCustoms,
			RequirementLevel: spec.RequirementLevelRequired,
			Reason:           "Required for border agency compliance and customs clearance release",
		},
		{
			DocType:          spec.DocTypeCertificateOfOrigin,
			Name:             "Certificate of Origin",
			Category:         spec.DocCategoryCommercial,
			RequirementLevel: spec.RequirementLevelOptional,
			Reason:           "Required if preferential tariff or FTA duty concessions apply",
		},
		{
			DocType:          spec.DocTypeInsuranceCert,
			Name:             "Cargo Insurance Certificate",
			Category:         spec.DocCategoryInsurance,
			RequirementLevel: spec.RequirementLevelOptional,
			Reason:           "Verifies marine cargo coverage and risk policy terms",
		},
	}
}

// EvaluateDocumentCompliance determines the real-time compliance readiness, missing documents queue, and risk state of a shipment's documents.
// In simple terms, this function:
// 1. Analyzes all uploaded documents attached to the shipment.
// 2. Checks expiry dates on each document to flag expiring or expired files.
// 3. Compares uploaded files against mandatory shipment requirements (MBL, Commercial Invoice, Packing List, Customs Entry).
// 4. Calculates counts for Available, Under Review, Approved, Rejected, Expired, and Expiring Soon.
// 5. Produces a list of Missing Document Requirements with priority levels (CRITICAL, REQUIRED, OPTIONAL).
// 6. Computes the overall compliance state: COMPLIANT, ATTENTION_REQUIRED, AT_RISK, or BLOCKED.
func EvaluateDocumentCompliance(sh *spec.Shipment, docs []*spec.ShipmentDocument) *spec.ShipmentDocumentComplianceSummary {
	requirements := getStandardDocumentRequirements(sh)

	// Index existing documents by normalized doc_type
	docsByType := make(map[string][]*spec.ShipmentDocument)
	var availableCount, underReviewCount, approvedCount, rejectedCount, expiredCount, expiringSoonCount int

	now := time.Now()
	for _, d := range docs {
		norm := NormalizeDocType(d.DocType)
		docsByType[norm] = append(docsByType[norm], d)

		status := strings.ToUpper(strings.TrimSpace(d.Status))

		// Evaluate expiry relative to now
		valStatus := ComputeDocumentValidityAt(d.ExpiresAt, now)
		d.ValidityStatus = valStatus

		if valStatus == spec.DocValidityExpired && status != spec.DocStatusRejected {
			status = spec.DocStatusExpired
			d.Status = spec.DocStatusExpired
			expiredCount++
		} else if valStatus == spec.DocValidityExpiringSoon && status != spec.DocStatusRejected {
			expiringSoonCount++
		}

		switch status {
		case spec.DocStatusApproved:
			approvedCount++
			availableCount++
		case spec.DocStatusUnderReview:
			underReviewCount++
			availableCount++
		case spec.DocStatusUploaded:
			availableCount++
		case spec.DocStatusRejected:
			rejectedCount++
		case spec.DocStatusExpired:
			// already counted in expiredCount
		}
	}

	missingQueue := make([]*spec.MissingDocumentRequirement, 0, len(requirements))
	var missingRequiredCount int
	var hasCriticalMissing bool
	var hasCriticalRejected bool
	var hasRequiredMissing bool
	var blockerReason string

	for _, req := range requirements {
		existing, ok := docsByType[req.DocType]
		
		isMissing := false
		currStatus := spec.DocStatusMissing

		if !ok || len(existing) == 0 {
			isMissing = true
			currStatus = spec.DocStatusMissing
		} else {
			hasValid := false
			for _, d := range existing {
				st := strings.ToUpper(strings.TrimSpace(d.Status))
				if st == spec.DocStatusRejected {
					currStatus = spec.DocStatusRejected
				} else if st == spec.DocStatusExpired {
					currStatus = spec.DocStatusExpired
				} else if st == spec.DocStatusApproved {
					hasValid = true
					currStatus = spec.DocStatusApproved
				} else if st == spec.DocStatusUnderReview || st == spec.DocStatusUploaded {
					hasValid = true
					if currStatus != spec.DocStatusApproved {
						currStatus = st
					}
				}
			}

			if !hasValid {
				isMissing = true
			}
		}

		if isMissing && req.RequirementLevel != spec.RequirementLevelOptional {
			missingRequiredCount++
			if req.RequirementLevel == spec.RequirementLevelCritical {
				hasCriticalMissing = true
				if currStatus == spec.DocStatusRejected {
					hasCriticalRejected = true
					blockerReason = fmt.Sprintf("Critical requirement '%s' was rejected", req.Name)
				} else if blockerReason == "" {
					blockerReason = fmt.Sprintf("Critical requirement '%s' is missing", req.Name)
				}
			} else {
				hasRequiredMissing = true
			}
		}

		missingQueue = append(missingQueue, &spec.MissingDocumentRequirement{
			DocType:          req.DocType,
			Category:         req.Category,
			Name:             req.Name,
			RequirementLevel: req.RequirementLevel,
			Reason:           req.Reason,
			IsMissing:        isMissing,
			Status:           currStatus,
		})
	}

	// Determine Compliance State
	// BLOCKED: Critical document is rejected or expired
	// ATTENTION_REQUIRED: Required or critical documents are missing
	// AT_RISK: Documents are expiring soon or under review with tight deadlines
	// COMPLIANT: All required documents uploaded, approved, or valid
	complianceState := spec.ComplianceStateCompliant
	if hasCriticalRejected || expiredCount > 0 {
		complianceState = spec.ComplianceStateBlocked
		if blockerReason == "" && expiredCount > 0 {
			blockerReason = fmt.Sprintf("%d operational document(s) have expired", expiredCount)
		}
	} else if hasCriticalMissing || hasRequiredMissing {
		complianceState = spec.ComplianceStateAttentionRequired
		if blockerReason == "" {
			blockerReason = fmt.Sprintf("%d required document(s) missing", missingRequiredCount)
		}
	} else if expiringSoonCount > 0 || underReviewCount > 0 {
		complianceState = spec.ComplianceStateAtRisk
		if blockerReason == "" && expiringSoonCount > 0 {
			blockerReason = fmt.Sprintf("%d document(s) expiring soon", expiringSoonCount)
		}
	}

	return &spec.ShipmentDocumentComplianceSummary{
		TotalRequired:    4, // MBL, Commercial Invoice, Packing List, Customs Entry
		Available:        availableCount,
		Missing:          missingRequiredCount,
		UnderReview:      underReviewCount,
		Approved:         approvedCount,
		Rejected:         rejectedCount,
		Expired:          expiredCount,
		ExpiringSoon:     expiringSoonCount,
		ComplianceState:  complianceState,
		BlockerReason:    blockerReason,
		MissingDocuments: missingQueue,
	}
}
