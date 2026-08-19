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
	DescriptionsBefore     []string
	DescriptionsAfter      []string
	Items                  []*FormItem
	Valid                  bool
	CSRFToken              string
	HTMLAfter              template.HTML
	AutosubmitFirstTime    bool
	AutosubmitOnDataChange bool
	ScriptPaths            []string
	ItemVersion            int64
	BoxHeader              *BoxHeader
}

func (form *Form) HeaderName() string {
	return form.GetBoxHeader().Name
}

func (form *Form) HeaderIcon() string {
	return form.GetBoxHeader().Icon
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
		v.SetValue(params.Get(v.ID))
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
	item.SetUUID()
	form.AddItem(item)
	return item
}

// AddText to form
func (form *Form) AddText(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input")
	return input
}

func (form *Form) AddNumber(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_int")
	return input
}

// AddTextarea to form
func (form *Form) AddTextarea(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_textarea")
	return input
}

// AddEmail to form
func (form *Form) AddEmail(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_email")
	return input
}

// AddPassword to form
func (form *Form) AddPassword(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_password")
	return input
}

// AddFile to form
func (form *Form) AddFile(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_file")
	return input
}

// AddCAPTCHA to form
func (form *Form) AddCAPTCHA(name, description string) *FormItem {
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

// AddDatePicker to form
func (form *Form) AddDatePicker(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_date")
	return input
}

func (form *Form) AddDateTimePicker(name, description string) *FormItem {
	input := form.addInput(name, description, "form_input_datetime")
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

func (form *Form) AddCSRFToken(request *Request) *Form {
	form.CSRFToken = request.csrfToken()
	return form
}
