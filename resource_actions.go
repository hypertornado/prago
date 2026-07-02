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

	resource.initListAction()

	resource.formAction("new", func(form *Form, request *Request) {
		var item any = reflect.New(resource.typ).Interface()
		queryData := request.Request().URL.Query()
		for k, v := range resource.defaultValues {
			queryData.Set(k, v(request))
		}
		resource.bindData(item, request, queryData)
		form.initWithResourceItem(resource, item, request)
		submitItem := form.AddSubmit(messages.Get(request.Locale(), "create"))
		submitItem.Icon = iconAdd

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
				SetPreName(messages.Get(request.Locale(), "item_created")).
				Flash(request))
			vc.Redirect(resource.getItemURL(item, "", request))
		}
	}).Icon(iconAdd).setPriority(defaultHighPriority).StyleCreate().Permission(resource.canCreate).Name(func(locale string) string {
		return resource.newItemName(locale)
	})

	resource.itemActionUi("", func(item any, request *Request, pd *pageData) {
		if item == nil {
			renderErrorPage(request, 404)
			return
		}
		id := resource.previewer(request, item).ID()
		pd.BoxHeader = resource.getBoxHeader(id, item, request)
		pd.ViewFields = resource.getViewFields(id, item, request)
		pd.RelationViews = resource.getRelationViews(id, request)
	},
	).Icon("glyphicons-basic-588-book-open-text.svg").Name(messages.GetNameFunction("view")).setPriority(defaultHighPriority).Permission(resource.canView)

	resource.formItemAction(
		"edit",
		func(item any, form *Form, request *Request) {
			form.initWithResourceItem(resource, item, request)
			submitItem := form.AddSubmit(messages.Get(request.Locale(), "edit"))
			submitItem.Icon = iconEdit
		},
		func(_ any, vc FormValidation, request *Request) {
			params := request.Params()

			fieldsMap := getFieldsFilterMap(request.Request().PostForm.Get("_fields"))

			//fix browsers not sending empty checkbox values
			resource.addBoleanFalseValuesAsEmpty(params, fieldsMap)
			item, validation := resource.editItemWithLogAndValues(request, params)

			if validation.Valid() {
				resource.app.Notification(resource.previewer(request, item).Name()).
					SetImage(resource.previewer(request, item).ThumbnailURL()).
					SetPreName(messages.Get(request.Locale(), "item_edited")).
					Flash(request)
				vc.Redirect(resource.getItemURL(item, "", request))
			} else {
				vc.(*formValidation).validationData.Errors = validation.errors
			}
		},
	).Icon(iconEdit).setPriority(defaultHighPriority).StyleAccented().Name(messages.GetNameFunction("edit")).Permission(resource.canUpdate)

	resource.formItemAction(
		"delete",
		func(item any, form *Form, request *Request) {
			form.AddDeleteSubmit(messages.Get(request.Locale(), "delete"))
			preview := resource.previewer(request, item)
			itemName := preview.Name()
			form.ItemVersion = resource.currentItemVersion(preview.ID())
			form.Title = messages.Get(request.Locale(), "delete_confirmation_name", itemName)
		},
		func(item any, fv FormValidation, request *Request) {
			preview := resource.previewer(request, item)
			vc := resource.validateDelete(item, request)
			resource.validateConflict(request, vc, preview.ID())
			//TODO: its not working
			fmt.Println(vc.errors)
			fv.(*formValidation).validationData.Errors = vc.errors
			if vc.Valid() {
				must(resource.deleteWithLog(item, request))
				request.AddFlashMessage(messages.Get(request.Locale(), "item_deleted"))
				fv.Redirect(resource.getURL(""))
			}
		},
	).Icon(iconDelete).setPriority(-defaultHighPriority).StyleDestroy().Permission(resource.canDelete).Name(messages.GetNameFunction("delete"))

	resource.initDefaultResourceMultipleActions()

	if resource.previewFn != nil {
		resource.itemActionHandler("preview",
			func(item any, request *Request) {
				request.Redirect(
					resource.previewFn(item),
				)
			}).Icon("glyphicons-basic-52-eye.svg").setPriority(defaultHighPriority).Name(messages.GetNameFunction("preview"))
	}

	bindResourceExportCSV(resource)

	if resource.activityLog {
		resource.formAction("_history", func(form *Form, request *Request) {
			form.AddNumberInput("page", "Stránka").Value = "1"
			form.AddRelationMultiple("user", "Uživatel", "user")
			form.AutosubmitFirstTime = true
			form.AddSubmit("Zobrazit")

		}, func(vc FormValidation, request *Request) {
			table := resource.app.getHistoryTable(request, resource, 0, request.Param("page"))
			vc.AfterContent(table.ExecuteHTML())

		}).
			Icon(iconActivity).
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("history")).
			Permission(resource.canView)

		resource.
			formItemAction(
				"_history",
				func(item any, form *Form, request *Request) {
					form.AddNumberInput("page", "Stránka").Value = "1"
					form.AddRelationMultiple("user", "Uživatel", "user")
					form.AutosubmitFirstTime = true
					form.AddSubmit("Zobrazit")
				},
				func(item any, vc FormValidation, request *Request) {
					id := resource.previewer(request, item).ID()
					table := resource.app.getHistoryTable(request, resource, id, request.Param("page"))
					vc.AfterContent(table.ExecuteHTML())
				},
			).
			Icon(iconActivity).
			setPriority(defaultHighPriority).
			Name(messages.GetNameFunction("history")).
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

	err = resource.logActivity(userData, nil, item)
	if err != nil {
		return err
	}

	return resource.updateCachedCount()

}

func DeleteWithLog[T any](item *T, request *Request) error {
	resource := getResource[T](request.app)
	return resource.deleteWithLog(item, request)
}

func (resource *Resource) deleteWithLog(item any, request UserData) error {

	err := resource.logActivity(request, item, nil)
	if err != nil {
		return err
	}

	id := resource.previewer(request, item).ID()

	err = resource.delete(context.Background(), id)
	if err != nil {
		return fmt.Errorf("can't delete item id '%d': %s", id, err)
	}

	resource.updateCachedCount()

	return nil
}

func (resource *Resource) editItemWithLogAndValues(request *Request, values url.Values) (any, *itemValidation) {
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

	err = resource.bindData(
		item, user, values,
	)
	must(err)

	itemValidation := resource.validateUpdate(item, request)
	resource.validateConflict(request, itemValidation, int64(id))
	if !itemValidation.Valid() {
		return nil, itemValidation
	}

	err = resource.updateWithLog(item, request)
	must(err)
	return item, itemValidation
}

func (resource *Resource) addBoleanFalseValuesAsEmpty(values url.Values, fieldsMap map[string]bool) {
	for _, field := range resource.fields {
		if fieldsMap != nil && !fieldsMap[field.id] {
			continue
		}
		if field.typeID() == "bool" && !values.Has(field.id) {
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

	err := resource.update(request.r.Context(), item, nil)
	if err != nil {
		return fmt.Errorf("can't save item (%d): %s", id, err)
	}

	must(
		resource.logActivity(request, beforeItem, item),
	)

	return nil
}
