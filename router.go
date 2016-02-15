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
	ANY
)

func requestMiddlewareDispatcher(p Request, next func()) {
	if p.IsProcessed() {
		return
	}

	router := p.App().data["router"].(*Router)
	if router == nil {
		panic("couldnt find router")
	}
	router.Process(p)

	next()
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
	for _, route := range router.routes {
		params, match := route.match(request.Request().Method, request.Request().URL.Path)
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
		ANY:    "ANY",
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
	if r.method != "ANY" && len(r.method) > 0 && r.method != method {
		return
	}

	if !strings.HasPrefix(path, "/") {
		return
	}
	items := strings.Split(path, "/")
	m := make(map[string]string)

	if len(r.items) == 1 && strings.HasPrefix(r.items[0], "*") {
		ok = true
		if len(r.items[0]) > 1 {
			m[r.items[0][1:]] = path
			ret = m
		}
		return
	}

	if len(items) != len(r.items) {
		return
	}

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
