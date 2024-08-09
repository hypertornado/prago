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

func logEmailSent(email *Email, err error) {

	var attachements []string

	for _, attachement := range email.attachements {
		attachements = append(attachements, attachement.Filename)
	}

	fmt.Println(len(email.plainTextContent))
	fmt.Println(email.plainTextContent)

	logEmail := &emailSent{
		Name: fmt.Sprintf("%s %s", email.to.Name, email.to.Email),
		From: fmt.Sprintf("%s %s", email.from.Name, email.from.Email),

		Subject: email.subject,

		Attachements: strings.Join(attachements, ", "),

		PlainTextContent: trimTextForMYSQLTextType(email.plainTextContent),
		HTMLContent:      trimTextForMYSQLTextType(email.htmlContent),
	}

	if err != nil {
		logEmail.Error = err.Error()
	}

	must(CreateItem(email.app, logEmail))
}

func trimTextForMYSQLTextType(in string) string {
	if len(in) > 65000 {
		in = in[0:65000]
	}
	return in
}
