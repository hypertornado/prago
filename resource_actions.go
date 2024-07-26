package prago

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
)

const defaultHighPriority int64 = 1000

func (resource *Resource) initDefaultResourceActions() {

	icon := resource.icon
	if icon == "" {
		icon = iconBoard
	}

	resource.action("").Icon(icon).setPriority(defaultHighPriority).
		Permission(resource.canView).Name(resource.pluralName).ui(
		func(request *Request, pd *pageData) {
			pd.BoardView = resource.resourceBoard.boardView(request)
		})

	resource.action("list").Icon(iconTable).setPriority(defaultHighPriority).Permission(resource.canView).Name(messages.GetNameFunction("admin_list")).
		ui(func(request *Request, pd *pageData) {
			listData, err := resource.getListHeader(request)
			must(err)
			pd.List = &listData
		},
		)

	resource.formAction("new", func(form *Form, request *Request) {
		var item interface{} = reflect.New(resource.typ).Interface()
		queryData := request.Request().URL.Query()
		for k, v := range resource.defaultValues {
			queryData.Set(k, v(request))
		}
		resource.bindData(item, request, queryData)
		form.addResourceItems(resource, item, request)
		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(vc FormValidation, request *Request) {
		var item interface{} = reflect.New(resource.typ).Interface()
		resource.bindData(item, request, request.Params())
		for _, v := range resource.updateValidations {
			v(item, vc, request)
		}
		if vc.Valid() {
			if resource.orderField != nil {
				count, _ := resource.query(request.Request().Context()).count()
				resource.setOrderPosition(item, count+1)
			}
			must(resource.createWithLog(item, request))

			must(resource.app.Notification(resource.previewer(request, item).Name()).
				SetImage(resource.previewer(request, item).ThumbnailURL()).
				SetPreName(messages.Get(request.Locale(), "admin_item_created")).
				Flash(request))
			vc.Validation().RedirectionLocation = resource.getItemURL(item, "", request)
		}
	}).Icon(iconAdd).setPriority(defaultHighPriority).Permission(resource.canCreate).Name(messages.GetNameFunction("admin_new"))

	resource.itemActionUi("", func(item any, request *Request, pd *pageData) {
		if item == nil {
			renderErrorPage(request, 404)
			return
		}
		pd.Views = resource.getViews(request.r.Context(), item, request)
	},
	).Icon("glyphicons-basic-588-book-open-text.svg").setPriority(defaultHighPriority).Permission(resource.canView)

	resource.formItemAction(
		"edit",
		func(item any, form *Form, request *Request) {
			form.addResourceItems(resource, item, request)
			form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
		},
		func(_ any, vc FormValidation, request *Request) {
			params := request.Params()
			resource.fixBooleanParams(request, params)

			item, validation, err := resource.editItemWithLogAndValues(request, params)
			if err != nil && err != errValidation {
				panic(err)
			}

			if validation.Valid() {
				user := request
				id, err := strconv.Atoi(request.Param("id"))
				must(err)

				resource.app.Notification(resource.previewer(user, item).Name()).
					SetImage(resource.previewer(request, item).ThumbnailURL()).
					SetPreName(messages.Get(request.Locale(), "admin_item_edited")).
					Flash(request)

				vc.Validation().RedirectionLocation = resource.getURL(fmt.Sprintf("%d", id))
			} else {
				//TODO: ugly hack with copying two validation contexts
				vc.Validation().Errors = validation.Validation().Errors
				vc.Validation().ItemErrors = validation.Validation().ItemErrors
			}
		},
	).Icon(iconEdit).setPriority(defaultHighPriority).Name(messages.GetNameFunction("admin_edit")).Permission(resource.canUpdate)

	resource.formItemAction(
		"delete",
		func(item any, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.Locale(), "admin_delete"))
			itemName := resource.previewer(request, item).Name()
			form.Title = messages.Get(request.Locale(), "admin_delete_confirmation_name", itemName)
		},
		func(item any, vc FormValidation, request *Request) {
			for _, v := range resource.deleteValidations {
				v(item, vc, request)
			}
			if vc.Valid() {
				must(resource.deleteWithLog(item, request))
				request.AddFlashMessage(messages.Get(request.Locale(), "admin_item_deleted"))
				vc.Validation().RedirectionLocation = resource.getURL("")
			}
		},
	).Icon(iconDelete).setPriority(-defaultHighPriority).Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete"))

	if resource.previewFn != nil {
		resource.itemActionHandler("preview",
			func(item any, request *Request) {
				request.Redirect(
					resource.previewFn(item),
				)
			}).Icon("glyphicons-basic-52-eye.svg").setPriority(defaultHighPriority).Name(messages.GetNameFunction("admin_preview"))
	}

	bindResourceExportCSV(resource)

	if resource.activityLog {
		resource.formAction("history", func(f *Form, r *Request) {
			f.AddTextInput("page", "Stránka").Value = "1"
			f.AutosubmitFirstTime = true

		}, func(vc FormValidation, request *Request) {
			table := resource.app.getHistoryTable(request, resource, 0, request.Param("page"))
			vc.Validation().AfterContent = table.ExecuteHTML()

		}).
			Icon("glyphicons-basic-58-history.svg").
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("admin_history")).
			Permission(resource.canUpdate)

		resource.
			formItemAction(
				"history",
				func(item any, f *Form, r *Request) {
					f.AddTextInput("page", "Stránka").Value = "1"
					f.AddSubmit("Zobrazit")
					f.AutosubmitFirstTime = true
				},
				func(item any, vc FormValidation, request *Request) {
					id := resource.previewer(request, item).ID()
					table := resource.app.getHistoryTable(request, resource, id, request.Param("page"))
					vc.Validation().AfterContent = table.ExecuteHTML()
				},
			).
			Icon("glyphicons-basic-58-history.svg").
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("admin_history")).
			Permission(resource.canUpdate)
	}
}

func CreateWithLog[T any](item *T, request *Request) error {
	resource := getResource[T](request.app)
	return resource.createWithLog(item, request)
}

func (resource *Resource) createWithLog(item any, userData UserData) error {
	err := resource.create(context.Background(), item)
	if err != nil {
		return err
	}

	if resource.activityLog {
		err := resource.logActivity(userData, nil, item)
		if err != nil {
			return err
		}

	}
	return resource.updateCachedCount()

}

func DeleteWithLog[T any](item *T, request *Request) error {
	resource := getResource[T](request.app)
	return resource.deleteWithLog(item, request)
}

func (resource *Resource) deleteWithLog(item any, request UserData) error {
	if resource.activityLog {
		err := resource.logActivity(request, item, nil)
		if err != nil {
			return err
		}
	}

	id := resource.previewer(request, item).ID()

	err := resource.delete(context.Background(), id)
	if err != nil {
		return fmt.Errorf("can't delete item id '%d': %s", id, err)
	}

	resource.updateCachedCount()

	return nil
}

func (resource *Resource) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, ItemValidation, error) {
	user := request
	id, err := strconv.Atoi(values.Get("id"))
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse id %d: %s", id, err)
	}

	beforeItem := resource.query(request.r.Context()).ID(id)
	if beforeItem == nil {
		return nil, nil, fmt.Errorf("can't get beforeitem with id %d: %s", id, err)
	}

	beforeVal := reflect.ValueOf(beforeItem).Elem()
	itemVal := beforeVal

	item := itemVal.Addr().Interface()

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

	vv := newValuesValidation(request.r.Context(), resource.app, user, allValues)
	for _, v := range resource.updateValidations {
		v(item, vv, request)
	}
	if !vv.Valid() {
		return nil, vv, errValidation
	}

	err = resource.updateWithLog(item, request)
	if err != nil {
		return nil, nil, err
	}

	return item, vv, nil
}

func UpdateWithLog[T any](item *T, request *Request) error {
	resource := getResource[T](request.app)
	return resource.updateWithLog(item, request)
}

func (resource *Resource) updateWithLog(item any, request *Request) error {
	id := resource.previewer(request, item).ID()

	beforeItem := resource.query(request.r.Context()).ID(id)
	if beforeItem == nil {
		return errors.New("can't find before item")
	}

	err := resource.update(request.r.Context(), item)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	if resource.activityLog {
		must(
			resource.logActivity(request, beforeItem, item),
		)
	}

	return nil
}
