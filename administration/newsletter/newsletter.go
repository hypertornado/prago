package newsletter

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"github.com/chris-ramon/douceur/inliner"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration"
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

var nmMiddleware *NewsletterMiddleware

type NewsletterMiddleware struct {
	Name            string
	baseUrl         string
	Admin           *administration.Administration
	SenderEmail     string
	SenderName      string
	Randomness      string
	SendgridKey     string
	Renderer        NewsletterRenderer
	Authenticatizer administration.Permission
	sendgridClient  *sendgrid.SGClient
	controller      *prago.Controller
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func InitNewsletterHelper(app *prago.App, nm NewsletterMiddleware) {
	if nmMiddleware != nil {
		app.Log().Errorf("cant initialize more then one instance of newsletter")
		return
	}
	nmMiddleware = &nm
	nmMiddleware.controller = app.MainController().SubController()
	nmMiddleware.baseUrl = app.Config.GetString("baseUrl")

	nmMiddleware.sendgridClient = sendgrid.NewSendGridClientWithApiKey(
		app.Config.GetString("sendgridApi"),
	)

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
		err := nmMiddleware.AddEmail(email, name, false)
		if err == nil {
			err := nmMiddleware.sendConfirmEmail(name, email)
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
		err := nmMiddleware.Admin.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Confirmed = true
		err = nmMiddleware.Admin.Save(&person)
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
		err := nmMiddleware.Admin.Query().WhereIs("email", email).Get(&person)
		if err != nil {
			panic("can't find user")
		}

		person.Unsubscribed = true
		err = nmMiddleware.Admin.Save(&person)
		if err != nil {
			panic(err)
		}

		request.SetData("show_back_button", true)
		request.SetData("title", "Odhlášení z odebírání newsletteru proběhlo úspěšně.")
		request.SetData("yield", "newsletter_empty")
		request.RenderView("newsletter_layout")
	})

	nmMiddleware.Admin.CreateResource(Newsletter{}, initNewsletterResource)
	nmMiddleware.Admin.CreateResource(NewsletterPersons{}, initNewsletterPersonsResource)
}

func (nm NewsletterMiddleware) sendConfirmEmail(name, email string) error {
	message := sendgrid.NewMail()

	address := mail.Address{
		Name:    nm.SenderName,
		Address: nm.SenderEmail,
	}

	message.SetFromEmail(&address)
	message.AddTo(email)
	message.AddToName(name)
	message.SetSubject("Potvrďte prosím odběr newsletteru " + nm.Name)
	message.SetText(nm.confirmEmailBody(name, email))
	return nm.sendgridClient.Send(message)
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
	io.WriteString(h, fmt.Sprintf("secret%s%s", nm.Randomness, email))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (nm NewsletterMiddleware) CSFR(request prago.Request) string {
	h := md5.New()
	io.WriteString(h, fmt.Sprintf("%s%s", nm.Randomness, request.Request().UserAgent()))
	return fmt.Sprintf("%x", h.Sum(nil))

}

func (nm NewsletterMiddleware) AddEmail(email, name string, confirm bool) error {
	if !strings.Contains(email, "@") {
		return errors.New("Wrong email format")
	}

	err := nm.Admin.Query().WhereIs("email", email).Get(&NewsletterPersons{})
	if err == nil {
		return ErrEmailAlreadyInList
	}

	person := NewsletterPersons{
		Name:      name,
		Email:     email,
		Confirmed: confirm,
	}
	return nm.Admin.Create(&person)
}

type Newsletter struct {
	ID            int64
	Name          string `prago-preview:"true" prago-description:"Jméno newsletteru"`
	Body          string `prago-type:"markdown"`
	PreviewSentAt time.Time
	SentAt        time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func initNewsletterResource(resource *administration.Resource) {
	a := resource.Admin
	resource.ActivityLog = true
	resource.CanView = nmMiddleware.Authenticatizer

	resource.ResourceController.AddBeforeAction(func(request prago.Request) {
		ret, err := a.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Count(&NewsletterPersons{})
		if err != nil {
			panic(err)
		}
		request.SetData("recipients_count", ret)
	})

	previewAction := administration.Action{
		Name: func(string) string { return "Náhled" },
		URL:  "preview",
		Handler: func(resource administration.Resource, request prago.Request, user administration.User) {
			var newsletter Newsletter
			err := resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}

			body, err := nmMiddleware.GetBody(newsletter, "")
			if err != nil {
				panic(err)
			}

			request.Response().WriteHeader(200)
			request.Response().Write([]byte(body))
		},
	}

	doSendPreviewAction := administration.Action{
		URL:    "send-preview",
		Method: "post",
		Handler: func(resource administration.Resource, request prago.Request, user administration.User) {
			var newsletter Newsletter
			must(resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter))
			newsletter.PreviewSentAt = time.Now()
			must(resource.Admin.Save(&newsletter))

			emails := parseEmails(request.Params().Get("emails"))
			nmMiddleware.SendEmails(newsletter, emails)
			administration.AddFlashMessage(request, "Náhled newsletteru odeslán.")
			request.Redirect(resource.GetURL(""))
		},
	}

	doSendAction := administration.Action{
		URL:    "send",
		Method: "post",
		Handler: func(resource administration.Resource, request prago.Request, user administration.User) {
			var newsletter Newsletter
			err := resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}
			newsletter.SentAt = time.Now()
			resource.Admin.Save(&newsletter)

			recipients, err := nmMiddleware.GetRecipients()
			if err != nil {
				panic(err)
			}

			go nmMiddleware.SendEmails(newsletter, recipients)

			request.SetData("recipients", recipients)
			request.SetData("recipients_count", len(recipients))
			request.SetData("admin_yield", "newsletter_sent")
			request.RenderView("admin_layout")
		},
	}

	resource.AddItemAction(previewAction)
	resource.AddItemAction(
		administration.CreateNavigationalItemAction(
			"send-preview",
			func(string) string { return "Odeslat náhled" },
			"newsletter_send",
			nil,
		),
	)
	resource.AddItemAction(doSendPreviewAction)
	resource.AddItemAction(administration.CreateNavigationalItemAction(
		"send",
		func(string) string { return "Odeslat" },
		"newsletter_send",
		func(administration.Resource, prago.Request, administration.User) interface{} {
			recipients, err := nmMiddleware.GetRecipients()
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

func (nm *NewsletterMiddleware) GetRecipients() ([]string, error) {
	ret := []string{}

	var persons []*NewsletterPersons
	err := nm.Admin.Query().WhereIs("confirmed", true).WhereIs("unsubscribed", false).Get(&persons)
	if err != nil {
		return nil, err
	}

	for _, v := range persons {
		ret = append(ret, v.Email)
	}
	return ret, nil
}

func (nm *NewsletterMiddleware) SendEmails(n Newsletter, emails []string) error {
	for _, v := range emails {
		body, err := nm.GetBody(n, v)
		if err == nil {
			message := sendgrid.NewMail()

			address := mail.Address{
				Name:    nm.SenderName,
				Address: nm.SenderEmail,
			}
			message.SetFromEmail(&address)

			message.AddTo(v)
			message.SetSubject(n.Name)
			message.SetHTML(body)
			err = nm.sendgridClient.Send(message)
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

func initNewsletterPersonsResource(resource *administration.Resource) {
	resource.CanView = "sysadmin"
	resource.ActivityLog = true
}