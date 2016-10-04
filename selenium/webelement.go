package selenium

import (
	"fmt"
	"strings"
)

type WebElement struct {
	ELEMENT string `json:"ELEMENT"`
	session *Session
}

func (e *WebElement) GetElementByClassName(value string) (*WebElement, error) {
	return e.getElementUsing("class name", value)
}

func (e *WebElement) GetElementByCSSSelector(value string) (*WebElement, error) {
	return e.getElementUsing("css selector", value)
}

func (e *WebElement) GetElementByID(value string) (*WebElement, error) {
	return e.getElementUsing("id", value)
}

func (e *WebElement) GetElementByName(value string) (*WebElement, error) {
	return e.getElementUsing("name", value)
}

func (e *WebElement) GetElementByLinkText(value string) (*WebElement, error) {
	return e.getElementUsing("link text", value)
}

func (e *WebElement) GetElementByPartialLinkText(value string) (*WebElement, error) {
	return e.getElementUsing("partial link text", value)
}

func (e *WebElement) GetElementByTagName(value string) (*WebElement, error) {
	return e.getElementUsing("tag name", value)
}

func (e *WebElement) GetElementByXPath(value string) (*WebElement, error) {
	return e.getElementUsing("xpath", value)
}

func (e *WebElement) getElementUsing(using, value string) (element *WebElement, err error) {
	data := map[string]string{
		"using": using,
		"value": value,
	}
	err = e.postRequestWithReturnValue("element", data, &element)
	if err != nil {
		return nil, err
	}
	element.session = e.session
	return
}

func (e *WebElement) GetElementsByClassName(value string) ([]*WebElement, error) {
	return e.getElementsUsing("class name", value)
}

func (e *WebElement) GetElementsByCSSSelector(value string) ([]*WebElement, error) {
	return e.getElementsUsing("css selector", value)
}

func (e *WebElement) GetElementsByID(value string) ([]*WebElement, error) {
	return e.getElementsUsing("id", value)
}

func (e *WebElement) GetElementsByName(value string) ([]*WebElement, error) {
	return e.getElementsUsing("name", value)
}

func (e *WebElement) GetElementsByLinkText(value string) ([]*WebElement, error) {
	return e.getElementsUsing("link text", value)
}

func (e *WebElement) GetElementsByPartialLinkText(value string) ([]*WebElement, error) {
	return e.getElementsUsing("partial link text", value)
}

func (e *WebElement) GetElementsByTagName(value string) ([]*WebElement, error) {
	return e.getElementsUsing("tag name", value)
}

func (e *WebElement) GetElementsByXPath(value string) ([]*WebElement, error) {
	return e.getElementsUsing("xpath", value)
}

func (e *WebElement) getElementsUsing(using, value string) (elements []*WebElement, err error) {
	data := map[string]string{
		"using": using,
		"value": value,
	}
	err = e.postRequestWithReturnValue("elements", data, &elements)
	if err != nil {
		return nil, err
	}
	for _, v := range elements {
		v.session = e.session
	}
	return
}

func (e *WebElement) Click() error {
	return e.postRequest("click", nil)
}

func (e *WebElement) Submit() error {
	return e.postRequest("submit", nil)
}

func (e *WebElement) Text() (text string, err error) {
	err = e.getValue("text", &text)
	return
}

func (e *WebElement) SendKeys(sequence string) error {
	keys := make([]string, len(sequence))
	for i, k := range sequence {
		keys[i] = string(k)
	}
	data := map[string]interface{}{
		"value": keys,
	}
	return e.postRequest("value", data)
}

func (e *WebElement) GetTagName() (name string, err error) {
	err = e.getValue("name", &name)
	return
}

func (e *WebElement) Clear() error {
	return e.postRequest("clear", nil)
}

func (e *WebElement) Selected() (selected bool, err error) {
	err = e.getValue("selected", &selected)
	return
}

func (e *WebElement) Enabled() (enabled bool, err error) {
	err = e.getValue("enabled", &enabled)
	return
}

func (e *WebElement) GetAttribute(name string) (attribute string, err error) {
	err = e.getValue(fmt.Sprintf("attribute/%s", name), &attribute)
	return
}

func (e *WebElement) Equals(other *WebElement) (equals bool, err error) {
	err = e.getValue(fmt.Sprintf("equals/%s", other.ELEMENT), &equals)
	return
}

func (e *WebElement) Displayed() (displayed bool, err error) {
	err = e.getValue("displayed", &displayed)
	return
}

func (e *WebElement) Location() (x, y int, err error) {
	location := &location{}
	err = e.getValue("location", location)
	if err != nil {
		return -1, -1, err
	}
	return location.X, location.Y, nil
}

func (e *WebElement) Size() (width, height int, err error) {
	size := &size{}
	err = e.getValue("size", size)
	if err != nil {
		return -1, -1, err
	}
	return size.Width, size.Height, nil
}

func (e *WebElement) GetComputedCSSValue(propertyName string) (value string, err error) {
	err = e.getValue(fmt.Sprintf("css/%s", propertyName), &value)
	return
}

func (e *WebElement) MoveTo() error {
	data := map[string]interface{}{
		"element": e.ELEMENT,
	}
	return e.session.postRequest("moveto", data)
}

func (e *WebElement) MoveToWithOffset(xoffset, yoffset int) error {
	data := map[string]interface{}{
		"element": e.ELEMENT,
		"xoffset": xoffset,
		"yoffset": yoffset,
	}
	return e.session.postRequest("moveto", data)
}

func (e *WebElement) TouchClick() error {
	return e.postRequest("touch/click", nil)
}

func (e *WebElement) TouchScroll(xoffset, yoffset int) error {
	data := map[string]interface{}{
		"element": e.ELEMENT,
		"xoffset": xoffset,
		"yoffset": yoffset,
	}
	return e.session.postRequest("touch/scroll", data)
}

func (e *WebElement) TouchDoubleClick() error {
	return e.postRequest("touch/doubleclick", nil)
}

func (e *WebElement) TouchLongClick() error {
	return e.postRequest("touch/longclick", nil)
}

func (e *WebElement) TouchFlick(xoffset, yoffset, speed int) error {
	data := map[string]interface{}{
		"element": e.ELEMENT,
		"xoffset": xoffset,
		"yoffset": yoffset,
		"speed":   speed,
	}
	return e.session.postRequest("touch/flick", data)
}

func (e *WebElement) HasClass(className string) bool {
	classesStr, err := e.GetAttribute("class")
	if err != nil {
		return false
	}
	for _, class := range strings.Split(classesStr, " ") {
		if class == className {
			return true
		}
	}
	return false
}
