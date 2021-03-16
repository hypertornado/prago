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

//ErrEmailAlreadyInList is returned when user is already in newsletter list
var ErrEmailAlreadyInList = errors.New("email already in newsletter list")

//NewsletterMiddleware represents users newsletter
type newsletterMiddleware struct {
	baseURL    string
	renderer   NewsletterRenderer
	app        *App
	randomness string
}

//InitNewsletter inits apps newsletter function
func (app *App) InitNewsletter(renderer NewsletterRenderer) {
	if app.newsletter != nil {
		panic("newsletter already initialized")
	}
	app.newsletter = &newsletterMiddleware{
		baseURL:  app.ConfigurationGetString("baseUrl"),
		renderer: renderer,

		app:        app,
		randomness: app.ConfigurationGetString("random"),
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

	controller := app.mainController.subController()
	controller.addBeforeAction(func(request Request) {
		request.SetData("site", app.name("en"))
	})

	controller.get("/newsletter-subscribe", func(request Request) {
		request.SetData("title", "Přihlásit se k odběru newsletteru")
		request.SetData("csrf", app.newsletter.CSFR(request))
		request.SetData("yield", "newsletter_subscribe")
		request.SetData("show_back_button", true)
		request.RenderView("newsletter_layout")
	})

	controller.post("/newsletter-subscribe", func(request Request) {
		if app.newsletter.CSFR(request) != request.Params().Get("csrf") {
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

	controller.get("/newsletter-confirm", func(request Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.newsletter.secret(email) != secret {
			panic("wrong secret")
		}

		var person newsletterPersons
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

	controller.get("/newsletter-unsubscribe", func(request Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.newsletter.secret(email) != secret {
			panic("wrong secret")
		}

		var person newsletterPersons
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

	initNewsletterResource(app.Resource(newsletter{}))
	initNewsletterSection(app.Resource(newsletterSection{}))
	initNewsletterPersonsResource(app.Resource(newsletterPersons{}))

	//newsletterResource.AddRelation(newsletterSectionResource, "Newsletter", Unlocalized("Přidat sekci"))
}

func (app *App) NewsletterCSRF(request Request) string {
	return app.newsletter.CSFR(request)
}

func (app App) sendConfirmEmail(name, email string) error {

	text := app.newsletter.confirmEmailBody(name, email)

	return app.SendEmail(
		name,
		email,
		"Potvrďte prosím odběr newsletteru "+app.name("en"),
		text,
		text,
	)
}

func (nm newsletterMiddleware) confirmEmailBody(name, email string) string {
	values := make(url.Values)
	values.Set("email", email)
	values.Set("secret", nm.secret(email))

	u := fmt.Sprintf("%s/newsletter-confirm?%s",
		nm.baseURL,
		values.Encode(),
	)

	return fmt.Sprintf("Potvrďte prosím odběr newsletteru z webu %s kliknutím na adresu:\n\n%s",
		nm.app.name("en"),
		u,
	)
}

func (nm newsletterMiddleware) unsubscribeURL(email string) string {
	values := make(url.Values)
	values.Set("email", email)
	values.Set("secret", nm.secret(email))

	return fmt.Sprintf("%s/newsletter-unsubscribe?%s",
		nm.baseURL,
		values.Encode(),
	)
}

func (nm newsletterMiddleware) secret(email string) string {
	h := md5.New()

	io.WriteString(h, fmt.Sprintf("secret%s%s", nm.randomness, email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//CSFR returns csrf token for newsletter
func (nm newsletterMiddleware) CSFR(request Request) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", nm.randomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))

}

//AddEmail adds email to newsletter
func (app *App) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("Wrong email format")
	}

	err := app.Query().WhereIs("email", email).Get(&newsletterPersons{})
	if err == nil {
		return ErrEmailAlreadyInList
	}

	person := newsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return app.Create(&person)
}

//Newsletter represents newsletter
type newsletter struct {
	ID            int64     `prago-preview:"true" prago-order-desc:"true"`
	Name          string    `prago-preview:"true" prago-description:"Jméno newsletteru"`
	Body          string    `prago-type:"markdown"`
	PreviewSentAt time.Time `prago-preview:"true"`
	SentAt        time.Time `prago-preview:"true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterResource(resource *Resource) {
	resource.canView = "newsletter"

	resource.resourceController.addBeforeAction(func(request Request) {
		ret, err := resource.app.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Count(&newsletterPersons{})
		if err != nil {
			panic(err)
		}
		request.SetData("recipients_count", ret)
	})

	/*
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
		}*/
	resource.ItemAction("preview").Name(Unlocalized("Náhled")).Handler(
		func(request Request) {
			var newsletter newsletter
			err := resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			must(err)

			body, err := resource.app.newsletter.GetBody(newsletter, "")
			must(err)

			request.Response().WriteHeader(200)
			request.Response().Write([]byte(body))
			return
		},
	)

	resource.ItemAction("send-preview").Method("POST").Handler(
		func(request Request) {
			var newsletter newsletter
			must(resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))
			newsletter.PreviewSentAt = time.Now()
			must(resource.app.Save(&newsletter))

			emails := parseEmails(request.Params().Get("emails"))
			resource.app.sendEmails(newsletter, emails)
			request.AddFlashMessage("Náhled newsletteru odeslán.")
			request.Redirect(resource.getURL(""))
		},
	)

	resource.ItemAction("send").Method("POST").Template("newsletter_sent").DataSource(
		func(request Request) interface{} {
			var newsletter newsletter
			err := resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}
			newsletter.SentAt = time.Now()
			resource.app.Save(&newsletter)

			recipients, err := resource.app.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}

			go resource.app.sendEmails(newsletter, recipients)

			var ret = map[string]interface{}{}

			ret["recipients"] = recipients
			ret["recipients_count"] = len(recipients)

			return ret
		},
	)

	resource.ItemAction("duplicate").Method("POST").Handler(
		func(request Request) {
			var newsletter newsletter
			must(resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))

			var sections []*newsletterSection
			err := resource.app.Query().WhereIs("newsletter", newsletter.ID).Order("orderposition").Get(&sections)

			newsletter.ID = 0
			must(resource.app.Create(&newsletter))

			if err == nil {
				for _, v := range sections {
					section := *v
					section.ID = 0
					section.Newsletter = newsletter.ID
					must(resource.app.Create(&section))
				}
			}
			request.Redirect(resource.getItemURL(&newsletter, "edit"))
		},
	)

	resource.ItemAction("send-preview").Name(Unlocalized("Odeslat náhled")).Template("newsletter_send_preview")

	resource.ItemAction("send").Name(Unlocalized("Odeslat")).Template("newsletter_send").DataSource(
		func(Request) interface{} {
			recipients, err := resource.app.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}
			return map[string]interface{}{
				"recipients":       recipients,
				"recipients_count": len(recipients),
			}
		},
	)

	resource.ItemAction("duplicate").Name(Unlocalized("Duplikovat")).Template("newsletter_duplicate")

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

	var persons []*newsletterPersons
	err := app.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Get(&persons)
	if err != nil {
		return nil, err
	}

	for _, v := range persons {
		ret = append(ret, v.Email)
	}
	return ret, nil
}

func (app *App) sendEmails(n newsletter, emails []string) error {
	for _, v := range emails {
		body, err := app.newsletter.GetBody(n, v)
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

//GetBody gets body of newsletter
func (nm *newsletterMiddleware) GetBody(n newsletter, email string) (string, error) {
	content := markdown.New(markdown.HTML(true)).RenderToString([]byte(n.Body))
	params := map[string]interface{}{
		"id":          n.ID,
		"baseUrl":     nm.baseURL,
		"site":        nm.app.name("en"),
		"title":       n.Name,
		"unsubscribe": nm.unsubscribeURL(email),
		"content":     template.HTML(content),
		"preview":     utils.CropMarkdown(n.Body, 200),
		"sections":    nm.getNewsletterSectionData(n),
	}

	if nm.renderer != nil {
		return nm.renderer(params)
	}
	return defaultNewsletterRenderer(params)
}

//NewsletterRenderer represent newsletter renderer
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

//NewsletterSection represents section of newsletter
type newsletterSection struct {
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
	resource.canView = "newsletter"
}

//NewsletterPersons represents person of newsletter
type newsletterPersons struct {
	ID           int64
	Name         string `prago-preview:"true" prago-description:"Jméno příjemce"`
	Email        string `prago-preview:"true"`
	Confirmed    bool   `prago-preview:"true"`
	Unsubscribed bool   `prago-preview:"true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func initNewsletterPersonsResource(resource *Resource) {
	resource.canView = "sysadmin"
}

//NewsletterSectionData represents data of newsletter section
type newsletterSectionData struct {
	Name   string
	Text   string
	Button string
	URL    string
	Image  string
}

func (nm *newsletterMiddleware) getNewsletterSectionData(n newsletter) []newsletterSectionData {
	var sections []*newsletterSection
	err := nm.app.Query().WhereIs("newsletter", n.ID).Order("orderposition").Get(&sections)
	if err != nil {
		return nil
	}

	var ret []newsletterSectionData

	for _, v := range sections {
		button := "Zjistit více"
		if v.Button != "" {
			button = v.Button
		}

		url := nm.baseURL
		if v.URL != "" {
			url = v.URL
		}

		image := ""
		files := nm.app.GetFiles(v.Image)
		if len(files) > 0 {
			image = files[0].GetMedium()
		}

		ret = append(ret, newsletterSectionData{
			Name:   v.Name,
			Text:   v.Text,
			Button: button,
			URL:    url,
			Image:  image,
		})
	}
	return ret
}
