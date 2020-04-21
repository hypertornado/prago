package setup

import (
	"fmt"
	"os"
	"path"

	"github.com/hypertornado/prago/utils"
)

func createSkeleton(workingDirectory, projectName string) {
	if !utils.ConsoleQuestion("Do you want to construct app skeleton?") {
		return
	}

	createTemplates(workingDirectory, projectName)
	createGoFiles(workingDirectory, projectName)

	createDirectory(path.Join(workingDirectory, "resources"))
	createDirectory(path.Join(workingDirectory, "resources", "js"))
	createDirectory(path.Join(workingDirectory, "resources", "css"))

	createCSSFiles(workingDirectory, projectName)
	createJSFiles(workingDirectory, projectName)

}

func createTemplates(workingDirectory, projectName string) {
	createDirectory(path.Join(workingDirectory, "templates"))
	createFile(path.Join(workingDirectory, "templates", "_layout.tmpl"), `
{{define "layout"}}
	<!doctype html>
	<html>
		<head>
			<meta charset="utf-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
			<title>{{.app_name}}</title>
			<meta name="description" content="">
			<meta name="viewport" content="width=device-width, initial-scale=1">

			<link rel="stylesheet" href="/style.css?={{.version}}">
			<script type="text/javascript" src="/script.js?v={{.version}}"></script>

		</head>
		<body>
			<h1 class="hero">{{.app_name}}</h1>
			{{tmpl .yield .}}

			<div>&copy; - {{.app_name}}</div>
		</body>
	</html>
{{end}}
`)

	createFile(path.Join(workingDirectory, "templates", "index.tmpl"), `
{{define "index"}}
	<h2>Index page of project</h2>
{{end}}
`)

	createFile(path.Join(workingDirectory, "templates", "404.tmpl"), `
{{define "404"}}
	<h2>Page not found</h2>
{{end}}
`)
}

func createGoFiles(workingDirectory, projectName string) {
	createFile(path.Join(workingDirectory, "main.go"), `
package main

import (
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration"
	"github.com/hypertornado/prago/build"
	"github.com/hypertornado/prago/development"
)

var appName = "`+projectName+`"
var appVersion = "0.0.1"

func main() {
	prago.NewApp(appName, appVersion, func(app *prago.App) {
		administration.NewAdministration(app, func(admin *administration.Administration) {

		})

		app.LoadTemplatePath("templates/*")

		build.CreateBuildHelper(app, build.BuildSettings{
			Copy: [][2]string{{"public", ""}, {"templates", ""}},
		})

		development.CreateDevelopmentHelper(app, development.DevelopmentSettings{
			Less: []development.Less{
				{"resources/css", "public/style.css"},
			},
			TypeScript: []string{
				"resources/js",
			},
		})

		app.MainController().Get("/", func(request prago.Request) {
			request.SetData("app_name", appName)
			request.SetData("version", appVersion)
			request.SetData("yield", "index")
			request.RenderView("layout")
		})

		app.MainController().Get("*", func(request prago.Request) {
			request.SetData("app_name", appName)
			request.SetData("version", appVersion)
			request.SetData("yield", "404")
			request.RenderViewWithCode("layout", 404)
		})
	})
}		
`)
}

func createCSSFiles(workingDirectory, projectName string) {
	createDirectory(path.Join(workingDirectory, "resources", "css", "layout"))
	createDirectory(path.Join(workingDirectory, "resources", "css", "base"))
	createDirectory(path.Join(workingDirectory, "resources", "css", "components"))
	createDirectory(path.Join(workingDirectory, "resources", "css", "core"))

	createFile(path.Join(workingDirectory, "resources", "css", "index.less"), `
@import "layout/normalize";

@import "core/variables";

@import "components/hero";

	`)

	createFile(path.Join(workingDirectory, "resources", "css", "layout", "normalize.less"), normalizeCSS)

	createFile(path.Join(workingDirectory, "resources", "css", "core", "variables.less"), `
@baseColor: red;
`)

	createFile(path.Join(workingDirectory, "resources", "css", "components", "hero.less"), `
.hero {
	background-color: @baseColor;
}
`)
}

func createJSFiles(workingDirectory, projectName string) {
	createDirectory(path.Join(workingDirectory, "resources", "js"))

	createFile(path.Join(workingDirectory, "resources", "js", "tsconfig.json"), `
{
	"compilerOptions": {
		"noImplicitAny": true,
		"removeComments": true,
		"preserveConstEnums": true,
		"sourceMap": false,
		"target": "es5",
		"outFile": "../../public/script.js"
	},
	"files": [
		"init.ts"
	]
}	
`)

	createFile(path.Join(workingDirectory, "resources", "js", "init.ts"), `
console.log("inited");
`)

}

func createFile(path, content string) {

	fmt.Println("Writing to file:", path)
	fmt.Println(content)

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		panic(err)
	}
	f.Truncate(0)

	_, err = f.WriteString(content)
	if err != nil {
		panic(err)
	}

	if err := f.Close(); err != nil {
		panic(err)
	}
}
