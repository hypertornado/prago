package prago

import (
	"html/template"
	"net/url"
)

// Form represents admin form
type Form struct {
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
}

// FormItem represents item of form
type FormItem struct {
	ID          string
	Icon        string
	Name        string
	Description string
	Required    bool
	Focused     bool
	Readonly    bool
	HiddenName  bool
	Hidden      bool
	Template    string
	Value       string
	Data        interface{}
	UUID        string
	form        *Form
}

// NewForm creates new form
func NewForm(action string) *Form {
	ret := &Form{
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
	item.AddUUID()
	f.AddItem(item)
	return item
}

// AddTextInput to form
func (f *Form) AddTextInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_input")
}

// AddTextareaInput to form
func (f *Form) AddTextareaInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_textarea")
}

// AddEmailInput to form
func (f *Form) AddEmailInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_email")
}

// AddPasswordInput to form
func (f *Form) AddPasswordInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_password")
}

// AddFileInput to form
func (f *Form) AddFileInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_file")
}

// AddCAPTCHAInput to form
func (f *Form) AddCAPTCHAInput(name, description string) *FormItem {
	return f.addInput(name, description, "admin_item_captcha")
}

// AddSubmit to form
func (f *Form) AddSubmit(description string) *FormItem {
	input := f.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "admin_item_submit"
	return input
}

// AddDeleteSubmit to form
func (f *Form) AddDeleteSubmit(description string) *FormItem {
	input := f.addInput("_submit", description, "")
	input.HiddenName = true
	input.Template = "admin_item_delete"
	return input
}

// AddCheckbox to form
func (f *Form) AddCheckbox(name, description string) *FormItem {
	input := f.addInput(name, description, "admin_item_checkbox")
	input.HiddenName = true
	return input
}

// AddHidden to form
func (f *Form) AddHidden(name string) *FormItem {
	input := f.addInput(name, "", "")
	input.Template = "admin_item_hidden"
	input.Hidden = true
	return input
}

// AddSelect to form
func (f *Form) AddSelect(name, description string, values [][2]string) *FormItem {
	input := f.addInput(name, description, "admin_item_select")
	input.Data = values
	return input
}

func (f *Form) AddRadio(name, description string, values [][2]string) *FormItem {
	input := f.addInput(name, description, "admin_item_select_radio")
	input.Data = values
	return input
}

// AddDatePicker to form
func (f *Form) AddDatePicker(name, description string) *FormItem {
	input := f.addInput(name, description, "admin_item_date")
	return input
}

func (f *Form) AddDateTimePicker(name, description string) *FormItem {
	input := f.addInput(name, description, "admin_item_datetime")
	return input
}

func (f *Form) AddRelation(name, description string, relatedResourceID string) *FormItem {
	input := f.addInput(name, description, "admin_item_relation")
	input.Data = relationFormDataSource{
		RelatedID: relatedResourceID,
	}
	return input
}

func (f *Form) AddRelationMultiple(name, description string, relatedResourceID string) *FormItem {
	input := f.addInput(name, description, "admin_item_relation")
	input.Data = relationFormDataSource{
		RelatedID:     relatedResourceID,
		MultiRelation: true,
	}
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
