package administration

import (
	"encoding/json"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
	"strconv"
	"strings"
)

type buttonData struct {
	Name   string
	URL    string
	Params map[string]string
}

type Action struct {
	Name    func(string) string
	Auth    Authenticatizer
	Method  string
	URL     string
	Handler func(Administration, Resource, prago.Request, User)
}

func (ra *Action) GetName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.URL
}

var actionList = Action{
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		listData, err := resource.getListHeader(admin, user)
		if err != nil {
			if err == ErrItemNotFound {
				render404(request)
				return
			}
			panic(err)
		}

		navigation := admin.getResourceNavigation(resource, user, "")
		navigation.Wide = true

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   navigation,
			PageTemplate: "admin_list",
			PageData:     listData,
		})
	},
}

var actionNew = Action{
	URL: "new",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		var item interface{}
		resource.newItem(&item)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		must(err)

		form.Classes = append(form.Classes, "form_leavealert")
		form.Action = "../" + resource.ID
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
		AddCSRFToken(form, request)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, true)
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(resource, user, "new"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var actionCreate = Action{
	Method: "post",
	URL:    "",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		var item interface{}
		resource.newItem(&item)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, true)
		}

		resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.getFilter())
		must(admin.Create(item))

		if resource.ActivityLog {
			admin.createNewActivityLog(resource, user, item)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		request.Redirect(resource.GetItemURL(item, ""))
	},
}

var actionView = Action{
	URL: "",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {

		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		err = admin.Query().WhereIs("id", int64(id)).Get(item)
		if err != nil {
			if err == ErrItemNotFound {
				render404(request)
				return
			}
			panic(err)
		}

		view, err := resource.StructCache.getView(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		must(err)

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, ""),
			PageTemplate: "admin_view",
			PageData:     view,
		})
	},
}

var actionEdit = Action{
	URL: "edit",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		err = admin.Query().WhereIs("id", int64(id)).Get(item)
		if err != nil {
			if err == ErrItemNotFound {
				render404(request)
				return
			}
			panic(err)
		}

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		must(err)

		form.Classes = append(form.Classes, "form_leavealert")
		form.Action = "edit"
		form.AddHidden("_submit_and_stay_clicked")
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
		form.AddSubmit("_submit_and_stay", messages.Messages.Get(GetLocale(request), "admin_edit_and_stay"))
		AddCSRFToken(form, request)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, false)
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, "edit"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var actionUpdate = Action{
	URL:    "edit",
	Method: "post",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		must(admin.Query().WhereIs("id", int64(id)).Get(item))

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, false)
		}

		var beforeData []byte
		if resource.ActivityLog {
			beforeData, err = json.Marshal(item)
			must(err)
		}

		must(
			resource.StructCache.BindData(
				item, request.Params(), request.Request().MultipartForm, form.getFilter(),
			),
		)
		must(admin.Save(item))

		if resource.ActivityLog {
			afterData, err := json.Marshal(item)
			if err != nil {
				panic(err)
			}

			admin.createEditActivityLog(resource, user, int64(id), beforeData, afterData)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))

		if request.Params().Get("_submit_and_stay_clicked") == "true" {
			request.Redirect(request.Request().URL.RequestURI())
		} else {
			request.Redirect(resource.GetURL(fmt.Sprintf("%d", id)))
		}

	},
}

var actionHistory = Action{
	URL: "history",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(resource, user, "history"),
			PageTemplate: "admin_history",
			PageData:     admin.getHistory(&resource, 0, 0),
		})
	},
}

var actionItemHistory = Action{
	URL: "history",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		must(admin.Query().WhereIs("id", int64(id)).Get(item))

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, "history"),
			PageTemplate: "admin_history",
			PageData:     admin.getHistory(&resource, 0, int64(id)),
		})
	},
}

var actionExport = CreateNavigationalAction(
	"export",
	messages.Messages.GetNameFunction("admin_export"),
	"admin_export",
	func(admin Administration, resource Resource, request prago.Request, user User) interface{} {
		return resource.getExportFormData(user, resource.VisibilityFilter)
	},
)

var actionDoExport = Action{
	URL:     "export",
	Method:  "POST",
	Handler: exportHandler,
}

var actionDelete = CreateNavigationalItemAction(
	"delete",
	messages.Messages.GetNameFunction("admin_delete"),
	"admin_form",
	func(admin Administration, resource Resource, request prago.Request, user User) interface{} {
		form := NewForm()
		form.Method = "POST"
		AddCSRFToken(form, request)
		form.AddSubmit("send", messages.Messages.Get(user.Locale, "admin_delete"))
		return form
	},
)

var actionDoDelete = Action{
	URL:    "delete",
	Method: "post",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		must(err)

		var item interface{}
		resource.newItem(&item)
		_, err = admin.Query().WhereIs("id", int64(id)).Delete(item)
		must(err)

		if resource.ActivityLog {
			admin.createDeleteActivityLog(resource, user, int64(id), item)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		request.Redirect(resource.GetURL(""))
	},
}

var actionOrder = Action{
	URL:    "order",
	Method: "post",
	Handler: func(admin Administration, resource Resource, request prago.Request, user User) {
		decoder := json.NewDecoder(request.Request().Body)
		var t = map[string][]int{}
		must(decoder.Decode(&t))

		order, ok := t["order"]
		if !ok {
			panic("wrong format")
		}

		for i, id := range order {
			var item interface{}
			resource.newItem(&item)
			must(admin.Query().WhereIs("id", int64(id)).Get(item))
			must(resource.StructCache.BindOrder(item, int64(i)))
			must(admin.Save(item))
		}
		request.RenderJSON(true)
	},
}

func bindResourceAction(admin *Administration, resource *Resource, action Action) error {
	return bindAction(admin, resource, action, false)
}

func bindResourceItemAction(admin *Administration, resource *Resource, action Action) error {
	return bindAction(admin, resource, action, true)
}

func bindAction(admin *Administration, resource *Resource, action Action, isItemAction bool) error {
	if strings.HasPrefix(action.URL, "/") {
		return nil
	}

	var url string
	if resource == nil {
		url = admin.GetURL(action.URL)
	} else {
		if isItemAction {
			if action.URL != "" {
				url = resource.GetURL(":id/" + action.URL)
			} else {
				url = resource.GetURL(":id")
			}
		} else {
			url = resource.GetURL(action.URL)
		}
	}

	method := strings.ToLower(action.Method)
	var controller *prago.Controller
	if resource != nil {
		controller = resource.ResourceController
	} else {
		controller = admin.AdminController
	}

	var fn func(request prago.Request) = func(request prago.Request) {
		user := request.GetData("currentuser").(*User)
		if action.Auth != nil {
			if !action.Auth(user) {
				render403(request)
				return
			}
		}
		if resource != nil {
			action.Handler(*admin, *resource, request, *user)
		} else {
			action.Handler(*admin, Resource{}, request, *user)
		}
	}

	constraints := []func(map[string]string) bool{}
	if isItemAction {
		constraints = append(constraints, utils.ConstraintInt("id"))
	}

	switch method {
	case "post":
		controller.Post(url, fn, constraints...)
	default:
		controller.Get(url, fn, constraints...)
	}
	return nil
}

func initResourceActions(a *Administration, resource *Resource) {
	for _, v := range resource.relations {
		resource.bindRelationActions(v)
	}

	for _, v := range resource.resourceActions {
		bindResourceAction(a, resource, v)
	}

	for _, v := range resource.resourceItemActions {
		bindResourceItemAction(a, resource, v)
	}

	if !resource.HasModel || !resource.HasView {
		return
	}

	bindResourceAction(a, resource, actionList)
	bindResourceAction(a, resource, actionOrder)

	if resource.ActivityLog {
		bindResourceAction(a, resource, actionHistory)
	}

	if resource.CanCreate {
		bindResourceAction(a, resource, actionNew)
		bindResourceAction(a, resource, actionCreate)
	}

	bindResourceItemAction(a, resource, actionView)

	if resource.CanEdit {
		bindResourceItemAction(a, resource, actionEdit)
		bindResourceItemAction(a, resource, actionUpdate)
		bindResourceItemAction(a, resource, actionDelete)
		bindResourceItemAction(a, resource, actionDoDelete)
	}

	if resource.CanExport {
		bindResourceAction(a, resource, actionExport)
		bindResourceAction(a, resource, actionDoExport)
	}

	if resource.ActivityLog {
		bindResourceItemAction(a, resource, actionItemHistory)
	}

}

func (resource *Resource) getResourceActionsButtonData(user *User, admin *Administration) (ret []buttonData) {
	navigation := admin.getResourceNavigation(*resource, *user, "")
	for _, v := range navigation.Tabs {
		ret = append(ret, buttonData{
			Name: v.Name,
			URL:  v.URL,
		})
	}
	return
}

func (admin *Administration) getListItemActions(user User, item interface{}, id int64, resource Resource) listItemActions {
	ret := listItemActions{}

	ret.VisibleButtons = append(ret.VisibleButtons, buttonData{
		Name: messages.Messages.Get(user.Locale, "admin_view"),
		URL:  resource.GetURL(fmt.Sprintf("%d", id)),
	})

	navigation := admin.getItemNavigation(resource, user, item, "")

	for _, v := range navigation.Tabs {
		if !v.Selected {
			ret.MenuButtons = append(ret.MenuButtons, buttonData{
				Name: v.Name,
				URL:  v.URL,
			})
		}
	}

	if resource.CanEdit && resource.StructCache.OrderColumnName != "" {
		ret.ShowOrderButton = true
	}

	return ret
}

func (resource *Resource) AddItemAction(action Action) {
	resource.resourceItemActions = append(resource.resourceItemActions, action)
}

func (resource *Resource) AddAction(action Action) {
	resource.resourceActions = append(resource.resourceActions, action)
}
