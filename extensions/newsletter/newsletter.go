package newsletter

import (
	"bytes"
	"github.com/chris-ramon/douceur/inliner"
	"github.com/golang-commonmark/markdown"
	"github.com/hypertornado/prago"
	administration "github.com/hypertornado/prago/extensions/admin"
	"html/template"
	"time"
)

//https://github.com/chris-ramon/douceur
//https://github.com/aymerick/douceur

const newsletterTemplate = `

<style>
p {
	color: red;
}
</style>

<h1>HEADER</h1>

{{.content}}

`

type NewsletterMiddleware struct {
	Admin       *administration.Admin
	SenderEmail string
}

func (nm NewsletterMiddleware) Init(app *prago.App) error {
	_, err := nm.Admin.CreateResource(Newsletter{})
	if err != nil {
		return err
	}

	_, err = nm.Admin.CreateResource(NewsletterPersons{})
	if err != nil {
		return err
	}

	return nil
}

type Newsletter struct {
	ID            int64
	Name          string `prago-preview:"true" prago-description:"Jméno newsletteru"`
	Subject       string
	Body          string `prago-type:"markdown"`
	PreviewSentAt time.Time
	SentAt        time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Newsletter) InitResource(a *administration.Admin, resource *administration.Resource) error {
	previewAction := administration.ResourceAction{
		Name: func(string) string { return "Preview" },
		Url:  "preview",
		Handler: func(admin *administration.Admin, resource *administration.Resource, request prago.Request) {
			var newsletter Newsletter
			err := admin.Query().WhereIs("id", request.Params().Get("id")).Get(&newsletter)
			if err != nil {
				panic(err)
			}

			body, err := newsletter.GetBody()
			if err != nil {
				panic(err)
			}

			request.SetData("body", []byte(body))
			request.SetData("statusCode", 200)
		},
	}

	resource.AddResourceItemAction(previewAction)
	return nil
}

func (n Newsletter) GetBody() (string, error) {
	t, err := template.New("newsletter").Parse(newsletterTemplate)
	if err != nil {
		return "", err
	}

	content := markdown.New().RenderToString([]byte(n.Body))

	buf := new(bytes.Buffer)
	err = t.ExecuteTemplate(buf, "newsletter", map[string]interface{}{
		"content": template.HTML(content),
	})
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
	Email        string
	Confirmed    bool
	Unsubscribed bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
