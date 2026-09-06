package approvals

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
)

type Service interface {
	ListApprovals(ctx context.Context, orgID int64) ([]*ApprovalRequest, error)
	GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error)
	CreateApproval(ctx context.Context, orgID int64, input *CreateApprovalInput, actorName string) (*ApprovalRequest, error)
	ApproveRequest(ctx context.Context, orgID int64, id int64, actorName string, notes string) (*ApprovalRequest, error)
	RejectRequest(ctx context.Context, orgID int64, id int64, actorName string, reason string, notes string) (*ApprovalRequest, error)
	GetStats(ctx context.Context, orgID int64) (*ApprovalStats, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) ListApprovals(ctx context.Context, orgID int64) ([]*ApprovalRequest, error) {
	return s.repo.GetApprovalsByOrg(ctx, orgID)
}

func (s *service) GetApprovalByID(ctx context.Context, orgID int64, id int64) (*ApprovalRequest, error) {
	return s.repo.GetApprovalByID(ctx, orgID, id)
}

func (s *service) CreateApproval(ctx context.Context, orgID int64, input *CreateApprovalInput, actorName string) (*ApprovalRequest, error) {
	if input.Title == "" {
		return nil, fmt.Errorf("title is required")
	}

	category := strings.ToUpper(input.Category)
	if category == "" {
		category = "DOCUMENTS"
	}

	appType := input.Type
	if appType == "" {
		if category == "DOCUMENTS" {
			appType = "Document Approval"
		} else if category == "COMMERCIAL" {
			appType = "Commercial Approval"
		} else if category == "OPERATIONS" {
			appType = "Operations Approval"
		} else if category == "FINANCE" {
			appType = "Finance Approval"
		} else {
			appType = "General Approval"
		}
	}

	priority := strings.ToUpper(input.Priority)
	if priority == "" {
		priority = "MEDIUM"
	}

	reqCode := fmt.Sprintf("%s-APP-%d", category[:3], 1000+rand.Intn(9000))

	requester := input.RequestedByName
	if requester == "" {
		requester = actorName
	}
	if requester == "" {
		requester = "Varun Kanade"
	}

	avatar := "VK"
	parts := strings.Fields(requester)
	if len(parts) >= 2 {
		avatar = fmt.Sprintf("%c%c", parts[0][0], parts[1][0])
	} else if len(parts) == 1 {
		avatar = fmt.Sprintf("%c", parts[0][0])
	}

	var dueDatePtr *time.Time
	dueText := "7 days left"
	if input.DueDate != "" {
		if parsed, err := time.Parse("2006-01-02", input.DueDate); err == nil {
			dueDatePtr = &parsed
		} else if parsed, err := time.Parse(time.RFC3339, input.DueDate); err == nil {
			dueDatePtr = &parsed
		}
	}

	req := &ApprovalRequest{
		OrgID:           orgID,
		RequestCode:     reqCode,
		Title:           input.Title,
		Category:        category,
		Type:            appType,
		Status:          "Pending",
		Priority:        priority,
		RequestedByName: requester,
		Avatar:          &avatar,
		DueDate:         dueDatePtr,
		DueText:         &dueText,
	}

	if input.RelatedRef != "" {
		req.RelatedRef = &input.RelatedRef
	}
	if input.RelatedEntityType != "" {
		req.RelatedEntityType = &input.RelatedEntityType
	}
	if input.RelatedEntityID > 0 {
		req.RelatedEntityID = &input.RelatedEntityID
	}
	if input.CustomerName != "" {
		req.CustomerName = &input.CustomerName
	}
	if input.CustomerID > 0 {
		req.CustomerID = &input.CustomerID
	}
	if input.ShipmentID > 0 {
		req.ShipmentID = &input.ShipmentID
	}
	if input.DocumentID > 0 {
		req.DocumentID = &input.DocumentID
	}
	if input.BookingID > 0 {
		req.BookingID = &input.BookingID
	}
	if input.Department != "" {
		req.Department = &input.Department
	} else {
		dept := "Operations"
		req.Department = &dept
	}
	if input.Description != "" {
		req.Description = &input.Description
	}

	err := s.repo.CreateApproval(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create approval request: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    requester,
		Action:       domain.ActionCreate,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", req.ID),
		ResourceName: req.RequestCode,
		Description:  fmt.Sprintf("Created approval request %s (%s)", req.RequestCode, req.Title),
		Result:       domain.ResultSuccess,
	})

	return req, nil
}

func (s *service) ApproveRequest(ctx context.Context, orgID int64, id int64, actorName string, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "Varun Kanade"
	}
	app, err := s.repo.UpdateApprovalStatus(ctx, orgID, id, "Approved", actorName, notes, "")
	if err != nil {
		return nil, err
	}

	// Cross-module sync with customer_invoices if this approval is for an Invoice
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "INVOICE" && app.RelatedEntityID != nil {
		invID := *app.RelatedEntityID
		_ = s.repo.(*repository).db.SelectContext // Or direct Exec
		_, _ = s.repo.(*repository).db.ExecContext(ctx, "UPDATE customer_invoices SET status = 'Issued', updated_at = NOW() WHERE id = ? AND org_id = ?", invID, orgID)
		_, _ = s.repo.(*repository).db.ExecContext(ctx, "INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at) VALUES (?, ?, 'Invoice Approved', 'Manager approved request; status changed to Issued', ?, NOW())", orgID, invID, actorName)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       domain.ActionApprove,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Approved request %s (%s)", app.RequestCode, app.Title),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) RejectRequest(ctx context.Context, orgID int64, id int64, actorName string, reason string, notes string) (*ApprovalRequest, error) {
	if actorName == "" {
		actorName = "Varun Kanade"
	}
	app, err := s.repo.UpdateApprovalStatus(ctx, orgID, id, "Rejected", actorName, notes, reason)
	if err != nil {
		return nil, err
	}

	// Cross-module sync with customer_invoices if this approval is for an Invoice
	if app != nil && app.RelatedEntityType != nil && strings.ToUpper(*app.RelatedEntityType) == "INVOICE" && app.RelatedEntityID != nil {
		invID := *app.RelatedEntityID
		_, _ = s.repo.(*repository).db.ExecContext(ctx, "UPDATE customer_invoices SET status = 'Draft', updated_at = NOW() WHERE id = ? AND org_id = ?", invID, orgID)
		desc := "Approval request rejected by manager; status reverted to Draft"
		if reason != "" {
			desc += ". Reason: " + reason
		}
		_, _ = s.repo.(*repository).db.ExecContext(ctx, "INSERT INTO customer_invoice_history (org_id, invoice_id, title, description, user_name, created_at) VALUES (?, ?, 'Approval Rejected', ?, ?, NOW())", orgID, invID, desc, actorName)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		ActorName:    actorName,
		Action:       domain.ActionReject,
		Module:       domain.ModuleApprovals,
		ResourceType: "APPROVAL",
		ResourceID:   fmt.Sprintf("%d", id),
		ResourceName: app.RequestCode,
		Description:  fmt.Sprintf("Rejected request %s. Reason: %s", app.RequestCode, reason),
		Result:       domain.ResultSuccess,
	})

	return app, nil
}

func (s *service) GetStats(ctx context.Context, orgID int64) (*ApprovalStats, error) {
	return s.repo.GetApprovalStats(ctx, orgID)
}
