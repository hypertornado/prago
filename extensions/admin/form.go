package admin

import (
	"github.com/asaskevich/govalidator"
	"net/url"
)

//TODO: tests

type Form struct {
	Method  string
	Action  string
	Items   []*FormItem
	ItemMap map[string]*FormItem
	Errors  []string
	Valid   bool
}

type ItemValidator interface {
	Validate(*FormItem)
}

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

func (f *Form) Validate() {
	for _, item := range f.Items {
		for _, validator := range item.validators {
			validator.Validate(item)
		}
	}
}

func NewForm() *Form {
	ret := &Form{}
	ret.ItemMap = make(map[string]*FormItem)
	ret.Valid = true
	return ret
}

func (f *Form) AddItem(item *FormItem) {
	item.form = f
	f.ItemMap[item.Name] = item
	f.Items = append(f.Items, item)
}

func (f *Form) BindData(params url.Values) {
	for _, v := range f.Items {
		v.Value = params.Get(v.Name)
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

func (f *Form) AddTextInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_input", validators)
}

func (f *Form) AddEmailInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_email", validators)
}

func (f *Form) AddPasswordInput(name, description string, validators ...ItemValidator) *FormItem {
	return f.addInput(name, description, "admin_item_password", validators)
}

func (f *Form) AddSubmit(name, description string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "", validators)
	input.Template = "admin_item_submit"
	return input
}

func (f *Form) AddCheckbox(name, description string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "admin_item_checkbox", validators)
	input.HiddenName = true
	return input
}

func (f *Form) AddHidden(name string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, "", "", validators)
	input.Template = "admin_item_hidden"
	return input
}

func (f *Form) AddSelect(name, description string, values [][2]string, validators ...ItemValidator) *FormItem {
	input := f.addInput(name, description, "admin_item_select", validators)
	input.Values = values
	return input
}

func (f *FormItem) AddError(err string) {
	f.Errors = append(f.Errors, err)
	f.form.Valid = false
}

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

func EmailValidator(Error string) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if !govalidator.IsEmail(field.Value) {
			return false
		}
		return true
	}, Error)
}

func NonEmptyValidator(Error string) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) == 0 {
			return false
		}
		return true
	}, Error)
}

func MinLengthValidator(Error string, minLength int) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) <= minLength {
			return false
		}
		return true
	}, Error)
}

func MaxLengthValidator(Error string, maxLength int) ItemValidator {
	return NewValidator(func(field *FormItem) bool {
		if len(field.Value) >= maxLength {
			return false
		}
		return true
	}, Error)
}
