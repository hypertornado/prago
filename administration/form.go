package administration

import (
	"github.com/asaskevich/govalidator"
	"net/url"
)

//Form represents admin form
type Form struct {
	Method  string
	Action  string
	Items   []*FormItem
	Errors  []string
	Valid   bool
	Classes []string
}

//ItemValidator represents validator for form item
type ItemValidator interface {
	Validate(*FormItem)
}

//FormItem represents item of form
type FormItem struct {
	Name        string
	NameHuman   string
	Required    bool
	Focused     bool
	Readonly    bool
	HiddenName  bool
	SubTemplate string
	Template    string
	Errors      []string
	Value       string
	Values      interface{}
	form        *Form
	validators  []ItemValidator
}

//Validate form
func (f *Form) Validate() {
	for _, item := range f.Items {
		for _, validator := range item.validators {
			validator.Validate(item)
		}
	}
}

//NewForm creates new form
func NewForm() *Form {
	ret := &Form{}
	ret.Valid = true
	return ret
}

//GetItemByName returns form item found by name
func (f *Form) GetItemByName(name string) *FormItem {
	for _, v := range f.Items {
		if v.Name == name {
			return v
		}
	}
	return nil
}

//AddItem adds form item
func (f *Form) AddItem(item *FormItem) {
	item.form = f
	f.Items = append(f.Items, item)
}

//BindData to form
func (f *Form) BindData(params url.Values) {
	for _, v := range f.Items {
		v.Value = params.Get(v.Name)
	}
}

//GetFilter returns struct filter
func (f *Form) getFilter() structFieldFilter {
	allowed := make(map[string]bool)
	for _, v := range f.Items {
		if !v.Readonly {
			allowed[v.Name] = true
		}
	}
	return func(resource Resource, user User, ff field) bool {
		return allowed[ff.Name]
	}
}

func (f *Form) addInput(name, description, template string, validators []ItemValidator) *FormItem {
	item := &FormItem{
		Name:        name,
		SubTemplate: template,
		NameHuman:   description,
	}
	item.validators = validators
	f.AddItem(item)
	return item
}

//AddTextInput to form
func (f *Form) AddTextInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_input", validators)
}

//AddTextareaInput to form
func (f *Form) AddTextareaInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_textarea", validators)
}

//AddEmailInput to form
func (f *Form) AddEmailInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_email", validators)
}

//AddPasswordInput to form
func (f *Form) AddPasswordInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_password", validators)
}

//AddFileInput to form
func (f *Form) AddFileInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_file", validators)
}

//AddSubmit to form
func (f *Form) AddSubmit(name, description string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "", validators)
	input.Template = "admin_item_submit"
	return input
}

//AddCheckbox to form
func (f *Form) AddCheckbox(name, description string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "admin_item_checkbox", validators)
	input.HiddenName = true
	return input
}

//AddHidden to form
func (f *Form) AddHidden(name string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, "", "", validators)
	input.Template = "admin_item_hidden"
	return input
}

//AddSelect to form
func (f *Form) AddSelect(name, description string, values [][2]string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "admin_item_select", validators)
	input.Values = values
	return input
}

//AddError to form
func (f *FormItem) AddError(err string) {
	f.Errors = append(f.Errors, err)
	f.form.Valid = false
}

//NewValidator creates new item validator with error message
func NewValidator(fn func(field *FormItem) bool, message string) ItemValidator {
	return validator{
		fn:      fn,
		message: message,
	}
}

type validator struct {
	fn      func(field *FormItem) bool
	message string
}

func (v validator) Validate(field *FormItem) {
	if !v.fn(field) {
		field.AddError(v.message)
	}
}

//EmailValidator for validation of email inputs
func EmailValidator(Error string) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if !govalidator.IsEmail(field.Value) {
			return false
		}
		return true
	}, Error)
}

//NonEmptyValidator for validation of nonempty inputs
func NonEmptyValidator(Error string) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) == 0 {
			return false
		}
		return true
	}, Error)
}

//MinLengthValidator for validation of min length of field
func MinLengthValidator(Error string, minLength int) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) < minLength {
			return false
		}
		return true
	}, Error)
}

//MaxLengthValidator for validation of max length of field
func MaxLengthValidator(Error string, maxLength int) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) >= maxLength {
			return false
		}
		return true
	}, Error)
}
