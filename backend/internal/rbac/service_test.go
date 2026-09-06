package rbac

import (
	"context"
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestGetRoles(t *testing.T) {
	type args struct {
		ctx   context.Context
		orgID int64
	}
	tests := []struct {
		name        string
		args        args
		want        []RoleResponse
		wantErr     bool
		prepareMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successfully retrieves roles",
			args: args{
				ctx:   context.Background(),
				orgID: 1,
			},
			want: []RoleResponse{
				{ID: 1, OrgID: 1, Name: "SUPER_ADMIN", Description: "Full access"},
			},
			wantErr: false,
			prepareMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "org_id", "name", "description"}).
					AddRow(1, 1, "SUPER_ADMIN", "Full access")
				mock.ExpectQuery(`SELECT r\.id, r\.org_id, r\.name, COALESCE\(r\.description, ''\) AS description, COUNT\(rp\.permission_id\) AS permission_count FROM roles r LEFT JOIN role_permissions rp ON r\.id = rp\.role_id WHERE r\.org_id = \? GROUP BY r\.id, r\.org_id, r\.name, r\.description ORDER BY r\.id ASC`).
					WithArgs(int64(1)).
					WillReturnRows(rows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock db: %v", err)
			}
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if tt.prepareMock != nil {
				tt.prepareMock(mock)
			}

			s := &service{db: sqlxDB}
			got, err := s.GetRoles(tt.args.ctx, tt.args.orgID)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.GetRoles() = %v, want %v", got, tt.want)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetRolePermissions(t *testing.T) {
	type args struct {
		ctx    context.Context
		orgID  int64
		roleID int64
	}
	tests := []struct {
		name        string
		args        args
		want        *RolePermissionsResponse
		wantErr     bool
		prepareMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successfully retrieves permissions",
			args: args{
				ctx:    context.Background(),
				orgID:  1,
				roleID: 1,
			},
			want: &RolePermissionsResponse{
				RoleID:   1,
				RoleName: "SALES",
				Permissions: []PermissionNode{
					{ID: 1, Resource: "COMPANIES", Action: "READ"},
				},
			},
			wantErr: false,
			prepareMock: func(mock sqlmock.Sqlmock) {
				// 1. Verify role
				mock.ExpectQuery(`SELECT name FROM roles WHERE id = \? AND org_id = \?`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("SALES"))

				// 2. Get permissions
				permsRows := sqlmock.NewRows([]string{"id", "resource", "action"}).
					AddRow(1, "COMPANIES", "READ")
				mock.ExpectQuery(`SELECT p\.id, p\.resource, p\.action FROM role_permissions rp JOIN permissions p ON rp\.permission_id = p\.id WHERE rp\.role_id = \? ORDER BY p\.resource, p\.action`).
					WithArgs(int64(1)).
					WillReturnRows(permsRows)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock db: %v", err)
			}
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if tt.prepareMock != nil {
				tt.prepareMock(mock)
			}

			s := &service{db: sqlxDB}
			got, err := s.GetRolePermissions(tt.args.ctx, tt.args.orgID, tt.args.roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.GetRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("service.GetRolePermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateRolePermissions(t *testing.T) {
	type args struct {
		ctx    context.Context
		orgID  int64
		roleID int64
		req    UpdatePermissionsRequest
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		prepareMock func(mock sqlmock.Sqlmock)
	}{
		{
			name: "Successfully updates permissions",
			args: args{
				ctx:    context.Background(),
				orgID:  1,
				roleID: 1,
				req: UpdatePermissionsRequest{
					Permissions: []PermissionNode{
						{Resource: "COMPANIES", Action: "READ"},
					},
				},
			},
			wantErr: false,
			prepareMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT name FROM roles WHERE id = \? AND org_id = \?`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"name"}).AddRow("SALES"))

				// 2. Begin tx
				mock.ExpectBegin()

				// 3. Delete existing
				mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = \?`).
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// 4. Insert new
				mock.ExpectPrepare(`INSERT INTO role_permissions \(role_id, permission_id\) SELECT \?, id FROM permissions WHERE resource = \? AND action = \?`)
				mock.ExpectExec(`INSERT INTO role_permissions`).
					WithArgs(int64(1), "COMPANIES", "READ").
					WillReturnResult(sqlmock.NewResult(0, 1))

				// 5. Commit
				mock.ExpectCommit()
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Failed to create mock db: %v", err)
			}
			defer db.Close()
			sqlxDB := sqlx.NewDb(db, "sqlmock")

			if tt.prepareMock != nil {
				tt.prepareMock(mock)
			}

			s := &service{db: sqlxDB}
			err = s.UpdateRolePermissions(tt.args.ctx, tt.args.orgID, tt.args.roleID, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("service.UpdateRolePermissions() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
