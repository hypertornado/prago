package administration

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/chris-ramon/douceur/inliner"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/utils"
	"github.com/sendgrid/sendgrid-go"
	"html/template"
	"io"
	"net/mail"
	"net/url"
	"strings"
	"time"
)

var (
	ErrEmailAlreadyInList = errors.New("email already in newsletter list")
)

const defaultNewsletterTemplate = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"> 
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>{{.title}}</title>

<style type="text/css">
  body {
    background-color: #edfaff;
    font-style: normal;
    font-size: 15px;
    line-height: 1.5em;
    font-weight: 400;
    color: #01354a;
    font-family: Arial, sans-serif !important;
  }

  .middle {
    background-color: #fff;
    padding: 10px;
  }

  img {
    max-width: 100%;
  }

  a {
    color: #009ee0;
  }

  a:hover {
    text-decoration: none;
  }

  .unsubscribe {
    color: #999;
    display: block;
    text-align: center;
    font-size: 11px;
  }

  h1 {
    text-align: center;
    line-height: 1.2em;
  }

  hr {
    border-top: 1px solid #009ee0;
    border-bottom: none;
  }

  table {
    margin-top: 5px;
  }

  td {
    padding: 0px 5px;
    vertical-align: top;
  }

</style>

</head>
<body>

<table width="100%" border="0" cellspacing="0" cellpadding="0"><tr><td width="100%" align="center">
  <table width="450" border="0" align="center" cellpadding="0" cellspacing="0">
    <tr><td width="450" align="left" class="middle">
          <div class="middle_header">
            <a href="{{.baseUrl}}/?utm_source=newsletter&utm_medium=prago&utm_campaign={{.id}}">{{.site}}</a>
            <h1>{{.title}}</h1>
          </div>
          {{.content}}

          <a href="{{.unsubscribe}}" class="unsubscribe">Odhlásit odběr novinek</a>
    </td></tr>
  </table>
</td></tr></table>

</body>
</html>
`

type NewsletterMiddleware struct {
	Name        string
	baseUrl     string
	SenderEmail string
	SenderName  string
	//Randomness  string
	//SendgridKey     string
	Renderer        NewsletterRenderer
	Authenticatizer Permission
	controller      *prago.Controller

	admin      *Administration
	randomness string
}

func (admin *Administration) InitNewsletterHelper(nm NewsletterMiddleware) {
	nm.admin = admin
	nm.randomness = admin.App.Config.GetString("random")

	admin.Newsletter = &nm

	app := admin.App

	nmMiddleware := &nm
	nmMiddleware.controller = app.MainController().SubController()
	nmMiddleware.baseUrl = app.Config.GetString("baseUrl")

	nmMiddleware.controller.AddBeforeAction(func(request prago.Request) {
		request.SetData("site", nmMiddleware.Name)
	})

	nmMiddleware.controller.Get("/newsletter-subscribe", func(request prago.Request) {
		request.SetData("title", "Přihlásit se k odběru newsletteru")
		request.SetData("csrf", nmMiddleware.CSFR(request))
		request.SetData("yield", "newsletter_subscribe")
		request.SetData("show_back_button", true)
		request.RenderView("newsletter_layout")
	})

	nmMiddleware.controller.Post("/newsletter-subscribe", func(request prago.Request) {
		if nmMiddleware.CSFR(request) != request.Params().Get("csrf") {
			panic("wrong csrf")
		}

		email := request.Params().Get("email")
		email = strings.Trim(email, " ")
		name := request.Params().Get("name")

		var message string
		err := admin.AddEmail(email, name, false)
		if err == nil {
			err := admin.sendConfirmEmail(name, email)
			if err != nil {
				panic(err)
			}
			message = "Na váš email " + email + " potvrzovací email k odebírání newsletteru."
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

	nmMiddleware.controller.Get("/newsletter-confirm", func(request prago.Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if nmMiddleware.secret(email) != secret {
			panic("wrong secret")
		}

		var person NewsletterPersons
		err := admin.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Confirmed = true
		err = admin.Save(&person)
		if err != nil {
			panic(err)
		}

		request.SetData("show_back_button", true)
		request.SetData("title", "Odběr newsletteru potvrzen")
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	nmMiddleware.controller.Get("/newsletter-unsubscribe", func(request prago.Request) {
		email := request.Params().Get("email")
		secret := request.Params().Get("secret")

		if nmMiddleware.secret(email) != secret {
			panic("wrong secret")
		}

		var person NewsletterPersons
		err := admin.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Unsubscribed = true
		err = admin.Save(&person)
		if err != nil {
			panic(err)
		}

		request.SetData("show_back_button", true)
		request.SetData("title", "Odhlášení z odebírání newsletteru proběhlo úspěšně.")
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	admin.CreateResource(Newsletter{}, initNewsletterResource)
	admin.CreateResource(NewsletterPersons{}, initNewsletterPersonsResource)
}

func (admin Administration) sendConfirmEmail(name, email string) error {
	message := sendgrid.NewMail()

	address := mail.Address{
		Name:    admin.Newsletter.SenderName,
		Address: admin.Newsletter.SenderEmail,
	}

	message.SetFromEmail(&address)
	message.AddTo(email)
	message.AddToName(name)
	message.SetSubject("Potvrďte prosím odběr newsletteru " + admin.Newsletter.Name)
	message.SetText(admin.Newsletter.confirmEmailBody(name, email))
	return admin.sendgridClient.Send(message)
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
		nm.Name,
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

func (nm NewsletterMiddleware) CSFR(request prago.Request) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", nm.randomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (admin *Administration) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("Wrong email format")
	}

	err := admin.Query().WhereIs("email", email).Get(&NewsletterPersons{})
	if err == nil {
		return ErrEmailAlreadyInList
	}

	person := NewsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return admin.Create(&person)
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
	resource.CanView = resource.Admin.Newsletter.Authenticatizer

	resource.ResourceController.AddBeforeAction(func(request prago.Request) {
		ret, err := resource.Admin.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Count(&NewsletterPersons{})
		if err != nil {
			panic(err)
		}
		request.SetData("recipients_count", ret)
	})

	previewAction := Action{
		Name: func(string) string { return "Náhled" },
		URL:  "preview",
		Handler: func(resource Resource, request prago.Request, user User) {
			var newsletter Newsletter
			err := resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}

			body, err := resource.Admin.Newsletter.GetBody(newsletter, "")
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
		Handler: func(resource Resource, request prago.Request, user User) {
			var newsletter Newsletter
			must(resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))
			newsletter.PreviewSentAt = time.Now()
			must(resource.Admin.Save(&newsletter))

			emails := parseEmails(request.Params().Get("emails"))
			resource.Admin.sendEmails(newsletter, emails)
			AddFlashMessage(request, "Náhled newsletteru odeslán.")
			request.Redirect(resource.GetURL(""))
		},
	}

	doSendAction := Action{
		URL:    "send",
		Method: "post",
		Handler: func(resource Resource, request prago.Request, user User) {
			var newsletter Newsletter
			err := resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}
			newsletter.SentAt = time.Now()
			resource.Admin.Save(&newsletter)

			recipients, err := resource.Admin.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}

			go resource.Admin.sendEmails(newsletter, recipients)

			request.SetData("recipients", recipients)
			request.SetData("recipients_count", len(recipients))
			request.SetData("admin_yield", "newsletter_sent")
			request.RenderView("admin_layout")
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
		func(Resource, prago.Request, User) interface{} {
			recipients, err := resource.Admin.getNewsletterRecipients()
			if err != nil {
				panic(err)
			}
			return map[string]interface{}{
				"recipients":       recipients,
				"recipients_count": len(recipients),
			}
		},
	))
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

func (admin *Administration) getNewsletterRecipients() ([]string, error) {
	ret := []string{}

	var persons []*NewsletterPersons
	err := admin.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Get(&persons)
	if err != nil {
		return nil, err
	}

	for _, v := range persons {
		ret = append(ret, v.Email)
	}
	return ret, nil
}

func (admin *Administration) sendEmails(n Newsletter, emails []string) error {
	for _, v := range emails {
		body, err := admin.Newsletter.GetBody(n, v)
		if err == nil {
			message := sendgrid.NewMail()

			address := mail.Address{
				Name:    admin.Newsletter.SenderName,
				Address: admin.Newsletter.SenderEmail,
			}
			message.SetFromEmail(&address)

			message.AddTo(v)
			message.SetSubject(n.Name)
			message.SetHTML(body)
			err = admin.sendgridClient.Send(message)
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
		"site":        nm.Name,
		"title":       n.Name,
		"unsubscribe": nm.unsubscribeUrl(email),
		"content":     template.HTML(content),
		"preview":     utils.CropMarkdown(n.Body, 200),
	}

	if nm.Renderer != nil {
		return nm.Renderer(params)
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
