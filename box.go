package prago

type BoxHeader struct {
	Style              string
	Image              string
	Icon               string
	DescriptionsBefore []string
	Name               string
	DescriptionsAfter  []string

	Buttons []*Button
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

	var style string
	if form.action != nil {
		style = form.action.style

	}

	ret := &BoxHeader{
		Icon:  icon,
		Name:  name,
		Image: form.image,
		Style: style,
	}

	ret.DescriptionsBefore = form.DescriptionsBefore
	ret.DescriptionsAfter = form.DescriptionsAfter

	return ret
}

func (form *Form) Description(text string) {

	form.DescriptionsAfter = append(form.DescriptionsAfter, text)

}
