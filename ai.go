package prago

import (
	"fmt"
	"html/template"
	"log"
	"mime/multipart"
	"reflect"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/net/context"

	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func (app *App) getGeminiAPIKey() string {
	return app.mustGetSetting("gemini_api_key")
}

func (app *App) getAvailableAImodels() (ret [][2]string) {

	ret = append(ret, [2]string{"", ""})

	ctx := context.Background()
	// Připojení pomocí API klíče
	client, err := genai.NewClient(ctx, option.WithAPIKey(app.getGeminiAPIKey()))
	if err != nil {
		return nil
	}
	defer client.Close()

	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		ret = append(ret, [2]string{m.Name, m.Name})
		//fmt.Printf("- %s\n", m.Name)
	}

	return ret

	//return app.mustGetSetting("gemini_api_key")
}

func (app *App) initAI() {

	app.Setting("gemini_api_key", "sysadmin")

	PopupForm(app, "_aicontextresource", func(form *Form, request *Request) {

		resource := app.getResourceByID(request.Param("_resource"))

		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			checkboxItem := form.AddCheckbox(field.id, field.name(request.Locale()))
			checkboxItem.Value = "on"
		}

		for _, stat := range resource.itemStats {
			if !request.Authorize(stat.Permission) {
				continue
			}
			checkboxItem := form.AddCheckbox(stat.id, stat.Name(request.Locale()))
			checkboxItem.Value = "on"
		}

		if resource.previewFn != nil {
			form.AddCheckbox("_previewurl", "Ukázat preview URL").Value = "on"
		}

		form.AddHidden("_resource").Value = resource.id
		form.AddRelationMultiple("_items", resource.pluralName(request.Locale()), resource.id).Value = request.Param("_ids")

		form.AddSubmit("Zobrazit")

	}, func(fv FormValidation, request *Request) {

		var fields []*Field

		resource := app.getResourceByID(request.Param("_resource"))
		for _, field := range resource.fields {
			if !request.Authorize(field.canView) {
				continue
			}
			if request.Param(field.id) == "on" {
				fields = append(fields, field)
			}
		}

		ids := MultirelationStringToArray(request.Param("_items"))
		if len(ids) == 0 {
			fv.AddItemError("_items", "Není vybrána žádná položka")
		}

		if !fv.Valid() {
			return
		}

		var strVal string

		for _, id := range ids {
			item := resource.query(context.Background()).ID(id)

			for _, field := range fields {
				ifaceVal := reflect.ValueOf(item).Elem().FieldByName(field.fieldClassName).Interface()

				cellData := getCellViewData(request, field, ifaceVal)
				strVal += fmt.Sprintf("%s: %v\n", field.name(request.Locale()), cellData.Name)
			}

			for _, stat := range resource.itemStats {
				if !request.Authorize(stat.Permission) {
					continue
				}

				if request.Param(stat.id) == "on" {
					strVal += fmt.Sprintf("%s: %v\n", stat.Name(request.Locale()), stat.Handler(item))
				}
			}

			if resource.previewFn != nil && request.Param("_previewurl") == "on" {
				prevURL := resource.previewFn(item)
				strVal += fmt.Sprintf("Veřejné URL: %v\n", prevURL)
			}

			strVal += "\n\n---------\n\n"
		}

		fv.AfterContent(template.HTML(fmt.Sprintf("<textarea class=\"input\">%s</textarea>", template.HTMLEscapeString(strVal))))

	}).Permission(loggedPermission).Name(unlocalized("AI Kontext"))

	ActionForm(app, "_aichat", func(form *Form, request *Request) {
		form.AddSelect("model", "Model", app.getAvailableAImodels()).Value = "models/gemini-flash-latest"
		form.AddTextareaInput("text", "Text").Focused = true
		fileInput := form.AddFileInput("files", "Soubory")
		fileInput.FileMultiple = true
		form.AddSubmit("Odeslat")
	}, func(fv FormValidation, request *Request) {

		model := request.Param("model")
		if model == "" {
			fv.AddItemError("model", "Vyberte model")
		}

		if !fv.Valid() {
			return
		}

		var files []*multipart.FileHeader
		if request.Request().MultipartForm != nil {
			files = request.Request().MultipartForm.File["files"]
		}
		answer, err := app.getAIAnswer(request.Param("text"), model, files)
		if err != nil {
			fv.AddError(err.Error())
			return
		}
		fv.AfterContent(template.HTML(answer))

	}).Permission("sysadmin").Name(unlocalized("AI")).Board(app.optionsBoard).Icon(iconAI)
}

func (app *App) getAIAnswer(in, modelID string, files []*multipart.FileHeader) (string, error) {

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(app.getGeminiAPIKey()))
	if err != nil {
		return "", err
	}
	defer client.Close()

	model := client.GenerativeModel(modelID)

	parts := []genai.Part{genai.Text(in)}

	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			log.Printf("opening uploaded file %s: %v", fh.Filename, err)
			continue
		}
		uploaded, err := client.UploadFile(ctx, "", f, &genai.UploadFileOptions{
			DisplayName: fh.Filename,
			MIMEType:    fh.Header.Get("Content-Type"),
		})
		f.Close()
		if err != nil {
			log.Printf("uploading file %s to Gemini: %v", fh.Filename, err)
			continue
		}
		parts = append(parts, genai.FileData{URI: uploaded.URI})
	}

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", err
	}

	part := resp.Candidates[0].Content.Parts[0]
	if text, ok := part.(genai.Text); ok {
		return string(text), nil
	}
	return "", nil
}
