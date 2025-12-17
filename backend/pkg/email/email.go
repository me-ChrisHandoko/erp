package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"

	"backend/internal/config"
)

// EmailService handles email sending operations
// Supports SMTP with TLS, template rendering, and multiple recipient types
type EmailService struct {
	cfg *config.Config
}

// NewEmailService creates a new email service instance
func NewEmailService(cfg *config.Config) *EmailService {
	return &EmailService{
		cfg: cfg,
	}
}

// EmailData holds common email data for template rendering
type EmailData struct {
	RecipientName string
	CompanyName   string
	Subject       string
	// Template-specific fields will be added via map[string]interface{}
}

// SendPasswordResetEmail sends a password reset email with token link
// Reference: PHASE2-MVP-ANALYSIS.md lines 180-220
func (s *EmailService) SendPasswordResetEmail(to, recipientName, resetToken string) error {
	// Build reset URL
	// TODO: Get frontend URL from config
	resetURL := fmt.Sprintf("http://localhost:3000/reset-password?token=%s", resetToken)

	// Prepare template data
	data := map[string]interface{}{
		"RecipientName": recipientName,
		"CompanyName":   s.cfg.Server.AppName,
		"ResetURL":      resetURL,
		"ExpiryMinutes": int(s.cfg.Email.PasswordResetExpiry.Minutes()),
	}

	// Render email templates
	htmlBody, err := s.renderTemplate("password_reset.html", data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	plainBody, err := s.renderTemplate("password_reset.txt", data)
	if err != nil {
		return fmt.Errorf("failed to render plain text template: %w", err)
	}

	// Send email
	subject := "Reset Your Password"
	return s.sendEmail(to, subject, htmlBody, plainBody)
}

// SendEmailVerificationEmail sends email verification link
// Placeholder for future implementation (Phase 2)
func (s *EmailService) SendEmailVerificationEmail(to, recipientName, verificationToken string) error {
	// Build verification URL
	verificationURL := fmt.Sprintf("http://localhost:3000/verify-email?token=%s", verificationToken)

	data := map[string]interface{}{
		"RecipientName":   recipientName,
		"CompanyName":     s.cfg.Server.AppName,
		"VerificationURL": verificationURL,
		"ExpiryHours":     int(s.cfg.Email.VerificationExpiry.Hours()),
	}

	htmlBody, err := s.renderTemplate("email_verification.html", data)
	if err != nil {
		return fmt.Errorf("failed to render HTML template: %w", err)
	}

	plainBody, err := s.renderTemplate("email_verification.txt", data)
	if err != nil {
		return fmt.Errorf("failed to render plain text template: %w", err)
	}

	subject := "Verify Your Email Address"
	return s.sendEmail(to, subject, htmlBody, plainBody)
}

// sendEmail sends an email via SMTP with both HTML and plain text versions
func (s *EmailService) sendEmail(to, subject, htmlBody, plainBody string) error {
	// Validate SMTP configuration
	if s.cfg.Email.SMTPHost == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	// Build email message with MIME multipart
	from := fmt.Sprintf("%s <%s>", s.cfg.Email.SMTPFromName, s.cfg.Email.SMTPFromEmail)

	var msg bytes.Buffer
	msg.WriteString(fmt.Sprintf("From: %s\r\n", from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: multipart/alternative; boundary=\"boundary-string\"\r\n")
	msg.WriteString("\r\n")

	// Plain text version
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(plainBody)
	msg.WriteString("\r\n")

	// HTML version
	msg.WriteString("--boundary-string\r\n")
	msg.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(htmlBody)
	msg.WriteString("\r\n")

	msg.WriteString("--boundary-string--\r\n")

	// SMTP authentication
	auth := smtp.PlainAuth("", s.cfg.Email.SMTPUser, s.cfg.Email.SMTPPassword, s.cfg.Email.SMTPHost)

	// SMTP address
	addr := fmt.Sprintf("%s:%d", s.cfg.Email.SMTPHost, s.cfg.Email.SMTPPort)

	// Send email with TLS if enabled
	if s.cfg.Email.SMTPTLS {
		return s.sendEmailTLS(addr, auth, s.cfg.Email.SMTPFromEmail, []string{to}, msg.Bytes())
	}

	// Send without TLS (not recommended for production)
	return smtp.SendMail(addr, auth, s.cfg.Email.SMTPFromEmail, []string{to}, msg.Bytes())
}

// sendEmailTLS sends email with explicit TLS connection
// Recommended for production use to ensure encrypted transmission
func (s *EmailService) sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte) error {
	// Create TLS configuration
	tlsConfig := &tls.Config{
		ServerName: s.cfg.Email.SMTPHost,
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2+
	}

	// Connect to SMTP server
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, s.cfg.Email.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send email data
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send DATA command: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return nil
}

// renderTemplate renders an email template with given data
func (s *EmailService) renderTemplate(templateName string, data interface{}) (string, error) {
	// Template path
	templatePath := filepath.Join("pkg", "email", "templates", templateName)

	// Parse template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template %s: %w", templateName, err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return buf.String(), nil
}
