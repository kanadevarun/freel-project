package customers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/utils"
	"github.com/go-chi/chi/v5"
	httptransport "github.com/go-kit/kit/transport/http"
)

var (
	ErrUnauthorized = errors.New("unauthorized access")
	ErrInvalidInput = errors.New("invalid input data")
)

func AddCustomerHandlers(r chi.Router, endpoints Endpoints, authMiddleware func(http.Handler) http.Handler, rbacGuard *middleware.RBACMiddleware) {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Route("/customers", func(r chi.Router) {
		r.Use(authMiddleware)

		// List Customers
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/", httptransport.NewServer(
				endpoints.ListCustomersEndpoint,
				decodeListCustomersReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Get Customer KPIs
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/kpis", httptransport.NewServer(
				endpoints.GetKPIsEndpoint,
				decodeOrgReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Duplicate Check
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Post("/check-duplicate", httptransport.NewServer(
				endpoints.CheckDuplicateEndpoint,
				decodeCheckDuplicateReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Create Customer
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)).
			Post("/", httptransport.NewServer(
				endpoints.CreateCustomerEndpoint,
				decodeCreateCustomerReq,
				encodeCreatedResponse,
				opts...,
			).ServeHTTP)

		// Convert Lead to Customer
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)).
			Post("/convert-lead", httptransport.NewServer(
				endpoints.ConvertLeadEndpoint,
				decodeConvertLeadReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Customer Details
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}", httptransport.NewServer(
				endpoints.GetCustomerEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Customer 360° Dashboard & Cross-Module Intelligence (Task 22)
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/dashboard", httptransport.NewServer(
				endpoints.Get360DashboardEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/rfqs", httptransport.NewServer(
				endpoints.GetCustomerRFQsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/quotations", httptransport.NewServer(
				endpoints.GetCustomerQuotationsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/bookings", httptransport.NewServer(
				endpoints.GetCustomerBookingsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/shipments", httptransport.NewServer(
				endpoints.GetCustomerShipmentsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/contracts", httptransport.NewServer(
				endpoints.GetCustomerContractsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/activity", httptransport.NewServer(
				endpoints.GetCustomerTimelineEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Financial & Relationship Management (Task 23)
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/financial-profile", httptransport.NewServer(
				endpoints.GetFinancialProfileEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Put("/{id:[0-9]+}/financial-profile", httptransport.NewServer(
				endpoints.UpdateFinancialProfileEndpoint,
				decodeUpdateFinancialProfileReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/commercial-metrics", httptransport.NewServer(
				endpoints.GetCommercialMetricsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/account-ownership", httptransport.NewServer(
				endpoints.GetAccountOwnershipEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Put("/{id:[0-9]+}/account-ownership", httptransport.NewServer(
				endpoints.UpdateAccountOwnershipEndpoint,
				decodeUpdateOwnershipReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/ownership-history", httptransport.NewServer(
				endpoints.GetOwnershipHistoryEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/relationship-summary", httptransport.NewServer(
				endpoints.GetRelationshipSummaryEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Customer Intelligence & Risk Engine (Task 24)
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/intelligence/summary", httptransport.NewServer(
				endpoints.GetIntelligenceSummaryEndpoint,
				decodeOrgReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/intelligence/attention", httptransport.NewServer(
				endpoints.GetAttentionItemsEndpoint,
				decodeOrgReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Post("/{id:[0-9]+}/intelligence/evaluate", httptransport.NewServer(
				endpoints.EvaluateCustomerIntelligenceEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/risks", httptransport.NewServer(
				endpoints.GetCustomerRisksEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/opportunities", httptransport.NewServer(
				endpoints.GetCustomerOpportunitiesEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Post("/{id:[0-9]+}/risks/{risk_id:[0-9]+}/resolve", httptransport.NewServer(
				endpoints.ResolveCustomerRiskEndpoint,
				decodeResolveRiskReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Update Customer
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Put("/{id:[0-9]+}", httptransport.NewServer(
				endpoints.UpdateCustomerEndpoint,
				decodeUpdateCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Archive Customer
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionDelete)).
			Post("/{id:[0-9]+}/archive", httptransport.NewServer(
				endpoints.ArchiveCustomerEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Reactivate Customer
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Post("/{id:[0-9]+}/reactivate", httptransport.NewServer(
				endpoints.ReactivateCustomerEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Contacts Sub-resource
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/contacts", httptransport.NewServer(
				endpoints.ListContactsEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)).
			Post("/{id:[0-9]+}/contacts", httptransport.NewServer(
				endpoints.AddContactEndpoint,
				decodeAddContactReq,
				encodeCreatedResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Put("/{id:[0-9]+}/contacts/{contact_id:[0-9]+}", httptransport.NewServer(
				endpoints.UpdateContactEndpoint,
				decodeUpdateContactReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionDelete)).
			Delete("/{id:[0-9]+}/contacts/{contact_id:[0-9]+}", httptransport.NewServer(
				endpoints.DeleteContactEndpoint,
				decodeDeleteContactReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		// Addresses Sub-resource
		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionRead)).
			Get("/{id:[0-9]+}/addresses", httptransport.NewServer(
				endpoints.ListAddressesEndpoint,
				decodeCustomerReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionCreate)).
			Post("/{id:[0-9]+}/addresses", httptransport.NewServer(
				endpoints.AddAddressEndpoint,
				decodeAddAddressReq,
				encodeCreatedResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionUpdate)).
			Put("/{id:[0-9]+}/addresses/{address_id:[0-9]+}", httptransport.NewServer(
				endpoints.UpdateAddressEndpoint,
				decodeUpdateAddressReq,
				encodeResponse,
				opts...,
			).ServeHTTP)

		r.With(rbacGuard.RequirePermission(rbac.ResourceCompanies, rbac.ActionDelete)).
			Delete("/{id:[0-9]+}/addresses/{address_id:[0-9]+}", httptransport.NewServer(
				endpoints.DeleteAddressEndpoint,
				decodeDeleteAddressReq,
				encodeResponse,
				opts...,
			).ServeHTTP)
	})
}

func decodeListCustomersReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	ownerID, _ := strconv.ParseInt(q.Get("account_owner_id"), 10, 64)

	params := ListFilterParams{
		Search:         q.Get("search"),
		Status:         q.Get("status"),
		CustomerType:   q.Get("customer_type"),
		Country:        q.Get("country"),
		AccountOwnerID: ownerID,
		IncludeArchived: q.Get("include_archived") == "true",
		SortBy:         q.Get("sort_by"),
		SortOrder:      q.Get("sort_order"),
		Page:           page,
		Limit:          limit,
	}

	return ListReq{OrgID: userCtx.OrgID, Params: params}, nil
}

func decodeOrgReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	return OrgReq{OrgID: userCtx.OrgID}, nil
}

func decodeCustomerReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	return CustomerReq{OrgID: userCtx.OrgID, CustomerID: id}, nil
}

func decodeCreateCustomerReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	var payload CreateCustomerReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return CreateReq{OrgID: userCtx.OrgID, Payload: payload}, nil
}

func decodeUpdateCustomerReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	var payload UpdateCustomerReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return UpdateReq{OrgID: userCtx.OrgID, CustomerID: id, Payload: payload}, nil
}

func decodeCheckDuplicateReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	var payload CheckDuplicateReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return CheckDupReq{OrgID: userCtx.OrgID, Payload: payload}, nil
}

func decodeConvertLeadReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	var payload ConvertLeadReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	var actorID *int64
	if userCtx.UserID > 0 {
		actorID = &userCtx.UserID
	}
	return ConvertReq{OrgID: userCtx.OrgID, ActorUserID: actorID, Payload: payload}, nil
}

func decodeAddContactReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	var payload CreateContactReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return ContactReq{OrgID: userCtx.OrgID, CustomerID: id, Payload: payload}, nil
}

func decodeUpdateContactReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	cIDStr := chi.URLParam(r, "contact_id")
	cID, err := strconv.ParseInt(cIDStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	var payload CreateContactReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return ContactReq{OrgID: userCtx.OrgID, CustomerID: id, ContactID: cID, Payload: payload}, nil
}

func decodeDeleteContactReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	cIDStr := chi.URLParam(r, "contact_id")
	cID, _ := strconv.ParseInt(cIDStr, 10, 64)
	return ContactReq{OrgID: userCtx.OrgID, CustomerID: id, ContactID: cID}, nil
}

func decodeAddAddressReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	var payload CreateAddressReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return AddressReq{OrgID: userCtx.OrgID, CustomerID: id, Payload: payload}, nil
}

func decodeUpdateAddressReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	aIDStr := chi.URLParam(r, "address_id")
	aID, err := strconv.ParseInt(aIDStr, 10, 64)
	if err != nil {
		return nil, ErrInvalidInput
	}
	var payload CreateAddressReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return AddressReq{OrgID: userCtx.OrgID, CustomerID: id, AddressID: aID, Payload: payload}, nil
}

func decodeDeleteAddressReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.ParseInt(idStr, 10, 64)
	aIDStr := chi.URLParam(r, "address_id")
	aID, _ := strconv.ParseInt(aIDStr, 10, 64)
	return AddressReq{OrgID: userCtx.OrgID, CustomerID: id, AddressID: aID}, nil
}

func decodeUpdateFinancialProfileReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, ErrInvalidInput
	}
	var payload UpdateFinancialProfileReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	return UpdateFinancialProfileReqInternal{OrgID: userCtx.OrgID, CustomerID: id, Payload: payload}, nil
}

func decodeUpdateOwnershipReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, ErrInvalidInput
	}
	var payload UpdateOwnershipReq
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, ErrInvalidInput
	}
	var actorID *int64
	if userCtx.UserID > 0 {
		aid := userCtx.UserID
		actorID = &aid
	}
	return UpdateOwnershipReqInternal{OrgID: userCtx.OrgID, CustomerID: id, ActorUserID: actorID, Payload: payload}, nil
}

func decodeResolveRiskReq(ctx context.Context, r *http.Request) (interface{}, error) {
	userCtx, ok := middleware.GetUserContext(ctx)
	if !ok || userCtx.OrgID <= 0 {
		return nil, ErrUnauthorized
	}
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return nil, ErrInvalidInput
	}
	riskIdStr := chi.URLParam(r, "risk_id")
	riskID, err := strconv.ParseInt(riskIdStr, 10, 64)
	if err != nil || riskID <= 0 {
		return nil, ErrInvalidInput
	}
	var payload ResolveRiskReq
	_ = json.NewDecoder(r.Body).Decode(&payload)

	var actorID *int64
	if userCtx.UserID > 0 {
		aid := userCtx.UserID
		actorID = &aid
	}
	return ResolveRiskReqInternal{OrgID: userCtx.OrgID, CustomerID: id, RiskID: riskID, ActorUserID: actorID, Payload: payload}, nil
}



func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	utils.Success(w, http.StatusOK, "Success", response)
	return nil
}

func encodeCreatedResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	utils.Success(w, http.StatusCreated, "Created successfully", response)
	return nil
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if errors.Is(err, ErrUnauthorized) {
		utils.Error(w, http.StatusUnauthorized, "Unauthorized", "AUTH_REQUIRED")
		return
	}
	if errors.Is(err, ErrInvalidInput) {
		utils.Error(w, http.StatusBadRequest, "Invalid input data", "INVALID_INPUT")
		return
	}
	utils.Error(w, http.StatusInternalServerError, err.Error(), "SERVER_ERROR")
}

