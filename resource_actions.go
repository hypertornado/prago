package prago

import (
	"fmt"
	"net/url"
	"strconv"
)

func initDefaultResourceActions(resource *resource) {
	app := resource.app

	newResourceAction(resource, "").priority().Permission(resource.canView).Name(resource.name).IsWide().Template("admin_list").DataSource(
		func(request *Request) interface{} {
			listData, err := resource.getListHeader(request.user)
			must(err)
			return listData
		},
	)

	newResourceFormAction(resource, "new").priority().Permission(resource.canCreate).Name(messages.GetNameFunction("admin_new")).Form(
		func(form *Form, request *Request) {
			var item interface{}
			resource.newItem(&item)
			resource.bindData(&item, request.user, request.Request().URL.Query())
			resource.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		for _, v := range resource.validations {
			v(vc)
		}
		request := vc.Request()
		if vc.Valid() {
			var item interface{}
			resource.newItem(&item)

			resource.bindData(item, request.user, request.Params())
			if resource.orderField != nil {
				resource.setOrderPosition(&item, resource.count()+1)
			}
			must(app.create(item))

			if app.search != nil {
				go func() {
					err := app.search.saveItem(resource, item)
					if err != nil {
						app.Log().Println(fmt.Errorf("%s", err))
					}
					app.search.flush()
				}()
			}

			if resource.activityLog {
				must(
					app.LogActivity("new", request.UserID(), resource.id, getItemID(item), nil, item),
				)
			}

			must(resource.updateCachedCount())

			app.Notification(getItemName(item)).
				SetImage(app.getItemImage(item)).
				SetPreName(messages.Get(request.user.Locale, "admin_item_created")).
				Flash(request)
			vc.Validation().RedirectionLocaliton = resource.getItemURL(item, "")
		}
	})

	newResourceItemAction(resource, "").priority().IsWide().Template("admin_views").Permission(resource.canView).DataSource(
		func(request *Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = app.is("id", int64(id)).get(item)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return nil
				}
				panic(err)
			}
			return resource.getViews(id, item, request.user)
		},
	)

	newResourceItemFormAction(resource, "edit").priority().Name(messages.GetNameFunction("admin_edit")).Permission(resource.canEdit).Form(
		func(form *Form, request *Request) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			app.is("id", int64(id)).mustGet(item)

			resource.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		request := vc.Request()
		params := request.Params()

		resource.fixBooleanParams(vc.Request().user, params)

		item, validation, err := resource.editItemWithLog(request.user, params)
		if err != nil && err != errValidation {
			panic(err)
		}

		if validation.Valid() {
			user := request.user
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			app.Notification(getItemName(item)).
				SetImage(app.getItemImage(item)).
				SetPreName(messages.Get(user.Locale, "admin_item_edited")).
				Flash(request)

			vc.Validation().RedirectionLocaliton = resource.getURL(fmt.Sprintf("%d", id))
		} else {
			//TODO: ugly hack with copying two validation contexts
			vc.Validation().Errors = validation.Validation().Errors
			vc.Validation().ItemErrors = validation.Validation().ItemErrors
		}
	})

	newResourceItemFormAction(resource, "delete").priority().Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete")).Form(
		func(form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.user.Locale, "admin_delete"))

			var item interface{}
			resource.newItem(&item)
			app.is("id", request.Params().Get("id")).mustGet(item)
			itemName := getItemName(item)
			form.Title = messages.Get(request.user.Locale, "admin_delete_confirmation_name", itemName)
		},
	).Validation(func(vc ValidationContext) {
		for _, v := range resource.deleteValidations {
			v(vc)
		}
		if vc.Valid() {
			id, err := strconv.Atoi(vc.GetValue("id"))
			must(err)

			must(resource.deleteItemWithLog(vc.Request().user, int64(id)))
			must(resource.updateCachedCount())
			vc.Request().AddFlashMessage(messages.Get(vc.Request().user.Locale, "admin_item_deleted"))
			vc.Validation().RedirectionLocaliton = resource.getURL("")
		}
	})

	if resource.previewURL != nil {
		newResourceItemAction(resource, "preview").priority().Name(messages.GetNameFunction("admin_preview")).Handler(
			func(request *Request) {
				var item interface{}
				resource.newItem(&item)
				app.is("id", request.Params().Get("id")).mustGet(item)
				request.Redirect(
					resource.previewURL(item),
				)
			},
		)
	}

	if resource.activityLog {
		newResourceAction(resource, "history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.canEdit).DataSource(
			func(request *Request) interface{} {
				return app.getHistory(resource, 0)
			},
		)

		newResourceItemAction(resource, "history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Permission(resource.canEdit).Template("admin_history").DataSource(
			func(request *Request) interface{} {
				id, err := strconv.Atoi(request.Params().Get("id"))
				must(err)

				var item interface{}
				resource.newItem(&item)
				app.is("id", int64(id)).mustGet(item)

				return app.getHistory(resource, int64(id))
			},
		)

	}
}

func (resource *resource) deleteItemWithLog(user *user, id int64) error {
	var beforeItem interface{}
	resource.newItem(&beforeItem)
	err := resource.app.is("id", id).get(beforeItem)
	if err != nil {
		return fmt.Errorf("can't find item for deletion id '%d': %s", id, err)
	}

	if resource.activityLog {
		err = resource.app.LogActivity("delete", user.ID, resource.id, id, beforeItem, nil)
		if err != nil {
			return err
		}
	}

	var item interface{}
	resource.newItem(&item)
	_, err = resource.app.is("id", id).Debug().delete(item)
	if err != nil {
		return fmt.Errorf("can't delete item id '%d': %s", id, err)
	}

	if resource.app.search != nil {
		err = resource.app.search.deleteItem(resource, id)
		if err != nil {
			resource.app.Log().Println(fmt.Errorf("%s", err))
		}
		resource.app.search.flush()
	}

	return nil
}

func (resource *resource) editItemWithLog(user *user, values url.Values) (interface{}, ValidationContext, error) {
	app := resource.app

	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	//TODO: remove this ugly hack and copy values via reflect package
	var beforeItem, item interface{}
	resource.newItem(&beforeItem)
	err = app.is("id", id).get(beforeItem)
	if err != nil {
		return nil, nil, fmt.Errorf("can't get beforeitem with id %d: %s", id, err)
	}

	resource.newItem(&item)
	err = app.is("id", id).get(item)
	if err != nil {
		return nil, nil, fmt.Errorf("can't get item with id %d: %s", id, err)
	}

	err = resource.bindData(
		item, user, values,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("can't bind data (%d): %s", id, err)
	}

	stringableValues := resource.getItemStringEditableValues(item, user)
	var allValues url.Values = make(map[string][]string)
	for k, v := range stringableValues {
		allValues.Add(k, v)
	}

	vv := newValuesValidation(user.Locale, allValues)
	for _, v := range resource.validations {
		v(vv)
	}
	if !vv.Valid() {
		return nil, vv, errValidation
	}

	err = app.save(item)
	if err != nil {
		return nil, nil, fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if app.search != nil {
		go func() {
			err = app.search.saveItem(resource, item)
			if err != nil {
				app.Log().Println(fmt.Errorf("%s", err))
			}
			app.search.flush()
		}()
	}

	if resource.activityLog {
		must(
			app.LogActivity("edit", user.ID, resource.id, int64(id), beforeItem, item),
		)
	}

	return item, vv, nil

}
