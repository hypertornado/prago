package prago

import (
	"net/url"

	"github.com/asaskevich/govalidator"
)

//Form represents admin form
type form struct {
	Method    string
	Action    string
	Items     []*formItem
	Errors    []string
	Valid     bool
	Classes   []string
	CSRFToken string
}

//ItemValidator represents validator for form item
type itemValidator interface {
	Validate(*formItem)
}

//FormItem represents item of form
type formItem struct {
	Name       string
	NameHuman  string
	Required   bool
	Focused    bool
	Readonly   bool
	HiddenName bool
	Template   string
	Errors     []string
	Value      string
	Data       interface{}
	UUID       string
	form       *form
	validators []itemValidator
}

//Validate form
func (f *form) Validate() bool {
	f.Valid = true
	for _, item := range f.Items {
		for _, validator := range item.validators {
			validator.Validate(item)
		}
	}
	return f.Valid
}

//NewForm creates new form
func newForm() *form {
	ret := &form{}
	ret.Valid = true
	return ret
}

//GetItemByName returns form item found by name
func (f *form) GetItemByName(name string) *formItem {
	for _, v := range f.Items {
		if v.Name == name {
			return v
		}
	}
	return nil
}

//AddItem adds form item
func (f *form) AddItem(item *formItem) {
	item.form = f
	f.Items = append(f.Items, item)
}

//BindData to form
func (f *form) BindData(params url.Values) {
	for _, v := range f.Items {
		v.Value = params.Get(v.Name)
	}
}

func (f *form) addInput(name, description, template string, validators []itemValidator) *formItem {
	item := &formItem{
		Name:      name,
		Template:  template,
		NameHuman: description,
	}
	item.AddUUID()
	item.validators = validators
	f.AddItem(item)
	return item
}

//AddTextInput to form
func (f *form) AddTextInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_input", validators)
}

//AddTextareaInput to form
func (f *form) AddTextareaInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_textarea", validators)
}

//AddEmailInput to form
func (f *form) AddEmailInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_email", validators)
}

//AddPasswordInput to form
func (f *form) AddPasswordInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_password", validators)
}

//AddFileInput to form
func (f *form) AddFileInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_file", validators)
}

//AddCAPTCHAInput to form
func (f *form) AddCAPTCHAInput(name, description string, validators ...itemValidator) *formItem {
	return f.addInput(name, description, "admin_item_captcha", validators)
}

//AddSubmit to form
func (f *form) AddSubmit(name, description string, validators ...itemValidator) *formItem {
	input := f.addInput(name, description, "", validators)
	input.HiddenName = true
	input.Template = "admin_item_submit"
	return input
}

//AddDeleteSubmit to form
func (f *form) AddDeleteSubmit(name, description string, validators ...itemValidator) *formItem {
	input := f.addInput(name, description, "", validators)
	input.HiddenName = true
	input.Template = "admin_item_delete"
	return input
}

//AddCheckbox to form
func (f *form) AddCheckbox(name, description string, validators ...itemValidator) *formItem {
	input := f.addInput(name, description, "admin_item_checkbox", validators)
	input.HiddenName = true
	return input
}

//AddHidden to form
func (f *form) AddHidden(name string, validators ...itemValidator) *formItem {
	input := f.addInput(name, "", "", validators)
	input.Template = "admin_item_hidden"
	return input
}

//AddSelect to form
func (f *form) AddSelect(name, description string, values [][2]string, validators ...itemValidator) *formItem {
	input := f.addInput(name, description, "admin_item_select", validators)
	input.Data = values
	return input
}

//AddError to form
func (f *formItem) AddError(err string) {
	f.Errors = append(f.Errors, err)
	f.form.Valid = false
}

//AddUUID to form
func (f *formItem) AddUUID() {
	f.UUID = "id-" + randomString(5)
}

//NewValidator creates new item validator with error message
func newValidator(fn func(field *formItem) bool, message string) itemValidator {
	return validator{
		fn:      fn,
		message: message,
	}
}

type validator struct {
	fn      func(field *formItem) bool
	message string
}

func (v validator) Validate(field *formItem) {
	if !v.fn(field) {
		field.AddError(v.message)
	}
}

//EmailValidator for validation of email inputs
func emailValidator(Error string) itemValidator {
	return newValidator(func(field *formItem) bool {
		if !govalidator.IsEmail(field.Value) {
			return false
		}
		return true
	}, Error)
}

//ValueValidator for validation field value
func valueValidator(ExpectedValue, Error string) itemValidator {
	return newValidator(func(field *formItem) bool {
		if field.Value != ExpectedValue {
			return false
		}
		return true
	}, Error)
}

//NonEmptyValidator for validation of nonempty inputs
func nonEmptyValidator(Error string) itemValidator {
	return newValidator(func(field *formItem) bool {
		if len(field.Value) == 0 {
			return false
		}
		return true
	}, Error)
}

//MinLengthValidator for validation of min length of field
func minLengthValidator(Error string, minLength int) itemValidator {
	return newValidator(func(field *formItem) bool {
		if len(field.Value) < minLength {
			return false
		}
		return true
	}, Error)
}

//MaxLengthValidator for validation of max length of field
func maxLengthValidator(Error string, maxLength int) itemValidator {
	return newValidator(func(field *formItem) bool {
		if len(field.Value) >= maxLength {
			return false
		}
		return true
	}, Error)
}
