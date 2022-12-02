package prago

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

func (resourceData *resourceData) initDefaultResourceActions() {
	resourceData.action("").Icon(iconTable).priority().Permission(resourceData.canView).Name(resourceData.pluralName).Template("admin_list").DataSource(
		func(request *Request) interface{} {
			listData, err := resourceData.getListHeader(request.user)
			must(err)
			return listData
		},
	)

	resourceData.FormAction("new").Icon(iconAdd).priority().Permission(resourceData.canCreate).Name(messages.GetNameFunction("admin_new")).Form(
		func(form *Form, request *Request) {
			//var item T
			var item interface{} = reflect.New(resourceData.typ).Interface()
			resourceData.bindData(item, request.user, request.Request().URL.Query())
			resourceData.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(vc ValidationContext) {
		for _, v := range resourceData.validations {
			v(vc)
		}
		request := vc.Request()
		if vc.Valid() {
			var item interface{} = reflect.New(resourceData.typ).Interface()
			resourceData.bindData(item, request.user, request.Params())
			if resourceData.orderField != nil {
				count, _ := resourceData.query().count()
				resourceData.setOrderPosition(item, count+1)
			}
			must(resourceData.CreateWithLog(item, request))

			resourceData.app.Notification(resourceData.previewer(request.user, item).Name()).
				SetImage(resourceData.previewer(request.user, item).ThumbnailURL()).
				SetPreName(messages.Get(request.user.Locale, "admin_item_created")).
				Flash(request)
			vc.Validation().RedirectionLocaliton = resourceData.getItemURL(item, "", request.user)
		}
	})

	resourceData.ItemAction("").Icon("glyphicons-basic-588-book-open-text.svg").priority().Template("admin_views").Permission(resourceData.canView).DataSource(
		func(item any, request *Request) interface{} {
			if item == nil {
				render404(request)
				return nil
			}
			return resourceData.getViews(item, request.user)
		},
	)

	resourceData.FormItemAction("edit").Icon("glyphicons-basic-31-pencil.svg").priority().Name(messages.GetNameFunction("admin_edit")).Permission(resourceData.canUpdate).Form(
		func(item any, form *Form, request *Request) {
			resourceData.addFormItems(item, request.user, form)
			form.AddSubmit(messages.Get(request.user.Locale, "admin_save"))
		},
	).Validation(func(_ any, vc ValidationContext) {
		request := vc.Request()
		params := request.Params()

		resourceData.fixBooleanParams(vc.Request().user, params)

		item, validation, err := resourceData.editItemWithLogAndValues(request, params)
		if err != nil && err != errValidation {
			panic(err)
		}

		if validation.Valid() {
			user := request.user
			id, err := strconv.Atoi(request.Param("id"))
			must(err)

			resourceData.app.Notification(resourceData.previewer(user, item).Name()).
				SetImage(resourceData.previewer(request.user, item).ThumbnailURL()).
				SetPreName(messages.Get(user.Locale, "admin_item_edited")).
				Flash(request)

			vc.Validation().RedirectionLocaliton = resourceData.getURL(fmt.Sprintf("%d", id))
		} else {
			//TODO: ugly hack with copying two validation contexts
			vc.Validation().Errors = validation.Validation().Errors
			vc.Validation().ItemErrors = validation.Validation().ItemErrors
		}
	})

	resourceData.FormItemAction("delete").Icon("glyphicons-basic-17-bin.svg").priority().Permission(resourceData.canDelete).Name(messages.GetNameFunction("admin_delete")).Form(
		func(item any, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.user.Locale, "admin_delete"))
			itemName := resourceData.previewer(request.user, item).Name()
			form.Title = messages.Get(request.user.Locale, "admin_delete_confirmation_name", itemName)
		},
	).Validation(func(item any, vc ValidationContext) {
		for _, v := range resourceData.deleteValidations {
			v(vc)
		}
		if vc.Valid() {
			must(resourceData.DeleteWithLog(item, vc.Request()))
			vc.Request().AddFlashMessage(messages.Get(vc.Request().user.Locale, "admin_item_deleted"))
			vc.Validation().RedirectionLocaliton = resourceData.getURL("")
		}
	})

	if resourceData.previewURLFunction != nil {
		resourceData.ItemAction("preview").Icon("glyphicons-basic-52-eye.svg").priority().Name(messages.GetNameFunction("admin_preview")).Handler(
			func(item any, request *Request) {
				request.Redirect(
					resourceData.previewURLFunction(item),
				)
			},
		)
	}

	if resourceData.activityLog {
		resourceData.FormAction("history").
			Icon("glyphicons-basic-58-history.svg").
			priority().
			Name(messages.GetNameFunction("admin_history")).
			//Template("admin_history").
			Permission(resourceData.canUpdate).
			Form(func(f *Form, r *Request) {

				f.AddTextInput("page", "Stránka").Value = "1"

				f.AddSubmit("Zobrazit")
				table := resourceData.app.getHistoryTable(r.user, resourceData, 0, "1")
				f.HTMLAfter = table.ExecuteHTML()

			}).
			Validation(func(vc ValidationContext) {
				table := resourceData.app.getHistoryTable(vc.Request().user, resourceData, 0, vc.GetValue("page"))
				vc.Validation().AfterContent = table.ExecuteHTML()

			})
			/*DataSource(
				func(request *Request) interface{} {
					return resourceData.app.getHistory(resourceData, 0)
				},
			)*/

		resourceData.
			FormItemAction("history").
			Icon("glyphicons-basic-58-history.svg").
			priority().
			Name(messages.GetNameFunction("admin_history")).
			Permission(resourceData.canUpdate).
			Form(func(item any, f *Form, r *Request) {
				f.AddTextInput("page", "Stránka").Value = "1"
				f.AddSubmit("Zobrazit")

				id := resourceData.previewer(r.user, item).ID()
				table := resourceData.app.getHistoryTable(r.user, resourceData, id, "1")
				f.HTMLAfter = table.ExecuteHTML()

			}).
			Validation(func(item any, vc ValidationContext) {
				id := resourceData.previewer(vc.Request().user, item).ID()
				table := resourceData.app.getHistoryTable(vc.Request().user, resourceData, id, vc.GetValue("page"))
				vc.Validation().AfterContent = table.ExecuteHTML()

			})
		//Template("admin_history").
		/*DataSource(
			func(item any, request *Request) interface{} {
				if item == nil {
					return nil
				}
				return resourceData.app.getHistory(resourceData, resourceData.previewer(request.user, item).ID())
			},
		)*/
	}
}

func (resource *Resource[T]) CreateWithLog(item *T, request *Request) error {
	return resource.data.CreateWithLog(item, request)
}

func (resourceData *resourceData) CreateWithLog(item any, request *Request) error {
	err := resourceData.Create(item)
	if err != nil {
		return err
	}

	if resourceData.app.search != nil {
		go func() {
			err := resourceData.saveSearchItem(item)
			if err != nil {
				resourceData.app.Log().Println(fmt.Errorf("%s", err))
			}
			resourceData.app.search.flush()
		}()
	}

	if resourceData.activityLog {
		err := resourceData.LogActivity(request.user, nil, item)
		if err != nil {
			return err
		}

	}
	return resourceData.updateCachedCount()

}

func (resource *Resource[T]) DeleteWithLog(item *T, request *Request) error {
	return resource.data.DeleteWithLog(item, request)
}

func (resourceData *resourceData) DeleteWithLog(item any, request *Request) error {
	if resourceData.activityLog {
		err := resourceData.LogActivity(request.user, item, nil)
		if err != nil {
			return err
		}
	}

	id := resourceData.previewer(request.user, item).ID()

	err := resourceData.Delete(id)
	if err != nil {
		return fmt.Errorf("can't delete item id '%d': %s", id, err)
	}

	if resourceData.app.search != nil {
		err = resourceData.app.search.deleteItem(resourceData, id)
		if err != nil {
			resourceData.app.Log().Println(fmt.Errorf("%s", err))
		}
		resourceData.app.search.flush()
	}

	resourceData.updateCachedCount()

	return nil
}

func (resourceData *resourceData) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, ValidationContext, error) {
	user := request.user
	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	beforeItem := resourceData.ID(id)
	if beforeItem == nil {
		return nil, nil, fmt.Errorf("can't get beforeitem with id %d: %s", id, err)
	}

	beforeVal := reflect.ValueOf(beforeItem).Elem()
	itemVal := beforeVal

	item := itemVal.Addr().Interface()

	err = resourceData.bindData(
		item, user, values,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("can't bind data (%d): %s", id, err)
	}

	stringableValues := resourceData.getItemStringEditableValues(item, user)
	var allValues url.Values = make(map[string][]string)
	for k, v := range stringableValues {
		allValues.Add(k, v)
	}

	vv := newValuesValidation(resourceData.app, user, allValues)
	for _, v := range resourceData.validations {
		v(vv)
	}
	if !vv.Valid() {
		return nil, vv, errValidation
	}

	err = resourceData.UpdateWithLog(item, request)
	if err != nil {
		return nil, nil, err
	}

	return item, vv, nil
}

func (resource *Resource[T]) UpdateWithLog(item *T, request *Request) error {
	return resource.data.UpdateWithLog(item, request)
}

func (resourceData *resourceData) UpdateWithLog(item any, request *Request) error {
	id := resourceData.previewer(request.user, item).ID()
	beforeItem := resourceData.ID(id)
	if beforeItem == nil {
		return errors.New("can't find before item")
	}

	err := resourceData.Update(item)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if resourceData.app.search != nil {
		go func() {
			err = resourceData.saveSearchItem(item)
			if err != nil {
				resourceData.app.Log().Println(fmt.Errorf("%s", err))
			}
			resourceData.app.search.flush()
		}()
	}

	if resourceData.activityLog {
		must(
			resourceData.LogActivity(request.user, beforeItem, item),
		)
	}

	return nil
}
