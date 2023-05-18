package prago

import (
	"bytes"
	"fmt"
	"html/template"
	"runtime/debug"
	"time"
)

func (app App) recoveryFunction(request *Request, recoveryData interface{}) {
	duration := time.Since(request.receivedAt)

	if app.developmentMode {
		temp, err := template.New("development_error").Parse(recoveryTmpl)
		if err != nil {
			panic(err)
		}
		byteData := fmt.Sprintf("%s", recoveryData)

		buf := new(bytes.Buffer)
		err = temp.ExecuteTemplate(buf, "development_error", map[string]interface{}{
			"name":    byteData,
			"subname": fmt.Sprintf("500 Internal Server Error (errorid %s)", request.uuid),
			"stack":   string(debug.Stack()),
		})
		if err != nil {
			panic(err)
		}

		if !request.Written {
			request.Response().Header().Add("Content-type", "text/html")
			request.Response().WriteHeader(500)
			request.Response().Write(buf.Bytes())
		}
	} else {
		request.Response().WriteHeader(500)
		request.Response().Write([]byte(fmt.Sprintf("We are sorry, some error occured. (errorid %s)", request.uuid)))
	}

	request.Written = true

	var userID int64 = request.UserID()

	message := fmt.Sprintf("500 - application error\nuserid=%d\nmessage=%s\nuuid=%s\ntook=%v\n%s",
		userID,
		recoveryData,
		request.uuid,
		duration,
		string(debug.Stack()),
	)
	app.Log().panicln(message)
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
