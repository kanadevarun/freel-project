package organization

import (
	"context"
	"time"

	"github.com/freel/backend/internal/common/events"
)

// Service defines the business logic for Organization.
type Service interface {
	CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error)
	UpdateOrganization(ctx context.Context, id int64, req UpdateOrgRequest) (*Organization, error)
	InviteUser(ctx context.Context, orgID int64, req InviteRequest) error
}

type service struct {
	repo     Repository
	eventBus events.Bus
}

// NewService creates a new organization service.
// Simple meaning: It builds the tool you need to manage companies.
// Example: svc := NewService(myRepo, myEventBus)
func NewService(repo Repository, eventBus events.Bus) Service {
	return &service{repo: repo, eventBus: eventBus}
}

// CreateOrganization handles the business logic to create a new organization.
// Simple meaning: You give it a name, and it sets up a brand new tenant (company workspace) in the system.
// Example: org, err := svc.CreateOrganization(ctx, CreateOrgRequest{Name: "Acme Corp"})
func (s *service) CreateOrganization(ctx context.Context, req CreateOrgRequest) (*Organization, error) {
	org := &Organization{
		Name:      req.Name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.repo.Create(ctx, org)
	if err != nil {
		return nil, err
	}

	// Publish OrgCreated event (e.g. to seed roles)
	s.eventBus.Publish(events.Event{
		ID:        "evt-org-created", // Typically generate UUID here
		Type:      events.EventOrgCreated,
		Payload:   org,
		Timestamp: time.Now(),
	})

	return org, nil
}

// UpdateOrganization updates the details of an existing organization.
// Simple meaning: You want to rename a company workspace. You give it the ID and the new name.
// Example: org, err := svc.UpdateOrganization(ctx, 123, UpdateOrgRequest{Name: "Acme Global"})
func (s *service) UpdateOrganization(ctx context.Context, id int64, req UpdateOrgRequest) (*Organization, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	org.Name = req.Name
	org.UpdatedAt = time.Now()
	
	err = s.repo.Update(ctx, org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

// InviteUser triggers the process of inviting a new person to join the organization.
// Simple meaning: You want your colleague to join your workspace. This function fires off an email invite.
// Example: err := svc.InviteUser(ctx, 123, InviteRequest{Email: "bob@acme.com", Role: "SALES"})
func (s *service) InviteUser(ctx context.Context, orgID int64, req InviteRequest) error {
	// 1. Create a pending invite record in DB (Users module dependency, or handled here temporarily)
	// 2. Publish UserInvited event
	s.eventBus.Publish(events.Event{
		ID:        "evt-user-invited",
		Type:      events.EventUserInvited,
		Payload: map[string]interface{}{
			"org_id": orgID,
			"email":  req.Email,
			"role":   req.Role,
		},
		Timestamp: time.Now(),
	})
	return nil
}
