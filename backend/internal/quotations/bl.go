package quotations

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/rates"
	"github.com/freel/backend/internal/svcerror"
)

// Service defines the business-logic operations for the quotation domain.
type Service interface {
	CreateQuotation(ctx context.Context, orgID, userID int64, req *CreateQuotationRequest) (*Quotation, error)
	UpdateQuotation(ctx context.Context, orgID, quotationID, userID int64, req *UpdateQuotationRequest) (*Quotation, error)
	GetQuotation(ctx context.Context, orgID, quotationID int64) (*QuotationDetail, error)
	ListQuotations(ctx context.Context, filters *QuotationListFilters) (*QuotationsListResponse, error)
	GetQuotationSummary(ctx context.Context, orgID int64) (*QuotationSummary, error)

	// Pricing & Commercial Calculation Engine (Task 18.2)
	GetQuotationPricing(ctx context.Context, orgID, quotationID int64) (*QuotationPricing, error)
	AddQuotationCharge(ctx context.Context, orgID, quotationID, userID int64, req *CreateQuotationChargeRequest) (*QuotationPricing, error)
	UpdateQuotationCharge(ctx context.Context, orgID, quotationID, chargeID, userID int64, req *UpdateQuotationChargeRequest) (*QuotationPricing, error)
	DeleteQuotationCharge(ctx context.Context, orgID, quotationID, chargeID, userID int64) (*QuotationPricing, error)
	ReorderQuotationCharges(ctx context.Context, orgID, quotationID int64, req *ReorderQuotationChargesRequest) (*QuotationPricing, error)
	GetRateCandidates(ctx context.Context, orgID, quotationID int64) ([]RateCandidate, error)
	ImportRateCharges(ctx context.Context, orgID, quotationID, userID int64, req *ImportRateChargesRequest) (*QuotationPricing, error)

	// Reusable Quotation Templates & Commercial Terms (Task 18.3)
	ListQuotationTemplates(ctx context.Context, orgID int64, activeOnly bool) ([]*QuotationTemplate, error)
	GetQuotationTemplate(ctx context.Context, orgID, templateID int64) (*QuotationTemplateDetail, error)
	CreateQuotationTemplate(ctx context.Context, orgID, userID int64, req *CreateQuotationTemplateRequest) (*QuotationTemplate, error)
	UpdateQuotationTemplate(ctx context.Context, orgID, templateID, userID int64, req *UpdateQuotationTemplateRequest) (*QuotationTemplate, error)
	DeleteQuotationTemplate(ctx context.Context, orgID, templateID int64) error
	CreateTemplateFromQuotation(ctx context.Context, orgID, quotationID, userID int64, req *CreateTemplateFromQuotationRequest) (*QuotationTemplate, error)
	ApplyTemplateToQuotation(ctx context.Context, orgID, quotationID, userID int64, req *ApplyQuotationTemplateRequest) (*QuotationPricing, error)
	UpdateQuotationCommercialTerms(ctx context.Context, orgID, quotationID, userID int64, req *UpdateQuotationCommercialTermsRequest) (*QuotationDetail, error)

	// Approval & Lifecycle (Task 18.4)
	SubmitQuotationForReview(ctx context.Context, orgID, quotationID, userID int64, req *SubmitQuotationForReviewRequest) (*QuotationDetail, error)
	ApproveQuotation(ctx context.Context, orgID, quotationID, userID int64, req *ApproveQuotationRequest) (*QuotationDetail, error)
	RequestQuotationChanges(ctx context.Context, orgID, quotationID, userID int64, req *RequestQuotationChangesRequest) (*QuotationDetail, error)
	SendQuotation(ctx context.Context, orgID, quotationID, userID int64, req *SendQuotationRequest) (*QuotationDetail, error)
	GetCustomerQuotationPreview(ctx context.Context, orgID, quotationID int64) (*CustomerQuotationPreview, error)
	GetQuotationApprovalHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationApprovalHistory, error)
	GetQuotationApprovalStatus(ctx context.Context, orgID, quotationID int64) (*QuotationApprovalStatus, error)
	MarkQuotationViewed(ctx context.Context, orgID, quotationID int64, req *MarkQuotationViewedRequest) error
	AcceptQuotation(ctx context.Context, orgID, quotationID, userID int64, req *AcceptQuotationRequest) (*QuotationDetail, error)
	DeclineQuotation(ctx context.Context, orgID, quotationID, userID int64, req *DeclineQuotationRequest) (*QuotationDetail, error)
	CancelQuotation(ctx context.Context, orgID, quotationID, userID int64, req *CancelQuotationRequest) (*QuotationDetail, error)

	// Documents & Public Sharing (Task 18.5)
	GenerateQuotationDocument(ctx context.Context, orgID, quotationID, userID int64, docType string) (*QuotationDocument, error)
	ListQuotationDocuments(ctx context.Context, orgID, quotationID int64) ([]*QuotationDocument, error)
	GetQuotationDocument(ctx context.Context, orgID, quotationID, docID int64) (*QuotationDocument, []byte, error)
	CreateQuotationPublicLink(ctx context.Context, orgID, quotationID, userID int64, req *CreateQuotationPublicLinkRequest) (*QuotationPublicLink, error)
	ListQuotationPublicLinks(ctx context.Context, orgID, quotationID int64) ([]*QuotationPublicLink, error)
	RevokeQuotationPublicLink(ctx context.Context, orgID, quotationID, linkID, userID int64, req *RevokeQuotationPublicLinkRequest) error
	GetPublicQuotationByToken(ctx context.Context, token string, clientIP, userAgent string) (*QuotationPublicViewResponse, error)
	PublicAcceptQuotation(ctx context.Context, token string, clientIP, userAgent string, req *PublicAcceptQuotationRequest) error
	PublicDeclineQuotation(ctx context.Context, token string, clientIP, userAgent string, req *PublicDeclineQuotationRequest) error

	// Quotation-to-Booking Operational Conversion & Commercial Handover (Task 18.6)
	GetQuotationConversionPreview(ctx context.Context, orgID, quotationID int64) (*QuotationConversionPreview, error)
	ConvertQuotationToBooking(ctx context.Context, orgID, quotationID, userID int64, req *ConvertQuotationToBookingRequest) (*QuotationConversionResult, error)
	GetQuotationConversionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationConversionHistory, error)

	// Booking Confirmation, Commercial Handover & Lineage Traceability (Task 18.7)
	GetQuotationOperationalHandover(ctx context.Context, orgID, quotationID int64) (*QuotationOperationalHandover, error)
	ConfirmQuotationBookingHandover(ctx context.Context, orgID, quotationID, userID int64, req *ConfirmQuotationHandoverRequest) (*QuotationOperationalHandover, error)
	GetQuotationOperationalChanges(ctx context.Context, orgID, quotationID int64) ([]*OperationalChange, error)
	GetQuotationHandoverHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationOperationalHandoverHistory, error)

	// Quotation Analytics, Performance & Intelligence (Task 18.8)
	GetQuotationAnalyticsOverview(ctx context.Context, orgID int64) (*QuotationAnalyticsOverview, error)
	GetQuotationAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]*QuotationTrendDataPoint, error)
	GetCustomerQuotationPerformance(ctx context.Context, orgID int64) ([]*CustomerQuotationPerformance, error)
	GetQuotationPerformanceByMode(ctx context.Context, orgID int64) ([]*QuotationPerformanceByMode, error)
	GetQuotationExpiryRisk(ctx context.Context, orgID int64) ([]*QuotationRiskItem, error)

	// Rate-to-Quotation Integration & Commercial Selection (Task 19.5)
	GetQuotationRateCandidates(ctx context.Context, orgID, quotationID int64) (*QuotationRateCandidatesResponse, error)
	GetQuotationRateSelection(ctx context.Context, orgID, quotationID int64) (*QuotationRateSelection, error)
	SelectQuotationRate(ctx context.Context, req *SelectQuotationRateRequest) (*QuotationRateSnapshot, error)
	ReplaceQuotationRate(ctx context.Context, req *ReplaceQuotationRateRequest) (*QuotationRateSnapshot, error)
	RemoveQuotationRate(ctx context.Context, orgID, quotationID int64, user string) error
	GetQuotationRateSnapshot(ctx context.Context, orgID, quotationID int64) (*QuotationRateSnapshot, error)
	GetQuotationRateSelectionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateSelectionHistory, error)

	// Rate Lifecycle Intelligence & Commercial Risk Management (Task 19.6)
	GetQuotationRateRisks(ctx context.Context, orgID, quotationID int64) (*QuotationRateRiskSummary, error)
	ResolveQuotationRateRisk(ctx context.Context, orgID, quotationID, riskID int64, user string) error
	GetRateReplacementCandidates(ctx context.Context, orgID, quotationID int64) ([]*RateReplacementCandidate, error)
	GetCommercialImpactAnalysis(ctx context.Context, orgID, quotationID int64, replacementRateID, replacementSpotID *int64) (*CommercialImpactAnalysis, error)
	EvaluateQuotationRateRisksForOrg(ctx context.Context, orgID int64) (int, error)
}

var quotationNumberMu sync.Mutex

type service struct {
	repo    Repository
	rateSvc rates.Service
	docGen  DocumentGenerator
}

// NewService creates a new quotation service with repository and rate intelligence dependencies.
func NewService(repo Repository, rateSvc rates.Service) Service {
	return &service{
		repo:    repo,
		rateSvc: rateSvc,
		docGen:  NewDocumentGenerator("./storage/quotations"),
	}
}


func (s *service) CreateQuotation(ctx context.Context, orgID, userID int64, req *CreateQuotationRequest) (*Quotation, error) {
	actor := resolveActor(ctx)

	quotationNumber, err := s.generateQuotationNumber(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("generate quotation number: %w", err)
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	paymentTerms := req.PaymentTerms
	if paymentTerms == "" {
		paymentTerms = PaymentTermsPrepaid
	}

	q := &Quotation{
		OrgID:           orgID,
		QuotationNumber: quotationNumber,
		CustomerID:      req.CustomerID,
		RFQID:           req.RFQID,
		Status:          QuotationStatusDraft,
		Origin:          req.Origin,
		OriginCode:      req.OriginCode,
		Destination:     req.Destination,
		DestinationCode: req.DestinationCode,
		ServiceType:     req.ServiceType,
		TransportMode:   req.TransportMode,
		Currency:        currency,
		PaymentTerms:    paymentTerms,
		Notes:           req.Notes,
		CommercialTerms: req.CommercialTerms,
		CustomerNotes:   req.CustomerNotes,
		InternalNotes:   req.InternalNotes,
		TemplateID:      req.TemplateID,
		CreatedBy:       actor,
		UpdatedBy:       actor,
	}

	if req.CustomerID != nil && *req.CustomerID > 0 {
		if ci, err := s.repo.GetCustomerInfo(ctx, orgID, *req.CustomerID); err == nil {
			q.CustomerName = ci.Name
		}
	}

	if req.RFQID != nil {
		q.RFQNumber = fmt.Sprintf("RFQ-%d", *req.RFQID)
	}

	if req.ValidFrom != nil && *req.ValidFrom != "" {
		if t, err := time.Parse("2006-01-02", *req.ValidFrom); err == nil {
			q.ValidFrom = &t
		}
	}
	if req.ValidUntil != nil && *req.ValidUntil != "" {
		if t, err := time.Parse("2006-01-02", *req.ValidUntil); err == nil {
			q.ValidUntil = &t
		}
	}

	if err := s.repo.CreateQuotation(ctx, q); err != nil {
		return nil, fmt.Errorf("persist quotation: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  q.ID,
		ActivityType: QuotationCreated,
		Description:  fmt.Sprintf("Quotation %s created as DRAFT.", q.QuotationNumber),
		Actor:        actor,
	})

	// If a template_id was provided at creation, apply its snapshot charges
	if req.TemplateID != nil && *req.TemplateID > 0 {
		_, _ = s.ApplyTemplateToQuotation(ctx, orgID, q.ID, userID, &ApplyQuotationTemplateRequest{
			TemplateID:      *req.TemplateID,
			OverrideCharges: false,
		})
	}

	return q, nil
}

func (s *service) UpdateQuotation(ctx context.Context, orgID, quotationID, userID int64, req *UpdateQuotationRequest) (*Quotation, error) {
	existing, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[existing.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("quotation in status %s cannot be edited (must be DRAFT)", existing.Status))
	}

	actor := resolveActor(ctx)

	if req.CustomerID != nil {
		existing.CustomerID = req.CustomerID
		if *req.CustomerID > 0 {
			if ci, err := s.repo.GetCustomerInfo(ctx, orgID, *req.CustomerID); err == nil {
				existing.CustomerName = ci.Name
			}
		}
	}
	if req.Origin != nil {
		existing.Origin = *req.Origin
	}
	if req.OriginCode != nil {
		existing.OriginCode = *req.OriginCode
	}
	if req.Destination != nil {
		existing.Destination = *req.Destination
	}
	if req.DestinationCode != nil {
		existing.DestinationCode = *req.DestinationCode
	}
	if req.ServiceType != nil {
		existing.ServiceType = *req.ServiceType
	}
	if req.TransportMode != nil {
		existing.TransportMode = *req.TransportMode
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.PaymentTerms != nil {
		existing.PaymentTerms = *req.PaymentTerms
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}
	if req.CommercialTerms != nil {
		existing.CommercialTerms = *req.CommercialTerms
	}
	if req.CustomerNotes != nil {
		existing.CustomerNotes = *req.CustomerNotes
	}
	if req.InternalNotes != nil {
		existing.InternalNotes = *req.InternalNotes
	}
	if req.ValidFrom != nil {
		if *req.ValidFrom != "" {
			if t, err := time.Parse("2006-01-02", *req.ValidFrom); err == nil {
				existing.ValidFrom = &t
			}
		} else {
			existing.ValidFrom = nil
		}
	}
	if req.ValidUntil != nil {
		if *req.ValidUntil != "" {
			if t, err := time.Parse("2006-01-02", *req.ValidUntil); err == nil {
				existing.ValidUntil = &t
			}
		} else {
			existing.ValidUntil = nil
		}
	}

	existing.UpdatedBy = actor

	if err := s.repo.UpdateQuotation(ctx, existing); err != nil {
		return nil, err
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationUpdated,
		Description:  fmt.Sprintf("Quotation %s updated.", existing.QuotationNumber),
		Actor:        actor,
	})

	return existing, nil
}

func (s *service) GetQuotation(ctx context.Context, orgID, quotationID int64) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	// Auto-expire if past valid_until and still SENT/VIEWED
	if q.ValidUntil != nil && q.ValidUntil.Before(time.Now()) &&
		(q.Status == QuotationStatusSent || q.Status == QuotationStatusViewed) {
		now := time.Now()
		q.Status = QuotationStatusExpired
		q.ExpiredAt = &now
		_ = s.repo.UpdateQuotation(ctx, q)
	}

	detail := &QuotationDetail{
		Quotation: q,
	}

	// Pricing
	if pricing, err := s.GetQuotationPricing(ctx, orgID, quotationID); err == nil {
		detail.Pricing = pricing
	}

	// Commercial Terms
	commercial := &QuotationCommercialTerms{
		PaymentTerms:    q.PaymentTerms,
		CommercialTerms: q.CommercialTerms,
		CustomerNotes:   q.CustomerNotes,
		InternalNotes:   q.InternalNotes,
		ValidityStatus:  q.ValidityStatus,
	}
	if q.ValidFrom != nil {
		s := q.ValidFrom.Format("2006-01-02")
		commercial.ValidFrom = &s
	}
	if q.ValidUntil != nil {
		s := q.ValidUntil.Format("2006-01-02")
		commercial.ValidUntil = &s
	}
	detail.CommercialTerms = commercial

	// Customer Info
	if q.CustomerID != nil && *q.CustomerID > 0 {
		if ci, err := s.repo.GetCustomerInfo(ctx, orgID, *q.CustomerID); err == nil {
			detail.Customer = ci
		}
	}

	// Activity Timeline
	detail.Activity, _ = s.repo.GetActivity(ctx, orgID, quotationID)

	return detail, nil
}

func (s *service) ListQuotations(ctx context.Context, filters *QuotationListFilters) (*QuotationsListResponse, error) {
	quotations, total, err := s.repo.ListQuotations(ctx, filters)
	if err != nil {
		return nil, err
	}
	return &QuotationsListResponse{
		Quotations: quotations,
		Total:      total,
		Page:       filters.Page,
		Limit:      filters.Limit,
	}, nil
}

func (s *service) GetQuotationSummary(ctx context.Context, orgID int64) (*QuotationSummary, error) {
	return s.repo.GetQuotationSummary(ctx, orgID)
}

// ── Pricing & Charges Business Logic ─────────────────────────────────────────

func (s *service) GetQuotationPricing(ctx context.Context, orgID, quotationID int64) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	charges, err := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	summary, err := CalculateQuotationPricing(q, charges)
	if err != nil {
		return nil, fmt.Errorf("calculate pricing: %w", err)
	}

	return &QuotationPricing{
		QuotationID: quotationID,
		Currency:    summary.Currency,
		ChargeItems: charges,
		Summary:     summary,
	}, nil
}

func (s *service) AddQuotationCharge(ctx context.Context, orgID, quotationID, userID int64, req *CreateQuotationChargeRequest) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot add charges to quotation in status %s (must be DRAFT)", q.Status))
	}

	actor := resolveActor(ctx)

	if strings.TrimSpace(req.ChargeName) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("charge_name is required"))
	}
	if req.Quantity <= 0 {
		req.Quantity = 1.0
	}
	if req.UnitPrice < 0 {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("unit_price cannot be negative"))
	}
	if req.CostAmount < 0 {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("cost_amount cannot be negative"))
	}
	if req.TaxRate < 0 || req.TaxRate > 100 {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("tax_rate must be between 0 and 100"))
	}
	if req.DiscountValue < 0 {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("discount_value cannot be negative"))
	}

	currency := req.Currency
	if currency == "" {
		currency = q.Currency
	}
	if currency == "" {
		currency = "USD"
	}

	category := strings.ToUpper(req.ChargeCategory)
	if category == "" {
		category = QuotationChargeCategoryOther
	}

	basis := strings.ToUpper(req.CalculationBasis)
	if basis == "" {
		basis = QuotationChargeBasisFlat
	}

	chargeType := strings.ToUpper(req.ChargeType)
	if chargeType == "" {
		chargeType = QuotationChargeTypeSell
	}

	discountType := strings.ToUpper(req.DiscountType)
	if discountType == "" {
		discountType = QuotationDiscountTypeNone
	}

	chargeCode := strings.TrimSpace(req.ChargeCode)
	if chargeCode == "" {
		chargeCode = strings.ToUpper(category[:3])
	}

	displayOrder := 0
	if req.DisplayOrder != nil {
		displayOrder = *req.DisplayOrder
	} else {
		maxOrder, _ := s.repo.GetMaxDisplayOrder(ctx, orgID, quotationID)
		displayOrder = maxOrder + 1
	}

	item := &QuotationChargeItem{
		OrgID:            orgID,
		QuotationID:      quotationID,
		ChargeCode:       chargeCode,
		ChargeName:       req.ChargeName,
		ChargeCategory:   category,
		ChargeType:       chargeType,
		CalculationBasis: basis,
		Quantity:         req.Quantity,
		UnitPrice:        req.UnitPrice,
		CostAmount:       req.CostAmount,
		Currency:         currency,
		ExchangeRate:     1.0,
		TaxRate:          req.TaxRate,
		DiscountType:     discountType,
		DiscountValue:    req.DiscountValue,
		DisplayOrder:     displayOrder,
		IsOptional:       req.IsOptional,
		Notes:            req.Notes,
		CreatedBy:        actor,
		UpdatedBy:        actor,
	}

	_ = CalculateChargeItem(item)

	if err := s.repo.CreateQuotationCharge(ctx, item); err != nil {
		return nil, fmt.Errorf("create charge: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationChargeAdded,
		Description:  fmt.Sprintf("Added charge '%s' (%s %0.2f).", item.ChargeName, item.Currency, item.TotalSell),
		Actor:        actor,
	})

	return s.recalculate(ctx, orgID, quotationID)
}

func (s *service) UpdateQuotationCharge(ctx context.Context, orgID, quotationID, chargeID, userID int64, req *UpdateQuotationChargeRequest) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot edit charges for quotation in status %s (must be DRAFT)", q.Status))
	}

	existing, err := s.repo.GetQuotationChargeByID(ctx, orgID, quotationID, chargeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	actor := resolveActor(ctx)

	if req.ChargeName != nil && strings.TrimSpace(*req.ChargeName) != "" {
		existing.ChargeName = *req.ChargeName
	}
	if req.ChargeCode != nil {
		existing.ChargeCode = *req.ChargeCode
	}
	if req.ChargeCategory != nil {
		existing.ChargeCategory = strings.ToUpper(*req.ChargeCategory)
	}
	if req.ChargeType != nil {
		existing.ChargeType = strings.ToUpper(*req.ChargeType)
	}
	if req.CalculationBasis != nil {
		existing.CalculationBasis = strings.ToUpper(*req.CalculationBasis)
	}
	if req.Quantity != nil {
		if *req.Quantity <= 0 {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("quantity must be greater than 0"))
		}
		existing.Quantity = *req.Quantity
	}
	if req.UnitPrice != nil {
		if *req.UnitPrice < 0 {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("unit_price cannot be negative"))
		}
		existing.UnitPrice = *req.UnitPrice
	}
	if req.CostAmount != nil {
		if *req.CostAmount < 0 {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("cost_amount cannot be negative"))
		}
		existing.CostAmount = *req.CostAmount
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.TaxRate != nil {
		if *req.TaxRate < 0 || *req.TaxRate > 100 {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("tax_rate must be between 0 and 100"))
		}
		existing.TaxRate = *req.TaxRate
	}
	if req.DiscountType != nil {
		existing.DiscountType = strings.ToUpper(*req.DiscountType)
	}
	if req.DiscountValue != nil {
		if *req.DiscountValue < 0 {
			return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("discount_value cannot be negative"))
		}
		existing.DiscountValue = *req.DiscountValue
	}
	if req.DisplayOrder != nil {
		existing.DisplayOrder = *req.DisplayOrder
	}
	if req.IsOptional != nil {
		existing.IsOptional = *req.IsOptional
	}
	if req.Notes != nil {
		existing.Notes = *req.Notes
	}

	existing.UpdatedBy = actor
	_ = CalculateChargeItem(existing)

	if err := s.repo.UpdateQuotationCharge(ctx, existing); err != nil {
		return nil, fmt.Errorf("update charge: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationChargeUpdated,
		Description:  fmt.Sprintf("Updated charge '%s' (%s %0.2f).", existing.ChargeName, existing.Currency, existing.TotalSell),
		Actor:        actor,
	})

	return s.recalculate(ctx, orgID, quotationID)
}

func (s *service) DeleteQuotationCharge(ctx context.Context, orgID, quotationID, chargeID, userID int64) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot delete charges for quotation in status %s (must be DRAFT)", q.Status))
	}

	existing, err := s.repo.GetQuotationChargeByID(ctx, orgID, quotationID, chargeID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	actor := resolveActor(ctx)

	if err := s.repo.DeleteQuotationCharge(ctx, orgID, quotationID, chargeID); err != nil {
		return nil, fmt.Errorf("delete charge: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationChargeRemoved,
		Description:  fmt.Sprintf("Removed charge '%s'.", existing.ChargeName),
		Actor:        actor,
	})

	return s.recalculate(ctx, orgID, quotationID)
}

func (s *service) ReorderQuotationCharges(ctx context.Context, orgID, quotationID int64, req *ReorderQuotationChargesRequest) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot reorder charges for quotation in status %s", q.Status))
	}

	if len(req.ChargeIDs) > 0 {
		if err := s.repo.ReorderQuotationCharges(ctx, orgID, quotationID, req.ChargeIDs); err != nil {
			return nil, err
		}
	}

	return s.GetQuotationPricing(ctx, orgID, quotationID)
}

func (s *service) GetRateCandidates(ctx context.Context, orgID, quotationID int64) ([]RateCandidate, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if s.rateSvc == nil {
		return []RateCandidate{}, nil
	}

	originPort := q.OriginCode
	if originPort == "" {
		originPort = q.Origin
	}
	destPort := q.DestinationCode
	if destPort == "" {
		destPort = q.Destination
	}

	if originPort == "" || destPort == "" {
		return []RateCandidate{}, nil
	}

	searchRes, err := s.rateSvc.SearchRates(ctx, rates.RateQuery{
		OrgID:           orgID,
		OriginPort:      originPort,
		DestinationPort: destPort,
		EquipmentType:   "40GP",
		MaxResults:      10,
	})
	if err != nil {
		return []RateCandidate{}, nil
	}

	candidates := make([]RateCandidate, 0, len(searchRes.Rates))
	for _, r := range searchRes.Rates {
		candSurcharges := make([]RateCandidateSurcharge, 0, len(r.Surcharges))
		for _, sc := range r.Surcharges {
			candSurcharges = append(candSurcharges, RateCandidateSurcharge{
				Code:        sc.Code,
				Description: sc.Description,
				Amount:      sc.Amount,
				Unit:        string(sc.Unit),
				Included:    sc.Included,
			})
		}

		candidates = append(candidates, RateCandidate{
			ID:                 r.ID,
			Source:             string(r.Source),
			CarrierSCAC:        r.CarrierSCAC,
			CarrierName:        r.CarrierName,
			OriginPort:         r.OriginPort,
			DestinationPort:    r.DestinationPort,
			EquipmentType:      r.EquipmentType,
			OceanFreight:       r.OceanFreight,
			OriginCharges:      r.OriginCharges,
			DestinationCharges: r.DestinationCharges,
			TotalBuyPrice:      r.TotalBuyPrice,
			Currency:           "USD",
			TransitDays:        r.TransitDays,
			ConfidenceScore:    r.ConfidenceScore,
			Surcharges:         candSurcharges,
		})
	}

	return candidates, nil
}

func (s *service) ImportRateCharges(ctx context.Context, orgID, quotationID, userID int64, req *ImportRateChargesRequest) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot import rate into quotation in status %s (must be DRAFT)", q.Status))
	}

	if s.rateSvc == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, fmt.Errorf("rate service is not available"))
	}

	rate, err := s.rateSvc.GetRateByID(ctx, orgID, req.RateID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("rate not found: %w", err))
	}

	actor := resolveActor(ctx)
	displayOrder, _ := s.repo.GetMaxDisplayOrder(ctx, orgID, quotationID)

	if rate.OceanFreight > 0 {
		displayOrder++
		freightItem := &QuotationChargeItem{
			OrgID:            orgID,
			QuotationID:      quotationID,
			ChargeCode:       "BAS",
			ChargeName:       fmt.Sprintf("Ocean Freight (%s - %s)", rate.CarrierName, rate.EquipmentType),
			ChargeCategory:   QuotationChargeCategoryFreight,
			ChargeType:       QuotationChargeTypeSell,
			CalculationBasis: QuotationChargeBasisPerContainer,
			Quantity:         1.0,
			UnitPrice:        round2(rate.OceanFreight * 1.15),
			CostAmount:       rate.OceanFreight,
			Currency:         "USD",
			ExchangeRate:     1.0,
			TaxRate:          0.0,
			DiscountType:     QuotationDiscountTypeNone,
			DisplayOrder:     displayOrder,
			Notes:            fmt.Sprintf("Imported from Rate #%s (%s)", rate.ID, rate.Source),
			CreatedBy:        actor,
			UpdatedBy:        actor,
		}
		_ = CalculateChargeItem(freightItem)
		_ = s.repo.CreateQuotationCharge(ctx, freightItem)
	}

	if rate.OriginCharges > 0 {
		displayOrder++
		originItem := &QuotationChargeItem{
			OrgID:            orgID,
			QuotationID:      quotationID,
			ChargeCode:       "ORG",
			ChargeName:       fmt.Sprintf("Origin Terminal / Handling Charges (%s)", rate.OriginPort),
			ChargeCategory:   QuotationChargeCategoryOrigin,
			ChargeType:       QuotationChargeTypeSell,
			CalculationBasis: QuotationChargeBasisPerContainer,
			Quantity:         1.0,
			UnitPrice:        round2(rate.OriginCharges * 1.10),
			CostAmount:       rate.OriginCharges,
			Currency:         "USD",
			ExchangeRate:     1.0,
			DisplayOrder:     displayOrder,
			CreatedBy:        actor,
			UpdatedBy:        actor,
		}
		_ = CalculateChargeItem(originItem)
		_ = s.repo.CreateQuotationCharge(ctx, originItem)
	}

	if rate.DestinationCharges > 0 {
		displayOrder++
		destItem := &QuotationChargeItem{
			OrgID:            orgID,
			QuotationID:      quotationID,
			ChargeCode:       "DST",
			ChargeName:       fmt.Sprintf("Destination Terminal / Handling Charges (%s)", rate.DestinationPort),
			ChargeCategory:   QuotationChargeCategoryDestination,
			ChargeType:       QuotationChargeTypeSell,
			CalculationBasis: QuotationChargeBasisPerContainer,
			Quantity:         1.0,
			UnitPrice:        round2(rate.DestinationCharges * 1.10),
			CostAmount:       rate.DestinationCharges,
			Currency:         "USD",
			ExchangeRate:     1.0,
			DisplayOrder:     displayOrder,
			CreatedBy:        actor,
			UpdatedBy:        actor,
		}
		_ = CalculateChargeItem(destItem)
		_ = s.repo.CreateQuotationCharge(ctx, destItem)
	}

	for _, sc := range rate.Surcharges {
		if sc.Included {
			continue
		}
		displayOrder++
		basis := QuotationChargeBasisPerContainer
		if sc.Unit == rates.SurchargeUnitPerShipment {
			basis = QuotationChargeBasisPerShipment
		}
		scItem := &QuotationChargeItem{
			OrgID:            orgID,
			QuotationID:      quotationID,
			ChargeCode:       sc.Code,
			ChargeName:       sc.Description,
			ChargeCategory:   QuotationChargeCategorySurcharge,
			ChargeType:       QuotationChargeTypeSell,
			CalculationBasis: basis,
			Quantity:         1.0,
			UnitPrice:        sc.Amount,
			CostAmount:       sc.Amount,
			Currency:         "USD",
			ExchangeRate:     1.0,
			DisplayOrder:     displayOrder,
			CreatedBy:        actor,
			UpdatedBy:        actor,
		}
		_ = CalculateChargeItem(scItem)
		_ = s.repo.CreateQuotationCharge(ctx, scItem)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationRateImported,
		Description:  fmt.Sprintf("Imported rate components from %s (%s).", rate.CarrierName, rate.ID),
		Actor:        actor,
	})

	return s.recalculate(ctx, orgID, quotationID)
}

// ── Reusable Quotation Templates & Commercial Terms (Task 18.3) ───────────────

func (s *service) ListQuotationTemplates(ctx context.Context, orgID int64, activeOnly bool) ([]*QuotationTemplate, error) {
	// Auto seed default templates if organization has 0 templates
	_ = s.repo.SeedDefaultTemplatesIfEmpty(ctx, orgID)
	return s.repo.GetQuotationTemplates(ctx, orgID, activeOnly)
}

func (s *service) GetQuotationTemplate(ctx context.Context, orgID, templateID int64) (*QuotationTemplateDetail, error) {
	tmpl, charges, err := s.repo.GetQuotationTemplateByID(ctx, orgID, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}
	return &QuotationTemplateDetail{
		Template: tmpl,
		Charges:  charges,
	}, nil
}

func (s *service) CreateQuotationTemplate(ctx context.Context, orgID, userID int64, req *CreateQuotationTemplateRequest) (*QuotationTemplate, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("template name is required"))
	}

	actor := resolveActor(ctx)
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	paymentTerms := req.PaymentTerms
	if paymentTerms == "" {
		paymentTerms = PaymentTermsPrepaid
	}
	validityDays := req.ValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}

	tmpl := &QuotationTemplate{
		OrgID:           orgID,
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		ShipmentMode:    req.ShipmentMode,
		TransportMode:   req.TransportMode,
		Origin:          req.Origin,
		Destination:     req.Destination,
		Currency:        currency,
		ValidityDays:    validityDays,
		PaymentTerms:    paymentTerms,
		CommercialTerms: req.CommercialTerms,
		CustomerNotes:   req.CustomerNotes,
		InternalNotes:   req.InternalNotes,
		IsActive:        true,
		CreatedBy:       actor,
	}

	charges := make([]*QuotationTemplateChargeItem, 0, len(req.Charges))
	for idx, c := range req.Charges {
		order := c.DisplayOrder
		if order <= 0 {
			order = idx + 1
		}
		charges = append(charges, &QuotationTemplateChargeItem{
			OrgID:            orgID,
			ChargeCategory:   c.ChargeCategory,
			ChargeCode:       c.ChargeCode,
			ChargeName:       c.ChargeName,
			CalculationBasis: c.CalculationBasis,
			Quantity:         c.Quantity,
			UnitPrice:        c.UnitPrice,
			CostAmount:       c.CostAmount,
			DiscountType:     c.DiscountType,
			DiscountValue:    c.DiscountValue,
			TaxRate:          c.TaxRate,
			Currency:         c.Currency,
			DisplayOrder:     order,
			IsOptional:       c.IsOptional,
			Notes:            c.Notes,
		})
	}

	if err := s.repo.CreateQuotationTemplate(ctx, tmpl, charges); err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}

	tmpl.ChargeCount = len(charges)
	return tmpl, nil
}

func (s *service) UpdateQuotationTemplate(ctx context.Context, orgID, templateID, userID int64, req *UpdateQuotationTemplateRequest) (*QuotationTemplate, error) {
	existing, _, err := s.repo.GetQuotationTemplateByID(ctx, orgID, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Description != nil {
		existing.Description = *req.Description
	}
	if req.ShipmentMode != nil {
		existing.ShipmentMode = *req.ShipmentMode
	}
	if req.TransportMode != nil {
		existing.TransportMode = *req.TransportMode
	}
	if req.Origin != nil {
		existing.Origin = *req.Origin
	}
	if req.Destination != nil {
		existing.Destination = *req.Destination
	}
	if req.Currency != nil {
		existing.Currency = *req.Currency
	}
	if req.ValidityDays != nil && *req.ValidityDays > 0 {
		existing.ValidityDays = *req.ValidityDays
	}
	if req.PaymentTerms != nil {
		existing.PaymentTerms = *req.PaymentTerms
	}
	if req.CommercialTerms != nil {
		existing.CommercialTerms = *req.CommercialTerms
	}
	if req.CustomerNotes != nil {
		existing.CustomerNotes = *req.CustomerNotes
	}
	if req.InternalNotes != nil {
		existing.InternalNotes = *req.InternalNotes
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	var charges []*QuotationTemplateChargeItem
	if req.Charges != nil {
		charges = make([]*QuotationTemplateChargeItem, 0, len(req.Charges))
		for idx, c := range req.Charges {
			order := c.DisplayOrder
			if order <= 0 {
				order = idx + 1
			}
			charges = append(charges, &QuotationTemplateChargeItem{
				OrgID:            orgID,
				TemplateID:       templateID,
				ChargeCategory:   c.ChargeCategory,
				ChargeCode:       c.ChargeCode,
				ChargeName:       c.ChargeName,
				CalculationBasis: c.CalculationBasis,
				Quantity:         c.Quantity,
				UnitPrice:        c.UnitPrice,
				CostAmount:       c.CostAmount,
				DiscountType:     c.DiscountType,
				DiscountValue:    c.DiscountValue,
				TaxRate:          c.TaxRate,
				Currency:         c.Currency,
				DisplayOrder:     order,
				IsOptional:       c.IsOptional,
				Notes:            c.Notes,
			})
		}
	}

	if err := s.repo.UpdateQuotationTemplate(ctx, existing, charges); err != nil {
		return nil, fmt.Errorf("update template: %w", err)
	}

	if charges != nil {
		existing.ChargeCount = len(charges)
	}

	return existing, nil
}

func (s *service) DeleteQuotationTemplate(ctx context.Context, orgID, templateID int64) error {
	_, _, err := s.repo.GetQuotationTemplateByID(ctx, orgID, templateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return err
	}
	return s.repo.ArchiveQuotationTemplate(ctx, orgID, templateID)
}

func (s *service) CreateTemplateFromQuotation(ctx context.Context, orgID, quotationID, userID int64, req *CreateTemplateFromQuotationRequest) (*QuotationTemplate, error) {
	if strings.TrimSpace(req.Name) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("template name is required"))
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	charges, err := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	actor := resolveActor(ctx)
	validityDays := req.ValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}

	tmpl := &QuotationTemplate{
		OrgID:           orgID,
		Name:            strings.TrimSpace(req.Name),
		Description:     req.Description,
		ShipmentMode:    q.ServiceType,
		TransportMode:   q.TransportMode,
		Origin:          q.Origin,
		Destination:     q.Destination,
		Currency:        q.Currency,
		ValidityDays:    validityDays,
		PaymentTerms:    q.PaymentTerms,
		CommercialTerms: q.CommercialTerms,
		CustomerNotes:   q.CustomerNotes,
		InternalNotes:   q.InternalNotes,
		IsActive:        true,
		CreatedBy:       actor,
	}

	tmplCharges := make([]*QuotationTemplateChargeItem, 0, len(charges))
	for idx, c := range charges {
		tmplCharges = append(tmplCharges, &QuotationTemplateChargeItem{
			OrgID:            orgID,
			ChargeCategory:   c.ChargeCategory,
			ChargeCode:       c.ChargeCode,
			ChargeName:       c.ChargeName,
			CalculationBasis: c.CalculationBasis,
			Quantity:         c.Quantity,
			UnitPrice:        c.UnitPrice,
			CostAmount:       c.CostAmount,
			DiscountType:     c.DiscountType,
			DiscountValue:    c.DiscountValue,
			TaxRate:          c.TaxRate,
			Currency:         c.Currency,
			DisplayOrder:     idx + 1,
			IsOptional:       c.IsOptional,
			Notes:            c.Notes,
		})
	}

	if err := s.repo.CreateQuotationTemplate(ctx, tmpl, tmplCharges); err != nil {
		return nil, fmt.Errorf("create template from quotation: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationTemplateCreated,
		Description:  fmt.Sprintf("Saved quotation setup as reusable template '%s'.", tmpl.Name),
		Actor:        actor,
	})

	tmpl.ChargeCount = len(tmplCharges)
	return tmpl, nil
}

func (s *service) ApplyTemplateToQuotation(ctx context.Context, orgID, quotationID, userID int64, req *ApplyQuotationTemplateRequest) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot apply template to quotation in status %s (must be DRAFT)", q.Status))
	}

	tmpl, charges, err := s.repo.GetQuotationTemplateByID(ctx, orgID, req.TemplateID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	actor := resolveActor(ctx)

	// If override_charges is requested, clear existing quotation charge items
	var startOrder int
	if req.OverrideCharges {
		_ = s.repo.ClearQuotationCharges(ctx, orgID, quotationID)
		startOrder = 0
	} else {
		startOrder, _ = s.repo.GetMaxDisplayOrder(ctx, orgID, quotationID)
	}

	// 1. Copy snapshot charge items into quotation_charge_items
	for _, tc := range charges {
		startOrder++
		qItem := &QuotationChargeItem{
			OrgID:            orgID,
			QuotationID:      quotationID,
			ChargeCode:       tc.ChargeCode,
			ChargeName:       tc.ChargeName,
			ChargeCategory:   tc.ChargeCategory,
			ChargeType:       QuotationChargeTypeSell,
			CalculationBasis: tc.CalculationBasis,
			Quantity:         tc.Quantity,
			UnitPrice:        tc.UnitPrice,
			CostAmount:       tc.CostAmount,
			Currency:         tc.Currency,
			ExchangeRate:     1.0,
			TaxRate:          tc.TaxRate,
			DiscountType:     tc.DiscountType,
			DiscountValue:    tc.DiscountValue,
			DisplayOrder:     startOrder,
			IsOptional:       tc.IsOptional,
			Notes:            tc.Notes,
			CreatedBy:        actor,
			UpdatedBy:        actor,
		}
		_ = CalculateChargeItem(qItem)
		_ = s.repo.CreateQuotationCharge(ctx, qItem)
	}

	// 2. Update commercial terms and validity on quotation
	now := time.Now()
	validFromStr := now.Format("2006-01-02")
	validityDays := tmpl.ValidityDays
	if validityDays <= 0 {
		validityDays = 30
	}
	validUntilStr := now.AddDate(0, 0, validityDays).Format("2006-01-02")

	terms := &QuotationCommercialTerms{
		PaymentTerms:    tmpl.PaymentTerms,
		CommercialTerms: tmpl.CommercialTerms,
		CustomerNotes:   tmpl.CustomerNotes,
		InternalNotes:   tmpl.InternalNotes,
		ValidFrom:       &validFromStr,
		ValidUntil:      &validUntilStr,
	}
	_ = s.repo.UpdateQuotationCommercialTerms(ctx, orgID, quotationID, terms)

	// Set template lineage on quotation
	q.TemplateID = &tmpl.ID
	q.PaymentTerms = tmpl.PaymentTerms
	q.CommercialTerms = tmpl.CommercialTerms
	q.CustomerNotes = tmpl.CustomerNotes
	q.InternalNotes = tmpl.InternalNotes
	_ = s.repo.UpdateQuotation(ctx, q)

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationTemplateApplied,
		Description:  fmt.Sprintf("Applied template '%s' (%d charges snapshot copied).", tmpl.Name, len(charges)),
		Actor:        actor,
	})

	// 3. Recalculate quotation pricing using Task 18.2 pricing engine
	return s.recalculate(ctx, orgID, quotationID)
}

func (s *service) UpdateQuotationCommercialTerms(ctx context.Context, orgID, quotationID, userID int64, req *UpdateQuotationCommercialTermsRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !EditableStatuses[q.Status] {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot update commercial terms on quotation in status %s (must be DRAFT)", q.Status))
	}

	actor := resolveActor(ctx)

	paymentTerms := q.PaymentTerms
	if req.PaymentTerms != nil && *req.PaymentTerms != "" {
		paymentTerms = *req.PaymentTerms
	}

	commercialTerms := q.CommercialTerms
	if req.CommercialTerms != nil {
		commercialTerms = *req.CommercialTerms
	}

	customerNotes := q.CustomerNotes
	if req.CustomerNotes != nil {
		customerNotes = *req.CustomerNotes
	}

	internalNotes := q.InternalNotes
	if req.InternalNotes != nil {
		internalNotes = *req.InternalNotes
	}

	validFromStr := ""
	if req.ValidFrom != nil {
		validFromStr = *req.ValidFrom
	} else if q.ValidFrom != nil {
		validFromStr = q.ValidFrom.Format("2006-01-02")
	}

	validUntilStr := ""
	if req.ValidUntil != nil {
		validUntilStr = *req.ValidUntil
	} else if q.ValidUntil != nil {
		validUntilStr = q.ValidUntil.Format("2006-01-02")
	}

	terms := &QuotationCommercialTerms{
		PaymentTerms:    paymentTerms,
		CommercialTerms: commercialTerms,
		CustomerNotes:   customerNotes,
		InternalNotes:   internalNotes,
		ValidFrom:       &validFromStr,
		ValidUntil:      &validUntilStr,
	}

	if err := s.repo.UpdateQuotationCommercialTerms(ctx, orgID, quotationID, terms); err != nil {
		return nil, fmt.Errorf("update commercial terms: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationCommercialTermsUpdated,
		Description:  "Commercial terms, payment terms, and validity rules updated.",
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// recalculate loads all quotation charges, runs the centralized calculation engine,
// updates the quotation row in DB with authoritative summary values, and returns QuotationPricing.
func (s *service) recalculate(ctx context.Context, orgID, quotationID int64) (*QuotationPricing, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	charges, err := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	summary, err := CalculateQuotationPricing(q, charges)
	if err != nil {
		return nil, fmt.Errorf("calculate pricing summary: %w", err)
	}

	if err := s.repo.UpdateQuotationTotals(
		ctx, orgID, quotationID,
		summary.Subtotal, summary.Surcharges, summary.Taxes,
		summary.TotalAmount, summary.TotalCost, summary.GrossProfit, summary.GrossMarginPercentage,
	); err != nil {
		return nil, fmt.Errorf("persist quotation totals: %w", err)
	}

	return &QuotationPricing{
		QuotationID: quotationID,
		Currency:    summary.Currency,
		ChargeItems: charges,
		Summary:     summary,
	}, nil
}

func (s *service) generateQuotationNumber(ctx context.Context, orgID int64) (string, error) {
	quotationNumberMu.Lock()
	defer quotationNumberMu.Unlock()

	year := time.Now().Year()
	last, err := s.repo.GetLastQuotationNumber(ctx, orgID, year)
	if err != nil {
		last = 0
	}
	next := last + 1
	return fmt.Sprintf("QT-%d-%04d", year, next), nil
}

func resolveActor(ctx context.Context) string {
	if uc, ok := middleware.GetUserContext(ctx); ok && uc.UserID > 0 {
		if uc.Role != "" {
			return fmt.Sprintf("User #%d (%s)", uc.UserID, uc.Role)
		}
		return fmt.Sprintf("User #%d", uc.UserID)
	}
	return "System"
}

func resolveActorUserID(ctx context.Context, fallbackUserID int64) *int64 {
	if uc, ok := middleware.GetUserContext(ctx); ok && uc.UserID > 0 {
		return &uc.UserID
	}
	if fallbackUserID > 0 {
		return &fallbackUserID
	}
	return nil
}

func isApprovalRequired(ctx context.Context, orgID int64) bool {
	// In the future, this can query org-level policy settings.
	return true
}

// SubmitQuotationForReview transitions a draft quotation to READY_FOR_REVIEW.
func (s *service) SubmitQuotationForReview(ctx context.Context, orgID, quotationID, userID int64, req *SubmitQuotationForReviewRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusReadyForReview, isApprovalRequired(ctx, orgID)); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	actorUserID := resolveActorUserID(ctx, userID)

	comments := ""
	if req != nil {
		comments = req.Comments
	}

	if err := s.repo.SubmitQuotationForReview(ctx, orgID, quotationID, actor, comments); err != nil {
		return nil, fmt.Errorf("submit quotation for review: %w", err)
	}

	// Record approval history
	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionSubmitted,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusReadyForReview,
		ActorUserID:    actorUserID,
		ActorName:      actor,
		Comments:       comments,
	})

	// Record timeline activity
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationSubmittedForReview,
		Description:  fmt.Sprintf("Quotation %s submitted for management review", q.QuotationNumber),
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// ApproveQuotation approves a submitted quotation.
func (s *service) ApproveQuotation(ctx context.Context, orgID, quotationID, userID int64, req *ApproveQuotationRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusApproved, isApprovalRequired(ctx, orgID)); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	actorUserID := resolveActorUserID(ctx, userID)

	approvalNotes := ""
	if req != nil {
		approvalNotes = req.ApprovalNotes
	}

	if err := s.repo.ApproveQuotation(ctx, orgID, quotationID, actor, approvalNotes); err != nil {
		return nil, fmt.Errorf("approve quotation: %w", err)
	}

	// Record approval history
	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionApproved,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusApproved,
		ActorUserID:    actorUserID,
		ActorName:      actor,
		Comments:       approvalNotes,
	})

	// Record timeline activity
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationApproved,
		Description:  fmt.Sprintf("Quotation %s approved by management", q.QuotationNumber),
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// RequestQuotationChanges requests modifications on a submitted quotation.
func (s *service) RequestQuotationChanges(ctx context.Context, orgID, quotationID, userID int64, req *RequestQuotationChangesRequest) (*QuotationDetail, error) {
	if req == nil || strings.TrimSpace(req.Reason) == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("reason is required when requesting changes"))
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusChangesRequested, isApprovalRequired(ctx, orgID)); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	actorUserID := resolveActorUserID(ctx, userID)

	if err := s.repo.RequestQuotationChanges(ctx, orgID, quotationID, actor, req.Reason); err != nil {
		return nil, fmt.Errorf("request quotation changes: %w", err)
	}

	// Record approval history
	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionChangesRequested,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusChangesRequested,
		ActorUserID:    actorUserID,
		ActorName:      actor,
		Comments:       req.Reason,
	})

	// Record timeline activity
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationChangesRequested,
		Description:  fmt.Sprintf("Changes requested on quotation %s: %s", q.QuotationNumber, req.Reason),
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// SendQuotation marks the quotation as sent to the customer and locks commercial edits.
func (s *service) SendQuotation(ctx context.Context, orgID, quotationID, userID int64, req *SendQuotationRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusSent, isApprovalRequired(ctx, orgID)); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	actorUserID := resolveActorUserID(ctx, userID)

	comments := ""
	if req != nil {
		comments = req.Comments
	}

	if err := s.repo.MarkQuotationSent(ctx, orgID, quotationID, actor); err != nil {
		return nil, fmt.Errorf("mark quotation sent: %w", err)
	}

	// Record approval history
	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionSent,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusSent,
		ActorUserID:    actorUserID,
		ActorName:      actor,
		Comments:       comments,
	})

	// Record timeline activity
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationSent,
		Description:  fmt.Sprintf("Quotation %s marked as sent to customer", q.QuotationNumber),
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// GetCustomerQuotationPreview builds the dedicated, safe customer view model.
// Internal costs, profits, gross margin pct, internal notes, rate candidate details are strictly omitted.
func (s *service) GetCustomerQuotationPreview(ctx context.Context, orgID, quotationID int64) (*CustomerQuotationPreview, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	charges, err := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("get charges: %w", err)
	}

	var customerCharges []CustomerQuotationChargeItem
	var discountTotal float64
	var taxTotal float64

	for _, c := range charges {
		discountTotal += c.DiscountAmount
		taxTotal += c.TaxAmount

		customerCharges = append(customerCharges, CustomerQuotationChargeItem{
			ID:               c.ID,
			ChargeCode:       c.ChargeCode,
			ChargeName:       c.ChargeName,
			ChargeCategory:   c.ChargeCategory,
			CalculationBasis: c.CalculationBasis,
			Quantity:         c.Quantity,
			UnitPrice:        c.UnitPrice,
			DiscountType:     c.DiscountType,
			DiscountValue:    c.DiscountValue,
			DiscountAmount:   c.DiscountAmount,
			TaxRate:          c.TaxRate,
			TaxAmount:        c.TaxAmount,
			FinalAmount:      c.SellAmount,
			Currency:         c.Currency,
			IsOptional:       c.IsOptional,
			SortOrder:        c.DisplayOrder,
		})
	}

	validityStatus := CalculateQuotationValidityStatus(q.ValidUntil)

	preview := &CustomerQuotationPreview{
		QuotationID:     q.ID,
		QuotationNumber: q.QuotationNumber,
		Status:          q.Status,
		CustomerID:      q.CustomerID,
		CustomerName:    q.CustomerName,
		Origin:          q.Origin,
		OriginCode:      q.OriginCode,
		Destination:     q.Destination,
		DestinationCode: q.DestinationCode,
		ServiceType:     q.ServiceType,
		TransportMode:   q.TransportMode,
		Currency:        q.Currency,
		PaymentTerms:    q.PaymentTerms,
		ValidFrom:       q.ValidFrom,
		ValidUntil:      q.ValidUntil,
		ValidityStatus:  validityStatus,
		Subtotal:        q.Subtotal,
		DiscountTotal:   discountTotal,
		TaxTotal:        taxTotal,
		TotalAmount:     q.TotalAmount,
		CommercialTerms: q.CommercialTerms,
		CustomerNotes:   q.CustomerNotes,
		SentAt:          q.SentAt,
		ViewedAt:        q.ViewedAt,
		AcceptedAt:      q.AcceptedAt,
		DeclinedAt:      q.DeclinedAt,
		Charges:         customerCharges,
		CompanyName:     "LogisticsHQ Global Freight",
		CompanyAddress:  "742 Evergreen Terrace, Logistics Park, Suite 400",
		CompanyContact:  "support@logisticshq.io • +1 (800) 555-LOGI",
	}

	return preview, nil
}

// GetQuotationApprovalHistory retrieves the full approval and lifecycle event log.
func (s *service) GetQuotationApprovalHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationApprovalHistory, error) {
	_, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	return s.repo.GetApprovalHistory(ctx, orgID, quotationID)
}

// GetQuotationApprovalStatus returns the capability flags for the quotation.
func (s *service) GetQuotationApprovalStatus(ctx context.Context, orgID, quotationID int64) (*QuotationApprovalStatus, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	approvalReq := isApprovalRequired(ctx, orgID)
	canEdit := IsQuotationCommerciallyEditable(q.Status)

	status := &QuotationApprovalStatus{
		Status:             q.Status,
		ApprovalRequired:   approvalReq,
		CanSubmitForReview: q.Status == QuotationStatusDraft || q.Status == QuotationStatusChangesRequested,
		CanApprove:         q.Status == QuotationStatusReadyForReview,
		CanRequestChanges:  q.Status == QuotationStatusReadyForReview,
		CanSend:            (q.Status == QuotationStatusApproved) || (!approvalReq && q.Status == QuotationStatusDraft),
		CanAccept:          q.Status == QuotationStatusSent || q.Status == QuotationStatusViewed,
		CanDecline:         q.Status == QuotationStatusSent || q.Status == QuotationStatusViewed,
		CanCancel:          q.Status != QuotationStatusAccepted && q.Status != QuotationStatusDeclined && q.Status != QuotationStatusCancelled && q.Status != QuotationStatusExpired,
		CanEdit:            canEdit,
	}

	return status, nil
}

// MarkQuotationViewed records a view event on the quotation.
func (s *service) MarkQuotationViewed(ctx context.Context, orgID, quotationID int64, req *MarkQuotationViewedRequest) error {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return fmt.Errorf("get quotation: %w", err)
	}

	var viewerName, viewerEmail, ipAddress, userAgent string
	if req != nil {
		viewerName = req.ViewerName
		viewerEmail = req.ViewerEmail
		ipAddress = req.IPAddress
		userAgent = req.UserAgent
	}

	if err := s.repo.MarkQuotationViewed(ctx, orgID, quotationID, viewerName, viewerEmail, ipAddress, userAgent); err != nil {
		return fmt.Errorf("mark viewed: %w", err)
	}

	_ = s.repo.RecordPublicView(ctx, orgID, quotationID, viewerName, viewerEmail, ipAddress, userAgent)

	if q.Status == QuotationStatusSent {
		_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
			OrgID:          orgID,
			QuotationID:    quotationID,
			Action:         QuotationApprovalActionViewed,
			PreviousStatus: QuotationStatusSent,
			NewStatus:      QuotationStatusViewed,
			ActorName:      viewerName,
			Comments:       fmt.Sprintf("Viewed from IP %s", ipAddress),
		})

		_ = s.repo.CreateActivity(ctx, &QuotationActivity{
			OrgID:        orgID,
			QuotationID:  quotationID,
			ActivityType: QuotationViewed,
			Description:  fmt.Sprintf("Quotation %s viewed by customer", q.QuotationNumber),
			Actor:        viewerName,
		})
	}

	return nil
}

// AcceptQuotation records customer acceptance.
func (s *service) AcceptQuotation(ctx context.Context, orgID, quotationID, userID int64, req *AcceptQuotationRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusAccepted, false); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	acceptedBy := actor
	comments := ""
	if req != nil {
		if req.AcceptedBy != "" {
			acceptedBy = req.AcceptedBy
		}
		comments = req.Comments
	}

	if err := s.repo.AcceptQuotation(ctx, orgID, quotationID, acceptedBy, comments); err != nil {
		return nil, fmt.Errorf("accept quotation: %w", err)
	}

	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionAccepted,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusAccepted,
		ActorName:      acceptedBy,
		Comments:       comments,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationAccepted,
		Description:  fmt.Sprintf("Quotation %s accepted by customer (%s)", q.QuotationNumber, acceptedBy),
		Actor:        acceptedBy,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// DeclineQuotation records customer decline with a mandatory or recommended reason.
func (s *service) DeclineQuotation(ctx context.Context, orgID, quotationID, userID int64, req *DeclineQuotationRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusDeclined, false); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	declinedBy := actor
	reason := "Commercial terms not accepted"
	if req != nil {
		if req.DeclinedBy != "" {
			declinedBy = req.DeclinedBy
		}
		if req.Reason != "" {
			reason = req.Reason
		}
	}

	if err := s.repo.DeclineQuotation(ctx, orgID, quotationID, declinedBy, reason); err != nil {
		return nil, fmt.Errorf("decline quotation: %w", err)
	}

	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionDeclined,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusDeclined,
		ActorName:      declinedBy,
		Comments:       reason,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationDeclined,
		Description:  fmt.Sprintf("Quotation %s declined by customer: %s", q.QuotationNumber, reason),
		Actor:        declinedBy,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// CancelQuotation cancels an active quotation.
func (s *service) CancelQuotation(ctx context.Context, orgID, quotationID, userID int64, req *CancelQuotationRequest) (*QuotationDetail, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("get quotation: %w", err)
	}

	if err := CanTransitionQuotationStatus(q.Status, QuotationStatusCancelled, false); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, err)
	}

	actor := resolveActor(ctx)
	reason := "Cancelled by operational administrator"
	if req != nil && req.Reason != "" {
		reason = req.Reason
	}

	if err := s.repo.CancelQuotation(ctx, orgID, quotationID, actor, reason); err != nil {
		return nil, fmt.Errorf("cancel quotation: %w", err)
	}

	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          orgID,
		QuotationID:    quotationID,
		Action:         QuotationApprovalActionCancelled,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusCancelled,
		ActorName:      actor,
		Comments:       reason,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationCancelled,
		Description:  fmt.Sprintf("Quotation %s cancelled: %s", q.QuotationNumber, reason),
		Actor:        actor,
	})

	return s.GetQuotation(ctx, orgID, quotationID)
}

// ── Document Generation & Retrieval (Task 18.5) ─────────────────────────────

func (s *service) GenerateQuotationDocument(ctx context.Context, orgID, quotationID, userID int64, docType string) (*QuotationDocument, error) {
	if docType == "" {
		docType = QuotationDocumentTypePDF
	}

	preview, err := s.GetCustomerQuotationPreview(ctx, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("get quotation preview for doc gen: %w", err)
	}

	curVersion, _ := s.repo.GetLatestDocumentVersion(ctx, orgID, quotationID, docType)
	nextVersion := curVersion + 1

	filePath, fileName, _, err := s.docGen.GenerateQuotationPDF(ctx, preview, nextVersion)
	if err != nil {
		return nil, fmt.Errorf("generate document (%s): %w", docType, err)
	}

	actor := resolveActor(ctx)
	doc := &QuotationDocument{
		OrgID:        orgID,
		QuotationID:  quotationID,
		DocumentType: docType,
		FileName:     fileName,
		FilePath:     filePath,
		Version:      nextVersion,
		GeneratedAt:  time.Now(),
	}
	if userID > 0 {
		doc.GeneratedBy = &userID
	}

	if err := s.repo.CreateQuotationDocument(ctx, doc); err != nil {
		return nil, fmt.Errorf("record quotation document: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationDocumentGenerated,
		Description:  fmt.Sprintf("Generated %s document (v%d): %s", strings.ToUpper(docType), nextVersion, fileName),
		Actor:        actor,
	})

	return doc, nil
}

func (s *service) ListQuotationDocuments(ctx context.Context, orgID, quotationID int64) ([]*QuotationDocument, error) {
	return s.repo.GetQuotationDocuments(ctx, orgID, quotationID)
}

func (s *service) GetQuotationDocument(ctx context.Context, orgID, quotationID, docID int64) (*QuotationDocument, []byte, error) {
	doc, err := s.repo.GetQuotationDocumentByID(ctx, orgID, quotationID, docID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, nil, fmt.Errorf("get quotation document: %w", err)
	}

	content, err := os.ReadFile(doc.FilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("read document file: %w", err)
	}

	return doc, content, nil
}

// ── Public Access & Sharing (Task 18.5) ──────────────────────────────────────

func (s *service) CreateQuotationPublicLink(ctx context.Context, orgID, quotationID, userID int64, req *CreateQuotationPublicLinkRequest) (*QuotationPublicLink, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	if !IsQuotationSharable(q.Status) {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf(
			"quotation in status %s cannot be shared publicly. Quotation must be approved or sent first",
			q.Status,
		))
	}

	token, err := GenerateSecurePublicToken()
	if err != nil {
		return nil, fmt.Errorf("generate secure token: %w", err)
	}

	actor := resolveActor(ctx)
	days := 14
	if req != nil && req.ValidityDays != nil && *req.ValidityDays > 0 {
		days = *req.ValidityDays
	}
	expTime := time.Now().AddDate(0, 0, days)

	link := &QuotationPublicLink{
		OrgID:       orgID,
		QuotationID: quotationID,
		PublicToken: token,
		Status:      QuotationPublicLinkActive,
		ExpiresAt:   &expTime,
		CreatedBy:   userID,
	}

	if err := s.repo.CreateQuotationPublicLink(ctx, link); err != nil {
		return nil, fmt.Errorf("create public link: %w", err)
	}

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationPublicLinkCreated,
		Description:  fmt.Sprintf("Created secure public sharing link (valid for %d days)", days),
		Actor:        actor,
	})

	return link, nil
}

func (s *service) ListQuotationPublicLinks(ctx context.Context, orgID, quotationID int64) ([]*QuotationPublicLink, error) {
	return s.repo.GetQuotationPublicLinks(ctx, orgID, quotationID)
}

func (s *service) RevokeQuotationPublicLink(ctx context.Context, orgID, quotationID, linkID, userID int64, req *RevokeQuotationPublicLinkRequest) error {
	reason := "Revoked by user"
	if req != nil && req.Reason != "" {
		reason = req.Reason
	}

	if err := s.repo.RevokeQuotationPublicLink(ctx, orgID, quotationID, linkID, userID, reason); err != nil {
		return fmt.Errorf("revoke public link: %w", err)
	}

	actor := resolveActor(ctx)
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationActivityLinkRevoked,
		Description:  fmt.Sprintf("Revoked public sharing link: %s", reason),
		Actor:        actor,
	})

	return nil
}

func (s *service) GetPublicQuotationByToken(ctx context.Context, token string, clientIP, userAgent string) (*QuotationPublicViewResponse, error) {
	if token == "" {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("token is required"))
	}

	link, err := s.repo.GetQuotationPublicLinkByToken(ctx, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, fmt.Errorf("fetch public link: %w", err)
	}

	if link.Status == QuotationPublicLinkRevoked {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("this quotation link has been revoked"))
	}
	if link.Status == QuotationPublicLinkExpired || IsPublicLinkExpired(link) {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("this quotation link has expired"))
	}

	_ = s.repo.IncrementPublicLinkAccess(ctx, link.ID)

	preview, err := s.GetCustomerQuotationPreview(ctx, link.OrgID, link.QuotationID)
	if err != nil {
		return nil, fmt.Errorf("load quotation details: %w", err)
	}

	// Update quotation status if in SENT
	if preview.Status == QuotationStatusSent {
		_ = s.repo.MarkQuotationViewed(ctx, link.OrgID, link.QuotationID, "Customer (Public Link)", "", clientIP, userAgent)
		preview.Status = QuotationStatusViewed
	}

	_ = s.repo.RecordPublicView(ctx, link.OrgID, link.QuotationID, "Customer (Public Link)", "", clientIP, userAgent)

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        link.OrgID,
		QuotationID:  link.QuotationID,
		ActivityType: QuotationPublicLinkAccessed,
		Description:  fmt.Sprintf("Quotation accessed via public portal from IP: %s", clientIP),
		Actor:        "Customer Portal",
	})

	canAccept := preview.Status == QuotationStatusSent || preview.Status == QuotationStatusViewed
	canDecline := canAccept

	return &QuotationPublicViewResponse{
		PublicToken: link.PublicToken,
		Status:      link.Status,
		ExpiresAt:   link.ExpiresAt,
		Quotation:   preview,
		AccessCount: link.AccessCount + 1,
		CanAccept:   canAccept,
		CanDecline:  canDecline,
	}, nil
}

func (s *service) PublicAcceptQuotation(ctx context.Context, token string, clientIP, userAgent string, req *PublicAcceptQuotationRequest) error {
	if token == "" {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("token is required"))
	}
	if req == nil || strings.TrimSpace(req.AcceptedBy) == "" {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("accepted_by is required to accept quotation"))
	}

	link, err := s.repo.GetQuotationPublicLinkByToken(ctx, token)
	if err != nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}
	if link.Status == QuotationPublicLinkRevoked || link.Status == QuotationPublicLinkExpired || IsPublicLinkExpired(link) {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("this quotation link is no longer valid"))
	}

	q, err := s.repo.GetQuotationByID(ctx, link.OrgID, link.QuotationID)
	if err != nil {
		return err
	}

	if q.Status != QuotationStatusSent && q.Status != QuotationStatusViewed {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("quotation in status %s cannot be accepted", q.Status))
	}

	actor := strings.TrimSpace(req.AcceptedBy)

	if err := s.repo.AcceptQuotation(ctx, link.OrgID, link.QuotationID, actor, req.Comments); err != nil {
		return fmt.Errorf("accept quotation: %w", err)
	}

	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          link.OrgID,
		QuotationID:    link.QuotationID,
		Action:         QuotationApprovalActionAccepted,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusAccepted,
		ActorName:      actor,
		Comments:       req.Comments,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        link.OrgID,
		QuotationID:  link.QuotationID,
		ActivityType: QuotationAccepted,
		Description:  fmt.Sprintf("Quotation accepted by customer %s via public portal", actor),
		Actor:        actor,
	})

	return nil
}

func (s *service) PublicDeclineQuotation(ctx context.Context, token string, clientIP, userAgent string, req *PublicDeclineQuotationRequest) error {
	if token == "" {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("token is required"))
	}
	if req == nil || strings.TrimSpace(req.Reason) == "" {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("reason is required to decline quotation"))
	}

	link, err := s.repo.GetQuotationPublicLinkByToken(ctx, token)
	if err != nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}
	if link.Status == QuotationPublicLinkRevoked || link.Status == QuotationPublicLinkExpired || IsPublicLinkExpired(link) {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("this quotation link is no longer valid"))
	}

	q, err := s.repo.GetQuotationByID(ctx, link.OrgID, link.QuotationID)
	if err != nil {
		return err
	}

	if q.Status != QuotationStatusSent && q.Status != QuotationStatusViewed {
		return svcerror.WrapServiceError(svcerror.ErrInvalidArgument, fmt.Errorf("quotation in status %s cannot be declined", q.Status))
	}

	actor := "Customer"
	if req.DeclinedBy != "" {
		actor = strings.TrimSpace(req.DeclinedBy)
	}

	if err := s.repo.DeclineQuotation(ctx, link.OrgID, link.QuotationID, actor, req.Reason); err != nil {
		return fmt.Errorf("decline quotation: %w", err)
	}

	_ = s.repo.CreateApprovalHistory(ctx, &QuotationApprovalHistory{
		OrgID:          link.OrgID,
		QuotationID:    link.QuotationID,
		Action:         QuotationApprovalActionDeclined,
		PreviousStatus: q.Status,
		NewStatus:      QuotationStatusDeclined,
		ActorName:      actor,
		Comments:       req.Reason,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        link.OrgID,
		QuotationID:  link.QuotationID,
		ActivityType: QuotationDeclined,
		Description:  fmt.Sprintf("Quotation declined by customer %s: %s", actor, req.Reason),
		Actor:        actor,
	})

	return nil
}

// ─── Quotation-to-Booking Operational Conversion Implementation (Task 18.6) ──

func (s *service) GetQuotationConversionPreview(ctx context.Context, orgID, quotationID int64) (*QuotationConversionPreview, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quotation %d not found: %w", quotationID, err))
	}

	charges, err := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	if err != nil {
		charges = []*QuotationChargeItem{}
	}

	var cust *QuotationCustomerInfo
	if q.CustomerID != nil && *q.CustomerID > 0 {
		cust, _ = s.repo.GetCustomerInfo(ctx, orgID, *q.CustomerID)
	}

	preview := BuildQuotationConversionPreview(q, charges, cust)
	return preview, nil
}

func (s *service) ConvertQuotationToBooking(ctx context.Context, orgID, quotationID, userID int64, req *ConvertQuotationToBookingRequest) (*QuotationConversionResult, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	if req == nil {
		req = &ConvertQuotationToBookingRequest{}
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, fmt.Errorf("quotation %d not found: %w", quotationID, err))
	}

	// 1. Idempotency Check: if quotation already converted, return existing booking immediately
	if q.ConvertedBookingID != nil && q.ConversionStatus == QuotationConversionStatusConverted {
		bookingNumber := fmt.Sprintf("BK-%s", q.QuotationNumber)
		shNumber := ""
		if q.ConvertedShipmentID != nil {
			shNumber = fmt.Sprintf("SH-%d", *q.ConvertedShipmentID)
		}
		var shNumberPtr *string
		if shNumber != "" {
			shNumberPtr = &shNumber
		}
		convTime := time.Now()
		if q.ConvertedAt != nil {
			convTime = *q.ConvertedAt
		}

		return &QuotationConversionResult{
			Success:          true,
			QuotationID:      q.ID,
			QuotationNumber:  q.QuotationNumber,
			BookingID:        *q.ConvertedBookingID,
			BookingNumber:    bookingNumber,
			ShipmentID:       q.ConvertedShipmentID,
			ShipmentNumber:   shNumberPtr,
			Message:          "Quotation was already converted. Returned existing booking reference.",
			ConversionStatus: QuotationConversionStatusConverted,
			ConvertedAt:      convTime,
			AlreadyConverted: true,
		}, nil
	}

	// 2. Strict Acceptance Verification
	if q.Status != QuotationStatusAccepted {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("quotation cannot be converted in status %s. Only ACCEPTED quotations can be converted to operational bookings", q.Status))
	}

	// 3. Evaluate Eligibility
	charges, _ := s.repo.GetQuotationCharges(ctx, orgID, quotationID)
	canConvert, blockingReasons := CanConvertQuotationToBooking(q, charges)
	if !canConvert {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("conversion blocked: %s", strings.Join(blockingReasons, "; ")))
	}

	creator := "Operations Team"
	if userID > 0 {
		creator = fmt.Sprintf("User #%d", userID)
	}

	// Record start activity
	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationConversionStarted,
		Description:  fmt.Sprintf("Commercial handover started for quotation %s by %s", q.QuotationNumber, creator),
		Actor:        creator,
	})

	// 4. Execute Transactional Booking Creation & Shipment Linking
	bookingID, bookingNumber, shipmentID, shipmentNumber, err := s.repo.CreateBookingFromQuotationTx(ctx, orgID, q, req, creator)
	if err != nil {
		// Record failure
		_ = s.repo.MarkQuotationConversionFailed(ctx, orgID, quotationID, creator, err.Error())
		_ = s.repo.CreateQuotationConversionHistory(ctx, &QuotationConversionHistory{
			OrgID:       orgID,
			QuotationID: quotationID,
			Action:      "CONVERT_TO_BOOKING",
			Status:      QuotationConversionStatusFailed,
			Message:     fmt.Sprintf("Failed to convert quotation: %v", err),
			PerformedBy: creator,
		})
		_ = s.repo.CreateActivity(ctx, &QuotationActivity{
			OrgID:        orgID,
			QuotationID:  quotationID,
			ActivityType: QuotationConversionFailed,
			Description:  fmt.Sprintf("Commercial conversion failed: %v", err),
			Actor:        creator,
		})
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, fmt.Errorf("failed to create booking from quotation: %w", err))
	}

	// 5. Update Quotation Status & References
	notes := req.OperationalNotes
	if err := s.repo.MarkQuotationConverted(ctx, orgID, quotationID, bookingID, shipmentID, creator, notes); err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, fmt.Errorf("failed to update quotation conversion references: %w", err))
	}

	// 6. Record Audit History and Activity Event
	successMsg := fmt.Sprintf("Quotation converted to Booking %s", bookingNumber)
	if shipmentID != nil {
		successMsg += fmt.Sprintf(" with linked Shipment #%d", *shipmentID)
	}

	_ = s.repo.CreateQuotationConversionHistory(ctx, &QuotationConversionHistory{
		OrgID:       orgID,
		QuotationID: quotationID,
		BookingID:   &bookingID,
		ShipmentID:  shipmentID,
		Action:      "CONVERTED_TO_BOOKING",
		Status:      "SUCCESS",
		Message:     successMsg,
		PerformedBy: creator,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: QuotationConvertedToBooking,
		Description:  successMsg,
		Actor:        creator,
	})

	return &QuotationConversionResult{
		Success:          true,
		QuotationID:      quotationID,
		QuotationNumber:  q.QuotationNumber,
		BookingID:        bookingID,
		BookingNumber:    bookingNumber,
		ShipmentID:       shipmentID,
		ShipmentNumber:   shipmentNumber,
		Message:          "Quotation converted to operational booking successfully",
		ConversionStatus: QuotationConversionStatusConverted,
		ConvertedAt:      time.Now(),
		AlreadyConverted: false,
	}, nil
}

func (s *service) GetQuotationConversionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationConversionHistory, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetQuotationConversionHistory(ctx, orgID, quotationID)
}

// ─── Booking Confirmation, Commercial Handover & Lineage Traceability (Task 18.7) ──

func (s *service) GetQuotationOperationalHandover(ctx context.Context, orgID, quotationID int64) (*QuotationOperationalHandover, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	var booking *RawOperationalBooking
	if q.ConvertedBookingID != nil {
		booking, _ = s.repo.GetOperationalBooking(ctx, orgID, *q.ConvertedBookingID)
	} else {
		booking, _ = s.repo.GetOperationalBookingByQuotationID(ctx, orgID, quotationID)
	}

	var shipment *RawOperationalShipment
	if q.ConvertedShipmentID != nil {
		shipment, _ = s.repo.GetOperationalShipment(ctx, orgID, *q.ConvertedShipmentID)
	} else if booking != nil {
		shipment, _ = s.repo.GetOperationalShipmentByBookingID(ctx, orgID, booking.ID)
	}

	handover := BuildQuotationOperationalHandover(q, booking, shipment)
	return handover, nil
}

func (s *service) ConfirmQuotationBookingHandover(ctx context.Context, orgID, quotationID, userID int64, req *ConfirmQuotationHandoverRequest) (*QuotationOperationalHandover, error) {
	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
		}
		return nil, err
	}

	var booking *RawOperationalBooking
	if q.ConvertedBookingID != nil {
		booking, _ = s.repo.GetOperationalBooking(ctx, orgID, *q.ConvertedBookingID)
	} else {
		booking, _ = s.repo.GetOperationalBookingByQuotationID(ctx, orgID, quotationID)
	}

	if booking == nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("cannot confirm handover: quotation has not yet been converted to a booking"))
	}

	if booking.CommercialHandoverStatus == CommercialHandoverBookingConfirmed || booking.CommercialHandoverStatus == CommercialHandoverCompleted {
		return nil, svcerror.WrapServiceError(svcerror.ErrInvalidArgument,
			fmt.Errorf("booking handover is already confirmed"))
	}

	actor := resolveActor(ctx)
	err = s.repo.UpdateBookingHandoverStatus(ctx, orgID, booking.ID, CommercialHandoverBookingConfirmed, actor)
	if err != nil {
		return nil, fmt.Errorf("update booking handover status: %w", err)
	}

	desc := fmt.Sprintf("Commercial handover confirmed for Booking %s (Quote %s)", booking.BookingNumber, q.QuotationNumber)
	if req != nil && req.ConfirmationNotes != "" {
		desc += fmt.Sprintf(" — Notes: %s", req.ConfirmationNotes)
	}

	var metaJSON []byte
	if req != nil {
		metaJSON = []byte(fmt.Sprintf(`{"notes":%q,"notify_operations":%t}`, req.ConfirmationNotes, req.NotifyOperations))
	}

	_ = s.repo.CreateOperationalHandoverHistory(ctx, &QuotationOperationalHandoverHistory{
		OrgID:       orgID,
		QuotationID: quotationID,
		BookingID:   &booking.ID,
		ShipmentID:  q.ConvertedShipmentID,
		EventType:   BookingConfirmedFromQuotation,
		Description: desc,
		Metadata:    metaJSON,
		PerformedBy: actor,
	})

	_ = s.repo.CreateActivity(ctx, &QuotationActivity{
		OrgID:        orgID,
		QuotationID:  quotationID,
		ActivityType: BookingConfirmedFromQuotation,
		Description:  desc,
		Actor:        actor,
	})

	return s.GetQuotationOperationalHandover(ctx, orgID, quotationID)
}

func (s *service) GetQuotationOperationalChanges(ctx context.Context, orgID, quotationID int64) ([]*OperationalChange, error) {
	handover, err := s.GetQuotationOperationalHandover(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}
	return handover.OperationalChanges, nil
}

func (s *service) GetQuotationHandoverHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationOperationalHandoverHistory, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetOperationalHandoverHistory(ctx, orgID, quotationID)
}

// ─── Quotation Analytics, Performance & Intelligence (Task 18.8) ────────────

func (s *service) GetQuotationAnalyticsOverview(ctx context.Context, orgID int64) (*QuotationAnalyticsOverview, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	overview, err := s.repo.GetQuotationAnalyticsOverview(ctx, orgID)
	if err != nil {
		return nil, err
	}

	customers, _ := s.repo.GetCustomerQuotationPerformance(ctx, orgID)
	risks, _ := s.repo.GetQuotationExpiryRisk(ctx, orgID)

	// Evaluate deterministic insights
	overview.Insights = GenerateQuotationOperationalInsights(overview, customers, risks)

	return overview, nil
}

func (s *service) GetQuotationAnalyticsTrends(ctx context.Context, orgID int64, days int) ([]*QuotationTrendDataPoint, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	if days <= 0 {
		days = 30
	}
	return s.repo.GetQuotationAnalyticsTrends(ctx, orgID, days)
}

func (s *service) GetCustomerQuotationPerformance(ctx context.Context, orgID int64) ([]*CustomerQuotationPerformance, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetCustomerQuotationPerformance(ctx, orgID)
}

func (s *service) GetQuotationPerformanceByMode(ctx context.Context, orgID int64) ([]*QuotationPerformanceByMode, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetQuotationPerformanceByMode(ctx, orgID)
}

func (s *service) GetQuotationExpiryRisk(ctx context.Context, orgID int64) ([]*QuotationRiskItem, error) {
	if orgID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetQuotationExpiryRisk(ctx, orgID)
}

// ── Task 19.5: Rate-to-Quotation Integration Service Implementations ──────────

func (s *service) GetQuotationRateCandidates(ctx context.Context, orgID, quotationID int64) (*QuotationRateCandidatesResponse, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	candidates, err := s.repo.GetQuotationRateCandidates(ctx, orgID, quotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rate candidates: %w", err)
	}

	// Add risk warnings
	now := time.Now()
	for i := range candidates {
		candidates[i].RiskWarnings = DetectCandidateRiskWarnings(&candidates[i], now)
	}

	// Evaluate recommendations (Cheapest, Fastest, Best Value)
	evaluated := EvaluateCandidateRecommendations(candidates, q.Currency)

	resp := &QuotationRateCandidatesResponse{
		QuotationID:     quotationID,
		Origin:          q.Origin,
		Destination:     q.Destination,
		TransportMode:   q.TransportMode,
		ServiceType:     q.ServiceType,
		Candidates:      evaluated,
		TotalCandidates: len(evaluated),
	}

	for _, c := range evaluated {
		for _, tag := range c.RecommendationTags {
			candCopy := c
			if tag == "CHEAPEST" && resp.CheapestCandidate == nil {
				resp.CheapestCandidate = &candCopy
			}
			if tag == "FASTEST" && resp.FastestCandidate == nil {
				resp.FastestCandidate = &candCopy
			}
			if tag == "BEST_VALUE" && resp.BestValueCandidate == nil {
				resp.BestValueCandidate = &candCopy
			}
		}
	}

	return resp, nil
}

func (s *service) GetQuotationRateSelection(ctx context.Context, orgID, quotationID int64) (*QuotationRateSelection, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetActiveQuotationRateSelection(ctx, orgID, quotationID)
}

func (s *service) SelectQuotationRate(ctx context.Context, req *SelectQuotationRateRequest) (*QuotationRateSnapshot, error) {
	if req == nil || req.OrgID <= 0 || req.QuotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// 1. Verify quotation is editable
	q, err := s.repo.GetQuotationByID(ctx, req.OrgID, req.QuotationID)
	if err != nil {
		return nil, err
	}
	if err := ValidateQuotationForRateSelection(q); err != nil {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// 2. Fetch candidates & find selected target
	candidatesResp, err := s.GetQuotationRateCandidates(ctx, req.OrgID, req.QuotationID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch rate candidates: %w", err)
	}

	var target *QuotationRateCandidate
	for _, c := range candidatesResp.Candidates {
		if req.SourceType == RateSourceManaged && req.RateID != nil && c.RateID != nil && *c.RateID == *req.RateID {
			target = &c
			break
		}
		if req.SourceType == RateSourceSpot && req.SpotRateResponseID != nil && c.SpotRateResponseID != nil && *c.SpotRateResponseID == *req.SpotRateResponseID {
			target = &c
			break
		}
	}

	if target == nil {
		return nil, svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	// 3. Build selection model
	sel := &QuotationRateSelection{
		OrgID:              req.OrgID,
		QuotationID:        req.QuotationID,
		RateID:             req.RateID,
		SpotRateRequestID:  target.SpotRateRequestID,
		SpotRateResponseID: req.SpotRateResponseID,
		RateSourceType:     req.SourceType,
		SelectedBy:         req.User,
		Notes:              req.Notes,
	}

	// 4. Build snapshot model
	snap, err := BuildQuotationRateSnapshot(req.OrgID, req.QuotationID, 0, target, req.User)
	if err != nil {
		return nil, fmt.Errorf("build rate snapshot: %w", err)
	}

	// 5. Construct quotation charge items from candidate base rate
	baseCharge := &QuotationChargeItem{
		OrgID:            req.OrgID,
		QuotationID:      req.QuotationID,
		ChargeCategory:   "FREIGHT",
		ChargeName:       fmt.Sprintf("Base Freight (%s - %s)", target.CarrierName, target.EquipmentType),
		CalculationBasis: "FLAT",
		Quantity:         1.0,
		UnitPrice:        target.BaseRate,
		SellAmount:       target.BaseRate,
		Currency:         target.Currency,
		DisplayOrder:     0,
		CreatedBy:        req.User,
	}

	charges := []*QuotationChargeItem{baseCharge}
	totals := map[string]float64{
		"subtotal":     target.BaseRate,
		"surcharges":   0.0,
		"taxes":        0.0,
		"total_amount": target.BaseRate,
	}

	// 6. Build audit history event
	desc := fmt.Sprintf("Selected rate from carrier %s (%s %.2f)", target.CarrierName, target.Currency, target.CommercialTotal)
	if req.SourceType == RateSourceSpot {
		desc = fmt.Sprintf("Selected spot rate response from %s (%s %.2f)", target.CarrierName, target.Currency, target.CommercialTotal)
	}

	eventType := RateEventSelected
	if req.SourceType == RateSourceSpot {
		eventType = RateEventSpotSelected
	}

	hist := &QuotationRateSelectionHistory{
		OrgID:       req.OrgID,
		QuotationID: req.QuotationID,
		EventType:   eventType,
		Description: desc,
		PerformedBy: req.User,
	}

	// 7. Execute transactional write
	if err := s.repo.CreateQuotationRateSelectionTx(ctx, req.OrgID, sel, snap, hist, charges, totals); err != nil {
		return nil, fmt.Errorf("failed to select quotation rate: %w", err)
	}

	return s.repo.GetLatestQuotationRateSnapshot(ctx, req.OrgID, req.QuotationID)
}

func (s *service) ReplaceQuotationRate(ctx context.Context, req *ReplaceQuotationRateRequest) (*QuotationRateSnapshot, error) {
	if req == nil || req.OrgID <= 0 || req.QuotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	// Fetch active selection before replacing
	activeSel, _ := s.repo.GetActiveQuotationRateSelection(ctx, req.OrgID, req.QuotationID)

	selectReq := &SelectQuotationRateRequest{
		OrgID:              req.OrgID,
		QuotationID:        req.QuotationID,
		RateID:             req.RateID,
		SpotRateResponseID: req.SpotRateResponseID,
		SourceType:         req.SourceType,
		Notes:              req.Notes,
		User:               req.User,
	}

	snap, err := s.SelectQuotationRate(ctx, selectReq)
	if err != nil {
		return nil, err
	}

	// Record RATE_REPLACED audit event
	var prevID *int64
	if activeSel != nil {
		prevID = &activeSel.ID
	}

	hist := &QuotationRateSelectionHistory{
		OrgID:               req.OrgID,
		QuotationID:         req.QuotationID,
		EventType:           RateEventReplaced,
		PreviousSelectionID: prevID,
		Description:         fmt.Sprintf("Replaced previous commercial rate selection with %s (%s %.2f)", snap.CarrierName, snap.Currency, snap.CommercialTotal),
		PerformedBy:         req.User,
	}

	_ = s.repo.CreateQuotationRateSelectionHistory(ctx, hist)
	return snap, nil
}

func (s *service) RemoveQuotationRate(ctx context.Context, orgID, quotationID int64, user string) error {
	if orgID <= 0 || quotationID <= 0 {
		return svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	q, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return err
	}
	if err := ValidateQuotationForRateSelection(q); err != nil {
		return svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	activeSel, err := s.repo.GetActiveQuotationRateSelection(ctx, orgID, quotationID)
	if err != nil || activeSel == nil {
		return svcerror.NewServiceError(svcerror.ErrResourceNotFound)
	}

	if err := s.repo.DeactivateQuotationRateSelection(ctx, orgID, activeSel.ID); err != nil {
		return err
	}

	hist := &QuotationRateSelectionHistory{
		OrgID:               orgID,
		QuotationID:         quotationID,
		EventType:           RateEventRemoved,
		PreviousSelectionID: &activeSel.ID,
		Description:         "Removed commercial rate selection from quotation",
		PerformedBy:         user,
	}

	return s.repo.CreateQuotationRateSelectionHistory(ctx, hist)
}

func (s *service) GetQuotationRateSnapshot(ctx context.Context, orgID, quotationID int64) (*QuotationRateSnapshot, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetLatestQuotationRateSnapshot(ctx, orgID, quotationID)
}

func (s *service) GetQuotationRateSelectionHistory(ctx context.Context, orgID, quotationID int64) ([]*QuotationRateSelectionHistory, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.GetQuotationRateSelectionHistory(ctx, orgID, quotationID)
}

// ── Task 19.6: Quotation Rate Risk & Commercial Impact Implementations ─────────

func (s *service) GetQuotationRateRisks(ctx context.Context, orgID, quotationID int64) (*QuotationRateRiskSummary, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	quote, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	snapshot, _ := s.repo.GetLatestQuotationRateSnapshot(ctx, orgID, quotationID)
	risks, err := s.repo.GetQuotationRateRisks(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	summary := &QuotationRateRiskSummary{
		QuotationID:       quotationID,
		QuotationNumber:   quote.QuotationNumber,
		Status:            quote.Status,
		HasActiveSnapshot: snapshot != nil,
		Risks:             risks,
		TotalRisks:        len(risks),
	}

	for _, rk := range risks {
		if !rk.IsResolved {
			switch rk.Severity {
			case "CRITICAL":
				summary.CriticalRisks++
			case "WARNING":
				summary.WarningRisks++
			case "INFO":
				summary.InfoRisks++
			}
		}
	}

	// Replacement count
	candidates, err := s.repo.GetQuotationRateCandidates(ctx, orgID, quotationID)
	if err == nil {
		summary.ReplacementCount = len(candidates)
	}

	return summary, nil
}

func (s *service) ResolveQuotationRateRisk(ctx context.Context, orgID, quotationID, riskID int64, user string) error {
	if orgID <= 0 || quotationID <= 0 || riskID <= 0 {
		return svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}
	return s.repo.ResolveQuotationRateRisk(ctx, orgID, quotationID, riskID, user)
}

func (s *service) GetRateReplacementCandidates(ctx context.Context, orgID, quotationID int64) ([]*RateReplacementCandidate, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	candidates, err := s.repo.GetQuotationRateCandidates(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	result := make([]*RateReplacementCandidate, 0, len(candidates))
	for _, c := range candidates {
		cand := &RateReplacementCandidate{
			SourceType:         c.SourceType,
			RateID:             c.RateID,
			SpotRateResponseID: c.SpotRateResponseID,
			CarrierName:        c.CarrierName,
			CarrierCode:        c.CarrierCode,
			RateType:           c.RateType,
			VersionNumber:      c.RateVersion,
			Currency:           c.Currency,
			BaseRate:           c.BaseRate,
			CommercialTotal:    c.CommercialTotal,
			TransitDays:        c.TransitDays,
			RecommendationTags: c.RecommendationTags,
		}
		if c.ValidUntil != nil {
			v := c.ValidUntil.Format("2006-01-02")
			cand.ValidUntil = &v
		}
		result = append(result, cand)
	}

	return result, nil
}

func (s *service) GetCommercialImpactAnalysis(ctx context.Context, orgID, quotationID int64, replacementRateID, replacementSpotID *int64) (*CommercialImpactAnalysis, error) {
	if orgID <= 0 || quotationID <= 0 {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	quote, err := s.repo.GetQuotationByID(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	snapshot, err := s.repo.GetLatestQuotationRateSnapshot(ctx, orgID, quotationID)
	if err != nil || snapshot == nil {
		return nil, fmt.Errorf("no active rate snapshot found on quotation %d", quotationID)
	}

	analysis := &CommercialImpactAnalysis{
		QuotationID:            quotationID,
		QuotationNumber:        quote.QuotationNumber,
		QuotationStatus:        quote.Status,
		CurrentCarrierName:     snapshot.CarrierName,
		CurrentCommercialTotal: snapshot.CommercialTotal,
		CurrentCurrency:        snapshot.Currency,
		CurrentTransitDays:     25, // default estimated
	}
	if snapshot.ValidUntil != nil {
		v := snapshot.ValidUntil.Format("2006-01-02")
		analysis.CurrentValidUntil = &v
	}

	// Find candidate
	candidates, err := s.repo.GetQuotationRateCandidates(ctx, orgID, quotationID)
	if err != nil {
		return nil, err
	}

	var targetCand *QuotationRateCandidate
	for i := range candidates {
		c := &candidates[i]
		if replacementRateID != nil && c.RateID != nil && *c.RateID == *replacementRateID {
			targetCand = c
			break
		}
		if replacementSpotID != nil && c.SpotRateResponseID != nil && *c.SpotRateResponseID == *replacementSpotID {
			targetCand = c
			break
		}
	}

	if targetCand == nil && len(candidates) > 0 {
		targetCand = &candidates[0]
	}

	if targetCand != nil {
		analysis.ReplacementSourceType = targetCand.SourceType
		analysis.ReplacementCarrierName = targetCand.CarrierName
		analysis.ReplacementCommercialTotal = targetCand.CommercialTotal
		analysis.ReplacementCurrency = targetCand.Currency
		analysis.ReplacementTransitDays = targetCand.TransitDays
		if targetCand.ValidUntil != nil {
			v := targetCand.ValidUntil.Format("2006-01-02")
			analysis.ReplacementValidUntil = &v
		}

		if targetCand.Currency != snapshot.Currency {
			analysis.CurrencyMismatch = true
			analysis.ImpactSummary = fmt.Sprintf("Currency difference: current quote is %s %.2f, replacement is %s %.2f. Cross-currency rate comparison requires manual review.",
				snapshot.Currency, snapshot.CommercialTotal, targetCand.Currency, targetCand.CommercialTotal)
		} else {
			diff := targetCand.CommercialTotal - snapshot.CommercialTotal
			analysis.PriceDifferenceAmount = diff
			if snapshot.CommercialTotal > 0 {
				analysis.PriceDifferencePercentage = (diff / snapshot.CommercialTotal) * 100.0
			}
			analysis.TransitDifferenceDays = targetCand.TransitDays - analysis.CurrentTransitDays
			analysis.IsCheaper = diff < 0
			analysis.IsFaster = targetCand.TransitDays < analysis.CurrentTransitDays

			if diff < 0 {
				analysis.ImpactSummary = fmt.Sprintf("Cost reduction of %s %.2f (%.1f%% cheaper) with %s.",
					snapshot.Currency, -diff, -analysis.PriceDifferencePercentage, targetCand.CarrierName)
			} else if diff > 0 {
				analysis.ImpactSummary = fmt.Sprintf("Price increase of %s %.2f (+%.1f%%) with %s.",
					snapshot.Currency, diff, analysis.PriceDifferencePercentage, targetCand.CarrierName)
			} else {
				analysis.ImpactSummary = fmt.Sprintf("Identical commercial rate of %s %.2f with %s.",
					snapshot.Currency, snapshot.CommercialTotal, targetCand.CarrierName)
			}
		}
	} else {
		analysis.ImpactSummary = "No alternative carrier rate candidates available for this trade lane and transport mode."
	}

	return analysis, nil
}

func (s *service) EvaluateQuotationRateRisksForOrg(ctx context.Context, orgID int64) (int, error) {
	details, err := s.repo.GetQuotationsWithActiveRateSelection(ctx, orgID)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	totalRisksCreated := 0

	for _, d := range details {
		input := rates.RateRiskEvaluationInput{
			QuotationID:              d.QuotationID,
			QuotationNumber:          d.QuotationNumber,
			QuotationStatus:          d.QuotationStatus,
			SnapshotCarrierName:      d.SnapshotCarrierName,
			SnapshotCurrency:         d.SnapshotCurrency,
			SnapshotCommercialTotal:  d.SnapshotCommercialTotal,
			SnapshotValidUntil:       d.SnapshotValidUntil,
			SourceRateID:             d.SourceRateID,
			SourceRateStatus:         d.SourceRateStatus,
			SourceRateValidUntil:     d.SourceRateValidUntil,
			SourceRateVersion:        d.SourceRateVersion,
			LatestRateVersion:        d.LatestRateVersion,
			SourceContractID:         d.SourceContractID,
			SourceContractCode:       d.SourceContractCode,
			SourceContractStatus:     d.SourceContractStatus,
			SourceContractEndDate:    d.SourceContractEndDate,
			SourceSpotRateResponseID: d.SourceSpotRateResponseID,
			SourceSpotValidUntil:     d.SourceSpotValidUntil,
			SourceSpotStatus:         d.SourceSpotStatus,
		}

		risks := rates.DetectQuotationRateRisks(input, now)
		for _, rk := range risks {
			_ = s.repo.CreateQuotationRateRiskEvent(ctx, &QuotationRateRisk{
				OrgID:                    orgID,
				QuotationID:              d.QuotationID,
				QuotationRateSnapshotID:  &d.SnapshotID,
				SourceRateID:             d.SourceRateID,
				SourceContractID:         d.SourceContractID,
				SourceSpotRateResponseID: d.SourceSpotRateResponseID,
				RiskType:                 rk.RiskType,
				Severity:                 rk.Severity,
				Headline:                 rk.Headline,
				Description:              rk.Description,
				RecommendedAction:        rk.RecommendedAction,
			})
			totalRisksCreated++
		}
	}

	return totalRisksCreated, nil
}







