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
	//Method string
	//Action string
	Items []*formItemView
	//Errors    []string
	Valid     bool
	Classes   []string
	CSRFToken string
	Form      *Form
}

//ItemValidator represents validator for form item
type itemValidator interface {
	Validate(*formItemView)
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
	validators []itemValidator
}

//Validate form
func (f *formView) Validate() bool {
	f.Valid = true
	for _, item := range f.Items {
		for _, validator := range item.validators {
			validator.Validate(item)
		}
	}
	return f.Valid
}

//NewForm creates new form
func newForm() *Form {
	ret := &Form{}
	//ret.Valid = true
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

func (f *formView) addInput(id, description, template string, validators []itemValidator) *formItemView {
	item := &formItemView{
		ID:        id,
		Template:  template,
		NameHuman: description,
	}
	item.AddUUID()
	item.validators = validators
	f.AddItem(item)
	return item
}

//AddTextInput to form
func (f *formView) AddTextInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_input", validators)
}

//AddTextareaInput to form
func (f *formView) AddTextareaInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_textarea", validators)
}

//AddEmailInput to form
func (f *formView) AddEmailInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_email", validators)
}

//AddPasswordInput to form
func (f *formView) AddPasswordInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_password", validators)
}

//AddFileInput to form
func (f *formView) AddFileInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_file", validators)
}

//AddCAPTCHAInput to form
func (f *formView) AddCAPTCHAInput(name, description string, validators ...itemValidator) *formItemView {
	return f.addInput(name, description, "admin_item_captcha", validators)
}

//AddSubmit to form
func (f *formView) AddSubmit(name, description string, validators ...itemValidator) *formItemView {
	input := f.addInput(name, description, "", validators)
	input.HiddenName = true
	input.Template = "admin_item_submit"
	return input
}

//AddDeleteSubmit to form
func (f *formView) AddDeleteSubmit(name, description string, validators ...itemValidator) *formItemView {
	input := f.addInput(name, description, "", validators)
	input.HiddenName = true
	input.Template = "admin_item_delete"
	return input
}

//AddCheckbox to form
func (f *formView) AddCheckbox(name, description string, validators ...itemValidator) *formItemView {
	input := f.addInput(name, description, "admin_item_checkbox", validators)
	input.HiddenName = true
	return input
}

//AddHidden to form
func (f *formView) AddHidden(name string, validators ...itemValidator) *formItemView {
	input := f.addInput(name, "", "", validators)
	input.Template = "admin_item_hidden"
	input.Hidden = true
	return input
}

//AddSelect to form
func (f *formView) AddSelect(name, description string, values [][2]string, validators ...itemValidator) *formItemView {
	input := f.addInput(name, description, "admin_item_select", validators)
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

//NewValidator creates new item validator with error message
func newValidator(fn func(field *formItemView) bool, message string) itemValidator {
	return validator{
		fn:      fn,
		message: message,
	}
}

type validator struct {
	fn      func(field *formItemView) bool
	message string
}

func (v validator) Validate(field *formItemView) {
	if !v.fn(field) {
		field.AddError(v.message)
	}
}

//EmailValidator for validation of email inputs
/*func emailValidator(Error string) itemValidator {
	return newValidator(func(field *formItemView) bool {
		if !govalidator.IsEmail(field.Value) {
			return false
		} else {
			return true
		}
	}, Error)
}

//ValueValidator for validation field value
func valueValidator(ExpectedValue, Error string) itemValidator {
	return newValidator(func(field *formItemView) bool {
		if field.Value != ExpectedValue {
			return false
		} else {
			return true
		}
	}, Error)
}

//NonEmptyValidator for validation of nonempty inputs
func nonEmptyValidator(Error string) itemValidator {
	return newValidator(func(field *formItemView) bool {
		if len(field.Value) == 0 {
			return false
		} else {
			return true
		}
	}, Error)
}*/

//MinLengthValidator for validation of min length of field
func minLengthValidator(Error string, minLength int) itemValidator {
	return newValidator(func(field *formItemView) bool {
		if len(field.Value) < minLength {
			return false
		} else {
			return true
		}
	}, Error)
}
