package prago

import (
	"encoding/json"
	"net/http"
)

//Must panics when error is not nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

//Redirect redirects request to new url
func Redirect(request Request, url string) {
	request.Header().Set("Location", url)
	request.Response().WriteHeader(http.StatusFound)
}

//WriteAPI writes data as JSON response to request with http code
func WriteAPI(request Request, data interface{}, code int) {
	request.Response().Header().Add("Content-type", "application/json")

	pretty := false
	if request.Params().Get("pretty") == "true" {
		pretty = true
	}

	var responseToWrite interface{}
	if code >= 400 {
		responseToWrite = map[string]interface{}{"error": data, "errorCode": code}
	} else {
		responseToWrite = data
	}

	var result []byte
	var e error

	if pretty == true {
		result, e = json.MarshalIndent(responseToWrite, "", "  ")
	} else {
		result, e = json.Marshal(responseToWrite)
	}

	if e != nil {
		panic("error while generating JSON output")
	}
	request.Response().WriteHeader(code)
	request.Response().Write(result)
}
