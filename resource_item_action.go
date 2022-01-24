package prago

type ResourceItemAction[T any] struct {
	resource *Resource[T]
	action   *Action
}

func (resource *Resource[T]) ItemAction(url string) *ResourceItemAction[T] {
	ret := &ResourceItemAction[T]{
		resource: resource,
	}
	action := newAction(resource.app, url)
	action.resource = resource
	action.isItemAction = true
	action.permission = resource.canView
	resource.itemActions = append(resource.itemActions, action)
	ret.action = action
	return ret
}

func (action *ResourceItemAction[T]) priority() *ResourceItemAction[T] {
	action.action.priority()
	return action
}

func (action *ResourceItemAction[T]) IsWide() *ResourceItemAction[T] {
	action.action.IsWide()
	return action
}

func (action *ResourceItemAction[T]) Template(template string) *ResourceItemAction[T] {
	action.action.Template(template)
	return action
}

func (action *ResourceItemAction[T]) Permission(permission Permission) *ResourceItemAction[T] {
	action.action.Permission(permission)
	return action
}

func (action *ResourceItemAction[T]) DataSource(dataSource func(*T, *Request) interface{}) *ResourceItemAction[T] {
	action.action.DataSource(func(request *Request) interface{} {
		item := action.resource.Query().Is("id", request.Params().Get("id")).First()
		if item == nil {
			//TODO: fix http: superfluous response.WriteHeader call from github.com/hypertornado/prago.Request.RenderViewWithCode
			render404(request)
			return nil
		}
		return dataSource(item, request)
	})
	return action
}

func (action *ResourceItemAction[T]) Name(name func(string) string) *ResourceItemAction[T] {
	action.action.Name(name)
	return action
}

func (action *ResourceItemAction[T]) Method(method string) *ResourceItemAction[T] {
	action.action.Method(method)
	return action
}

func (action *ResourceItemAction[T]) Handler(fn func(*T, *Request)) *ResourceItemAction[T] {
	action.action.Handler(func(request *Request) {
		item := action.resource.Is("id", request.Params().Get("id")).First()
		if item == nil {
			render404(request)
			return
		}
		fn(item, request)
	})
	return action
}
