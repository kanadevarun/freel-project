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
				mock.ExpectQuery(`SELECT id, org_id, name, description FROM roles WHERE org_id = \$1 ORDER BY id ASC`).
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
				RoleID: 1,
				Permissions: []PermissionNode{
					{Resource: "COMPANIES", Action: "READ"},
				},
			},
			wantErr: false,
			prepareMock: func(mock sqlmock.Sqlmock) {
				// 1. Verify role
				mock.ExpectQuery(`SELECT 1 FROM roles WHERE id = \$1 AND org_id = \$2`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

				// 2. Get permissions
				permsRows := sqlmock.NewRows([]string{"resource", "action"}).
					AddRow("COMPANIES", "READ")
				mock.ExpectQuery(`SELECT p\.resource, p\.action FROM role_permissions rp JOIN permissions p ON rp\.permission_id = p\.id WHERE rp\.role_id = \$1`).
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
				// 1. Verify role
				mock.ExpectQuery(`SELECT 1 FROM roles WHERE id = \$1 AND org_id = \$2`).
					WithArgs(int64(1), int64(1)).
					WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(1))

				// 2. Begin tx
				mock.ExpectBegin()

				// 3. Delete existing
				mock.ExpectExec(`DELETE FROM role_permissions WHERE role_id = \$1`).
					WithArgs(int64(1)).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// 4. Insert new
				mock.ExpectPrepare(`INSERT INTO role_permissions \(role_id, permission_id\) SELECT \$1, id FROM permissions WHERE resource = \$2 AND action = \$3`)
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
