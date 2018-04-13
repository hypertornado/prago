package prago

import (
	"github.com/hypertornado/prago/utils"
	"testing"
)

func TestRouterNormal(t *testing.T) {
	r := newRoute(post, "/a/:id/:name/aa", nil, nil, nil)
	params, ok := r.match("POST", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	if params["id"] != "123" {
		t.Fatal(params["id"])
	}
	if params["name"] != "ondra" {
		t.Fatal(params["name"])
	}
	if len(params) != 2 {
		t.Fatal(len(params))
	}

	r = newRoute(get, "/a/:id/:name/aa", nil, nil, nil)
	_, ok = r.match("POST", "/a/123/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/b/123/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123/ondra/aa/")
	if ok != false {
		t.Fatal(ok)
	}

	_, ok = r.match("GET", "/a/123/o/aa")
	if ok != true {
		t.Fatal(ok)
	}

	constraint := func(m map[string]string) bool {
		item, ok := m["name"]
		if !ok || len(item) <= 2 {
			return false
		}
		return true
	}

	r = newRoute(get, "/a/:id/:name/aa", nil, nil, []func(map[string]string) bool{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123/o/aa")
	if ok != false {
		t.Fatal(ok)
	}

	constraint = utils.ConstraintInt("id")
	r = newRoute(get, "/a/:id/:name/aa", nil, nil, []func(map[string]string) bool{constraint})
	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/0/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123AA/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}

	constraint = utils.ConstraintWhitelist("name", []string{"ondra", "pepa"})
	r = newRoute(get, "/a/:id/:name/aa", nil, nil, []func(map[string]string) bool{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123/karel/aa")
	if ok != false {
		t.Fatal(ok)
	}
}

func TestRouterFallback(t *testing.T) {
	r := newRoute(get, "*some", nil, nil, nil)
	params, ok := r.match("GET", "/XXX")
	if ok != true {
		t.Fatal(ok)
	}
	if params["some"] != "/XXX" {
		t.Fatal(params["some"])
	}

	r = newRoute(get, "/a/b/*some", nil, nil, nil)
	params, ok = r.match("GET", "/a/b/c/d")
	if ok != true {
		t.Fatal(ok)
	}
	if params["some"] != "c/d" {
		t.Fatal(params["some"])
	}

}

func TestRouterAny(t *testing.T) {
	r := newRoute(any, "/hello", nil, nil, nil)
	_, ok := r.match("GET", "/hello")
	if ok != true {
		t.Fatal(ok)
	}

	_, ok = r.match("POST", "/hello")
	if ok != true {
		t.Fatal(ok)
	}
}
