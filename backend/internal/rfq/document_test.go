package rfq_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) rfq.BusinessLogic {
	db, err := database.Connect("root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true")
	require.NoError(t, err)

	// Seed test orgs
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (7701, 'Test Org 7701', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (7702, 'Test Org 7702', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (7703, 'Test Org 7703', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")
	_, _ = db.Exec("INSERT INTO organizations (id, name, created_at, updated_at) VALUES (7704, 'Test Org 7704', NOW(), NOW()) ON DUPLICATE KEY UPDATE name=VALUES(name)")

	dl := rfq.NewDataLayer(db)
	bus := events.NewInProcessBus()
	return rfq.NewBusinessLogic(dl, bus, nil, nil, nil, nil)
}



func TestDocument_Lifecycle_Transitions(t *testing.T) {
	// Test valid transitions
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusMissing, spec.DocStatusRequested))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusMissing, spec.DocStatusUploaded))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusRequested, spec.DocStatusUploaded))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusUploaded, spec.DocStatusUnderReview))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusUploaded, spec.DocStatusApproved))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusUploaded, spec.DocStatusRejected))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusUnderReview, spec.DocStatusApproved))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusUnderReview, spec.DocStatusRejected))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusRejected, spec.DocStatusUploaded))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusApproved, spec.DocStatusExpired))
	assert.NoError(t, rfq.ValidateStatusTransition(spec.DocStatusApproved, spec.DocStatusUnderReview))

	// Test invalid transitions
	assert.Error(t, rfq.ValidateStatusTransition(spec.DocStatusMissing, spec.DocStatusApproved))
	assert.Error(t, rfq.ValidateStatusTransition(spec.DocStatusMissing, spec.DocStatusUnderReview))
	assert.Error(t, rfq.ValidateStatusTransition(spec.DocStatusRejected, spec.DocStatusApproved))
}

func TestDocument_FullWorkflowAndOrgIsolation(t *testing.T) {
	bl := setupTestDB(t)
	ctx := context.Background()

	orgID := int32(7701)
	otherOrgID := int32(7702)

	// Create test RFQ in org 7701
	origin := "INNSA"
	dest := "DEHAM"
	incoterms := "FOB"
	targetDate := time.Now().Add(14 * 24 * time.Hour)
	rfqCreated, err := bl.CreateRFQ(ctx, spec.CreateRFQRequest{
		OrgID:       orgID,
		CustomerID:  1,
		Origin:      &origin,
		Destination: &dest,
		Incoterms:   &incoterms,
		TargetDate:  &targetDate,
		Items: []spec.RFQItem{
			{Description: "Industrial Machinery Parts", Quantity: 10},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, rfqCreated)

	// 1. Initial State: Commercial Invoice and Packing List are MISSING
	docsResp, err := bl.GetDocuments(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, docsResp.Summary.TotalDocuments)
	assert.Equal(t, 2, docsResp.Summary.RequiredDocuments)
	assert.Equal(t, 2, docsResp.Summary.MissingDocuments)
	assert.Equal(t, 0, docsResp.Summary.ReadinessPercentage)

	// 2. Attach / Upload Commercial Invoice
	fileName := "commercial_invoice_7701.pdf"
	fileURL := "http://localhost:8080/uploads/commercial_invoice_7701.pdf"
	ciDoc, err := bl.CreateDocument(ctx, orgID, rfqCreated.ID, spec.CreateDocumentRequest{
		DocumentType: spec.DocTypeCommercialInvoice,
		DocumentName: "Commercial Invoice - INV-2026-001",
		FileName:     &fileName,
		FileURL:      &fileURL,
	}, "John Operator")
	require.NoError(t, err)
	assert.Equal(t, spec.DocStatusUploaded, ciDoc.Status)
	assert.Equal(t, spec.DocTypeCommercialInvoice, ciDoc.DocumentType)

	// 3. Verify Documents summary after upload
	docsResp2, err := bl.GetDocuments(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, docsResp2.Summary.TotalDocuments)
	assert.Equal(t, 1, docsResp2.Summary.UnderReviewDocuments)

	// 4. Move to UNDER_REVIEW
	ciDocReviewed, err := bl.UpdateDocumentStatus(ctx, orgID, rfqCreated.ID, ciDoc.ID, spec.UpdateDocumentStatusRequest{
		Status: spec.DocStatusUnderReview,
	}, "Alice Reviewer")
	require.NoError(t, err)
	assert.Equal(t, spec.DocStatusUnderReview, ciDocReviewed.Status)

	// 5. Approve Commercial Invoice
	ciDocApproved, err := bl.UpdateDocumentStatus(ctx, orgID, rfqCreated.ID, ciDoc.ID, spec.UpdateDocumentStatusRequest{
		Status: spec.DocStatusApproved,
	}, "Alice Reviewer")
	require.NoError(t, err)
	assert.Equal(t, spec.DocStatusApproved, ciDocApproved.Status)

	// 6. Verify Requirements Engine reflects SATISFIED for Commercial Invoice
	reqsResp, err := bl.GetRequirements(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	var ciReq *spec.DocumentRequirement
	for _, d := range reqsResp.DocumentRequirements {
		if d.DocType == spec.ReqDocCommercialInvoice {
			ciReq = &d
			break
		}
	}
	require.NotNil(t, ciReq)
	assert.Equal(t, spec.ReqStatusSatisfied, ciReq.Status)
	assert.Equal(t, spec.DocStatusApproved, ciReq.DocumentStatus)

	// 7. Verify Activity Timeline has real DOCUMENT_UPLOADED and DOCUMENT_APPROVED events
	actResp, err := bl.GetActivity(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)
	hasDocUploaded := false
	hasDocApproved := false
	for _, ev := range actResp.Events {
		if ev.Type == spec.ActivityDocumentUploaded {
			hasDocUploaded = true
		}
		if ev.Type == spec.ActivityDocumentApproved {
			hasDocApproved = true
		}
	}
	assert.True(t, hasDocUploaded, "Activity timeline must contain DOCUMENT_UPLOADED event")
	assert.True(t, hasDocApproved, "Activity timeline must contain DOCUMENT_APPROVED event")

	// 8. Strict Multi-Tenant Organization Isolation
	// Org 7702 must NOT be able to access or mutate Org 7701's documents
	_, err = bl.GetDocuments(ctx, otherOrgID, rfqCreated.ID)
	assert.Error(t, err, "Org 7702 must not get Org 7701 documents")

	_, err = bl.CreateDocument(ctx, otherOrgID, rfqCreated.ID, spec.CreateDocumentRequest{
		DocumentType: spec.DocTypePackingList,
		DocumentName: "Hacked Packing List",
	}, "Attacker")
	assert.Error(t, err, "Org 7702 must not create document on Org 7701 RFQ")

	_, err = bl.UpdateDocumentStatus(ctx, otherOrgID, rfqCreated.ID, ciDoc.ID, spec.UpdateDocumentStatusRequest{
		Status: spec.DocStatusRejected,
	}, "Attacker")
	assert.Error(t, err, "Org 7702 must not update document on Org 7701 RFQ")

	err = bl.DeleteDocument(ctx, otherOrgID, rfqCreated.ID, ciDoc.ID)
	assert.Error(t, err, "Org 7702 must not delete document on Org 7701 RFQ")

	// Clean up created document
	err = bl.DeleteDocument(ctx, orgID, rfqCreated.ID, ciDoc.ID)
	assert.NoError(t, err)
}

func TestDocument_FutureStage_NonBlocking(t *testing.T) {
	bl := setupTestDB(t)
	ctx := context.Background()

	orgID := int32(7703)
	origin := "INNSA"
	dest := "DEHAM"
	incoterms := "FOB"
	targetDate := time.Now().Add(14 * 24 * time.Hour)

	rfqCreated, err := bl.CreateRFQ(ctx, spec.CreateRFQRequest{
		OrgID:       orgID,
		CustomerID:  1,
		Origin:      &origin,
		Destination: &dest,
		Incoterms:   &incoterms,
		TargetDate:  &targetDate,
		Items: []spec.RFQItem{
			{Description: "Machinery", Quantity: 5},
		},
	})
	require.NoError(t, err)

	docsResp, err := bl.GetDocuments(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)

	// Verify future stage documents are categorized under FutureStageDocuments
	assert.True(t, len(docsResp.FutureStageDocuments) > 0, "Must have future stage documents")
	hasBL := false
	for _, fDoc := range docsResp.FutureStageDocuments {
		if fDoc.DocType == spec.ReqDocBillOfLading || fDoc.DocType == spec.ReqDocHBL || fDoc.DocType == spec.ReqDocMBL {
			hasBL = true
			assert.Equal(t, spec.ReqStatusNotApplicable, fDoc.Status)
		}
	}
	assert.True(t, hasBL, "Must contain Bill of Lading in future stage documents")
	assert.Equal(t, 2, docsResp.Summary.RequiredDocuments, "Only current stage docs are counted in required summary")
}

func TestDocument_DangerousGoods_MSDS_Requirement(t *testing.T) {
	bl := setupTestDB(t)
	ctx := context.Background()

	orgID := int32(7704)
	origin := "INNSA"
	dest := "SGSIN"
	incoterms := "CIF"
	targetDate := time.Now().Add(10 * 24 * time.Hour)

	// Create RFQ with hazardous chemical cargo
	rfqCreated, err := bl.CreateRFQ(ctx, spec.CreateRFQRequest{
		OrgID:       orgID,
		CustomerID:  1,
		Origin:      &origin,
		Destination: &dest,
		Incoterms:   &incoterms,
		TargetDate:  &targetDate,
		Items: []spec.RFQItem{
			{Description: "Flammable solvent UN1263 Class 3 Dangerous Goods", Quantity: 20},
		},
	})
	require.NoError(t, err)

	docsResp, err := bl.GetDocuments(ctx, orgID, rfqCreated.ID)
	require.NoError(t, err)

	// Verify Dangerous Goods Declaration is activated under ConditionalDocuments
	hasDGDoc := false
	for _, cDoc := range docsResp.ConditionalDocuments {
		if cDoc.DocType == spec.ReqDocDGDeclaration {
			hasDGDoc = true
			assert.True(t, cDoc.IsRequired, "DG Declaration must be required for dangerous goods cargo")
			assert.Equal(t, spec.ReqStatusMissing, cDoc.Status)
		}
	}
	assert.True(t, hasDGDoc, "Conditional documents must include dg_declaration")
}

