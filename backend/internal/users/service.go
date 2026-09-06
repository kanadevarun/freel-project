package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/freel/backend/internal/audit"
	"github.com/freel/backend/internal/audit/domain"
	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/rbac"
)

// Service defines the business logic operations for the users module.
// It acts as the intermediary between the HTTP handlers and the database repository,
// enforcing business rules (like generating secure tokens) and orchestrating side effects
// (like sending invitation emails).
type Service interface {
	// InviteUser handles the process of inviting a new team member to an organization.
	// It performs the following steps:
	// 1. Generates a secure, random 32-byte hexadecimal token.
	// 2. Constructs an Invitation record with a 7-day expiration.
	// 3. Persists the invitation via the Repository.
	// 4. Sends an email to the invitee via the Notifications service containing the accept link.
	// Returns an error if any of these steps fail.
	InviteUser(ctx context.Context, orgID int64, req InviteUserRequest) error

	// ListUsers retrieves all members of a specific organization.
	// It simply delegates to the repository and returns the flattened OrgMemberResponse list.
	ListUsers(ctx context.Context, orgID int64) ([]OrgMemberResponse, error)

	// ListInvitations retrieves all pending invitations for a specific organization.
	ListInvitations(ctx context.Context, orgID int64) ([]InvitationResponse, error)

	// CancelInvitation deletes a pending invitation.
	CancelInvitation(ctx context.Context, orgID, inviteID int64) error

	// UpdateRole changes the role of an existing organization member.
	// It calls the repository to update the role_id in the org_members table.
	UpdateRole(ctx context.Context, orgID, userID int64, req UpdateRoleRequest) error

	// RemoveUser revokes a user's access to an organization.
	// It calls the repository to delete the org_members record linking the user to the org.
	RemoveUser(ctx context.Context, orgID, userID int64) error
}

// serviceImpl is the concrete implementation of the users Service interface.
// It holds references to the user repository and the notification service.
type serviceImpl struct {
	repo         Repository
	notifService notifications.Service
}

// NewService constructs and returns a new users Service implementation.
// It requires a valid user Repository and a notifications Service to function.
func NewService(repo Repository, notifService notifications.Service) Service {
	return &serviceImpl{
		repo:         repo,
		notifService: notifService,
	}
}

// InviteUser executes the business logic for inviting a user.
// It generates a secure token, saves the invitation to the database, and uses the
// notification service to send an email. If the email fails, the process fails,
// but note that the invitation might still be saved in the database (in a real
// production scenario, this might use a transaction or saga pattern).
func (s *serviceImpl) InviteUser(ctx context.Context, orgID int64, req InviteUserRequest) error {
	// 1. Generate a secure random token for the invite link.
	// We generate 32 random bytes, which is cryptographically secure.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate invite token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// 2. Construct the invitation record.
	// We set the expiration time to 7 days from the current time.
	invitation := &Invitation{
		OrgID:     orgID,
		RoleID:    req.RoleID,
		Email:     req.Email,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	// 3. Persist the invitation in the database.
	// If the user has already been invited to this org, the unique constraint (org_id, email)
	// will trigger an error, preventing duplicate active invites.
	err := s.repo.CreateInvitation(ctx, invitation)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return fmt.Errorf("INVITATION_ALREADY_EXISTS: An invitation has already been sent to this email address.")
		}
		return fmt.Errorf("failed to create invitation record: %w", err)
	}

	// 4. Send the beautifully formatted HTML invitation email.
	// We pass the target email, the secure token, and the organization name to the template.
	err = s.notifService.SendInviteEmail(ctx, req.Email, token, req.OrgName)
	if err != nil {
		// Log the error. Ideally, we might want to clean up the invitation here or enqueue a retry.
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionInvite,
		Module:       domain.ModuleUsers,
		ResourceType: "USER",
		ResourceName: req.Email,
		Description:  fmt.Sprintf("Invited user %s", req.Email),
		Result:       domain.ResultSuccess,
	})

	return nil
}

// ListUsers fetches all active members for the specified organization ID.
// It directly invokes the corresponding repository method.
func (s *serviceImpl) ListUsers(ctx context.Context, orgID int64) ([]OrgMemberResponse, error) {
	members, err := s.repo.ListOrgMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}
	return members, nil
}

// ListInvitations fetches all pending invitations for the specified organization ID.
func (s *serviceImpl) ListInvitations(ctx context.Context, orgID int64) ([]InvitationResponse, error) {
	invites, err := s.repo.ListInvitations(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve invitations: %w", err)
	}
	return invites, nil
}

// CancelInvitation deletes a pending invitation.
func (s *serviceImpl) CancelInvitation(ctx context.Context, orgID, inviteID int64) error {
	err := s.repo.DeleteInvitation(ctx, orgID, inviteID)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionDelete,
		Module:       domain.ModuleUsers,
		ResourceType: "INVITATION",
		ResourceID:   fmt.Sprintf("%d", inviteID),
		Description:  fmt.Sprintf("Cancelled invitation #%d", inviteID),
		Result:       domain.ResultSuccess,
	})

	return nil
}

// UpdateRole modifies the role assigned to a specific user within the organization.
// It maps the HTTP request payload directly to the repository update method.
func (s *serviceImpl) UpdateRole(ctx context.Context, orgID, userID int64, req UpdateRoleRequest) error {
	// 1. Safety check: Protect the last SUPER_ADMIN
	currentRole, err := s.repo.GetMemberRoleName(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to check current role: %w", err)
	}

	if currentRole == rbac.RoleSuperAdmin {
		// Is this the *only* SUPER_ADMIN?
		superAdminCount, err := s.repo.CountActiveSuperAdmins(ctx, orgID)
		if err != nil {
			return fmt.Errorf("failed to check super admin count: %w", err)
		}
		if superAdminCount <= 1 {
			// This is the last SUPER_ADMIN (or 0 due to some anomaly). Prevent changing their role.
			// We return an error that will map to a 403/409 (handled generically for now, but msg is clear)
			return fmt.Errorf("Cannot change the role of the last SUPER_ADMIN. The organization must have at least one active SUPER_ADMIN.")
		}
	}

	// 2. Perform the update
	err = s.repo.UpdateMemberRole(ctx, orgID, userID, req.RoleID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionRoleChanged,
		Module:       domain.ModuleRolesPermissions,
		ResourceType: "USER",
		ResourceID:   fmt.Sprintf("%d", userID),
		Description:  fmt.Sprintf("Changed user #%d role from %s", userID, currentRole),
		Before:       map[string]interface{}{"role": currentRole},
		After:        map[string]interface{}{"role_id": req.RoleID},
		Result:       domain.ResultSuccess,
	})

	return nil
}

// RemoveUser deletes the association between a user and an organization.
// It does not delete the user account entirely, only their access to this specific organization.
func (s *serviceImpl) RemoveUser(ctx context.Context, orgID, userID int64) error {
	// 1. Safety check: Protect the last SUPER_ADMIN
	currentRole, err := s.repo.GetMemberRoleName(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to check current role before removal: %w", err)
	}

	if currentRole == rbac.RoleSuperAdmin {
		superAdminCount, err := s.repo.CountActiveSuperAdmins(ctx, orgID)
		if err != nil {
			return fmt.Errorf("failed to check super admin count: %w", err)
		}
		if superAdminCount <= 1 {
			return fmt.Errorf("Cannot remove the last SUPER_ADMIN from the organization")
		}
	}

	err = s.repo.RemoveMember(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}

	_, _ = audit.Record(ctx, domain.CreateAuditLogParams{
		OrgID:        orgID,
		Action:       domain.ActionDelete,
		Module:       domain.ModuleUsers,
		ResourceType: "USER",
		ResourceID:   fmt.Sprintf("%d", userID),
		Description:  fmt.Sprintf("Removed user #%d from organization", userID),
		Result:       domain.ResultSuccess,
	})

	return nil
}
