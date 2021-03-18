package prago

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime/debug"
	"time"
)

func (app App) recoveryFunction(p *Request, recoveryData interface{}) {
	duration := time.Now().Sub(p.receivedAt)

	if app.developmentMode {
		temp, err := template.New("development_error").Parse(recoveryTmpl)
		if err != nil {
			panic(err)
		}
		byteData := fmt.Sprintf("%s", recoveryData)

		buf := new(bytes.Buffer)
		err = temp.ExecuteTemplate(buf, "development_error", map[string]interface{}{
			"name":    byteData,
			"subname": fmt.Sprintf("500 Internal Server Error (errorid %s)", p.uuid),
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
		p.Response().Write([]byte(fmt.Sprintf("We are sorry, some error occured. (errorid %s)", p.uuid)))
	}

	app.Log().Printf("500 - application error\nmessage=%s\nuuid=%s\ntook=%v\n%s",
		recoveryData,
		p.uuid,
		duration,
		string(debug.Stack()),
	)
}

const recoveryTmpl = `
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
      line-height: 1.2em;
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
