package prago

import "fmt"

type QuickAction[T any] struct {
	data *quickActionData
}

type quickActionData struct {
	url          string
	resource     *resourceData
	permission   Permission
	singularName func(locale string) string
	pluralName   func(locale string) string
	typ          quickActionType

	validation func(any, *Request) bool
	handler    func(any, *Request) error
}

/*type quickActionIface interface {
	getData() *quickActionData
}*/

type quickActionType int64

const (
	quickTypeBasic  quickActionType = 1
	quickTypeDelete quickActionType = 2
	quickTypeGreen  quickActionType = 3
	quickTypeBlue   quickActionType = 4
)

//listMultipleAction

func (resource *Resource[T]) QuickAction(url string) *QuickAction[T] {
	ret := &QuickAction[T]{
		data: &quickActionData{
			url:        url,
			resource:   resource.data,
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

	//ret := &QuickAction[T]{data}

	resource.data.quickActions = append(resource.data.quickActions, ret.data)
	return ret
}

func (resourceData *resourceData) getMultipleActionsFromQuickActions(userData UserData) (ret []listMultipleAction) {

	for _, action := range resourceData.quickActions {
		a := action.getMultipleAction(userData)
		if a != nil {
			ret = append(ret, *a)
		}
	}

	return
}

func (data *quickActionData) getMultipleAction(userData UserData) *listMultipleAction {
	if !userData.Authorize(data.permission) {
		return nil
	}

	ret := listMultipleAction{
		ID:   data.url,
		Name: data.pluralName(userData.Locale()),
	}
	return &ret
}

func (data *quickActionData) getApiURL(id int64) string {
	return fmt.Sprintf("/admin/%s/api/quick-action?action=%s&itemid=%d", data.resource.getID(), data.url, id)
}

func (qa *QuickAction[T]) Permission(permission Permission) *QuickAction[T] {
	must(qa.data.resource.app.validatePermission(permission))
	qa.data.permission = permission
	return qa
}

func (qa *QuickAction[T]) Name(singular, plural func(string) string) *QuickAction[T] {
	qa.data.singularName = singular
	qa.data.pluralName = plural
	return qa
}

func (qa *QuickAction[T]) Validation(validation func(*T, *Request) bool) *QuickAction[T] {

	qa.data.validation = func(a any, r *Request) bool {
		return validation(a.(*T), r)
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
