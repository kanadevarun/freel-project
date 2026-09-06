package notifications

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ses"
	"github.com/aws/aws-sdk-go-v2/service/ses/types"
)

// sesServiceImpl is the concrete implementation of the notifications Service using AWS SES.
type sesServiceImpl struct {
	sesClient   *ses.Client
	fromAddress string
	templateDir string
}

// NewSESService creates and returns a new SES-backed notification service.
// It requires an initialized AWS SES client, the verified 'from' email address,
// and the path to the HTML email templates directory.
func NewSESService(client *ses.Client, fromAddress, templateDir string) Service {
	return &sesServiceImpl{
		sesClient:   client,
		fromAddress: fromAddress,
		templateDir: templateDir,
	}
}

// SendEmail sends a raw text email to a recipient.
// This is the generic email method used for basic alerts or notifications.
func (s *sesServiceImpl) SendEmail(ctx context.Context, to, subject, body string) error {
	input := &ses.SendEmailInput{
		Source: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(subject),
			},
			Body: &types.Body{
				Text: &types.Content{
					Data: aws.String(body),
				},
			},
		},
	}

	_, err := s.sesClient.SendEmail(ctx, input)
	if err != nil {
		log.Printf("Failed to send email via SES to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// SendWhatsApp sends a WhatsApp message (placeholder implementation).
// In a real scenario, this would integrate with Twilio or Meta's WhatsApp API.
func (s *sesServiceImpl) SendWhatsApp(ctx context.Context, phone, message string) error {
	log.Printf("Mock WhatsApp message sent to %s: %s", phone, message)
	return nil
}

// SendInviteEmail sends a beautifully formatted HTML email for inviting users.
// It loads the `invite.html` template, injects the dynamic link and organization name,
// and dispatches the email via AWS SES.
func (s *sesServiceImpl) SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error {
	// Parse the HTML template from the filesystem
	tmplPath := filepath.Join(s.templateDir, "invite.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse invite email template: %w", err)
	}

	// Prepare the data payload for the template
	inviteLink := fmt.Sprintf("https://logisticshq.in/accept-invite?token=%s", token)
	data := struct {
		OrgName    string
		InviteLink string
	}{
		OrgName:    orgName,
		InviteLink: inviteLink,
	}

	// Execute the template with the provided data
	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to execute invite email template: %w", err)
	}

	// Construct the SES email input with HTML content
	input := &ses.SendEmailInput{
		Source: aws.String(s.fromAddress),
		Destination: &types.Destination{
			ToAddresses: []string{toEmail},
		},
		Message: &types.Message{
			Subject: &types.Content{
				Data: aws.String(fmt.Sprintf("You’re invited to join %s on LogisticsHQ", orgName)),
			},
			Body: &types.Body{
				Html: &types.Content{
					Data: aws.String(htmlBody.String()),
				},
				Text: &types.Content{
					Data: aws.String(fmt.Sprintf(`You're invited to join %s on LogisticsHQ.

You've been invited to collaborate with your organization on LogisticsHQ and help manage your logistics operations from one place.

Accept your invitation:
%s

If you weren't expecting this invitation, you can safely ignore this email.

© LogisticsHQ
This is an automated invitation email. Please do not reply to this message.`, orgName, inviteLink)),
				},
			},
		},
	}

	// Dispatch the email
	_, err = s.sesClient.SendEmail(ctx, input)
	if err != nil {
		log.Printf("Failed to send invite email to %s: %v", toEmail, err)
		return fmt.Errorf("failed to send invite email: %w", err)
	}

	return nil
}

func (s *sesServiceImpl) GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error) {
	return []Notification{}, nil
}

func (s *sesServiceImpl) MarkAsRead(ctx context.Context, orgID int32, notifID int32) error {
	return nil
}
