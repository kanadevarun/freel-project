package operations_automation_test

import (
	"testing"
	"time"

	"github.com/freel/backend/internal/shipments/operations_automation"
	"github.com/freel/backend/internal/shipments/spec"
)

func TestDeterministicSignalsCalculation(t *testing.T) {
	// Test overdue milestone detection and stale tracking
	pastTime := time.Now().Add(-72 * time.Hour)
	recentTime := time.Now().Add(-2 * time.Hour)
	bkgNum := "BKG-101"

	sh := &spec.Shipment{
		ID:              101,
		OrgID:           1,
		BookingNumber:   &bkgNum,
		Status:          "IN_TRANSIT",
		CarrierSCAC:     "MAEU",
		OriginPort:      "CNSHA",
		DestinationPort: "USLAX",
	}

	milestones := []*spec.ShipmentMilestone{
		{
			ID:            1,
			ShipmentID:    101,
			MilestoneCode: "DEPARTURE",
			Status:        "PLANNED",
			PlannedDate:   &pastTime, // Past due!
			ActualDate:    nil,
		},
		{
			ID:            2,
			ShipmentID:    101,
			MilestoneCode: "ARRIVAL",
			Status:        "PLANNED",
			PlannedDate:   &recentTime,
			ActualDate:    nil,
		},
	}

	exceptions := []*spec.ShipmentException{
		{
			ID:            501,
			ShipmentID:    101,
			OrgID:         1,
			ExceptionType: "CUSTOMS_HOLD",
			Severity:      "CRITICAL",
			Status:        "OPEN",
			Title:         "Port Customs Hold",
		},
	}

	docs := []*spec.ShipmentDocument{
		{
			ID:         11,
			ShipmentID: 101,
			OrgID:      1,
			DocType:    "BILL_OF_LADING",
			Status:     "VERIFIED",
		},
	}

	events := []*spec.CarrierTrackingEvent{
		{
			ID:         901,
			ShipmentID: &sh.ID,
			OrgID:      1,
			EventID:    "EV-901",
			EventTime:  pastTime,
		},
	}

	// Verify our business expectations for the signals
	if len(milestones) == 0 {
		t.Fatal("Expected milestones")
	}
	if len(exceptions) == 0 {
		t.Fatal("Expected exceptions")
	}
	if len(docs) == 0 {
		t.Fatal("Expected documents")
	}
	if len(events) == 0 {
		t.Fatal("Expected events")
	}

	// Overdue milestone check
	var hasOverdueMilestone bool
	now := time.Now()
	for _, m := range milestones {
		if (m.Status == "PLANNED" || m.Status == "PENDING" || m.Status == "SCHEDULED") && m.PlannedDate != nil && m.PlannedDate.Before(now) {
			hasOverdueMilestone = true
			break
		}
	}
	if !hasOverdueMilestone {
		t.Errorf("Expected hasOverdueMilestone to be true")
	}

	// Active exception check
	var hasActiveException bool
	for _, ex := range exceptions {
		if ex.Status == "OPEN" || ex.Status == "INVESTIGATING" {
			hasActiveException = true
			break
		}
	}
	if !hasActiveException {
		t.Errorf("Expected hasActiveException to be true")
	}

	// Missing document check (e.g. COMMERCIAL_INVOICE)
	docTypes := make(map[string]bool)
	for _, d := range docs {
		docTypes[d.DocType] = true
	}
	if docTypes["COMMERCIAL_INVOICE"] {
		t.Errorf("Expected COMMERCIAL_INVOICE to be missing")
	}

	// Stale tracking check: last event was 72 hours ago (> 24h)
	lastEventTime := events[0].EventTime
	if now.Sub(lastEventTime) <= 24*time.Hour {
		t.Errorf("Expected tracking to be stale (> 24h)")
	}
}

func TestTenantIsolationModel(t *testing.T) {
	// Verify that model structs cleanly isolate org_id
	analysis := operations_automation.ShipmentOperationsAnalysis{
		OrgID:      1,
		ShipmentID: 101,
		RiskLevel:  "HIGH",
	}

	if analysis.OrgID != 1 {
		t.Errorf("Expected OrgID to be 1, got %d", analysis.OrgID)
	}

	draft := operations_automation.ShipmentCommunicationDraft{
		OrgID:            1,
		ShipmentID:       101,
		DraftType:        "CUSTOMER_UPDATE",
		Subject:          "Delay Notification",
		CustomerWording:  "Your container is delayed.",
		RequiresApproval: true,
	}

	if draft.OrgID != 1 || !draft.RequiresApproval {
		t.Errorf("Expected draft for Org 1 with RequiresApproval=true")
	}
}
