package selenium

type WebElementTest struct {
	t       *Test
	Element *WebElement
}

func (e *WebElementTest) GetElementByClassName(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByClassName(value))
}
func (e *WebElementTest) GetElementByCSSSelector(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByCSSSelector(value))
}
func (e *WebElementTest) GetElementByID(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByID(value))
}
func (e *WebElementTest) GetElementByName(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByName(value))
}
func (e *WebElementTest) GetElementByLinkText(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByLinkText(value))
}
func (e *WebElementTest) GetElementByPartialLinkText(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByPartialLinkText(value))
}
func (e *WebElementTest) GetElementByTagName(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByTagName(value))
}
func (e *WebElementTest) GetElementByXPath(value string) *WebElementTest {
	return e.t.getEl(e.Element.GetElementByXPath(value))
}

func (e *WebElementTest) GetElementsByClassName(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByClassName(value))
}
func (e *WebElementTest) GetElementsByCSSSelector(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByCSSSelector(value))
}
func (e *WebElementTest) GetElementsByID(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByID(value))
}
func (e *WebElementTest) GetElementsByName(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByName(value))
}
func (e *WebElementTest) GetElementsByLinkText(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByLinkText(value))
}
func (e *WebElementTest) GetElementsByPartialLinkText(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByPartialLinkText(value))
}
func (e *WebElementTest) GetElementsByTagName(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByTagName(value))
}
func (e *WebElementTest) GetElementsByXPath(value string) []*WebElementTest {
	return e.t.getEls(e.Element.GetElementsByXPath(value))
}

func (e *WebElementTest) Click()                   { e.t.err(e.Element.Click()) }
func (e *WebElementTest) Submit()                  { e.t.err(e.Element.Submit()) }
func (e *WebElementTest) Text() string             { return e.t.getStr(e.Element.Text()) }
func (e *WebElementTest) SendKeys(sequence string) { e.t.err(e.Element.SendKeys(sequence)) }
func (e *WebElementTest) GetTagName() string       { return e.t.getStr(e.Element.GetTagName()) }
func (e *WebElementTest) Clear()                   { e.t.err(e.Element.Clear()) }
func (e *WebElementTest) Selected() bool           { return e.t.getBool(e.Element.Selected()) }
func (e *WebElementTest) Enabled() bool            { return e.t.getBool(e.Element.Enabled()) }

func (e *WebElementTest) GetAttribute(name string) string {
	return e.t.getStr(e.Element.GetAttribute(name))
}

func (e *WebElementTest) Equals(o *WebElementTest) bool {
	return e.t.getBool(e.Element.Equals(o.Element))
}

func (e *WebElementTest) Displayed() bool { return e.t.getBool(e.Element.Displayed()) }

func (e *WebElementTest) Location() (x, y int) {
	return e.t.getInts(e.Element.Location())
}

func (e *WebElementTest) Size() (width, height int) {
	return e.t.getInts(e.Element.Size())
}

func (e *WebElementTest) GetComputedCSSValue(propertyName string) (value string) {
	return e.t.getStr(e.Element.GetComputedCSSValue(propertyName))
}

func (e *WebElementTest) MoveTo()                   { e.t.err(e.Element.MoveTo()) }
func (e *WebElementTest) MoveToWithOffset(x, y int) { e.t.err(e.Element.MoveToWithOffset(x, y)) }
func (e *WebElementTest) TouchClick()               { e.t.err(e.Element.TouchClick()) }
func (e *WebElementTest) TouchScroll(xoffset, yoffset int) {
	e.t.err(e.Element.TouchScroll(xoffset, yoffset))
}
func (e *WebElementTest) TouchDoubleClick() { e.t.err(e.Element.TouchDoubleClick()) }
func (e *WebElementTest) TouchLongClick()   { e.t.err(e.Element.TouchLongClick()) }
func (e *WebElementTest) TouchFlick(xoffset, yoffset, speed int) {
	e.t.err(e.Element.TouchFlick(xoffset, yoffset, speed))
}

func (e *WebElementTest) HasClass(className string) bool {
	return e.Element.HasClass(className)
}
