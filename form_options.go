package prago

import (
	"fmt"
	"net/url"
)

type FormOption struct {
	ID                string
	Name              string
	DescriptionBefore string
	DescriptionAfter  string
	TextAfter         string
	Color             string
	Icon              string
	Style             string
	ImageURL          string
	Button            *Button
	//Google Fonts family name (e.g. "Playfair Display"), option name is rendered in this font
	Font string
}

func (fo *FormOption) GetColor() string {
	if fo.Color != "" {
		return fo.Color
	}
	return getStyleColor(fo.Style)
}

func (fo *FormOption) FontURL() string {
	return "https://fonts.googleapis.com/css2?family=" + url.QueryEscape(fo.Font) + "&display=swap"
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

func previewerToFormOption(preview *previewer) *FormOption {
	ret := &FormOption{
		ID:        fmt.Sprintf("%d", preview.ID()),
		Name:      preview.Name(),
		TextAfter: preview.DescriptionExtended(nil),
		ImageURL:  preview.ImageURL(),
	}

	return ret
}
