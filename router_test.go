package prago

import (
	"testing"
)

func TestRouterNormal(t *testing.T) {
	r := NewRoute(POST, "/a/:id/:name/aa", nil, []Constraint{})
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

	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{})
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

	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123/o/aa")
	if ok != false {
		t.Fatal(ok)
	}

	constraint = ConstraintInt("id")
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})
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

	constraint = ConstraintWhitelist("name", []string{"ondra", "pepa"})
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})

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
	r := NewRoute(GET, "*some", nil, []Constraint{})
	params, ok := r.match("GET", "/XXX")
	if ok != true {
		t.Fatal(ok)
	}
	if params["some"] != "/XXX" {
		t.Fatal(params["some"])
	}

	r = NewRoute(GET, "/a/b/*some", nil, []Constraint{})
	params, ok = r.match("GET", "/a/b/c/d")
	if ok != true {
		t.Fatal(ok)
	}
	if params["some"] != "c/d" {
		t.Fatal(params["some"])
	}

}

func TestRouterAny(t *testing.T) {
	r := NewRoute(ANY, "/hello", nil, []Constraint{})
	_, ok := r.match("GET", "/hello")
	if ok != true {
		t.Fatal(ok)
	}

	_, ok = r.match("POST", "/hello")
	if ok != true {
		t.Fatal(ok)
	}
}
