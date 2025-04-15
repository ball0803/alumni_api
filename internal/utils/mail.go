package utils

import (
	"alumni_api/config"
	"alumni_api/internal/utils/mail_format"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	// "gopkg.in/gomail.v2"
)

func sendEmail(toEmail, subject, body string) error {
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
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendOneTimeRegistryEmailFail(email, ref string) error {
	subject := "Alumni One Time Registration"
	body := fmt.Sprintf(mail_format.OneTimeRegistryFail, ref)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendVerificationEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(mail_format.VerifyMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendVerificationChangeEmail(email, token string) error {
	subject := "Alumni Verification"
	body := fmt.Sprintf(mail_format.VerifyChangeMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}

func SendResetMail(email, token string) error {
	subject := "Alumni Password Reset"
	body := fmt.Sprintf(mail_format.ResetPasswordMail, token)
	if err := sendEmail(email, subject, body); err != nil {
		return err
	}
	return nil
}
