package selenium

import ()

type SessionTest struct {
	t       *Test
	Session *Session
}

func (s *SessionTest) GetCapabilities() (c *Capabilities) {
	c, err := s.Session.GetCapabilities()
	s.t.err(err)
	return
}

func (s *SessionTest) Delete() { s.t.err(s.Session.Delete()) }
func (s *SessionTest) SetScriptTimeout(milliseconds uint) {
	s.t.err(s.Session.SetScriptTimeout(milliseconds))
}
func (s *SessionTest) SetImplicitTimeout(milliseconds uint) {
	s.t.err(s.Session.SetImplicitTimeout(milliseconds))
}
func (s *SessionTest) SetPageLoadTimeout(milliseconds uint) {
	s.t.err(s.Session.SetPageLoadTimeout(milliseconds))
}
func (s *SessionTest) SetAsyncScriptTimeout(milliseconds uint) {
	s.t.err(s.Session.SetAsyncScriptTimeout(milliseconds))
}
func (s *SessionTest) SetImplicitWaitTimeout(milliseconds uint) {
	s.t.err(s.Session.SetImplicitWaitTimeout(milliseconds))
}

func (s *SessionTest) GetCurrentWindow() *WindowTest {
	window, err := s.Session.GetCurrentWindow()
	s.t.err(err)
	return &WindowTest{s.t, window}
}

func (s *SessionTest) GetWindows() []*WindowTest {
	windows, err := s.Session.GetWindows()
	s.t.err(err)
	ret := []*WindowTest{}
	for _, v := range windows {
		ret = append(ret, &WindowTest{s.t, v})
	}
	return ret
}

func (s *SessionTest) GetURL() string    { return s.t.getStr(s.Session.GetURL()) }
func (s *SessionTest) SetURL(url string) { s.t.err(s.Session.SetURL(url)) }
func (s *SessionTest) Forward()          { s.t.err(s.Session.Forward()) }
func (s *SessionTest) Back()             { s.t.err(s.Session.Back()) }
func (s *SessionTest) Refresh()          { s.t.err(s.Session.Refresh()) }

func (s *SessionTest) Execute(script string, args []interface{}) (ret interface{}) {
	return s.t.getIface(s.Session.Execute(script, args))
}

func (s *SessionTest) ExecuteAsync(script string, args []interface{}) (ret interface{}) {
	return s.t.getIface(s.Session.ExecuteAsync(script, args))
}

func (s *SessionTest) Screenshot() (screenshot string) { return s.t.getStr(s.Session.Screenshot()) }

func (s *SessionTest) GetImeAvailableEngines() (engines []string) {
	engines, err := s.Session.GetImeAvailableEngines()
	s.t.err(err)
	return
}

func (s *SessionTest) GetImeActiveEngine() (engine string) {
	return s.t.getStr(s.Session.GetImeActiveEngine())
}
func (s *SessionTest) ImeActivated() (activated bool) {
	return s.t.getBool(s.Session.ImeActivated())
}
func (s *SessionTest) ImeDeactivate()            { s.t.err(s.Session.ImeDeactivate()) }
func (s *SessionTest) ImeActivate(engine string) { s.t.err(s.Session.ImeActivate(engine)) }

func (s *SessionTest) FocusParent()              { s.t.err(s.Session.FocusParent()) }
func (s *SessionTest) FocusFrame(id interface{}) { s.t.err(s.Session.FocusFrame(id)) }
func (s *SessionTest) FocusWindow(name string)   { s.t.err(s.Session.FocusWindow(name)) }
func (s *SessionTest) DeleteWindow()             { s.t.err(s.Session.DeleteWindow()) }

func (s *SessionTest) GetCurrentWindowSize() (width, height int) {
	return s.t.getInts(s.Session.GetCurrentWindowSize())
}

func (s *SessionTest) SetCurrentWindowSize(width, height int) {
	s.t.err(s.Session.SetCurrentWindowSize(width, height))
}

func (s *SessionTest) GetCurrentWindowPosition() (x, y int) {
	return s.t.getInts(s.Session.GetCurrentWindowPosition())
}

func (s *SessionTest) SetCurrentWindowPosition(x, y int) {
	s.t.err(s.Session.SetCurrentWindowPosition(x, y))
}

func (s *SessionTest) MaximizeCurrentWindow() { s.t.err(s.Session.MaximizeCurrentWindow()) }

func (s *SessionTest) GetCookies() (cookies []*Cookie) {
	cookies, err := s.Session.GetCookies()
	s.t.err(err)
	return
}

func (s *SessionTest) SetCookie(cookie *Cookie) { s.t.err(s.Session.SetCookie(cookie)) }
func (s *SessionTest) DeleteCookies()           { s.t.err(s.Session.DeleteCookies()) }
func (s *SessionTest) DeleteCookie(name string) { s.t.err(s.Session.DeleteCookie(name)) }
func (s *SessionTest) GetSource() string        { return s.t.getStr(s.Session.GetSource()) }
func (s *SessionTest) GetTitle() string         { return s.t.getStr(s.Session.GetTitle()) }

func (s *SessionTest) GetElementByClassName(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByClassName(value))
}
func (s *SessionTest) GetElementByCSSSelector(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByCSSSelector(value))
}
func (s *SessionTest) GetElementByID(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByID(value))
}
func (s *SessionTest) GetElementByName(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByName(value))
}
func (s *SessionTest) GetElementByLinkText(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByLinkText(value))
}
func (s *SessionTest) GetElementByPartialLinkText(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByPartialLinkText(value))
}
func (s *SessionTest) GetElementByTagName(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByTagName(value))
}
func (s *SessionTest) GetElementByXPath(value string) *WebElementTest {
	return s.t.getEl(s.Session.GetElementByXPath(value))
}

func (s *SessionTest) GetElementsByClassName(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByClassName(value))
}
func (s *SessionTest) GetElementsByCSSSelector(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByCSSSelector(value))
}
func (s *SessionTest) GetElementsByID(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByID(value))
}
func (s *SessionTest) GetElementsByName(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByName(value))
}
func (s *SessionTest) GetElementsByLinkText(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByLinkText(value))
}
func (s *SessionTest) GetElementsByPartialLinkText(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByPartialLinkText(value))
}
func (s *SessionTest) GetElementsByTagName(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByTagName(value))
}
func (s *SessionTest) GetElementsByXPath(value string) []*WebElementTest {
	return s.t.getEls(s.Session.GetElementsByXPath(value))
}

func (s *SessionTest) GetActiveElement() *WebElementTest {
	return s.t.getEl(s.Session.GetActiveElement())
}

func (s *SessionTest) GetOrientation() string {
	return s.t.getStr(s.Session.GetOrientation())
}

func (s *SessionTest) SetOrientation(orientation string) {
	s.t.err(s.Session.SetOrientation(orientation))
}

func (s *SessionTest) GetAlertText() string {
	return s.t.getStr(s.Session.GetAlertText())
}

func (s *SessionTest) SetAlertText(text string) {
	s.t.err(s.Session.SetAlertText(text))
}

func (s *SessionTest) AcceptAlert()            { s.t.err(s.Session.AcceptAlert()) }
func (s *SessionTest) DismissAlert()           { s.t.err(s.Session.DismissAlert()) }
func (s *SessionTest) MoveToRelative(x, y int) { s.t.err(s.Session.MoveToRelative(x, y)) }

func (s *SessionTest) Click()             { s.t.err(s.Session.Click()) }
func (s *SessionTest) LeftButtonClick()   { s.t.err(s.Session.LeftButtonClick()) }
func (s *SessionTest) MiddleButtonClick() { s.t.err(s.Session.MiddleButtonClick()) }
func (s *SessionTest) RightButtonClick()  { s.t.err(s.Session.RightButtonClick()) }
func (s *SessionTest) LeftButtonDown()    { s.t.err(s.Session.LeftButtonDown()) }
func (s *SessionTest) MiddleButtonDown()  { s.t.err(s.Session.MiddleButtonDown()) }
func (s *SessionTest) RightButtonDown()   { s.t.err(s.Session.RightButtonDown()) }
func (s *SessionTest) LeftButtonUp()      { s.t.err(s.Session.LeftButtonUp()) }
func (s *SessionTest) MiddleButtonUp()    { s.t.err(s.Session.MiddleButtonUp()) }
func (s *SessionTest) RightButtonUp()     { s.t.err(s.Session.RightButtonUp()) }
func (s *SessionTest) DoubleClick()       { s.t.err(s.Session.DoubleClick()) }

func (s *SessionTest) SendKeys(sequence string) { s.t.err(s.Session.SendKeys(sequence)) }

func (s *SessionTest) TouchDown(x, y int) { s.t.err(s.Session.TouchDown(x, y)) }
func (s *SessionTest) TouchUp(x, y int)   { s.t.err(s.Session.TouchUp(x, y)) }
func (s *SessionTest) TouchMove(x, y int) { s.t.err(s.Session.TouchMove(x, y)) }
func (s *SessionTest) TouchScroll(xoffset, yoffset int) {
	s.t.err(s.Session.TouchScroll(xoffset, yoffset))
}
func (s *SessionTest) TouchFlick(xspeed, yspeed int) {
	s.t.err(s.Session.TouchFlick(xspeed, yspeed))
}

func (s *SessionTest) GetLocation() (latitude, longitude, altitude float64) {
	latitude, longitude, altitude, err := s.Session.GetLocation()
	if err != nil {
		s.t.err(err)
	}
	return
}

func (s *SessionTest) SetLocation(latitude, longitude, altitude float64) {
	s.t.err(s.Session.SetLocation(latitude, longitude, altitude))
}

func (s *SessionTest) GetLocalStorage() (keys []string) {
	keys, err := s.Session.GetLocalStorage()
	s.t.err(err)
	return
}
func (s *SessionTest) SetLocalStorage(key, value string) {
	s.t.err(s.Session.SetLocalStorage(key, value))
}
func (s *SessionTest) DeleteLocalStorage() { s.t.err(s.Session.DeleteLocalStorage()) }
func (s *SessionTest) GetLocalStorageValue(key string) (value string) {
	return s.t.getStr(s.Session.GetLocalStorageValue(key))
}
func (s *SessionTest) DeleteLocalStorageValue(key string) {
	s.t.err(s.Session.DeleteLocalStorageValue(key))
}
func (s *SessionTest) GetLocalStorageSize() (size int) {
	return s.t.getInt(s.Session.GetLocalStorageSize())
}

func (s *SessionTest) GetSessionStorage() (keys []string) {
	keys, err := s.Session.GetSessionStorage()
	s.t.err(err)
	return
}
func (s *SessionTest) SetSessionStorage(key, value string) {
	s.t.err(s.Session.SetSessionStorage(key, value))
}
func (s *SessionTest) DeleteSessionStorage() { s.t.err(s.Session.DeleteSessionStorage()) }
func (s *SessionTest) GetSessionStorageValue(key string) (value string) {
	return s.t.getStr(s.Session.GetSessionStorageValue(key))
}
func (s *SessionTest) DeleteSessionStorageValue(key string) {
	s.t.err(s.Session.DeleteSessionStorageValue(key))
}
func (s *SessionTest) GetSessionStorageSize() (size int) {
	return s.t.getInt(s.Session.GetSessionStorageSize())
}

func (s *SessionTest) GetLog(logType string) (entries []LogEntry) {
	entries, err := s.Session.GetLog(logType)
	s.t.err(err)
	return
}
func (s *SessionTest) GetLogTypes() (types []string) {
	types, err := s.Session.GetLogTypes()
	s.t.err(err)
	return
}
