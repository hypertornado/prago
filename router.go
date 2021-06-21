package prago

import (
	"strings"
)

type method int

const (
	get method = iota
	head
	post
	put
	del
)

type router struct {
	priorityRoutes []*route
	routes         []*route
}

func newRouter() *router {
	return &router{
		[]*route{},
		[]*route{},
	}
}

func (r *router) addRoute(route *route) {
	if route.controller.priorityRouter {
		r.priorityRoutes = append(r.priorityRoutes, route)
	} else {
		r.routes = append(r.routes, route)
	}
}

func (r *router) process(request *Request) bool {
	for _, routes := range [][]*route{
		r.priorityRoutes,
		r.routes,
	} {
		for _, route := range routes {
			params, match := route.match(request.Request().Method, request.Request().URL.Path)
			if match {
				for k, v := range params {
					request.Params().Add(k, v)
				}

				route.controller.callArounds(request, 0, func() {
					route.fn(request)
				}, true)

				return true
			}
		}
	}
	return false
}

func (r *router) export() (ret [][2]string) {
	ret = append(ret, [2]string{"PRIORITY ROUTES", ""})
	for _, v := range r.priorityRoutes {
		ret = append(ret, [2]string{v.method, v.path})
	}

	ret = append(ret, [2]string{"NORMAL ROUTES", ""})
	for _, v := range r.routes {
		ret = append(ret, [2]string{v.method, v.path})
	}
	return
}

type route struct {
	method      string
	path        string
	constraints []func(map[string]string) bool
	pathMatcher pathMatcherFn
	controller  *controller
	fn          func(p *Request)
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

func newRoute(m method, path string, controller *controller, fn func(p *Request), constraints []func(map[string]string) bool) (ret *route) {
	methodName := map[method]string{
		get:  "GET",
		head: "HEAD",
		post: "POST",
		put:  "PUT",
		del:  "DELETE",
	}

	ret = &route{
		method:      methodName[m],
		path:        path,
		constraints: constraints,
		controller:  controller,
		fn:          fn,
	}

	for _, v := range []func(string) pathMatcherFn{matcherStar, matcherStarMiddle, matcherBasic} {
		if ret.pathMatcher != nil {
			break
		}
		ret.pathMatcher = v(path)
	}
	return
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
		if !ok {
			return nil, false
		}
	}

	return m, true
}

func methodMatch(m1, m2 string) bool {
	if len(m1) > 0 && m1 != m2 {
		return false
	}
	return true
}
