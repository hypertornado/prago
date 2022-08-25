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

type QuickActionAPIResponse struct {
	Redirect string
	Error    string
}

func (qa *QuickAction[T]) getView() *QuickActionView {
	return nil
}

func (resourceData *resourceData) getQuickActionViews(itemIface any, user *user) (ret []QuickActionView) {
	//item := itemIface.(*T)
	for _, v := range resourceData.quickActions {
		quickActionData := v.getData()
		if quickActionData.validation == nil || quickActionData.validation(itemIface, user) {
			var typStr string
			switch quickActionData.typ {
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
				ActionURL: quickActionData.getApiURL(getItemID(itemIface)),
				Name:      quickActionData.singularName(user.Locale),
				TypStr:    typStr,
			}
			ret = append(ret, view)
		}
	}
	return ret
}

func quickActionAPIHandler(resourceData *resourceData, request *Request) {
	var actionName = request.Param("action")

	itemID, err := strconv.Atoi(request.Param("itemid"))
	must(err)

	err = resourceData.runQuickAction(actionName, int64(itemID), request)
	if err != nil {
		renderAPIMessage(request, 500, err.Error())
		return
	}
	renderAPIMessage(request, 200, "")
}

func (resourceData *resourceData) runQuickAction(actionName string, itemID int64, request *Request) error {
	item := resourceData.query().ID(itemID)
	if item == nil {
		return errors.New("nelze nalézt položku")
	}
	for _, v := range resourceData.quickActions {
		actionData := v.getData()
		if actionData.url == actionName {
			if actionData.handler == nil {
				return errors.New("není přiřazena žádná akce")
			}
			return actionData.handler(item, request)
		}
	}
	return errors.New("chyba akce")
}
