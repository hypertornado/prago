package prago

//Controller struct
type Controller struct {
	parent        *Controller
	app           *App
	aroundActions []func(p Request, next func())
}

//MainController returns main controller of application
//all controllers in app are children of this controller
func (a *App) MainController() (ret *Controller) {
	ret = a.data["mainController"].(*Controller)
	if ret == nil {
		panic("couldnt find controller")
	}
	return
}

func newMainController(app *App) *Controller {
	return &Controller{
		parent:        nil,
		app:           app,
		aroundActions: []func(p Request, next func()){},
	}
}

//SubController returns subcontroller of controller
func (c *Controller) SubController() *Controller {
	return &Controller{
		parent:        c,
		app:           c.app,
		aroundActions: []func(p Request, next func()){},
	}
}

//AddBeforeAction adds action which is executed before main router action is called
func (c *Controller) AddBeforeAction(fn func(p Request)) {
	c.AddAroundAction(func(p Request, next func()) {
		fn(p)
		next()
	})
}

//AddAfterAction adds action which is executed after main router action is called
func (c *Controller) AddAfterAction(fn func(p Request)) {
	c.AddAroundAction(func(p Request, next func()) {
		next()
		fn(p)
	})
}

//AddAroundAction adds action which is executed before and after action
//next function needs to be called in fn function
func (c *Controller) AddAroundAction(fn func(p Request, next func())) {
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

//Get creates new route for GET request
func (c *Controller) Get(path string, action func(p Request), constraints ...Constraint) {
	c.app.route(get, path, c, action, constraints...)
}

//Post creates new route for POST request
func (c *Controller) Post(path string, action func(p Request), constraints ...Constraint) {
	c.app.route(post, path, c, action, constraints...)
}

//Put creates new route for PUT request
func (c *Controller) Put(path string, action func(p Request), constraints ...Constraint) {
	c.app.route(put, path, c, action, constraints...)
}

//Delete creates new route for DELETE request
func (c *Controller) Delete(path string, action func(p Request), constraints ...Constraint) {
	c.app.route(del, path, c, action, constraints...)
}
