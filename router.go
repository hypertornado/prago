package prago

import (
	"regexp"
	"strings"
)

type method int

const (
	GET method = iota
	HEAD
	POST
	PUT
	DELETE
)

func MiddlewareDispatcher(p Request) {
	if p.IsProcessed() {
		return
	}
	p.App().Router().Process(p)
}

type Router struct {
	routes []*Route
}

func NewRouter() *Router {
	return &Router{[]*Route{}}
}

func (r *Router) AddRoute(route *Route) {
	r.routes = append(r.routes, route)
}

func (router *Router) Process(request Request) {
	_, r := request.HttpIO()
	for _, route := range router.routes {
		params, match := route.match(r.Method, r.URL.Path)
		if match {
			for k, v := range params {
				request.Params().Add(k, v)
			}
			route.action.call(request)
			return
		}
	}
}

type Route struct {
	items       []string
	method      string
	constraints []Constraint
	action      *Action
}

func NewRoute(m method, str string, action *Action, constraints []Constraint) *Route {
	items := strings.Split(str, "/")

	methodName := map[method]string{
		GET:    "GET",
		HEAD:   "HEAD",
		POST:   "POST",
		PUT:    "PUT",
		DELETE: "DELETE",
	}

	return &Route{items, methodName[m], constraints, action}
}

type Constraint func(map[string]string) bool

func ConstraintInt(item string) func(map[string]string) bool {
	reg, _ := regexp.Compile("^[1-9][0-9]*$")
	f := ConstraintRegexp(item, reg)
	return f
}

func ConstraintWhitelist(item string, allowedValues []string) func(map[string]string) bool {
	allowedMap := make(map[string]bool)
	for _, v := range allowedValues {
		allowedMap[v] = true
	}
	return func(m map[string]string) bool {
		if value, ok := m[item]; ok {
			return allowedMap[value]
		} else {
			return false
		}
	}
}

func ConstraintRegexp(item string, reg *regexp.Regexp) func(map[string]string) bool {
	return func(m map[string]string) bool {
		if value, ok := m[item]; ok {
			return reg.Match([]byte(value))
		} else {
			return false
		}
	}
}

func (r *Route) match(method, path string) (ret map[string]string, ok bool) {
	ok = false
	if len(r.method) > 0 && r.method != method {
		return
	}

	if !strings.HasPrefix(path, "/") {
		return
	}
	items := strings.Split(path, "/")

	if len(items) != len(r.items) {
		return
	}
	m := make(map[string]string)
	for i := 0; i < len(items); i++ {
		expect := r.items[i]
		if len(expect) > 1 && strings.HasPrefix(expect, ":") {
			m[expect[1:]] = items[i]
		} else {
			if expect != items[i] {
				return
			}
		}
	}

	for _, constraint := range r.constraints {
		ok = constraint(m)
		if ok != true {
			return ret, false
		}
	}

	ret = m
	return ret, true
}
