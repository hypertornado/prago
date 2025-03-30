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

	resource.action("list").Icon(iconTable).setPriority(defaultHighPriority).
		Permission(resource.canView).Name(messages.GetNameFunction("admin_list")).
		ui(func(request *Request, pd *pageData) {
			listData, err := resource.getListHeader(request)
			must(err)
			pd.List = &listData
		},
		)

	resource.formAction("new", func(form *Form, request *Request) {
		var item any = reflect.New(resource.typ).Interface()
		queryData := request.Request().URL.Query()
		for k, v := range resource.defaultValues {
			queryData.Set(k, v(request))
		}
		resource.bindData(item, request, queryData)
		form.initWithResourceItem(resource, item, request)
		form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
	}, func(vc FormValidation, request *Request) {
		var item any = reflect.New(resource.typ).Interface()
		resource.bindData(item, request, request.Params())

		itemValidation := resource.validateUpdate(item, request)

		vc.(*formValidation).validationData.Errors = itemValidation.errors

		if vc.Valid() {
			if resource.orderField != nil {
				count, _ := resource.query(request.Request().Context()).count()
				resource.setOrderPosition(item, count+1)
			}
			must(resource.createWithLog(item, request))

			preview := resource.previewer(request, item).Preview(nil)
			vc.(*formValidation).validationData.Data = preview

			must(resource.app.Notification(resource.previewer(request, item).Name()).
				SetImage(resource.previewer(request, item).ThumbnailURL()).
				SetPreName(messages.Get(request.Locale(), "admin_item_created")).
				Flash(request))
			vc.Redirect(resource.getItemURL(item, "", request))
		}
	}).Icon(iconAdd).setPriority(defaultHighPriority).Permission(resource.canCreate).Name(func(locale string) string {
		return resource.newItemName(locale)
	})

	resource.itemActionUi("", func(item any, request *Request, pd *pageData) {
		if item == nil {
			renderErrorPage(request, 404)
			return
		}
		pd.Views = resource.getViews(request.r.Context(), item, request)
	},
	).Icon("glyphicons-basic-588-book-open-text.svg").Name(messages.GetNameFunction("admin_view")).setPriority(defaultHighPriority).Permission(resource.canView)

	resource.formItemAction(
		"edit",
		func(item any, form *Form, request *Request) {
			form.initWithResourceItem(resource, item, request)
			form.AddSubmit(messages.Get(request.Locale(), "admin_save"))
		},
		func(_ any, vc FormValidation, request *Request) {
			item, validation := resource.editItemWithLogAndValues(request, request.Params())

			if validation.Valid() {
				resource.app.Notification(resource.previewer(request, item).Name()).
					SetImage(resource.previewer(request, item).ThumbnailURL()).
					SetPreName(messages.Get(request.Locale(), "admin_item_edited")).
					Flash(request)
				vc.Redirect(resource.getItemURL(item, "", request))
			} else {
				vc.(*formValidation).validationData.Errors = validation.errors
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
		func(item any, fv FormValidation, request *Request) {
			vc := resource.validateDelete(item, request)
			fv.(*formValidation).validationData.Errors = vc.errors
			if vc.Valid() {
				must(resource.deleteWithLog(item, request))
				request.AddFlashMessage(messages.Get(request.Locale(), "admin_item_deleted"))
				fv.Redirect(resource.getURL(""))
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
			vc.AfterContent(table.ExecuteHTML())

		}).
			Icon(iconActivity).
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
					vc.AfterContent(table.ExecuteHTML())
				},
			).
			Icon(iconActivity).
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("admin_history")).
			Permission(resource.canView)
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

func (resource *Resource) editItemWithLogAndValues(request *Request, values url.Values) (interface{}, *itemValidation) {
	user := request
	id, err := strconv.Atoi(values.Get("id"))
	must(err)

	beforeItem := resource.query(request.r.Context()).ID(id)
	if beforeItem == nil {
		panic(fmt.Sprintf("can't get beforeitem with id %d: %s", id, err))
	}

	beforeVal := reflect.ValueOf(beforeItem).Elem()
	itemVal := beforeVal

	item := itemVal.Addr().Interface()

	resource.addBoleanFalseValuesAsEmpty(values)
	err = resource.bindData(
		item, user, values,
	)
	must(err)

	itemValidation := resource.validateUpdate(item, request)
	if !itemValidation.Valid() {
		return nil, itemValidation
	}

	err = resource.updateWithLog(item, request)
	must(err)
	return item, itemValidation
}

func (resource *Resource) addBoleanFalseValuesAsEmpty(values url.Values) {
	for _, field := range resource.fields {
		if field.typ.Kind() == reflect.Bool && !values.Has(field.id) {
			values.Set(field.id, "")
		}
	}
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
