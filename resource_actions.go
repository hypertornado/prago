package prago

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

const defaultHighPriority int64 = 1000

func (resourceData *Resource) initDefaultResourceActions() {

	icon := resourceData.icon
	if icon == "" {
		icon = iconBoard
	}

	resourceData.action("").Icon(icon).setPriority(defaultHighPriority).
		Permission(resourceData.canView).Name(resourceData.pluralName).ui(
		func(request *Request, pd *pageData) {
			pd.BoardView = resourceData.resourceBoard.boardView(request)
		})

	resourceData.action("list").Icon(iconTable).setPriority(defaultHighPriority).Permission(resourceData.canView).Name(messages.GetNameFunction("admin_list")).
		ui(func(request *Request, pd *pageData) {
			listData, err := resourceData.getListHeader(request)
			must(err)
			pd.List = &listData
		},
		)

	resourceData.formAction("new", func(form *Form, request *Request) {
		var item interface{} = reflect.New(resourceData.typ).Interface()
		resourceData.bindData(item, request, request.Request().URL.Query())
		resourceData.addFormItems(item, request, form)
		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(vc ValidationContext) {
		for _, v := range resourceData.validations {
			v(vc)
		}
		request := vc.Request()
		if vc.Valid() {
			var item interface{} = reflect.New(resourceData.typ).Interface()
			resourceData.bindData(item, request, request.Params())
			if resourceData.orderField != nil {
				count, _ := resourceData.query(vc.Context()).count()
				resourceData.setOrderPosition(item, count+1)
			}
			must(resourceData.createWithLog(item, request))

			resourceData.app.Notification(resourceData.previewer(request, item).Name()).
				SetImage(resourceData.previewer(request, item).ThumbnailURL(vc.Context())).
				SetPreName(messages.Get(request.Locale(), "admin_item_created")).
				Flash(request)
			vc.Validation().RedirectionLocaliton = resourceData.getItemURL(item, "", request)
		}
	}).Icon(iconAdd).setPriority(defaultHighPriority).Permission(resourceData.canCreate).Name(messages.GetNameFunction("admin_new"))

	resourceData.itemActionUi("", func(item any, request *Request, pd *pageData) {
		if item == nil {
			renderErrorPage(request, 404)
			return
		}
		pd.Views = resourceData.getViews(request.r.Context(), item, request)
	},
	).Icon("glyphicons-basic-588-book-open-text.svg").setPriority(defaultHighPriority).Permission(resourceData.canView)

	resourceData.formItemAction(
		"edit",
		func(item any, form *Form, request *Request) {
			resourceData.addFormItems(item, request, form)
			form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
		},
		func(_ any, vc ValidationContext) {
			request := vc.Request()
			params := request.Params()

			resourceData.fixBooleanParams(vc.Request(), params)

			item, validation, err := resourceData.editItemWithLogAndValues(request, params)
			if err != nil && err != errValidation {
				panic(err)
			}

			if validation.Valid() {
				user := request
				id, err := strconv.Atoi(request.Param("id"))
				must(err)

				resourceData.app.Notification(resourceData.previewer(user, item).Name()).
					SetImage(resourceData.previewer(request, item).ThumbnailURL(vc.Context())).
					SetPreName(messages.Get(request.Locale(), "admin_item_edited")).
					Flash(request)

				vc.Validation().RedirectionLocaliton = resourceData.getURL(fmt.Sprintf("%d", id))
			} else {
				//TODO: ugly hack with copying two validation contexts
				vc.Validation().Errors = validation.Validation().Errors
				vc.Validation().ItemErrors = validation.Validation().ItemErrors
			}
		},
	).Icon("glyphicons-basic-31-pencil.svg").setPriority(defaultHighPriority).Name(messages.GetNameFunction("admin_edit")).Permission(resourceData.canUpdate)

	resourceData.formItemAction(
		"delete",
		func(item any, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.Locale(), "admin_delete"))
			itemName := resourceData.previewer(request, item).Name()
			form.Title = messages.Get(request.Locale(), "admin_delete_confirmation_name", itemName)
		},
		func(item any, vc ValidationContext) {
			for _, v := range resourceData.deleteValidations {
				v(vc)
			}
			if vc.Valid() {
				must(resourceData.deleteWithLog(item, vc.Request()))
				vc.Request().AddFlashMessage(messages.Get(vc.Request().Locale(), "admin_item_deleted"))
				vc.Validation().RedirectionLocaliton = resourceData.getURL("")
			}
		},
	).Icon("glyphicons-basic-17-bin.svg").setPriority(-defaultHighPriority).Permission(resourceData.canDelete).Name(messages.GetNameFunction("admin_delete"))

	if resourceData.previewFn != nil {
		resourceData.itemActionHandler("preview",
			func(item any, request *Request) {
				request.Redirect(
					resourceData.previewFn(item),
				)
			}).Icon("glyphicons-basic-52-eye.svg").setPriority(defaultHighPriority).Name(messages.GetNameFunction("admin_preview"))
	}

	bindResourceExportCSV(resourceData)

	if resourceData.activityLog {
		resourceData.formAction("history", func(f *Form, r *Request) {
			f.AddTextInput("page", "Stránka").Value = "1"
			f.AutosubmitFirstTime = true

		}, func(vc ValidationContext) {
			table := resourceData.app.getHistoryTable(vc.Request(), resourceData, 0, vc.GetValue("page"))
			vc.Validation().AfterContent = table.ExecuteHTML()

		}).
			Icon("glyphicons-basic-58-history.svg").
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("admin_history")).
			Permission(resourceData.canUpdate)

		resourceData.
			formItemAction(
				"history",
				func(item any, f *Form, r *Request) {
					f.AddTextInput("page", "Stránka").Value = "1"
					f.AddSubmit("Zobrazit")
					f.AutosubmitFirstTime = true
				},
				func(item any, vc ValidationContext) {
					id := resourceData.previewer(vc.Request(), item).ID()
					table := resourceData.app.getHistoryTable(vc.Request(), resourceData, id, vc.GetValue("page"))
					vc.Validation().AfterContent = table.ExecuteHTML()
				},
			).
			Icon("glyphicons-basic-58-history.svg").
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("admin_history")).
			Permission(resourceData.canUpdate)
	}
}

func CreateWithLog[T any](item *T, request *Request) error {
	resource := GetResource[T](request.app)
	return resource.createWithLog(item, request)
}

func (resourceData *Resource) createWithLog(item any, request *Request) error {
	err := resourceData.create(request.r.Context(), item)
	if err != nil {
		return err
	}

	if resourceData.activityLog {
		err := resourceData.logActivity(request, nil, item)
		if err != nil {
			return err
		}

	}
	return resourceData.updateCachedCount(request.r.Context())

}

func DeleteWithLog[T any](item *T, request *Request) error {
	resource := GetResource[T](request.app)
	return resource.deleteWithLog(item, request)
}

func (resourceData *Resource) deleteWithLog(item any, request *Request) error {
	if resourceData.activityLog {
		err := resourceData.logActivity(request, item, nil)
		if err != nil {
			return err
		}
	}

	id := resourceData.previewer(request, item).ID()

	err := resourceData.delete(request.r.Context(), id)
	if err != nil {
		return fmt.Errorf("can't delete item id '%d': %s", id, err)
	}

	resourceData.updateCachedCount(request.r.Context())

	return nil
}

func (resourceData *Resource) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, ValidationContext, error) {
	user := request
	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	beforeItem := resourceData.query(request.r.Context()).ID(id)
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

	vv := newValuesValidation(request.r.Context(), resourceData.app, user, allValues)
	for _, v := range resourceData.validations {
		v(vv)
	}
	if !vv.Valid() {
		return nil, vv, errValidation
	}

	err = resourceData.updateWithLog(item, request)
	if err != nil {
		return nil, nil, err
	}

	return item, vv, nil
}

func UpdateWithLog[T any](item *T, request *Request) error {
	resource := GetResource[T](request.app)
	return resource.updateWithLog(item, request)
}

func (resourceData *Resource) updateWithLog(item any, request *Request) error {
	id := resourceData.previewer(request, item).ID()

	beforeItem := resourceData.query(request.r.Context()).ID(id)
	if beforeItem == nil {
		return errors.New("can't find before item")
	}

	err := resourceData.update(request.r.Context(), item)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if resourceData.activityLog {
		must(
			resourceData.logActivity(request, beforeItem, item),
		)
	}

	return nil
}
