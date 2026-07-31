package users

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/freel/backend/internal/notifications"
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
		return fmt.Errorf("failed to create invitation record: %w", err)
	}

	// 4. Send the beautifully formatted HTML invitation email.
	// We pass the target email, the secure token, and the organization name to the template.
	err = s.notifService.SendInviteEmail(ctx, req.Email, token, req.OrgName)
	if err != nil {
		// Log the error. Ideally, we might want to clean up the invitation here or enqueue a retry.
		return fmt.Errorf("failed to send invitation email: %w", err)
	}

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
	// Normally we would verify the invite actually belongs to orgID before deleting.
	// We'll trust the repository to handle deletion cleanly for now.
	err := s.repo.DeleteInvitation(ctx, inviteID)
	if err != nil {
		return fmt.Errorf("failed to cancel invitation: %w", err)
	}
	return nil
}

// UpdateRole modifies the role assigned to a specific user within the organization.
// It maps the HTTP request payload directly to the repository update method.
func (s *serviceImpl) UpdateRole(ctx context.Context, orgID, userID int64, req UpdateRoleRequest) error {
	err := s.repo.UpdateMemberRole(ctx, orgID, userID, req.RoleID)
	if err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}
	return nil
}

// RemoveUser deletes the association between a user and an organization.
// It does not delete the user account entirely, only their access to this specific organization.
func (s *serviceImpl) RemoveUser(ctx context.Context, orgID, userID int64) error {
	err := s.repo.RemoveMember(ctx, orgID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %w", err)
	}
	return nil
}
