package prago

import (
	"testing"
)

func TestUser(t *testing.T) {
	test := NewTest(t)
	test.EqualTrue(true)

	r := NewRoute(POST, "/a/:id/:name/aa", nil, []Constraint{})
	params, ok := r.match("POST", "/a/123/ondra/aa")
	test.EqualTrue(ok)
	test.EqualString(params["id"], "123")
	test.EqualString(params["name"], "ondra")
	test.EqualInt(len(params), 2)

	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{})
	_, ok = r.match("POST", "/a/123/ondra/aa")
	test.EqualFalse(ok)
	_, ok = r.match("GET", "/b/123/ondra/aa")
	test.EqualFalse(ok)
	_, ok = r.match("GET", "/a/123/ondra/aa/")
	test.EqualFalse(ok)

	_, ok = r.match("GET", "/a/123/o/aa")
	test.EqualTrue(ok)

	constraint := func(m map[string]string) bool {
		item, ok := m["name"]
		if !ok || len(item) <= 2 {
			return false
		}
		return true
	}

	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	test.EqualTrue(ok)
	_, ok = r.match("GET", "/a/123/o/aa")
	test.EqualFalse(ok)

	constraint = ConstraintInt("id")
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})
	_, ok = r.match("GET", "/a/123/ondra/aa")
	test.EqualTrue(ok)
	_, ok = r.match("GET", "/a/0/ondra/aa")
	test.EqualFalse(ok)
	_, ok = r.match("GET", "/a/123AA/ondra/aa")
	test.EqualFalse(ok)

	constraint = ConstraintWhitelist("name", []string{"ondra", "pepa"})
	r = NewRoute(GET, "/a/:id/:name/aa", nil, []Constraint{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	test.EqualTrue(ok)
	_, ok = r.match("GET", "/a/123/karel/aa")
	test.EqualFalse(ok)

}
