package prago

import "html/template"

type Button struct {
	Icon     string
	Name     string
	URL      string
	OnClick  template.JS
	Title    string
	Selected bool
	Style    string
}

func (btn *Button) SafeURL() template.URL {
	return template.URL(btn.URL)
}

func (btn *Button) ButtonColor() string {
	return getStyleColor(btn.Style)
}

func (btn Button) GetTitle() string {
	if btn.Title != "" {
		return btn.Title
	}
	return btn.Name
}

func (btn *Button) StyleAccented() *Button {
	btn.Style = styleAccented
	return btn
}

func (btn *Button) StyleCreate() *Button {
	btn.Style = styleCreate
	return btn
}

func (btn *Button) StyleDestroy() *Button {
	btn.Style = styleDestroy
	return btn
}
