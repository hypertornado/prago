package prago

import (
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"mime/multipart"
	"strconv"
	"strings"

	"github.com/hypertornado/prago/pragoai"
)

func multipartFilesToAIFiles(files []*multipart.FileHeader) ([]pragoai.File, error) {
	var ret []pragoai.File
	for _, header := range files {
		file, err := header.Open()
		if err != nil {
			return nil, err
		}
		data, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return nil, err
		}
		ret = append(ret, pragoai.File{
			Data:      data,
			MediaType: header.Header.Get("Content-Type"),
		})
	}
	return ret, nil
}

func parseAIConsoleEnumValues(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	for i, part := range parts {
		parts[i] = strings.TrimSpace(part)
	}
	return parts
}

type aiTestStruct struct {
	Name          string `json:"name" ai:"Jméno"`
	Description   string `json:"description" ai:"Popis"`
	HomeTeamGoals int64  `json:"home" ai:"Počet gólů domací"`
	AwayTeamGoals int64  `json:"away" ai:"Počet gólů hosté"`
}

func (app *App) initAIConsole() {

	app.ActionForm("_aiconsole", func(form *Form, request *Request) {

		var modelOptions [][2]string
		models, err := app.AI.GetModels()
		Must(err)
		for _, model := range models {
			modelOptions = append(modelOptions, [2]string{
				model.ID,
				model.Name,
			})
		}

		form.AddSelect("model", "Modely", modelOptions).SetValue(pragoai.DefaultModelLow)

		fileInput := form.AddFile("file", "Soubory")
		fileInput.FileMultiple = true

		form.AddTextarea("text", "Text").SetFocused()

		form.AddSelect("output", "Výstup", [][2]string{
			{"text", "Text"},
			{"boolean", "Boolean"},
			{"int", "Int"},
			{"int64", "Int64"},
			{"float64", "Float64"},
			{"text_list", "Text (seznam)"},
			{"boolean_list", "Boolean (seznam)"},
			{"int_list", "Int (seznam)"},
			{"int64_list", "Int64 (seznam)"},
			{"float64_list", "Float64 (seznam)"},
			{"ai_test_struct", "AI Test struct"},
		})

		form.AddText("output_description", "Popis výstupní hodnoty")

		form.AddText("enum_values", "Enum hodnoty (odděleny čárkou)")

		effortOptions := [][2]string{{"", ""}}
		for _, effort := range pragoai.AvailableEfforts {
			effortOptions = append(effortOptions, [2]string{effort, effort})
		}
		form.AddSelect("effort", "Effort", effortOptions)

		form.AddNumber("max_tokens", "Maximum tokenů").SetValue(fmt.Sprintf("%d", pragoai.DefaultMaxTokens))

		form.AddSubmit("Odeslat")
	}, func(fv FormValidation, request *Request) {

		text := request.Param("text")
		if text == "" {
			fv.AddItemError("text", "Zadejte text")
		}

		multipartFiles := request.Request().MultipartForm.File["file"]

		maxTokens, err := strconv.Atoi(request.Param("max_tokens"))
		if err != nil || maxTokens <= 0 {
			fv.AddItemError("max_tokens", "Chyba")
		}

		if !fv.Valid() {
			return
		}

		aiFiles, err := multipartFilesToAIFiles(multipartFiles)
		if err != nil {
			fv.AddError(err.Error())
			return
		}

		outputType := request.Param("output")

		baseRequest := pragoai.Request{
			Text:   text,
			Model:  request.Param("model"),
			Effort: request.Param("effort"),
			Files:  aiFiles,
		}

		var outputHTML template.HTML

		if outputType == "text" {
			baseRequest.ResponseDescription = request.Param("output_description")
			baseRequest.EnumValues = parseAIConsoleEnumValues(request.Param("enum_values"))

			response, err := app.AI.Get[string](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML, err = markdownToHTML(response)
			if err != nil {
				fv.AddError(err.Error())
				return
			}
		}

		if outputType == "boolean" {
			response, err := app.AI.Get[bool](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))

		}

		if outputType == "int" {
			response, err := app.AI.Get[int](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))

		}

		if outputType == "int64" {
			response, err := app.AI.Get[int64](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))

		}

		if outputType == "float64" {
			response, err := app.AI.Get[float64](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))

		}

		if outputType == "text_list" {
			baseRequest.ResponseDescription = request.Param("output_description")
			baseRequest.EnumValues = parseAIConsoleEnumValues(request.Param("enum_values"))

			response, err := app.AI.Get[[]string](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			var items []string
			for _, item := range response {
				items = append(items, "<li>"+html.EscapeString(item)+"</li>")
			}
			outputHTML = template.HTML("<ul>" + strings.Join(items, "") + "</ul>")
		}

		if outputType == "boolean_list" {
			baseRequest.ResponseDescription = request.Param("output_description")

			response, err := app.AI.Get[[]bool](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))
		}

		if outputType == "int_list" {
			baseRequest.ResponseDescription = request.Param("output_description")

			response, err := app.AI.Get[[]int](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))
		}

		if outputType == "int64_list" {
			baseRequest.ResponseDescription = request.Param("output_description")

			response, err := app.AI.Get[[]int64](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))
		}

		if outputType == "float64_list" {
			baseRequest.ResponseDescription = request.Param("output_description")

			response, err := app.AI.Get[[]float64](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML(fmt.Sprintf("%v", response))
		}

		if outputType == "ai_test_struct" {
			baseRequest.ResponseDescription = request.Param("output_description")

			response, err := app.AI.Get[aiTestStruct](baseRequest)

			if err != nil {
				fv.AddError(err.Error())
				return
			}

			data, err := json.MarshalIndent(response, " ", " ")
			if err != nil {
				fv.AddError(err.Error())
				return
			}

			outputHTML = template.HTML("<pre>" + string(data) + "</pre>")
		}

		fv.AfterContent(outputHTML)

	}).Permission("sysadmin").Name(Unlocalized("AI Console")).Board(aiBoard).Icon("glyphicons-basic-118-terminal-isolated.svg")

}
