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
	method      string
	constraints []Constraint
	action      *Action
	pathMatcher pathMatcherFn
}

type pathMatcherFn func(string) (map[string]string, bool)

func matcherBasic(route string) pathMatcherFn {
	routeItems := strings.Split(route, "/")

	return func(path string) (m map[string]string, ok bool) {
		items := strings.Split(path, "/")
		m = make(map[string]string)

		if len(items) != len(routeItems) {
			return
		}

		for i := 0; i < len(items); i++ {
			expect := routeItems[i]
			if len(expect) > 1 && strings.HasPrefix(expect, ":") {
				m[expect[1:]] = items[i]
			} else {
				if expect != items[i] {
					return
				}
			}
		}
		return m, true
	}
}

func matcherStar(route string) pathMatcherFn {
	if !strings.HasPrefix(route, "*") {
		return nil
	}
	routeName := route[1:]
	return func(path string) (m map[string]string, ok bool) {
		m = make(map[string]string)
		if len(routeName) > 0 {
			m[routeName] = path
		}
		return m, true
	}
}

func matcherStarMiddle(route string) pathMatcherFn {
	starIndex := strings.Index(route, "/*")
	if starIndex <= 0 {
		return nil
	}
	prefix := route[0 : starIndex+1]
	routeName := route[starIndex+2:]
	return func(path string) (m map[string]string, ok bool) {
		if !strings.HasPrefix(path, prefix) {
			return nil, false
		}
		m = make(map[string]string)
		if len(routeName) > 0 {
			m[routeName] = path[starIndex+1:]
		}
		return m, true
	}
}

func NewRoute(m method, path string, action *Action, constraints []Constraint) (route *Route) {
	methodName := map[method]string{
		GET:    "GET",
		HEAD:   "HEAD",
		POST:   "POST",
		PUT:    "PUT",
		DELETE: "DELETE",
		ANY:    "ANY",
	}

	route = &Route{
		method:      methodName[m],
		constraints: constraints,
		action:      action,
	}

	for _, v := range []func(string) pathMatcherFn{matcherStar, matcherStarMiddle, matcherBasic} {
		if route.pathMatcher != nil {
			break
		}
		route.pathMatcher = v(path)
	}
	return
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

func (r *Route) match(method, path string) (map[string]string, bool) {
	if !methodMatch(r.method, method) {
		return nil, false
	}

	if !strings.HasPrefix(path, "/") {
		return nil, false
	}

	m, ok := r.pathMatcher(path)
	if !ok {
		return nil, false
	}

	for _, constraint := range r.constraints {
		ok = constraint(m)
		if ok != true {
			return nil, false
		}
	}

	return m, true
}

func methodMatch(m1, m2 string) bool {
	if m1 != "ANY" && len(m1) > 0 && m1 != m2 {
		return false
	}
	return true
}
