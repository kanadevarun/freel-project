package subscription

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/freel/backend/internal/middleware"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

// GetWorkspace returns the current subscription, usage, add-ons, invoices, etc.
func (h *Handler) GetWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	workspace, err := h.svc.GetWorkspace(ctx, userCtx.OrgID)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch workspace data"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workspace)
}

// GetPlans returns all available plans to subscribe to.
func (h *Handler) GetPlans(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Still protected by auth, but technically org-independent
	plans, err := h.svc.GetAvailablePlans(ctx)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch plans"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(plans)
}

// ChangePlan upgrades or downgrades the subscription.
func (h *Handler) ChangePlan(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.BillingCycle != "monthly" && req.BillingCycle != "annual" {
		http.Error(w, `{"error": "Invalid billing cycle"}`, http.StatusBadRequest)
		return
	}

	err := h.svc.ChangePlan(ctx, userCtx.OrgID, req)
	if err == ErrPlanNotFound {
		http.Error(w, `{"error": "Plan not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to change plan"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// CancelSubscription marks the subscription to cancel at the end of the period.
func (h *Handler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	err := h.svc.CancelSubscription(ctx, userCtx.OrgID)
	if err != nil {
		http.Error(w, `{"error": "Failed to cancel subscription"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// Checkout creates a new Stripe checkout session
func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	if req.BillingCycle != "monthly" && req.BillingCycle != "annual" {
		http.Error(w, `{"error": "Invalid billing cycle"}`, http.StatusBadRequest)
		return
	}

	sessionURL, err := h.svc.CreateCheckoutSession(ctx, userCtx.OrgID, req)
	if err == ErrPlanNotFound {
		http.Error(w, `{"error": "Plan not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to create checkout session: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": sessionURL})
}

// ReactivateSubscription removes the cancel_at_period_end flag to resume an active plan.
func (h *Handler) ReactivateSubscription(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	err := h.svc.ReactivateSubscription(ctx, userCtx.OrgID)
	if err != nil {
		http.Error(w, `{"error": "Failed to reactivate subscription: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}

// CreateCustomerPortal generates a Stripe Customer Portal link.
func (h *Handler) CreateCustomerPortal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	url, err := h.svc.CreateCustomerPortalSession(ctx, userCtx.OrgID)
	if err != nil {
		http.Error(w, `{"error": "Failed to create portal session: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

// Webhook handles Stripe webhook events
func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Read body (limit to 64KB for security)
	r.Body = http.MaxBytesReader(w, r.Body, 65536)
	
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
		return
	}

	sigHeader := r.Header.Get("Stripe-Signature")
	
	err = h.svc.ProcessStripeWebhook(ctx, payload, sigHeader)
	if err != nil {
		// Log the error but return 400 for Stripe to know it failed validation/processing
		http.Error(w, "Webhook processing failed", http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// PreviewPlanChange previews a subscription upgrade/downgrade.
func (h *Handler) PreviewPlanChange(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	preview, err := h.svc.PreviewPlanChange(ctx, userCtx.OrgID, req)
	if err == ErrPlanNotFound {
		http.Error(w, `{"error": "Plan not found"}`, http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, `{"error": "Failed to preview plan change"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preview)
}

// GetAddonConfigs returns all available add-ons.
func (h *Handler) GetAddonConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	addons, err := h.svc.GetAddonConfigs(ctx)
	if err != nil {
		http.Error(w, `{"error": "Failed to fetch addons"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(addons)
}

// UpdateAddons updates the organization's addons.
func (h *Handler) UpdateAddons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req UpdateAddonsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	err := h.svc.UpdateAddons(ctx, userCtx.OrgID, req)
	if err != nil {
		http.Error(w, `{"error": "Failed to update addons: `+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"success": true}`))
}
