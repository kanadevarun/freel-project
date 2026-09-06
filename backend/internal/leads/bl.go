package leads

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"encoding/json"

	"github.com/freel/backend/internal/ai"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/common/crypto"
	"github.com/freel/backend/internal/common/events"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/leads/spec"
	"github.com/freel/backend/internal/middleware"
	"github.com/freel/backend/internal/organization"
	"github.com/freel/backend/internal/svcerror"
)

type InboundEmail struct {
	MailboxID        int64
	RawEmailID       string
	RFCMessageID     string
	ThreadID         string
	From             string
	To               string
	Subject          string
	Body             string
	MessageID        string // Keep for compatibility
	InReplyTo        string
	ReferencesHeader string
	Sender           string
	Recipients       string
	CCRecipients     string
	ReceivedAt       time.Time
}

type BusinessLogic interface {
	CreateLead(ctx context.Context, req spec.CreateLeadRequest) (*spec.Lead, error)
	BulkCreate(ctx context.Context, req spec.ImportLeadsRequest) (int, error)
	GetLead(ctx context.Context, orgID int32, id int32) (*spec.Lead, error)
	ListLeads(ctx context.Context, req spec.ListLeadsRequest) (*spec.ListLeadsResponse, error)
	UpdateLead(ctx context.Context, req spec.UpdateLeadRequest) (*spec.Lead, error)
	DeleteLead(ctx context.Context, req spec.DeleteLeadRequest) error
	GetLeadByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error)
	LogInteraction(ctx context.Context, orgID int32, inter *LeadInteraction) error
	ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error)
	FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error)
	GetInteractionByRawEmailID(ctx context.Context, orgID int32, rawEmailID string) (*LeadInteraction, error)
	GetCustomerIDByCompanyName(ctx context.Context, orgID int32, name string) (int32, error)
	CreateAITask(ctx context.Context, orgID int64, entityType, entityID, taskType string, payload map[string]interface{}) error
	UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error
	UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error
	GetLeadTimeline(ctx context.Context, orgID int32, leadID int32) ([]spec.TimelineEvent, error)
	BulkUpdateLeads(ctx context.Context, req spec.BulkUpdateLeadsRequest) (*spec.BulkUpdateReport, error)
	ProcessInboundEmail(ctx context.Context, orgID int32, email InboundEmail) (*LeadInteraction, error)
	SendClarificationEmail(ctx context.Context, orgID int64, interactionID int64, draftedReply string, summary string) error
	GetInteractionByID(ctx context.Context, orgID int32, id int64) (*LeadInteraction, error)
	ReplyToInteraction(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64, from string, to string, cc string, subject string, body string) (*LeadInteraction, error)
	RetryEmailInteraction(ctx context.Context, orgID int64, leadID int64, interactionID int64) (*LeadInteraction, error)
	GetDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) (*LeadEmailDraft, error)
	SaveDraft(ctx context.Context, draft *LeadEmailDraft) error
	DeleteDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) error
	ClassifyEmailRelevanceWithAI(ctx context.Context, from, subject, body string) (bool, string, error)
}

type businessLogic struct {
	dl            Datalayer
	eventBus      events.Bus
	orgRepo       organization.Repository
	gmailProvider *organization.GmailProvider
	aiGateway     ai.Gateway
}

func NewBusinessLogic(dl Datalayer, eventBus events.Bus, orgRepo organization.Repository, gmailProvider *organization.GmailProvider, aiGateway ai.Gateway) BusinessLogic {
	return &businessLogic{
		dl:            dl,
		eventBus:      eventBus,
		orgRepo:       orgRepo,
		gmailProvider: gmailProvider,
		aiGateway:     aiGateway,
	}
}

func (b *businessLogic) CreateLead(ctx context.Context, req spec.CreateLeadRequest) (*spec.Lead, error) {
	log.Printf("[Leads Service] CreateLead called for OrgID: %d, CompanyName: %s", req.OrgID, req.CompanyName)
	if req.CompanyName == "" {
		return nil, svcerror.NewServiceError(svcerror.ErrInvalidArgument)
	}

	var assignedAt *time.Time
	if req.AssignedTo != nil && *req.AssignedTo > 0 {
		ok, err := b.dl.UserExistsInOrg(ctx, req.OrgID, *req.AssignedTo)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		if !ok {
			return nil, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
		}
		now := time.Now()
		assignedAt = &now
	} else {
		req.AssignedTo = nil
	}

	lead := &spec.Lead{
		OrgID:       req.OrgID,
		CompanyName: req.CompanyName,
		ContactName: req.ContactName,
		Email:       req.Email,
		Phone:       req.Phone,
		Status:      "NEW",
		Source:      req.Source,
		Notes:       req.Notes,
		Location:    req.Location,
		AssignedTo:  req.AssignedTo,
		AssignedAt:  assignedAt,
	}

	err := b.dl.Create(ctx, lead)
	if err != nil {
		log.Printf("[Leads Service] CreateLead failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if len(req.Tags) > 0 {
		err = b.dl.SetTags(ctx, lead.ID, req.Tags)
		if err != nil {
			log.Printf("[Leads Service] CreateLead failed to set tags: %v", err)
		}
		tags, _ := b.dl.GetTags(ctx, lead.ID)
		lead.Tags = tags
	} else {
		lead.Tags = []string{}
	}

	var actorUserID *int64
	if userCtx, ok := middleware.GetUserContext(ctx); ok {
		actorUserID = &userCtx.UserID
	}
	_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "CREATED", "Lead was manually added to the system.", actorUserID)

	// Publish Event
	b.eventBus.Publish(events.Event{
		Type: events.EventLeadCreated,
		Payload: map[string]interface{}{
			"lead_id": lead.ID,
			"org_id":  lead.OrgID,
		},
	})

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(lead.OrgID),
		Action:       domain.ActionCreate,
		Module:       domain.ModuleLeads,
		ResourceType: "LEAD",
		ResourceID:   fmt.Sprintf("%d", lead.ID),
		ResourceName: lead.CompanyName,
		Description:  fmt.Sprintf("Created lead %s", lead.CompanyName),
		Result:       domain.ResultSuccess,
	})

	log.Printf("[Leads Service] CreateLead succeeded for Lead ID: %d", lead.ID)
	return lead, nil
}

func (b *businessLogic) BulkCreate(ctx context.Context, req spec.ImportLeadsRequest) (int, error) {
	log.Printf("[Leads Service] BulkCreate called for OrgID: %d with %d leads", req.OrgID, len(req.Leads))
	createdCount := 0
	for _, lReq := range req.Leads {
		lReq.OrgID = req.OrgID
		_, err := b.CreateLead(ctx, *lReq)
		if err == nil {
			createdCount++
		}
	}
	log.Printf("[Leads Service] BulkCreate completed. Imported %d/%d leads", createdCount, len(req.Leads))
	return createdCount, nil
}

func (b *businessLogic) GetLead(ctx context.Context, orgID int32, id int32) (*spec.Lead, error) {
	log.Printf("[Leads Service] GetLead called for OrgID: %d, ID: %d", orgID, id)
	lead, err := b.dl.GetByID(ctx, orgID, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Leads Service] GetLead resource not found for ID: %d", id)
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] GetLead failed for ID: %d: %v", id, err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	tags, err := b.dl.GetTags(ctx, id)
	if err == nil {
		lead.Tags = tags
	} else {
		lead.Tags = []string{}
	}

	return lead, nil
}

func (b *businessLogic) ListLeads(ctx context.Context, req spec.ListLeadsRequest) (*spec.ListLeadsResponse, error) {
	log.Printf("[Leads Service] ListLeads called for OrgID: %d (Limit: %d, Offset: %d)", req.OrgID, req.Limit, req.Offset)
	if req.Limit <= 0 {
		req.Limit = 50
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Offset < 0 {
		req.Offset = 0
	}

	// Auto-purge any historical non-logistics / bank alert / spam leads from database
	_ = b.dl.PurgeNonLogisticsLeads(ctx, req.OrgID)

	data, total, err := b.dl.List(ctx, req.OrgID, req.Limit, req.Offset, req.Status, req.Search, req.Source)
	if err != nil {
		log.Printf("[Leads Service] ListLeads failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	if data == nil {
		data = []*spec.Lead{}
	}

	if len(data) > 0 {
		var leadIDs []int32
		for _, l := range data {
			leadIDs = append(leadIDs, l.ID)
		}
		tagsMap, err := b.dl.GetTagsBatch(ctx, leadIDs)
		if err == nil {
			for _, l := range data {
				l.Tags = tagsMap[l.ID]
			}
		} else {
			log.Printf("[Leads Service] ListLeads failed to fetch tags batch: %v", err)
			for _, l := range data {
				l.Tags = []string{}
			}
		}
	}

	log.Printf("[Leads Service] ListLeads succeeded. Found %d leads (Total count: %d)", len(data), total)
	return &spec.ListLeadsResponse{
		Data:       data,
		TotalCount: total,
	}, nil
}

func (b *businessLogic) UpdateLead(ctx context.Context, req spec.UpdateLeadRequest) (*spec.Lead, error) {
	log.Printf("[Leads Service] UpdateLead called for OrgID: %d, ID: %d", req.OrgID, req.ID)
	lead, err := b.dl.GetByID(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Leads Service] UpdateLead resource not found for ID: %d", req.ID)
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] UpdateLead failed to fetch lead: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	tags, err := b.dl.GetTags(ctx, lead.ID)
	if err == nil {
		lead.Tags = tags
	}

	beforeMap := map[string]interface{}{
		"company_name": lead.CompanyName,
		"status":       lead.Status,
		"assigned_to":  lead.AssignedTo,
	}

	var actorUserID *int64
	if userCtx, ok := middleware.GetUserContext(ctx); ok {
		actorUserID = &userCtx.UserID
	}

	if req.AssignedTo != nil {
		oldAssignedTo := lead.AssignedTo
		newAssignedTo := req.AssignedTo
		if newAssignedTo != nil && *newAssignedTo <= 0 {
			newAssignedTo = nil
		}

		if (oldAssignedTo == nil && newAssignedTo != nil) || (oldAssignedTo != nil && newAssignedTo == nil) || (oldAssignedTo != nil && newAssignedTo != nil && *oldAssignedTo != *newAssignedTo) {
			if newAssignedTo != nil {
				ok, err := b.dl.UserExistsInOrg(ctx, req.OrgID, *newAssignedTo)
				if err != nil {
					return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
				}
				if !ok {
					return nil, svcerror.NewServiceError(svcerror.ErrInsufficientResourceAccess)
				}
				lead.AssignedTo = newAssignedTo
				now := time.Now()
				lead.AssignedAt = &now
				_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "OWNER_CHANGED", fmt.Sprintf("Lead assigned to user ID %d.", *newAssignedTo), actorUserID)
			} else {
				lead.AssignedTo = nil
				lead.AssignedAt = nil
				_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "OWNER_CHANGED", "Lead owner assignment was removed.", actorUserID)
			}
		}
	}

	if req.CompanyName != nil {
		lead.CompanyName = *req.CompanyName
	}
	if req.ContactName != nil {
		lead.ContactName = req.ContactName
	}
	if req.Email != nil {
		lead.Email = req.Email
	}
	if req.Phone != nil {
		lead.Phone = req.Phone
	}
	if req.Status != nil {
		oldStatus := lead.Status
		newStatus := *req.Status

		if oldStatus != newStatus {
			lead.Status = newStatus
			_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "STATUS_CHANGED", fmt.Sprintf("Status changed from %s to %s.", oldStatus, newStatus), actorUserID)

			if oldStatus != "CONVERTED" && lead.Status == "CONVERTED" {
				log.Printf("[Leads Service] Lead converted to customer. Ensuring customer record exists.")
				_ = b.dl.EnsureCustomerForLead(ctx, lead)
			}
		}
	}
	if req.Source != nil {
		lead.Source = req.Source
	}
	if req.AIScore != nil {
		lead.AIScore = *req.AIScore
	}
	if req.AIResearchReport != nil {
		lead.AIResearchReport = req.AIResearchReport
	}
	if req.Notes != nil {
		lead.Notes = req.Notes
	}
	if req.Location != nil {
		lead.Location = req.Location
	}

	err = b.dl.Update(ctx, lead)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] UpdateLead save failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	afterMap := map[string]interface{}{
		"company_name": lead.CompanyName,
		"status":       lead.Status,
		"assigned_to":  lead.AssignedTo,
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        int64(lead.OrgID),
		Action:       domain.ActionUpdate,
		Module:       domain.ModuleLeads,
		ResourceType: "LEAD",
		ResourceID:   fmt.Sprintf("%d", lead.ID),
		ResourceName: lead.CompanyName,
		Description:  fmt.Sprintf("Updated lead %s", lead.CompanyName),
		Before:       beforeMap,
		After:        afterMap,
		Result:       domain.ResultSuccess,
	})

	if req.Tags != nil {
		err = b.dl.SetTags(ctx, lead.ID, req.Tags)
		if err != nil {
			log.Printf("[Leads Service] UpdateLead failed to update tags: %v", err)
		} else {
			_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "TAGS_CHANGED", fmt.Sprintf("Tags updated to: %s.", strings.Join(req.Tags, ", ")), actorUserID)
		}
		updatedTags, _ := b.dl.GetTags(ctx, lead.ID)
		lead.Tags = updatedTags
	}

	reloadedLead, err := b.dl.GetByID(ctx, req.OrgID, lead.ID)
	if err == nil {
		reloadedLead.Tags = lead.Tags
		lead = reloadedLead
	}

	_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "UPDATED", "Lead details were updated.", actorUserID)

	log.Printf("[Leads Service] UpdateLead succeeded for Lead ID: %d", lead.ID)
	return lead, nil
}

func (b *businessLogic) DeleteLead(ctx context.Context, req spec.DeleteLeadRequest) error {
	log.Printf("[Leads Service] DeleteLead called for OrgID: %d, ID: %d", req.OrgID, req.ID)
	err := b.dl.Delete(ctx, req.OrgID, req.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Leads Service] DeleteLead resource not found for ID: %d", req.ID)
			return svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] DeleteLead failed: %v", err)
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	log.Printf("[Leads Service] DeleteLead succeeded for Lead ID: %d", req.ID)
	return nil
}

func (b *businessLogic) GetLeadByEmail(ctx context.Context, orgID int32, email string) (*spec.Lead, error) {
	log.Printf("[Leads Service] GetLeadByEmail called for OrgID: %d, Email: %s", orgID, email)
	lead, err := b.dl.GetByEmail(ctx, orgID, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("[Leads Service] GetLeadByEmail not found for Email: %s", email)
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] GetLeadByEmail failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return lead, nil
}

func (b *businessLogic) LogInteraction(ctx context.Context, orgID int32, inter *LeadInteraction) error {
	log.Printf("[Leads Service] LogInteraction called for OrgID: %d, LeadID: %d, Direction: %s", orgID, inter.LeadID, inter.Direction)
	inter.OrgID = int64(orgID)
	err := b.dl.LogInteraction(ctx, inter)
	if err != nil {
		log.Printf("[Leads Service] LogInteraction failed: %v", err)
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) ListInteractions(ctx context.Context, orgID int32, leadID int32) ([]*LeadInteraction, error) {
	log.Printf("[Leads Service] ListInteractions called for OrgID: %d, LeadID: %d", orgID, leadID)
	list, err := b.dl.ListInteractions(ctx, orgID, leadID)
	if err != nil {
		log.Printf("[Leads Service] ListInteractions failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return list, nil
}

func (b *businessLogic) FindByThreadID(ctx context.Context, orgID int32, threadID string) ([]*LeadInteraction, error) {
	log.Printf("[Leads Service] FindByThreadID called for OrgID: %d, ThreadID: %s", orgID, threadID)
	list, err := b.dl.FindByThreadID(ctx, orgID, threadID)
	if err != nil {
		log.Printf("[Leads Service] FindByThreadID failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return list, nil
}

func (b *businessLogic) GetInteractionByRawEmailID(ctx context.Context, orgID int32, rawEmailID string) (*LeadInteraction, error) {
	log.Printf("[Leads Service] GetInteractionByRawEmailID called for OrgID: %d, RawEmailID: %s", orgID, rawEmailID)
	inter, err := b.dl.GetInteractionByRawEmailID(ctx, orgID, rawEmailID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		log.Printf("[Leads Service] GetInteractionByRawEmailID failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return inter, nil
}

func (b *businessLogic) GetCustomerIDByCompanyName(ctx context.Context, orgID int32, name string) (int32, error) {
	log.Printf("[Leads Service] GetCustomerIDByCompanyName called for OrgID: %d, Name: %s", orgID, name)
	customerID, err := b.dl.GetCustomerIDByCompanyName(ctx, orgID, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return 0, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return customerID, nil
}



func (b *businessLogic) CreateAITask(ctx context.Context, orgID int64, entityType, entityID, taskType string, payload map[string]interface{}) error {
	log.Printf("[Leads Service] CreateAITask called for OrgID: %d, EntityType: %s, EntityID: %s, TaskType: %s", orgID, entityType, entityID, taskType)
	err := b.dl.CreateAITask(ctx, orgID, entityType, entityID, taskType, payload)
	if err != nil {
		log.Printf("[Leads Service] CreateAITask failed: %v", err)
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) UpdateInteractionAI(ctx context.Context, orgID int64, id int64, intent string, sentiment string, confidence int, linkedRFQID *int64, aiSummary string, draftedReply string) error {
	log.Printf("[Leads Service] UpdateInteractionAI called for OrgID: %d, InteractionID: %d, Intent: %s, Sentiment: %s", orgID, id, intent, sentiment)
	err := b.dl.UpdateInteractionAI(ctx, orgID, id, intent, sentiment, confidence, linkedRFQID, aiSummary, draftedReply)
	if err != nil {
		log.Printf("[Leads Service] UpdateInteractionAI failed: %v", err)
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) UpdateInteractionContext(ctx context.Context, orgID int64, id int64, partialCtx map[string]interface{}) error {
	log.Printf("[Leads Service] UpdateInteractionContext called for OrgID: %d, InteractionID: %d", orgID, id)
	err := b.dl.UpdateInteractionContext(ctx, orgID, id, partialCtx)
	if err != nil {
		log.Printf("[Leads Service] UpdateInteractionContext failed: %v", err)
		return svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return nil
}

func (b *businessLogic) GetLeadTimeline(ctx context.Context, orgID int32, leadID int32) ([]spec.TimelineEvent, error) {
	log.Printf("[Leads Service] GetLeadTimeline called for OrgID: %d, LeadID: %d", orgID, leadID)
	
	_, err := b.dl.GetByID(ctx, orgID, leadID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, svcerror.WrapServiceError(svcerror.ErrResourceNotFound, err)
		}
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	events, err := b.dl.GetActivities(ctx, orgID, leadID)
	if err != nil {
		log.Printf("[Leads Service] GetLeadTimeline failed: %v", err)
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}

	return events, nil
}

func (b *businessLogic) BulkUpdateLeads(ctx context.Context, req spec.BulkUpdateLeadsRequest) (*spec.BulkUpdateReport, error) {
	log.Printf("[Leads Service] BulkUpdateLeads called for OrgID: %d, count: %d", req.OrgID, len(req.IDs))

	report := &spec.BulkUpdateReport{
		SuccessIDs: []int32{},
		FailedIDs:  make(map[string]string),
	}

	var actorUserID *int64
	if userCtx, ok := middleware.GetUserContext(ctx); ok {
		actorUserID = &userCtx.UserID
	}

	if req.Status != nil && *req.Status == "CONVERTED" {
		for _, id := range req.IDs {
			report.FailedIDs[strconv.Itoa(int(id))] = "Direct bulk conversion to CONVERTED is not allowed. Convert leads individually via RFQ creation."
		}
		return report, nil
	}

	if req.AssignedTo != nil && *req.AssignedTo > 0 {
		ok, err := b.dl.UserExistsInOrg(ctx, req.OrgID, *req.AssignedTo)
		if err != nil {
			return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
		}
		if !ok {
			for _, id := range req.IDs {
				report.FailedIDs[strconv.Itoa(int(id))] = "Assigned user does not exist or does not belong to the same organization."
			}
			return report, nil
		}
	}

	for _, id := range req.IDs {
		lead, err := b.dl.GetByID(ctx, req.OrgID, id)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				report.FailedIDs[strconv.Itoa(int(id))] = "Lead not found in this organization."
			} else {
				report.FailedIDs[strconv.Itoa(int(id))] = fmt.Sprintf("Database error: %v", err)
			}
			continue
		}

		dirty := false

		if req.Status != nil {
			oldStatus := lead.Status
			newStatus := *req.Status
			if oldStatus != newStatus {
				lead.Status = newStatus
				dirty = true
				_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "STATUS_CHANGED", fmt.Sprintf("Status changed bulk from %s to %s.", oldStatus, newStatus), actorUserID)
			}
		}

		if req.AssignedTo != nil {
			oldOwner := lead.AssignedTo
			newOwner := req.AssignedTo
			if newOwner != nil && *newOwner <= 0 {
				newOwner = nil
			}

			if (oldOwner == nil && newOwner != nil) || (oldOwner != nil && newOwner == nil) || (oldOwner != nil && newOwner != nil && *oldOwner != *newOwner) {
				lead.AssignedTo = newOwner
				if newOwner != nil {
					now := time.Now()
					lead.AssignedAt = &now
					_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "OWNER_CHANGED", fmt.Sprintf("Owner assigned bulk to user ID %d.", *newOwner), actorUserID)
				} else {
					lead.AssignedAt = nil
					_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "OWNER_CHANGED", "Owner assignment removed bulk.", actorUserID)
				}
				dirty = true
			}
		}

		if dirty {
			err = b.dl.Update(ctx, lead)
			if err != nil {
				report.FailedIDs[strconv.Itoa(int(id))] = fmt.Sprintf("Failed to update lead: %v", err)
				continue
			}
		}

		tagUpdated := false
		var finalTags []string

		if len(req.AddTags) > 0 || len(req.RemoveTags) > 0 {
			currentTags, err := b.dl.GetTags(ctx, id)
			if err != nil {
				report.FailedIDs[strconv.Itoa(int(id))] = fmt.Sprintf("Failed to retrieve current tags: %v", err)
				continue
			}

			tagMap := make(map[string]bool)
			for _, t := range currentTags {
				tagMap[strings.ToLower(t)] = true
			}

			for _, tag := range req.AddTags {
				trimmed := strings.TrimSpace(tag)
				if trimmed == "" {
					continue
				}
				tagMap[strings.ToLower(trimmed)] = true
			}

			for _, tag := range req.RemoveTags {
				trimmed := strings.TrimSpace(tag)
				delete(tagMap, strings.ToLower(trimmed))
			}

			seen := make(map[string]bool)
			for _, tag := range currentTags {
				lower := strings.ToLower(tag)
				if tagMap[lower] && !seen[lower] {
					seen[lower] = true
					finalTags = append(finalTags, tag)
				}
			}
			for _, tag := range req.AddTags {
				trimmed := strings.TrimSpace(tag)
				lower := strings.ToLower(trimmed)
				if trimmed != "" && tagMap[lower] && !seen[lower] {
					seen[lower] = true
					finalTags = append(finalTags, trimmed)
				}
			}

			err = b.dl.SetTags(ctx, id, finalTags)
			if err != nil {
				report.FailedIDs[strconv.Itoa(int(id))] = fmt.Sprintf("Failed to update tags: %v", err)
				continue
			}
			tagUpdated = true
			_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "TAGS_CHANGED", fmt.Sprintf("Tags updated bulk to: %s.", strings.Join(finalTags, ", ")), actorUserID)
		}

		if dirty || tagUpdated {
			_ = b.dl.CreateActivity(ctx, lead.OrgID, "LEAD", lead.ID, "UPDATED", "Lead updated via bulk operation.", actorUserID)
		}

		report.SuccessIDs = append(report.SuccessIDs, id)
	}

	return report, nil
}

// IsBlacklistedOrNonLogisticsEmail detects bank alerts, transactional notifications, security alerts, and non-logistics bots/senders.
func IsBlacklistedOrNonLogisticsEmail(from, subject, body string) bool {
	fromLower := strings.ToLower(strings.TrimSpace(from))
	subjectLower := strings.ToLower(strings.TrimSpace(subject))
	bodyLower := strings.ToLower(strings.TrimSpace(body))

	// 1. Blacklisted Senders & Domains (Bank alerts, automated system emails, newsletters, OTP bots)
	blacklistedSenders := []string{
		"alerts.sbi.bank.in", "sbi.co.in", "hdfcbank.com", "icicibank.com", "axisbank.com",
		"kotak.com", "chase.com", "bankofamerica.com", "cbsalerts", "no-reply", "noreply",
		"do-not-reply", "donotreply", "mailer-daemon", "postmaster", "bounce", "security-alert",
		"info-alert", "notifications@", "marketing@", "promotions@", "newsletter@", "paypal.com",
		"stripe.com", "razorpay.com", "bank.in", "bank.co.in", "subscribe@", "indianexpressonline.in",
		"indianexpress.com", "news@", "newsletters@", "billing@", "receipts@", "alert@", "alerts@",
	}
	for _, pattern := range blacklistedSenders {
		if strings.Contains(fromLower, pattern) {
			return true
		}
	}

	// 2. Blacklisted Subject & Body Phrases (Banking alerts, OTPs, Subscriptions, Financial Receipts)
	blacklistedPhrases := []string{
		"cbssbi alert", "cbs sbi", "debit alert", "credit alert", "hold for inr",
		"deposit ac", "greetings from sbi", "otp", "verification code", "one time password",
		"security alert", "monthly statement", "bank statement", "account balance",
		"transaction alert", "do not reply to this auto generated email", "contact branch",
		"do not reply to this email", "password reset", "login alert", "your invoice from",
		"indian express", "breaking news", "newsletter", "unsubscribe",
	}
	for _, phrase := range blacklistedPhrases {
		if strings.Contains(subjectLower, phrase) || strings.Contains(bodyLower, phrase) {
			return true
		}
	}

	return false
}

// IsLogisticsEmail inspects email text for logistics, shipping, freight, ports, containers, and RFQ terms.
func IsLogisticsEmail(subject, body string) bool {
	if IsBlacklistedOrNonLogisticsEmail("", subject, body) {
		return false
	}

	combined := strings.ToLower(subject + " " + body)
	if strings.TrimSpace(combined) == "" {
		return false
	}

	keywords := []string{
		"rfq", "quote", "quotation", "freight", "shipping", "shipment", "cargo",
		"container", "fcl", "lcl", "origin", "destination", "pol", "pod",
		"weight", "volume", "cbm", "kg", "tons", "tonnes", "incoterm", "incoterms",
		"fob", "cif", "exw", "ddu", "ddp", "port", "vessel", "consignee", "shipper",
		"bill of lading", "lading", "bl", "ocean", "sea freight", "air freight",
		"logistics", "customs", "pallet", "rate", "booking", "hamburg", "mumbai",
		"nhava", "rotterdam", "singapore", "shanghai", "ningbo", "felixstowe",
		"dubai", "jebel ali", "transport", "carrier", "teu", "feu", "drayage", "haulage",
	}

	for _, kw := range keywords {
		if strings.Contains(combined, kw) {
			return true
		}
	}
	return false
}

// ClassifyEmailRelevanceWithAI uses AI Gateway (Google Gemini / OpenAI) to analyze intent, sentiment, and logistics relevance.
func (b *businessLogic) ClassifyEmailRelevanceWithAI(ctx context.Context, from, subject, body string) (bool, string, error) {
	if b.aiGateway == nil {
		return IsLogisticsEmail(subject, body), "KEYWORD_FALLBACK", nil
	}

	prompt := fmt.Sprintf(`You are an AI Email Relevance Classifier for a Global Logistics & Freight Forwarding platform (LogisticsHQ).
Analyze the incoming email below and determine if it is a genuine freight forwarding / shipping / RFQ / cargo transport inquiry or customer reply.

SENDER: %s
SUBJECT: %s
BODY CONTENT:
%s

CLASSIFICATION RULES:
- Return is_logistics_related: true ONLY if the email is about freight, shipping rates, cargo, containers, transport, RFQs, customs, or logistics business.
- Return is_logistics_related: false for bank alerts, OTPs, job applications, newsletters, personal messages, invoice receipts, marketing blasts, IT alerts.

RESPOND STRICTLY IN THIS EXACT JSON FORMAT (no markdown wrappers):
{
  "is_logistics_related": true,
  "intent": "RFQ_INQUIRY" | "GENERAL_LOGISTICS" | "NON_LOGISTICS" | "BANK_ALERT" | "SPAM",
  "sentiment": "POSITIVE" | "NEUTRAL" | "URGENT" | "NEGATIVE",
  "reasoning": "Short 1-sentence reasoning"
}`, from, subject, body)

	aiCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	resp, err := b.aiGateway.ExecutePrompt(aiCtx, prompt)
	if err != nil {
		log.Printf("[Leads Service] AI Email Classification error: %v. Using keyword fallback.", err)
		return IsLogisticsEmail(subject, body), "KEYWORD_FALLBACK_ERROR", nil
	}

	cleanedJSON := strings.TrimSpace(resp)
	if strings.HasPrefix(cleanedJSON, "```") {
		cleanedJSON = strings.TrimPrefix(cleanedJSON, "```json")
		cleanedJSON = strings.TrimPrefix(cleanedJSON, "```")
		cleanedJSON = strings.TrimSuffix(cleanedJSON, "```")
		cleanedJSON = strings.TrimSpace(cleanedJSON)
	}

	var result struct {
		IsLogisticsRelated bool   `json:"is_logistics_related"`
		Intent             string `json:"intent"`
		Sentiment          string `json:"sentiment"`
		Reasoning          string `json:"reasoning"`
	}

	if err := json.Unmarshal([]byte(cleanedJSON), &result); err != nil {
		log.Printf("[Leads Service] AI Classification JSON parse error: %v. Raw: %s", err, resp)
		return IsLogisticsEmail(subject, body), "KEYWORD_FALLBACK_JSON", nil
	}

	log.Printf("[Leads Service] 🤖 AI Classifier Result -> LogisticsRelated: %v, Intent: %s, Sentiment: %s, Reasoning: %q",
		result.IsLogisticsRelated, result.Intent, result.Sentiment, result.Reasoning)

	return result.IsLogisticsRelated, result.Intent, nil
}

func (b *businessLogic) ProcessInboundEmail(ctx context.Context, orgID int32, email InboundEmail) (*LeadInteraction, error) {
	log.Printf("[Leads Service] ProcessInboundEmail called for OrgID: %d, From: %s, RawEmailID: %s", orgID, email.From, email.RawEmailID)

	// 0A. STRICT FRONT-DOOR FILTERING: Drop blacklisted bank alerts, bots, and system emails immediately.
	if IsBlacklistedOrNonLogisticsEmail(email.From, email.Subject, email.Body) {
		log.Printf("[Leads Service] Blacklisted email DROPPED at front door (From: %s, Subject: %q). Skipping lead creation & storage.", email.From, email.Subject)
		return &LeadInteraction{
			OrgID:      int64(orgID),
			Channel:    "EMAIL",
			Direction:  "INBOUND",
			Subject:    email.Subject,
			Content:    email.Body,
			Sender:     email.From,
			Status:     "IGNORED",
			CreatedAt:  time.Now(),
		}, nil
	}

	// 0B. AI-POWERED RELEVANCE & SENTIMENT CLASSIFICATION (Google Gemini / OpenAI)
	isLogisticsRelated, aiIntent, _ := b.ClassifyEmailRelevanceWithAI(ctx, email.From, email.Subject, email.Body)
	if !isLogisticsRelated {
		log.Printf("[Leads Service] 🤖 AI Provider classified email as NON-LOGISTICS (%s). DROPPED at front door (From: %s, Subject: %q).", aiIntent, email.From, email.Subject)
		return &LeadInteraction{
			OrgID:      int64(orgID),
			Channel:    "EMAIL",
			Direction:  "INBOUND",
			Subject:    email.Subject,
			Content:    email.Body,
			Sender:     email.From,
			Status:     "IGNORED",
			CreatedAt:  time.Now(),
		}, nil
	}

	// 0B. Check for duplicate RawEmailID to ensure idempotency.
	if email.RawEmailID != "" {
		existingInter, err := b.dl.GetInteractionByRawEmailID(ctx, orgID, email.RawEmailID)
		if err == nil && existingInter != nil {
			log.Printf("[Leads Service] Duplicate Gmail message ignored: %s", email.RawEmailID)
			existingInter.IsIdempotent = true
			return existingInter, nil
		}
	}

	// 1. Resolve Lead and Parent Interaction ID using strict conversation-matching priorities.
	// We prioritize thread-identity matching over sender email matching to handle multiple independent inquiries correctly.
	threadID := email.ThreadID
	var matchedLeadID int64
	var parentInteractionID *int64
	var priorRFQContext map[string]interface{}

	// Priority 1: Match by Gmail Thread ID (Primary match mechanism for connected mailboxes)
	if threadID != "" {
		existingInteractions, err := b.dl.FindByThreadID(ctx, orgID, threadID)
		if err != nil {
			log.Printf("[Leads Service] FindByThreadID error: %v", err)
		} else if len(existingInteractions) > 0 {
			matchedLeadID = existingInteractions[0].LeadID
			
			// Use the most recent existing interaction in the same Gmail thread as the parent interaction
			mostRecent := existingInteractions[len(existingInteractions)-1]
			parentInteractionID = &mostRecent.ID
		}
	}

	// Priority 2: Match by RFC In-Reply-To header (Fallback if thread ID did not resolve)
	if matchedLeadID == 0 && email.InReplyTo != "" {
		matched, err := b.dl.GetInteractionByRFCMessageID(ctx, orgID, email.InReplyTo)
		if err != nil {
			log.Printf("[Leads Service] GetInteractionByRFCMessageID error: %v", err)
		} else if matched != nil {
			matchedLeadID = matched.LeadID
			parentInteractionID = &matched.ID
		}
	}

	// Priority 3: Match by RFC References header (Fallback if In-Reply-To did not match)
	if matchedLeadID == 0 && email.ReferencesHeader != "" {
		refIDs := strings.Fields(email.ReferencesHeader)
		for _, refID := range refIDs {
			matched, err := b.dl.GetInteractionByRFCMessageID(ctx, orgID, refID)
			if err != nil {
				log.Printf("[Leads Service] GetInteractionByRFCMessageID error: %v", err)
			} else if matched != nil {
				matchedLeadID = matched.LeadID
				parentInteractionID = &matched.ID
				break
			}
		}
	}

	// Unified conversation-scoped prior context retrieval
	if matchedLeadID > 0 {
		var existingInteractions []*LeadInteraction
		var err error
		if threadID != "" {
			existingInteractions, err = b.dl.FindByThreadID(ctx, orgID, threadID)
		}
		if len(existingInteractions) == 0 {
			existingInteractions, err = b.dl.ListInteractions(ctx, int32(orgID), int32(matchedLeadID))
		}
		if err == nil {
			for i := len(existingInteractions) - 1; i >= 0; i-- {
				prev := existingInteractions[i]
				prev.UnmarshalPartialRFQContext()
				if len(prev.PartialRFQContext) > 0 {
					priorRFQContext = prev.PartialRFQContext
					break
				}
			}
		}
	}

	// Resolve the Lead object based on matching results
	var lead *spec.Lead
	var err error
	if matchedLeadID > 0 {
		// Log correct conversation matching tags as specified
		log.Printf("[Leads Service] Existing conversation matched.")
		log.Printf("[Leads Service] Matched Lead ID: %d", matchedLeadID)
		log.Printf("[Leads Service] Creating inbound interaction for existing Lead.")
		lead, err = b.GetLead(ctx, orgID, int32(matchedLeadID))
		if err != nil {
			return nil, fmt.Errorf("fetch matched lead by ID %d: %w", matchedLeadID, err)
		}
	} else {
		// If no conversation match could be found, we create a new Lead. We do not use sender email fallback.
		log.Printf("[Leads Service] No existing conversation found. Creating new Lead.")
		contactName := email.From
		companyName := "Inbound Lead (" + email.From + ")"
		createReq := spec.CreateLeadRequest{
			OrgID:       orgID,
			CompanyName: companyName,
			ContactName: strPtr(contactName),
			Email:       strPtr(email.From),
			Source:      strPtr("EMAIL"),
		}
		lead, err = b.CreateLead(ctx, createReq)
		if err != nil {
			return nil, fmt.Errorf("auto-create lead: %w", err)
		}
	}

	// Fallback: If no thread matches, generate a new conversation thread ID
	if threadID == "" {
		if email.RawEmailID != "" {
			threadID = email.RawEmailID
		} else {
			threadID = fmt.Sprintf("thread_%d_%d", lead.ID, time.Now().UnixNano())
		}
	}

	// 4. Log the interaction as INBOUND.
	receivedAt := time.Now()
	if !email.ReceivedAt.IsZero() {
		receivedAt = email.ReceivedAt
	}

	var mailboxIDPtr *int64
	if email.MailboxID > 0 {
		mID := email.MailboxID
		mailboxIDPtr = &mID
	}

	inter := &LeadInteraction{
		OrgID:               int64(orgID),
		LeadID:              int64(lead.ID),
		MailboxID:           mailboxIDPtr,
		Channel:             "EMAIL",
		Direction:           "INBOUND",
		Subject:             email.Subject,
		Content:             email.Body,
		RawEmailID:          email.RawEmailID,
		ThreadID:            threadID,
		Sentiment:           "NEUTRAL",
		Intent:              "RFQ_REQUEST",
		AIConfidence:        0,
		ParentInteractionID: parentInteractionID,
		RFCMessageID:        email.RFCMessageID,
		InReplyTo:           email.InReplyTo,
		ReferencesHeader:    email.ReferencesHeader,
		Sender:              email.Sender,
		Recipients:          email.Recipients,
		CCRecipients:        email.CCRecipients,
		CreatedAt:           receivedAt,
	}
	err = b.dl.LogInteraction(ctx, inter)
	if err != nil {
		return nil, fmt.Errorf("log lead interaction: %w", err)
	}

	// Resolve backend URL
	backendBaseURL := os.Getenv("GO_BACKEND_URL")
	backendBaseURL = strings.TrimRight(backendBaseURL, "/")
	if backendBaseURL == "" {
		backendBaseURL = "http://localhost:8080"
	}

	// 5. Enqueue the EMAIL_PARSE task for the Python AI sidecar.
	isReply := parentInteractionID != nil
	taskPayload := map[string]interface{}{
		"from":           email.From,
		"subject":        email.Subject,
		"body":           email.Body,
		"message_id":     email.MessageID,
		"thread_id":      threadID,
		"interaction_id": inter.ID,
		"lead_id":        lead.ID,
		"callback_url":   backendBaseURL + "/internal/sales/callback",
		// Thread-awareness fields
		"is_reply":              isReply,
		"parent_interaction_id": parentInteractionID,
		"prior_rfq_context":     priorRFQContext,
	}

	err = b.CreateAITask(
		ctx,
		int64(orgID),
		"LEAD",
		strconv.FormatInt(inter.ID, 10),
		"EMAIL_PARSE",
		taskPayload,
	)
	if err != nil {
		return nil, fmt.Errorf("create AI parse task: %w", err)
	}

	return inter, nil
}

func strPtr(s string) *string {
	return &s
}

func (b *businessLogic) SendClarificationEmail(ctx context.Context, orgID int64, interactionID int64, draftedReply string, summary string) error {
	log.Printf("[Leads Service] SendClarificationEmail called for OrgID: %d, InteractionID: %d", orgID, interactionID)

	// 1. Fetch parent inbound interaction
	parentInter, err := b.dl.GetInteractionByID(ctx, int32(orgID), interactionID)
	if err != nil {
		return fmt.Errorf("get parent interaction: %w", err)
	}

	// 2. Outbound Idempotency Check: check if outbound interaction is already sent for this parent interaction
	existingList, err := b.dl.ListInteractions(ctx, int32(orgID), int32(parentInter.LeadID))
	if err == nil {
		for _, item := range existingList {
			if item.ParentInteractionID != nil && *item.ParentInteractionID == interactionID && item.Direction == "OUTBOUND" {
				log.Printf("[Leads Service] Outbound clarification email already sent for interaction %d (idempotent)", interactionID)
				return nil
			}
		}
	}

	// 3. Mailbox Selection Rules
	mailboxes, err := b.orgRepo.GetConnectedMailboxes(ctx, orgID)
	if err != nil {
		return fmt.Errorf("get connected mailboxes: %w", err)
	}

	var gmailMailboxes []organization.ConnectedMailbox
	for _, mb := range mailboxes {
		if mb.Provider == "GMAIL" && mb.Status != "DISCONNECTED" {
			gmailMailboxes = append(gmailMailboxes, mb)
		}
	}
	if len(gmailMailboxes) == 0 {
		errNoGmail := fmt.Errorf("no active connected Gmail mailbox found for organization %d", orgID)
		_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(parentInter.LeadID), "EMAIL_OUTBOUND_FAILED", errNoGmail.Error(), nil)
		return errNoGmail
	}

	var selectedMailbox *organization.ConnectedMailbox
	// Rule 1: preferentially match the mailbox that received the original email
	for i := range gmailMailboxes {
		mb := &gmailMailboxes[i]
		if strings.Contains(strings.ToLower(parentInter.Recipients), strings.ToLower(mb.Email)) {
			selectedMailbox = mb
			break
		}
	}

	// Rule 2: Fall back to primary Gmail mailbox
	if selectedMailbox == nil {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if mb.IsPrimary {
				selectedMailbox = mb
				break
			}
		}
	}

	// Rule 3: Fall back to first available Gmail mailbox
	if selectedMailbox == nil {
		selectedMailbox = &gmailMailboxes[0]
	}

	// 4. Set up OAuth token update callback
	onTokenUpdate := func(c context.Context, accessToken string, expiry time.Time) error {
		selectedMailbox.TokenExpiry = &expiry
		encKey := os.Getenv(config.EnvMailboxEncryptionKey)
		encAccess, err := crypto.Encrypt(accessToken, encKey)
		if err != nil {
			return err
		}
		selectedMailbox.AccessTokenEncrypted = &encAccess
		return b.orgRepo.UpdateConnectedMailbox(c, selectedMailbox)
	}

	// 5. Send outbound clarification email through Gmail Provider
	gmailMsgID, rfcMsgID, threadID, err := b.gmailProvider.SendEmail(
		ctx,
		selectedMailbox,
		parentInter.Sender,
		parentInter.Subject,
		draftedReply,
		parentInter.ThreadID,
		parentInter.RFCMessageID,
		parentInter.RFCMessageID+" "+parentInter.ReferencesHeader,
		onTokenUpdate,
	)

	if err != nil {
		// Log failed activity in timeline to allow retry
		desc := fmt.Sprintf("Failed to send clarification email for interaction %d: %v", interactionID, err)
		_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(parentInter.LeadID), "EMAIL_OUTBOUND_FAILED", desc, nil)
		return fmt.Errorf("gmail provider send failed: %w", err)
	}

	// 6. Log successful outbound interaction
	outboundInter := &LeadInteraction{
		OrgID:               orgID,
		LeadID:              parentInter.LeadID,
		MailboxID:           &selectedMailbox.ID,
		Channel:             "EMAIL",
		Direction:           "OUTBOUND",
		Subject:             "Re: " + parentInter.Subject,
		Content:             draftedReply,
		RawEmailID:          gmailMsgID,
		ThreadID:            threadID,
		Sentiment:           "NEUTRAL",
		Intent:              "RFQ_REQUEST_INCOMPLETE",
		ParentInteractionID: &interactionID,
		RFCMessageID:        rfcMsgID,
		InReplyTo:           parentInter.RFCMessageID,
		ReferencesHeader:    parentInter.RFCMessageID + " " + parentInter.ReferencesHeader,
		Sender:              selectedMailbox.Email,
		Recipients:          parentInter.Sender,
		CreatedAt:           time.Now(),
	}

	err = b.dl.LogInteraction(ctx, outboundInter)
	if err != nil {
		log.Printf("[Leads Service] Warning: failed to log outbound interaction in DB: %v", err)
	}

	// 7. Log successful activity event on timeline with bullet points
	missingStr := ""
	if idx := strings.Index(summary, "Missing mandatory fields:"); idx != -1 {
		missingStr = summary[idx+len("Missing mandatory fields:"):]
		missingStr = strings.TrimSpace(strings.TrimSuffix(missingStr, "."))
	}

	desc := "Clarification email sent"
	if missingStr != "" {
		desc += "\n\nThe customer was asked to provide:\n"
		fields := strings.Split(missingStr, ",")
		for _, f := range fields {
			desc += fmt.Sprintf("• %s\n", strings.TrimSpace(f))
		}
	} else {
		desc += fmt.Sprintf("\n\n%s", draftedReply)
	}

	err = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(parentInter.LeadID), "EMAIL_OUTBOUND", desc, nil)
	if err != nil {
		log.Printf("[Leads Service] Warning: failed to log outbound activity in DB: %v", err)
	}

	return nil
}

func (b *businessLogic) GetInteractionByID(ctx context.Context, orgID int32, id int64) (*LeadInteraction, error) {
	log.Printf("[Leads Service] GetInteractionByID called for OrgID: %d, InteractionID: %d", orgID, id)
	inter, err := b.dl.GetInteractionByID(ctx, orgID, id)
	if err != nil {
		return nil, svcerror.WrapServiceError(svcerror.ErrInternal, err)
	}
	return inter, nil
}

func (b *businessLogic) ReplyToInteraction(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64, from string, to string, cc string, subject string, body string) (*LeadInteraction, error) {
	log.Printf("[Leads Service] ReplyToInteraction called for OrgID: %d, LeadID: %d, ParentInteractionID: %d, From: %s", orgID, leadID, parentInteractionID, from)

	// =========================================================================
	// 1. Parent Interaction Verification (Supports 0 for fresh outbound emails)
	// =========================================================================
	var parentInter *LeadInteraction
	var parentThreadID string
	var parentInterIDPtr *int64

	if parentInteractionID > 0 {
		var err error
		parentInter, err = b.dl.GetInteractionByID(ctx, int32(orgID), parentInteractionID)
		if err != nil {
			return nil, fmt.Errorf("get parent interaction: %w", err)
		}

		if parentInter.LeadID != leadID || parentInter.OrgID != orgID {
			return nil, fmt.Errorf("interaction/lead organization mismatch")
		}
		parentThreadID = parentInter.ThreadID
		parentInterIDPtr = &parentInteractionID
	} else {
		// Find latest existing interaction for this lead to maintain thread continuity in Gmail & UI
		existingInteractions, err := b.dl.ListInteractions(ctx, int32(orgID), int32(leadID))
		if err == nil && len(existingInteractions) > 0 {
			for i := len(existingInteractions) - 1; i >= 0; i-- {
				if existingInteractions[i].ThreadID != "" {
					parentInter = existingInteractions[i]
					parentThreadID = existingInteractions[i].ThreadID
					parentInterIDPtr = &existingInteractions[i].ID
					break
				}
			}
		}
	}

	// =========================================================================
	// 2. Mailbox Selection Priority Rules
	// =========================================================================
	mailboxes, err := b.orgRepo.GetConnectedMailboxes(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get connected mailboxes: %w", err)
	}

	var gmailMailboxes []organization.ConnectedMailbox
	for _, mb := range mailboxes {
		if mb.Provider == "GMAIL" && mb.Status != "DISCONNECTED" {
			gmailMailboxes = append(gmailMailboxes, mb)
		}
	}
	if len(gmailMailboxes) == 0 {
		return nil, fmt.Errorf("no active connected Gmail mailbox found for organization %d", orgID)
	}

	var selectedMailbox *organization.ConnectedMailbox

	// Rule Priority 0: Explicitly selected From sender address
	if strings.TrimSpace(from) != "" {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if strings.EqualFold(strings.TrimSpace(mb.Email), strings.TrimSpace(from)) {
				selectedMailbox = mb
				break
			}
		}
	}

	// Rule Priority 1: Match the parent interaction's Mailbox ID (if parent exists).
	if selectedMailbox == nil && parentInter != nil && parentInter.MailboxID != nil {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if mb.ID == *parentInter.MailboxID {
				selectedMailbox = mb
				break
			}
		}
	}

	// Rule Priority 2: Match the email recipient address.
	if selectedMailbox == nil && parentInter != nil {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if strings.Contains(strings.ToLower(parentInter.Recipients), strings.ToLower(mb.Email)) {
				selectedMailbox = mb
				break
			}
		}
	}

	// Rule Priority 3: Fall back to the organization's primary connected Gmail mailbox.
	if selectedMailbox == nil {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if mb.IsPrimary {
				selectedMailbox = mb
				break
			}
		}
	}

	// Rule Priority 4: Fall back to the first available connected Gmail mailbox.
	if selectedMailbox == nil {
		selectedMailbox = &gmailMailboxes[0]
	}

	// =========================================================================
	// 3. OAuth Token Refresh Callback Setup
	// =========================================================================
	onTokenUpdate := func(c context.Context, accessToken string, expiry time.Time) error {
		selectedMailbox.TokenExpiry = &expiry
		encKey := os.Getenv(config.EnvMailboxEncryptionKey)
		encAccess, err := crypto.Encrypt(accessToken, encKey)
		if err != nil {
			return err
		}
		selectedMailbox.AccessTokenEncrypted = &encAccess
		return b.orgRepo.UpdateConnectedMailbox(c, selectedMailbox)
	}

	// =========================================================================
	// 4. Threading Headers Preparation (RFC-2822 compliance)
	// =========================================================================
	inReplyTo := ""
	references := ""
	if parentInter != nil {
		inReplyTo = parentInter.RFCMessageID
		references = parentInter.ReferencesHeader
		if references != "" {
			if !strings.Contains(references, parentInter.RFCMessageID) {
				references = references + " " + parentInter.RFCMessageID
			}
		} else {
			references = parentInter.RFCMessageID
		}
	}

	// Generate a temporary local Message-ID for tracking the PENDING state in the database.
	tempOutboundMessageID := fmt.Sprintf("<%d.reply.%s@freel-platform.local>", time.Now().UnixNano(), selectedMailbox.Email)

	// =========================================================================
	// 5. Log outbound interaction as PENDING before sending (anti-loss design)
	// =========================================================================
	outboundInter := &LeadInteraction{
		OrgID:               orgID,
		LeadID:              leadID,
		MailboxID:           &selectedMailbox.ID,
		Channel:             "EMAIL",
		Direction:           "OUTBOUND",
		Subject:             subject,
		Content:             body,
		RawEmailID:          "",
		ThreadID:            parentThreadID,
		Sentiment:           "NEUTRAL",
		Intent:              "RFQ_REQUEST_INCOMPLETE",
		ParentInteractionID: parentInterIDPtr,
		RFCMessageID:        tempOutboundMessageID,
		InReplyTo:           inReplyTo,
		ReferencesHeader:    references,
		Sender:              selectedMailbox.Email,
		Recipients:          to,
		CCRecipients:        cc,
		Status:              "PENDING", // Initiated as PENDING
		CreatedAt:           time.Now(),
	}

	err = b.dl.LogInteraction(ctx, outboundInter)
	if err != nil {
		return nil, fmt.Errorf("failed to log PENDING outbound interaction: %w", err)
	}

	// =========================================================================
	// 6. Call Gmail API Provider to Send the Email
	// =========================================================================
	gmailMsgID, rfcMsgID, threadID, err := b.gmailProvider.SendEmail(
		ctx,
		selectedMailbox,
		to,
		subject,
		body,
		parentThreadID,
		inReplyTo,
		references,
		onTokenUpdate,
	)

	// =========================================================================
	// 7. Resolve Delivery State (SENT or FAILED)
	// =========================================================================
	if err != nil {
		_ = b.dl.UpdateInteractionStatusAndIDs(ctx, orgID, outboundInter.ID, "FAILED", "", tempOutboundMessageID, parentThreadID)
		outboundInter.Status = "FAILED"
		
		desc := fmt.Sprintf("Failed to send manual reply to %s: %v", to, err)
		_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(leadID), "EMAIL_OUTBOUND_FAILED", desc, nil)
		return outboundInter, fmt.Errorf("gmail provider send failed: %w", err)
	}

	// On success:
	err = b.dl.UpdateInteractionStatusAndIDs(ctx, orgID, outboundInter.ID, "SENT", gmailMsgID, rfcMsgID, threadID)
	if err != nil {
		log.Printf("[Leads Service] Warning: failed to update interaction to SENT in DB: %v", err)
	}

	outboundInter.Status = "SENT"
	outboundInter.RawEmailID = gmailMsgID
	outboundInter.RFCMessageID = rfcMsgID
	outboundInter.ThreadID = threadID

	// Log successful activity event on timeline
	desc := fmt.Sprintf("Email reply sent to %s: \"%s\"", to, subject)
	_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(parentInter.LeadID), "EMAIL_OUTBOUND_SENT", desc, nil)

	return outboundInter, nil
}

func (b *businessLogic) RetryEmailInteraction(ctx context.Context, orgID int64, leadID int64, interactionID int64) (*LeadInteraction, error) {
	log.Printf("[Leads Service] RetryEmailInteraction called for OrgID: %d, LeadID: %d, InteractionID: %d", orgID, leadID, interactionID)

	// 1. Fetch target interaction to validate first
	inter, err := b.dl.GetInteractionByID(ctx, int32(orgID), interactionID)
	if err != nil {
		return nil, fmt.Errorf("get interaction: %w", err)
	}

	// Validate org, lead, and retry eligibility (must be OUTBOUND and FAILED)
	if inter.LeadID != leadID || inter.OrgID != orgID {
		return nil, fmt.Errorf("interaction/lead organization mismatch")
	}
	if inter.Direction != "OUTBOUND" || inter.Status != "FAILED" {
		return nil, fmt.Errorf("interaction is not eligible for retry")
	}

	// 2. Atomic Database State Lock Acquisition
	// This atomically transitions status FAILED -> PENDING, retry_count++, and last_retry_at = NOW()
	// only if status was indeed FAILED. If RowsAffected is 0, concurrent retry requests are blocked.
	locked, err := b.dl.LockInteractionForRetry(ctx, orgID, interactionID)
	if err != nil {
		return nil, err
	}
	if !locked {
		return nil, fmt.Errorf("retry lock already in progress or interaction state changed")
	}

	// Retrieve updated interaction values for execution
	inter, _ = b.dl.GetInteractionByID(ctx, int32(orgID), interactionID)

	// Log retry start to activity timeline
	_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(leadID), "EMAIL_RETRY_STARTED", fmt.Sprintf("Retrying failed email to %s: \"%s\"", inter.Recipients, inter.Subject), nil)

	// 3. Resolve Mailbox Selection
	mailboxes, err := b.orgRepo.GetConnectedMailboxes(ctx, orgID)
	if err != nil {
		safeErr := "Could not send this email. Please try again."
		_ = b.dl.UpdateInteractionRetry(ctx, orgID, interactionID, "FAILED", &safeErr, false, "", "", inter.ThreadID)
		return inter, fmt.Errorf("get connected mailboxes: %w", err)
	}

	var gmailMailboxes []organization.ConnectedMailbox
	for _, mb := range mailboxes {
		if mb.Provider == "GMAIL" && mb.Status != "DISCONNECTED" {
			gmailMailboxes = append(gmailMailboxes, mb)
		}
	}

	var selectedMailbox *organization.ConnectedMailbox

	// Rule Priority 1: Match the original mailbox_id
	if inter.MailboxID != nil {
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if mb.ID == *inter.MailboxID {
				selectedMailbox = mb
				break
			}
		}
	}

	// Fallbacks if the original mailbox is unavailable
	if selectedMailbox == nil {
		// Rule Priority 2: Match the recipient address matching parent's sender
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if strings.Contains(strings.ToLower(inter.Sender), strings.ToLower(mb.Email)) {
				selectedMailbox = mb
				break
			}
		}
	}
	if selectedMailbox == nil {
		// Rule Priority 3: Fall back to primary connected GMAIL mailbox
		for i := range gmailMailboxes {
			mb := &gmailMailboxes[i]
			if mb.IsPrimary {
				selectedMailbox = mb
				break
			}
		}
	}
	if selectedMailbox == nil && len(gmailMailboxes) > 0 {
		// Rule Priority 4: Fall back to first available GMAIL mailbox
		selectedMailbox = &gmailMailboxes[0]
	}

	if selectedMailbox == nil {
		safeErr := "No active connected Gmail mailbox found for retry."
		_ = b.dl.UpdateInteractionRetry(ctx, orgID, interactionID, "FAILED", &safeErr, false, "", "", inter.ThreadID)
		return inter, fmt.Errorf("no connected Gmail mailbox found")
	}

	// 4. Update the interaction mailbox association in case we had to use a fallback
	if inter.MailboxID == nil || *inter.MailboxID != selectedMailbox.ID {
		inter.MailboxID = &selectedMailbox.ID
		query := `UPDATE lead_interactions SET mailbox_id = ? WHERE id = ?`
		_, _ = b.dl.(*dataLayer).db.ExecContext(ctx, query, selectedMailbox.ID, interactionID)
	}

	// 5. Token Refresh Callback
	onTokenUpdate := func(c context.Context, accessToken string, expiry time.Time) error {
		selectedMailbox.TokenExpiry = &expiry
		encKey := os.Getenv(config.EnvMailboxEncryptionKey)
		encAccess, err := crypto.Encrypt(accessToken, encKey)
		if err != nil {
			return err
		}
		selectedMailbox.AccessTokenEncrypted = &encAccess
		return b.orgRepo.UpdateConnectedMailbox(c, selectedMailbox)
	}

	// 6. Execute Send via Gmail
	gmailMsgID, rfcMsgID, threadID, err := b.gmailProvider.SendEmail(
		ctx,
		selectedMailbox,
		inter.Recipients,
		inter.Subject,
		inter.Content,
		inter.ThreadID,
		inter.InReplyTo,
		inter.ReferencesHeader,
		onTokenUpdate,
	)

	// 7. Resolve Delivery State
	if err != nil {
		// Sanitize the error message to avoid exposing raw token/Gmail details to client
		safeErr := "Could not send this email. Please try again."
		log.Printf("[Leads Service] Retry send failed for interaction %d: %v", interactionID, err)
		
		_ = b.dl.UpdateInteractionRetry(ctx, orgID, interactionID, "FAILED", &safeErr, false, "", "", inter.ThreadID)
		_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(leadID), "EMAIL_SEND_FAILED", fmt.Sprintf("Failed manual reply retry to %s: %s", inter.Recipients, safeErr), nil)
		
		inter, _ = b.dl.GetInteractionByID(ctx, int32(orgID), interactionID)
		return inter, err
	}

	// Update to SENT and reset the error field
	err = b.dl.UpdateInteractionRetry(ctx, orgID, interactionID, "SENT", nil, false, gmailMsgID, rfcMsgID, threadID)
	if err != nil {
		log.Printf("[Leads Service] Failed to finalize retry status in DB: %v", err)
	}

	_ = b.dl.CreateActivity(ctx, int32(orgID), "LEAD", int32(leadID), "EMAIL_SENT", fmt.Sprintf("Email reply sent to %s: \"%s\" (via: %s)", inter.Recipients, inter.Subject, selectedMailbox.Email), nil)

	inter, _ = b.dl.GetInteractionByID(ctx, int32(orgID), interactionID)
	return inter, nil
}

func (b *businessLogic) GetDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) (*LeadEmailDraft, error) {
	log.Printf("[Leads Service] GetDraft called for OrgID: %d, LeadID: %d, ParentInteractionID: %d", orgID, leadID, parentInteractionID)
	draft, err := b.dl.GetDraft(ctx, orgID, leadID, parentInteractionID)
	if err != nil {
		return nil, err
	}
	if draft == nil {
		return nil, nil
	}
	if draft.MailboxID != nil {
		mailboxes, err := b.orgRepo.GetConnectedMailboxes(ctx, orgID)
		if err == nil {
			for i := range mailboxes {
				mb := &mailboxes[i]
				if mb.ID == *draft.MailboxID {
					draft.From = mb.Email
					break
				}
			}
		}
	}
	return draft, nil
}

func (b *businessLogic) SaveDraft(ctx context.Context, draft *LeadEmailDraft) error {
	log.Printf("[Leads Service] SaveDraft called for OrgID: %d, LeadID: %d, ParentInteractionID: %d, From: %s", draft.OrgID, draft.LeadID, draft.ParentInteractionID, draft.From)
	
	if draft.MailboxID == nil && strings.TrimSpace(draft.From) != "" {
		mailboxes, err := b.orgRepo.GetConnectedMailboxes(ctx, draft.OrgID)
		if err == nil {
			for i := range mailboxes {
				mb := &mailboxes[i]
				if strings.EqualFold(strings.TrimSpace(mb.Email), strings.TrimSpace(draft.From)) {
					mbIDCopy := mb.ID
					draft.MailboxID = &mbIDCopy
					break
				}
			}
		}
	}
	
	return b.dl.SaveDraft(ctx, draft)
}

func (b *businessLogic) DeleteDraft(ctx context.Context, orgID int64, leadID int64, parentInteractionID int64) error {
	log.Printf("[Leads Service] DeleteDraft called for OrgID: %d, LeadID: %d, ParentInteractionID: %d", orgID, leadID, parentInteractionID)
	return b.dl.DeleteDraft(ctx, orgID, leadID, parentInteractionID)
}
