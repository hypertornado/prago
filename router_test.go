package prago

import (
	"testing"
)

func TestUser(t *testing.T) {
	r := NewRoute(POST, "/a/:id/:name/aa", nil, []Constraint{})
	params, ok := r.match("POST", "/a/123/ondra/aa")
	if ok != true {
		t.Error(ok)
	}
	if params["id"] != "123" {
		t.Error(params["id"])
	}
	if params["name"] != "ondra" {
		t.Error(params["name"])
	}
	if len(params) != 2 {
		t.Error(len(params))
	}

	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{})
	_, ok = r.match("POST", "/a/123/ondra/aa")
	if ok != false {
		t.Error(ok)
	}
	_, ok = r.match("GET", "/b/123/ondra/aa")
	if ok != false {
		t.Error(ok)
	}
	_, ok = r.match("GET", "/a/123/ondra/aa/")
	if ok != false {
		t.Error(ok)
	}

	_, ok = r.match("GET", "/a/123/o/aa")
	if ok != true {
		t.Error(ok)
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
		t.Error(ok)
	}
	_, ok = r.match("GET", "/a/123/o/aa")
	if ok != false {
		t.Error(ok)
	}

	constraint = ConstraintInt("id")
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})
	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Error(ok)
	}
	_, ok = r.match("GET", "/a/0/ondra/aa")
	if ok != false {
		t.Error(ok)
	}
	_, ok = r.match("GET", "/a/123AA/ondra/aa")
	if ok != false {
		t.Error(ok)
	}

	constraint = ConstraintWhitelist("name", []string{"ondra", "pepa"})
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Error(ok)
	}
	_, ok = r.match("GET", "/a/123/karel/aa")
	if ok != false {
		t.Error(ok)
	}

}
