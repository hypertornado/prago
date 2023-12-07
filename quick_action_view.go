package prago

/*type QuickActionView struct {
	ActionURL string
	Name      string
	TypStr    string
	ItemID    int64
}

type QuickActionAPIResponse struct {
	Redirect string
	Error    string
}

func (resourceData *resourceData) getQuickActionViews(itemIface any, request *Request) (ret []QuickActionView) {
	for _, quickActionData := range resourceData.quickActions {
		if quickActionData.validation == nil || quickActionData.validation(itemIface, request) {
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
				ActionURL: quickActionData.getApiURL(resourceData.previewer(request, itemIface).ID()),
				Name:      quickActionData.singularName(request.Locale()),
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
	item := resourceData.query(request.r.Context()).ID(itemID)
	if item == nil {
		return errors.New("nelze nalézt položku")
	}
	for _, actionData := range resourceData.quickActions {
		if actionData.url == actionName {
			if actionData.handler == nil {
				return errors.New("není přiřazena žádná akce")
			}
			return actionData.handler(item, request)
		}
	}
	return errors.New("chyba akce")
}
*/
