package rfq

// document_lifecycle.go — Task 12: RFQ Document Lifecycle Validator & Resolution
//
// Enforces deterministic state transitions for RFQ documents and provides
// resolution helpers for mapping document requirements to real DB records.

import (
	"fmt"
	"strings"

	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/svcerror"
)


// ValidateStatusTransition ensures that document lifecycle state transitions
// adhere strictly to freight operational audit standards.
//
// Valid paths:
//   MISSING      → REQUESTED | UPLOADED
//   REQUESTED    → UPLOADED | NOT_REQUIRED
//   UPLOADED     → UNDER_REVIEW | APPROVED | REJECTED
//   UNDER_REVIEW → APPROVED | REJECTED | UPLOADED
//   APPROVED     → EXPIRED | UNDER_REVIEW
//   REJECTED     → UPLOADED | REQUESTED
//   EXPIRED      → UPLOADED | REQUESTED
//   NOT_REQUIRED → REQUESTED | UPLOADED
func ValidateStatusTransition(currentStatus, newStatus string) error {
	curr := strings.ToUpper(strings.TrimSpace(currentStatus))
	target := strings.ToUpper(strings.TrimSpace(newStatus))

	if curr == target {
		return nil
	}

	valid := false
	switch curr {
	case spec.DocStatusMissing, "":
		valid = (target == spec.DocStatusRequested || target == spec.DocStatusUploaded)
	case spec.DocStatusRequested:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusNotRequired)
	case spec.DocStatusUploaded:
		valid = (target == spec.DocStatusUnderReview || target == spec.DocStatusApproved || target == spec.DocStatusRejected)
	case spec.DocStatusUnderReview:
		valid = (target == spec.DocStatusApproved || target == spec.DocStatusRejected || target == spec.DocStatusUploaded)
	case spec.DocStatusApproved:
		valid = (target == spec.DocStatusExpired || target == spec.DocStatusUnderReview)
	case spec.DocStatusRejected:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusRequested)
	case spec.DocStatusExpired:
		valid = (target == spec.DocStatusUploaded || target == spec.DocStatusRequested)
	case spec.DocStatusNotRequired:
		valid = (target == spec.DocStatusRequested || target == spec.DocStatusUploaded)
	}

	if !valid {
		return svcerror.WrapServiceError(
			svcerror.ErrInvalidArgument,
			fmt.Errorf("invalid document status transition from %s to %s", curr, target),
		)
	}

	return nil
}

// NormalizeDocType converts different casings or aliases to canonical uppercase DocumentType.
func NormalizeDocType(raw string) string {
	cleaned := strings.ToUpper(strings.TrimSpace(raw))
	cleaned = strings.ReplaceAll(cleaned, "-", "_")
	cleaned = strings.ReplaceAll(cleaned, " ", "_")

	switch cleaned {
	case spec.DocTypeCommercialInvoice, spec.DocAliasInvoice, spec.DocAliasCI:
		return spec.DocTypeCommercialInvoice
	case spec.DocTypePackingList, spec.DocAliasPL, spec.DocAliasPackingListNoSpace:
		return spec.DocTypePackingList
	case spec.DocTypeProformaInvoice, spec.DocAliasPI, spec.DocAliasProforma:
		return spec.DocTypeProformaInvoice
	case spec.DocTypePurchaseOrder, spec.DocAliasPO:
		return spec.DocTypePurchaseOrder
	case spec.DocTypeSalesContract, spec.DocAliasSC, spec.DocAliasContract:
		return spec.DocTypeSalesContract
	case spec.DocTypeBillOfLading, spec.DocAliasBL, spec.DocAliasBOL:
		return spec.DocTypeBillOfLading
	case spec.DocTypeHBL, spec.DocAliasHouseBL, spec.DocAliasHouseBillOfLading:
		return spec.DocTypeHBL
	case spec.DocTypeMBL, spec.DocAliasMasterBL, spec.DocAliasMasterBillOfLading:
		return spec.DocTypeMBL
	case spec.DocTypeAirWaybill, spec.DocAliasAWB, spec.DocAliasHAWB, spec.DocAliasMAWB:
		return spec.DocTypeAirWaybill
	case spec.DocTypeSeaWaybill, spec.DocAliasSWB:
		return spec.DocTypeSeaWaybill
	case spec.DocTypeDeliveryOrder, spec.DocAliasDO:
		return spec.DocTypeDeliveryOrder
	case spec.DocTypeCertificateOfOrigin, spec.DocAliasCOO:
		return spec.DocTypeCertificateOfOrigin
	case spec.DocTypeExportDeclaration, spec.DocAliasExportCustoms:
		return spec.DocTypeExportDeclaration
	case spec.DocTypeImportDeclaration, spec.DocAliasImportCustoms:
		return spec.DocTypeImportDeclaration
	case spec.DocTypeCustomsEntry, spec.DocAliasCustomsClearance:
		return spec.DocTypeCustomsEntry
	case spec.DocTypeCommercialLicense:
		return spec.DocTypeCommercialLicense
	case spec.DocTypeMSDS, spec.DocAliasMaterialSafetyDataSheet:
		return spec.DocTypeMSDS
	case spec.DocTypeDangerousGoodsDeclaration, spec.DocAliasDGDeclaration, spec.DocAliasDGD:
		return spec.DocTypeDangerousGoodsDeclaration
	case spec.DocTypeIMDGDeclaration, spec.DocAliasIMDG:
		return spec.DocTypeIMDGDeclaration
	case spec.DocTypeCargoSpecification, spec.DocAliasSpecification:
		return spec.DocTypeCargoSpecification
	case spec.DocTypeProductCertificate:
		return spec.DocTypeProductCertificate
	case spec.DocTypeQualityCertificate:
		return spec.DocTypeQualityCertificate
	case spec.DocTypeInspectionCertificate:
		return spec.DocTypeInspectionCertificate
	case spec.DocTypeInsuranceCertificate, spec.DocAliasInsurance:
		return spec.DocTypeInsuranceCertificate
	case spec.DocTypeCargoInsurancePolicy:
		return spec.DocTypeCargoInsurancePolicy
	case spec.DocTypeBookingConfirmation, spec.DocAliasBC:
		return spec.DocTypeBookingConfirmation
	case spec.DocTypeShippingInstruction, spec.DocAliasSI:
		return spec.DocTypeShippingInstruction
	case spec.DocTypeContainerInspection:
		return spec.DocTypeContainerInspection
	case spec.DocTypeVGMDeclaration, spec.DocAliasVGM:
		return spec.DocTypeVGMDeclaration
	default:
		if cleaned == "" {
			return spec.DocTypeOther
		}
		return cleaned
	}
}


// BuildGetDocumentsResponse constructs a full GetDocumentsResponse from an RFQ,
// its resolved document requirements, and all attached document records.
func BuildGetDocumentsResponse(rfq *spec.RFQ, docReqs []spec.DocumentRequirement, allDocs []spec.RFQDocument) *spec.GetDocumentsResponse {
	// Index real docs by normalized doc_type
	docsByType := make(map[string]*spec.RFQDocument)
	for i := range allDocs {
		d := &allDocs[i]
		normType := NormalizeDocType(d.DocumentType)
		// Prefer APPROVED, then UNDER_REVIEW, then UPLOADED, then newest
		existing, ok := docsByType[normType]
		if !ok || isHigherPriorityStatus(d.Status, existing.Status) {
			docsByType[normType] = d
		}
	}

	currentStage := []spec.ResolvedDocumentRequirement{}
	conditional := []spec.ResolvedDocumentRequirement{}
	futureStage := []spec.ResolvedDocumentRequirement{}

	var reqCount, receivedCount, missingCount, underReviewCount, approvedCount, rejectedCount, futureCount int

	for _, req := range docReqs {
		normType := NormalizeDocType(req.DocType)
		matchedDoc := docsByType[normType]

		res := spec.ResolvedDocumentRequirement{
			DocumentRequirement: req,
			DocumentRecord:      matchedDoc,
		}

		if matchedDoc != nil {
			res.DocumentID = &matchedDoc.ID
			res.DocumentStatus = matchedDoc.Status
			res.FileName = matchedDoc.FileName
			res.FileURL = matchedDoc.FileURL
			res.UploadedAt = matchedDoc.UploadedAt
			res.ReviewedAt = matchedDoc.ReviewedAt
		}

		if req.ApplicableStage == spec.DocStageRFQ || req.ApplicableStage == spec.DocStageQuotation {
			if req.IsConditional {
				conditional = append(conditional, res)
			} else {
				currentStage = append(currentStage, res)
			}
		} else {
			futureStage = append(futureStage, res)
			futureCount++
		}

		// Count metrics
		if req.IsRequired && (req.ApplicableStage == spec.DocStageRFQ || req.ApplicableStage == spec.DocStageQuotation) {
			reqCount++
			if matchedDoc != nil && matchedDoc.Status == spec.DocStatusApproved {
				receivedCount++
				approvedCount++
			} else if matchedDoc != nil && (matchedDoc.Status == spec.DocStatusUnderReview || matchedDoc.Status == spec.DocStatusUploaded) {
				underReviewCount++
			} else if matchedDoc != nil && matchedDoc.Status == spec.DocStatusRejected {
				rejectedCount++
				missingCount++
			} else {
				missingCount++
			}
		} else if matchedDoc != nil {
			if matchedDoc.Status == spec.DocStatusApproved {
				approvedCount++
			} else if matchedDoc.Status == spec.DocStatusUnderReview || matchedDoc.Status == spec.DocStatusUploaded {
				underReviewCount++
			} else if matchedDoc.Status == spec.DocStatusRejected {
				rejectedCount++
			}
		}
	}

	readinessPct := 100
	if reqCount > 0 {
		readinessPct = int((float64(receivedCount) / float64(reqCount)) * 100.0)
		if readinessPct > 100 {
			readinessPct = 100
		}
	}

	summary := spec.DocumentSummary{
		TotalDocuments:       len(allDocs),
		RequiredDocuments:    reqCount,
		ReceivedDocuments:    receivedCount,
		MissingDocuments:     missingCount,
		UnderReviewDocuments: underReviewCount,
		ApprovedDocuments:    approvedCount,
		RejectedDocuments:    rejectedCount,
		FutureStageDocuments: futureCount,
		ReadinessPercentage:  readinessPct,
	}

	return &spec.GetDocumentsResponse{
		Summary:               summary,
		CurrentStageDocuments: currentStage,
		ConditionalDocuments:  conditional,
		FutureStageDocuments:  futureStage,
		AllDocuments:          allDocs,
	}
}

func isHigherPriorityStatus(s1, s2 string) bool {
	prio := map[string]int{
		spec.DocStatusApproved:    5,
		spec.DocStatusUnderReview: 4,
		spec.DocStatusUploaded:    3,
		spec.DocStatusRequested:   2,
		spec.DocStatusRejected:    1,
		spec.DocStatusMissing:     0,
	}
	return prio[s1] > prio[s2]
}
