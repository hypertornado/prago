package prago

type BoxHeader struct {
	Icon      string
	Name      string
	TextAfter string
	Image     string
	Tags      []BoxTag
}

type BoxTag struct {
	URL  string
	Icon string
	Name string
}

func (form *Form) GetBoxHeader() *BoxHeader {

	icon := form.Icon

	if icon == "" && form.action != nil {
		icon = form.action.icon
	}

	name := form.Title
	if name == "" && form.action != nil {
		name = form.action.name("cs")
	}

	description := form.Description

	return &BoxHeader{
		Icon:      icon,
		Name:      name,
		Image:     form.image,
		TextAfter: description,
	}
}
