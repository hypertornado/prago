package prago

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/chris-ramon/douceur/inliner"
	"github.com/golang-commonmark/markdown"
)

//ErrEmailAlreadyInList is returned when user is already in newsletter list
var ErrEmailAlreadyInList = errors.New("email already in newsletter list")

//NewsletterMiddleware represents users newsletter
type Newsletters struct {
	baseURL    string
	renderer   NewsletterRenderer
	app        *App
	randomness string

	newsletterResource        *Resource
	newsletterSectionResource *Resource
}

func (newsletters *Newsletters) Renderer(renderer NewsletterRenderer) *Newsletters {
	newsletters.renderer = renderer
	return newsletters
}

func (newsletters *Newsletters) Permission(permission Permission) *Newsletters {
	newsletters.newsletterResource.PermissionView(permission)
	newsletters.newsletterSectionResource.PermissionView(permission)

	return newsletters
}

//InitNewsletters inits apps newsletter function
func (app *App) Newsletters() *Newsletters {
	if app.newsletters != nil {
		panic("newsletter already initialized")
	}
	app.newsletters = &Newsletters{
		baseURL:  app.ConfigurationGetString("baseUrl"),
		renderer: defaultNewsletterRenderer,

		app:        app,
		randomness: app.ConfigurationGetString("random"),
	}

	controller := app.mainController.subController()
	controller.addBeforeAction(func(request *Request) {
		request.SetData("site", app.name("en"))
	})

	controller.get("/newsletter-subscribe", func(request *Request) {
		request.SetData("title", "Přihlásit se k odběru newsletteru")
		request.SetData("csrf", app.newsletters.CSRF(request))
		request.SetData("yield", "newsletter_subscribe")
		request.SetData("show_back_button", true)
		request.RenderView("newsletter_layout")
	})

	controller.post("/newsletter-subscribe", func(request *Request) {
		if app.newsletters.CSRF(request) != request.Params().Get("csrf") {
			panic("wrong csrf")
		}

		email := request.Params().Get("email")
		email = strings.Trim(email, " ")
		name := request.Params().Get("name")

		var message string
		err := app.newsletters.AddEmail(email, name, false)
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

	controller.get("/newsletter-confirm", func(request *Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.newsletters.secret(email) != secret {
			panic("wrong secret")
		}

		var person newsletterPersons
		err := app.Is("email", email).Get(&person)
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

	controller.get("/newsletter-unsubscribe", func(request *Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if app.newsletters.secret(email) != secret {
			panic("wrong secret")
		}

		var person newsletterPersons
		err := app.Is("email", email).Get(&person)
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

	app.newsletters.newsletterResource = app.Resource(newsletter{}).Name(unlocalized("Newsletter"))
	initNewsletterResource(
		app.newsletters.newsletterResource,
	)

	app.newsletters.newsletterSectionResource = app.Resource(newsletterSection{}).Name(unlocalized("Newsletter - sekce"))
	initNewsletterSection(
		app.newsletters.newsletterSectionResource,
	)

	app.Resource(newsletterPersons{}).PermissionView(sysadminPermission).Name(unlocalized("Newsletter - osoby"))
	return app.newsletters
}

func (app App) sendConfirmEmail(name, email string) error {
	text := app.newsletters.confirmEmailBody(name, email)
	return app.Email().To(name, email).Subject("Potvrďte prosím odběr newsletteru " + app.name("en")).TextContent(text).Send()
}

func (nm Newsletters) confirmEmailBody(name, email string) string {
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

func (nm Newsletters) unsubscribeURL(email string) string {
	values := make(url.Values)
	values.Set("email", email)
	values.Set("secret", nm.secret(email))

	return fmt.Sprintf("%s/newsletter-unsubscribe?%s",
		nm.baseURL,
		values.Encode(),
	)
}

func (nm Newsletters) secret(email string) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("secret%s%s", nm.randomness, email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

//CSRF returns csrf token for newsletter
func (nm *Newsletters) CSRF(request *Request) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", nm.randomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (nm *Newsletters) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("wrong email format")
	}

	err := nm.app.Is("email", email).Get(&newsletterPersons{})
	if err == nil {
		return ErrEmailAlreadyInList
	}

	person := newsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return nm.app.Create(&person)
}

//Newsletter represents newsletter
type newsletter struct {
	ID            int64     `prago-preview:"true" prago-order-desc:"true"`
	Name          string    `prago-preview:"true" prago-name:"Jméno newsletteru"`
	Body          string    `prago-type:"markdown"`
	PreviewSentAt time.Time `prago-preview:"true"`
	SentAt        time.Time `prago-preview:"true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterResource(resource *Resource) {
	resource.canView = sysadminPermission

	resource.resourceController.addBeforeAction(func(request *Request) {
		ret, err := resource.app.Is("confirmed", true).Is("unsubscribed", false).Count(&newsletterPersons{})
		if err != nil {
			panic(err)
		}
		request.SetData("recipients_count", ret)
	})

	resource.ItemAction("preview").Permission(loggedPermission).Name(unlocalized("Náhled")).Handler(
		func(request *Request) {
			var newsletter newsletter
			resource.app.Is("id", request.Params().Get("id")).MustGet(&newsletter)

			body, err := resource.app.newsletters.GetBody(newsletter, "")
			must(err)

			request.Response().WriteHeader(200)
			request.Response().Write([]byte(body))
		},
	)

	resource.ItemAction("send-preview").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			var newsletter newsletter
			resource.app.Is("id", request.Params().Get("id")).MustGet(&newsletter)
			newsletter.PreviewSentAt = time.Now()
			resource.app.MustSave(&newsletter)

			emails := parseEmails(request.Params().Get("emails"))
			resource.app.sendEmails(newsletter, emails)
			request.AddFlashMessage("Náhled newsletteru odeslán.")
			request.Redirect(resource.getURL(""))
		},
	)

	resource.ItemAction("send").Permission(loggedPermission).Method("POST").Template("newsletter_sent").DataSource(
		func(request *Request) interface{} {
			var newsletter newsletter
			resource.app.Is("id", request.Params().Get("id")).MustGet(&newsletter)

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

	resource.ItemAction("duplicate").Permission(loggedPermission).Method("POST").Handler(
		func(request *Request) {
			var newsletter newsletter
			resource.app.Is("id", request.Params().Get("id")).MustGet(&newsletter)

			var sections []*newsletterSection
			err := resource.app.Is("newsletter", newsletter.ID).Order("orderposition").Get(&sections)

			newsletter.ID = 0
			resource.app.MustCreate(&newsletter)

			if err == nil {
				for _, v := range sections {
					section := *v
					section.ID = 0
					section.Newsletter = newsletter.ID
					resource.app.MustCreate(&section)
				}
			}
			request.Redirect(resource.getItemURL(&newsletter, "edit"))
		},
	)

	resource.ItemAction("send-preview").Permission(loggedPermission).Name(unlocalized("Odeslat náhled")).Template("newsletter_send_preview")

	resource.ItemAction("send").Permission(loggedPermission).Name(unlocalized("Odeslat")).Template("newsletter_send").DataSource(
		func(*Request) interface{} {
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

	resource.ItemAction("duplicate").Permission(loggedPermission).Name(unlocalized("Duplikovat")).Template("newsletter_duplicate")

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
	err := app.Is("confirmed", true).Is("unsubscribed", false).Get(&persons)
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
		body, err := app.newsletters.GetBody(n, v)
		if err == nil {
			err = app.Email().To("", v).Subject(n.Name).HTMLContent(body).Send()
		}
		if err != nil {
			app.Log().Println("ERROR", err.Error())
		}
	}
	return nil
}

//GetBody gets body of newsletter
func (nm *Newsletters) GetBody(n newsletter, email string) (string, error) {
	content := markdown.New(markdown.HTML(true)).RenderToString([]byte(n.Body))
	params := map[string]interface{}{
		"id":          n.ID,
		"baseUrl":     nm.baseURL,
		"site":        nm.app.name("en"),
		"title":       n.Name,
		"unsubscribe": nm.unsubscribeURL(email),
		"content":     template.HTML(content),
		"preview":     cropMarkdown(n.Body, 200),
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

	ret, err := inliner.Inline(buf.String())
	if err != nil {
		return "", err
	}

	return ret, nil
}

//NewsletterSection represents section of newsletter
type newsletterSection struct {
	ID            int64
	Newsletter    int64  `prago-type:"relation" prago-preview:"true"`
	Name          string `prago-name:"Jméno sekce"`
	Text          string `prago-type:"text"`
	Button        string `prago-name:"Tlačítko"`
	URL           string `prago-name:"Odkaz"`
	Image         string `prago-type:"image" prago-preview:"true"`
	OrderPosition int64  `prago-type:"order"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterSection(resource *Resource) {
	resource.canView = sysadminPermission
}

//NewsletterPersons represents person of newsletter
type newsletterPersons struct {
	ID           int64
	Name         string `prago-preview:"true" prago-name:"Jméno příjemce"`
	Email        string `prago-preview:"true"`
	Confirmed    bool   `prago-preview:"true"`
	Unsubscribed bool   `prago-preview:"true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

//NewsletterSectionData represents data of newsletter section
type newsletterSectionData struct {
	Name   string
	Text   string
	Button string
	URL    string
	Image  string
}

func (nm *Newsletters) getNewsletterSectionData(n newsletter) []newsletterSectionData {
	var sections []*newsletterSection
	err := nm.app.Is("newsletter", n.ID).Order("orderposition").Get(&sections)
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
