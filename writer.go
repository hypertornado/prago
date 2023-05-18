package prago

type Writer[T any] struct {
	app      *App
	template string

	beforeFN func(*Request, *T)
	afterFN  func(*Request, *T)
}

func NewWriter[T any](app *App, template string) *Writer[T] {
	return &Writer[T]{
		app:      app,
		template: template,
	}
}

func (w *Writer[T]) Before(fn func(*Request, *T)) *Writer[T] {
	w.beforeFN = fn
	return w
}

func (w *Writer[T]) After(fn func(*Request, *T)) *Writer[T] {
	w.afterFN = fn
	return w
}

func (w *Writer[T]) GET(path string, handler func(*Request, *T), constraints ...routerConstraint) {
	w.route(get, path, handler, constraints...)

}

func (w *Writer[T]) POST(path string, handler func(*Request, *T), constraints ...routerConstraint) {
	w.route(post, path, handler, constraints...)

}

func (w *Writer[T]) route(method method, path string, handler func(*Request, *T), constraints ...routerConstraint) {

	action := func(request *Request) {
		var d T
		dp := &d
		request.ResponseTemplate = w.template

		if w.beforeFN != nil {
			w.beforeFN(request, dp)
		}

		handler(request, dp)

		if w.afterFN != nil {
			w.afterFN(request, dp)
		}

		if !request.Written && request.ResponseTemplate != "" {
			request.WriteHTML(request.ResponseStatus, request.ResponseTemplate, dp)
		}
	}

	w.app.appController.router.route(method, path, w.app.appController, action, constraints...)

}
