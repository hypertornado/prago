package prago

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func (resource *Resource[T]) initDefaultResourceActions() {
	resource.Action("").priority().Permission(resource.data.canView).Name(resource.data.pluralName).Template("admin_list").DataSource(
		func(request *Request) interface{} {
			listData, err := resource.data.getListHeader(request.user)
			must(err)
			return listData
		},
	)

	resource.FormAction("new").priority().Permission(resource.data.canCreate).Name(messages.GetNameFunction("admin_new")).Form(
		func(form *Form, request *Request) {
			var item T
			resource.bindData(&item, request.user, request.Request().URL.Query())
			resource.addFormItems(&item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		for _, v := range resource.data.validations {
			v(vc)
		}
		request := vc.Request()
		if vc.Valid() {
			var item T
			resource.bindData(&item, request.user, request.Params())
			if resource.data.orderField != nil {
				count, _ := resource.Query().Count()
				resource.data.setOrderPosition(&item, count+1)
			}
			must(resource.CreateWithLog(&item, request))

			resource.data.app.Notification(getItemName(&item)).
				SetImage(resource.data.app.getItemImage(&item)).
				SetPreName(messages.Get(request.user.Locale, "admin_item_created")).
				Flash(request)
			vc.Validation().RedirectionLocaliton = resource.data.getItemURL(&item, "")
		}
	})

	resource.ItemAction("").priority().Template("admin_views").Permission(resource.data.canView).DataSource(
		func(item *T, request *Request) interface{} {
			if item == nil {
				render404(request)
				return nil
			}
			return resource.data.getViews(item, request.user)
		},
	)

	resource.FormItemAction("edit").priority().Name(messages.GetNameFunction("admin_edit")).Permission(resource.data.canUpdate).Form(
		func(item *T, form *Form, request *Request) {
			resource.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(_ *T, vc ValidationContext) {
		request := vc.Request()
		params := request.Params()

		resource.data.fixBooleanParams(vc.Request().user, params)

		item, validation, err := resource.editItemWithLogAndValues(request, params)
		if err != nil && err != errValidation {
			panic(err)
		}

		if validation.Valid() {
			user := request.user
			id, err := strconv.Atoi(request.Param("id"))
			must(err)

			resource.data.app.Notification(getItemName(item)).
				SetImage(resource.data.app.getItemImage(item)).
				SetPreName(messages.Get(user.Locale, "admin_item_edited")).
				Flash(request)

			vc.Validation().RedirectionLocaliton = resource.data.getURL(fmt.Sprintf("%d", id))
		} else {
			//TODO: ugly hack with copying two validation contexts
			vc.Validation().Errors = validation.Validation().Errors
			vc.Validation().ItemErrors = validation.Validation().ItemErrors
		}
	})

	/*resource.QuickAction("test_quick").Name(unlocalized("Test buttonu 1"), unlocalized("Testy buttonu 1"))
	resource.QuickAction("test_quick2").DeleteType()
	resource.QuickAction("test_quick_green").GreenType().Handler(func(t *T, r *Request) error {
		return errors.New("green error")
	})
	resource.QuickAction("test_quick_blue").BlueType().Handler(func(t *T, r *Request) error {
		return nil
	})*/

	resource.FormItemAction("delete").priority().Permission(resource.data.canDelete).Name(messages.GetNameFunction("admin_delete")).Form(
		func(item *T, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.user.Locale, "admin_delete"))
			itemName := getItemName(item)
			form.Title = messages.Get(request.user.Locale, "admin_delete_confirmation_name", itemName)
		},
	).Validation(func(item *T, vc ValidationContext) {
		for _, v := range resource.data.deleteValidations {
			v(vc)
		}
		if vc.Valid() {
			must(resource.DeleteWithLog(item, vc.Request()))
			vc.Request().AddFlashMessage(messages.Get(vc.Request().user.Locale, "admin_item_deleted"))
			vc.Validation().RedirectionLocaliton = resource.data.getURL("")
		}
	})

	if resource.previewURLFunction != nil {
		resource.ItemAction("preview").priority().Name(messages.GetNameFunction("admin_preview")).Handler(
			func(item *T, request *Request) {
				request.Redirect(
					resource.previewURLFunction(item),
				)
			},
		)
	}

	if resource.data.activityLog {
		resource.Action("history").priority().Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.data.canUpdate).DataSource(
			func(request *Request) interface{} {
				return resource.data.app.getHistory(resource.data, 0)
			},
		)

		resource.ItemAction("history").priority().Name(messages.GetNameFunction("admin_history")).Permission(resource.data.canUpdate).Template("admin_history").DataSource(
			func(item *T, request *Request) interface{} {
				if item == nil {
					return nil
				}
				return resource.data.app.getHistory(resource.data, getItemID(item))
			},
		)
	}
}

func (resource *Resource[T]) CreateWithLog(item *T, request *Request) error {
	err := resource.Create(item)
	if err != nil {
		return err
	}

	if resource.data.app.search != nil {
		go func() {
			err := resource.saveSearchItem(item)
			if err != nil {
				resource.data.app.Log().Println(fmt.Errorf("%s", err))
			}
			resource.data.app.search.flush()
		}()
	}

	if resource.data.activityLog {
		err := resource.LogActivity(request.user, nil, item)
		if err != nil {
			return err
		}

	}
	return resource.data.updateCachedCount()

}

func (resource *Resource[T]) DeleteWithLog(item *T, request *Request) error {
	if resource.data.activityLog {
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

	if resource.data.app.search != nil {
		err = resource.data.app.search.deleteItem(resource.data, id)
		if err != nil {
			resource.data.app.Log().Println(fmt.Errorf("%s", err))
		}
		resource.data.app.search.flush()
	}

	resource.data.updateCachedCount()

	return nil
}

func (resource *Resource[T]) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, ValidationContext, error) {
	user := request.user
	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	beforeItem := resource.ID(id)
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
	for _, v := range resource.data.validations {
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
	beforeItem := resource.ID(id)
	if beforeItem == nil {
		return errors.New("can't find before item")
	}

	err := resource.Update(item)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if resource.data.app.search != nil {
		go func() {
			err = resource.saveSearchItem(item)
			if err != nil {
				resource.data.app.Log().Println(fmt.Errorf("%s", err))
			}
			resource.data.app.search.flush()
		}()
	}

	if resource.data.activityLog {
		must(
			resource.LogActivity(request.user, beforeItem, item),
		)
	}

	return nil
}
