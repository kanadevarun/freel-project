package notifications

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"html/template"
	"log"
	"net/smtp"
	"path/filepath"
	"time"
)

// smtpServiceImpl is the concrete implementation of the notifications Service using SMTP.
type smtpServiceImpl struct {
	host        string
	port        string
	username    string
	password    string
	fromAddress string
	templateDir string
	frontendURL string
}

// NewSMTPService creates and returns a new SMTP-backed notification service.
func NewSMTPService(host, port, username, password, fromAddress, templateDir, frontendURL string) Service {
	return &smtpServiceImpl{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromAddress: fromAddress,
		templateDir: templateDir,
		frontendURL: frontendURL,
	}
}

func (s *smtpServiceImpl) send(to []string, subject string, htmlBody string, textBody string) error {
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Generate a Message-ID
	randBytes := make([]byte, 16)
	rand.Read(randBytes)
	msgID := fmt.Sprintf("<%x@logisticshq.in>", randBytes)

	// Construct email headers
	boundary := fmt.Sprintf("boundary-%x", randBytes)
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: LogisticsHQ <%s>\r\n", s.fromAddress))
	buf.WriteString(fmt.Sprintf("Reply-To: no-reply@logisticshq.in\r\n"))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", to[0]))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString(fmt.Sprintf("Message-ID: %s\r\n", msgID))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")

	// Plain text part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(textBody)
	buf.WriteString("\r\n\r\n")

	// HTML part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n\r\n")

	// End boundary
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	err := smtp.SendMail(addr, auth, s.fromAddress, to, buf.Bytes())
	if err != nil {
		return fmt.Errorf("failed to send SMTP email: %w", err)
	}

	return nil
}

// SendEmail sends a generic email.
func (s *smtpServiceImpl) SendEmail(ctx context.Context, to, subject, body string) error {
	// For basic text emails, we can wrap it in basic HTML or just send as HTML
	return s.send([]string{to}, subject, body, body)
}

func (s *smtpServiceImpl) SendWhatsApp(ctx context.Context, phone, message string) error {
	log.Printf("Mock WhatsApp message sent to %s: %s", phone, message)
	return nil
}

// SendInviteEmail sends an HTML invite email.
func (s *smtpServiceImpl) SendInviteEmail(ctx context.Context, toEmail, token, orgName string) error {
	tmplPath := filepath.Join(s.templateDir, "invite.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return fmt.Errorf("failed to parse invite email template: %w", err)
	}

	// Make sure frontendURL doesn't end with slash
	baseURL := s.frontendURL
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	inviteLink := fmt.Sprintf("%s/invite/accept?token=%s", baseURL, token)
	data := struct {
		OrgName    string
		InviteLink string
	}{
		OrgName:    orgName,
		InviteLink: inviteLink,
	}

	var htmlBody bytes.Buffer
	if err := tmpl.Execute(&htmlBody, data); err != nil {
		return fmt.Errorf("failed to execute invite email template: %w", err)
	}

	subject := fmt.Sprintf("You’re invited to join %s on LogisticsHQ", orgName)
	textBody := fmt.Sprintf(`You're invited to join %s on LogisticsHQ.

You've been invited to collaborate with your organization on LogisticsHQ and help manage your logistics operations from one place.

Accept your invitation:
%s

If you weren't expecting this invitation, you can safely ignore this email.

© LogisticsHQ
This is an automated invitation email. Please do not reply to this message.`, orgName, inviteLink)

	err = s.send([]string{toEmail}, subject, htmlBody.String(), textBody)
	if err != nil {
		log.Printf("Failed to send invite email to %s via SMTP: %v", toEmail, err)
		return err
	}

	return nil
}

func (s *smtpServiceImpl) GetUnreadNotifications(ctx context.Context, orgID int32) ([]Notification, error) {
	return []Notification{}, nil
}

func (s *smtpServiceImpl) MarkAsRead(ctx context.Context, orgID int32, notifID int32) error {
	return nil
}
