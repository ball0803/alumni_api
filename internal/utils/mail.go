package utils

import (
	"alumni_api/config"
	"alumni_api/internal/utils/mail_format"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net/smtp"
	"strings"
	// "gopkg.in/gomail.v2"
)

func sendEmailHTML(toEmail, subject, html string) error {
	fromEmail := config.GetEnv("SENDER_GMAIL", "")
	fromName := "CPE Alumni"
	password := config.GetEnv("SMTP_PASSWORD", "")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	fromHeader := fmt.Sprintf("%s <%s>", fromName, fromEmail)

	// Properly formatted MIME headers
	headers := make(map[string]string)
	headers["From"] = fromHeader
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=\"UTF-8\""

	// Build the message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + html) // Separate headers and body with \r\n

	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		fromEmail,
		[]string{toEmail},
		[]byte(msg.String()),
	)
	return err
}

func sendEmail(toEmail, subject, body string) error {
	// Gmail SMTP Configuration
	from := config.GetEnv("", "")
	password := config.GetEnv("SMTP_PASSWORD", "")
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	// Email Content
	msg := fmt.Sprintf("Subject: %s\n\n%s", subject, body)
	// SMTP Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)
	// Send Email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%s", smtpHost, smtpPort),
		auth,
		from,
		[]string{toEmail},
		[]byte(msg),
	)
	return err
}

func sendEmailSendGrid(toEmail, subject, body string) error {
	from := mail.NewEmail("CPE Alumni", "phurin.reongsang@gmail.com")
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, body, body)
	client := sendgrid.NewSendClient(config.GetEnv("SENDGUN_API_KEY", ""))

	_, err := client.Send(message)
	return err
}

func SendOneTimeRegistryEmailSucc(email, token, ref string) error {
	subject := "Alumni One Time Registration"
	body := fmt.Sprintf(mail_format.OneTimeRegistrySucc, token, ref)
	if err := sendEmailHTML(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendOneTimeRegistryEmailFail(email, ref string) error {
	subject := "Alumni One Time Registration"
	body := fmt.Sprintf(mail_format.OneTimeRegistryFail, ref)
	if err := sendEmailHTML(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendVerificationEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(mail_format.VerifyMail, token)
	if err := sendEmailHTML(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendVerificationChangeEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(mail_format.VerifyChangeMail, token)
	if err := sendEmailHTML(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendResetMail(email, token string) error {
	subject := "Alumni Password Reset"
	body := fmt.Sprintf(mail_format.ResetPasswordMail, token)
	if err := sendEmailHTML(email, subject, body); err != nil {
		return err
	}
	return nil
}
