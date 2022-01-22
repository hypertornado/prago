package prago

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (resource *Resource[T]) initDefaultResourceActions() {
	resource.Action("").priority().Permission(resource.canView).Name(resource.name).IsWide().Template("admin_list").DataSource(
		func(request *Request) interface{} {
			listData, err := resource.getListHeader(request.user)
			must(err)
			return listData
		},
	)

	resource.FormAction("new").priority().Permission(resource.canCreate).Name(messages.GetNameFunction("admin_new")).Form(
		func(form *Form, request *Request) {
			var item T
			resource.bindData(&item, request.user, request.Request().URL.Query())
			resource.addFormItems(&item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		for _, v := range resource.validations {
			v(vc)
		}
		request := vc.Request()
		if vc.Valid() {
			var item T
			resource.bindData(&item, request.user, request.Params())
			if resource.orderField != nil {
				count, _ := resource.Query().Count()
				resource.setOrderPosition(&item, count+1)
			}
			must(resource.CreateWithLog(&item, request))

			resource.app.Notification(getItemName(&item)).
				SetImage(resource.app.getItemImage(&item)).
				SetPreName(messages.Get(request.user.Locale, "admin_item_created")).
				Flash(request)
			vc.Validation().RedirectionLocaliton = resource.getItemURL(&item, "")
		}
	})

	resource.ItemAction("").priority().IsWide().Template("admin_views").Permission(resource.canView).DataSource(
		func(request *Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			item := resource.Is("id", int64(id)).First()
			if item == nil {
				render404(request)
				return nil
			}
			return resource.getViews(id, item, request.user)
		},
	)

	resource.FormItemAction("edit").priority().Name(messages.GetNameFunction("admin_edit")).Permission(resource.canUpdate).Form(
		func(form *Form, request *Request) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			item := resource.Query().Is("id", int64(id)).First()
			if item == nil {
				render404(request)
				return
			}

			resource.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		request := vc.Request()
		params := request.Params()

		resource.fixBooleanParams(vc.Request().user, params)

		item, validation, err := resource.editItemWithLogAndValues(request, params)
		if err != nil && err != errValidation {
			panic(err)
		}

		if validation.Valid() {
			user := request.user
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			resource.app.Notification(getItemName(item)).
				SetImage(resource.app.getItemImage(item)).
				SetPreName(messages.Get(user.Locale, "admin_item_edited")).
				Flash(request)

			vc.Validation().RedirectionLocaliton = resource.getURL(fmt.Sprintf("%d", id))
		} else {
			//TODO: ugly hack with copying two validation contexts
			vc.Validation().Errors = validation.Validation().Errors
			vc.Validation().ItemErrors = validation.Validation().ItemErrors
		}
	})

	resource.FormItemAction("delete").priority().Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete")).Form(
		func(form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.user.Locale, "admin_delete"))

			item := resource.Is("id", request.Params().Get("id")).First()
			if item == nil {
				render404(request)
				return
			}
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

			item := resource.Is("id", id).First()
			if item == nil {
				panic(fmt.Sprintf("can't find item for deletion id '%d'", id))
			}

			must(resource.DeleteWithLog(item, vc.Request()))
			vc.Request().AddFlashMessage(messages.Get(vc.Request().user.Locale, "admin_item_deleted"))
			vc.Validation().RedirectionLocaliton = resource.getURL("")
		}
	})

	if resource.previewURLFunction != nil {
		resource.ItemAction("preview").priority().Name(messages.GetNameFunction("admin_preview")).Handler(
			func(request *Request) {
				item := resource.Is("id", request.Params().Get("id")).First()
				if item == nil {
					render404(request)
					return
				}
				request.Redirect(
					resource.previewURLFunction(item),
				)
			},
		)
	}

	if resource.activityLog {
		resource.Action("history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.canUpdate).DataSource(
			func(request *Request) interface{} {
				return resource.app.getHistory(resource, 0)
			},
		)

		resource.ItemAction("history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Permission(resource.canUpdate).Template("admin_history").DataSource(
			func(request *Request) interface{} {
				id, err := strconv.Atoi(request.Params().Get("id"))
				must(err)

				item := resource.Query().Is("id", int64(id)).First()
				if item == nil {
					render404(request)
					return nil
				}

				return resource.app.getHistory(resource, int64(id))
			},
		)

	}
}

func (resource *Resource[T]) CreateWithLog(item *T, request *Request) error {
	err := resource.Create(item)
	if err != nil {
		return err
	}

	if resource.app.search != nil {
		go func() {
			err := resource.saveSearchItem(item)
			if err != nil {
				resource.app.Log().Println(fmt.Errorf("%s", err))
			}
			resource.app.search.flush()
		}()
	}

	if resource.activityLog {
		err := resource.LogActivity(request.user, nil, item)
		if err != nil {
			return err
		}

	}
	return resource.updateCachedCount()

}

func (resource *Resource[T]) DeleteWithLog(item *T, request *Request) error {
	if resource.activityLog {
		err := resource.LogActivity(request.user, item, nil)
		if err != nil {
			return err
		}
	}

	id := getItemID(item)

	err := resource.Delete(id)
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

	resource.updateCachedCount()

	return nil
}

func (resource *Resource[T]) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, ValidationContext, error) {
	user := request.user
	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	beforeItem := resource.Is("id", id).First()
	if beforeItem == nil {
		return nil, nil, fmt.Errorf("can't get beforeitem with id %d: %s", id, err)
	}

	cloned := *beforeItem
	item := &cloned

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

	err = resource.UpdateWithLog(item, request)
	if err != nil {
		return nil, nil, err
	}

	return item, vv, nil
}

func (resource *Resource[T]) UpdateWithLog(item *T, request *Request) error {
	id := getItemID(item)
	beforeItem := resource.Is("id", id).First()
	if beforeItem == nil {
		return errors.New("can't find before item")
	}

	err := resource.Update(item)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if resource.app.search != nil {
		go func() {
			err = resource.saveSearchItem(item)
			if err != nil {
				resource.app.Log().Println(fmt.Errorf("%s", err))
			}
			resource.app.search.flush()
		}()
	}

	if resource.activityLog {
		must(
			resource.LogActivity(request.user, beforeItem, item),
		)
	}

	return nil
}
