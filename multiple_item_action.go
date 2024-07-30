package prago

import (
	"fmt"
	"reflect"
	"time"
)

type MultipleItemAction struct {
	ID         string
	ActionType string
	Icon       string
	Name       func(string) string
	Permission Permission
	Handler    func(items []any, request UserData, response *MultipleItemActionResponse)
}

func (app *App) initMultipleItemActions() {
	for _, resource := range app.resources {
		resource.addDefaultMultipleActions()
	}
}

func AddMultipleItemsAction[T any](
	app *App,
	name func(string) string, permission Permission, icon string,
	handler func(items []*T, request UserData, response *MultipleItemActionResponse)) {

	resource := getResource[T](app)

	resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "customaction-" + randomString(10),
		Icon:       icon,
		Name:       name,
		Permission: permission,
		Handler: func(items []any, request UserData, response *MultipleItemActionResponse) {
			var arr []*T
			for _, item := range items {
				arr = append(arr, item.(*T))

			}
			handler(arr, request, response)
		},
	})

}

type MultipleItemActionResponse struct {
	FlashMessage string
	RedirectURL  string
}

func (resource *Resource) addDefaultMultipleActions() {

	resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "edit",
		ActionType: "mutiple_edit",
		Icon:       iconEdit,
		Name:       unlocalized("Upravit"),
		Permission: resource.canUpdate,
	})

	resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "clone",
		Icon:       iconDuplicate,
		Name:       unlocalized("Naklonovat"),
		Permission: resource.canCreate,
		Handler: func(items []any, request UserData, response *MultipleItemActionResponse) {
			for _, item := range items {
				val := reflect.ValueOf(item).Elem()
				val.FieldByName("ID").SetInt(0)
				timeVal := reflect.ValueOf(time.Now())
				for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
					field := val.FieldByName(fieldName)
					if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
						field.Set(timeVal)
					}
				}

				err := resource.createWithLog(item, request)
				if err != nil {
					panic(fmt.Sprintf("can't create item for clone %v: %s", item, err))
				}

				if resource.activityLog {
					must(
						resource.logActivity(request, nil, item),
					)
				}
			}

			response.FlashMessage = fmt.Sprintf("%d položek naklonováno", len(items))
		},
	})

	resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "delete",
		Icon:       iconDelete,
		Name:       unlocalized("Smazat"),
		Permission: resource.canDelete,
		Handler: func(items []any, request UserData, response *MultipleItemActionResponse) {
			for _, item := range items {
				valValidation := resource.validateDelete(item, request)
				if !valValidation.Valid() {
					panic("cant validate delete")
				}

				err := resource.deleteWithLog(item, request)
				must(err)
				response.FlashMessage = fmt.Sprintf("%d položek smazáno", len(items))
			}
		},
	})

}

func (resource *Resource) allowsMultipleActions(userData UserData) (ret bool) {
	return len(resource.getMultipleActions(userData)) > 0
}

func (resource *Resource) getMultipleActions(userData UserData) (ret []listMultipleAction) {
	for _, ma := range resource.multipleActions {
		if !userData.Authorize(ma.Permission) {
			continue
		}
		ret = append(ret, listMultipleAction{
			ID:         ma.ID,
			ActionType: ma.ActionType,
			Icon:       ma.Icon,
			Name:       ma.Name(userData.Locale()),
		})
	}
	return
}
