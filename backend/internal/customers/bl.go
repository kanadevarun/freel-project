package customers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
)

type BusinessLogic interface {
	ListCustomers(ctx context.Context, orgID int64, params ListFilterParams) ([]Customer, int, error)
	GetCustomerKPIs(ctx context.Context, orgID int64) (CustomerKPIs, error)
	GetCustomerByID(ctx context.Context, orgID int64, customerID int64) (*Customer, error)
	CreateCustomer(ctx context.Context, orgID int64, req CreateCustomerReq) (*Customer, error)
	UpdateCustomer(ctx context.Context, orgID int64, customerID int64, req UpdateCustomerReq) (*Customer, error)
	ArchiveCustomer(ctx context.Context, orgID int64, customerID int64) error
	ReactivateCustomer(ctx context.Context, orgID int64, customerID int64) error
	CheckDuplicate(ctx context.Context, orgID int64, req CheckDuplicateReq) ([]DuplicateMatchResult, error)
	ConvertLeadToCustomer(ctx context.Context, orgID int64, actorUserID *int64, req ConvertLeadReq) (*ConvertLeadResp, error)

	ListContacts(ctx context.Context, orgID int64, customerID int64) ([]CustomerContact, error)
	AddContact(ctx context.Context, orgID int64, customerID int64, req CreateContactReq) (*CustomerContact, error)
	UpdateContact(ctx context.Context, orgID int64, customerID int64, contactID int64, req CreateContactReq) (*CustomerContact, error)
	DeleteContact(ctx context.Context, orgID int64, customerID int64, contactID int64) error

	ListAddresses(ctx context.Context, orgID int64, customerID int64) ([]CustomerAddress, error)
	AddAddress(ctx context.Context, orgID int64, customerID int64, req CreateAddressReq) (*CustomerAddress, error)
	UpdateAddress(ctx context.Context, orgID int64, customerID int64, addressID int64, req CreateAddressReq) (*CustomerAddress, error)
	DeleteAddress(ctx context.Context, orgID int64, customerID int64, addressID int64) error

	GetCustomer360Dashboard(ctx context.Context, orgID int64, customerID int64) (*Customer360Dashboard, error)
	GetCustomerRFQs(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerRFQ, error)
	GetCustomerQuotations(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerQuotation, error)
	GetCustomerBookings(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerBooking, error)
	GetCustomerShipments(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerShipment, error)
	GetCustomerContracts(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerContract, error)
	GetCustomerTimeline(ctx context.Context, orgID int64, customerID int64) ([]CustomerTimelineEvent, error)

	GetFinancialProfile(ctx context.Context, orgID int64, customerID int64) (*CustomerFinancialProfile, error)
	UpdateFinancialProfile(ctx context.Context, orgID int64, customerID int64, req UpdateFinancialProfileReq) (*CustomerFinancialProfile, error)
	GetCommercialMetrics(ctx context.Context, orgID int64, customerID int64) (CustomerCommercialMetrics, error)
	GetAccountOwnership(ctx context.Context, orgID int64, customerID int64) (*CustomerAccountOwnership, error)
	UpdateAccountOwnership(ctx context.Context, orgID int64, customerID int64, actorUserID *int64, req UpdateOwnershipReq) (*CustomerAccountOwnership, error)
	GetOwnershipHistory(ctx context.Context, orgID int64, customerID int64) ([]CustomerOwnershipHistoryItem, error)
	GetRelationshipSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerRelationshipSummary, error)

	EvaluateAndPersistCustomerIntelligence(ctx context.Context, orgID int64, customerID int64) (*CustomerIntelligenceProfile, error)
	GetIntelligenceSummary(ctx context.Context, orgID int64) (CustomerIntelligenceSummary, error)
	GetAttentionItems(ctx context.Context, orgID int64) ([]CustomerAttentionItem, error)
	GetCustomerRisks(ctx context.Context, orgID int64, customerID int64, includeResolved bool) ([]CustomerRiskEvent, error)
	GetCustomerOpportunities(ctx context.Context, orgID int64, customerID int64) ([]CustomerOpportunityEvent, error)
	ResolveCustomerRisk(ctx context.Context, orgID int64, customerID int64, riskID int64, actorUserID *int64, note string) error
}

type businessLogic struct {
	dl DataLayer
}

func NewBusinessLogic(dl DataLayer) BusinessLogic {
	return &businessLogic{dl: dl}
}

func (b *businessLogic) ListCustomers(ctx context.Context, orgID int64, params ListFilterParams) ([]Customer, int, error) {
	return b.dl.List(ctx, orgID, params)
}

func (b *businessLogic) GetCustomerKPIs(ctx context.Context, orgID int64) (CustomerKPIs, error) {
	return b.dl.GetKPIs(ctx, orgID)
}

func (b *businessLogic) GetCustomerByID(ctx context.Context, orgID int64, customerID int64) (*Customer, error) {
	return b.dl.GetByID(ctx, orgID, customerID)
}

func (b *businessLogic) CreateCustomer(ctx context.Context, orgID int64, req CreateCustomerReq) (*Customer, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, fmt.Errorf("customer legal name is required")
	}

	custType := strings.ToUpper(strings.TrimSpace(req.CustomerType))
	if custType == "" {
		custType = TypeShipper
	}

	currency := strings.ToUpper(strings.TrimSpace(req.Currency))
	if currency == "" {
		currency = "USD"
	}

	paymentTerms := strings.ToUpper(strings.TrimSpace(req.PaymentTerms))
	if paymentTerms == "" {
		paymentTerms = "NET30"
	}

	c := &Customer{
		OrgID:          orgID,
		Name:           strings.TrimSpace(req.Name),
		TradingName:    req.TradingName,
		CustomerType:   custType,
		Domain:         req.Domain,
		Industry:       req.Industry,
		TaxID:          req.TaxID,
		PANNumber:      req.PANNumber,
		EORINumber:     req.EORINumber,
		Currency:       currency,
		PaymentTerms:   paymentTerms,
		CreditLimit:    req.CreditLimit,
		HealthScore:    80,
		AccountOwnerID: req.AccountOwnerID,
		Website:        req.Website,
		Country:        req.Country,
		City:           req.City,
		ContactName:    req.ContactName,
		ContactEmail:   req.ContactEmail,
		ContactPhone:   req.ContactPhone,
		Notes:          req.Notes,
		Status:         StatusActive,
	}

	tx, err := b.dl.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	id, err := b.dl.Create(ctx, tx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to insert customer: %w", err)
	}
	c.ID = id

	// Optional primary contact
	if req.PrimaryContact != nil && req.PrimaryContact.FirstName != "" {
		contact := &CustomerContact{
			OrgID:      orgID,
			CustomerID: id,
			FirstName:  req.PrimaryContact.FirstName,
			LastName:   req.PrimaryContact.LastName,
			Email:      req.PrimaryContact.Email,
			Phone:      req.PrimaryContact.Phone,
			JobTitle:   req.PrimaryContact.JobTitle,
			Department: req.PrimaryContact.Department,
			IsPrimary:  true,
			Notes:      req.PrimaryContact.Notes,
		}
		_, _ = b.dl.CreateContact(ctx, tx, contact)
	} else if req.ContactName != nil && *req.ContactName != "" {
		parts := strings.SplitN(*req.ContactName, " ", 2)
		fn := parts[0]
		ln := ""
		if len(parts) > 1 {
			ln = parts[1]
		}
		contact := &CustomerContact{
			OrgID:      orgID,
			CustomerID: id,
			FirstName:  fn,
			LastName:   ln,
			Email:      req.ContactEmail,
			Phone:      req.ContactPhone,
			IsPrimary:  true,
		}
		_, _ = b.dl.CreateContact(ctx, tx, contact)
	}

	// Optional primary billing address
	if req.BillingAddress != nil && req.BillingAddress.AddressLine1 != "" {
		addr := &CustomerAddress{
			OrgID:             orgID,
			CustomerID:        id,
			AddressType:       AddressTypeBilling,
			Label:             req.BillingAddress.Label,
			AddressLine1:      req.BillingAddress.AddressLine1,
			AddressLine2:      req.BillingAddress.AddressLine2,
			City:              req.BillingAddress.City,
			State:             req.BillingAddress.State,
			PostalCode:        req.BillingAddress.PostalCode,
			CountryCode:       req.BillingAddress.CountryCode,
			Country:           req.BillingAddress.Country,
			IsPrimaryBilling:  true,
			IsPrimaryShipping: req.BillingAddress.IsPrimaryShipping,
			ContactName:       req.BillingAddress.ContactName,
			ContactPhone:      req.BillingAddress.ContactPhone,
			ContactEmail:      req.BillingAddress.ContactEmail,
		}
		_, _ = b.dl.CreateAddress(ctx, tx, addr)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleCustomers,
		ResourceType: "CUSTOMER",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: c.Name,
		Description:  fmt.Sprintf("Created customer %s", c.Name),
		Result:       domain.ResultSuccess,
	})

	return b.GetCustomerByID(ctx, orgID, id)
}

func (b *businessLogic) UpdateCustomer(ctx context.Context, orgID int64, customerID int64, req UpdateCustomerReq) (*Customer, error) {
	c, err := b.dl.GetByID(ctx, orgID, customerID)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("customer not found")
	}

	oldName := c.Name
	oldStatus := c.Status

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		c.Name = strings.TrimSpace(*req.Name)
	}
	if req.TradingName != nil {
		c.TradingName = req.TradingName
	}
	if req.CustomerType != nil && *req.CustomerType != "" {
		c.CustomerType = strings.ToUpper(*req.CustomerType)
	}
	if req.Domain != nil {
		c.Domain = req.Domain
	}
	if req.Industry != nil {
		c.Industry = req.Industry
	}
	if req.TaxID != nil {
		c.TaxID = req.TaxID
	}
	if req.PANNumber != nil {
		c.PANNumber = req.PANNumber
	}
	if req.EORINumber != nil {
		c.EORINumber = req.EORINumber
	}
	if req.Currency != nil && *req.Currency != "" {
		c.Currency = strings.ToUpper(*req.Currency)
	}
	if req.PaymentTerms != nil && *req.PaymentTerms != "" {
		c.PaymentTerms = strings.ToUpper(*req.PaymentTerms)
	}
	if req.CreditLimit != nil {
		c.CreditLimit = *req.CreditLimit
	}
	if req.HealthScore != nil {
		c.HealthScore = *req.HealthScore
	}
	if req.AccountOwnerID != nil {
		c.AccountOwnerID = req.AccountOwnerID
	}
	if req.Website != nil {
		c.Website = req.Website
	}
	if req.Country != nil {
		c.Country = req.Country
	}
	if req.City != nil {
		c.City = req.City
	}
	if req.ContactName != nil {
		c.ContactName = req.ContactName
	}
	if req.ContactEmail != nil {
		c.ContactEmail = req.ContactEmail
	}
	if req.ContactPhone != nil {
		c.ContactPhone = req.ContactPhone
	}
	if req.Notes != nil {
		c.Notes = req.Notes
	}
	if req.Status != nil && *req.Status != "" {
		c.Status = strings.ToUpper(*req.Status)
	}

	err = b.dl.Update(ctx, c)
	if err != nil {
		return nil, err
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleCustomers,
		ResourceType: "CUSTOMER",
		ResourceID:   fmt.Sprintf("%d", customerID),
		ResourceName: c.Name,
		Description:  fmt.Sprintf("Updated customer %s", c.Name),
		Before:       map[string]interface{}{"name": oldName, "status": oldStatus},
		After:        map[string]interface{}{"name": c.Name, "status": c.Status},
		Result:       domain.ResultSuccess,
	})

	return b.GetCustomerByID(ctx, orgID, customerID)
}

func (b *businessLogic) ArchiveCustomer(ctx context.Context, orgID int64, customerID int64) error {
	err := b.dl.Archive(ctx, orgID, customerID)
	if err == nil {
		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			Action:       domain.ActionDelete,
			Module:       domain.ModuleCustomers,
			ResourceType: "CUSTOMER",
			ResourceID:   fmt.Sprintf("%d", customerID),
			Description:  fmt.Sprintf("Archived customer #%d", customerID),
			Result:       domain.ResultSuccess,
		})
	}
	return err
}

func (b *businessLogic) ReactivateCustomer(ctx context.Context, orgID int64, customerID int64) error {
	err := b.dl.Reactivate(ctx, orgID, customerID)
	if err == nil {
		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        orgID,
			Action:       domain.ActionEnable,
			Module:       domain.ModuleCustomers,
			ResourceType: "CUSTOMER",
			ResourceID:   fmt.Sprintf("%d", customerID),
			Description:  fmt.Sprintf("Reactivated customer #%d", customerID),
			Result:       domain.ResultSuccess,
		})
	}
	return err
}

func (b *businessLogic) CheckDuplicate(ctx context.Context, orgID int64, req CheckDuplicateReq) ([]DuplicateMatchResult, error) {
	candidates, err := b.dl.GetCandidatesForDuplicateCheck(ctx, orgID)
	if err != nil {
		return nil, err
	}

	results := []DuplicateMatchResult{}
	for _, candidate := range candidates {
		score, reason := EvaluateDuplicateScore(req, candidate)
		if score >= 50 {
			primaryContactStr := ""
			if candidate.ContactName != nil {
				primaryContactStr = *candidate.ContactName
			}
			results = append(results, DuplicateMatchResult{
				CustomerID:      candidate.ID,
				CustomerCode:    candidate.CustomerCode,
				CustomerName:    candidate.Name,
				ConfidenceScore: score,
				MatchReason:     reason,
				ExistingStatus:  candidate.Status,
				PrimaryContact:  primaryContactStr,
			})
		}
	}
	return results, nil
}

func (b *businessLogic) ConvertLeadToCustomer(ctx context.Context, orgID int64, actorUserID *int64, req ConvertLeadReq) (*ConvertLeadResp, error) {
	lead, err := b.dl.GetLeadByID(ctx, orgID, req.LeadID)
	if err != nil {
		return nil, fmt.Errorf("lead not found: %w", err)
	}

	tx, err := b.dl.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var targetCustomerID int64
	var customerCode string
	isNewAccount := false

	// Scenario A: User selected to link to an existing customer account
	if req.LinkToExistingCustomerID != nil && *req.LinkToExistingCustomerID > 0 {
		targetCustomerID = *req.LinkToExistingCustomerID
		existingCust, err := b.dl.GetByID(ctx, orgID, targetCustomerID)
		if err != nil {
			return nil, fmt.Errorf("selected customer account not found: %w", err)
		}
		customerCode = existingCust.CustomerCode

		// Optionally attach contact from lead if not present
		if lead.ContactName != nil && *lead.ContactName != "" {
			parts := strings.SplitN(*lead.ContactName, " ", 2)
			fn := parts[0]
			ln := ""
			if len(parts) > 1 {
				ln = parts[1]
			}
			contact := &CustomerContact{
				OrgID:      orgID,
				CustomerID: targetCustomerID,
				FirstName:  fn,
				LastName:   ln,
				Email:      lead.Email,
				Phone:      lead.Phone,
				IsPrimary:  false,
			}
			_, _ = b.dl.CreateContact(ctx, tx, contact)
		}
	} else {
		// Scenario B: Create a brand new customer entity
		isNewAccount = true

		custType := strings.ToUpper(strings.TrimSpace(req.CustomerType))
		if custType == "" {
			custType = TypeShipper
		}

		c := &Customer{
			OrgID:          orgID,
			Name:           lead.CompanyName,
			CustomerType:   custType,
			TaxID:          req.TaxID,
			Currency:       "USD",
			PaymentTerms:   "NET30",
			CreditLimit:    0.00,
			HealthScore:    80,
			AccountOwnerID: req.AccountOwnerID,
			ContactName:    lead.ContactName,
			ContactEmail:   lead.Email,
			ContactPhone:   lead.Phone,
			Country:        lead.Location,
			Notes:          req.Notes,
			Status:         StatusActive,
		}

		id, err := b.dl.Create(ctx, tx, c)
		if err != nil {
			return nil, fmt.Errorf("failed to create customer from lead: %w", err)
		}
		targetCustomerID = id
		customerCode = fmt.Sprintf("CUST-%d-%05d", time.Now().Year(), id)

		// Create primary contact
		if lead.ContactName != nil && *lead.ContactName != "" {
			parts := strings.SplitN(*lead.ContactName, " ", 2)
			fn := parts[0]
			ln := ""
			if len(parts) > 1 {
				ln = parts[1]
			}
			contact := &CustomerContact{
				OrgID:      orgID,
				CustomerID: targetCustomerID,
				FirstName:  fn,
				LastName:   ln,
				Email:      lead.Email,
				Phone:      lead.Phone,
				IsPrimary:  true,
			}
			_, _ = b.dl.CreateContact(ctx, tx, contact)
		}
	}

	// Record lead link audit record
	noteStr := fmt.Sprintf("Converted from Lead #%d", lead.ID)
	if req.Notes != nil && *req.Notes != "" {
		noteStr += ": " + *req.Notes
	}
	link := &CustomerLeadLink{
		OrgID:             orgID,
		CustomerID:        targetCustomerID,
		LeadID:            lead.ID,
		ConvertedByUserID: actorUserID,
		ConversionNotes:   &noteStr,
	}
	_, err = b.dl.CreateLeadLink(ctx, tx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to record lead conversion link: %w", err)
	}

	// Update lead status to CONVERTED
	err = b.dl.UpdateLeadStatusToConverted(ctx, tx, orgID, lead.ID, targetCustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to update lead status: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	msg := fmt.Sprintf("Lead successfully converted to Customer Account %s", customerCode)
	if !isNewAccount {
		msg = fmt.Sprintf("Lead successfully linked to existing Customer Account %s", customerCode)
	}

	return &ConvertLeadResp{
		CustomerID:   targetCustomerID,
		CustomerCode: customerCode,
		LeadID:       lead.ID,
		IsNewAccount: isNewAccount,
		Message:      msg,
	}, nil
}

func (b *businessLogic) ListContacts(ctx context.Context, orgID int64, customerID int64) ([]CustomerContact, error) {
	return b.dl.ListContacts(ctx, orgID, customerID)
}

func (b *businessLogic) AddContact(ctx context.Context, orgID int64, customerID int64, req CreateContactReq) (*CustomerContact, error) {
	cRole := req.ContactRole
	if cRole == "" {
		cRole = ContactRoleCommercial
	}
	contact := &CustomerContact{
		OrgID:       orgID,
		CustomerID:  customerID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		Mobile:      req.Mobile,
		JobTitle:    req.JobTitle,
		Department:  req.Department,
		ContactRole: cRole,
		IsPrimary:   req.IsPrimary,
		Notes:       req.Notes,
	}
	id, err := b.dl.CreateContact(ctx, nil, contact)
	if err != nil {
		return nil, err
	}
	contact.ID = id
	return contact, nil
}

func (b *businessLogic) UpdateContact(ctx context.Context, orgID int64, customerID int64, contactID int64, req CreateContactReq) (*CustomerContact, error) {
	cRole := req.ContactRole
	if cRole == "" {
		cRole = ContactRoleCommercial
	}
	contact := &CustomerContact{
		ID:          contactID,
		OrgID:       orgID,
		CustomerID:  customerID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Email:       req.Email,
		Phone:       req.Phone,
		Mobile:      req.Mobile,
		JobTitle:    req.JobTitle,
		Department:  req.Department,
		ContactRole: cRole,
		IsPrimary:   req.IsPrimary,
		Notes:       req.Notes,
	}
	err := b.dl.UpdateContact(ctx, contact)
	if err != nil {
		return nil, err
	}
	return contact, nil
}

func (b *businessLogic) DeleteContact(ctx context.Context, orgID int64, customerID int64, contactID int64) error {
	return b.dl.DeleteContact(ctx, orgID, contactID)
}

func (b *businessLogic) ListAddresses(ctx context.Context, orgID int64, customerID int64) ([]CustomerAddress, error) {
	return b.dl.ListAddresses(ctx, orgID, customerID)
}

func (b *businessLogic) AddAddress(ctx context.Context, orgID int64, customerID int64, req CreateAddressReq) (*CustomerAddress, error) {
	addrType := strings.ToUpper(strings.TrimSpace(req.AddressType))
	if addrType == "" {
		addrType = AddressTypeBilling
	}
	addr := &CustomerAddress{
		OrgID:             orgID,
		CustomerID:        customerID,
		AddressType:       addrType,
		Label:             req.Label,
		AddressLine1:      req.AddressLine1,
		AddressLine2:      req.AddressLine2,
		City:              req.City,
		State:             req.State,
		PostalCode:        req.PostalCode,
		CountryCode:       req.CountryCode,
		Country:           req.Country,
		IsPrimaryBilling:  req.IsPrimaryBilling,
		IsPrimaryShipping: req.IsPrimaryShipping,
		ContactName:       req.ContactName,
		ContactPhone:      req.ContactPhone,
		ContactEmail:      req.ContactEmail,
	}
	id, err := b.dl.CreateAddress(ctx, nil, addr)
	if err != nil {
		return nil, err
	}
	addr.ID = id
	return addr, nil
}

func (b *businessLogic) UpdateAddress(ctx context.Context, orgID int64, customerID int64, addressID int64, req CreateAddressReq) (*CustomerAddress, error) {
	addrType := strings.ToUpper(strings.TrimSpace(req.AddressType))
	if addrType == "" {
		addrType = AddressTypeBilling
	}
	addr := &CustomerAddress{
		ID:                addressID,
		OrgID:             orgID,
		CustomerID:        customerID,
		AddressType:       addrType,
		Label:             req.Label,
		AddressLine1:      req.AddressLine1,
		AddressLine2:      req.AddressLine2,
		City:              req.City,
		State:             req.State,
		PostalCode:        req.PostalCode,
		CountryCode:       req.CountryCode,
		Country:           req.Country,
		IsPrimaryBilling:  req.IsPrimaryBilling,
		IsPrimaryShipping: req.IsPrimaryShipping,
		ContactName:       req.ContactName,
		ContactPhone:      req.ContactPhone,
		ContactEmail:      req.ContactEmail,
	}
	err := b.dl.UpdateAddress(ctx, addr)
	if err != nil {
		return nil, err
	}
	return addr, nil
}

func (b *businessLogic) DeleteAddress(ctx context.Context, orgID int64, customerID int64, addressID int64) error {
	return b.dl.DeleteAddress(ctx, orgID, addressID)
}

func (b *businessLogic) GetCustomer360Dashboard(ctx context.Context, orgID int64, customerID int64) (*Customer360Dashboard, error) {
	return b.dl.GetCustomer360Dashboard(ctx, orgID, customerID)
}

func (b *businessLogic) GetCustomerRFQs(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerRFQ, error) {
	return b.dl.GetCustomerRFQs(ctx, orgID, customerID, limit)
}

func (b *businessLogic) GetCustomerQuotations(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerQuotation, error) {
	return b.dl.GetCustomerQuotations(ctx, orgID, customerID, limit)
}

func (b *businessLogic) GetCustomerBookings(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerBooking, error) {
	return b.dl.GetCustomerBookings(ctx, orgID, customerID, limit)
}

func (b *businessLogic) GetCustomerShipments(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerShipment, error) {
	return b.dl.GetCustomerShipments(ctx, orgID, customerID, limit)
}

func (b *businessLogic) GetCustomerContracts(ctx context.Context, orgID int64, customerID int64, limit int) ([]CustomerContract, error) {
	return b.dl.GetCustomerContracts(ctx, orgID, customerID, limit)
}

func (b *businessLogic) GetCustomerTimeline(ctx context.Context, orgID int64, customerID int64) ([]CustomerTimelineEvent, error) {
	return b.dl.GetCustomerTimeline(ctx, orgID, customerID)
}

func (b *businessLogic) GetFinancialProfile(ctx context.Context, orgID int64, customerID int64) (*CustomerFinancialProfile, error) {
	return b.dl.GetFinancialProfile(ctx, orgID, customerID)
}

func (b *businessLogic) UpdateFinancialProfile(ctx context.Context, orgID int64, customerID int64, req UpdateFinancialProfileReq) (*CustomerFinancialProfile, error) {
	return b.dl.UpdateFinancialProfile(ctx, orgID, customerID, req)
}

func (b *businessLogic) GetCommercialMetrics(ctx context.Context, orgID int64, customerID int64) (CustomerCommercialMetrics, error) {
	return b.dl.GetCommercialMetrics(ctx, orgID, customerID)
}

func (b *businessLogic) GetAccountOwnership(ctx context.Context, orgID int64, customerID int64) (*CustomerAccountOwnership, error) {
	return b.dl.GetAccountOwnership(ctx, orgID, customerID)
}

func (b *businessLogic) UpdateAccountOwnership(ctx context.Context, orgID int64, customerID int64, actorUserID *int64, req UpdateOwnershipReq) (*CustomerAccountOwnership, error) {
	return b.dl.UpdateAccountOwnership(ctx, orgID, customerID, actorUserID, req)
}

func (b *businessLogic) GetOwnershipHistory(ctx context.Context, orgID int64, customerID int64) ([]CustomerOwnershipHistoryItem, error) {
	return b.dl.GetOwnershipHistory(ctx, orgID, customerID)
}

func (b *businessLogic) GetRelationshipSummary(ctx context.Context, orgID int64, customerID int64) (*CustomerRelationshipSummary, error) {
	return b.dl.GetRelationshipSummary(ctx, orgID, customerID)
}

func (b *businessLogic) EvaluateAndPersistCustomerIntelligence(ctx context.Context, orgID int64, customerID int64) (*CustomerIntelligenceProfile, error) {
	return b.dl.EvaluateAndPersistCustomerIntelligence(ctx, orgID, customerID)
}

func (b *businessLogic) GetIntelligenceSummary(ctx context.Context, orgID int64) (CustomerIntelligenceSummary, error) {
	return b.dl.GetIntelligenceSummary(ctx, orgID)
}

func (b *businessLogic) GetAttentionItems(ctx context.Context, orgID int64) ([]CustomerAttentionItem, error) {
	return b.dl.GetAttentionItems(ctx, orgID)
}

func (b *businessLogic) GetCustomerRisks(ctx context.Context, orgID int64, customerID int64, includeResolved bool) ([]CustomerRiskEvent, error) {
	return b.dl.GetCustomerRisks(ctx, orgID, customerID, includeResolved)
}

func (b *businessLogic) GetCustomerOpportunities(ctx context.Context, orgID int64, customerID int64) ([]CustomerOpportunityEvent, error) {
	return b.dl.GetCustomerOpportunities(ctx, orgID, customerID)
}

func (b *businessLogic) ResolveCustomerRisk(ctx context.Context, orgID int64, customerID int64, riskID int64, actorUserID *int64, note string) error {
	return b.dl.ResolveCustomerRisk(ctx, orgID, customerID, riskID, actorUserID, note)
}



