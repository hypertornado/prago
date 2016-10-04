package selenium

import (
	"fmt"
)

type Session struct {
	SessionID string `json:"sessionId"`
	driver    *Driver
}

type SessionResponse struct {
	SessionID string      `json:"sessionId"`
	Status    int         `json:"status"`
	Value     interface{} `json:"value"`
}

func (s *Session) GetCapabilities() (*Capabilities, error) {
	var capabilities = &Capabilities{}
	err := s.getValue("", capabilities)
	return capabilities, err
}

func (s *Session) Delete() error {
	return s.deleteRequest("")
}

func (s *Session) SetScriptTimeout(milliseconds uint) error {
	return s.setTimeout("script", milliseconds)
}

func (s *Session) SetImplicitTimeout(milliseconds uint) error {
	return s.setTimeout("implicit", milliseconds)
}

func (s *Session) SetPageLoadTimeout(milliseconds uint) error {
	return s.setTimeout("page load", milliseconds)
}

func (s *Session) setTimeout(timeoutType string, milliseconds uint) error {
	data := map[string]interface{}{
		"type": timeoutType,
		"ms":   milliseconds,
	}
	return s.postRequest("timeouts", data)
}

func (s *Session) SetAsyncScriptTimeout(milliseconds uint) error {
	data := map[string]interface{}{
		"ms": milliseconds,
	}
	return s.postRequest("timeouts/async_script", data)
}

func (s *Session) SetImplicitWaitTimeout(milliseconds uint) error {
	data := map[string]interface{}{
		"ms": milliseconds,
	}
	return s.postRequest("timeouts/implicit_wait", data)
}

func (s *Session) GetCurrentWindow() (window *Window, err error) {
	var id string
	err = s.getValue("window_handle", &id)
	if err != nil {
		return nil, err
	}
	return &Window{id, s}, nil
}

func (s *Session) GetWindows() (windows []*Window, err error) {
	var ids []string
	err = s.getValue("window_handles", &ids)
	if err != nil {
		return
	}
	windows = []*Window{}
	for _, v := range ids {
		windows = append(windows, &Window{v, s})
	}
	return
}

func (s *Session) GetURL() (url string, err error) {
	err = s.getValue("url", &url)
	return
}

func (s *Session) SetURL(url string) error {
	data := map[string]interface{}{
		"url": url,
	}
	return s.postRequest("url", data)
}

func (s *Session) Forward() error {
	return s.postRequest("forward", nil)
}

func (s *Session) Back() error {
	return s.postRequest("back", nil)
}

func (s *Session) Refresh() error {
	return s.postRequest("refresh", nil)
}

func (s *Session) Execute(script string, args []interface{}) (ret interface{}, err error) {
	data := map[string]interface{}{
		"script": script,
		"args":   args,
	}
	err = s.postRequestWithReturnValue("execute", data, &ret)
	return
}

func (s *Session) ExecuteAsync(script string, args []interface{}) (ret interface{}, err error) {
	data := map[string]interface{}{
		"script": script,
		"args":   args,
	}
	err = s.postRequestWithReturnValue("execute_async", data, &ret)
	return
}

func (s *Session) Screenshot() (screenshot string, err error) {
	err = s.getValue("screenshot", &screenshot)
	return
}

func (s *Session) GetImeAvailableEngines() (engines []string, err error) {
	err = s.getValue("ime/available_engines", &engines)
	return
}

func (s *Session) GetImeActiveEngine() (engine string, err error) {
	err = s.getValue("ime/active_engine", &engine)
	return
}

func (s *Session) ImeActivated() (activated bool, err error) {
	err = s.getValue("ime/activated", &activated)
	return
}

func (s *Session) ImeDeactivate() error {
	return s.postRequest("ime/deactivate", nil)
}

func (s *Session) ImeActivate(engine string) error {
	data := map[string]interface{}{
		"engine": engine,
	}
	return s.postRequest("ime/activate", data)
}

func (s *Session) FocusParent() error {
	return s.postRequest("frame/parent", nil)
}

func (s *Session) FocusFrame(id interface{}) error {
	data := map[string]interface{}{
		"id": id,
	}
	return s.postRequest("frame", data)
}

func (s *Session) FocusWindow(name string) error {
	data := map[string]interface{}{
		"name": name,
	}
	return s.postRequest("window", data)
}

func (s *Session) DeleteWindow() error {
	return s.deleteRequest("window")
}

func (s *Session) currentWindow() *Window {
	return &Window{"current", s}
}

func (s *Session) GetCurrentWindowSize() (width, height int, err error) {
	return s.currentWindow().GetSize()
}

func (s *Session) SetCurrentWindowSize(width, height int) error {
	return s.currentWindow().SetSize(width, height)
}

func (s *Session) GetCurrentWindowPosition() (x, y int, err error) {
	return s.currentWindow().GetPosition()
}

func (s *Session) SetCurrentWindowPosition(x, y int) error {
	return s.currentWindow().SetPosition(x, y)
}

func (s *Session) MaximizeCurrentWindow() error {
	return s.currentWindow().Maximize()
}

func (s *Session) GetCookies() (cookies []*Cookie, err error) {
	err = s.getValue("cookie", &cookies)
	return
}

func (s *Session) SetCookie(cookie *Cookie) error {
	data := map[string]interface{}{
		"cookie": cookie,
	}
	return s.postRequest("cookie", data)
}

func (s *Session) DeleteCookies() error {
	return s.deleteRequest("cookie")
}

func (s *Session) DeleteCookie(name string) error {
	return s.deleteRequest(fmt.Sprintf("cookie/%s", name))
}

func (s *Session) GetSource() (source string, err error) {
	err = s.getValue("source", &source)
	return
}

func (s *Session) GetTitle() (ret string, err error) {
	err = s.getValue("title", &ret)
	return
}

func (s *Session) GetElementByClassName(value string) (*WebElement, error) {
	return s.getElementUsing("class name", value)
}

func (s *Session) GetElementByCSSSelector(value string) (*WebElement, error) {
	return s.getElementUsing("css selector", value)
}

func (s *Session) GetElementByID(value string) (*WebElement, error) {
	return s.getElementUsing("id", value)
}

func (s *Session) GetElementByName(value string) (*WebElement, error) {
	return s.getElementUsing("name", value)
}

func (s *Session) GetElementByLinkText(value string) (*WebElement, error) {
	return s.getElementUsing("link text", value)
}

func (s *Session) GetElementByPartialLinkText(value string) (*WebElement, error) {
	return s.getElementUsing("partial link text", value)
}

func (s *Session) GetElementByTagName(value string) (*WebElement, error) {
	return s.getElementUsing("tag name", value)
}

func (s *Session) GetElementByXPath(value string) (*WebElement, error) {
	return s.getElementUsing("xpath", value)
}

func (s *Session) getElementUsing(using, value string) (element *WebElement, err error) {
	data := map[string]string{
		"using": using,
		"value": value,
	}
	err = s.postRequestWithReturnValue("element", data, &element)
	if err == nil {
		element.session = s
	}
	return
}

func (s *Session) GetElementsByClassName(value string) ([]*WebElement, error) {
	return s.getElementsUsing("class name", value)
}

func (s *Session) GetElementsByCSSSelector(value string) ([]*WebElement, error) {
	return s.getElementsUsing("css selector", value)
}

func (s *Session) GetElementsByID(value string) ([]*WebElement, error) {
	return s.getElementsUsing("id", value)
}

func (s *Session) GetElementsByName(value string) ([]*WebElement, error) {
	return s.getElementsUsing("name", value)
}

func (s *Session) GetElementsByLinkText(value string) ([]*WebElement, error) {
	return s.getElementsUsing("link text", value)
}

func (s *Session) GetElementsByPartialLinkText(value string) ([]*WebElement, error) {
	return s.getElementsUsing("partial link text", value)
}

func (s *Session) GetElementsByTagName(value string) ([]*WebElement, error) {
	return s.getElementsUsing("tag name", value)
}

func (s *Session) GetElementsByXPath(value string) ([]*WebElement, error) {
	return s.getElementsUsing("xpath", value)
}

func (s *Session) getElementsUsing(using, value string) (elements []*WebElement, err error) {
	data := map[string]string{
		"using": using,
		"value": value,
	}
	err = s.postRequestWithReturnValue("elements", data, &elements)
	if err == nil {
		for _, v := range elements {
			v.session = s
		}
	}
	return
}

func (s *Session) GetActiveElement() (element *WebElement, err error) {
	err = s.postRequestWithReturnValue("element/active", nil, &element)
	if err == nil {
		element.session = s
	}
	return
}

func (s *Session) GetOrientation() (orientation string, err error) {
	err = s.getValue("orientation", &orientation)
	return
}

func (s *Session) SetOrientation(orientation string) error {
	data := map[string]string{
		"orientation": orientation,
	}
	return s.postRequest("orientation", data)
}

func (s *Session) GetAlertText() (alertText string, err error) {
	err = s.getValue("alert_text", &alertText)
	return
}

func (s *Session) SetAlertText(text string) error {
	data := map[string]interface{}{
		"text": text,
	}
	return s.postRequest("alert_text", data)
}

func (s *Session) AcceptAlert() error  { return s.postRequest("accept_alert", nil) }
func (s *Session) DismissAlert() error { return s.postRequest("dismiss_alert", nil) }

func (s *Session) MoveToRelative(x, y int) error {
	data := map[string]interface{}{
		"xoffset": x,
		"yoffset": y,
	}
	return s.postRequest("moveto", data)
}

func (s *Session) Click() error             { return s.click(0) }
func (s *Session) LeftButtonClick() error   { return s.click(0) }
func (s *Session) MiddleButtonClick() error { return s.click(1) }
func (s *Session) RightButtonClick() error  { return s.click(2) }
func (s *Session) click(b int) error        { return s.buttonAction("click", b) }
func (s *Session) LeftButtonDown() error    { return s.buttonDown(0) }
func (s *Session) MiddleButtonDown() error  { return s.buttonDown(1) }
func (s *Session) RightButtonDown() error   { return s.buttonDown(2) }
func (s *Session) buttonDown(b int) error   { return s.buttonAction("buttondown", b) }
func (s *Session) LeftButtonUp() error      { return s.buttonUp(0) }
func (s *Session) MiddleButtonUp() error    { return s.buttonUp(1) }
func (s *Session) RightButtonUp() error     { return s.buttonUp(2) }
func (s *Session) buttonUp(b int) error     { return s.buttonAction("buttonup", b) }

func (s *Session) buttonAction(action string, b int) error {
	data := map[string]interface{}{
		"button": b,
	}
	return s.postRequest(action, data)
}

func (s *Session) DoubleClick() error {
	return s.postRequest("doubleclick", nil)
}

func (s *Session) SendKeys(sequence string) error {
	keys := make([]string, len(sequence))
	for i, k := range sequence {
		keys[i] = string(k)
	}
	data := map[string]interface{}{
		"value": keys,
	}
	return s.postRequest("keys", data)
}

func (s *Session) TouchDown(x, y int) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	return s.postRequest("touch/down", data)
}

func (s *Session) TouchUp(x, y int) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	return s.postRequest("touch/up", data)
}

func (s *Session) TouchMove(x, y int) error {
	data := map[string]interface{}{
		"x": x,
		"y": y,
	}
	return s.postRequest("touch/move", data)
}

func (s *Session) TouchScroll(xoffset, yoffset int) error {
	data := map[string]interface{}{
		"xoffset": xoffset,
		"yoffset": yoffset,
	}
	return s.postRequest("touch/scroll", data)
}

func (s *Session) TouchFlick(xspeed, yspeed int) error {
	data := map[string]interface{}{
		"xspeed": xspeed,
		"yspeed": yspeed,
	}
	return s.postRequest("touch/flick", data)
}

type geolocation struct {
	Latitude  float64
	Longitude float64
	Altitude  float64
}

func (s *Session) GetLocation() (latitude, longitude, altitude float64, err error) {
	var geolocation geolocation
	err = s.getValue("location", &geolocation)
	if err != nil {
		return -1, -1, -1, err
	}
	return geolocation.Latitude, geolocation.Longitude, geolocation.Altitude, nil
}

func (s *Session) SetLocation(latitude, longitude, altitude float64) error {
	data := map[string]interface{}{
		"latitude":  latitude,
		"longitude": longitude,
		"altitude":  altitude,
	}
	return s.postRequest("location", data)
}

func (s *Session) GetLocalStorage() (keys []string, err error) {
	err = s.getValue("local_storage", &keys)
	return
}
func (s *Session) SetLocalStorage(key, value string) error {
	data := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	return s.postRequest("local_storage", data)
}
func (s *Session) DeleteLocalStorage() error {
	return s.deleteRequest("local_storage")
}
func (s *Session) GetLocalStorageValue(key string) (value string, err error) {
	err = s.getValue(fmt.Sprintf("local_storage/key/%s", key), &value)
	return
}
func (s *Session) DeleteLocalStorageValue(key string) (err error) {
	return s.deleteRequest(fmt.Sprintf("local_storage/key/%s", key))
}
func (s *Session) GetLocalStorageSize() (size int, err error) {
	err = s.getValue("local_storage/size", &size)
	return
}

func (s *Session) GetSessionStorage() (keys []string, err error) {
	err = s.getValue("session_storage", &keys)
	return
}
func (s *Session) SetSessionStorage(key, value string) error {
	data := map[string]interface{}{
		"key":   key,
		"value": value,
	}
	return s.postRequest("session_storage", data)
}
func (s *Session) DeleteSessionStorage() error {
	return s.deleteRequest("session_storage")
}
func (s *Session) GetSessionStorageValue(key string) (value string, err error) {
	err = s.getValue(fmt.Sprintf("session_storage/key/%s", key), &value)
	return
}
func (s *Session) DeleteSessionStorageValue(key string) (err error) {
	return s.deleteRequest(fmt.Sprintf("session_storage/key/%s", key))
}
func (s *Session) GetSessionStorageSize() (size int, err error) {
	err = s.getValue("session_storage/size", &size)
	return
}

func (s *Session) GetLog(logType string) (entries []LogEntry, err error) {
	data := map[string]interface{}{
		"type": logType,
	}
	err = s.postRequestWithReturnValue("log", data, &entries)
	return
}

func (s *Session) GetLogTypes() (types []string, err error) {
	err = s.getValue("log/types", &types)
	return
}
