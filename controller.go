package prago

type Controller struct {
	parent        *Controller
	app           *App
	aroundActions []func(p Request, next func())
}

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

func (c *Controller) SubController() (controller *Controller) {
	return &Controller{
		parent:        c,
		app:           c.app,
		aroundActions: []func(p Request, next func()){},
	}
}

func (c *Controller) AddBeforeAction(fn func(p Request)) {
	c.AddAroundAction(func(p Request, next func()) {
		fn(p)
		next()
	})
}

func (c *Controller) AddAfterAction(fn func(p Request)) {
	c.AddAroundAction(func(p Request, next func()) {
		next()
		fn(p)
	})
}

func (c *Controller) AddAroundAction(fn func(p Request, next func())) {
	c.aroundActions = append(c.aroundActions, fn)
}

type Action struct {
	controller *Controller
	fn         func(p Request)
}

func (c *Controller) NewAction(fn func(p Request)) *Action {
	return &Action{c, fn}
}

func (a *Action) call(p Request) {
	a.controller.callArounds(p, 0, func() {
		a.fn(p)
	}, true)
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

func (c *Controller) Get(path string, action func(p Request), constraints ...Constraint) {
	c.app.Route(get, path, c, action, constraints...)
}

func (c *Controller) Post(path string, action func(p Request), constraints ...Constraint) {
	c.app.Route(post, path, c, action, constraints...)
}

func (c *Controller) Put(path string, action func(p Request), constraints ...Constraint) {
	c.app.Route(put, path, c, action, constraints...)
}

func (c *Controller) Delete(path string, action func(p Request), constraints ...Constraint) {
	c.app.Route(del, path, c, action, constraints...)
}
