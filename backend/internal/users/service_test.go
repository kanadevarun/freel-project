package users_test

import (
	"context"
	"errors"
	"testing"

	"github.com/freel/backend/internal/notifications"
	"github.com/freel/backend/internal/users"
	"go.uber.org/mock/gomock"
)

// mocks groups the generated mock dependencies for the users Service tests.
type mocks struct {
	repo         *users.MockRepository
	notifService *notifications.MockService
}

// TestInviteUser tests the InviteUser method of the users Service.
// It uses a table-driven approach to thoroughly verify behavior under different
// success and failure conditions, including repository errors and notification failures.
func TestInviteUser(t *testing.T) {
	type args struct {
		ctx   context.Context
		orgID int64
		req   users.InviteUserRequest
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		prepareTest func(mock *mocks)
	}{
		{
			name: "Failed to create invitation record",
			args: args{
				ctx:   context.TODO(),
				orgID: 1,
				req: users.InviteUserRequest{
					Email:  "test@logisticshq.in",
					RoleID: 2,
				},
			},
			wantErr: true,
			prepareTest: func(mock *mocks) {
				// We expect the repository's CreateInvitation method to be called,
				// and we simulate it returning an error (e.g., duplicate invite).
				mock.repo.EXPECT().
					CreateInvitation(gomock.Any(), gomock.Any()).
					Return(errors.New("duplicate entry"))
			},
		},
		{
			name: "Failed to send email notification",
			args: args{
				ctx:   context.TODO(),
				orgID: 1,
				req: users.InviteUserRequest{
					Email:  "test@logisticshq.in",
					RoleID: 2,
				},
			},
			wantErr: true,
			prepareTest: func(mock *mocks) {
				// The repository succeeds.
				mock.repo.EXPECT().
					CreateInvitation(gomock.Any(), gomock.Any()).
					Return(nil)

				// However, the notification service fails to send the email.
				mock.notifService.EXPECT().
					SendInviteEmail(gomock.Any(), "test@logisticshq.in", gomock.Any(), gomock.Any()).
					Return(errors.New("SES delivery failed"))
			},
		},
		{
			name: "Successfully invites user",
			args: args{
				ctx:   context.TODO(),
				orgID: 1,
				req: users.InviteUserRequest{
					Email:  "test@logisticshq.in",
					RoleID: 2,
				},
			},
			wantErr: false,
			prepareTest: func(mock *mocks) {
				// Both the repository and the notification service succeed.
				mock.repo.EXPECT().
					CreateInvitation(gomock.Any(), gomock.Any()).
					Return(nil)

				mock.notifService.EXPECT().
					SendInviteEmail(gomock.Any(), "test@logisticshq.in", gomock.Any(), gomock.Any()).
					Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// Initialize the mocks
			m := &mocks{
				repo:         users.NewMockRepository(ctrl),
				notifService: notifications.NewMockService(ctrl),
			}

			// Setup the expectations using the provided prepareTest function
			if tt.prepareTest != nil {
				tt.prepareTest(m)
			}

			// Instantiate the service with the mocked dependencies
			svc := users.NewService(m.repo, m.notifService)

			// Execute the method under test
			err := svc.InviteUser(tt.args.ctx, tt.args.orgID, tt.args.req)
			
			// Validate the outcome
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.InviteUser() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestListUsers tests the ListUsers method of the users Service.
func TestListUsers(t *testing.T) {
	type args struct {
		ctx   context.Context
		orgID int64
	}

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		wantLen     int
		prepareTest func(mock *mocks)
	}{
		{
			name: "Failed to list users from repository",
			args: args{
				ctx:   context.TODO(),
				orgID: 1,
			},
			wantErr: true,
			wantLen: 0,
			prepareTest: func(mock *mocks) {
				mock.repo.EXPECT().
					ListOrgMembers(gomock.Any(), int64(1)).
					Return(nil, errors.New("db error"))
			},
		},
		{
			name: "Successfully retrieves users",
			args: args{
				ctx:   context.TODO(),
				orgID: 1,
			},
			wantErr: false,
			wantLen: 2,
			prepareTest: func(mock *mocks) {
				mock.repo.EXPECT().
					ListOrgMembers(gomock.Any(), int64(1)).
					Return([]users.OrgMemberResponse{
						{UserID: 1, Email: "user1@logisticshq.in"},
						{UserID: 2, Email: "user2@logisticshq.in"},
					}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			m := &mocks{
				repo:         users.NewMockRepository(ctrl),
				notifService: notifications.NewMockService(ctrl),
			}

			if tt.prepareTest != nil {
				tt.prepareTest(m)
			}

			svc := users.NewService(m.repo, m.notifService)

			got, err := svc.ListUsers(tt.args.ctx, tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Service.ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr && len(got) != tt.wantLen {
				t.Errorf("Service.ListUsers() returned %d items, want %d", len(got), tt.wantLen)
			}
		})
	}
}
