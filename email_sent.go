package prago

import (
	"fmt"
	"strings"
	"time"
)

type emailSent struct {
	ID               int64 `prago-order-desc:"true"`
	Name             string
	From             string
	Subject          string
	Attachements     string
	PlainTextContent string `prago-type:"text"`
	HTMLContent      string `prago-type:"text"`
	Error            string `prago-type:"text"`

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (app *App) initEmailSentResource() {

	resource := NewResource[emailSent](app)
	resource.Name(
		unlocalized("Log odeslaného emailu"),
		unlocalized("Log odeslaných emailů"),
	)
	resource.PermissionView("sysadmin")
	resource.Board(sysadminBoard)
}

/*
TODO: fix utf error
https://forum.golangbridge.org/t/golang-not-supporting-utf8-for-some-reason/21429/3

it does not work with emoji, encoding of database shoul be utf8mb4, not utf8

*/

func logEmailSent(email *Email, err error) {

	var attachements []string

	for _, attachement := range email.attachements {
		attachements = append(attachements, attachement.Filename)
	}

	var emailsStr []string
	for _, v := range email.to {
		emailsStr = append(emailsStr, fmt.Sprintf("%s %s", v.Name, v.Email))
	}

	logEmail := &emailSent{
		Name: strings.Join(emailsStr, ", "),
		From: fmt.Sprintf("%s %s", email.from.Name, email.from.Email),

		Subject: email.subject,

		Attachements: strings.Join(attachements, ", "),

		PlainTextContent: trimTextForMYSQLTextType(email.plainTextContent),
		HTMLContent:      trimTextForMYSQLTextType(email.htmlContent),
	}

	if err != nil {
		logEmail.Error = err.Error()
	}

	err = CreateItem(email.app, logEmail)
	if err != nil {
		email.app.Log().Errorf("can't save log error: %s", err)
	}
}

func trimTextForMYSQLTextType(in string) string {
	if len(in) > 65000 {
		in = in[0:65000]
	}
	return in
}
