package selenium

//https://github.com/SeleniumHQ/selenium/wiki/JsonWireProtocol

//Starting ChromeDriver 2.20.353124 (035346203162d32c80f1dce587c8154a1efa0c3b) on port 9515
//Only local connections are allowed.

//https://code.google.com/p/selenium/wiki/Grid2
//https://github.com/nightwatchjs/nightwatch/wiki/Enable-Firebug-in-Firefox-for-Nightwatch-tests

//"/Applications/Firefox.app/Contents/MacOS/firefox-bin"

import (
	"net/http"
	"testing"
)

var debugApi = false

type Driver struct {
	client               *http.Client
	BaseURL              string
	DesiredCapabilities  Capabilities
	RequiredCapabilities Capabilities
}

func NewDriver(url string) *Driver {
	return &Driver{
		client:               &http.Client{},
		BaseURL:              url,
		RequiredCapabilities: map[string]interface{}{},
		DesiredCapabilities:  map[string]interface{}{},
	}
}

type Capabilities map[string]interface{}

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Secure   bool   `json:"secure"`
	HTTPOnly bool   `json:"httpOnly"`
	Expiry   int64  `json:"expiry"`
}

//TODO: find how to enable timestamp (fix reimportJSON in utils.go )
type LogEntry struct {
	//Timestamp int64  `json:"timestamp"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

type Status struct {
	Build Build
	OS    OS
}

type Build struct {
	Version  string
	Revision string
	Time     string
}

type OS struct {
	Arch    string
	Name    string
	Version string
}

func (d *Driver) Status() (*Status, error) {
	response, err := d.apiRequest("GET", "status", nil)

	if err != nil {
		return nil, err
	}

	status := &Status{}
	err = reimportJSON(response.Value, status)
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (d *Driver) NewTestSession(test *testing.T) *SessionTest {
	session, err := d.NewSession()
	if err != nil {
		test.Fatal(err)
	}
	return &SessionTest{&Test{test}, session}
}

func (d *Driver) NewSession() (*Session, error) {

	reqData := map[string]interface{}{
		"desiredCapabilities":  d.DesiredCapabilities,
		"requiredCapabilities": d.RequiredCapabilities,
	}
	resp, err := d.apiRequest("POST", "session", reqData)
	if err != nil {
		return nil, err
	}

	return &Session{
		SessionID: resp.SessionID,
		driver:    d,
	}, nil

}
