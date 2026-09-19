package actions

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/approvals"
	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/rbac"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Service interface {
	Execute(ctx context.Context, req ActionExecutionRequest) (*ActionExecutionResponse, error)
	ListActions() []ActionDescriptor
	GetRegistry() Registry
	SetApprovalsService(svc approvals.Service)
}

type defaultService struct {
	registry     Registry
	store        IdempotencyStore
	rbacSvc      rbac.Service
	db           *sqlx.DB
	approvalsSvc approvals.Service
}

func NewService(registry Registry, store IdempotencyStore, rbacSvc rbac.Service, db *sqlx.DB) Service {
	return &defaultService{
		registry: registry,
		store:    store,
		rbacSvc:  rbacSvc,
		db:       db,
	}
}

func (s *defaultService) SetApprovalsService(svc approvals.Service) {
	s.approvalsSvc = svc
}

func (s *defaultService) GetRegistry() Registry {
	return s.registry
}

func (s *defaultService) ListActions() []ActionDescriptor {
	registered := s.registry.ListActions()
	descriptors := make([]ActionDescriptor, 0, len(registered))

	for _, a := range registered {
		res, act := a.RequiredPermission()
		perm := ""
		if res != "" && act != "" {
			perm = res + "." + act
		}
		descriptors = append(descriptors, ActionDescriptor{
			Name:                 a.Name(),
			Module:               a.Module(),
			Description:          a.Description(),
			Category:             a.Category(),
			RequiredPermission:   perm,
			RequiresConfirmation: a.RequiresConfirmation(),
			InputSchema:          a.InputSchema(),
		})
	}
	return descriptors
}

func (s *defaultService) Execute(ctx context.Context, req ActionExecutionRequest) (*ActionExecutionResponse, error) {
	correlationID := uuid.New().String()

	// 1. Validate mandatory fields
	if strings.TrimSpace(req.ActionName) == "" {
		return &ActionExecutionResponse{
			Success:       false,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: "Validation", Message: "action_name is required"},
		}, nil
	}

	if req.OrgID <= 0 {
		return &ActionExecutionResponse{
			Success:       false,
			ActionName:    req.ActionName,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: "Validation", Message: "org_id must be a positive integer"},
		}, nil
	}

	if s.db != nil {
		var orgCount int
		_ = s.db.GetContext(ctx, &orgCount, "SELECT count(*) FROM organizations WHERE id = ?", req.OrgID)
		if orgCount == 0 {
			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error:         &ActionError{Type: "NotFound", Message: fmt.Sprintf("Organization #%d not found", req.OrgID)},
			}, nil
		}
	}

	if req.ActorType == "" {
		req.ActorType = ActorTypeAIAgent
	}

	// 2. Resolve Action from Registry
	action, err := s.registry.GetAction(req.ActionName)
	if err != nil {
		return &ActionExecutionResponse{
			Success:       false,
			ActionName:    req.ActionName,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: "NotFound", Message: fmt.Sprintf("Action '%s' is not registered", req.ActionName)},
		}, nil
	}

	// 3. Resolve Actor Role
	var roleName string
	if req.ActingUserID > 0 {
		if s.db != nil {
			var userRole string
			query := `
				SELECT r.name 
				FROM org_members om 
				JOIN roles r ON om.role_id = r.id 
				WHERE om.user_id = ? AND om.org_id = ? AND om.status = 'ACTIVE' 
				LIMIT 1
			`
			err := s.db.GetContext(ctx, &userRole, query, req.ActingUserID, req.OrgID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					return &ActionExecutionResponse{
						Success:       false,
						ActionName:    req.ActionName,
						CorrelationID: correlationID,
						Error:         &ActionError{Type: "Unauthorized", Message: fmt.Sprintf("Acting user #%d not found in organization #%d", req.ActingUserID, req.OrgID)},
					}, nil
				}
				return &ActionExecutionResponse{
					Success:       false,
					ActionName:    req.ActionName,
					CorrelationID: correlationID,
					Error:         &ActionError{Type: "DatabaseError", Message: "Failed to resolve acting user: " + err.Error()},
				}, nil
			}
			roleName = userRole
		} else {
			roleName = rbac.RoleSuperAdmin
		}
	} else {
		// AI Agent acting autonomously: resolve canonical role by domain module
		switch action.Module() {
		case "pricing":
			roleName = rbac.RolePricing
		case "shipments", "operations":
			roleName = rbac.RoleOperations
		case "sales", "leads", "outreach":
			roleName = rbac.RoleSales
		case "contracts", "compliance", "documents":
			roleName = rbac.RoleDocumentation
		case "finance", "billing":
			roleName = rbac.RoleFinance
		default:
			roleName = rbac.RoleSuperAdmin
		}
	}

	// 4. Verify RBAC Permissions
	res, act := action.RequiredPermission()
	if res != "" && act != "" && s.rbacSvc != nil {
		hasPerm, err := s.rbacSvc.HasPermission(ctx, roleName, req.OrgID, res, act)
		if (!hasPerm || err != nil) && (roleName == rbac.RoleSuperAdmin || roleName == "ADMIN") {
			altRole := "ADMIN"
			if roleName == "ADMIN" {
				altRole = rbac.RoleSuperAdmin
			}
			hasPerm, err = s.rbacSvc.HasPermission(ctx, altRole, req.OrgID, res, act)
		}
		if err != nil || !hasPerm {
			// Record unauthorized attempt in audit log
			_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
				OrgID:        req.OrgID,
				ActorType:    string(req.ActorType),
				ActorName:    fmt.Sprintf("AI Agent (%s)", req.ActionName),
				ActorRole:    roleName,
				Action:       "UNAUTHORIZED_ATTEMPT",
				Module:       strings.ToUpper(action.Module()),
				ResourceType: "ACTION",
				ResourceID:   req.ActionName,
				Description:  fmt.Sprintf("Role '%s' denied permission '%s.%s' on action '%s'", roleName, res, act, req.ActionName),
				Result:       domain.ResultFailed,
				ErrorMessage: "Forbidden: Missing required RBAC permission",
			})

			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error: &ActionError{
					Type:    "Unauthorized",
					Message: fmt.Sprintf("Role '%s' lacks required permission '%s.%s' for action '%s'", roleName, res, act, req.ActionName),
				},
			}, nil
		}
	}

	// 5. Check Idempotency Conflict and Stored Results
	if req.IdempotencyKey != "" && s.store != nil {
		conflict, err := s.store.CheckConflict(ctx, req.OrgID, req.IdempotencyKey, req.ActionName)
		if err != nil {
			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error:         &ActionError{Type: "DatabaseError", Message: "Idempotency check error: " + err.Error()},
			}, nil
		}
		if conflict {
			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error:         &ActionError{Type: "Conflict", Message: fmt.Sprintf("Idempotency key '%s' was already used for a different action", req.IdempotencyKey)},
			}, nil
		}

		completed, storedResult, err := s.store.Start(ctx, req.OrgID, req.IdempotencyKey)
		if err != nil {
			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error:         &ActionError{Type: "DatabaseError", Message: "Idempotency start error: " + err.Error()},
			}, nil
		}
		if completed && storedResult != nil {
			return &ActionExecutionResponse{
				Success:          storedResult.Success,
				ActionName:       req.ActionName,
				CorrelationID:    correlationID,
				IdempotentReplay: true,
				Data:             storedResult.Data,
				Error:            storedResult.Error,
			}, nil
		}
	}

	// 6. Confirmation Gate: Protect High-Risk Actions
	if action.RequiresConfirmation() {
		// AI Agents can NEVER self-confirm high-risk actions
		if req.ActorType == ActorTypeAIAgent && req.IsConfirmed {
			return &ActionExecutionResponse{
				Success:       false,
				ActionName:    req.ActionName,
				CorrelationID: correlationID,
				Error: &ActionError{
					Type:    "Unauthorized",
					Message: fmt.Sprintf("AI agent cannot self-confirm high-risk action '%s'; explicit human approval required", req.ActionName),
				},
			}, nil
		}

		if req.IsConfirmed {
			// When confirmed without a direct human user, verify that an approved record actually exists in approval_requests
			if s.db != nil && (req.ActingUserID == 0 || req.ActorType == ActorTypeAIAgent) {
				var approvedCount int
				approvalRef := strings.TrimPrefix(req.IdempotencyKey, "approval-")
				query := `
					SELECT count(*) 
					FROM approval_requests 
					WHERE org_id = ? AND action_name = ? AND status = 'APPROVED'
					  AND (approval_reference = ? OR (thread_id = ? AND ? != '') OR (idempotency_key = ? AND ? != ''))
				`
				_ = s.db.GetContext(ctx, &approvedCount, query, req.OrgID, req.ActionName, approvalRef, req.ThreadID, req.ThreadID, req.IdempotencyKey, req.IdempotencyKey)
				if approvedCount == 0 {
					return &ActionExecutionResponse{
						Success:       false,
						ActionName:    req.ActionName,
						CorrelationID: correlationID,
						Error: &ActionError{
							Type:    "Unauthorized",
							Message: fmt.Sprintf("High-risk action '%s' claims confirmation, but no valid approved record was found in approval_requests", req.ActionName),
						},
					}, nil
				}
			}
		} else {
			approvalRef := fmt.Sprintf("approval-%s-%s", req.ActionName, correlationID[:8])

		// Log proposal in universal audit log
		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        req.OrgID,
			ActorType:    string(req.ActorType),
			ActorName:    fmt.Sprintf("AI Agent (%s)", req.ActionName),
			ActorRole:    roleName,
			Action:       "PROPOSE_ACTION",
			Module:       strings.ToUpper(action.Module()),
			ResourceType: "APPROVAL_REQUEST",
			ResourceID:   approvalRef,
			Description:  fmt.Sprintf("High-risk action %s proposed, awaiting human confirmation", req.ActionName),
			Result:       "PENDING",
		})

		// Bridge to canonical approval_requests table without duplication
		if s.approvalsSvc != nil {
			inputBytes, _ := json.Marshal(req.Input)
			appr, err := s.approvalsSvc.ProposeAIApproval(ctx, req.OrgID, &approvals.ProposeAIApprovalInput{
				Title:              fmt.Sprintf("AI Action Confirmation: %s", req.ActionName),
				Category:           strings.ToUpper(action.Module()),
				Type:               fmt.Sprintf("AI Action: %s", req.ActionName),
				Priority:           "HIGH",
				ActorType:          string(req.ActorType),
				Source:             req.Source,
				ActionName:         req.ActionName,
				RiskLevel:          string(action.Category()),
				RequiredPermission: res + ":" + act,
				ThreadID:           req.ThreadID,
				ProposedPayload:    string(inputBytes),
				ApprovalReference:  approvalRef,
				CorrelationID:      correlationID,
				IdempotencyKey:     req.IdempotencyKey,
				ExpiresInHours:     48,
			})
			if err == nil && appr != nil && appr.ApprovalReference != nil && *appr.ApprovalReference != "" {
				approvalRef = *appr.ApprovalReference
			}
		}

		return &ActionExecutionResponse{
			Success:              false,
			ActionName:           req.ActionName,
			CorrelationID:        correlationID,
			ConfirmationRequired: true,
			ApprovalReference:    approvalRef,
			Error: &ActionError{
				Type:    "ConfirmationRequired",
				Message: fmt.Sprintf("Action '%s' is high-risk and requires explicit human confirmation before execution", req.ActionName),
			},
		}, nil
	}
	}

	// 7. Assemble ActionContext
	actionCtx := NewActionContext(ctx, req.OrgID, req.ActingUserID, req.ActorType, correlationID)
	actionCtx.WithMetadata(req.Source, req.TaskID, req.ThreadID, req.IsConfirmed)
	if req.IdempotencyKey != "" {
		actionCtx.WithIdempotencyKey(req.IdempotencyKey)
	}

	// 8. Marshal and Execute Action
	inputBytes, err := json.Marshal(req.Input)
	if err != nil {
		return &ActionExecutionResponse{
			Success:       false,
			ActionName:    req.ActionName,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: "Validation", Message: "Failed to serialize input: " + err.Error()},
		}, nil
	}

	result, err := action.Execute(actionCtx, inputBytes)
	if err != nil {
		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        req.OrgID,
			ActorType:    string(req.ActorType),
			ActorName:    fmt.Sprintf("AI Agent (%s)", req.ActionName),
			ActorRole:    roleName,
			Action:       string(action.Category()),
			Module:       strings.ToUpper(action.Module()),
			ResourceType: "ACTION",
			ResourceID:   req.ActionName,
			Description:  fmt.Sprintf("Action %s execution failed: %v", req.ActionName, err),
			Result:       domain.ResultFailed,
			ErrorMessage: err.Error(),
		})

		return &ActionExecutionResponse{
			Success:       false,
			ActionName:    req.ActionName,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: "ExecutionError", Message: err.Error()},
		}, nil
	}

	if !result.Success {
		errType := "BusinessRule"
		errMsg := "Action execution did not succeed"
		if result.Error != nil {
			errType = result.Error.Type
			errMsg = result.Error.Message
		}

		_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
			OrgID:        req.OrgID,
			ActorType:    string(req.ActorType),
			ActorName:    fmt.Sprintf("AI Agent (%s)", req.ActionName),
			ActorRole:    roleName,
			Action:       string(action.Category()),
			Module:       strings.ToUpper(action.Module()),
			ResourceType: result.ResourceType,
			ResourceID:   result.ResourceID,
			Description:  fmt.Sprintf("Action %s rejected: %s", req.ActionName, errMsg),
			Result:       domain.ResultFailed,
			ErrorMessage: errMsg,
		})

		return &ActionExecutionResponse{
			Success:       false,
			ActionName:    req.ActionName,
			CorrelationID: correlationID,
			Error:         &ActionError{Type: errType, Message: errMsg},
		}, nil
	}

	// 9. Persist Idempotency on Success
	if req.IdempotencyKey != "" && s.store != nil {
		_ = s.store.Complete(ctx, req.OrgID, req.IdempotencyKey, result, 24*time.Hour)
	}

	// 10. Record Universal Audit Log
	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        req.OrgID,
		ActorType:    string(req.ActorType),
		ActorName:    fmt.Sprintf("AI Agent (%s)", req.ActionName),
		ActorRole:    roleName,
		Action:       string(action.Category()),
		Module:       strings.ToUpper(action.Module()),
		ResourceType: result.ResourceType,
		ResourceID:   result.ResourceID,
		Description:  result.Summary,
		Result:       domain.ResultSuccess,
	})

	return &ActionExecutionResponse{
		Success:       true,
		ActionName:    req.ActionName,
		CorrelationID: correlationID,
		Data:          result.Data,
	}, nil
}
