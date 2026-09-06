package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/audit/repository"
	"github.com/freel/backend/internal/middleware"
	"github.com/jmoiron/sqlx"
)

// Service provides high-level business logic for recording and inspecting universal audit logs.
type Service interface {
	Record(ctx context.Context, params domain.CreateAuditLogParams) (*domain.AuditLog, error)
	RecordAsync(ctx context.Context, params domain.CreateAuditLogParams)
	List(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogListResponse, error)
	GetByID(ctx context.Context, orgID int64, id int64) (*domain.AuditLog, error)
}

type auditService struct {
	repo repository.Repository
	db   *sqlx.DB
}

// NewService creates a new universal AuditLog service instance.
func NewService(repo repository.Repository, db *sqlx.DB) Service {
	return &auditService{
		repo: repo,
		db:   db,
	}
}

// generateDefaultDescription constructs a clean, human-readable description for business users.
func generateDefaultDescription(module, action, resourceType, resourceID, resourceName string) string {
	resLabel := resourceType
	if resLabel == "" {
		resLabel = "Record"
	}
	resLabel = strings.Title(strings.ToLower(strings.ReplaceAll(resLabel, "_", " ")))

	target := resourceID
	if resourceName != "" {
		target = resourceName
	}

	switch action {
	case domain.ActionCreate:
		if target != "" {
			return fmt.Sprintf("%s %s created", resLabel, target)
		}
		return fmt.Sprintf("%s created", resLabel)
	case domain.ActionUpdate:
		if target != "" {
			return fmt.Sprintf("%s %s updated", resLabel, target)
		}
		return fmt.Sprintf("%s updated", resLabel)
	case domain.ActionDelete:
		if target != "" {
			return fmt.Sprintf("%s %s deleted", resLabel, target)
		}
		return fmt.Sprintf("%s deleted", resLabel)
	case domain.ActionLogin:
		return "User authentication login successful"
	case domain.ActionLogout:
		return "User logged out"
	case domain.ActionLoginFailed:
		return "User authentication login failed"
	case domain.ActionConnect:
		if target != "" {
			return fmt.Sprintf("Carrier %s integration connected", target)
		}
		return "Carrier integration connected"
	case domain.ActionDisconnect:
		if target != "" {
			return fmt.Sprintf("Carrier %s integration disconnected", target)
		}
		return "Carrier integration disconnected"
	case domain.ActionEnable:
		return fmt.Sprintf("%s %s enabled", resLabel, target)
	case domain.ActionDisable:
		return fmt.Sprintf("%s %s disabled", resLabel, target)
	case domain.ActionSync:
		return fmt.Sprintf("%s synchronization completed", resLabel)
	case domain.ActionApprove:
		return fmt.Sprintf("%s %s approved", resLabel, target)
	case domain.ActionReject:
		return fmt.Sprintf("%s %s rejected", resLabel, target)
	case domain.ActionSend:
		return fmt.Sprintf("%s %s sent", resLabel, target)
	case domain.ActionFollowUpSent:
		return fmt.Sprintf("Follow-up email sent for %s", target)
	case domain.ActionPaymentRecorded:
		return fmt.Sprintf("Payment recorded for invoice %s", target)
	case domain.ActionRoleChanged:
		return fmt.Sprintf("User role updated for %s", target)
	case domain.ActionPermissionChanged:
		return fmt.Sprintf("Permissions updated for role %s", target)
	case domain.ActionInvite:
		return fmt.Sprintf("Team member invitation sent to %s", target)
	case domain.ActionExport:
		return fmt.Sprintf("%s data exported", resLabel)
	default:
		return fmt.Sprintf("%s %s performed", resLabel, strings.ToLower(action))
	}
}

// lookupUserName resolves the user's display name from the database if not provided in context.
func (s *auditService) lookupUserName(ctx context.Context, userID int64) (string, string) {
	if s.db == nil || userID <= 0 {
		return "", ""
	}

	var row struct {
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		Email     string `db:"email"`
		Role      string `db:"role"`
	}

	query := `SELECT first_name, last_name, email, role FROM users WHERE id = ? LIMIT 1`
	if err := s.db.GetContext(ctx, &row, query, userID); err == nil {
		fullName := strings.TrimSpace(row.FirstName + " " + row.LastName)
		if fullName == "" {
			fullName = row.Email
		}
		return fullName, row.Role
	}
	return "", ""
}

func (s *auditService) Record(ctx context.Context, params domain.CreateAuditLogParams) (*domain.AuditLog, error) {
	// 1. Auto-derive organization and actor info from context if missing
	userCtx, hasUserCtx := middleware.GetUserContext(ctx)

	orgID := params.OrgID
	if orgID <= 0 && hasUserCtx {
		orgID = userCtx.OrgID
	}
	if orgID <= 0 {
		orgID = 1 // Fallback root organization
	}

	actorID := params.ActorID
	if actorID == nil && hasUserCtx && userCtx.UserID > 0 {
		uid := userCtx.UserID
		actorID = &uid
	}

	actorRole := params.ActorRole
	if actorRole == "" && hasUserCtx {
		actorRole = userCtx.Role
	}

	actorType := params.ActorType
	if actorType == "" {
		if actorID != nil && *actorID > 0 {
			actorType = domain.ActorTypeUser
		} else {
			actorType = domain.ActorTypeSystem
		}
	}

	actorName := params.ActorName
	if actorName == "" && actorID != nil && *actorID > 0 {
		name, role := s.lookupUserName(ctx, *actorID)
		if name != "" {
			actorName = name
		}
		if actorRole == "" && role != "" {
			actorRole = role
		}
	}
	if actorName == "" {
		if actorType == domain.ActorTypeAIAgent {
			actorName = "Operations Agent"
		} else if actorType == domain.ActorTypeSystem {
			actorName = "System"
		} else {
			actorName = "User"
		}
	}

	// 2. Action & Module resolution
	action := strings.ToUpper(strings.TrimSpace(params.Action))
	if action == "" {
		action = domain.ActionUpdate
	}

	module := strings.ToUpper(strings.TrimSpace(params.Module))
	if module == "" {
		module = domain.ModuleSettings
	}

	resourceType := strings.ToUpper(strings.TrimSpace(params.ResourceType))
	if resourceType == "" {
		resourceType = strings.ToLower(module)
	}

	result := strings.ToUpper(strings.TrimSpace(params.Result))
	if result != domain.ResultFailed {
		result = domain.ResultSuccess
	}

	// 3. Human-readable description
	description := strings.TrimSpace(params.Description)
	if description == "" {
		description = generateDefaultDescription(module, action, resourceType, params.ResourceID, params.ResourceName)
	}

	// 4. Automatic Sensitive-Data Sanitization (MANDATORY)
	sanitizedBefore := domain.SanitizeMap(params.Before)
	sanitizedAfter := domain.SanitizeMap(params.After)
	sanitizedMetadata := domain.SanitizeMap(params.Metadata)

	// 5. Automatic structured change computation (Field, Before, After)
	changes := params.Changes
	if (changes == nil || len(changes) == 0) && (params.Before != nil || params.After != nil) {
		changes = domain.ComputeChanges(params.Before, params.After)
	}

	auditEntry := &domain.AuditLog{
		OrgID:        orgID,
		ActorID:      actorID,
		ActorType:    actorType,
		ActorName:    actorName,
		ActorRole:    actorRole,
		Action:       action,
		Module:       module,
		ResourceType: resourceType,
		ResourceID:   params.ResourceID,
		ResourceName: params.ResourceName,
		Description:  description,
		Result:       result,
		ErrorMessage: params.ErrorMessage,
		BeforeData:   sanitizedBefore,
		AfterData:    sanitizedAfter,
		Changes:      changes,
		Metadata:     sanitizedMetadata,
		IPAddress:    params.IPAddress,
		UserAgent:    params.UserAgent,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, auditEntry); err != nil {
		log.Printf("Universal AuditLog Error: failed to record event for org %d: %v", orgID, err)
		return nil, err
	}

	return auditEntry, nil
}

// RecordAsync dispatches audit log persistence in a non-blocking background goroutine.
func (s *auditService) RecordAsync(ctx context.Context, params domain.CreateAuditLogParams) {
	// Copy context parameters for safe asynchronous execution
	userCtx, hasUserCtx := middleware.GetUserContext(ctx)

	go func() {
		bgCtx := context.Background()
		if hasUserCtx {
			bgCtx = context.WithValue(bgCtx, middleware.UserContextKey, userCtx)
		}
		_, _ = s.Record(bgCtx, params)
	}()
}

func (s *auditService) List(ctx context.Context, filter domain.AuditLogFilter) (*domain.AuditLogListResponse, error) {
	if filter.OrgID <= 0 {
		userCtx, ok := middleware.GetUserContext(ctx)
		if ok && userCtx.OrgID > 0 {
			filter.OrgID = userCtx.OrgID
		} else {
			filter.OrgID = 1
		}
	}

	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	items, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(total) / float64(filter.Limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &domain.AuditLogListResponse{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *auditService) GetByID(ctx context.Context, orgID int64, id int64) (*domain.AuditLog, error) {
	if orgID <= 0 {
		userCtx, ok := middleware.GetUserContext(ctx)
		if ok && userCtx.OrgID > 0 {
			orgID = userCtx.OrgID
		} else {
			orgID = 1
		}
	}
	return s.repo.GetByID(ctx, orgID, id)
}
