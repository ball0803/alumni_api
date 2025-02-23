package utils

import (
	"fmt"
	"gopkg.in/gomail.v2"
)

func SendVerificationEmail(email, token string) error {
	// Mailtrap SMTP configuration
	smtpHost := "sandbox.smtp.mailtrap.io"
	smtpPort := 2525
	smtpUser := "20f24dedb0b692"
	smtpPass := "1be3c10883319e"

	// Create the email
	m := gomail.NewMessage()
	m.SetHeader("From", "no-reply@kmutt.cpe.alumni.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Verify Your Email Address")
	m.SetBody("text/html", fmt.Sprintf(`
		<p>Hi,</p>
		<p>Thank you for registering! Please click the link below to verify your email address:</p>
		<p><a href="http://localhost:3000/v1/auth/verify-email?token=%s">Verify Email</a></p>
		<p>If you did not create an account, please ignore this email.</p>
	`, token))

	// Send the email
	d := gomail.NewDialer(smtpHost, smtpPort, smtpUser, smtpPass)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
