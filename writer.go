package prago

type Writer[T any] struct {
	app          *App
	templateName string
	templates    *PragoTemplates

	beforeFN func(*Request, *T)
	afterFN  func(*Request, *T)
}

func NewWriter[T any](app *App, templates *PragoTemplates, templateName string) *Writer[T] {
	return &Writer[T]{
		app:          app,
		templateName: templateName,
		templates:    templates,
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
	w.route("GET", path, handler, constraints...)
}

func (w *Writer[T]) POST(path string, handler func(*Request, *T), constraints ...routerConstraint) {
	w.route("POST", path, handler, constraints...)

}

func (w *Writer[T]) route(method string, path string, handler func(*Request, *T), constraints ...routerConstraint) {
	action := func(request *Request) {
		var d T
		dp := &d
		//request.ResponseTemplates = w.templates
		request.ResponseTemplateName = w.templateName

		if w.beforeFN != nil {
			w.beforeFN(request, dp)
		}

		handler(request, dp)

		if w.afterFN != nil {
			w.afterFN(request, dp)
		}

		if !request.Written && request.ResponseTemplateName != "" {
			request.WriteHTML(request.ResponseStatus, request.app.Templates, request.ResponseTemplateName, dp)
		}
	}

	w.app.router.route(method, path, w.app.appController, action, constraints...)
}
