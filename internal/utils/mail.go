package utils

import (
	"alumni_api/config"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	// "gopkg.in/gomail.v2"
)

func sendEmail(toEmail, subject, body string) error {
	from := mail.NewEmail("CPE Alumni", "phurin.reongsang@gmail.com") // Your verified sender email
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, body, body)
	client := sendgrid.NewSendClient(config.GetEnv("SENDGUN_API_KEY", ""))

	_, err := client.Send(message)
	return err
}

func SendVerificationEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(VerifyMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendVerificationChangeEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(VerifyChangeMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendResetMail(email, token string) error {
	subject := "Alumni Password Reset"
	body := fmt.Sprintf(ResetPasswordMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}
