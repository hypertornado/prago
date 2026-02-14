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
func (f *Form) AddItem(item *FormItem) {
	item.form = f
	f.Items = append(f.Items, item)
}

// BindData to form
func (f *Form) BindData(params url.Values) {
	for _, v := range f.Items {
		v.Value = params.Get(v.ID)
	}
}

func (f *Form) addInput(id, description, template string) *FormItem {
	item := &FormItem{
		ID:       id,
		Template: template,
		Name:     description,
	}
	if description == "" {
		item.HiddenName = true
	}
	item.AddUUID()
	f.AddItem(item)
	return item
}

// AddTextInput to form
func (f *Form) AddTextInput(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input")
	if description != "" {
		input.Icon = iconText
	}
	return input
}

func (f *Form) AddNumberInput(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_int")
	input.Icon = iconNumber
	return input
}

// AddTextareaInput to form
func (f *Form) AddTextareaInput(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_textarea")
	if description != "" {
		input.Icon = iconText
	}
	return input
}

// AddEmailInput to form
func (f *Form) AddEmailInput(name, description string) *FormItem {
	return f.addInput(name, description, "form_input_email")
}

// AddPasswordInput to form
func (f *Form) AddPasswordInput(name, description string) *FormItem {
	return f.addInput(name, description, "form_input_password")
}

// AddFileInput to form
func (f *Form) AddFileInput(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_file")
	input.Icon = iconImage
	return input
}

// AddCAPTCHAInput to form
func (f *Form) AddCAPTCHAInput(name, description string) *FormItem {
	return f.addInput(name, description, "form_input_captcha")
}

// AddSubmit to form
func (f *Form) AddSubmit(description string) *FormItem {
	input := f.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "form_input_submit"
	return input
}

// AddDeleteSubmit to form
func (f *Form) AddDeleteSubmit(description string) *FormItem {
	input := f.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "form_input_delete"
	input.Icon = iconDelete
	return input
}

// AddCheckbox to form
func (f *Form) AddCheckbox(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_checkbox")
	input.HiddenName = true
	return input
}

// AddHidden to form
func (f *Form) AddHidden(name string) *FormItem {
	input := f.addInput(name, "", "")
	input.Template = "form_input_hidden"
	input.Hidden = true
	return input
}

// AddSelect to form
func (f *Form) AddSelect(name, description string, values [][2]string) *FormItem {
	input := f.addInput(name, description, "form_input_select")
	input.Data = values
	input.Icon = iconSelect
	return input
}

func (f *Form) AddRadio(name, description string, values [][2]string) *FormItem {
	input := f.addInput(name, description, "form_input_select_radio")
	input.Data = values
	input.Icon = iconSelect
	return input
}

// AddDatePicker to form
func (f *Form) AddDatePicker(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_date")
	input.Icon = iconDate
	return input
}

func (f *Form) AddDateTimePicker(name, description string) *FormItem {
	input := f.addInput(name, description, "form_input_datetime")
	input.Icon = iconDateTime
	return input
}

func (f *Form) AddRelation(name, description string, relatedResourceID string) *FormItem {
	input := f.addInput(name, description, "form_input_relation")
	input.Data = relationFormDataSource{
		App:       f.app,
		RelatedID: columnName(relatedResourceID),
	}
	input.Icon = f.app.getResourceByID(relatedResourceID).icon
	return input
}

func (f *Form) AddRelationMultiple(name, description string, relatedResourceID string) *FormItem {
	input := f.addInput(name, description, "form_input_relation")
	input.Data = relationFormDataSource{
		App:           f.app,
		RelatedID:     columnName(relatedResourceID),
		MultiRelation: true,
	}
	input.Icon = f.app.getResourceByID(relatedResourceID).icon
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
