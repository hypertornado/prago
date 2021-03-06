package prago

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/chris-ramon/douceur/inliner"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago/utils"
)

var (
	ErrEmailAlreadyInList = errors.New("email already in newsletter list")
)

type NewsletterMiddleware struct {
	baseUrl    string
	renderer   NewsletterRenderer
	app        *App
	randomness string
}

func (app *App) InitNewsletter(renderer NewsletterRenderer) {
	app.Newsletter = &NewsletterMiddleware{
		baseUrl:  app.Config.GetString("baseUrl"),
		renderer: renderer,

		app:        app,
		randomness: app.Config.GetString("random"),
	}

	var importPath string
	app.AddCommand("newsletter", "import").StringArgument(&importPath).Callback(func() {
		file, err := os.Open(importPath)
		must(err)

		data, err := ioutil.ReadAll(file)
		must(err)

		lines := []string{string(data)}
		for _, sep := range []string{"\r\n", "\r", "\n"} {
			l2 := []string{}
			for _, v := range lines {
				l2 = append(l2, strings.Split(v, sep)...)
			}
			lines = l2
		}

		for _, v := range lines {
			err := app.AddEmail(v, "", true)
			if err != nil {
				fmt.Printf("error while importing %s: %s\n", v, err)
			}
		}
	})

	controller := app.MainController().SubController()
	controller.AddBeforeAction(func(request Request) {
		request.SetData("site", app.HumanName)
	})

	controller.Get("/newsletter-subscribe", func(request Request) {
		request.SetData("title", "Přihlásit se k odběru newsletteru")
		request.SetData("csrf", app.Newsletter.CSFR(request))
		request.SetData("yield", "newsletter_subscribe")
		request.SetData("show_back_button", true)
		request.RenderView("newsletter_layout")
	})

	controller.Post("/newsletter-subscribe", func(request Request) {
		if app.Newsletter.CSFR(request) != request.Params().Get("csrf") {
			panic("wrong csrf")
		}

		email := request.Params().Get("email")
		email = strings.Trim(email, " ")
		name := request.Params().Get("name")

		var message string
		err := app.AddEmail(email, name, false)
		if err == nil {
			err := app.sendConfirmEmail(name, email)
			if err != nil {
				panic(err)
			}
			message = "Na váš email " + email + " jsme odeslali potvrzovací email k odebírání newsletteru."
		} else {
			if err == ErrEmailAlreadyInList {
				message = "Email se již nachází v naší emailové databázi"
			} else {
				panic(err)
			}
		}

		request.SetData("show_back_button", true)
		request.SetData("title", message)
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	controller.Get("/newsletter-confirm", func(request Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.Newsletter.secret(email) != secret {
			panic("wrong secret")
		}

		var person NewsletterPersons
		err := app.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Confirmed = true
		err = app.Save(&person)
		if err != nil {
			panic(err)
		}

		request.SetData("show_back_button", true)
		request.SetData("title", "Odběr newsletteru potvrzen")
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	controller.Get("/newsletter-unsubscribe", func(request Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.Newsletter.secret(email) != secret {
			panic("wrong secret")
		}

		var person NewsletterPersons
		err := app.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Unsubscribed = true
		err = app.Save(&person)
		if err != nil {
			panic(err)
		}

		request.SetData("show_back_button", true)
		request.SetData("title", "Odhlášení z odebírání newsletteru proběhlo úspěšně.")
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	app.CreateResource(Newsletter{}, initNewsletterResource)
	app.CreateResource(NewsletterSection{}, initNewsletterSection)
	app.CreateResource(NewsletterPersons{}, initNewsletterPersonsResource)

	//newsletterResource.AddRelation(newsletterSectionResource, "Newsletter", Unlocalized("Přidat sekci"))
}

func (app App) sendConfirmEmail(name, email string) error {

	text := app.Newsletter.confirmEmailBody(name, email)

	return app.SendEmail(
		name,
		email,
		"Potvrďte prosím odběr newsletteru "+app.HumanName,
		text,
		text,
	)
}

func (nm NewsletterMiddleware) confirmEmailBody(name, email string) string {
	values := make(url.Values)
	values.Set("email", email)
	values.Set("secret", nm.secret(email))

	u := fmt.Sprintf("%s/newsletter-confirm?%s",
		nm.baseUrl,
		values.Encode(),
	)

	return fmt.Sprintf("Potvrďte prosím odběr newsletteru z webu %s kliknutím na adresu:\n\n%s",
		nm.app.HumanName,
		u,
	)
}

func (nm NewsletterMiddleware) unsubscribeUrl(email string) string {
	values := make(url.Values)
	values.Set("email", email)
	values.Set("secret", nm.secret(email))

	return fmt.Sprintf("%s/newsletter-unsubscribe?%s",
		nm.baseUrl,
		values.Encode(),
	)
}

func (nm NewsletterMiddleware) secret(email string) string {
	h := md5.New()

	io.WriteString(h, fmt.Sprintf("secret%s%s", nm.randomness, email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (nm NewsletterMiddleware) CSFR(request Request) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", nm.randomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (app *App) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("Wrong email format")
	}

	err := app.Query().WhereIs("email", email).Get(&NewsletterPersons{})
	if err == nil {
		return ErrEmailAlreadyInList
	}

	person := NewsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return app.Create(&person)
}

type Newsletter struct {
	ID            int64     `prago-preview:"true" prago-order-desc:"true"`
	Name          string    `prago-preview:"true" prago-description:"Jméno newsletteru"`
	Body          string    `prago-type:"markdown"`
	PreviewSentAt time.Time `prago-preview:"true"`
	SentAt        time.Time `prago-preview:"true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterResource(resource *Resource) {
	resource.ActivityLog = true
	resource.CanView = "newsletter"

	resource.ResourceController.AddBeforeAction(func(request Request) {
		ret, err := resource.App.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Count(&NewsletterPersons{})
		if err != nil {
			panic(err)
		}
		request.SetData("recipients_count", ret)
	})

	previewAction := Action{
		Name: func(string) string { return "Náhled" },
		URL:  "preview",
		Handler: func(resource Resource, request Request, user User) {
			var newsletter Newsletter
			err := resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}

			body, err := resource.App.Newsletter.GetBody(newsletter, "")
			if err != nil {
				panic(err)
			}

			request.Response().WriteHeader(200)
			request.Response().Write([]byte(body))
		},
	}

	doSendPreviewAction := Action{
		URL:    "send-preview",
		Method: "post",
		Handler: func(resource Resource, request Request, user User) {
			var newsletter Newsletter
			must(resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))
			newsletter.PreviewSentAt = time.Now()
			must(resource.App.Save(&newsletter))

			emails := parseEmails(request.Params().Get("emails"))
			resource.App.sendEmails(newsletter, emails)
			AddFlashMessage(request, "Náhled newsletteru odeslán.")
			request.Redirect(resource.GetURL(""))
		},
	}

	doSendAction := Action{
		URL:    "send",
		Method: "post",
		Handler: func(resource Resource, request Request, user User) {
			var newsletter Newsletter
			err := resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}
			newsletter.SentAt = time.Now()
			resource.App.Save(&newsletter)

			recipients, err := resource.App.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}

			go resource.App.sendEmails(newsletter, recipients)

			request.SetData("recipients", recipients)
			request.SetData("recipients_count", len(recipients))
			request.SetData("admin_yield", "newsletter_sent")
			request.RenderView("admin_layout")
		},
	}

	doDuplicateAction := Action{
		URL:    "duplicate",
		Method: "post",
		Handler: func(resource Resource, request Request, user User) {
			var newsletter Newsletter
			must(resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))

			var sections []*NewsletterSection
			err := resource.App.Query().WhereIs("newsletter", newsletter.ID).Order("orderposition").Get(&sections)

			newsletter.ID = 0
			must(resource.App.Create(&newsletter))

			if err == nil {
				for _, v := range sections {
					section := *v
					section.ID = 0
					section.Newsletter = newsletter.ID
					must(resource.App.Create(&section))
				}
			}
			request.Redirect(resource.GetItemURL(&newsletter, "edit"))
		},
	}

	resource.AddItemAction(previewAction)
	resource.AddItemAction(
		CreateNavigationalItemAction(
			"send-preview",
			func(string) string { return "Odeslat náhled" },
			"newsletter_send_preview",
			nil,
		),
	)
	resource.AddItemAction(doSendPreviewAction)
	resource.AddItemAction(CreateNavigationalItemAction(
		"send",
		func(string) string { return "Odeslat" },
		"newsletter_send",
		func(Resource, Request, User) interface{} {
			recipients, err := resource.App.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}
			return map[string]interface{}{
				"recipients":       recipients,
				"recipients_count": len(recipients),
			}
		},
	))
	resource.AddItemAction(CreateNavigationalItemAction(
		"duplicate",
		func(string) string { return "Duplikovat" },
		"newsletter_duplicate",
		func(Resource, Request, User) interface{} {
			return nil
		},
	))
	resource.AddItemAction(doDuplicateAction)
	resource.AddItemAction(doSendAction)
}

func parseEmails(emails string) []string {
	ret := []string{}

	for _, v := range strings.Split(emails, "\n") {
		v = strings.Trim(v, " ")
		if len(v) > 0 {
			ret = append(ret, v)
		}
	}
	return ret
}

func (app *App) getNewsletterRecipients() ([]string, error) {
	ret := []string{}

	var persons []*NewsletterPersons
	err := app.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Get(&persons)
	if err != nil {
		return nil, err
	}

	for _, v := range persons {
		ret = append(ret, v.Email)
	}
	return ret, nil
}

func (app *App) sendEmails(n Newsletter, emails []string) error {
	for _, v := range emails {
		body, err := app.Newsletter.GetBody(n, v)
		if err == nil {

			err = app.SendEmail(
				"",
				v,
				n.Name,
				body,
				body,
			)
		}
		if err != nil {
			fmt.Println("ERROR", err.Error())
		}
	}
	return nil
}

func (nm *NewsletterMiddleware) GetBody(n Newsletter, email string) (string, error) {
	content := markdown.New(markdown.HTML(true)).RenderToString([]byte(n.Body))
	params := map[string]interface{}{
		"id":          n.ID,
		"baseUrl":     nm.baseUrl,
		"site":        nm.app.HumanName,
		"title":       n.Name,
		"unsubscribe": nm.unsubscribeUrl(email),
		"content":     template.HTML(content),
		"preview":     utils.CropMarkdown(n.Body, 200),
		"sections":    nm.getNewsletterSectionData(n),
	}

	if nm.renderer != nil {
		return nm.renderer(params)
	} else {
		return defaultNewsletterRenderer(params)
	}

}

type NewsletterRenderer func(map[string]interface{}) (string, error)

func defaultNewsletterRenderer(params map[string]interface{}) (string, error) {
	t, err := template.New("newsletter").Parse(defaultNewsletterTemplate)
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "newsletter", params)
	if err != nil {
		return "", err
	}

	ret, err := inliner.Inline(string(buf.Bytes()))
	if err != nil {
		return "", err
	}

	return ret, nil
}

type NewsletterSection struct {
	ID            int64
	Newsletter    int64  `prago-type:"relation" prago-preview:"true"`
	Name          string `prago-description:"Jméno sekce"`
	Text          string `prago-type:"text"`
	Button        string `prago-description:"Tlačítko"`
	URL           string `prago-description:"Odkaz"`
	Image         string `prago-type:"image" prago-preview:"true"`
	OrderPosition int64  `prago-type:"order"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterSection(resource *Resource) {
	resource.CanView = "newsletter"
}

type NewsletterPersons struct {
	ID           int64
	Name         string `prago-preview:"true" prago-description:"Jméno příjemce"`
	Email        string `prago-preview:"true"`
	Confirmed    bool   `prago-preview:"true"`
	Unsubscribed bool   `prago-preview:"true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func initNewsletterPersonsResource(resource *Resource) {
	resource.CanView = "sysadmin"
	resource.ActivityLog = true
}

type NewsletterSectionData struct {
	Name   string
	Text   string
	Button string
	URL    string
	Image  string
}

func (nm *NewsletterMiddleware) getNewsletterSectionData(n Newsletter) []NewsletterSectionData {
	var sections []*NewsletterSection
	err := nm.app.Query().WhereIs("newsletter", n.ID).Order("orderposition").Get(&sections)
	if err != nil {
		return nil
	}

	var ret []NewsletterSectionData

	for _, v := range sections {
		button := "Zjistit více"
		if v.Button != "" {
			button = v.Button
		}

		url := nm.baseUrl
		if v.URL != "" {
			url = v.URL
		}

		image := ""
		files := nm.app.GetFiles(v.Image)
		if len(files) > 0 {
			image = files[0].GetMedium()
		}

		ret = append(ret, NewsletterSectionData{
			Name:   v.Name,
			Text:   v.Text,
			Button: button,
			URL:    url,
			Image:  image,
		})
	}
	return ret
}
