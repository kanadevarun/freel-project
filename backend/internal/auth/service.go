package auth

import (
	"context"
	"errors"
	"log"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/freel/backend/internal/config"
	"github.com/freel/backend/internal/utils"
)

type Service struct {
	cfg    *config.Config
	client *cognitoidentityprovider.Client
}

func NewService(cfg *config.Config) *Service {
	awsCfg, err := awsConfig.LoadDefaultConfig(context.TODO(), awsConfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		log.Fatalf("unable to load AWS config, %v", err)
	}

	return &Service{
		cfg:    cfg,
		client: cognitoidentityprovider.NewFromConfig(awsCfg),
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

	allowedRoles := map[string]bool{
		"shipper":           true,
		"carrier":           true,
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
	return err
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
		return nil, err
	}

	if resp.AuthenticationResult == nil {
		return nil, errors.New("authentication result is missing")
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

func stringPtr(s string) *string {
	return &s
}
