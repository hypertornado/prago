package prago

// Controller struct
type controller struct {
	app            *App
	parent         *controller
	priorityRouter bool
	aroundActions  []func(p *Request, next func())
}

func newMainController(app *App) *controller {
	return &controller{
		app:           app,
		parent:        nil,
		aroundActions: []func(p *Request, next func()){},
	}
}

func (c *controller) subController() *controller {
	return &controller{
		app:            c.app,
		parent:         c,
		priorityRouter: c.priorityRouter,
		aroundActions:  []func(p *Request, next func()){},
	}
}

// AddBeforeAction adds action which is executed before main router action is called
func (c *controller) addBeforeAction(fn func(p *Request)) {
	c.addAroundAction(func(p *Request, next func()) {
		fn(p)
		next()
	})
}

// AddAfterAction adds action which is executed after main router action is called
func (c *controller) addAfterAction(fn func(p *Request)) {
	c.addAroundAction(func(p *Request, next func()) {
		next()
		fn(p)
	})
}

// AddAroundAction adds action which is executed before and after action
// next function needs to be called in fn function
func (c *controller) addAroundAction(fn func(p *Request, next func())) {
	c.aroundActions = append(c.aroundActions, fn)
}

func (c *controller) callAroundActions(p *Request, i int, finalFunc func(), down bool) {
	if down {
		if c.parent != nil {
			c.parent.callAroundActions(p, 0, func() {
				c.callAroundActions(p, 0, finalFunc, false)
			}, down)
		} else {
			c.callAroundActions(p, 0, finalFunc, false)
		}
		return
	}

	if i < len(c.aroundActions) {
		c.aroundActions[i](p, func() {
			c.callAroundActions(p, i+1, finalFunc, false)
		})
	} else {
		finalFunc()
	}
}

func (router *router) route(method string, path string, controller *controller, routeAction func(p *Request), constraints ...routerConstraint) {
	route := newRoute(method, path, controller, routeAction, constraints)
	router.addRoute(route)

}

func (c *controller) routeHandler(method, path string, action func(p *Request), constraints ...routerConstraint) {
	c.app.router.route(method, path, c, action, constraints...)
}

// Get creates new route for GET request
func (app *App) Handle(method, path string, action func(p *Request), constraints ...routerConstraint) {
	app.appController.routeHandler(method, path, action, constraints...)
}

// AddBeforeAction adds action which is executed before main router action is called
func (app *App) BeforeAction(fn func(p *Request)) {
	app.appController.addBeforeAction(fn)
}

// AddAfterAction adds action which is executed after main router action is called
func (app *App) AfterAction(fn func(p *Request)) {
	app.appController.addAfterAction(fn)
}
