package customers

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type Endpoints struct {
	ListCustomersEndpoint    endpoint.Endpoint
	GetKPIsEndpoint          endpoint.Endpoint
	GetCustomerEndpoint       endpoint.Endpoint
	CreateCustomerEndpoint   endpoint.Endpoint
	UpdateCustomerEndpoint   endpoint.Endpoint
	ArchiveCustomerEndpoint  endpoint.Endpoint
	ReactivateCustomerEndpoint endpoint.Endpoint
	CheckDuplicateEndpoint   endpoint.Endpoint
	ConvertLeadEndpoint      endpoint.Endpoint
	ListContactsEndpoint     endpoint.Endpoint
	AddContactEndpoint       endpoint.Endpoint
	UpdateContactEndpoint    endpoint.Endpoint
	DeleteContactEndpoint    endpoint.Endpoint
	ListAddressesEndpoint    endpoint.Endpoint
	AddAddressEndpoint       endpoint.Endpoint
	UpdateAddressEndpoint    endpoint.Endpoint
	DeleteAddressEndpoint    endpoint.Endpoint

	Get360DashboardEndpoint      endpoint.Endpoint
	GetCustomerRFQsEndpoint      endpoint.Endpoint
	GetCustomerQuotationsEndpoint endpoint.Endpoint
	GetCustomerBookingsEndpoint  endpoint.Endpoint
	GetCustomerShipmentsEndpoint endpoint.Endpoint
	GetCustomerContractsEndpoint endpoint.Endpoint
	GetCustomerTimelineEndpoint  endpoint.Endpoint

	GetFinancialProfileEndpoint      endpoint.Endpoint
	UpdateFinancialProfileEndpoint   endpoint.Endpoint
	GetCommercialMetricsEndpoint     endpoint.Endpoint
	GetAccountOwnershipEndpoint      endpoint.Endpoint
	UpdateAccountOwnershipEndpoint   endpoint.Endpoint
	GetOwnershipHistoryEndpoint      endpoint.Endpoint
	GetRelationshipSummaryEndpoint   endpoint.Endpoint

	GetIntelligenceSummaryEndpoint         endpoint.Endpoint
	GetAttentionItemsEndpoint              endpoint.Endpoint
	EvaluateCustomerIntelligenceEndpoint  endpoint.Endpoint
	GetCustomerRisksEndpoint               endpoint.Endpoint
	GetCustomerOpportunitiesEndpoint       endpoint.Endpoint
	ResolveCustomerRiskEndpoint            endpoint.Endpoint
}

func MakeEndpoints(s BusinessLogic) Endpoints {
	return Endpoints{
		ListCustomersEndpoint:        MakeListCustomersEndpoint(s),
		GetKPIsEndpoint:              MakeGetKPIsEndpoint(s),
		GetCustomerEndpoint:           MakeGetCustomerEndpoint(s),
		CreateCustomerEndpoint:       MakeCreateCustomerEndpoint(s),
		UpdateCustomerEndpoint:       MakeUpdateCustomerEndpoint(s),
		ArchiveCustomerEndpoint:      MakeArchiveCustomerEndpoint(s),
		ReactivateCustomerEndpoint:   MakeReactivateCustomerEndpoint(s),
		CheckDuplicateEndpoint:       MakeCheckDuplicateEndpoint(s),
		ConvertLeadEndpoint:          MakeConvertLeadEndpoint(s),
		ListContactsEndpoint:         MakeListContactsEndpoint(s),
		AddContactEndpoint:           MakeAddContactEndpoint(s),
		UpdateContactEndpoint:        MakeUpdateContactEndpoint(s),
		DeleteContactEndpoint:        MakeDeleteContactEndpoint(s),
		ListAddressesEndpoint:        MakeListAddressesEndpoint(s),
		AddAddressEndpoint:           MakeAddAddressEndpoint(s),
		UpdateAddressEndpoint:        MakeUpdateAddressEndpoint(s),
		DeleteAddressEndpoint:        MakeDeleteAddressEndpoint(s),

		Get360DashboardEndpoint:      MakeGet360DashboardEndpoint(s),
		GetCustomerRFQsEndpoint:      MakeGetCustomerRFQsEndpoint(s),
		GetCustomerQuotationsEndpoint: MakeGetCustomerQuotationsEndpoint(s),
		GetCustomerBookingsEndpoint:  MakeGetCustomerBookingsEndpoint(s),
		GetCustomerShipmentsEndpoint: MakeGetCustomerShipmentsEndpoint(s),
		GetCustomerContractsEndpoint: MakeGetCustomerContractsEndpoint(s),
		GetCustomerTimelineEndpoint:  MakeGetCustomerTimelineEndpoint(s),

		GetFinancialProfileEndpoint:    MakeGetFinancialProfileEndpoint(s),
		UpdateFinancialProfileEndpoint: MakeUpdateFinancialProfileEndpoint(s),
		GetCommercialMetricsEndpoint:   MakeGetCommercialMetricsEndpoint(s),
		GetAccountOwnershipEndpoint:    MakeGetAccountOwnershipEndpoint(s),
		UpdateAccountOwnershipEndpoint: MakeUpdateAccountOwnershipEndpoint(s),
		GetOwnershipHistoryEndpoint:    MakeGetOwnershipHistoryEndpoint(s),
		GetRelationshipSummaryEndpoint: MakeGetRelationshipSummaryEndpoint(s),

		GetIntelligenceSummaryEndpoint:        MakeGetIntelligenceSummaryEndpoint(s),
		GetAttentionItemsEndpoint:             MakeGetAttentionItemsEndpoint(s),
		EvaluateCustomerIntelligenceEndpoint: MakeEvaluateCustomerIntelligenceEndpoint(s),
		GetCustomerRisksEndpoint:              MakeGetCustomerRisksEndpoint(s),
		GetCustomerOpportunitiesEndpoint:      MakeGetCustomerOpportunitiesEndpoint(s),
		ResolveCustomerRiskEndpoint:           MakeResolveCustomerRiskEndpoint(s),
	}
}

type ListReq struct {
	OrgID  int64
	Params ListFilterParams
}

type ListResp struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

func MakeListCustomersEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListReq)
		list, total, err := s.ListCustomers(ctx, req.OrgID, req.Params)
		if err != nil {
			return nil, err
		}
		page := req.Params.Page
		if page <= 0 {
			page = 1
		}
		limit := req.Params.Limit
		if limit <= 0 {
			limit = 10
		}
		return ListResp{Customers: list, Total: total, Page: page, Limit: limit}, nil
	}
}

type OrgReq struct {
	OrgID int64
}

func MakeGetKPIsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OrgReq)
		return s.GetCustomerKPIs(ctx, req.OrgID)
	}
}

type CustomerReq struct {
	OrgID      int64
	CustomerID int64
}

func MakeGetCustomerEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerByID(ctx, req.OrgID, req.CustomerID)
	}
}

type CreateReq struct {
	OrgID   int64
	Payload CreateCustomerReq
}

func MakeCreateCustomerEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CreateReq)
		return s.CreateCustomer(ctx, req.OrgID, req.Payload)
	}
}

type UpdateReq struct {
	OrgID      int64
	CustomerID int64
	Payload    UpdateCustomerReq
}

func MakeUpdateCustomerEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateReq)
		return s.UpdateCustomer(ctx, req.OrgID, req.CustomerID, req.Payload)
	}
}

func MakeArchiveCustomerEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		err := s.ArchiveCustomer(ctx, req.OrgID, req.CustomerID)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": "Customer archived successfully"}, nil
	}
}

func MakeReactivateCustomerEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		err := s.ReactivateCustomer(ctx, req.OrgID, req.CustomerID)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": "Customer reactivated successfully"}, nil
	}
}

type CheckDupReq struct {
	OrgID   int64
	Payload CheckDuplicateReq
}

func MakeCheckDuplicateEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CheckDupReq)
		return s.CheckDuplicate(ctx, req.OrgID, req.Payload)
	}
}

type ConvertReq struct {
	OrgID       int64
	ActorUserID *int64
	Payload     ConvertLeadReq
}

func MakeConvertLeadEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ConvertReq)
		return s.ConvertLeadToCustomer(ctx, req.OrgID, req.ActorUserID, req.Payload)
	}
}

func MakeListContactsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.ListContacts(ctx, req.OrgID, req.CustomerID)
	}
}

type ContactReq struct {
	OrgID      int64
	CustomerID int64
	ContactID  int64
	Payload    CreateContactReq
}

func MakeAddContactEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ContactReq)
		return s.AddContact(ctx, req.OrgID, req.CustomerID, req.Payload)
	}
}

func MakeUpdateContactEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ContactReq)
		return s.UpdateContact(ctx, req.OrgID, req.CustomerID, req.ContactID, req.Payload)
	}
}

func MakeDeleteContactEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ContactReq)
		err := s.DeleteContact(ctx, req.OrgID, req.CustomerID, req.ContactID)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": "Contact deleted successfully"}, nil
	}
}

func MakeListAddressesEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.ListAddresses(ctx, req.OrgID, req.CustomerID)
	}
}

type AddressReq struct {
	OrgID      int64
	CustomerID int64
	AddressID  int64
	Payload    CreateAddressReq
}

func MakeAddAddressEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddressReq)
		return s.AddAddress(ctx, req.OrgID, req.CustomerID, req.Payload)
	}
}

func MakeUpdateAddressEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddressReq)
		return s.UpdateAddress(ctx, req.OrgID, req.CustomerID, req.AddressID, req.Payload)
	}
}

func MakeDeleteAddressEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddressReq)
		err := s.DeleteAddress(ctx, req.OrgID, req.CustomerID, req.AddressID)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": "Address deleted successfully"}, nil
	}
}

func MakeGet360DashboardEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomer360Dashboard(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeGetCustomerRFQsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerRFQs(ctx, req.OrgID, req.CustomerID, 50)
	}
}

func MakeGetCustomerQuotationsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerQuotations(ctx, req.OrgID, req.CustomerID, 50)
	}
}

func MakeGetCustomerBookingsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerBookings(ctx, req.OrgID, req.CustomerID, 50)
	}
}

func MakeGetCustomerShipmentsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerShipments(ctx, req.OrgID, req.CustomerID, 50)
	}
}

func MakeGetCustomerContractsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerContracts(ctx, req.OrgID, req.CustomerID, 50)
	}
}

func MakeGetCustomerTimelineEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerTimeline(ctx, req.OrgID, req.CustomerID)
	}
}

type UpdateFinancialProfileReqInternal struct {
	OrgID      int64
	CustomerID int64
	Payload    UpdateFinancialProfileReq
}

func MakeGetFinancialProfileEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetFinancialProfile(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeUpdateFinancialProfileEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateFinancialProfileReqInternal)
		return s.UpdateFinancialProfile(ctx, req.OrgID, req.CustomerID, req.Payload)
	}
}

func MakeGetCommercialMetricsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCommercialMetrics(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeGetAccountOwnershipEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetAccountOwnership(ctx, req.OrgID, req.CustomerID)
	}
}

type UpdateOwnershipReqInternal struct {
	OrgID       int64
	CustomerID  int64
	ActorUserID *int64
	Payload     UpdateOwnershipReq
}

func MakeUpdateAccountOwnershipEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateOwnershipReqInternal)
		return s.UpdateAccountOwnership(ctx, req.OrgID, req.CustomerID, req.ActorUserID, req.Payload)
	}
}

func MakeGetOwnershipHistoryEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetOwnershipHistory(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeGetRelationshipSummaryEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetRelationshipSummary(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeGetIntelligenceSummaryEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OrgReq)
		return s.GetIntelligenceSummary(ctx, req.OrgID)
	}
}

func MakeGetAttentionItemsEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(OrgReq)
		return s.GetAttentionItems(ctx, req.OrgID)
	}
}

func MakeEvaluateCustomerIntelligenceEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.EvaluateAndPersistCustomerIntelligence(ctx, req.OrgID, req.CustomerID)
	}
}

func MakeGetCustomerRisksEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerRisks(ctx, req.OrgID, req.CustomerID, true)
	}
}

func MakeGetCustomerOpportunitiesEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(CustomerReq)
		return s.GetCustomerOpportunities(ctx, req.OrgID, req.CustomerID)
	}
}

type ResolveRiskReqInternal struct {
	OrgID       int64
	CustomerID  int64
	RiskID      int64
	ActorUserID *int64
	Payload     ResolveRiskReq
}

func MakeResolveCustomerRiskEndpoint(s BusinessLogic) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ResolveRiskReqInternal)
		err := s.ResolveCustomerRisk(ctx, req.OrgID, req.CustomerID, req.RiskID, req.ActorUserID, req.Payload.ResolutionNote)
		if err != nil {
			return nil, err
		}
		return map[string]string{"message": "Risk event resolved successfully"}, nil
	}
}



