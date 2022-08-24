package prago

import "fmt"

type QuickAction[T any] struct {
	data *quickActionData
}

type quickActionData struct {
	url          string
	resource     resourceIface
	permission   Permission
	singularName func(locale string) string
	pluralName   func(locale string) string
	typ          quickActionType

	validation func(any, *user) bool
	handler    func(any, *Request) error
}

type quickActionIface interface {
	getData() *quickActionData
}

type quickActionType int64

const (
	quickTypeBasic  quickActionType = 1
	quickTypeDelete quickActionType = 2
	quickTypeGreen  quickActionType = 3
	quickTypeBlue   quickActionType = 4
)

func (resource *Resource[T]) QuickAction(url string) *QuickAction[T] {
	ret := &QuickAction[T]{
		data: &quickActionData{
			url:        url,
			resource:   resource,
			permission: sysadminPermission,
			singularName: func(string) string {
				return url
			},
			pluralName: func(string) string {
				return url
			},
			typ: quickTypeBasic,
		},
	}
	resource.data.quickActions = append(resource.data.quickActions, ret)
	return ret
}

func (qa *QuickAction[T]) getData() *quickActionData {
	return qa.data
}

func (data *quickActionData) getApiURL(id int64) string {
	return fmt.Sprintf("/admin/%s/api/quick-action?action=%s&itemid=%d", data.resource.getData().getID(), data.url, id)
}

func (qa *QuickAction[T]) Permission(permission Permission) *QuickAction[T] {
	must(qa.data.resource.getData().app.validatePermission(permission))
	qa.data.permission = permission
	return qa
}

func (qa *QuickAction[T]) Name(singular, plural func(string) string) *QuickAction[T] {
	qa.data.singularName = singular
	qa.data.pluralName = plural
	return qa
}

func (qa *QuickAction[T]) Validation(validation func(*T, *user) bool) *QuickAction[T] {

	qa.data.validation = func(a any, u *user) bool {
		return validation(a.(*T), u)
	}
	return qa
}

func (qa *QuickAction[T]) Handler(handler func(*T, *Request) error) *QuickAction[T] {
	qa.data.handler = func(a any, request *Request) error {
		return handler(a.(*T), request)
	}
	return qa
}

func (qa *QuickAction[T]) DeleteType() *QuickAction[T] {
	qa.data.typ = quickTypeDelete
	return qa
}

func (qa *QuickAction[T]) GreenType() *QuickAction[T] {
	qa.data.typ = quickTypeGreen
	return qa
}

func (qa *QuickAction[T]) BlueType() *QuickAction[T] {
	qa.data.typ = quickTypeBlue
	return qa
}
