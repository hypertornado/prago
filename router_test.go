package prago

import (
	"context"
	"net/url"
	"testing"
)

func TestRouterNormal(t *testing.T) {
	ctx := context.Background()
	r := newRoute("POST", "/a/:id/:name/aa", nil, nil, nil)
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

	r = newRoute("GET", "/a/:id/:name/aa", nil, nil, nil)
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
		item := values.Get("name")
		if !values.Has("name") || len(item) <= 2 {
			return false
		}
		return true
	}

	r = newRoute("GET", "/a/:id/:name/aa", nil, nil, []routerConstraint{constraint})

	_, ok = r.match(ctx, "GET", "/a/123/ondra/aa")
	if ok != true {
		t.Fatal(ok)
	}
	_, ok = r.match(ctx, "GET", "/a/123/o/aa")
	if ok != false {
		t.Fatal(ok)
	}
}

func TestRouterFallback(t *testing.T) {
	r := newRoute("GET", "*some", nil, nil, nil)
	params, ok := r.match(context.Background(), "GET", "/XXX")
	if ok != true {
		t.Fatal(ok)
	}
	if params.Get("some") != "/XXX" {
		t.Fatal(params["some"])
	}

	r = newRoute("GET", "/a/b/*some", nil, nil, nil)
	params, ok = r.match(context.Background(), "GET", "/a/b/c/d")
	if ok != true {
		t.Fatal(ok)
	}
	if params.Get("some") != "c/d" {
		t.Fatal(params["some"])
	}

}

func TestRouterAny(t *testing.T) {
	r := newRoute("ANY", "/some", nil, nil, nil)
	_, ok := r.match(context.Background(), "POST", "/some")
	if ok != true {
		t.Fatal(ok)
	}

}
