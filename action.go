package prago

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hypertornado/prago/utils"
	"github.com/tealeg/xlsx"
)

type buttonData struct {
	Name   string
	URL    string
	Params map[string]string
}

//Action represents action
type Action struct {
	name       func(string) string
	permission Permission
	method     string
	url        string
	handler    func(Request)
	template   string
	dataSource func(Request) interface{}

	app          *App
	resource     *Resource
	isItemAction bool
	isWide       bool
}

func (app *App) bindAllActions() {
	for _, v := range app.rootActions {
		err := v.bindAction()
		if err != nil {
			panic(fmt.Sprintf("error while binding root action %s %s: %s", v.method, v.name("en"), err))
		}
	}

	for _, resource := range app.resources {
		for _, v := range resource.actions {
			err := v.bindAction()
			if err != nil {
				panic(fmt.Sprintf("error while binding resource %s action %s %s: %s", resource.ID, v.method, v.name("en"), err))
			}
		}
		for _, v := range resource.itemActions {
			err := v.bindAction()
			if err != nil {
				panic(fmt.Sprintf("error while binding item resource %s action %s %s: %s", resource.ID, v.method, v.name("en"), err))
			}
		}

	}

}

func newAction(app *App, url string) *Action {
	return &Action{
		name:       Unlocalized(url),
		permission: "",
		method:     "GET",
		url:        url,
		app:        app,
	}
}

func (app *App) AddAction(url string) *Action {
	action := newAction(app, url)
	app.rootActions = append(app.rootActions, action)
	return action
}

func (resource *Resource) AddAction(url string) *Action {
	action := newAction(resource.App, url)
	action.resource = resource
	action.permission = resource.CanView
	resource.actions = append(resource.actions, action)
	return action
}

func (resource *Resource) AddItemAction(url string) *Action {
	action := newAction(resource.App, url)
	action.resource = resource
	action.isItemAction = true
	action.permission = resource.CanView
	resource.itemActions = append(resource.itemActions, action)
	return action
}

func (action *Action) Name(name func(string) string) *Action {
	action.name = name
	return action
}

func (action *Action) Permission(permission Permission) *Action {
	action.permission = permission
	return action
}

func (action *Action) Method(method string) *Action {
	action.method = strings.ToUpper(method)
	return action
}

func (action *Action) Template(template string) *Action {
	action.template = template
	return action
}

func (action *Action) Handler(handler func(Request)) *Action {
	action.handler = handler
	return action
}

func (action *Action) DataSource(dataSource func(Request) interface{}) *Action {
	action.dataSource = dataSource
	return action
}

func (action *Action) IsWide() *Action {
	action.isWide = true
	return action
}

/*
func (ra *Action) getName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.URL
}*/

func actionList(resource *Resource) *Action {

	return resource.AddAction("").Permission(resource.CanView).Name(resource.HumanName).Handler(
		func(request Request) {
			user := request.GetUser()
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(resource.App, user, request.Request().URL.Query())
				if err != nil {
					panic(err)
				}
				request.RenderJSON(listDataJSON)
				return
			}

			if request.Request().URL.Query().Get("_format") == "xlsx" {
				listData, err := resource.getListContent(resource.App, user, request.Request().URL.Query())
				if err != nil {
					panic(err)
				}

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

}

func actionNew(resource *Resource, permission Permission) *Action {
	return resource.AddAction("new").Permission(permission).Name(messages.GetNameFunction("admin_new")).Handler(
		func(request Request) {
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

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.getNavigation(user, "new"),
				PageTemplate: "admin_form",
				PageData:     form,
			})
		},
	)
}

func actionCreate(resource *Resource, permission Permission) *Action {

	return resource.AddAction("").Method("POST").Permission(permission).Handler(
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
			must(resource.App.Create(item))

			if resource.App.search != nil {
				err = resource.App.search.saveItem(resource, item)
				if err != nil {
					resource.App.Log().Println(fmt.Errorf("%s", err))
				}
				resource.App.search.flush()
			}

			if resource.ActivityLog {
				resource.App.createNewActivityLog(*resource, user, item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_created"))
			request.Redirect(resource.GetItemURL(item, ""))
		},
	)

}

func actionView(resource *Resource) *Action {

	return resource.AddItemAction("").IsWide().Template("admin_views").Permission(resource.CanView).DataSource(
		func(request Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = resource.App.Query().WhereIs("id", int64(id)).Get(item)
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
}

func actionEdit(resource *Resource, permission Permission) *Action {

	return resource.AddItemAction("edit").Name(messages.GetNameFunction("admin_edit")).Permission(permission).Template("admin_form").DataSource(
		func(request Request) interface{} {
			user := request.GetUser()
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = resource.App.Query().WhereIs("id", int64(id)).Get(item)
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

}

func actionUpdate(resource *Resource, permission Permission) *Action {

	return resource.AddItemAction("edit").Method("POST").Permission(permission).Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			must(resource.App.Query().WhereIs("id", int64(id)).Get(item))

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
			must(resource.App.Save(item))

			if resource.App.search != nil {
				err = resource.App.search.saveItem(resource, item)
				if err != nil {
					resource.App.Log().Println(fmt.Errorf("%s", err))
				}
				resource.App.search.flush()
			}

			if resource.ActivityLog {
				afterData, err := json.Marshal(item)
				if err != nil {
					panic(err)
				}

				resource.App.createEditActivityLog(*resource, user, int64(id), beforeData, afterData)
			}

			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_edited"))
			request.Redirect(resource.getURL(fmt.Sprintf("%d", id)))
		},
	)

}

func actionHistory(resource *Resource, permission Permission) *Action {

	return resource.AddAction("history").Name(messages.GetNameFunction("admin_history")).Permission(permission).Handler(
		func(request Request) {
			user := request.GetUser()
			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.getNavigation(user, "history"),
				PageTemplate: "admin_history",
				PageData:     resource.App.getHistory(resource, 0),
			})
		},
	)
}

func actionItemHistory(resource *Resource, permission Permission) *Action {
	return resource.AddItemAction("history").Name(messages.GetNameFunction("admin_history")).Permission(permission).Template("admin_history").DataSource(
		func(request Request) interface{} {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			must(resource.App.Query().WhereIs("id", int64(id)).Get(item))

			return resource.App.getHistory(resource, int64(id))
		},
	)
}

func actionDelete(resource *Resource, permission Permission) *Action {
	return resource.AddItemAction("delete").Permission(permission).Name(messages.GetNameFunction("admin_delete")).Template("admin_delete").DataSource(
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
			must(resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			itemName := getItemName(item)
			ret["delete_title"] = fmt.Sprintf("Chcete smazat položku %s?", itemName)
			ret["delete_title"] = messages.Get(user.Locale, "admin_delete_confirmation_name", itemName)
			return ret
		},
	)

}

func actionDoDelete(resource *Resource, permission Permission) *Action {
	return resource.AddItemAction("delete").Permission(permission).Method("POST").Handler(
		func(request Request) {
			user := request.GetUser()
			validateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			_, err = resource.App.Query().WhereIs("id", int64(id)).Delete(item)
			must(err)

			if resource.App.search != nil {
				err = resource.App.search.deleteItem(resource, int64(id))
				if err != nil {
					resource.App.Log().Println(fmt.Errorf("%s", err))
				}
				resource.App.search.flush()
			}

			if resource.ActivityLog {
				resource.App.createDeleteActivityLog(*resource, user, int64(id), item)
			}

			must(resource.updateCachedCount())
			request.AddFlashMessage(messages.Get(user.Locale, "admin_item_deleted"))
			request.Redirect(resource.getURL(""))
		},
	)
}

func actionPreview(resource *Resource, permission Permission) *Action {
	return resource.AddItemAction("preview").Name(messages.GetNameFunction("admin_preview")).Permission(permission).Handler(
		func(request Request) {
			var item interface{}
			resource.newItem(&item)
			must(resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			request.Redirect(
				resource.PreviewURLFunction(item),
			)
		},
	)
}

func (action *Action) getnavigation(request Request) adminItemNavigation {
	if action.resource != nil {
		user := request.GetUser()
		code := action.url
		if action.isItemAction {
			var item interface{}
			action.resource.newItem(&item)
			must(action.resource.App.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			return action.resource.getItemNavigation(user, item, code)
		} else {
			return action.resource.getNavigation(user, code)
		}
	}
	return adminItemNavigation{}

}

func (action *Action) bindAction() error {
	app := action.app
	if strings.HasPrefix(action.url, "/") {
		return errors.New("url can't start with / character")
	}

	var url string
	if action.resource == nil {
		url = app.GetAdminURL(action.url)
	} else {
		resource := action.resource
		if action.isItemAction {
			if action.url != "" {
				url = resource.getURL(":id/" + action.url)
			} else {
				url = resource.getURL(":id")
			}
		} else {
			url = resource.getURL(action.url)
		}
	}

	var controller *Controller
	if action.resource != nil {
		controller = action.resource.ResourceController
	} else {
		controller = app.AdminController
	}

	var fn = func(request Request) {
		user := request.GetUser()
		if !app.Authorize(user, action.permission) {
			render403(request)
			return
		}
		if action.handler != nil {
			action.handler(request)
		} else {
			var data interface{}
			if action.dataSource != nil {
				data = action.dataSource(request)
			}
			renderNavigationPage(request, adminNavigationPage{
				App:          app,
				Navigation:   action.getnavigation(request),
				PageTemplate: action.template,
				PageData:     data,
			})
		}
	}

	constraints := []func(map[string]string) bool{}
	if action.isItemAction {
		constraints = append(constraints, utils.ConstraintInt("id"))
	}

	switch action.method {
	case "POST":
		controller.Post(url, fn, constraints...)
	case "GET":
		controller.Get(url, fn, constraints...)
	case "PUT":
		controller.Put(url, fn, constraints...)
	case "DELETE":
		controller.Delete(url, fn, constraints...)
	default:
		return fmt.Errorf("unknown method %s", action.method)
	}
	return nil
}

func initResourceActions(a *App, resource *Resource) {
	if resource.CanCreate == "" {
		resource.CanCreate = resource.CanEdit
	}
	if resource.CanDelete == "" {
		resource.CanDelete = resource.CanEdit
	}

	resourceActions := []*Action{
		actionList(resource),
		actionNew(resource, resource.CanCreate),
		actionCreate(resource, resource.CanCreate),
	}
	if resource.ActivityLog {
		resourceActions = append(resourceActions, actionHistory(resource, resource.CanEdit))
	}

	itemActions := []*Action{
		actionView(resource),
	}

	itemActions = append(itemActions,
		actionEdit(resource, resource.CanEdit),
		actionUpdate(resource, resource.CanEdit),
		actionDelete(resource, resource.CanDelete),
		actionDoDelete(resource, resource.CanDelete),
	)

	if resource.PreviewURLFunction != nil {
		itemActions = append(itemActions, actionPreview(resource, resource.CanView))
	}

	if resource.ActivityLog {
		itemActions = append(itemActions, actionItemHistory(resource, resource.CanView))
	}

}

func (resource *Resource) getResourceActionsButtonData(user User, admin *App) (ret []buttonData) {
	navigation := resource.getNavigation(user, "")
	for _, v := range navigation.Tabs {
		ret = append(ret, buttonData{
			Name: v.Name,
			URL:  v.URL,
		})
	}
	return
}

func (app *App) getListItemActions(user User, item interface{}, id int64, resource Resource) listItemActions {
	ret := listItemActions{}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Name: messages.Get(user.Locale, "admin_view"),
		URL:  resource.getURL(fmt.Sprintf("%d", id)),
	})

	navigation := resource.getItemNavigation(user, item, "")

	for _, v := range navigation.Tabs {
		if !v.Selected {
			ret.MenuButtons = append(ret.MenuButtons, buttonData{
				Name: v.Name,
				URL:  v.URL,
			})
		}
	}

	if app.Authorize(user, resource.CanEdit) && resource.OrderColumnName != "" {
		ret.ShowOrderButton = true
	}

	return ret
}

/*
//AddItemAction adds item action
func (resource *Resource) AddItemAction(url string, name func(string) string, templateName string, dataGenerator func(Request) interface{}) {
	action := Action{
		Name:    name,
		URL:     url,
		Handler: createAdminHandler(url, templateName, dataGenerator, false),
	}
	resource.itemActions = append(resource.itemActions, action)
}

//AddItemAction adds item action
func (resource *Resource) AddItemPOSTAction(url string, handler func(Request)) {
	action := Action{
		Method: "post",
		URL:    url,
		Handler: func(resource Resource, request Request, user User) {
			handler(request)
		},
	}
	resource.itemActions = append(resource.itemActions, action)
}

//AddAction adds action
func (resource *Resource) AddAction(action Action) {
	resource.actions = append(resource.actions, action)
}*/
