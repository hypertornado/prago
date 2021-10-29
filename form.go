package prago

import (
	"net/url"
)

type Form struct {
	AJAX   bool
	Action string
	Items  []FormItem
}

type FormItem struct {
	ID   string
	Name func(string) string
}

func (form *Form) GetFormView(request *Request) *formView {
	return &formView{
		Form: form,
	}
}

//Form represents admin form
type formView struct {
	Items     []*formItemView
	Valid     bool
	Classes   []string
	CSRFToken string
	Form      *Form
}

//FormItem represents item of form
type formItemView struct {
	ID         string
	NameHuman  string
	Required   bool
	Focused    bool
	Readonly   bool
	HiddenName bool
	Hidden     bool
	Template   string
	Errors     []string
	Value      string
	Data       interface{}
	UUID       string
	form       *formView
}

//NewForm creates new form
func newForm() *Form {
	ret := &Form{}
	return ret
}

//GetItemByName returns form item found by name
func (f *formView) GetItemByID(id string) *formItemView {
	for _, v := range f.Items {
		if v.ID == id {
			return v
		}
	}
	return nil
}

//AddItem adds form item
func (f *formView) AddItem(item *formItemView) {
	item.form = f
	f.Items = append(f.Items, item)
}

//BindData to form
func (f *formView) BindData(params url.Values) {
	for _, v := range f.Items {
		v.Value = params.Get(v.ID)
	}
}

func (f *formView) addInput(id, description, template string) *formItemView {
	item := &formItemView{
		ID:        id,
		Template:  template,
		NameHuman: description,
	}
	item.AddUUID()
	f.AddItem(item)
	return item
}

//AddTextInput to form
func (f *formView) AddTextInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_input")
}

//AddTextareaInput to form
func (f *formView) AddTextareaInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_textarea")
}

//AddEmailInput to form
func (f *formView) AddEmailInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_email")
}

//AddPasswordInput to form
func (f *formView) AddPasswordInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_password")
}

//AddFileInput to form
func (f *formView) AddFileInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_file")
}

//AddCAPTCHAInput to form
func (f *formView) AddCAPTCHAInput(name, description string) *formItemView {
	return f.addInput(name, description, "admin_item_captcha")
}

//AddSubmit to form
func (f *formView) AddSubmit(name, description string) *formItemView {
	input := f.addInput(name, description, "")
	input.HiddenName = true
	input.Template = "admin_item_submit"
	return input
}

//AddDeleteSubmit to form
func (f *formView) AddDeleteSubmit(name, description string) *formItemView {
	input := f.addInput(name, description, "")
	input.HiddenName = true
	input.Template = "admin_item_delete"
	return input
}

//AddCheckbox to form
func (f *formView) AddCheckbox(name, description string) *formItemView {
	input := f.addInput(name, description, "admin_item_checkbox")
	input.HiddenName = true
	return input
}

//AddHidden to form
func (f *formView) AddHidden(name string) *formItemView {
	input := f.addInput(name, "", "")
	input.Template = "admin_item_hidden"
	input.Hidden = true
	return input
}

//AddSelect to form
func (f *formView) AddSelect(name, description string, values [][2]string) *formItemView {
	input := f.addInput(name, description, "admin_item_select")
	input.Data = values
	return input
}

//AddError to form
func (f *formItemView) AddError(err string) {
	f.Errors = append(f.Errors, err)
	f.form.Valid = false
}

//AddUUID to form
func (f *formItemView) AddUUID() {
	f.UUID = "id-" + randomString(5)
}

func (form *formView) AddCSRFToken(request *Request) *formView {
	form.CSRFToken = request.csrfToken()
	return form
}
