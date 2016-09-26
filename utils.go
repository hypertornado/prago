package prago

import (
	"encoding/json"
	"net/http"
)

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func Redirect(request Request, urlStr string) {
	request.Header().Set("Location", urlStr)
	request.Response().WriteHeader(http.StatusFound)
	request.SetProcessed()
}

func WriteAPI(r Request, data interface{}, code int) {
	r.SetProcessed()

	r.Response().Header().Add("Content-type", "application/json")

	pretty := false
	if r.Params().Get("pretty") == "true" {
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
	r.Response().WriteHeader(code)
	r.Response().Write(result)
}
