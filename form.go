package prago

import (
	"html/template"
	"net/url"
)

// Form represents admin form
type Form struct {
	app                    *App
	action                 *Action
	image                  string
	Action                 string
	Icon                   string
	Title                  string
	Description            string
	Items                  []*FormItem
	Valid                  bool
	CSRFToken              string
	HTMLAfter              template.HTML
	AutosubmitFirstTime    bool
	AutosubmitOnDataChange bool
	ScriptPaths            []string
	ItemVersion            int64
}

// FormItem represents item of form
type FormItem struct {
	ID                 string
	Icon               string
	Name               string
	Description        string
	DescriptionsBefore []string
	DescriptionsAfter  []string
	Placeholder        string
	Required           bool
	Focused            bool
	Readonly           bool
	HiddenName         bool
	Hidden             bool
	Template           string
	Value              string
	Data               interface{}

	Content template.HTML

	TextOver string

	UUID string
	form *Form

	Autocomplete string
	InputMode    string

	HelpURL string

	FileMultiple bool
	FileAccept   string

	FormFilterID string

	SuggestionURL string
}

func (fi *FormItem) GetContent() template.HTML {
	if fi.Content != "" {
		return fi.Content
	}
	if fi.Template != "" {
		return fi.form.app.adminTemplates.ExecuteToHTML(fi.Template, fi)
	}
	return ""
}

// NewForm creates new form
func (app *App) NewForm(action string) *Form {
	ret := &Form{
		app:    app,
		Action: action,
	}
	return ret
}

// AddItem adds form item
func (form *Form) AddItem(item *FormItem) {
	item.form = form
	form.Items = append(form.Items, item)
}

// BindData to form
func (form *Form) BindData(params url.Values) {
	for _, v := range form.Items {
		v.Value = params.Get(v.ID)
	}
}

func (form *Form) addInput(id, description, template string) *FormItem {
	item := &FormItem{
		ID:       id,
		Template: template,
		Name:     description,
	}
	if description == "" {
		item.HiddenName = true
	}
	item.AddUUID()
	form.AddItem(item)
	return item
}

// AddTextInput to form
func (form *Form) AddTextInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input")
	if description != "" {
		input.Icon = iconText
	}
	return input
}

func (form *Form) AddNumberInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_int")
	input.Icon = iconNumber
	return input
}

// AddTextareaInput to form
func (form *Form) AddTextareaInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_textarea")
	if description != "" {
		input.Icon = iconText
	}
	return input
}

// AddEmailInput to form
func (form *Form) AddEmailInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_email")
	input.Icon = iconEmail
	return input
}

// AddPasswordInput to form
func (form *Form) AddPasswordInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_password")
	input.Icon = iconPassword
	return input
}

// AddFileInput to form
func (form *Form) AddFileInput(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_file")
	input.Icon = iconImage
	return input
}

// AddCAPTCHAInput to form
func (form *Form) AddCAPTCHAInput(name, description string) *FormItem {
	return form.addInput(name, description, "form_input_captcha")
}

// AddSubmit to form
func (form *Form) AddSubmit(description string) *FormItem {
	input := form.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "form_input_submit"
	return input
}

// AddDeleteSubmit to form
func (form *Form) AddDeleteSubmit(description string) *FormItem {
	input := form.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "form_input_delete"
	input.Icon = iconDelete
	return input
}

// AddCheckbox to form
func (form *Form) AddCheckbox(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_checkbox")
	input.HiddenName = true
	return input
}

// AddHidden to form
func (form *Form) AddHidden(name string) *FormItem {
	input := form.addInput(name, "", "")
	input.Template = "form_input_hidden"
	input.Hidden = true
	return input
}

// AddSelect to form
func (form *Form) AddSelect(name, description string, values [][2]string) *FormItem {
	input := form.addInput(name, description, "form_input_select")
	input.Data = values
	input.Icon = iconSelect
	return input
}

type FormOption struct {
	ID                string
	Name              string
	DescriptionBefore string
	DescriptionAfter  string
	ImageURL          string
	Button            *Button
}

func (form *Form) AddOptions(name, description string, options []*FormOption) *FormItem {
	input := form.addInput(name, description, "form_input_select_radio")
	input.Data = options
	input.Icon = iconSelect
	return input
}

func (form *Form) AddRadio(name, description string, values [][2]string) *FormItem {
	var options []*FormOption
	for _, v := range values {
		options = append(options, &FormOption{
			ID:   v[0],
			Name: v[1],
		})
	}
	return form.AddOptions(name, description, options)
}

// AddDatePicker to form
func (form *Form) AddDatePicker(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_date")
	input.Icon = iconDate
	return input
}

func (form *Form) AddDateTimePicker(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_datetime")
	input.Icon = iconDateTime
	return input
}

func (form *Form) AddRelation(name, description string, relatedResourceID string) *FormItem {
	input := form.addInput(name, description, "form_input_relation")
	input.Data = relationFormDataSource{
		App:       form.app,
		RelatedID: columnName(relatedResourceID),
	}
	input.Icon = form.app.getResourceByID(relatedResourceID).icon
	return input
}

func (form *Form) AddRelationMultiple(name, description string, relatedResourceID string) *FormItem {
	input := form.addInput(name, description, "form_input_relation")
	input.Data = relationFormDataSource{
		App:           form.app,
		RelatedID:     columnName(relatedResourceID),
		MultiRelation: true,
	}
	input.Icon = form.app.getResourceByID(relatedResourceID).icon
	return input
}

// AddUUID to form
func (f *FormItem) AddUUID() {
	f.UUID = "id-" + randomString(5)
}

func (form *Form) AddCSRFToken(request *Request) *Form {
	form.CSRFToken = request.csrfToken()
	return form
}

func (f *FormItem) AddFromFilter(formFilter *FormFilter) {
	f.FormFilterID = formFilter.uuid
}
