package shipments_test

import (
	"context"
	"testing"
	"time"

	"github.com/freel/backend/internal/database"
	"github.com/freel/backend/internal/rfq"
	"github.com/freel/backend/internal/rfq/spec"
	"github.com/freel/backend/internal/shipments"
	shipmentspec "github.com/freel/backend/internal/shipments/spec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShipmentFoundationAndSync(t *testing.T) {
	// Connect to local test DB
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up database tables
	_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")

	// Seed Organizations
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, created_at, updated_at) 
		VALUES (9991, 'Org A', NOW(), NOW()), (9992, 'Org B', NOW(), NOW())
		ON DUPLICATE KEY UPDATE name = VALUES(name)
	`)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")
	}()

	// Seed Customers
	_, err = db.Exec(`
		INSERT INTO customers (id, org_id, name, contact_email)
		VALUES 
			(9301, 9991, 'Customer A', 'cust-a@example.com'),
			(9302, 9992, 'Customer B', 'cust-b@example.com')
	`)
	require.NoError(t, err)

	// Seed Carriers
	_, _ = db.Exec("INSERT INTO carriers (scac, name) VALUES ('MAEU', 'Maersk Line') ON DUPLICATE KEY UPDATE name = VALUES(name)")

	// Seed RFQs under Org A (9991) and Org B (9992)
	_, err = db.Exec(`
		INSERT INTO rfqs (id, org_id, rfq_number, customer_id, origin, destination, incoterms, status, stage)
		VALUES 
			(9001, 9991, 'RFQ-9001', 9301, 'INNSA', 'DEHAM', 'FOB', 'WON', 'COMPLETED'),
			(9002, 9992, 'RFQ-9002', 9302, 'INNSA', 'NLRTM', 'FOB', 'WON', 'COMPLETED')
	`)
	require.NoError(t, err)

	// Seed RFQ Quotes
	_, err = db.Exec(`
		INSERT INTO rfq_quotes (id, rfq_id, carrier_name, transit_time_days, sell_price, status)
		VALUES 
			(9101, 9001, 'Maersk Line', 14, 2500.0, 'APPROVED'),
			(9102, 9002, 'Maersk Line', 14, 2800.0, 'APPROVED')
	`)
	require.NoError(t, err)

	// Seed Bookings
	_, err = db.Exec(`
		INSERT INTO bookings (id, org_id, rfq_id, quote_id, booking_number, carrier_name, carrier_scac, status, origin_port, destination_port, vessel_name, voyage_number)
		VALUES 
			(9201, 9991, 9001, 9101, 'BKG-9201', 'Maersk Line', 'MAEU', 'CONFIRMED', 'INNSA', 'DEHAM', 'Maersk Hamburg', '2601W'),
			(9202, 9991, 9001, 9101, 'BKG-9202', 'Maersk Line', 'MAEU', 'DRAFT', 'INNSA', 'DEHAM', 'Maersk Hamburg', '2601W'),
			(9203, 9992, 9002, 9102, 'BKG-9203', 'Maersk Line', 'MAEU', 'CONFIRMED', 'INNSA', 'NLRTM', 'Maersk Rotterdam', '2602W'),
			(9204, 9991, 9002, 9102, 'BKG-9204', 'Maersk Line', 'MAEU', 'CONFIRMED', 'INNSA', 'NLRTM', 'Maersk Rotterdam', '2602W') -- Mismatched RFQ org lineage (booking is Org A, RFQ is Org B)
	`)
	require.NoError(t, err)

	// Initialize Datalayer & Services
	rfqDL := rfq.NewDataLayer(db)
	shipRepo := shipments.NewRepository(db)
	shipSvc := shipments.NewService(shipRepo, db, nil, "http://localhost:8080")
	ctx := context.Background()

	t.Run("Confirmed booking handoff creates shipment successfully", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{
			ContainerNumbers: []string{"MSKU1111222"},
		}
		sh, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9201, req, "Test User")
		require.NoError(t, err)
		require.NotNil(t, sh)
		assert.Equal(t, "BOOKED", sh.Status)
		assert.Equal(t, "BKG-9201", *sh.BookingNumber)
		assert.Equal(t, "MAEU", sh.CarrierSCAC)

		// Seed standard milestones manually for this shipment (mimics won rfq seed)
		err = db.GetContext(ctx, &sh.ID, "SELECT id FROM shipments WHERE booking_id = ? AND org_id = ?", 9201, 9991)
		require.NoError(t, err)

		_, err = db.Exec(`
			INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, status)
			VALUES 
				(?, 'BOOKED', 'Confirmed by shipping line', NOW(), 'PLANNED'),
				(?, 'DEPARTED', 'Vessel departed origin port', NOW(), 'PLANNED')
		`, sh.ID, sh.ID)
		require.NoError(t, err)
	})

	t.Run("Handoff from non-confirmed booking is rejected", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{}
		sh, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9202, req, "Test User")
		assert.Error(t, err)
		assert.Nil(t, sh)
		assert.Contains(t, err.Error(), "must be in CONFIRMED status")
	})

	t.Run("Handoff from non-existent booking is rejected", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{}
		sh, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9999, req, "Test User")
		assert.Error(t, err)
		assert.Nil(t, sh)
		assert.Contains(t, err.Error(), "booking 9999 not found")
	})

	t.Run("Handoff with mismatched RFQ lineage (unauthorized RFQ org) is rejected", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{}
		sh, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9204, req, "Test User")
		assert.Error(t, err)
		assert.Nil(t, sh)
		assert.Contains(t, err.Error(), "invalid lineage: RFQ 9002 not found or unauthorized")
	})

	t.Run("Idempotency: duplicate request returns the existing shipment", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{}
		sh1, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9201, req, "Test User")
		require.NoError(t, err)
		sh2, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9991, 9201, req, "Test User")
		require.NoError(t, err)
		assert.Equal(t, sh1.ID, sh2.ID)
	})

	t.Run("Tenant isolation: Org B cannot access or create shipments on Org A bookings", func(t *testing.T) {
		req := spec.CreateShipmentFromBookingRequest{}
		sh, err := rfqDL.CreateShipmentFromBookingTx(ctx, 9992, 9201, req, "Test User")
		assert.Error(t, err)
		assert.Nil(t, sh)
		assert.Contains(t, err.Error(), "booking 9201 not found")
	})

	t.Run("UpdateMilestone tenant safety check", func(t *testing.T) {
		// Fetch shipment ID created under Org A
		var shipmentID int64
		err = db.GetContext(ctx, &shipmentID, "SELECT id FROM shipments WHERE booking_id = ? AND org_id = ?", 9201, 9991)
		require.NoError(t, err)

		now := time.Now()
		location := "Port of Mumbai"
		notes := "Arrived at terminal"

		// Org B attempts to update milestone for Org A's shipment -> must fail
		err = shipSvc.UpdateMilestone(ctx, 9992, shipmentID, "DEPARTED", &now, &location, &notes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "access denied")

		// Org A updates milestone successfully
		err = shipSvc.UpdateMilestone(ctx, 9991, shipmentID, "DEPARTED", &now, &location, &notes)
		assert.NoError(t, err)

		// Verify status update activity event was logged in activities table
		var activityLogged bool
		err = db.GetContext(ctx, &activityLogged, `
			SELECT EXISTS(
				SELECT 1 FROM activities 
				WHERE org_id = 9991 AND entity_type = 'SHIPMENT' AND entity_id = ? AND action = 'SHIPMENT_STATUS_UPDATED'
			)
		`, shipmentID)
		require.NoError(t, err)
		assert.True(t, activityLogged, "Status update activity event must be recorded in unified activities timeline")
	})

	t.Run("GetShipmentsWorkspace search, status filter, pagination, and tenant isolation", func(t *testing.T) {
		// Org A queries workspace
		statusFilter := "DEPARTED"
		filter := shipmentspec.ShipmentListFilter{
			Page:   1,
			Limit:  5,
			Status: &statusFilter,
		}

		list, kpis, total, err := shipRepo.GetShipmentsWorkspace(ctx, 9991, filter)
		require.NoError(t, err)
		assert.NotEmpty(t, list)
		assert.Equal(t, int64(1), kpis.TotalShipments)
		assert.Equal(t, int64(1), kpis.InTransit) // "DEPARTED" transitioned status to "DEPARTED", wait is DEPARTED in_transit? No, status is DEPARTED. Let's check status count.
		assert.Equal(t, int64(1), kpis.TotalShipments)
		assert.Equal(t, 1, total)

		// Search filter
		searchVal := "BKG-9201"
		filterSearch := shipmentspec.ShipmentListFilter{
			Page:   1,
			Limit:  5,
			Search: &searchVal,
		}
		listSearch, _, totalSearch, err := shipRepo.GetShipmentsWorkspace(ctx, 9991, filterSearch)
		require.NoError(t, err)
		assert.Len(t, listSearch, 1)
		assert.Equal(t, 1, totalSearch)

		// Pagination limit test
		filterPage := shipmentspec.ShipmentListFilter{
			Page:  1,
			Limit: 1,
		}
		listPage, _, totalPage, err := shipRepo.GetShipmentsWorkspace(ctx, 9991, filterPage)
		require.NoError(t, err)
		assert.Len(t, listPage, 1)
		assert.Equal(t, 1, totalPage)

		// Tenant isolation test: Org B queries workspace -> should get 0 Org A shipments
		listOrgB, kpisOrgB, totalOrgB, err := shipRepo.GetShipmentsWorkspace(ctx, 9992, filterSearch)
		require.NoError(t, err)
		assert.Empty(t, listOrgB)
		assert.Equal(t, int64(0), kpisOrgB.TotalShipments)
		assert.Equal(t, 0, totalOrgB)
	})
}

func TestManualMilestoneUpdate(t *testing.T) {
	// Connect to local test DB
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up database tables
	_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")

	// Seed Organizations
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, created_at, updated_at) 
		VALUES (9991, 'Org A', NOW(), NOW()), (9992, 'Org B', NOW(), NOW())
	`)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")
	}()

	// Seed Customers
	_, err = db.Exec(`
		INSERT INTO customers (id, org_id, name, contact_email)
		VALUES 
			(9301, 9991, 'Customer A', 'cust-a@example.com'),
			(9302, 9992, 'Customer B', 'cust-b@example.com')
	`)
	require.NoError(t, err)

	// Seed Carriers
	_, _ = db.Exec("INSERT INTO carriers (scac, name) VALUES ('MAEU', 'Maersk Line') ON DUPLICATE KEY UPDATE name = VALUES(name)")

	// Seed RFQs
	_, err = db.Exec(`
		INSERT INTO rfqs (id, org_id, rfq_number, customer_id, origin, destination, incoterms, status, stage)
		VALUES 
			(9001, 9991, 'RFQ-9001', 9301, 'INNSA', 'DEHAM', 'FOB', 'WON', 'COMPLETED')
	`)
	require.NoError(t, err)

	// Seed RFQ Quotes
	_, err = db.Exec(`
		INSERT INTO rfq_quotes (id, rfq_id, carrier_name, transit_time_days, sell_price, status)
		VALUES (9101, 9001, 'Maersk Line', 14, 2500.0, 'APPROVED')
	`)
	require.NoError(t, err)

	// Seed Bookings
	_, err = db.Exec(`
		INSERT INTO bookings (id, org_id, rfq_id, quote_id, booking_number, carrier_name, carrier_scac, status, origin_port, destination_port, vessel_name, voyage_number)
		VALUES (9201, 9991, 9001, 9101, 'BKG-9201', 'Maersk Line', 'MAEU', 'CONFIRMED', 'INNSA', 'DEHAM', 'Maersk Hamburg', '2601W')
	`)
	require.NoError(t, err)

	// Initialize DL and Service
	shipRepo := shipments.NewRepository(db)
	shipSvc := shipments.NewService(shipRepo, db, nil, "http://localhost:8080")
	ctx := context.Background()

	// 1. Automatically create Shipment from Won RFQ Handoff
	sh, err := shipSvc.CreateFromRFQ(ctx, 9001)
	require.NoError(t, err)
	require.NotNil(t, sh)
	assert.Equal(t, "BOOKING_PENDING", sh.Status)

	// Verify milestones were seeded
	milestones, err := shipSvc.GetMilestones(ctx, sh.ID)
	require.NoError(t, err)
	assert.Len(t, milestones, 5)

	// 2. Perform valid milestone completion (DEPARTED) by authorized operator
	now := time.Now().Round(time.Second)
	location := "INNSA"
	notes := "Vessel departedNhava Sheva"
	err = shipSvc.UpdateMilestone(ctx, 9991, sh.ID, "DEPARTED", &now, &location, &notes)
	require.NoError(t, err)

	// Verify milestone is completed
	milestonesUpdated, err := shipSvc.GetMilestones(ctx, sh.ID)
	require.NoError(t, err)
	var departedMilestone *shipmentspec.ShipmentMilestone
	for _, m := range milestonesUpdated {
		if m.MilestoneCode == "DEPARTED" {
			departedMilestone = m
		}
	}
	require.NotNil(t, departedMilestone)
	assert.Equal(t, "COMPLETED", departedMilestone.Status)
	assert.Equal(t, location, *departedMilestone.Location)
	assert.Equal(t, notes, *departedMilestone.Notes)

	// Verify shipment status advanced to DEPARTED
	shUpdated, err := shipSvc.GetShipmentByID(ctx, 9991, sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "DEPARTED", shUpdated.Status)

	// 3. One-way status progression check (Completing lower-rank BOOKED should not regress status)
	err = shipSvc.UpdateMilestone(ctx, 9991, sh.ID, "BOOKED", &now, nil, nil)
	require.NoError(t, err)

	shRegressCheck, err := shipSvc.GetShipmentByID(ctx, 9991, sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "DEPARTED", shRegressCheck.Status) // Still DEPARTED!

	// 4. Tenant isolation: Org B tries to update Org A's milestone
	err = shipSvc.UpdateMilestone(ctx, 9992, sh.ID, "IN_TRANSIT", &now, nil, nil)
	assert.Error(t, err) // Should return error (access denied or shipment not found)
}

func TestShipmentExceptionsLifecycle(t *testing.T) {
	// Connect to local test DB
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up database tables
	_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_exceptions WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")

	// Seed Organizations
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, created_at, updated_at) 
		VALUES (9991, 'Org A', NOW(), NOW()), (9992, 'Org B', NOW(), NOW())
	`)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_exceptions WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM customers WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")
	}()

	// Seed Customers
	_, err = db.Exec(`
		INSERT INTO customers (id, org_id, name, contact_email)
		VALUES 
			(9301, 9991, 'Customer A', 'cust-a@example.com'),
			(9302, 9992, 'Customer B', 'cust-b@example.com')
	`)
	require.NoError(t, err)

	// Seed RFQs
	_, err = db.Exec(`
		INSERT INTO rfqs (id, org_id, rfq_number, customer_id, origin, destination, incoterms, status, stage)
		VALUES 
			(9001, 9991, 'RFQ-9001', 9301, 'INNSA', 'DEHAM', 'FOB', 'WON', 'COMPLETED')
	`)
	require.NoError(t, err)

	// Seed RFQ Quotes
	_, err = db.Exec(`
		INSERT INTO rfq_quotes (id, rfq_id, carrier_name, transit_time_days, sell_price, status)
		VALUES (9101, 9001, 'Maersk Line', 14, 2500.0, 'APPROVED')
	`)
	require.NoError(t, err)

	// Seed Bookings
	_, err = db.Exec(`
		INSERT INTO bookings (id, org_id, rfq_id, quote_id, booking_number, carrier_name, carrier_scac, status, origin_port, destination_port, vessel_name, voyage_number)
		VALUES (9201, 9991, 9001, 9101, 'BKG-9201', 'Maersk Line', 'MAEU', 'CONFIRMED', 'INNSA', 'DEHAM', 'Maersk Hamburg', '2601W')
	`)
	require.NoError(t, err)

	// Initialize DL and Service
	shipRepo := shipments.NewRepository(db)
	shipSvc := shipments.NewService(shipRepo, db, nil, "http://localhost:8080")
	ctx := context.Background()

	// 1. Create Shipment
	sh, err := shipSvc.CreateFromRFQ(ctx, 9001)
	require.NoError(t, err)
	require.NotNil(t, sh)

	// 2. Raise manual exception
	err = shipSvc.CreateShipmentException(ctx, 9991, sh.ID, "CUSTOMS_HOLD", "HIGH", "Customs Inspection Hold", "Cargo selected for custom hold", nil)
	require.NoError(t, err)

	// Verify exceptions raised
	exceptions, err := shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	require.Len(t, exceptions, 1)
	assert.Equal(t, "CUSTOMS_HOLD", exceptions[0].ExceptionType)
	assert.Equal(t, "OPEN", exceptions[0].Status)

	// 3. Exception status updates and lifecycle validation
	exceptionID := exceptions[0].ID

	// Try to update with invalid status
	err = shipSvc.UpdateShipmentException(ctx, 9991, sh.ID, exceptionID, "INVALID_STATUS", "", nil)
	assert.Error(t, err) // invalid status

	// Acknowledge exception
	err = shipSvc.AcknowledgeShipmentException(ctx, 9991, sh.ID, exceptionID)
	require.NoError(t, err)

	exceptions, err = shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "ACKNOWLEDGED", exceptions[0].Status)

	// Dismiss exception
	err = shipSvc.DismissShipmentException(ctx, 9991, sh.ID, exceptionID)
	require.NoError(t, err)

	exceptions, err = shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "DISMISSED", exceptions[0].Status)
	assert.True(t, exceptions[0].Resolved)

	// Cannot dismiss/resolve again
	err = shipSvc.ResolveShipmentException(ctx, 9991, sh.ID, exceptionID, "Resolution notes", 1)
	assert.Error(t, err)

	// 4. Deterministic engine delay validation
	// Let's create an overdue milestone
	// Set BOOKED planned_date to 48 hours ago
	pastDate := time.Now().Add(-48 * time.Hour)
	_, err = db.Exec("UPDATE shipment_milestones SET planned_date = ? WHERE shipment_id = ? AND milestone_code = 'BOOKED'", pastDate, sh.ID)
	require.NoError(t, err)

	// Run deterministic engine evaluation
	err = shipSvc.EvaluateShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)

	// Verify overdue exception was created
	exceptions, err = shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	// We should have 2 exceptions now: the dismissed customs hold + the newly evaluated overdue BOOKED milestone exception
	assert.Len(t, exceptions, 2)

	var overdueEx *shipmentspec.ShipmentException
	for _, ex := range exceptions {
		if ex.Status == "OPEN" {
			overdueEx = ex
		}
	}
	require.NotNil(t, overdueEx)
	assert.Equal(t, "SCHEDULE_DELAY", overdueEx.ExceptionType)
	assert.Equal(t, "HIGH", overdueEx.Severity)
	assert.Contains(t, *overdueEx.SourceEventID, "OVERDUE-BOOKED")

	// Idempotence check: running evaluate again should not duplicate exceptions
	err = shipSvc.EvaluateShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)

	exceptions, err = shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	assert.Len(t, exceptions, 2) // still 2 exceptions!

	// 5. Tenant isolation check: Org B tries to access Org A's exceptions
	_, err = shipSvc.GetShipmentExceptions(ctx, 9992, sh.ID)
	assert.Error(t, err) // Access denied or shipment not found

	err = shipSvc.AcknowledgeShipmentException(ctx, 9992, sh.ID, overdueEx.ID)
	assert.Error(t, err) // Access denied

	// 6. Invalid transitions (Dismissed Exception cannot be updated or resolved)
	err = shipSvc.UpdateShipmentException(ctx, 9991, sh.ID, exceptionID, "OPEN", "Reopen dismissed", nil)
	assert.Error(t, err)

	// 7. Multiple Exceptions Lifecycle: Resolve one while another remains open
	// Currently overdueEx (ID: overdueEx.ID) is OPEN. Let's raise another manual exception.
	err = shipSvc.CreateShipmentException(ctx, 9991, sh.ID, "PORT_CONGESTION", "CRITICAL", "Port Congestion Hold", "Vessel stuck outside port", nil)
	require.NoError(t, err)

	exceptions, err = shipSvc.GetShipmentExceptions(ctx, 9991, sh.ID)
	require.NoError(t, err)
	// We should now have 3 exceptions: 1 dismissed, 1 open (overdueEx), 1 open (port congestion)
	assert.Len(t, exceptions, 3)

	var portEx *shipmentspec.ShipmentException
	for _, ex := range exceptions {
		if ex.ExceptionType == "PORT_CONGESTION" {
			portEx = ex
		}
	}
	require.NotNil(t, portEx)

	// Since we have active exceptions, let's set the main shipment status to EXCEPTION (simulated)
	_, err = db.Exec("UPDATE shipments SET status = 'EXCEPTION' WHERE id = ?", sh.ID)
	require.NoError(t, err)

	// Resolve the overdue exception
	err = shipSvc.ResolveShipmentException(ctx, 9991, sh.ID, overdueEx.ID, "Resolved schedule delay", 1)
	require.NoError(t, err)

	// Verify overdueEx is resolved, but the shipment status is STILL EXCEPTION because portEx is still open!
	var updatedShipment shipmentspec.Shipment
	err = db.GetContext(ctx, &updatedShipment, "SELECT * FROM shipments WHERE id = ?", sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "EXCEPTION", updatedShipment.Status)

	// Resolve the final port congestion exception
	err = shipSvc.ResolveShipmentException(ctx, 9991, sh.ID, portEx.ID, "Resolved port congestion", 1)
	require.NoError(t, err)

	// Verify both exceptions are resolved and the shipment status has returned to the latest completed milestone (BOOKING_PENDING since no milestone completed yet)
	err = db.GetContext(ctx, &updatedShipment, "SELECT * FROM shipments WHERE id = ?", sh.ID)
	require.NoError(t, err)
	assert.Equal(t, "BOOKING_PENDING", updatedShipment.Status)

	// 8. Robustness under missing dates
	// Create another shipment with completely null milestone planned dates
	shNullDates, err := db.Exec(`
		INSERT INTO shipments (org_id, rfq_id, status, carrier_scac, origin_port, destination_port, created_at, updated_at)
		VALUES (9991, 9001, 'BOOKED', 'MAEU', 'INNSA', 'DEHAM', NOW(), NOW())
	`)
	require.NoError(t, err)
	shNullID, err := shNullDates.LastInsertId()
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, actual_date, status)
		VALUES (?, 'BOOKED', 'Confirmed', NULL, NULL, 'PLANNED')
	`, shNullID)
	require.NoError(t, err)

	// Run deterministic diagnostics engine on null dates shipment — should not panic or produce exceptions
	err = shipSvc.EvaluateShipmentExceptions(ctx, 9991, shNullID)
	require.NoError(t, err)

	nullExcs, err := shipSvc.GetShipmentExceptions(ctx, 9991, shNullID)
	require.NoError(t, err)
	assert.Len(t, nullExcs, 0)
}

func TestShipmentTrackingAndClosure(t *testing.T) {
	// Connect to local test DB
	dbURL := "root:@tcp(127.0.0.1:3306)/freel_mysql?parseTime=true&loc=UTC&multiStatements=true"
	db, err := database.Connect(dbURL)
	require.NoError(t, err)
	defer db.Close()

	// Clean up database tables
	_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_exceptions WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
	_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
	_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")

	// Seed Organization
	_, err = db.Exec(`
		INSERT INTO organizations (id, name, created_at, updated_at) 
		VALUES (9991, 'Org A', NOW(), NOW()), (9992, 'Org B', NOW(), NOW())
		ON DUPLICATE KEY UPDATE name = VALUES(name)
	`)
	require.NoError(t, err)

	defer func() {
		_, _ = db.Exec("DELETE FROM activities WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_exceptions WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id IN (SELECT id FROM shipments WHERE org_id IN (9991, 9992))")
		_, _ = db.Exec("DELETE FROM shipments WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM bookings WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM rfqs WHERE org_id IN (9991, 9992)")
		_, _ = db.Exec("DELETE FROM organizations WHERE id IN (9991, 9992)")
	}()

	// Seed Customer
	_, err = db.Exec(`
		INSERT INTO customers (id, org_id, name, contact_email)
		VALUES (9301, 9991, 'Customer A', 'cust-a@example.com')
	`)
	require.NoError(t, err)

	// Seed RFQ
	_, err = db.Exec(`
		INSERT INTO rfqs (id, org_id, rfq_number, customer_id, origin, destination, incoterms, status, stage)
		VALUES (9001, 9991, 'RFQ-9001', 9301, 'INNSA', 'DEHAM', 'FOB', 'WON', 'COMPLETED')
	`)
	require.NoError(t, err)

	// Create Shipment under Org A (9991)
	shRes, err := db.Exec(`
		INSERT INTO shipments (org_id, rfq_id, status, carrier_scac, origin_port, destination_port, created_at, updated_at, closure_status)
		VALUES (9991, 9001, 'BOOKING_PENDING', 'MAEU', 'INNSA', 'DEHAM', NOW(), NOW(), 'ACTIVE')
	`)
	require.NoError(t, err)
	shID, err := shRes.LastInsertId()
	require.NoError(t, err)

	// Setup service
	shipRepo := shipments.NewRepository(db)
	shipSvc := shipments.NewService(shipRepo, db, nil, "http://localhost:8080")
	ctx := context.Background()

	t.Run("Default progress calculation is correct", func(t *testing.T) {
		summary, err := shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, 10, summary.ProgressPercentage)
		assert.Equal(t, "ON_TRACK", summary.TrackingState)
		assert.Equal(t, "ACTIVE", summary.ClosureStatus)
	})

	t.Run("Progress calculation advances when milestones are completed", func(t *testing.T) {
		// Complete BOOKED milestone
		_, err = db.Exec(`
			INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, actual_date, status)
			VALUES (?, 'BOOKED', 'Confirmed', NOW(), NOW(), 'COMPLETED')
		`, shID)
		require.NoError(t, err)

		summary, err := shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, 30, summary.ProgressPercentage)
		assert.Equal(t, "BOOKED", summary.HighestCompletedMilestone)
	})

	t.Run("Schedule health delay check evaluates correctly", func(t *testing.T) {
		// Add DEPARTED milestone with planned date 2 days ago and still PLANNED
		pastDate := time.Now().Add(-48 * time.Hour)
		_, err = db.Exec(`
			INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, actual_date, status)
			VALUES (?, 'DEPARTED', 'Departed', ?, NULL, 'PLANNED')
		`, shID, pastDate)
		require.NoError(t, err)

		summary, err := shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "DELAYED", summary.TrackingState)
	})

	t.Run("Exception overrides schedule health", func(t *testing.T) {
		// Insert active high exception
		_, err = db.Exec(`
			INSERT INTO shipment_exceptions (org_id, shipment_id, exception_type, severity, status, title, description, resolved, source_event_id, created_at, updated_at)
			VALUES (9991, ?, 'CUSTOMS_HOLD', 'HIGH', 'OPEN', 'Customs Hold', 'Held at port', 0, 'EXC-1', NOW(), NOW())
		`, shID)
		require.NoError(t, err)

		summary, err := shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "EXCEPTION", summary.TrackingState)
	})

	t.Run("Closure validations block closure when required milestones are not completed", func(t *testing.T) {
		// Request closure should fail since other milestones (DEPARTED, etc) are not completed
		err = shipSvc.RequestClosure(ctx, 9991, shID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot request closure")

		eval, err := shipSvc.EvaluateClosure(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "ACTIVE", eval)
	})

	t.Run("Tenant security: cross-tenant requests must fail", func(t *testing.T) {
		// Org B (9992) tries to fetch tracking or request closure of Org A shipment
		_, err := shipSvc.GetShipmentTracking(ctx, 9992, shID)
		assert.Error(t, err)

		err = shipSvc.RequestClosure(ctx, 9992, shID)
		assert.Error(t, err)

		err = shipSvc.CompleteShipment(ctx, 9992, shID)
		assert.Error(t, err)

		err = shipSvc.ReopenShipment(ctx, 9992, shID)
		assert.Error(t, err)
	})

	t.Run("E2E Closure state machine transitions successfully", func(t *testing.T) {
		// Complete all remaining milestones (DEPARTED, IN_TRANSIT, ARRIVED, DELIVERED)
		_, _ = db.Exec("DELETE FROM shipment_milestones WHERE shipment_id = ?", shID)
		milestonesList := []string{"BOOKED", "DEPARTED", "IN_TRANSIT", "ARRIVED", "DELIVERED"}
		for _, mCode := range milestonesList {
			_, err = db.Exec(`
				INSERT INTO shipment_milestones (shipment_id, milestone_code, description, planned_date, actual_date, status)
				VALUES (?, ?, 'Milestone completed', NOW(), NOW(), 'COMPLETED')
			`, shID, mCode)
			require.NoError(t, err)
		}

		// Resolve open exceptions to ensure no blockage
		_, err = db.Exec("UPDATE shipment_exceptions SET resolved = 1, status = 'RESOLVED' WHERE shipment_id = ?", shID)
		require.NoError(t, err)

		// Evaluate closure -> should transition to READY_FOR_CLOSURE
		evalStatus, err := shipSvc.EvaluateClosure(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "READY_FOR_CLOSURE", evalStatus)

		// Check tracking summary shows READY_FOR_CLOSURE
		summary, err := shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "READY_FOR_CLOSURE", summary.ClosureStatus)
		assert.Equal(t, 100, summary.ProgressPercentage)

		// Complete Shipment E2E -> closure_status = CLOSED
		err = shipSvc.CompleteShipment(ctx, 9991, shID)
		require.NoError(t, err)

		// Verify closed status and DELIVERED shipment status
		summary, err = shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "CLOSED", summary.ClosureStatus)
		assert.Equal(t, "DELIVERED", summary.ShipmentStatus)

		// Reopen Shipment E2E -> closure_status = ACTIVE
		err = shipSvc.ReopenShipment(ctx, 9991, shID)
		require.NoError(t, err)

		summary, err = shipSvc.GetShipmentTracking(ctx, 9991, shID)
		require.NoError(t, err)
		assert.Equal(t, "ACTIVE", summary.ClosureStatus)
	})
}


