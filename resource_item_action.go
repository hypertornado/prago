package prago

import (
	"net/url"
	"strconv"

	"golang.org/x/net/context"
)

func ResourceItemView[T any](app *App, url string, template string, dataSource func(*T, *Request) interface{}) *Action {
	resource := GetResource[T](app)
	return resource.itemActionView(url, template, func(item any, request *Request) interface{} {
		return dataSource(item.(*T), request)
	})
}

func (resource *Resource) itemActionView(url, template string, dataSource func(any, *Request) interface{}) *Action {
	action := resource.newItemAction(url)

	action.View(template, func(request *Request) interface{} {
		item := resource.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		return dataSource(item, request)
	})
	return action
}

func (resourceData *Resource) newItemAction(itemUrl string) *Action {
	action := newAction(resourceData.app, itemUrl)
	action.resourceData = resourceData
	action.isItemAction = true
	action.permission = resourceData.canView
	action.addConstraint(func(ctx context.Context, values url.Values) bool {
		id, err := strconv.Atoi(values.Get("id"))
		if err != nil {
			return false
		}
		item := resourceData.query(ctx).ID(id)
		return item != nil
	})

	resourceData.itemActions = append(resourceData.itemActions, action)
	return action
}

func (resourceData *Resource) itemActionUi(itemURL string, handler func(any, *Request, *pageData)) *Action {
	action := resourceData.newItemAction(itemURL)

	action.ui(func(request *Request, pd *pageData) {
		item := resourceData.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		handler(item, request, pd)
	})

	return action
}

func ResourceItemHandler[T any](app *App, url string, fn func(*T, *Request)) *Action {
	resource := GetResource[T](app)
	return resource.itemActionHandler(url, func(item any, request *Request) {
		fn(item.(*T), request)
	})
}

func (resourceData *Resource) itemActionHandler(url string, fn func(any, *Request)) *Action {
	action := resourceData.newItemAction(url)

	return action.Handler(func(request *Request) {
		item := resourceData.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		fn(item, request)
	})
}
