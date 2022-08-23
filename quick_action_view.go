package prago

import (
	"errors"
	"strconv"
)

type QuickActionView struct {
	ActionURL string
	Name      string
	TypStr    string
	ItemID    int64
}

type QuickActionIface interface {
	//getName(string) string
}

type QuickActionAPIResponse struct {
	Redirect string
	Error    string
}

func (qa *QuickAction[T]) getView() *QuickActionView {
	return nil
}

func (resource *Resource[T]) getQuickActionViews(itemIface any, user *user) (ret []QuickActionView) {
	item := itemIface.(*T)
	for _, v := range resource.data.quickActions {
		quickAction := v.(*QuickAction[T])
		if quickAction.validation == nil || quickAction.validation(item, user) {
			var typStr string
			switch quickAction.typ {
			case quickTypeBasic:
				typStr = "basic"
			case quickTypeDelete:
				typStr = "delete"
			case quickTypeGreen:
				typStr = "green"
			case quickTypeBlue:
				typStr = "blue"
			}

			view := QuickActionView{
				ActionURL: quickAction.getApiURL(getItemID(item)),
				Name:      quickAction.singularName(user.Locale),
				TypStr:    typStr,
			}
			ret = append(ret, view)
		}
	}
	return ret
}

func quickActionAPIHandler(resource resourceIface, request *Request) {
	var actionName = request.Param("action")

	itemID, err := strconv.Atoi(request.Param("itemid"))
	must(err)

	err = resource.runQuickAction(actionName, int64(itemID), request)
	if err != nil {
		renderAPIMessage(request, 500, err.Error())
		return
	}
	renderAPIMessage(request, 200, "")
}

func (resource *Resource[T]) runQuickAction(actionName string, itemID int64, request *Request) error {
	item := resource.ID(itemID)
	if item == nil {
		return errors.New("Nelze nalézt položku")
	}
	for _, v := range resource.data.quickActions {
		action := v.(*QuickAction[T])
		if action.url == actionName {
			if action.handler == nil {
				return errors.New("není přiřazena žádná akce")
			}
			return action.handler(item, request)
		}
	}
	return errors.New("chyba akce")
}
