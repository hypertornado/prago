package prago

import (
	"context"
	"net/url"
	"testing"
)

func TestRouterNormal(t *testing.T) {
	ctx := context.Background()
	r := newRoute(post, "/a/:id/:name/aa", nil, nil, nil)
	params, ok := r.match(ctx, "POST", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	if params.Get("id") != "123" {
		t.Fatal(params.Get("id"))
	}
	if params.Get("name") != "ondra" {
		t.Fatal(params.Get("name"))
	}
	if len(params) != 2 {
		t.Fatal(len(params))
	}

	r = newRoute(get, "/a/:id/:name/aa", nil, nil, nil)
	_, ok = r.match(ctx, "POST", "/a/123/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}
	_, ok = r.match(ctx, "GET", "/b/123/ondra/aa")
	if ok != false {
		t.Fatal(ok)
	}
	_, ok = r.match(ctx, "GET", "/a/123/ondra/aa/")
	if ok != false {
		t.Fatal(ok)
	}

	_, ok = r.match(ctx, "GET", "/a/123/o/aa")
	if ok != true {
		t.Fatal(ok)
	}

	constraint := func(ctx context.Context, values url.Values) bool {
		item, ok := values["name"]
		if !ok || len(item) <= 2 {
			return false
		}
		return true
	}

	r = newRoute(get, "/a/:id/:name/aa", nil, nil, []routerConstraint{constraint})

	_, ok = r.match(ctx, "GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match(ctx, "GET", "/a/123/o/aa")
	if ok != false {
		t.Fatal(ok)
	}

	/*constraint = constraintInt("id")
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

	constraint = constraintWhitelist("name", []string{"ondra", "pepa"})
	r = newRoute(get, "/a/:id/:name/aa", nil, nil, []func(map[string]string) bool{constraint})

	_, ok = r.match("GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match("GET", "/a/123/karel/aa")
	if ok != false {
		t.Fatal(ok)
	}*/
}

func TestRouterFallback(t *testing.T) {
	r := newRoute(get, "*some", nil, nil, nil)
	params, ok := r.match(context.Background(), "GET", "/XXX")
	if ok != true {
		t.Fatal(ok)
	}
	if params.Get("some") != "/XXX" {
		t.Fatal(params["some"])
	}

	r = newRoute(get, "/a/b/*some", nil, nil, nil)
	params, ok = r.match(context.Background(), "GET", "/a/b/c/d")
	if ok != true {
		t.Fatal(ok)
	}
	if params.Get("some") != "c/d" {
		t.Fatal(params["some"])
	}

}
