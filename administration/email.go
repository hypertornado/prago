package administration

import (
	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (a Administration) SendEmail(name, email, subject, contentText, contentHTML string) error {
	from := mail.NewEmail(a.HumanName, a.noReplyEmail)
	to := mail.NewEmail(name, email)
	message := mail.NewSingleEmail(from, subject, to, contentText, contentHTML)
	client := sendgrid.NewSendClient(a.sendgridKey)
	_, err := client.Send(message)
	return err
}

func (a Administration) SendEmailFromTo(fromEmail, toEmail, subject, contentText, contentHTML string) error {
	from := mail.NewEmail("", fromEmail)
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, contentText, contentHTML)
	client := sendgrid.NewSendClient(a.sendgridKey)
	_, err := client.Send(message)
	return err
}
