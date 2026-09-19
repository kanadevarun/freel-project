package bcontext

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freel/backend/internal/rbac"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRBACService struct{}

func (m *mockRBACService) HasPermission(ctx context.Context, roleName string, orgID int64, resource, action string) (bool, error) {
	return true, nil
}
func (m *mockRBACService) SeedSystemPermissions(ctx context.Context) error {
	return nil
}
func (m *mockRBACService) SeedDefaultRolesForOrg(ctx context.Context, orgID int64) error {
	return nil
}
func (m *mockRBACService) GetRoles(ctx context.Context, orgID int64) ([]rbac.RoleResponse, error) {
	return nil, nil
}
func (m *mockRBACService) GetRolePermissions(ctx context.Context, orgID, roleID int64) (*rbac.RolePermissionsResponse, error) {
	return nil, nil
}
func (m *mockRBACService) UpdateRolePermissions(ctx context.Context, orgID, roleID int64, req rbac.UpdatePermissionsRequest) error {
	return nil
}
func (m *mockRBACService) GetStats(ctx context.Context, orgID int64) (*rbac.StatsResponse, error) {
	return nil, nil
}
func (m *mockRBACService) CreateRole(ctx context.Context, orgID int64, req rbac.CreateRoleRequest) (*rbac.RoleResponse, error) {
	return nil, nil
}
func (m *mockRBACService) UpdateRole(ctx context.Context, orgID, roleID int64, req rbac.UpdateRoleRequest) error {
	return nil
}
func (m *mockRBACService) DeleteRole(ctx context.Context, orgID, roleID int64) error {
	return nil
}

func setupTestService(t *testing.T) (Service, sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(mockDB, "sqlmock")
	svc := NewService(sqlxDB, &mockRBACService{})
	return svc, mock
}

func TestGetBusinessContext_OrganizationIsolation(t *testing.T) {
	svc, mock := setupTestService(t)
	ctx := context.Background()

	// Org 1 tries to access Customer 101 which belongs to Org 2 -> sql.ErrNoRows
	mock.ExpectQuery(`SELECT id, name, customer_code, status, credit_status, customer_type, currency, payment_terms, credit_limit, health_score, country, city, created_at, updated_at FROM customers WHERE id = \? AND org_id = \?`).
		WithArgs(int64(101), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // empty rows simulates foreign tenant isolation

	_, err := svc.GetBusinessContext(ctx, ContextRequest{
		OrgID:       1,
		UserID:      1,
		PrimaryType: EntityTypeCustomer,
		PrimaryID:   101,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "customer 101 not found in organization")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBusinessContext_InvalidInputs(t *testing.T) {
	svc, _ := setupTestService(t)
	ctx := context.Background()

	// 1. Invalid OrgID
	_, err := svc.GetBusinessContext(ctx, ContextRequest{
		OrgID:       0,
		PrimaryType: EntityTypeCustomer,
		PrimaryID:   1,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid organization ID")

	// 2. Invalid PrimaryID
	_, err = svc.GetBusinessContext(ctx, ContextRequest{
		OrgID:       1,
		PrimaryType: EntityTypeCustomer,
		PrimaryID:   -5,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid primary record ID")

	// 3. Unsupported entity type
	_, err = svc.GetBusinessContext(ctx, ContextRequest{
		OrgID:       1,
		PrimaryType: EntityType("UNKNOWN_MODULE"),
		PrimaryID:   1,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported entity type")
}

func TestGetBusinessContext_RFQSuccess(t *testing.T) {
	svc, mock := setupTestService(t)
	ctx := context.Background()
	now := time.Now().UTC()
	rfqNum := "RFQ-2026-1001"
	origin := "INNSA"
	dest := "DEHAM"
	incoterms := "FOB"
	stage := "ACTIVE"

	// Primary RFQ Query
	mock.ExpectQuery(`SELECT id, rfq_number, customer_id, lead_id, status, stage, agent_status, origin, destination, incoterms, health_score, target_date, created_at, updated_at FROM rfqs WHERE id = \? AND org_id = \?`).
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "rfq_number", "customer_id", "lead_id", "status", "stage", "agent_status", "origin", "destination", "incoterms", "health_score", "target_date", "created_at", "updated_at",
		}).AddRow(1, &rfqNum, 101, nil, "CONFIRMED", &stage, "IDLE", &origin, &dest, &incoterms, 85, &now, &now, &now))

	// Related Customer Query
	mock.ExpectQuery(`SELECT name FROM customers WHERE id = \? AND org_id = \?`).
		WithArgs(int64(101), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Apex Global Logistics Corp"))

	// Related Quotations Query
	mock.ExpectQuery(`SELECT id, quotation_number, status, total_amount, currency, created_at FROM quotations WHERE rfq_id = \? AND org_id = \?`).
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_number", "status", "total_amount", "currency", "created_at",
		}).AddRow(201, "QT-2026-001", "ACCEPTED", 3450.00, "USD", &now))

	bCtx, err := svc.GetBusinessContext(ctx, ContextRequest{
		OrgID:       1,
		UserID:      1,
		PrimaryType: EntityTypeRFQ,
		PrimaryID:   1,
	})

	require.NoError(t, err)
	assert.Equal(t, int64(1), bCtx.OrgID)
	assert.Equal(t, EntityTypeRFQ, bCtx.PrimaryType)
	assert.Equal(t, "RFQ-2026-1001", bCtx.PrimaryRecord.ReferenceNumber)
	assert.Equal(t, "INNSA", bCtx.PrimaryRecord.KeyAttributes["origin"])
	assert.Equal(t, "DEHAM", bCtx.PrimaryRecord.KeyAttributes["destination"])
	assert.True(t, bCtx.IsReadOnly)

	// Check related customer
	custs := bCtx.RelatedRecords[EntityTypeCustomer]
	require.Len(t, custs, 1)
	assert.Equal(t, "Apex Global Logistics Corp", custs[0].Title)

	// Check related quotations
	quotes := bCtx.RelatedRecords[EntityTypeQuotation]
	require.Len(t, quotes, 1)
	assert.Equal(t, "QT-2026-001", quotes[0].ReferenceNumber)
	assert.Equal(t, 3450.00, quotes[0].FinancialValues["total_amount"])

	// Check Source References
	assert.GreaterOrEqual(t, len(bCtx.SourceReferences), 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGenerateInsight_GroundedRFQ(t *testing.T) {
	svc, mock := setupTestService(t)
	ctx := context.Background()
	now := time.Now().UTC()
	rfqNum := "RFQ-2026-1001"
	origin := "INNSA"
	dest := "DEHAM"
	incoterms := "FOB"
	stage := "ACTIVE"

	// Mock queries for RFQ context
	mock.ExpectQuery(`SELECT id, rfq_number, customer_id, lead_id, status, stage, agent_status, origin, destination, incoterms, health_score, target_date, created_at, updated_at FROM rfqs WHERE id = \? AND org_id = \?`).
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "rfq_number", "customer_id", "lead_id", "status", "stage", "agent_status", "origin", "destination", "incoterms", "health_score", "target_date", "created_at", "updated_at",
		}).AddRow(1, &rfqNum, 101, nil, "CONFIRMED", &stage, "IDLE", &origin, &dest, &incoterms, 85, &now, &now, &now))

	mock.ExpectQuery(`SELECT name FROM customers WHERE id = \? AND org_id = \?`).
		WithArgs(int64(101), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("Apex Global Logistics Corp"))

	mock.ExpectQuery(`SELECT id, quotation_number, status, total_amount, currency, created_at FROM quotations WHERE rfq_id = \? AND org_id = \?`).
		WithArgs(int64(1), int64(1)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "quotation_number", "status", "total_amount", "currency", "created_at",
		}).AddRow(201, "QT-2026-001", "ACCEPTED", 3450.00, "USD", &now))

	insight, err := svc.GenerateInsight(ctx, InsightRequest{
		OrgID:      1,
		UserID:     1,
		EntityType: EntityTypeRFQ,
		EntityID:   1,
	})

	require.NoError(t, err)
	assert.True(t, insight.IsInformationalOnly)
	assert.Equal(t, "HIGH", insight.ConfidenceLevel)
	assert.Contains(t, insight.Title, "RFQ Intelligence: RFQ-2026-1001")
	assert.Contains(t, insight.Summary, "CONFIRMED")
	assert.GreaterOrEqual(t, len(insight.SupportingRecords), 2)
	assert.GreaterOrEqual(t, len(insight.SupportingFieldReferences), 3)
	assert.NoError(t, mock.ExpectationsWereMet())
}
