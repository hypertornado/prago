package prago

import (
	"bytes"
	"context"
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

// ErrEmailAlreadyInList is returned when user is already in newsletter list
var ErrEmailAlreadyInList = errors.New("email already in newsletter list")

// NewsletterMiddleware represents users newsletter
type Newsletters struct {
	renderer   NewsletterRenderer
	app        *App
	randomness string
	//board      *Board

	newsletterResource        *Resource[newsletter]
	newsletterSectionResource *Resource[newsletterSection]
}

func (newsletters *Newsletters) Renderer(renderer NewsletterRenderer) *Newsletters {
	newsletters.renderer = renderer
	return newsletters
}

/*func (newsletters *Newsletters) Board(board *Board) *Newsletters {
	newsletters.board = board
	return newsletters
}*/

func (newsletters *Newsletters) Permission(permission Permission) *Newsletters {
	newsletters.newsletterResource.PermissionView(permission)
	newsletters.newsletterSectionResource.PermissionView(permission)
	return newsletters
}

type NewsletterWriteData struct {
	Title          string
	Csrf           string
	Yield          string
	ShowBackButton bool
	Site           string
}

// InitNewsletters inits apps newsletter function
func (app *App) Newsletters(board *Board) *Newsletters {
	if app.newsletters != nil {
		return app.newsletters
	}
	app.newsletters = &Newsletters{
		renderer:   defaultNewsletterRenderer,
		app:        app,
		randomness: app.MustGetSetting(context.Background(), "random"),
		//board:      app.MainBoard,
	}

	app.GET("/newsletter-subscribe", func(request *Request) {

		data := &NewsletterWriteData{
			Title:          "Přihlásit se k odběru newsletteru",
			Csrf:           RequestCSRF(request),
			Yield:          "newsletter_subscribe",
			ShowBackButton: true,
			Site:           app.name("en"),
		}

		request.Write(200, "newsletter_layout", data)
	})

	app.POST("/newsletter-subscribe", func(request *Request) {
		if RequestCSRF(request) != request.Param("csrf") {
			panic("wrong csrf")
		}

		email := request.Param("email")
		email = strings.Trim(email, " ")
		name := request.Param("name")

		var message string
		err := app.newsletters.SubscribeWithConfirmationEmail(email, name)
		if err == nil {
			message = "Na váš email " + email + " jsme odeslali potvrzovací email k odebírání newsletteru."
		} else {
			if err == ErrEmailAlreadyInList {
				message = "Email se již nachází v naší emailové databázi"
			} else {
				panic(err)
			}
		}

		data := &NewsletterWriteData{
			ShowBackButton: true,
			Title:          message,
			Yield:          "newsletter_empty",
			Site:           app.name("en"),
		}

		request.Write(200, "newsletter_layout", data)
	})

	app.GET("/newsletter-confirm", func(request *Request) {
		email := request.Param("email")
		secret := request.Param("secret")

		if app.newsletters.secret(email) != secret {
			panic("wrong secret")
		}

		res := GetResource[newsletterPersons](app)

		person := res.Query(request.r.Context()).Is("email", email).First()
		if person == nil {
			panic("can't find user")
		}

		person.Confirmed = true

		err := res.Update(request.r.Context(), person)
		must(err)

		data := &NewsletterWriteData{
			ShowBackButton: true,
			Title:          "Odběr newsletteru potvrzen",
			Yield:          "newsletter_empty",
			Site:           app.name("en"),
		}

		request.Write(200, "newsletter_layout", data)
	})

	//TODO: add confirmation button and form
	app.GET("/newsletter-unsubscribe", func(request *Request) {
		email := request.Param("email")
		secret := request.Param("secret")

		if app.newsletters.secret(email) != secret {
			panic("wrong secret")
		}

		res := GetResource[newsletterPersons](app)

		person := res.Query(request.r.Context()).Is("email", email).First()
		if person == nil {
			panic("can't find user")
		}

		person.Unsubscribed = true
		err := res.Update(request.r.Context(), person)
		if err != nil {
			panic(err)
		}

		data := &NewsletterWriteData{
			ShowBackButton: true,
			Title:          "Odhlášení z odebírání newsletteru proběhlo úspěšně.",
			Yield:          "newsletter_empty",
			Site:           app.name("en"),
		}

		request.Write(200, "newsletter_layout", data)
	})

	app.newsletters.newsletterResource = NewResource[newsletter](app).Name(unlocalized("Newsletter"), unlocalized("Newslettery"))
	initNewsletterResource(
		GetResource[newsletter](app),
		board,
	)

	app.newsletters.newsletterSectionResource = NewResource[newsletterSection](app).
		Board(board).
		Name(unlocalized("Newsletter - sekce"), unlocalized("Newsletter - sekce"))

	NewResource[newsletterPersons](app).Board(board).PermissionView(sysadminPermission).Name(unlocalized("Newsletter - osoba"), unlocalized("Newsletter - osoby"))
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
		nm.app.MustGetSetting(context.TODO(), "base_url"),
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
		nm.app.MustGetSetting(context.TODO(), "base_url"),
		values.Encode(),
	)
}

func (nm Newsletters) secret(email string) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("secret%s%s", nm.randomness, email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (nm *Newsletters) SubscribeWithConfirmationEmail(email, name string) error {
	res := GetResource[newsletterPersons](nm.app)
	person := res.Query(context.TODO()).Is("email", email).First()
	if person != nil {
		return ErrEmailAlreadyInList
	}
	err := nm.app.sendConfirmEmail(name, email)
	if err != nil {
		return err
	}

	return nm.AddEmail(email, name, false)
}

func (nm *Newsletters) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("wrong email format")
	}

	res := GetResource[newsletterPersons](nm.app)

	person := res.Query(context.TODO()).Is("email", email).First()
	if person != nil {
		return ErrEmailAlreadyInList
	}

	person = &newsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return res.Create(context.TODO(), person)
}

// Newsletter represents newsletter
type newsletter struct {
	ID            int64     `prago-preview:"true" prago-order-desc:"true"`
	Name          string    `prago-preview:"true" prago-name:"Jméno newsletteru" prago-validations:"nonempty"`
	Body          string    `prago-type:"markdown" prago-validations:"nonempty"`
	PreviewSentAt time.Time `prago-preview:"true"`
	SentAt        time.Time `prago-preview:"true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterResource(resource *Resource[newsletter], board *Board) {
	resource.data.canView = sysadminPermission

	resource.Board(board)

	resource.ItemAction("preview").Permission(loggedPermission).Name(unlocalized("Náhled")).Handler(
		func(item *newsletter, request *Request) {
			body, err := resource.data.app.newsletters.GetBody(*item, "")
			must(err)

			request.Response().WriteHeader(200)
			request.Response().Write([]byte(body))
		},
	)

	resource.FormItemAction("send-preview").Permission(loggedPermission).Name(unlocalized("Odeslat náhled")).Form(
		func(item *newsletter, f *Form, r *Request) {
			f.AddTextareaInput("emails", "Seznam emailů na poslání preview (jeden email na řádek)").Focused = true
			f.AddSubmit("Odeslat náhled")
		},
	).Validation(func(newsletter *newsletter, vc ValidationContext) {
		newsletter.PreviewSentAt = time.Now()
		err := resource.Update(vc.Context(), newsletter)
		if err != nil {
			panic(err)
		}

		emails := parseEmails(vc.GetValue("emails"))
		if len(emails) == 0 {
			vc.AddError("Není zadán žádný email")
		}
		if vc.Valid() {
			err := resource.data.app.sendEmails(*newsletter, emails)
			if err != nil {
				vc.AddError(fmt.Sprintf("Chyba při odesílání emailů: %s", err))
			}
		}
		if vc.Valid() {
			vc.Request().AddFlashMessage("Náhled newsletteru odeslán.")
			vc.Validation().RedirectionLocaliton = resource.data.getItemURL(newsletter, "", vc.Request())
		}
	})

	resource.FormItemAction("send").Permission(loggedPermission).Name(unlocalized("Odeslat")).Form(
		func(newsletter *newsletter, form *Form, request *Request) {
			recipients, err := resource.data.app.getNewsletterRecipients()
			must(err)
			form.AddSubmit(fmt.Sprintf("Odelsat newsletter na %d emailů", len(recipients)))
		},
	).Validation(
		func(newsletter *newsletter, vc ValidationContext) {
			newsletter.SentAt = time.Now()
			//TODO: log sent emails
			must(resource.Update(vc.Context(), newsletter))

			recipients, err := resource.data.app.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}

			go resource.data.app.sendEmails(*newsletter, recipients)
			vc.Request().AddFlashMessage(fmt.Sprintf("Newsletter '%s' se odesílá na %d adres", newsletter.Name, len(recipients)))
			vc.Validation().RedirectionLocaliton = resource.data.getItemURL(newsletter, "", vc.Request())
		},
	)

	resource.FormItemAction("duplicate").Permission(loggedPermission).Name(unlocalized("Duplikovat")).Form(
		func(newsletter *newsletter, f *Form, r *Request) {
			f.AddSubmit("Duplikovat newsletter")
		},
	).Validation(func(newsletter *newsletter, vc ValidationContext) {
		newsletterSectionResource := GetResource[newsletterSection](vc.Request().app)
		sections := newsletterSectionResource.Query(vc.Context()).Is("newsletter", newsletter.ID).Order("orderposition").List()

		newsletter.ID = 0
		must(resource.CreateWithLog(newsletter, vc.Request()))

		for _, v := range sections {
			section := *v
			section.ID = 0
			section.Newsletter = newsletter.ID
			must(newsletterSectionResource.Create(vc.Context(), &section))
		}

		vc.Validation().RedirectionLocaliton = resource.data.getItemURL(newsletter, "edit", vc.Request())
	})
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
	persons := GetResource[newsletterPersons](app).Query(context.TODO()).Is("confirmed", true).Is("unsubscribed", false).List()
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

// GetBody gets body of newsletter
func (nm *Newsletters) GetBody(n newsletter, email string) (string, error) {
	content := markdown.New(markdown.HTML(true)).RenderToString([]byte(n.Body))
	params := map[string]interface{}{
		"id":          n.ID,
		"baseUrl":     nm.app.MustGetSetting(context.TODO(), "base_url"),
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

// NewsletterRenderer represent newsletter renderer
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

// NewsletterSection represents section of newsletter
type newsletterSection struct {
	ID            int64
	Newsletter    int64  `prago-type:"relation" prago-preview:"true" prago-validations:"nonempty"`
	Name          string `prago-name:"Jméno sekce" prago-validations:"nonempty"`
	Text          string `prago-type:"text" prago-validations:"nonempty"`
	Button        string `prago-name:"Tlačítko"`
	URL           string `prago-name:"Odkaz"`
	Image         string `prago-type:"image" prago-preview:"true"`
	OrderPosition int64  `prago-type:"order"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

//func initNewsletterSection(resource *resource) {
//resource.canView = sysadminPermission
//}

// NewsletterPersons represents person of newsletter
type newsletterPersons struct {
	ID           int64
	Name         string `prago-preview:"true" prago-name:"Jméno příjemce"`
	Email        string `prago-preview:"true"`
	Confirmed    bool   `prago-preview:"true"`
	Unsubscribed bool   `prago-preview:"true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewsletterSectionData represents data of newsletter section
type newsletterSectionData struct {
	Name   string
	Text   string
	Button string
	URL    string
	Image  string
}

func (nm *Newsletters) getNewsletterSectionData(n newsletter) []newsletterSectionData {
	//var sections []*newsletterSection
	sections := GetResource[newsletterSection](nm.app).Query(context.TODO()).Is("newsletter", n.ID).Order("orderposition").List()
	var ret []newsletterSectionData

	for _, v := range sections {
		button := "Zjistit více"
		if v.Button != "" {
			button = v.Button
		}

		url := nm.app.MustGetSetting(context.TODO(), "base_url")
		if v.URL != "" {
			url = v.URL
		}

		image := ""
		files := nm.app.GetFiles(context.TODO(), v.Image)
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
