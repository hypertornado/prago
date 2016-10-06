package development

import (
	"bytes"
	"fmt"
	"github.com/hypertornado/prago"
	"html/template"
	"runtime/debug"
)

var defaultPort = 8585

type DevelopmentSettings struct {
	Less []Less
}

type Less struct {
	SourceDir string
	Target    string
}

type MiddlewareDevelopment struct {
	Settings DevelopmentSettings
}

func (m MiddlewareDevelopment) Init(app *prago.App) error {
	app.RecoveryFunc = DevelopmentRecovery
	return m.initPlatform(app)
}

func DevelopmentRecovery(p *prago.Request, recoveryData interface{}) {
	if p.App().DevelopmentMode {
		temp, err := template.New("development_error").Parse(developmentErrorTmpl)
		if err != nil {
			panic(err)
		}

		byteData := fmt.Sprintf("%s", recoveryData)

		buf := new(bytes.Buffer)
		err = temp.ExecuteTemplate(buf, "development_error", map[string]interface{}{
			"name":    byteData,
			"subname": "500 Internal Server Error",
			"stack":   string(debug.Stack()),
		})
		if err != nil {
			panic(err)
		}

		p.Response().Header().Add("Content-type", "text/html")
		p.Response().WriteHeader(500)
		p.Response().Write(buf.Bytes())

	} else {
		p.Response().WriteHeader(500)
		p.Response().Write([]byte("We are sorry, some error occured. (500)"))
	}

	p.Log().Errorln(fmt.Sprintf("500 - error\n%s\nstack:\n", recoveryData))
	p.Log().Errorln(string(debug.Stack()))
}

const developmentErrorTmpl = `
<html>
<head>
  <title>{{.subname}}: {{.name}}</title>

  <style>
  	html, body{
  	  height: 100%;
  	  font-family: Roboto, -apple-system, BlinkMacSystemFont, "Helvetica Neue", "Segoe UI", Oxygen, Ubuntu, Cantarell, "Open Sans", sans-serif;
  	  font-size: 15px;
  	  line-height: 1.4em;
  	  margin: 0px;
  	  color: #333;
  	}
  	h1 {
  		border-bottom: 1px solid #dd2e4f;
  		background-color: #dd2e4f;
  		color: white;
  		padding: 10px 10px;
  		margin: 0px;
  	}

  	.err {
  		font-size: 15px;
  		margin-bottom: 5px;
  	}

  	pre {
  		margin: 5px 10px;
  	}

  </style>

</head>
<body>

<h1>
	<div class="err">{{.subname}}</div>
	{{.name}}
</h1>

<pre>{{.stack}}</pre>

</body>
</html>
`
