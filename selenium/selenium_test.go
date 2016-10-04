package selenium

import (
	"strings"
	"testing"
)

func init() {
	debugApi = false
}

func prepareSeleniumSession(t *testing.T) *SessionTest {
	StartTestServer()
	d := NewDriver("http://localhost:9515")
	return d.NewTestSession(t)
}

func TestNotEquals(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	el1 := session.GetElementByID("linkid")
	el2 := session.GetElementByID("id1")

	ok := el1.Equals(el2)
	if ok {
		t.Fatal("should not be equal")
	}
	session.Delete()
}

func TestProperties(t *testing.T) {
	session := prepareSeleniumSession(t)
	session.SetURL("http://localhost:8587/test")

	el := session.GetElementByClassName("foo")

	w, h := el.Size()
	if w != 20 || h != 30 {
		t.Error("bad sizes")
	}

	cssValue := el.GetComputedCSSValue("color")
	if cssValue != "rgba(255, 0, 0, 1)" {
		t.Error(cssValue)
	}

	attr := el.GetAttribute("data-id")
	if attr != "x" {
		t.Error(attr)
	}

	tagName := el.GetTagName()
	if tagName != "div" {
		t.Error(tagName)
	}

	x, y := el.Location()
	if x != 3 {
		t.Error(x)
	}
	if y != 2 {
		t.Error(y)
	}

	title := session.GetTitle()
	if title != "foo" {
		t.Error(title)
	}

	source := session.GetSource()
	if !strings.Contains(source, "foo") {
		t.Error("not source")
	}

	button := session.GetElementByID("btn")
	button.MoveTo()
	session.Click()

	text := button.Text()
	if text != "clicked" {
		t.Error(text)
	}

	session.DoubleClick()

	text = button.Text()
	if text != "dblclicked" {
		t.Error(text)
	}

	session.Delete()
}

func TestSelectors(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	elements := []*WebElementTest{}

	elements = append(elements, session.GetElementByClassName("link"))
	elements = append(elements, session.GetElementByCSSSelector("#linkid"))
	elements = append(elements, session.GetElementByID("linkid"))
	elements = append(elements, session.GetElementByName("ipsum"))
	elements = append(elements, session.GetElementByLinkText("lorem"))
	elements = append(elements, session.GetElementByPartialLinkText("lore"))
	elements = append(elements, session.GetElementByTagName("a"))
	elements = append(elements, session.GetElementByXPath("//a"))

	firstElement := elements[0]
	for i := 1; i < len(elements); i++ {
		el := elements[i]
		ok := firstElement.Equals(el)
		if !ok {
			t.Fatal(i)
		}
	}

	elements = session.GetElementsByClassName("baz")

	if len(elements) != 2 {
		t.Error(len(elements))
	}

	tagName := elements[0].GetTagName()
	if tagName != "div" {
		t.Error(tagName)
	}

	container := session.GetElementByClassName("container")
	el1 := container.GetElementByClassName("x")
	els1 := session.GetElementsByClassName("x")
	els2 := container.GetElementsByClassName("x")

	ok := el1.Equals(els1[1])
	if !ok {
		t.Error("not equal")
	}

	ok = el1.Equals(els2[0])
	if !ok {
		t.Error("not equal")
	}

	session.Delete()

}

func TestSelenium(t *testing.T) {
	session := prepareSeleniumSession(t)

	status, _ := session.Session.driver.Status()
	if len(status.Build.Version) == 0 {
		t.Fatal("short")
	}

	if len(status.OS.Name) == 0 {
		t.Fatal("short")
	}

	if len(session.Session.SessionID) == 0 {
		t.Fatal("short session")
	}

	session.SetURL("http://localhost:8587/test")
	title := session.GetTitle()
	if title != "foo" {
		t.Fatal(title)
	}

	el := session.GetElementByID("id1")

	txt := el.Text()
	if txt != "bars" {
		t.Fatal(txt)
	}

	_, err := session.Session.GetOrientation()
	if err != NotImplementedError {
		t.Error(err)
	}

	session.Execute("console.log(1);", []interface{}{})

	logTypes := session.GetLogTypes()
	if len(logTypes) == 0 {
		t.Fatal("zero log types")
	}
	for _, v := range logTypes {
		session.GetLog(v)
	}

	session.Delete()
}

func TestWindows(t *testing.T) {
	session := prepareSeleniumSession(t)

	windows := session.GetWindows()
	if len(windows) != 1 {
		t.Error(len(windows))
	}

	cw := session.GetCurrentWindow()
	cw.SetSize(400, 450)

	w, h := session.GetCurrentWindowSize()
	if w != 400 {
		t.Error(w)
	}
	if h != 450 {
		t.Error(h)
	}

	cw.SetPosition(35, 36)

	x, y := session.GetCurrentWindowPosition()
	if x != 35 {
		t.Error(x)
	}
	if y != 36 {
		t.Error(y)
	}

	session.MaximizeCurrentWindow()

	w, _ = cw.GetSize()
	if !(w > 400) {
		t.Error(w)
	}

	session.Delete()
}

func TestNavigation(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/time")
	source1 := session.GetSource()
	session.Refresh()
	source2 := session.GetSource()

	if source1 == source2 {
		t.Error(source1)
	}

	aUrl := "http://localhost:8587/a"
	bUrl := "http://localhost:8587/b"

	session.SetURL("http://localhost:8587/a")

	el := session.GetElementByTagName("a")
	el.Click()

	url := session.GetURL()
	if url != bUrl {
		t.Error(url)
	}

	session.Back()
	url = session.GetURL()
	if url != aUrl {
		t.Error(url)
	}

	session.Forward()
	url = session.GetURL()
	if url != bUrl {
		t.Error(url)
	}

	session.Delete()
}

func TestCookie(t *testing.T) {
	session := prepareSeleniumSession(t)

	cookie := &Cookie{}
	cookie.Name = "bar"
	cookie.Value = "baz"
	cookie.Path = "/"
	err := session.Session.SetCookie(cookie)
	if err == nil {
		t.Error("no url set yet")
	}

	session.SetURL("http://localhost:8587/test")

	cookie = &Cookie{}
	cookie.Name = "bar"
	cookie.Value = "baz"
	cookie.Path = "/"
	session.SetCookie(cookie)

	cookie = &Cookie{}
	cookie.Name = "bar2"
	cookie.Value = "baz2"
	cookie.Path = "/"
	session.SetCookie(cookie)

	cookies := session.GetCookies()
	if len(cookies) != 2 {
		t.Error(cookies)
	}

	session.DeleteCookie("bar2")
	cookies = session.GetCookies()
	if len(cookies) != 1 {
		t.Error(cookies)
	}

	cookie = cookies[0]
	if cookie.Name != "bar" {
		t.Error(cookie.Name)
	}
	if cookie.Value != "baz" {
		t.Error(cookie.Value)
	}

	session.DeleteCookies()
	cookies = session.GetCookies()
	if len(cookies) != 0 {
		t.Error(cookies)
	}

	session.Delete()
}

func TestScript(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	ret := session.Execute("return arguments[0]+'bar'.toUpperCase();", []interface{}{"x"})
	str, ok := ret.(string)
	if !ok {
		t.Error("not string")
	}
	if str != "xBAR" {
		t.Error(str)
	}

	err := session.Execute("alert('hello');", []interface{}{})

	text := session.GetAlertText()
	if text != "hello" {
		t.Error(text)
	}

	session.AcceptAlert()
	_, err = session.Session.GetAlertText()
	if err != NoAlertOpenError {
		t.Error(err)
	}

	session.Delete()
}

func TestKeys(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	textarea := session.GetElementByID("textarea")
	text := textarea.Text()
	if text != "lorem" {
		t.Fatal(text)
	}

	textarea.Click()

	textarea.SendKeys("abc")
	textarea.SendKeys(KeyCode.LeftArrow)
	textarea.SendKeys(KeyCode.LeftArrow)
	textarea.SendKeys(KeyCode.LeftArrow)
	textarea.SendKeys(KeyCode.Backspace)
	textarea.SendKeys(KeyCode.RightArrow)
	textarea.SendKeys("X")

	text = textarea.GetAttribute("value")
	if text != "loreaXbc" {
		t.Error(text)
	}
	session.Delete()
}

func TestSessionStorage(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	session.DeleteSessionStorage()
	session.SetSessionStorage("a", "lorem")
	keys := session.GetSessionStorage()
	if len(keys) != 1 {
		t.Error(len(keys))
	}
	if keys[0] != "a" {
		t.Error(keys[0])
	}

	if session.GetSessionStorageSize() != 1 {
		t.Error("not 1")
	}
	session.SetSessionStorage("b", "ipsum")
	if session.GetSessionStorageValue("a") != "lorem" {
		t.Error("wrong value")
	}
	if session.GetSessionStorageSize() != 2 {
		t.Error("not 2")
	}
	session.DeleteSessionStorageValue("a")
	if session.GetSessionStorageSize() != 1 {
		t.Error("not 1")
	}
	session.DeleteSessionStorage()
	if session.GetSessionStorageSize() != 0 {
		t.Error("not 0")
	}

	session.Delete()
}

func TestLocalStorage(t *testing.T) {
	session := prepareSeleniumSession(t)

	session.SetURL("http://localhost:8587/test")

	session.DeleteLocalStorage()
	session.SetLocalStorage("a", "lorem")
	keys := session.GetLocalStorage()
	if len(keys) != 1 {
		t.Error(len(keys))
	}
	if keys[0] != "a" {
		t.Error(keys[0])
	}

	if session.GetLocalStorageSize() != 1 {
		t.Error("not 1")
	}
	session.SetLocalStorage("b", "ipsum")
	if session.GetLocalStorageValue("a") != "lorem" {
		t.Error("wrong value")
	}
	if session.GetLocalStorageSize() != 2 {
		t.Error("not 2")
	}
	session.DeleteLocalStorageValue("a")
	if session.GetLocalStorageSize() != 1 {
		t.Error("not 1")
	}
	session.DeleteLocalStorage()
	if session.GetLocalStorageSize() != 0 {
		t.Error("not 0")
	}

	session.Delete()
}

func TestHasClass(t *testing.T) {
	session := prepareSeleniumSession(t)
	session.SetURL("http://localhost:8587/test")

	if !session.GetElementByID("cl1").HasClass("cl1") {
		t.Fatal()
	}
	if !session.GetElementByID("cl2").HasClass("cl1") {
		t.Fatal()
	}
	if session.GetElementByID("cl3").HasClass("cl1") {
		t.Fatal()
	}

	session.Delete()
}
