package selenium

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"testing"
)

type size struct {
	Width  int
	Height int
}

type position struct {
	X int
	Y int
}
type location position

func (d *Driver) apiRequest(method, actionUrl string, requestBodyData interface{}) (*SessionResponse, error) {
	var reqStream io.Reader = nil

	if requestBodyData != nil {
		reqData, err := json.Marshal(requestBodyData)
		if err != nil {
			if debugApi {
				panic(err)
			}
			return nil, UnknownError
		}
		reqStream = bytes.NewReader(reqData)
	}

	actionUrl = d.BaseURL + "/" + actionUrl

	if debugApi {
		fmt.Println("---")
		fmt.Println(method, " ", actionUrl)
		fmt.Println("-")

		marshaledBytes, err := json.Marshal(requestBodyData)
		if err != nil {
			panic(err)
		}
		fmt.Println(string(marshaledBytes))
	}

	req, err := http.NewRequest(method, actionUrl, reqStream)
	if err != nil {
		if debugApi {
			panic(err)
		}
		return nil, UnknownError
	}

	resp, err := d.client.Do(req)
	if err != nil {
		if debugApi {
			panic(err)
		}
		return nil, UnknownError
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		if debugApi {
			panic(err)
		}
		return nil, UnknownError
	}

	if debugApi {
		fmt.Println(resp.StatusCode)
		println(string(data))
	}

	err = HTTPErrors(resp.StatusCode)
	if err != nil {
		return nil, err
	}

	sessionResponse := SessionResponse{}
	err = json.Unmarshal(data, &sessionResponse)
	if err != nil {
		return nil, err
	}

	err = ErrorCodeToError(sessionResponse.Status)
	if err == UnknownErrorCode {
		byteData, _ := json.Marshal(sessionResponse.Value)
		err = errors.New(fmt.Sprintf("Unknown error code %d: %s", sessionResponse.Status, byteData))
	}
	if err != nil {
		return nil, err
	}

	return &sessionResponse, nil

}

//http://stackoverflow.com/questions/26744873/converting-map-to-struct
func reimportJSON(inData interface{}, ret interface{}) error {
	data, err := json.Marshal(inData)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &ret)
}

func tGetError(t *testing.T, fn func() error) {
	err := fn()
	if err != nil {
		t.Fatal(err)
	}
}

func tGetString(t *testing.T, fn func() (string, error)) string {
	str, err := fn()
	if err != nil {
		t.Fatal(err)
	}
	return str
}

//SESSION

func (s *Session) request(method, url string, data interface{}) (*SessionResponse, error) {
	slash := ""
	if len(url) > 0 {
		slash = "/"
	}
	actionUrl := fmt.Sprintf("session/%s%s%s", s.SessionID, slash, url)
	ret, err := s.driver.apiRequest(method, actionUrl, data)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func (s *Session) postRequest(url string, data interface{}) error {
	_, err := s.request("POST", url, data)
	return err
}

func (s *Session) postRequestWithReturnValue(url string, data interface{}, returnValue interface{}) error {
	response, err := s.request("POST", url, data)
	if err != nil {
		return err
	}
	return reimportJSON(response.Value, returnValue)
}

func (s *Session) deleteRequest(url string) error {
	_, err := s.request("DELETE", url, nil)
	return err
}

func (s *Session) getValue(url string, returnValue interface{}) (err error) {
	ret, err := s.request("GET", url, nil)
	if err != nil {
		return err
	}
	return reimportJSON(ret.Value, returnValue)
}

//ELEMENTS

func (e *WebElement) postRequest(url string, data interface{}) error {
	return e.session.postRequest(e.actionPath(url), data)
}

func (e *WebElement) postRequestWithReturnValue(url string, data interface{}, returnValue interface{}) error {
	return e.session.postRequestWithReturnValue(e.actionPath(url), data, returnValue)
}

func (e *WebElement) getValue(url string, returnValue interface{}) (err error) {
	return e.session.getValue(e.actionPath(url), returnValue)
}

func (e *WebElement) actionPath(action string) string {
	return fmt.Sprintf("element/%s/%s", e.ELEMENT, action)
}

//WINDOW
func (w *Window) postRequest(url string, data interface{}) error {
	return w.session.postRequest(fmt.Sprintf("window/%s/%s", w.Id, url), data)
}

func (w *Window) getValue(url string, returnValue interface{}) (err error) {
	return w.session.getValue(fmt.Sprintf("window/%s/%s", w.Id, url), returnValue)
}
