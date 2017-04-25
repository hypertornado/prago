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
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"> 
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1.0"/>
<title>Email title or subject</title>

<style type="text/css">
  body {
  	background-color: #edfaff;
  	font-style: normal;
  	font-size: 15px;
  	line-height: 1.5em;
  	font-weight: 400;
  	color: #01354a;
  	font-family: Arial, sans-serif;
  }

  .middle {
  	background-color: #fff;
  	padding: 10px;
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
  }

</style>

</head>
<body>
<table border="0" cellspacing="0" width="100%">
    <tr>
        <td></td>
        <td width="450" class="middle">
        	<div class="middle_header">
        		<a href="https://www.lazne-podebrady.cz">Novinky z Lázních Poděbrady</a>
        		<h1>{{.title}}</h1>
        	</div>
        	{{.content}}

        	<a href="{{.unsubscribe}}" class="unsubscribe">Odhlásit odběr novinek</a>
        </td>
        <td></td>
     </tr>
</table> 
</body>
</html>
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
		"title":       n.Name,
		"unsubscribe": "https://www.lazne-podebrady.cz/zrusit-newsletter",
		"content":     template.HTML(content),
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
