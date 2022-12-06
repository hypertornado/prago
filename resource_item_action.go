package prago

import (
	"context"
	"strconv"
)

type ResourceItemAction[T any] struct {
	data *resourceItemActionData
}

type resourceItemActionData struct {
	resourceData *resourceData
	action       *Action
}

func (resource *Resource[T]) ItemAction(url string) *ResourceItemAction[T] {
	return &ResourceItemAction[T]{
		data: resource.data.ItemAction(url),
	}
}

func (resourceData *resourceData) ItemAction(url string) *resourceItemActionData {
	action := newAction(resourceData.app, url)
	action.resourceData = resourceData
	action.isItemAction = true
	action.permission = resourceData.canView
	action.addConstraint(func(m map[string]string) bool {
		id, err := strconv.Atoi(m["id"])
		if err != nil {
			return false
		}
		item := resourceData.query(context.TODO()).ID(id)
		return item != nil
	})

	resourceData.itemActions = append(resourceData.itemActions, action)

	return &resourceItemActionData{
		resourceData: resourceData,
		action:       action,
	}
}

func (actionData *resourceItemActionData) priority() *resourceItemActionData {
	actionData.action.priority()
	return actionData
}

func (action *ResourceItemAction[T]) Icon(icon string) *ResourceItemAction[T] {
	action.data.Icon(icon)
	return action
}

func (actionData *resourceItemActionData) Icon(icon string) *resourceItemActionData {
	actionData.action.icon = icon
	return actionData
}

func (action *ResourceItemAction[T]) Template(template string) *ResourceItemAction[T] {
	action.data.Template(template)
	return action
}

func (actionData *resourceItemActionData) Template(template string) *resourceItemActionData {
	actionData.action.Template(template)
	return actionData
}

func (action *ResourceItemAction[T]) Permission(permission Permission) *ResourceItemAction[T] {
	action.data.Permission(permission)
	return action
}

func (actionData *resourceItemActionData) Permission(permission Permission) *resourceItemActionData {
	actionData.action.Permission(permission)
	return actionData
}

func (action *ResourceItemAction[T]) DataSource(dataSource func(*T, *Request) interface{}) *ResourceItemAction[T] {
	action.data.DataSource(func(t any, r *Request) interface{} {
		return dataSource(t.(*T), r)
	})
	return action
}

func (actionData *resourceItemActionData) DataSource(dataSource func(any, *Request) interface{}) *resourceItemActionData {
	actionData.action.DataSource(func(request *Request) interface{} {
		item := actionData.resourceData.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}

		return dataSource(item, request)
	})
	return actionData
}

func (action *ResourceItemAction[T]) Name(name func(string) string) *ResourceItemAction[T] {
	action.data.Name(name)
	return action
}

func (actionData *resourceItemActionData) Name(name func(string) string) *resourceItemActionData {
	actionData.action.Name(name)
	return actionData
}

func (action *ResourceItemAction[T]) Method(method string) *ResourceItemAction[T] {
	action.data.Method(method)
	return action
}

func (actionData *resourceItemActionData) Method(method string) *resourceItemActionData {
	actionData.action.Method(method)
	return actionData
}

func (action *ResourceItemAction[T]) Handler(fn func(*T, *Request)) *ResourceItemAction[T] {
	action.data.Handler(func(t any, r *Request) {
		fn(t.(*T), r)
	})
	return action
}

func (actionData *resourceItemActionData) Handler(fn func(any, *Request)) *resourceItemActionData {
	actionData.action.Handler(func(request *Request) {
		item := actionData.resourceData.query(request.r.Context()).ID(request.Param("id"))
		if item == nil {
			panic("can't find item")
		}
		fn(item, request)
	})
	return actionData
}
