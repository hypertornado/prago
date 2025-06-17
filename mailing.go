package prago

import (
	"html/template"
	"strings"
)

type MailingData struct {
	App     *App
	BaseURL string
	LogoURL string
	AppName string

	FromEmail string
	FromName  string

	Tos []*MailingRecipient

	Subject string

	PreName  string
	Name     string
	PostName string

	Images []*File

	Description template.HTML

	Text template.HTML

	Sections []*MailingDataSection

	Button *Button

	FooterDescription template.HTML
}

type MailingDataSection struct {
	Name string
	Text string
}

type MailingRecipient struct {
	Name  string
	Email string
}

func (app *App) initMailing() {

	ActionPlain(app, "mailing-preview", func(request *Request) {

		request.WriteHTML(200, app.adminTemplates, "mailing", getTestingMailingData(app))

	}).Permission("sysadmin").Board(sysadminBoard)

	ActionForm(app, "send-mailing-preview", func(f *Form, r *Request) {
		f.AddTextInput("emails", "Emails")
		f.AddSubmit("Odeslat")
	}, func(fv FormValidation, request *Request) {
		emails := strings.Split(request.Param("emails"), ",")

		md := getTestingMailingData(app)
		for _, v := range emails {
			md.AddRecipient("", v)
		}

		err := sendMailingData(md)
		if err != nil {
			fv.AddError(err.Error())
		} else {
			fv.AddError("Email odeslán")
		}

	}).Permission("sysadmin").Board(sysadminBoard)
}

func getTestingMailingData(app *App) *MailingData {
	data := initMailingData("cs", app)
	data.Name = "Hello world"

	data.PostName = "test name"

	data.Description = "This is some description"
	data.Text = "This is some text"

	data.Button = &Button{
		Name: "More info",
		URL:  "https://www.seznam.cz",
	}

	data.AddSection("Pohlaví", "žena")
	data.AddSection("Poznámka", "Nejdřív bylo škubnutí, pak pilotův třesoucí se hlas a nakonec náraz a horko. Tak médiím popsal okamžiky před čtvrteční havárií letadla Air India na západě Indie jediný přeživší Viswashkumar Ramesh. Muž cestoval s dalšími 241 lidmi, kteří při pádu letadla do obydlené oblasti za letištěm v Ahmadábádu.")

	data.FooterDescription = "q oejw <a href=\"/xxxx\">ifoejwifoe</a> e"

	return data

}

func (md *MailingData) AddRecipient(name, email string) *MailingData {
	md.Tos = append(md.Tos, &MailingRecipient{
		Name:  name,
		Email: email,
	})
	return md
}

func (md *MailingData) AddSection(name, text string) *MailingDataSection {
	ret := &MailingDataSection{
		Name: name,
		Text: text,
	}
	md.Sections = append(md.Sections, ret)
	return ret
}

func initMailingData(locale string, app *App) *MailingData {
	ai := app.GetAppInfo()

	data := &MailingData{
		App:     app,
		BaseURL: app.BaseURL(),
		LogoURL: app.IconURL(),
		AppName: ai.Name(locale),

		PreName: ai.Name(locale),

		FromEmail: app.mustGetSetting("no_reply_email"),
		FromName:  ai.Name(locale),
	}
	return data
}

func (app *App) Mailing(locale string, fn func(*MailingData)) error {
	data := initMailingData(locale, app)
	fn(data)
	return sendMailingData(data)
}

func sendMailingData(data *MailingData) error {
	var subject = data.Subject
	if subject == "" {
		subject = data.Name
	}
	if subject == "" {
		subject = data.FromName
	}

	htmlContent := data.App.adminTemplates.ExecuteToString("mailing", data)

	email := data.App.Email().From(data.FromName, data.FromEmail).HTMLContent(htmlContent).Subject(subject)
	for _, v := range data.Tos {
		email.To(v.Name, v.Email)
	}

	return email.Send()

}
