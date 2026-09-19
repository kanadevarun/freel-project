package recommendations

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/freel/backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerSecurityAndWorkflows(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRepository(db)
	gen := NewGenerator(db, repo)
	svc := NewService(repo, gen, nil)
	handler := NewHandler(svc)

	orgID := int64(8899)
	userCtx := middleware.UserContext{
		UserID: 42,
		OrgID:  orgID,
		Role:   "ADMIN",
	}

	// Clean test data
	ctx := context.Background()
	_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", orgID)
	defer func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM ai_recommendations WHERE org_id = ?", orgID)
	}()

	// 1. Unauthenticated request must return 401
	t.Run("Unauthenticated_Returns401", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/recommendations", nil)
		w := httptest.NewRecorder()
		handler.List(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		reqGen := httptest.NewRequest("POST", "/api/recommendations/generate", nil)
		wGen := httptest.NewRecorder()
		handler.Generate(wGen, reqGen)
		assert.Equal(t, http.StatusUnauthorized, wGen.Code)
	})

	// 2. Generate recommendations for org
	t.Run("Generate_Authenticated", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/recommendations/generate", nil)
		req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, userCtx))
		w := httptest.NewRecorder()

		handler.Generate(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp["success"].(bool))
	})

	// 3. Create explicit recommendation and test workflow endpoints
	t.Run("CRUD_Workflow_Endpoints", func(t *testing.T) {
		hash := computeDedupHash(orgID, SourceInvoice, 999, CategoryFinance, "HANDLER_TEST_RULE")
		cand := &Recommendation{
			OrgID:             orgID,
			SourceType:        SourceInvoice,
			SourceID:          999,
			SourceReference:   "INV-TEST-999",
			Title:             "Handler Test Overdue Invoice",
			Description:       "Test description for handler verification",
			Category:          CategoryFinance,
			Priority:          PriorityHigh,
			RiskLevel:         RiskHigh,
			Confidence:        "HIGH",
			ConfidenceScore:   0.94,
			EvidenceJSON:      `[{"source_module":"finance","source_entity_id":999,"source_ref":"INV-TEST-999","field_name":"amount","observed_value":12500}]`,
			RecommendedAction: "Review aging invoice and dispatch reminder",
			ActionType:        "REVIEW_INVOICE",
			Status:            StatusNew,
			RequiresApproval:  false,
			CorrelationID:     "hndlr-corr-1",
			CreatedBy:         "HANDLER_TEST",
			GeneratedBy:       "DETERMINISTIC_RULES",
			RuleApplied:       "HANDLER_TEST_RULE",
			DedupHash:         hash,
		}

		created, err := repo.Create(ctx, cand)
		require.NoError(t, err)
		require.NotNil(t, created)

		// A. List with pagination
		listReq := httptest.NewRequest("GET", "/api/recommendations?page=1&limit=10&category=finance", nil)
		listReq = listReq.WithContext(context.WithValue(listReq.Context(), middleware.UserContextKey, userCtx))
		listW := httptest.NewRecorder()
		handler.List(listW, listReq)
		assert.Equal(t, http.StatusOK, listW.Code)

		var listResp map[string]interface{}
		_ = json.Unmarshal(listW.Body.Bytes(), &listResp)
		assert.True(t, listResp["success"].(bool))
		recs := listResp["recommendations"].([]interface{})
		assert.GreaterOrEqual(t, len(recs), 1)

		// B. Get Stats
		statsReq := httptest.NewRequest("GET", "/api/recommendations/stats", nil)
		statsReq = statsReq.WithContext(context.WithValue(statsReq.Context(), middleware.UserContextKey, userCtx))
		statsW := httptest.NewRecorder()
		handler.GetStats(statsW, statsReq)
		assert.Equal(t, http.StatusOK, statsW.Code)

		// C. GetByID
		rContext := chi.NewRouteContext()
		rContext.URLParams.Add("id", "99999999") // non-existent
		badGetReq := httptest.NewRequest("GET", "/api/recommendations/99999999", nil)
		badGetReq = badGetReq.WithContext(context.WithValue(badGetReq.Context(), chi.RouteCtxKey, rContext))
		badGetReq = badGetReq.WithContext(context.WithValue(badGetReq.Context(), middleware.UserContextKey, userCtx))
		badGetW := httptest.NewRecorder()
		handler.GetByID(badGetW, badGetReq)
		assert.Equal(t, http.StatusNotFound, badGetW.Code)

		// Valid GetByID
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", string(rune('0'+created.ID)))
		getReq := httptest.NewRequest("GET", "/api/recommendations/"+jsonNumber(created.ID), nil)
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		getReq = getReq.WithContext(context.WithValue(getReq.Context(), chi.RouteCtxKey, rContext))
		getReq = getReq.WithContext(context.WithValue(getReq.Context(), middleware.UserContextKey, userCtx))
		getW := httptest.NewRecorder()
		handler.GetByID(getW, getReq)
		assert.Equal(t, http.StatusOK, getW.Code)

		// D. Review action
		reviewReq := httptest.NewRequest("POST", "/api/recommendations/"+jsonNumber(created.ID)+"/review", nil)
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		reviewReq = reviewReq.WithContext(context.WithValue(reviewReq.Context(), chi.RouteCtxKey, rContext))
		reviewReq = reviewReq.WithContext(context.WithValue(reviewReq.Context(), middleware.UserContextKey, userCtx))
		reviewW := httptest.NewRecorder()
		handler.MarkReviewed(reviewW, reviewReq)
		assert.Equal(t, http.StatusOK, reviewW.Code)

		// E. Assign action
		assignBody, _ := json.Marshal(AssignInput{AssigneeID: 55, AssigneeName: "Finance Lead"})
		assignReq := httptest.NewRequest("POST", "/api/recommendations/"+jsonNumber(created.ID)+"/assign", bytes.NewReader(assignBody))
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		assignReq = assignReq.WithContext(context.WithValue(assignReq.Context(), chi.RouteCtxKey, rContext))
		assignReq = assignReq.WithContext(context.WithValue(assignReq.Context(), middleware.UserContextKey, userCtx))
		assignW := httptest.NewRecorder()
		handler.Assign(assignW, assignReq)
		assert.Equal(t, http.StatusOK, assignW.Code)

		// F. Dismiss action (missing reason -> 400)
		badDismissBody, _ := json.Marshal(DismissInput{Reason: ""})
		badDismissReq := httptest.NewRequest("POST", "/api/recommendations/"+jsonNumber(created.ID)+"/dismiss", bytes.NewReader(badDismissBody))
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		badDismissReq = badDismissReq.WithContext(context.WithValue(badDismissReq.Context(), chi.RouteCtxKey, rContext))
		badDismissReq = badDismissReq.WithContext(context.WithValue(badDismissReq.Context(), middleware.UserContextKey, userCtx))
		badDismissW := httptest.NewRecorder()
		handler.Dismiss(badDismissW, badDismissReq)
		assert.Equal(t, http.StatusBadRequest, badDismissW.Code)

		// Valid dismiss
		dismissBody, _ := json.Marshal(DismissInput{Reason: "Payment received today"})
		dismissReq := httptest.NewRequest("POST", "/api/recommendations/"+jsonNumber(created.ID)+"/dismiss", bytes.NewReader(dismissBody))
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		dismissReq = dismissReq.WithContext(context.WithValue(dismissReq.Context(), chi.RouteCtxKey, rContext))
		dismissReq = dismissReq.WithContext(context.WithValue(dismissReq.Context(), middleware.UserContextKey, userCtx))
		dismissW := httptest.NewRecorder()
		handler.Dismiss(dismissW, dismissReq)
		assert.Equal(t, http.StatusOK, dismissW.Code)

		// G. Evidence retrieval
		evReq := httptest.NewRequest("GET", "/api/recommendations/"+jsonNumber(created.ID)+"/evidence", nil)
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("id", jsonNumber(created.ID))
		evReq = evReq.WithContext(context.WithValue(evReq.Context(), chi.RouteCtxKey, rContext))
		evReq = evReq.WithContext(context.WithValue(evReq.Context(), middleware.UserContextKey, userCtx))
		evW := httptest.NewRecorder()
		handler.GetEvidence(evW, evReq)
		assert.Equal(t, http.StatusOK, evW.Code)

		var evResp map[string]interface{}
		_ = json.Unmarshal(evW.Body.Bytes(), &evResp)
		assert.True(t, evResp["success"].(bool))
		evItems := evResp["evidence"].([]interface{})
		assert.Equal(t, 1, len(evItems))

		// H. List by source
		sourceReq := httptest.NewRequest("GET", "/api/recommendations/source/INVOICE/999", nil)
		rContext = chi.NewRouteContext()
		rContext.URLParams.Add("sourceType", "INVOICE")
		rContext.URLParams.Add("sourceId", "999")
		sourceReq = sourceReq.WithContext(context.WithValue(sourceReq.Context(), chi.RouteCtxKey, rContext))
		sourceReq = sourceReq.WithContext(context.WithValue(sourceReq.Context(), middleware.UserContextKey, userCtx))
		sourceW := httptest.NewRecorder()
		handler.ListBySource(sourceW, sourceReq)
		assert.Equal(t, http.StatusOK, sourceW.Code)
	})
}

func jsonNumber(n int64) string {
	bytes, _ := json.Marshal(n)
	return string(bytes)
}
