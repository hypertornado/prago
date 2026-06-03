package prago

// Controller struct
type controller struct {
	app            *App
	parent         *controller
	priorityRouter bool

	beforeActions []func(request *Request) bool
	afterActions  []func(request *Request) bool
}

func newController(app *App) *controller {
	return &controller{
		app: app,

		beforeActions: []func(request *Request) bool{},
		afterActions:  []func(request *Request) bool{},
	}

}

func (c *controller) subController() *controller {
	ret := newController(c.app)
	ret.parent = c
	ret.priorityRouter = c.priorityRouter
	return ret
}

func (c *controller) addBeforeAction(fn func(request *Request) bool) {
	c.beforeActions = append(c.beforeActions, fn)
}

func (c *controller) callBeforeActions(request *Request) bool {
	if c.parent != nil {
		ok := c.parent.callBeforeActions(request)
		if !ok {
			return false
		}
	}
	for _, action := range c.beforeActions {
		ok := action(request)
		if !ok {
			return false
		}
	}
	return true
}

func (c *controller) callAfterActions(request *Request) bool {
	if c.parent != nil {
		ok := c.parent.callAfterActions(request)
		if !ok {
			return false
		}
	}
	for _, action := range c.afterActions {
		ok := action(request)
		if !ok {
			return false
		}
	}
	return true
}

func (router *router) route(method string, path string, controller *controller, routeAction func(p *Request), constraints ...routerConstraint) {
	route := newRoute(method, path, controller, routeAction, constraints)
	router.addRoute(route)

}

func (c *controller) routeHandler(method, path string, action func(p *Request), constraints ...routerConstraint) {
	c.app.router.route(method, path, c, action, constraints...)
}

// Get creates new route for GET request
func (app *App) Handle(method, path string, action func(request *Request), constraints ...routerConstraint) {
	app.appController.routeHandler(method, path, action, constraints...)
}

// AddBeforeAction adds action which is executed before main router action is called
func (app *App) BeforeAction(fn func(request *Request) bool) {
	app.appController.addBeforeAction(fn)
}
