package transport_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/audit/transport"
	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
)

type mockAuditService struct {
	recordedLogs []domain.AuditLog
}

func (m *mockAuditService) Record(ctx context.Context, params domain.CreateAuditLogParams) (*domain.AuditLog, error) {
	log := domain.AuditLog{
		ID:           1,
		OrgID:        params.OrgID,
		Action:       params.Action,
		Module:       params.Module,
		ResourceType: params.ResourceType,
		ResourceID:   params.ResourceID,
		Description:  params.Description,
		Result:       domain.ResultSuccess,
		CreatedAt:    time.Now().UTC(),
	}
	m.recordedLogs = append(m.recordedLogs, log)
	return &log, nil
}

func (m *mockAuditService) RecordAsync(ctx context.Context, params domain.CreateAuditLogParams) {
	_, _ = m.Record(ctx, params)
}

func (m *mockAuditService) List(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogListResponse, error) {
	return &domain.AuditLogListResponse{
		Items: []domain.AuditLog{
			{
				ID:           101,
				OrgID:        filter.OrgID,
				ActorType:    "USER",
				ActorName:    "Varun",
				ActorRole:    "Admin",
				Action:       "CREATE",
				Module:       "SHIPMENTS",
				ResourceType: "SHIPMENT",
				ResourceID:   "SHP-1024",
				Description:  "Shipment SHP-1024 created",
				Result:       "SUCCESS",
				CreatedAt:    time.Now().UTC(),
			},
		},
		Total:      1,
		Page:       1,
		Limit:      20,
		TotalPages: 1,
	}, nil
}

func (m *mockAuditService) GetByID(ctx context.Context, orgID int64, id int64) (*domain.AuditLog, error) {
	return &domain.AuditLog{
		ID:           id,
		OrgID:        orgID,
		ActorType:    "USER",
		ActorName:    "Varun",
		ActorRole:    "Admin",
		Action:       "UPDATE",
		Module:       "SHIPMENTS",
		ResourceType: "SHIPMENT",
		ResourceID:   "SHP-1024",
		Description:  "Shipment updated",
		Result:       "SUCCESS",
		CreatedAt:    time.Now().UTC(),
	}, nil
}

func TestHTTPListAuditLogs(t *testing.T) {
	mockSvc := &mockAuditService{}
	handler := transport.NewHandler(mockSvc)

	r := chi.NewRouter()
	r.Get("/api/v1/settings/audit-logs", handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/audit-logs?page=1&limit=20", nil)
	// Inject user context
	userCtx := middleware.UserContext{
		UserID: 1,
		OrgID:  10,
		Role:   "ADMIN",
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, userCtx))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp domain.AuditLogListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Total != 1 || len(resp.Items) != 1 {
		t.Errorf("expected 1 item, got total=%d items=%d", resp.Total, len(resp.Items))
	}
	if resp.Items[0].OrgID != 10 {
		t.Errorf("expected org_id 10, got %d", resp.Items[0].OrgID)
	}
}

func TestHTTPGetAuditLogByID(t *testing.T) {
	mockSvc := &mockAuditService{}
	handler := transport.NewHandler(mockSvc)

	r := chi.NewRouter()
	r.Get("/api/v1/settings/audit-logs/{id}", handler.GetAuditLogByID)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/audit-logs/55", nil)
	userCtx := middleware.UserContext{
		UserID: 1,
		OrgID:  10,
		Role:   "ADMIN",
	}
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, userCtx))

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var entry domain.AuditLog
	if err := json.Unmarshal(rec.Body.Bytes(), &entry); err != nil {
		t.Fatalf("failed to decode entry: %v", err)
	}

	if entry.ID != 55 {
		t.Errorf("expected ID 55, got %d", entry.ID)
	}
	if entry.OrgID != 10 {
		t.Errorf("expected OrgID 10, got %d", entry.OrgID)
	}
}

func TestHTTPUnauthorizedWhenNoUserContext(t *testing.T) {
	mockSvc := &mockAuditService{}
	handler := transport.NewHandler(mockSvc)

	r := chi.NewRouter()
	r.Get("/api/v1/settings/audit-logs", handler.ListAuditLogs)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/settings/audit-logs", nil)
	// No user context
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", rec.Code)
	}
}
