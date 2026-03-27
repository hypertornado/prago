package prago

type MultipleItemAction struct {
	ID         string
	ActionType string
	Icon       string
	Name       func(string) string
	Permission Permission
	Handler    func(items []any, request *Request, response *MultipleItemActionResponse)
}

func (app *App) initMultipleItemActions() {
	for _, resource := range app.resources {
		resource.addDefaultMultipleActions()
	}
}

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

/*
func AddMultipleItemsAction[T any](
	app *App,
	name func(string) string, permission Permission, icon string,
	handler func(items []*T, request *Request, response *MultipleItemActionResponse)) {

	resource := getResource[T](app)

	resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "customaction-" + randomString(10),
		Icon:       icon,
		Name:       name,
		Permission: permission,
		Handler: func(items []any, request *Request, response *MultipleItemActionResponse) {
			var arr []*T
			for _, item := range items {
				arr = append(arr, item.(*T))

			}
			handler(arr, request, response)
		},
	})
}*/

type MultipleItemActionResponse struct {
	FlashMessage string
	ErrorStr     string
	RedirectURL  string
	FormURL      string
}

func (resource *Resource) addDefaultMultipleActions() {

	/*
		resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
			ID:         "edit",
			ActionType: "mutiple_edit",
			Icon:       iconEdit,
			Name:       unlocalized("Upravit"),
			Permission: resource.canUpdate,
		})*/

	/*
		resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
			ID:         "export",
			ActionType: "mutiple_export",
			Icon:       iconDownload,
			Name:       unlocalized("Export .xlsx"),
			Permission: resource.canExport,
		})*/

	/*
		resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
			ID:         "clone",
			Icon:       iconDuplicate,
			Name:       unlocalized("Naklonovat"),
			Permission: resource.canCreate,
			Handler: func(items []any, request *Request, response *MultipleItemActionResponse) {
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

					validation := resource.validateUpdate(item, request)
					if !validation.Valid() {
						response.ErrorStr = fmt.Sprintf("Nelze naklonovat položku: %s", validation.TextErrorReport(0, request.Locale()).Text)
						return
					}

					err := resource.createWithLog(item, request)
					if err != nil {
						panic(fmt.Sprintf("can't create item for clone %v: %s", item, err))
					}

					must(
						resource.logActivity(request, nil, item),
					)

				}

				response.FlashMessage = fmt.Sprintf("%d položek naklonováno", len(items))
			},
		})*/

	/*resource.multipleActions = append(resource.multipleActions, &MultipleItemAction{
		ID:         "ai-context",
		Icon:       iconAI,
		Name:       unlocalized("AI Kontext"),
		Permission: resource.canView,
		Handler: func(items []any, request *Request, response *MultipleItemActionResponse) {

			var values url.Values = make(url.Values)
			values.Add("_resource", resource.id)

			var ids []int64
			for _, item := range items {
				val := reflect.ValueOf(item).Elem()
				ids = append(ids, val.FieldByName("ID").Int())
			}

			values.Add("_ids", MultirelationArrayToString(ids))

			response.FormURL = fmt.Sprintf("/admin/_aicontextresource?%s", values.Encode())
		},
	})*/

	/*
		resource.formItemMultipleAction(
			"ai-context2",
			func(items []any, form *Form, request *Request) {
				for _, field := range resource.fields {
					if !request.Authorize(field.canView) {
						continue
					}
					checkboxItem := form.AddCheckbox(field.id, field.name(request.Locale()))
					checkboxItem.Value = "on"
				}

				for _, stat := range resource.itemStats {
					if !request.Authorize(stat.Permission) {
						continue
					}
					checkboxItem := form.AddCheckbox(stat.id, stat.Name(request.Locale()))
					checkboxItem.Value = "on"
				}

				if resource.previewFn != nil {
					form.AddCheckbox("_previewurl", "Ukázat preview URL").Value = "on"
				}

				form.AddHidden("_resource").Value = resource.id
				form.AddRelationMultiple("_items", resource.pluralName(request.Locale()), resource.id).Value = request.Param("_ids")

				form.AddSubmit("Zobrazit")
				//form.AddSubmit("AI Kontext")
			},
			func(items []any, fv FormValidation, request *Request) {
				var fields []*Field
				//resource := app.getResourceByID(request.Param("_resource"))
				for _, field := range resource.fields {
					if !request.Authorize(field.canView) {
						continue
					}
					if request.Param(field.id) == "on" {
						fields = append(fields, field)
					}
				}

				ids := MultirelationStringToArray(request.Param("_items"))
				if len(ids) == 0 {
					fv.AddItemError("_items", "Není vybrána žádná položka")
				}

				if !fv.Valid() {
					return
				}

				var strVal string

				for _, id := range ids {
					item := resource.query(context.Background()).ID(id)

					for _, field := range fields {
						ifaceVal := reflect.ValueOf(item).Elem().FieldByName(field.fieldClassName).Interface()

						cellData := getCellViewData(request, field, ifaceVal)
						strVal += fmt.Sprintf("%s: %v\n", field.name(request.Locale()), cellData.Name)
					}

					for _, stat := range resource.itemStats {
						if !request.Authorize(stat.Permission) {
							continue
						}

						if request.Param(stat.id) == "on" {
							strVal += fmt.Sprintf("%s: %v\n", stat.Name(request.Locale()), stat.Handler(item))
						}
					}

					if resource.previewFn != nil && request.Param("_previewurl") == "on" {
						prevURL := resource.previewFn(item)
						strVal += fmt.Sprintf("Veřejné URL: %v\n", prevURL)
					}

					strVal += "\n\n---------\n\n"
				}

				fv.AfterContent(template.HTML(fmt.Sprintf("<textarea class=\"input\">%s</textarea>", template.HTMLEscapeString(strVal))))
			},
		).Icon(iconDelete).Permission(resource.canDelete).Name(unlocalized("AI Kontext 2"))
	*/
}

func (resource *Resource) hasMultipleActions(userData UserData) (ret bool) {
	return len(resource.getMultipleActions(userData)) > 0
}

func (resource *Resource) getMultipleActions(userData UserData) (ret []listMultipleAction) {
	for _, ma := range resource.multipleActions {
		if !userData.Authorize(ma.Permission) {
			continue
		}
		ret = append(ret, listMultipleAction{
			ID:         ma.ID,
			ResourceID: resource.id,
			ActionType: ma.ActionType,
			Icon:       ma.Icon,
			Name:       ma.Name(userData.Locale()),
		})
	}

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
			ActionType: "multiple_action_form",
			Icon:       action.icon,
			Name:       action.name(userData.Locale()),
		})

	}

	return
}
