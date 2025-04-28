package utils

import (
	"alumni_api/config"
	"alumni_api/internal/utils/mail_format"
	"fmt"

	"crypto/tls"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"net"
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

	// 1. Build the email headers and body (unchanged from your original code)
	fromHeader := fmt.Sprintf("%s <%s>", fromName, fromEmail)
	headers := map[string]string{
		"From":         fromHeader,
		"To":           toEmail,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n" + html)

	// 2. Manually handle the SMTP connection with TLS
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", smtpHost, smtpPort))
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	// 3. Create an SMTP client
	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// 4. Enable STARTTLS with custom TLS config
	tlsConfig := &tls.Config{
		ServerName: smtpHost, // Verify the server's certificate
		// InsecureSkipVerify: true, // ⚠️ Uncomment ONLY for testing (disable cert verification)
	}
	if err = client.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %v", err)
	}

	// 5. Authenticate
	auth := smtp.PlainAuth("", fromEmail, password, smtpHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %v", err)
	}

	// 6. Send the email
	if err = client.Mail(fromEmail); err != nil {
		return fmt.Errorf("mail failed: %v", err)
	}
	if err = client.Rcpt(toEmail); err != nil {
		return fmt.Errorf("rcpt failed: %v", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %v", err)
	}
	defer w.Close()

	if _, err = fmt.Fprint(w, msg.String()); err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	return nil
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

func SendVerificationEmail(email, token, ref string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(mail_format.VerifyMail, token, ref)
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
