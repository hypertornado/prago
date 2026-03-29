package prago

func ActionResourceMultipleItemsForm[T any](
	app *App,
	url string,
	formGenerator func(items []*T, form *Form, request *Request),
	validation func(items []*T, fv FormValidation, request *Request),
) *Action {
	resource := getResource[T](app)

	return resource.formItemMultipleAction(
		url,
		func(a []any, f *Form, r *Request) {
			items := make([]*T, len(a))
			for i, v := range a {
				items[i] = v.(*T)
			}
			formGenerator(items, f, r)
		},
		func(a []any, fv FormValidation, r *Request) {
			items := make([]*T, len(a))
			for i, v := range a {
				items[i] = v.(*T)
			}
			validation(items, fv, r)
		},
	)
}

func (resource *Resource) hasMultipleActions(userData UserData) (ret bool) {
	return len(resource.getMultipleActions(userData)) > 0
}

func (resource *Resource) getMultipleActions(userData UserData) (ret []listMultipleAction) {
	for _, action := range resource.itemActions {
		if !action.isFormMultipleAction {
			continue
		}
		if action.method != "GET" {
			continue
		}
		if !userData.Authorize(action.permission) {
			continue
		}
		ret = append(ret, listMultipleAction{
			ID:         action.url,
			ResourceID: resource.id,
			Icon:       action.icon,
			Name:       action.name(userData.Locale()),
		})

	}

	return
}
