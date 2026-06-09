package prago

type boxHeader struct {
	Style      string
	Icon       string
	TextBefore string
	Name       string
	TextAfter  string
	Image      string

	Buttons []*buttonData
}

func (form *Form) GetBoxHeader() *boxHeader {

	icon := form.Icon

	if icon == "" && form.action != nil {
		icon = form.action.icon
	}

	name := form.Title
	if name == "" && form.action != nil {
		name = form.action.name("cs")
	}

	description := form.Description

	var style string
	if form.action != nil {
		style = form.action.style

	}

	return &boxHeader{
		Icon:      icon,
		Name:      name,
		Image:     form.image,
		TextAfter: description,
		Style:     style,
	}
}
