package prago

import (
	"encoding/json"
	"fmt"
	"strconv"
)

func initResourceActions(resource *Resource) {
	app := resource.app
	if resource.canCreate == "" {
		resource.canCreate = resource.canEdit
	}
	if resource.canDelete == "" {
		resource.canDelete = resource.canEdit
	}

	//list action
	resource.Action("").Permission(resource.canView).Name(resource.name).IsWide().Template("admin_list").DataSource(
		func(request Request) interface{} {
			listData, err := resource.getListHeader(request.GetUser())
			must(err)
			return listData
		},
	)

	resource.Action("new").Permission(resource.canCreate).Template("admin_form").Name(messages.GetNameFunction("admin_new")).DataSource(
		func(request Request) interface{} {
			user := request.GetUser()
			var item interface{}
			resource.newItem(&item)

			resource.bindData(&item, user, request.Request().URL.Query(), defaultEditabilityFilter)

			form, err := resource.getForm(item, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "../" + resource.id
			form.AddSubmit("_submit", messages.Get(user.Locale, "admin_save"))
			form.AddCSRFToken(request)
			return form
		},
	)
	resource.Action("").Method("POST").Permission(resource.canCreate).Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			var item interface{}
			resource.newItem(&item)

			form, err := resource.getForm(item, user)
			must(err)

			resource.bindData(item, user, request.Params(), form.getFilter())
			if resource.orderFieldName != "" {
				resource.setOrderPosition(&item, resource.count()+1)
			}
			must(app.Create(item))

			if app.search != nil {
				err = app.search.saveItem(resource, item)
				if err != nil {
					app.Log().Println(fmt.Errorf("%s", err))
				}
				app.search.flush()
			}

			if resource.activityLog {
				app.createNewActivityLog(*resource, user, item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_created"))
			request.Redirect(resource.getItemURL(item, ""))
		},
	)

	resource.ItemAction("").IsWide().Template("admin_views").Permission(resource.canView).DataSource(
		func(request Request) interface{} {
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

			return resource.getViews(id, item, request.GetUser())
		},
	)

	resource.ItemAction("edit").Name(messages.GetNameFunction("admin_edit")).Permission(resource.canEdit).Template("admin_form").DataSource(
		func(request Request) interface{} {
			user := request.GetUser()
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = app.Query().WhereIs("id", int64(id)).Get(item)
			must(err)

			form, err := resource.getForm(item, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "edit"
			form.AddSubmit("_submit", messages.Get(user.Locale, "admin_save"))
			form.AddCSRFToken(request)
			return form
		},
	)

	resource.ItemAction("edit").Method("POST").Permission(resource.canEdit).Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			must(app.Query().WhereIs("id", int64(id)).Get(item))

			form, err := resource.getForm(item, user)
			must(err)

			var beforeData []byte
			if resource.activityLog {
				beforeData, err = json.Marshal(item)
				must(err)
			}

			must(
				resource.bindData(
					item, user, request.Params(), form.getFilter(),
				),
			)
			must(app.Save(item))

			if app.search != nil {
				err = app.search.saveItem(resource, item)
				if err != nil {
					app.Log().Println(fmt.Errorf("%s", err))
				}
				app.search.flush()
			}

			if resource.activityLog {
				afterData, err := json.Marshal(item)
				if err != nil {
					panic(err)
				}

				app.createEditActivityLog(*resource, user, int64(id), beforeData, afterData)
			}

			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_edited"))
			request.Redirect(resource.getURL(fmt.Sprintf("%d", id)))
		},
	)

	resource.ItemAction("delete").Permission(resource.canDelete).Name(messages.GetNameFunction("admin_delete")).Template("admin_delete").DataSource(
		func(request Request) interface{} {
			user := request.GetUser()
			ret := map[string]interface{}{}
			form := newForm()
			form.Method = "POST"
			form.AddCSRFToken(request)
			form.AddDeleteSubmit("send", messages.Get(user.Locale, "admin_delete"))
			ret["form"] = form

			var item interface{}
			resource.newItem(&item)
			must(app.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			itemName := getItemName(item)
			ret["delete_title"] = messages.Get(user.Locale, "admin_delete_confirmation_name", itemName)
			return ret
		},
	)

	resource.ItemAction("delete").Permission(resource.canDelete).Method("POST").Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			_, err = app.Query().WhereIs("id", int64(id)).Delete(item)
			must(err)

			if app.search != nil {
				err = app.search.deleteItem(resource, int64(id))
				if err != nil {
					app.Log().Println(fmt.Errorf("%s", err))
				}
				app.search.flush()
			}

			if resource.activityLog {
				app.createDeleteActivityLog(*resource, user, int64(id), item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_deleted"))
			request.Redirect(resource.getURL(""))
		},
	)

	if resource.previewURL != nil {
		resource.ItemAction("preview").Name(messages.GetNameFunction("admin_preview")).Handler(
			func(request Request) {
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
		resource.Action("history").Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.canEdit).DataSource(
			func(request Request) interface{} {
				return app.getHistory(resource, 0)
			},
		)

		resource.ItemAction("history").Name(messages.GetNameFunction("admin_history")).Permission(resource.canEdit).Template("admin_history").DataSource(
			func(request Request) interface{} {
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
