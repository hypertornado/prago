package prago

type FormOption struct {
	ID                string
	Name              string
	DescriptionBefore string
	DescriptionAfter  string
	Color             string
	Icon              string
	Style             string
	ImageURL          string
	Button            *Button
}

func (fo *FormOption) GetColor() string {
	if fo.Color != "" {
		return fo.Color
	}
	return getStyleColor(fo.Style)
}

func (form *Form) AddSelectOptions(name, description string, options []*FormOption) *FormItem {
	input := form.addInput(name, description, "form_input_select")
	input.Data = options
	return input
}

func (form *Form) AddSelect(name, description string, values [][2]string) *FormItem {
	return form.AddSelectOptions(name, description, getFormOptions(values))

}

func (form *Form) AddRadioOptions(name, description string, options []*FormOption) *FormItem {
	input := form.addInput(name, description, "form_input_radio")
	input.Data = options
	return input
}

func (form *Form) AddRadio(name, description string, values [][2]string) *FormItem {
	return form.AddRadioOptions(name, description, getFormOptions(values))
}

func getFormOptions(values [][2]string) (ret []*FormOption) {
	for _, v := range values {
		ret = append(ret, &FormOption{
			ID:   v[0],
			Name: v[1],
		})
	}
	return ret
}
