package prago

import (
	"fmt"
	"regexp"
	"strings"
)

type method int

const (
	get method = iota
	head
	post
	put
	del
	any
)

type MiddlewareDispatcher struct {
	router *router
}

func (m MiddlewareDispatcher) Init(app *App) error {
	m.router = newRouter()
	app.data["router"] = m.router
	app.requestMiddlewares = append(app.requestMiddlewares, m.requestMiddlewareDispatcher)

	routerCommand := app.CreateCommand("routes", "Show routes")
	app.AddCommand(routerCommand, func(app *App) error {
		m.router.print()
		return nil
	})

	return nil
}

func (m MiddlewareDispatcher) requestMiddlewareDispatcher(p Request, next func()) {
	if p.IsProcessed() {
		return
	}

	m.router.process(p)

	next()
}

type router struct {
	routes []*route
}

func newRouter() *router {
	return &router{[]*route{}}
}

func (r *router) addRoute(route *route) {
	r.routes = append(r.routes, route)
}

func (r *router) process(request Request) {
	for _, route := range r.routes {
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

func (r *router) print() {
	for _, v := range r.routes {
		fmt.Printf("%s %s\n", v.method, v.path)
	}
}

type route struct {
	method      string
	path        string
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

func newRoute(m method, path string, action *Action, constraints []Constraint) (ret *route) {
	methodName := map[method]string{
		get:  "GET",
		head: "HEAD",
		post: "POST",
		put:  "PUT",
		del:  "DELETE",
		any:  "ANY",
	}

	ret = &route{
		method:      methodName[m],
		path:        path,
		constraints: constraints,
		action:      action,
	}

	for _, v := range []func(string) pathMatcherFn{matcherStar, matcherStarMiddle, matcherBasic} {
		if ret.pathMatcher != nil {
			break
		}
		ret.pathMatcher = v(path)
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

func (r *route) match(method, path string) (map[string]string, bool) {
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
