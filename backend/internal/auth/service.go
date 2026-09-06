package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/rbac"
	"github.com/freel/backend/internal/utils"
	"github.com/jmoiron/sqlx"
)

type Service struct {
	cfg    *config.Config
	client *cognitoidentityprovider.Client
	db     *sqlx.DB
}

func NewService(cfg *config.Config, db *sqlx.DB) *Service {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		log.Fatalf("unable to load AWS config, %v", err)
	}

	return &Service{
		cfg:    cfg,
		client: cognitoidentityprovider.NewFromConfig(awsCfg),
		db:     db,
	}
}

func (s *Service) Signup(ctx context.Context, req SignupRequest) error {
	if err := utils.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := utils.ValidatePassword(req.Password); err != nil {
		return err
	}
	if err := utils.ValidateRequired(req.FullName, "full_name"); err != nil {
		return err
	}
	if err := utils.ValidateRequired(req.CompanyName, "company_name"); err != nil {
		return err
	}

	// Phase 1: Only Freight Forwarders onboard via self-serve signup.
	allowedRoles := map[string]bool{
		"freight_forwarder": true,
	}

	if !allowedRoles[req.Role] {
		return errors.New("invalid role")
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)

	_, err := s.client.SignUp(ctx, &cognitoidentityprovider.SignUpInput{
		ClientId:   &s.cfg.CognitoClientID,
		Username:   &req.Email,
		Password:   &req.Password,
		SecretHash: &secretHash,
		UserAttributes: []types.AttributeType{
			{Name: stringPtr("email"), Value: &req.Email},
			{Name: stringPtr("name"), Value: &req.FullName},
			{Name: stringPtr("custom:company_name"), Value: &req.CompanyName},
			{Name: stringPtr("custom:role"), Value: &req.Role},
		},
	})
	log.Printf("[AUTH] Cognito signup successful for email=%s, company=%s", req.Email, req.CompanyName)
	return err
}

func (s *Service) VerifyEmail(ctx context.Context, req VerifyEmailRequest) error {
	if err := utils.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := utils.ValidateRequired(req.Code, "code"); err != nil {
		return err
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)

	_, err := s.client.ConfirmSignUp(ctx, &cognitoidentityprovider.ConfirmSignUpInput{
		ClientId:         &s.cfg.CognitoClientID,
		Username:         &req.Email,
		ConfirmationCode: &req.Code,
		SecretHash:       &secretHash,
	})
	if err != nil {
		log.Printf("[AUTH ERROR] ConfirmSignUp failed for email=%s: %v", req.Email, err)
		return err
	}
	log.Printf("[AUTH] Cognito user confirmed for email=%s", req.Email)

	// Immediately provision MySQL database workspace upon successful email verification
	cognitoSub := req.Email
	fullName := ""
	companyName := ""
	if s.cfg.CognitoUserPoolID != "" {
		adminUserResp, adminErr := s.client.AdminGetUser(ctx, &cognitoidentityprovider.AdminGetUserInput{
			UserPoolId: &s.cfg.CognitoUserPoolID,
			Username:   &req.Email,
		})
		if adminErr == nil && adminUserResp != nil {
			if adminUserResp.Username != nil {
				cognitoSub = *adminUserResp.Username
			}
			for _, attr := range adminUserResp.UserAttributes {
				if attr.Name != nil && attr.Value != nil {
					switch *attr.Name {
					case "sub":
						cognitoSub = *attr.Value
					case "name":
						fullName = *attr.Value
					case "custom:company_name":
						companyName = *attr.Value
					}
				}
			}
		}
	}

	if provErr := s.provisionTenantIfMissing(ctx, cognitoSub, req.Email, fullName, companyName); provErr != nil {
		log.Printf("[PROVISION ERROR] VerifyEmail onboarding failed for email=%s: %v", req.Email, provErr)
		return fmt.Errorf("workspace provisioning failed: %w", provErr)
	}

	return nil
}



func (s *Service) provisionTenantIfMissing(ctx context.Context, cognitoSub, email, fullName, companyName string) error {
	log.Printf("[PROVISION] Starting MySQL onboarding for email=%s, cognito_sub=%s, company=%s", email, cognitoSub, companyName)

	if companyName == "" {
		companyName = "Freight Forwarder Workspace"
	}
	if fullName == "" {
		fullName = strings.Split(email, "@")[0]
	}

	parts := strings.SplitN(fullName, " ", 2)
	firstName := parts[0]
	lastName := ""
	if len(parts) > 1 {
		lastName = parts[1]
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=begin_tx error=%v", err)
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// 1. Ensure system permissions are defined
	log.Printf("[PROVISION] Ensuring system permissions...")
	if err := rbac.NewService(s.db, nil).SeedSystemPermissions(ctx); err != nil {
		log.Printf("[PROVISION ERROR] step=ensure_permissions error=%v", err)
		return err
	}

	// 2. Check or create User
	log.Printf("[PROVISION] Creating user record: email=%s, cognito_sub=%s", email, cognitoSub)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (cognito_sub, email, first_name, last_name, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE cognito_sub = VALUES(cognito_sub), first_name = VALUES(first_name), last_name = VALUES(last_name), updated_at = NOW()
	`, cognitoSub, email, firstName, lastName)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=upsert_user error=%v", err)
		return fmt.Errorf("upsert user: %w", err)
	}

	var userID int64
	err = tx.QueryRowxContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, email).Scan(&userID)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=get_user_id error=%v", err)
		return fmt.Errorf("get user id: %w", err)
	}
	log.Printf("[PROVISION] User created/resolved: id=%d", userID)

	// 3. Check if user already has an active org membership
	var existingOrgID int64
	err = tx.QueryRowxContext(ctx, `SELECT org_id FROM org_members WHERE user_id = ? LIMIT 1`, userID).Scan(&existingOrgID)
	if err == nil && existingOrgID > 0 {
		log.Printf("[PROVISION] User id=%d already has active org_id=%d. Skipping creation.", userID, existingOrgID)
		return tx.Commit()
	}

	// 4. Create Organization
	log.Printf("[PROVISION] Creating organization: name=%s", companyName)
	res, err := tx.ExecContext(ctx, `
		INSERT INTO organizations (name, created_at, updated_at)
		VALUES (?, NOW(), NOW())
	`, companyName)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=create_organization error=%v", err)
		return fmt.Errorf("create organization: %w", err)
	}
	orgID, err := res.LastInsertId()
	if err != nil {
		log.Printf("[PROVISION ERROR] step=get_org_id error=%v", err)
		return fmt.Errorf("get org id: %w", err)
	}
	log.Printf("[PROVISION] Organization created: id=%d", orgID)

	// Step 5 (REMOVED): Previously created a separate 'companies' row here.
	// Now organizations table IS the company — no duplicate entry needed.
	log.Printf("[PROVISION] Organization is the company record: id=%d, name=%s", orgID, companyName)


	// 6. Create SUPER_ADMIN role for the new org
	log.Printf("[PROVISION] Creating SUPER_ADMIN role for org_id=%d", orgID)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO roles (org_id, name, description, created_at, updated_at)
		VALUES (?, 'SUPER_ADMIN', 'Super Admin role with full access', NOW(), NOW())
		ON DUPLICATE KEY UPDATE description = VALUES(description)
	`, orgID)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=create_role error=%v", err)
		return fmt.Errorf("create role: %w", err)
	}
	var roleID int64
	err = tx.QueryRowxContext(ctx, `SELECT id FROM roles WHERE org_id = ? AND name = 'SUPER_ADMIN' LIMIT 1`, orgID).Scan(&roleID)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=get_role_id error=%v", err)
		return fmt.Errorf("get role id: %w", err)
	}
	log.Printf("[PROVISION] SUPER_ADMIN role created/resolved: id=%d", roleID)

	// 7. Assign all permissions to SUPER_ADMIN
	log.Printf("[PROVISION] Assigning permissions to role_id=%d...", roleID)
	permRes, err := tx.ExecContext(ctx, `
		INSERT IGNORE INTO role_permissions (role_id, permission_id, created_at)
		SELECT ?, id, NOW() FROM permissions
	`, roleID)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=assign_role_permissions error=%v", err)
		return fmt.Errorf("assign role permissions: %w", err)
	}
	permCount, _ := permRes.RowsAffected()
	log.Printf("[PROVISION] Role permissions assigned: count=%d", permCount)

	// 8. Create org_member linking User + Org + Role
	log.Printf("[PROVISION] Creating org_members record: org_id=%d, user_id=%d, role_id=%d", orgID, userID, roleID)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO org_members (org_id, user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, 'ACTIVE', NOW(), NOW())
		ON DUPLICATE KEY UPDATE role_id = VALUES(role_id), status = 'ACTIVE', updated_at = NOW()
	`, orgID, userID, roleID)
	if err != nil {
		log.Printf("[PROVISION ERROR] step=create_org_member error=%v", err)
		return fmt.Errorf("create org_member: %w", err)
	}

	if err := tx.Commit(); err != nil {
		log.Printf("[PROVISION ERROR] step=commit_tx error=%v", err)
		return fmt.Errorf("commit tx: %w", err)
	}

	// 7. Seed remaining default roles using the RBAC service
	log.Printf("[PROVISION] Seeding remaining default roles for org_id=%d", orgID)
	if err := rbac.NewService(s.db, nil).SeedDefaultRolesForOrg(ctx, orgID); err != nil {
		log.Printf("[PROVISION ERROR] step=seed_roles error=%v", err)
		// We don't return an error here so that the user can still log in,
		// but the failure is logged.
	}

	log.Printf("[PROVISION] Provisioning completed successfully for user_id=%d, org_id=%d", userID, orgID)
	return nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponseData, error) {
	if err := utils.ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := utils.ValidateRequired(req.Password, "password"); err != nil {
		return nil, err
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)

	authInput := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		ClientId: &s.cfg.CognitoClientID,
		AuthParameters: map[string]string{
			"USERNAME":    req.Email,
			"PASSWORD":    req.Password,
			"SECRET_HASH": secretHash,
		},
	}

	resp, err := s.client.InitiateAuth(ctx, authInput)
	if err != nil {
		log.Printf("[AUTH ERROR] InitiateAuth failed for email=%s: %v", req.Email, err)
		return nil, err
	}

	if resp.AuthenticationResult == nil {
		return nil, errors.New("authentication result is missing")
	}

	// Fetch user info from Cognito
	getUserResp, err := s.client.GetUser(ctx, &cognitoidentityprovider.GetUserInput{
		AccessToken: resp.AuthenticationResult.AccessToken,
	})

	cognitoSub := ""
	fullName := ""
	companyName := ""
	if err == nil && getUserResp != nil {
		cognitoSub = *getUserResp.Username
		for _, attr := range getUserResp.UserAttributes {
			if attr.Name != nil && attr.Value != nil {
				switch *attr.Name {
				case "sub":
					cognitoSub = *attr.Value
				case "name":
					fullName = *attr.Value
				case "custom:company_name":
					companyName = *attr.Value
				}
			}
		}
	}
	if cognitoSub == "" {
		cognitoSub = req.Email
	}

	log.Printf("[AUTH] Login successful in Cognito. Email=%s, cognito_sub=%s, company=%s", req.Email, cognitoSub, companyName)

	// Auto-provision tenant if not yet in MySQL
	if provErr := s.provisionTenantIfMissing(ctx, cognitoSub, req.Email, fullName, companyName); provErr != nil {
		log.Printf("[PROVISION ERROR] Onboarding failed for email=%s: %v", req.Email, provErr)
		return nil, fmt.Errorf("failed to provision workspace: %w", provErr)
	}

	// Fetch user, org, and role details from Postgres
	var dbInfo struct {
		UserID    int64  `db:"user_id"`
		Email     string `db:"email"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		OrgID     int64  `db:"org_id"`
		OrgName   string `db:"org_name"`
		RoleName  string `db:"role_name"`
	}

	query := `
		SELECT u.id AS user_id, u.email, COALESCE(u.first_name, '') AS first_name, COALESCE(u.last_name, '') AS last_name,
		       o.id AS org_id, o.name AS org_name, r.name AS role_name
		FROM users u
		JOIN org_members om ON u.id = om.user_id
		JOIN organizations o ON om.org_id = o.id
		JOIN roles r ON om.role_id = r.id
		WHERE (u.email = ? OR u.cognito_sub = ?) AND om.status = 'ACTIVE'
		LIMIT 1
	`
	err = s.db.GetContext(ctx, &dbInfo, query, req.Email, cognitoSub)
	roleName := "GUEST"
	if err == nil {
		roleName = dbInfo.RoleName
	}

	var permissions []string
	if roleName != "GUEST" {
		permQuery := `
			SELECT CONCAT(p.resource, ':', p.action)
			FROM role_permissions rp
			JOIN permissions p ON rp.permission_id = p.id
			JOIN roles r ON rp.role_id = r.id
			JOIN org_members om ON r.id = om.role_id
			WHERE om.user_id = ?
		`
		_ = s.db.SelectContext(ctx, &permissions, permQuery, dbInfo.UserID)
	}

	displayRole := roleName
	if displayRole == "SUPER_ADMIN" {
		displayRole = "Super Admin"
	}

	resolvedFullName := fullName
	if resolvedFullName == "" && (dbInfo.FirstName != "" || dbInfo.LastName != "") {
		resolvedFullName = strings.TrimSpace(dbInfo.FirstName + " " + dbInfo.LastName)
	}
	if resolvedFullName == "" {
		resolvedFullName = strings.Split(req.Email, "@")[0]
	}

	var refreshToken string
	if resp.AuthenticationResult.RefreshToken != nil {
		refreshToken = *resp.AuthenticationResult.RefreshToken
	}

	return &LoginResponseData{
		AccessToken:  *resp.AuthenticationResult.AccessToken,
		IDToken:      *resp.AuthenticationResult.IdToken,
		RefreshToken: refreshToken,
		ExpiresIn:    resp.AuthenticationResult.ExpiresIn,
		Role: RoleResponse{
			Name:        roleName,
			DisplayName: displayRole,
			Permissions: permissions,
		},
		User: UserResponse{
			ID:        dbInfo.UserID,
			Email:     req.Email,
			FullName:  resolvedFullName,
			FirstName: dbInfo.FirstName,
			LastName:  dbInfo.LastName,
		},
		Org: OrgResponse{
			ID:   dbInfo.OrgID,
			Name: dbInfo.OrgName,
		},
	}, nil
}

func (s *Service) Refresh(ctx context.Context, req RefreshRequest) (*RefreshResponseData, error) {
	if req.RefreshToken == "" {
		return nil, errors.New("refresh_token is required")
	}

	authParams := map[string]string{
		"REFRESH_TOKEN": req.RefreshToken,
	}
	if s.cfg.CognitoClientSecret != "" && req.Email != "" {
		authParams["SECRET_HASH"] = ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)
	}

	authInput := &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow:       types.AuthFlowTypeRefreshTokenAuth,
		ClientId:       &s.cfg.CognitoClientID,
		AuthParameters: authParams,
	}

	resp, err := s.client.InitiateAuth(ctx, authInput)
	if err != nil {
		return nil, err
	}
	if resp.AuthenticationResult == nil {
		return nil, errors.New("authentication result is missing")
	}

	refreshToken := req.RefreshToken
	if resp.AuthenticationResult.RefreshToken != nil {
		refreshToken = *resp.AuthenticationResult.RefreshToken
	}

	return &RefreshResponseData{
		AccessToken:  *resp.AuthenticationResult.AccessToken,
		IDToken:      *resp.AuthenticationResult.IdToken,
		RefreshToken: refreshToken,
		ExpiresIn:    resp.AuthenticationResult.ExpiresIn,
	}, nil
}

func (s *Service) GetMe(ctx context.Context, userID int64) (*CurrentUserResponseData, error) {
	var dbInfo struct {
		UserID    int64  `db:"user_id"`
		Email     string `db:"email"`
		FirstName string `db:"first_name"`
		LastName  string `db:"last_name"`
		OrgID     int64  `db:"org_id"`
		OrgName   string `db:"org_name"`
		RoleName  string `db:"role_name"`
	}

	query := `
		SELECT u.id AS user_id, u.email, COALESCE(u.first_name, '') AS first_name, COALESCE(u.last_name, '') AS last_name,
		       o.id AS org_id, o.name AS org_name, r.name AS role_name
		FROM users u
		JOIN org_members om ON u.id = om.user_id
		JOIN organizations o ON om.org_id = o.id
		JOIN roles r ON om.role_id = r.id
		WHERE u.id = ? AND om.status = 'ACTIVE'
		LIMIT 1
	`
	err := s.db.GetContext(ctx, &dbInfo, query, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	var permissions []string
	permQuery := `
		SELECT CONCAT(p.resource, ':', p.action)
		FROM role_permissions rp
		JOIN permissions p ON rp.permission_id = p.id
		JOIN roles r ON rp.role_id = r.id
		JOIN org_members om ON r.id = om.role_id
		WHERE om.user_id = ?
	`
	_ = s.db.SelectContext(ctx, &permissions, permQuery, userID)

	displayRole := dbInfo.RoleName
	if displayRole == "SUPER_ADMIN" {
		displayRole = "Super Admin"
	}

	fullName := strings.TrimSpace(dbInfo.FirstName + " " + dbInfo.LastName)
	if fullName == "" {
		fullName = strings.Split(dbInfo.Email, "@")[0]
	}

	return &CurrentUserResponseData{
		User: UserResponse{
			ID:        dbInfo.UserID,
			Email:     dbInfo.Email,
			FullName:  fullName,
			FirstName: dbInfo.FirstName,
			LastName:  dbInfo.LastName,
		},
		Org: OrgResponse{
			ID:   dbInfo.OrgID,
			Name: dbInfo.OrgName,
		},
		Role: RoleResponse{
			Name:        dbInfo.RoleName,
			DisplayName: displayRole,
			Permissions: permissions,
		},
	}, nil
}

func (s *Service) ForgotPassword(ctx context.Context, req ForgotPasswordRequest) error {
	if err := utils.ValidateEmail(req.Email); err != nil {
		return err
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)

	_, err := s.client.ForgotPassword(ctx, &cognitoidentityprovider.ForgotPasswordInput{
		ClientId:   &s.cfg.CognitoClientID,
		Username:   &req.Email,
		SecretHash: &secretHash,
	})
	return err
}

func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	if err := utils.ValidateEmail(req.Email); err != nil {
		return err
	}
	if err := utils.ValidateRequired(req.Code, "code"); err != nil {
		return err
	}
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, req.Email, s.cfg.CognitoClientID)

	_, err := s.client.ConfirmForgotPassword(ctx, &cognitoidentityprovider.ConfirmForgotPasswordInput{
		ClientId:         &s.cfg.CognitoClientID,
		Username:         &req.Email,
		ConfirmationCode: &req.Code,
		Password:         &req.NewPassword,
		SecretHash:       &secretHash,
	})
	return err
}

type AcceptInviteRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type ValidateInviteResponse struct {
	Email    string `json:"email" db:"email"`
	OrgName  string `json:"org_name" db:"org_name"`
	RoleName string `json:"role_name" db:"role_name"`
}

func (s *Service) ValidateInvite(ctx context.Context, token string) (*ValidateInviteResponse, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	var res ValidateInviteResponse
	query := `
		SELECT i.email, o.name as org_name, r.name as role_name 
		FROM invitations i
		JOIN organizations o ON i.org_id = o.id
		JOIN roles r ON i.role_id = r.id
		WHERE i.token = ? AND i.expires_at > NOW()
	`
	err := s.db.GetContext(ctx, &res, query, token)
	if err != nil {
		return nil, errors.New("invalid or expired invitation token")
	}

	return &res, nil
}

func (s *Service) AcceptInvite(ctx context.Context, req AcceptInviteRequest) error {
	if req.Token == "" || req.Password == "" || req.FullName == "" {
		return errors.New("token, password, and full name are required")
	}

	var inv struct {
		ID        int64  `db:"id"`
		OrgID     int64  `db:"org_id"`
		RoleID    int64  `db:"role_id"`
		Email     string `db:"email"`
		ExpiresAt string `db:"expires_at"`
	}

	err := s.db.GetContext(ctx, &inv, `SELECT id, org_id, role_id, email, expires_at FROM invitations WHERE token = ? AND expires_at > NOW()`, req.Token)
	if err != nil {
		return errors.New("invalid or expired invitation token")
	}

	secretHash := ComputeSecretHash(s.cfg.CognitoClientSecret, inv.Email, s.cfg.CognitoClientID)

	// Fetch Organization Name for Cognito Custom Attribute if required
	var orgName string
	_ = s.db.GetContext(ctx, &orgName, `SELECT name FROM organizations WHERE id = ?`, inv.OrgID)

	signUpInput := &cognitoidentityprovider.SignUpInput{
		ClientId:   &s.cfg.CognitoClientID,
		Username:   &inv.Email,
		Password:   &req.Password,
		SecretHash: &secretHash,
		UserAttributes: []types.AttributeType{
			{Name: stringPtr("email"), Value: &inv.Email},
			{Name: stringPtr("name"), Value: &req.FullName},
			{Name: stringPtr("custom:company_name"), Value: &orgName},
		},
	}

	_, err = s.client.SignUp(ctx, signUpInput)
	if err != nil {
		// If user already exists in cognito, we might still want to link them to the org
		// For simplicity, we assume they are new or fail. To be robust, we'd check if error is UsernameExistsException.
		if !strings.Contains(err.Error(), "UsernameExistsException") {
			return fmt.Errorf("failed to register user in identity provider: %w", err)
		}
	} else {
		// Automatically confirm the user and mark email as verified since they used an invitation link
		if s.cfg.CognitoUserPoolID != "" {
			_, confirmErr := s.client.AdminConfirmSignUp(ctx, &cognitoidentityprovider.AdminConfirmSignUpInput{
				UserPoolId: &s.cfg.CognitoUserPoolID,
				Username:   &inv.Email,
			})
			if confirmErr != nil {
				log.Printf("[AUTH ERROR] AdminConfirmSignUp failed for invited user %s: %v", inv.Email, confirmErr)
			}
			
			_, updateErr := s.client.AdminUpdateUserAttributes(ctx, &cognitoidentityprovider.AdminUpdateUserAttributesInput{
				UserPoolId: &s.cfg.CognitoUserPoolID,
				Username:   &inv.Email,
				UserAttributes: []types.AttributeType{
					{Name: stringPtr("email_verified"), Value: stringPtr("true")},
				},
			})
			if updateErr != nil {
				log.Printf("[AUTH ERROR] AdminUpdateUserAttributes failed for invited user %s: %v", inv.Email, updateErr)
			}
		}
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert or update User record
	_, err = tx.ExecContext(ctx, `
		INSERT INTO users (cognito_sub, email, first_name)
		VALUES (?, ?, ?)
		ON DUPLICATE KEY UPDATE first_name = VALUES(first_name)
	`, inv.Email, inv.Email, req.FullName)
	if err != nil {
		return fmt.Errorf("failed to create user record: %w", err)
	}

	var userID int64
	err = tx.QueryRowxContext(ctx, `SELECT id FROM users WHERE email = ? LIMIT 1`, inv.Email).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to get user record id: %w", err)
	}

	// Add membership
	_, err = tx.ExecContext(ctx, `
		INSERT INTO org_members (org_id, user_id, role_id, status, created_at, updated_at)
		VALUES (?, ?, ?, 'ACTIVE', NOW(), NOW())
		ON DUPLICATE KEY UPDATE role_id = VALUES(role_id), status = 'ACTIVE'
	`, inv.OrgID, userID, inv.RoleID)

	if err != nil {
		return fmt.Errorf("failed to assign user to organization: %w", err)
	}

	// Invalidate token
	_, err = tx.ExecContext(ctx, `DELETE FROM invitations WHERE id = ?`, inv.ID)
	if err != nil {
		return fmt.Errorf("failed to clean up invitation: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func stringPtr(s string) *string {
	return &s
}
