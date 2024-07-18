package prago

import (
	"net/url"
	"strconv"

	"golang.org/x/net/context"
)

func ResourceItemView[T any](app *App, url string, templates *PragoTemplates, template string, dataSource func(*T, *Request) interface{}) *Action {
	resource := getResource[T](app)
	action := resource.newItemAction(url)
	action.uiView(templates, template, func(request *Request) interface{} {
		item := resource.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		return dataSource(item.(*T), request)
	})
	return action
}

func (resource *Resource) newItemAction(itemUrl string) *Action {
	action := newAction(resource.app, itemUrl)
	action.resource = resource
	action.isItemAction = true
	action.permission = resource.canView
	action.addConstraint(func(ctx context.Context, values url.Values) bool {
		id, err := strconv.Atoi(values.Get("id"))
		if err != nil {
			return false
		}
		item := resource.query(ctx).ID(id)
		return item != nil
	})

	resource.itemActions = append(resource.itemActions, action)
	return action
}

func (resource *Resource) itemActionUi(itemURL string, handler func(any, *Request, *pageData)) *Action {
	action := resource.newItemAction(itemURL)

	action.ui(func(request *Request, pd *pageData) {
		item := resource.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		handler(item, request, pd)
	})

	return action
}

func ResourceItemHandler[T any](app *App, url string, fn func(*T, *Request)) *Action {
	resource := getResource[T](app)
	return resource.itemActionHandler(url, func(item any, request *Request) {
		fn(item.(*T), request)
	})
}

func (resource *Resource) itemActionHandler(url string, fn func(any, *Request)) *Action {
	action := resource.newItemAction(url)

	return action.Handler(func(request *Request) {
		item := resource.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		fn(item, request)
	})
}
