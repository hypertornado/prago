package prago

import (
	"fmt"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

func (app *App) initEmail() {
	app.sendgridKey = app.ConfigurationGetStringWithFallback("sendgridApi", "")
	app.noReplyEmail = app.ConfigurationGetStringWithFallback("noReplyEmail", "")
	app.sendgridClient = sendgrid.NewSendClient(app.sendgridKey)
}

type Email struct {
	app              *App
	from             *emailAddress
	to               *emailAddress
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
	return &Email{
		from: newEmailAddress(app.name("en"), app.noReplyEmail),
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
	resp, err := email.app.sendgridClient.Send(emailMessage)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("email could not be sent, code %d: %s", resp.StatusCode, resp.Body)
	}
	return nil
}

//SendEmail from app
/*func (a App) SendEmailOLD(name, email, subject, contentText, contentHTML string) error {
	from := mail.NewEmail(a.name("en"), a.noReplyEmail)
	to := mail.NewEmail(name, email)
	message := mail.NewSingleEmail(from, subject, to, contentText, contentHTML)
	client := sendgrid.NewSendClient(a.sendgridKey)
	_, err := client.Send(message)
	return err
}*/

//SendEmailFromTo send email with from data
/*func (a App) SendEmailFromTo(fromEmail, toEmail, subject, contentText, contentHTML string) error {
	from := mail.NewEmail("", fromEmail)
	to := mail.NewEmail("", toEmail)
	message := mail.NewSingleEmail(from, subject, to, contentText, contentHTML)
	client := sendgrid.NewSendClient(a.sendgridKey)
	_, err := client.Send(message)
	return err
}*/
