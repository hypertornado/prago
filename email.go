package prago

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Email struct {
	app              *App
	from             *emailAddress
	to               *emailAddress
	attachements     []*mail.Attachment
	subject          string
	plainTextContent string
	htmlContent      string
}

type emailAddress struct {
	Name  string
	Email string
}

func (e *emailAddress) toSendgridEmail() *mail.Email {
	return mail.NewEmail(e.Name, e.Email)
}

func newEmailAddress(name, email string) *emailAddress {
	return &emailAddress{name, email}
}

func (app *App) Email() *Email {
	noReplyEmail := app.MustGetSetting(context.TODO(), "no_reply_email")
	return &Email{
		from: newEmailAddress(app.name("en"), noReplyEmail),
		app:  app,
	}
}

func (email *Email) From(name, emailAddress string) *Email {
	email.from = newEmailAddress(name, emailAddress)
	return email
}

func (email *Email) To(name, emailAddress string) *Email {
	email.to = newEmailAddress(name, emailAddress)
	return email
}

func (email *Email) Subject(subject string) *Email {
	email.subject = subject
	return email
}

func (email *Email) Attachement(filename string, content []byte) *Email {
	attachement := mail.NewAttachment()
	attachement.SetFilename(filename)
	attachement.SetContent(
		base64.StdEncoding.EncodeToString(content),
	)
	email.attachements = append(email.attachements, attachement)
	return email
}

func (email *Email) TextContent(content string) *Email {
	email.plainTextContent = content
	if email.htmlContent == "" {
		email.htmlContent = content
	}
	return email
}

func (email *Email) HTMLContent(content string) *Email {
	email.htmlContent = content
	if email.plainTextContent == "" {
		email.plainTextContent = content
	}
	return email
}

func (email *Email) Send() error {
	from := email.from.toSendgridEmail()
	to := email.to.toSendgridEmail()
	emailMessage := mail.NewSingleEmail(from, email.subject, to, email.plainTextContent, email.htmlContent)

	for _, v := range email.attachements {
		emailMessage.AddAttachment(v)
	}

	key, err := email.app.GetSetting(context.TODO(), "sendgrid_key")
	if err != nil {
		return err
	}

	sendGridClient := sendgrid.NewSendClient(
		key,
	)

	resp, err := sendGridClient.Send(emailMessage)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("email could not be sent, code %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}
