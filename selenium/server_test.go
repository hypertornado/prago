package selenium

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func testHandler(w http.ResponseWriter, r *http.Request) {
	content := map[string]string{
		"/a": `<html><a href="/b">linka</a></html>`,
		"/b": `<html><a href="/a">linkb</a></html>`,
		"/test": `
<html>
<head>
<title>foo</title>
</head>

<script>console.log("12");</script>

<div id="id1" class="baz">bars</div>
<div class="baz">bazs</div>
<div style="color: red;">
	<div class="foo" style="width: 20px; height: 30px; position: absolute; top: 2px; left: 3px;" data-id="x">foos</div>
</div>

<a href="" class="link" id="linkid" name="ipsum">lorem</a>

<div class="x">A</div>
<div class="container">
	<div class="x">B</div>
</div>

<button id="btn" onclick="this.textContent='clicked';" ondblclick="this.textContent='dblclicked';"></button>

<textarea id="textarea">lorem</textarea>

<div id="cl1" class="abc    cl1  cde"></div>
<div id="cl2" class="abc cl1"></div>
<div id="cl3" class="abc cde"></div>

</html>
    `,
	}

	data, ok := content[r.URL.Path]

	if r.URL.Path == "/time" {
		ok = true
		data = strconv.FormatInt(time.Now().UnixNano(), 10)
	}

	if !ok {
		w.WriteHeader(404)
		fmt.Fprint(w, "not found "+r.URL.Path)
		return
	}

	fmt.Fprint(w, data)
}

var testServerRunning = false

func StartTestServer() error {
	if testServerRunning == true {
		return nil
	}

	http.HandleFunc("/", testHandler)
	go http.ListenAndServe(":8587", nil)

	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		resp, _ := http.Get("http://localhost:8587/test")
		if resp.StatusCode == 200 {
			testServerRunning = true
			return nil
		}
	}

	return errors.New("timeout error")

}
