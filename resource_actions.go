package prago

import (
	"fmt"
	"net/url"
	"strconv"
)

func initDefaultResourceActions(resource *Resource) {
	app := resource.app

	//list action
	resource.Action("").priority().Permission(resource.canView).Name(resource.name).IsWide().Template("admin_list").DataSource(
		func(request *Request) interface{} {
			listData, err := resource.getListHeader(request.user)
			must(err)
			return listData
		},
	)

	resource.Action("new").priority().Permission(resource.canCreate).Template("admin_form").Name(messages.GetNameFunction("admin_new")).DataSource(
		func(request *Request) interface{} {
			var item interface{}
			resource.newItem(&item)

			resource.bindData(&item, request.user, request.Request().URL.Query(), nil)

			form, err := resource.getForm(item, request.user)
			must(err)

			form.Classes = append(form.Classes, "prago_form")
			form.Action = "../" + resource.id
			form.AddSubmit("_submit", messages.Get(request.user.Locale, "admin_save"))
			form.AddCSRFToken(request)
			return form
		},
	)
	resource.Action("").Method("POST").Permission(resource.canCreate).Handler(
		func(request *Request) {
			validateCSRF(request)
			var item interface{}
			resource.newItem(&item)

			resource.bindData(item, request.user, request.Params(), nil)
			if resource.orderField != nil {
				resource.setOrderPosition(&item, resource.count()+1)
			}
			must(app.Create(item))

			if app.search != nil {
				err := app.search.saveItem(resource, item)
				if err != nil {
					app.Log().Println(fmt.Errorf("%s", err))
				}
				app.search.flush()
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
			request.Redirect(resource.getItemURL(item, ""))
		},
	)

	resource.ItemAction("").priority().IsWide().Template("admin_views").Permission(resource.canView).DataSource(
		func(request *Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = app.Query().WhereIs("id", int64(id)).Get(item)
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

	resource.ItemAction("edit").priority().Name(messages.GetNameFunction("admin_edit")).Permission(resource.canEdit).Template("admin_form").DataSource(
		func(request *Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = app.Query().WhereIs("id", int64(id)).Get(item)
			must(err)

			form, err := resource.getForm(item, request.user)
			must(err)

			form.Classes = append(form.Classes, "prago_form")
			form.Action = "edit"
			form.AddSubmit("_submit", messages.Get(request.user.Locale, "admin_save"))
			form.AddCSRFToken(request)
			return form
		},
	)

	resource.ItemAction("edit").Method("POST").Permission(resource.canEdit).Handler(
		func(request *Request) {
			user := request.user
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			items, err := resource.editItemsWithLog(user, []int64{int64(id)}, request.Params(), nil)
			must(err)

			item := items[0]

			app.Notification(getItemName(item)).
				SetImage(app.getItemImage(item)).
				SetPreName(messages.Get(user.Locale, "admin_item_edited")).
				Flash(request)

			request.Redirect(resource.getURL(fmt.Sprintf("%d", id)))
		},
	)

	resource.ItemAction("delete").priority().Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete")).Template("admin_delete").DataSource(
		func(request *Request) interface{} {
			ret := map[string]interface{}{}
			form := newForm()
			form.Method = "POST"
			form.AddCSRFToken(request)
			form.AddDeleteSubmit("send", messages.Get(request.user.Locale, "admin_delete"))
			ret["form"] = form

			var item interface{}
			resource.newItem(&item)
			must(app.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			itemName := getItemName(item)
			ret["delete_title"] = messages.Get(request.user.Locale, "admin_delete_confirmation_name", itemName)
			return ret
		},
	)

	resource.ItemAction("delete").Permission(resource.canDelete).Method("POST").Handler(
		func(request *Request) {
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			must(resource.deleteItemWithLog(request.user, int64(id)))

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(request.user.Locale, "admin_item_deleted"))
			request.Redirect(resource.getURL(""))
		},
	)

	if resource.previewURL != nil {
		resource.ItemAction("preview").priority().Name(messages.GetNameFunction("admin_preview")).Handler(
			func(request *Request) {
				var item interface{}
				resource.newItem(&item)
				must(app.Query().WhereIs("id", request.Params().Get("id")).Get(item))
				request.Redirect(
					resource.previewURL(item),
				)
			},
		)
	}

	if resource.activityLog {
		resource.Action("history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.canEdit).DataSource(
			func(request *Request) interface{} {
				return app.getHistory(resource, 0)
			},
		)

		resource.ItemAction("history").priority().IsWide().Name(messages.GetNameFunction("admin_history")).Permission(resource.canEdit).Template("admin_history").DataSource(
			func(request *Request) interface{} {
				id, err := strconv.Atoi(request.Params().Get("id"))
				must(err)

				var item interface{}
				resource.newItem(&item)
				must(app.Query().WhereIs("id", int64(id)).Get(item))

				return app.getHistory(resource, int64(id))
			},
		)

	}
}

func (resource *Resource) deleteItemWithLog(user *user, id int64) error {

	var beforeItem interface{}
	resource.newItem(&beforeItem)
	err := resource.app.Query().WhereIs("id", id).Get(beforeItem)
	if err != nil {
		return fmt.Errorf("can't find item for deletion id '%d': %s", id, err)
	}

	var item interface{}
	resource.newItem(&item)
	_, err = resource.app.Query().WhereIs("id", id).Delete(item)
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

	if resource.activityLog {
		return resource.app.LogActivity("delete", user.ID, resource.id, id, beforeItem, nil)
	}

	return nil
}

func (resource *Resource) editItemsWithLog(user *user, ids []int64, values url.Values, bindedFieldIDs map[string]bool) ([]interface{}, error) {

	var items []interface{}
	app := resource.app

	for _, id := range ids {
		//TODO: remove this ugly hack and copy values via reflect package
		var beforeItem, item interface{}
		resource.newItem(&beforeItem)
		err := app.Query().WhereIs("id", id).Get(beforeItem)
		if err != nil {
			return nil, fmt.Errorf("can't get beforeitem with id %d: %s", id, err)
		}

		resource.newItem(&item)
		err = app.Query().WhereIs("id", id).Get(item)
		if err != nil {
			return nil, fmt.Errorf("can't get item with id %d: %s", id, err)
		}

		err = resource.bindData(
			item, user, values, bindedFieldIDs,
		)
		if err != nil {
			return nil, fmt.Errorf("can't bind data (%d): %s", id, err)
		}

		err = app.Save(item)
		if err != nil {
			return nil, fmt.Errorf("can't save item (%d): %s", id, err)
		}
		items = append(items, item)

		if app.search != nil {
			err = app.search.saveItem(resource, item)
			if err != nil {
				app.Log().Println(fmt.Errorf("%s", err))
			}
			app.search.flush()
		}

		if resource.activityLog {
			must(
				app.LogActivity("edit", user.ID, resource.id, int64(id), beforeItem, item),
			)

		}
	}

	return items, nil

}
