package prago

//Controller struct
type Controller struct {
	parent         *Controller
	router         *router
	priorityRouter bool
	aroundActions  []func(p Request, next func())
}

//MainController returns main controller of application
//all controllers in app are children of this controller
/*func (app *App) MainController() (ret *Controller) {
	return app.mainController
}*/

func newMainController() *Controller {
	return &Controller{
		parent:        nil,
		router:        newRouter(),
		aroundActions: []func(p Request, next func()){},
	}
}

func (c *Controller) dispatchRequest(request Request) bool {
	request.data = make(map[string]interface{})
	parseRequest(request)
	return c.router.process(request)
}

//SubController returns subcontroller of controller
func (c *Controller) subController() *Controller {
	return &Controller{
		parent:         c,
		router:         c.router,
		priorityRouter: c.priorityRouter,
		aroundActions:  []func(p Request, next func()){},
	}
}

//AddBeforeAction adds action which is executed before main router action is called
func (c *Controller) addBeforeAction(fn func(p Request)) {
	c.addAroundAction(func(p Request, next func()) {
		fn(p)
		next()
	})
}

//AddAfterAction adds action which is executed after main router action is called
func (c *Controller) addAfterAction(fn func(p Request)) {
	c.addAroundAction(func(p Request, next func()) {
		next()
		fn(p)
	})
}

//AddAroundAction adds action which is executed before and after action
//next function needs to be called in fn function
func (c *Controller) addAroundAction(fn func(p Request, next func())) {
	c.aroundActions = append(c.aroundActions, fn)
}

func (c *Controller) callArounds(p Request, i int, finalFunc func(), down bool) {
	if down {
		if c.parent != nil {
			c.parent.callArounds(p, 0, func() {
				c.callArounds(p, 0, finalFunc, false)
			}, down)
		} else {
			c.callArounds(p, 0, finalFunc, false)
		}
		return
	}

	if i < len(c.aroundActions) {
		c.aroundActions[i](p, func() {
			c.callArounds(p, i+1, finalFunc, false)
		})
	} else {
		finalFunc()
	}
}

func (router *router) route(m method, path string, controller *Controller, routeAction func(p Request), constraints ...func(map[string]string) bool) {
	route := newRoute(m, path, controller, routeAction, constraints)
	router.addRoute(route)

}

//Get creates new route for GET request
func (c *Controller) get(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	c.router.route(get, path, c, action, constraints...)
}

//Post creates new route for POST request
func (c *Controller) post(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	c.router.route(post, path, c, action, constraints...)
}

//Put creates new route for PUT request
func (c *Controller) put(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	c.router.route(put, path, c, action, constraints...)
}

//Delete creates new route for DELETE request
func (c *Controller) delete(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	c.router.route(del, path, c, action, constraints...)
}

//Get creates new route for GET request
func (app *App) GET(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	app.appController.get(path, action, constraints...)
}

//Post creates new route for POST request
func (app *App) POST(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	app.appController.post(path, action, constraints...)
}

//Put creates new route for PUT request
func (app *App) PUT(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	app.appController.put(path, action, constraints...)
}

//Delete creates new route for DELETE request
func (app *App) DELETE(path string, action func(p Request), constraints ...func(map[string]string) bool) {
	app.appController.delete(path, action, constraints...)
}

//AddBeforeAction adds action which is executed before main router action is called
func (app *App) AddBeforeAction(fn func(p Request)) {
	app.appController.addBeforeAction(fn)
}

//AddAfterAction adds action which is executed after main router action is called
func (app *App) AddAfterAction(fn func(p Request)) {
	app.appController.addAfterAction(fn)
}
