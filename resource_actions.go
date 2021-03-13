package prago

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

func initResourceActions(resource *Resource) {
	app := resource.app
	if resource.CanCreate == "" {
		resource.CanCreate = resource.CanEdit
	}
	if resource.CanDelete == "" {
		resource.CanDelete = resource.CanEdit
	}

	//list action
	resource.AddAction("").Permission(resource.CanView).Name(resource.HumanName).Handler(
		func(request Request) {
			user := request.GetUser()
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(app, user, request.Request().URL.Query())
				must(err)
				request.RenderJSON(listDataJSON)
				return
			}

			if request.Request().URL.Query().Get("_format") == "xlsx" {
				listData, err := resource.getListContent(app, user, request.Request().URL.Query())
				must(err)

				file := xlsx.NewFile()
				sheet, err := file.AddSheet("List 1")
				must(err)

				row := sheet.AddRow()
				columnsStr := request.Request().URL.Query().Get("_columns")
				if columnsStr == "" {
					columnsStr = resource.defaultVisibleFieldsStr()
				}
				columnsAr := strings.Split(columnsStr, ",")
				for _, v := range columnsAr {
					cell := row.AddCell()
					cell.SetValue(v)
				}

				for _, v1 := range listData.Rows {
					row := sheet.AddRow()
					for _, v2 := range v1.Items {
						cell := row.AddCell()
						if reflect.TypeOf(v2.OriginalValue) == reflect.TypeOf(time.Now()) {
							t := v2.OriginalValue.(time.Time)
							cell.SetString(t.Format("2006-01-02"))
						} else {
							cell.SetValue(v2.OriginalValue)
						}
					}
				}
				file.Write(request.Response())
				return
			}

			listData, err := resource.getListHeader(user)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			navigation := resource.getNavigation(user, "")
			navigation.Wide = true

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	)

	resource.AddAction("new").Permission(resource.CanCreate).Template("admin_form").Name(messages.GetNameFunction("admin_new")).DataSource(
		func(request Request) interface{} {
			user := request.GetUser()
			var item interface{}
			resource.newItem(&item)

			resource.bindData(&item, user, request.Request().URL.Query(), defaultEditabilityFilter)

			form, err := resource.getForm(item, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "../" + resource.ID
			form.AddSubmit("_submit", messages.Get(user.Locale, "admin_save"))
			form.AddCSRFToken(request)
			return form
		},
	)
	resource.AddAction("").Method("POST").Permission(resource.CanCreate).Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			var item interface{}
			resource.newItem(&item)

			form, err := resource.getForm(item, user)
			must(err)

			resource.bindData(item, user, request.Params(), form.getFilter())
			if resource.OrderFieldName != "" {
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

			if resource.ActivityLog {
				app.createNewActivityLog(*resource, user, item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_created"))
			request.Redirect(resource.GetItemURL(item, ""))
		},
	)

	resource.AddItemAction("").IsWide().Template("admin_views").Permission(resource.CanView).DataSource(
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

	resource.AddItemAction("edit").Name(messages.GetNameFunction("admin_edit")).Permission(resource.CanEdit).Template("admin_form").DataSource(
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

	resource.AddItemAction("edit").Method("POST").Permission(resource.CanEdit).Handler(
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
			if resource.ActivityLog {
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

			if resource.ActivityLog {
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

	resource.AddItemAction("delete").Permission(resource.CanDelete).Name(messages.GetNameFunction("admin_delete")).Template("admin_delete").DataSource(
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
			ret["delete_title"] = fmt.Sprintf("Chcete smazat položku %s?", itemName)
			ret["delete_title"] = messages.Get(user.Locale, "admin_delete_confirmation_name", itemName)
			return ret
		},
	)

	resource.AddItemAction("delete").Permission(resource.CanDelete).Method("POST").Handler(
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

			if resource.ActivityLog {
				app.createDeleteActivityLog(*resource, user, int64(id), item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_deleted"))
			request.Redirect(resource.getURL(""))
		},
	)

	if resource.PreviewURLFunction != nil {
		resource.AddItemAction("preview").Name(messages.GetNameFunction("admin_preview")).Handler(
			func(request Request) {
				var item interface{}
				resource.newItem(&item)
				must(app.Query().WhereIs("id", request.Params().Get("id")).Get(item))
				request.Redirect(
					resource.PreviewURLFunction(item),
				)
			},
		)
	}

	if resource.ActivityLog {
		resource.AddAction("history").Name(messages.GetNameFunction("admin_history")).Template("admin_history").Permission(resource.CanEdit).DataSource(
			func(request Request) interface{} {
				return app.getHistory(resource, 0)
			},
		)

		resource.AddItemAction("history").Name(messages.GetNameFunction("admin_history")).Permission(resource.CanEdit).Template("admin_history").DataSource(
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
