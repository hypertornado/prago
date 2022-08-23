package prago

import "fmt"

type QuickAction[T any] struct {
	resource     resourceIface
	url          string
	permission   Permission
	singularName func(locale string) string
	pluralName   func(locale string) string
	typ          quickActionType

	validation func(*T, *user) bool

	//confirmPrompt func(count int64, locale string) string

	handler func(*T, *Request) error
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
		resource:   resource,
		url:        url,
		permission: sysadminPermission,
		singularName: func(string) string {
			return url
		},
		pluralName: func(string) string {
			return url
		},
		typ: quickTypeBasic,
	}
	resource.data.quickActions = append(resource.data.quickActions, ret)
	return ret
}

func (qa *QuickAction[T]) getApiURL(id int64) string {
	return fmt.Sprintf("/admin/%s/api/quick-action?action=%s&itemid=%d", qa.resource.getData().getID(), qa.url, id)
}

func (qa *QuickAction[T]) Permission(permission Permission) *QuickAction[T] {
	must(qa.resource.getData().app.validatePermission(permission))
	qa.permission = permission
	return qa
}

func (qa *QuickAction[T]) Name(singular, plural func(string) string) *QuickAction[T] {
	qa.singularName = singular
	qa.pluralName = plural
	return qa
}

func (qa *QuickAction[T]) Validation(validation func(*T, *user) bool) *QuickAction[T] {
	qa.validation = validation
	return qa
}

func (qa *QuickAction[T]) Handler(handler func(*T, *Request) error) *QuickAction[T] {
	qa.handler = handler
	return qa
}

func (qa *QuickAction[T]) DeleteType() *QuickAction[T] {
	qa.typ = quickTypeDelete
	return qa
}

func (qa *QuickAction[T]) GreenType() *QuickAction[T] {
	qa.typ = quickTypeGreen
	return qa
}

func (qa *QuickAction[T]) BlueType() *QuickAction[T] {
	qa.typ = quickTypeBlue
	return qa
}
