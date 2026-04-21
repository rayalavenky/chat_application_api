package services

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"mime"
	"net/smtp"
	"os"
	"strings"
)

type welcomeEmailData struct {
	FirstName string
	Email     string
	Password  string
	AppName   string
}

const welcomeEmailTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Welcome to {{.AppName}}</title>
</head>
<body style="margin:0;padding:0;background:#f2f4f8;font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',Roboto,Helvetica,Arial,sans-serif;color:#1f2937;">
  <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f2f4f8;padding:40px 0;">
    <tr>
      <td align="center">
        <table role="presentation" width="560" cellpadding="0" cellspacing="0" style="background:#ffffff;border-radius:12px;box-shadow:0 4px 16px rgba(15,23,42,0.08);overflow:hidden;">
          <tr>
            <td style="background:#1E7A5A;padding:32px 40px;color:#ffffff;">
              <h1 style="margin:0;font-size:24px;font-weight:700;">Welcome to {{.AppName}}, {{.FirstName}}!</h1>
              <p style="margin:8px 0 0;font-size:14px;opacity:0.9;">Your account has been created successfully.</p>
            </td>
          </tr>
          <tr>
            <td style="padding:32px 40px;">
              <p style="margin:0 0 16px;font-size:15px;line-height:1.6;">
                Hi {{.FirstName}}, we're excited to have you on board. Use the credentials below to log in to your account.
              </p>
              <table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:#f8fafc;border:1px solid #e2e8f0;border-radius:8px;margin:24px 0;">
                <tr>
                  <td style="padding:16px 20px;font-size:14px;">
                    <div style="color:#64748b;margin-bottom:4px;">Email</div>
                    <div style="font-weight:600;color:#0f172a;">{{.Email}}</div>
                  </td>
                </tr>
                <tr>
                  <td style="padding:16px 20px;font-size:14px;border-top:1px solid #e2e8f0;">
                    <div style="color:#64748b;margin-bottom:4px;">Temporary password</div>
                    <div style="font-family:Menlo,Consolas,monospace;font-size:16px;font-weight:700;color:#4f46e5;letter-spacing:1px;">{{.Password}}</div>
                  </td>
                </tr>
              </table>
              <p style="margin:0 0 8px;font-size:14px;line-height:1.6;color:#475569;">
                For security, please change your password after logging in.
              </p>
              <p style="margin:0;font-size:13px;line-height:1.6;color:#94a3b8;">
                If you did not sign up for {{.AppName}}, you can safely ignore this email.
              </p>
            </td>
          </tr>
          <tr>
            <td style="padding:20px 40px;background:#f8fafc;border-top:1px solid #e2e8f0;text-align:center;font-size:12px;color:#94a3b8;">
              &copy; {{.AppName}}. All rights reserved.
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`

var welcomeTmpl = template.Must(template.New("welcome").Parse(welcomeEmailTemplate))

func SendWelcomeEmail(toEmail, firstName, password string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")
	appName := os.Getenv("APP_NAME")

	if smtpFrom == "" {
		smtpFrom = smtpUser
	}
	if appName == "" {
		appName = "OrbiTalk"
	}
	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		return errors.New("smtp configuration is not set")
	}

	if containsCRLF(toEmail) || containsCRLF(smtpFrom) || containsCRLF(firstName) || containsCRLF(appName) {
		return errors.New("invalid characters in email header fields")
	}

	var body bytes.Buffer
	data := welcomeEmailData{
		FirstName: firstName,
		Email:     toEmail,
		Password:  password,
		AppName:   appName,
	}
	if err := welcomeTmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to render email template: %w", err)
	}

	subject := mime.QEncoding.Encode("utf-8", fmt.Sprintf("Welcome to %s - Your Account Password", appName))
	headers := "From: " + smtpFrom + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n" +
		"\r\n"

	msg := append([]byte(headers), body.Bytes()...)

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	addr := smtpHost + ":" + smtpPort

	if err := smtp.SendMail(addr, auth, smtpFrom, []string{toEmail}, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func containsCRLF(s string) bool {
	return strings.ContainsAny(s, "\r\n")
}

func SendSimpleEmail(toEmail, subject, body string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	smtpFrom := os.Getenv("SMTP_FROM")

	if smtpFrom == "" {
		smtpFrom = smtpUser
	}

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPassword == "" {
		return errors.New("smtp configuration is not set")
	}

	// Prevent header injection
	if containsCRLF(toEmail) || containsCRLF(subject) {
		return errors.New("invalid characters in email")
	}

	// Build message
	msg := "From: " + smtpFrom + "\r\n" +
		"To: " + toEmail + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=\"utf-8\"\r\n" +
		"\r\n" +
		body

	auth := smtp.PlainAuth("", smtpUser, smtpPassword, smtpHost)
	addr := smtpHost + ":" + smtpPort

	err := smtp.SendMail(addr, auth, smtpFrom, []string{toEmail}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func SendOTPEmail(toEmail, firstName, otp string) error {
	subject := "Password reset otp"
	body := fmt.Sprintf("Hello %s,\n\nYour OTP is: %s\nIt will expire in 10 minutes.", firstName, otp)
	return SendSimpleEmail(toEmail, subject, body)

}
