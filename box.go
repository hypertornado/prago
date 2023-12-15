package prago

type boxHeader struct {
	Icon      string
	Name      string
	TextAfter string
	Image     string

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

	return &boxHeader{
		Icon:      icon,
		Name:      name,
		Image:     form.image,
		TextAfter: description,
	}
}
