package admin

import (
	"encoding/json"
	"fmt"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"strconv"
	"strings"
)

type ButtonData struct {
	Name   string
	Url    string
	Params map[string]string
}

type Action struct {
	Name    func(string) string
	Auth    Authenticatizer
	Method  string
	Url     string
	Handler func(Admin, Resource, prago.Request, User)
}

func (ra *Action) GetName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.Url
}

var actionList = Action{
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
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
	Url: "new",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		var item interface{}
		resource.newItem(&item)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

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
	Url:    "",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		var item interface{}
		resource.newItem(&item)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, true)
		}

		resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.getFilter())
		prago.Must(admin.Create(item))

		if resource.ActivityLog {
			admin.createNewActivityLog(resource, user, item)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, admin.GetItemURL(resource, item, ""))
	},
}

var actionView = Action{
	Url: "",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {

		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

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
		prago.Must(err)

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, id, ""),
			PageTemplate: "admin_view",
			PageData:     view,
		})
	},
}

var actionEdit = Action{
	Url: "edit",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

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
		prago.Must(err)

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
			Navigation:   admin.getItemNavigation(resource, user, item, id, "edit"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var actionUpdate = Action{
	Url:    "edit",
	Method: "post",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, false)
		}

		var beforeData []byte
		if resource.ActivityLog {
			beforeData, err = json.Marshal(item)
			prago.Must(err)
		}

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.getFilter())
		prago.Must(err)
		prago.Must(admin.Save(item))

		if resource.ActivityLog {
			afterData, err := json.Marshal(item)
			if err != nil {
				panic(err)
			}

			admin.createEditActivityLog(resource, user, int64(id), beforeData, afterData)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))

		if request.Params().Get("_submit_and_stay_clicked") == "true" {
			prago.Redirect(request, request.Request().URL.RequestURI())
		} else {
			prago.Redirect(request, admin.GetURL(&resource, fmt.Sprintf("%d", id)))
		}

	},
}

var actionHistory = Action{
	Url: "history",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(resource, user, "history"),
			PageTemplate: "admin_history",
			PageData:     admin.getHistory(&resource, 0, 0),
		})
	},
}

var actionItemHistory = Action{
	Url: "history",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, id, "history"),
			PageTemplate: "admin_history",
			PageData:     admin.getHistory(&resource, 0, int64(id)),
		})
	},
}

var actionExport = CreateNavigationalAction(
	"export",
	messages.Messages.GetNameFunction("admin_export"),
	"admin_export",
	func(admin Admin, resource Resource, request prago.Request, user User) interface{} {
		return resource.getExportFormData(user, resource.VisibilityFilter)
	},
)

var actionDoExport = Action{
	Url:     "export",
	Method:  "POST",
	Handler: exportHandler,
}

var actionDelete = CreateNavigationalItemAction(
	"delete",
	messages.Messages.GetNameFunction("admin_delete"),
	"admin_form",
	func(admin Admin, resource Resource, request prago.Request, user User) interface{} {
		form := NewForm()
		form.Method = "POST"
		AddCSRFToken(form, request)
		form.AddSubmit("send", messages.Messages.Get(user.Locale, "admin_delete"))
		return form
	},
)

var actionDoDelete = Action{
	Url:    "delete",
	Method: "post",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		_, err = admin.Query().WhereIs("id", int64(id)).Delete(item)
		prago.Must(err)

		if resource.ActivityLog {
			admin.createDeleteActivityLog(resource, user, int64(id), item)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		prago.Redirect(request, admin.GetURL(&resource, ""))
	},
}

var actionOrder = Action{
	Url:    "order",
	Method: "post",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		decoder := json.NewDecoder(request.Request().Body)
		var t = map[string][]int{}
		err := decoder.Decode(&t)
		prago.Must(err)

		order, ok := t["order"]
		if !ok {
			panic("wrong format")
		}

		for i, id := range order {
			var item interface{}
			resource.newItem(&item)
			prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))
			prago.Must(resource.StructCache.BindOrder(item, int64(i)))
			prago.Must(admin.Save(item))
		}

		prago.WriteAPI(request, true, 200)
	},
}

func bindResourceAction(a *Admin, resource *Resource, action Action) error {
	return bindAction(a, resource, action, false)
}

func bindResourceItemAction(a *Admin, resource *Resource, action Action) error {
	return bindAction(a, resource, action, true)
}

func bindAction(a *Admin, resource *Resource, action Action, isItemAction bool) error {
	var url string

	if resource == nil {
		url = a.Prefix + "/" + action.Url
	} else {
		if isItemAction {
			url = a.Prefix + "/" + resource.ID + "/:id"
			if len(action.Url) > 0 {
				url += "/" + action.Url
			}
		} else {
			url = a.GetURL(resource, action.Url)
		}
	}

	method := strings.ToLower(action.Method)
	var controller *prago.Controller
	if resource != nil {
		controller = resource.ResourceController
	} else {
		controller = a.AdminController
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
			action.Handler(*a, *resource, request, *user)
		} else {
			action.Handler(*a, Resource{}, request, *user)
		}
	}

	constraints := []prago.Constraint{}
	if isItemAction {
		constraints = []prago.Constraint{prago.ConstraintInt("id")}
	}

	switch method {
	case "post":
		controller.Post(url, fn, constraints...)
	default:
		controller.Get(url, fn, constraints...)
	}
	return nil
}

func initResourceActions(a *Admin, resource *Resource) error {
	for _, v := range resource.resourceActions {
		bindResourceAction(a, resource, v)
	}

	for _, v := range resource.resourceItemActions {
		bindResourceItemAction(a, resource, v)
	}

	if !resource.HasModel || !resource.HasView {
		return nil
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

	return nil
}

func (resource *Resource) ResourceActionsButtonData(user *User, admin *Admin) []ButtonData {
	ret := []ButtonData{}
	navigation := admin.getResourceNavigation(*resource, *user, "")
	for _, v := range navigation.Tabs {
		ret = append(ret, ButtonData{
			Name: v.Name,
			Url:  v.URL,
		})
	}
	return ret
}

func (admin *Admin) getListItemActions(user User, id int64, resource Resource) listItemActions {
	ret := listItemActions{}

	prefix := admin.GetURL(&resource, fmt.Sprintf("%d", id))

	ret.VisibleButtons = append(ret.VisibleButtons, ButtonData{
		Name: messages.Messages.Get(user.Locale, "admin_view"),
		Url:  prefix,
	})

	if resource.PreviewURLFunction != nil {
		var item interface{}
		resource.newItem(&item)
		err := admin.Query().WhereIs("id", id).Get(item)
		if err == nil {
			url := resource.PreviewURLFunction(item)
			if url != "" {
				ret.MenuButtons = append(ret.MenuButtons, ButtonData{
					Name: messages.Messages.Get(user.Locale, "admin_preview"),
					Url:  url,
					Params: map[string]string{
						"target": "_blank",
					},
				})
			}
		}
	}
	if resource.CanEdit {
		ret.MenuButtons = append(ret.MenuButtons, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_edit"),
			Url:  prefix + "/edit",
		})

		ret.MenuButtons = append(ret.MenuButtons, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_delete"),
			Url:  prefix + "/delete",
		})

		if resource.ActivityLog {
			ret.MenuButtons = append(ret.MenuButtons, ButtonData{
				Name: messages.Messages.Get(user.Locale, "admin_history"),
				Url:  prefix + "/edit",
			})
		}

		if resource.StructCache.OrderColumnName != "" {
			ret.ShowOrderButton = true
		}
	}

	if resource.CanExport {
		ret.MenuButtons = append(ret.MenuButtons, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_export"),
			Url:  prefix + "/export",
		})
	}

	for _, v := range resource.resourceItemActions {
		if v.Name == nil {
			continue
		}
		name := v.Url
		if v.Name != nil {
			name = v.Name(user.Locale)
		}

		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if v.Auth == nil || v.Auth(&user) {
				ret.MenuButtons = append(ret.MenuButtons, ButtonData{
					Name: name,
					Url:  prefix + "/" + v.Url,
				})
			}
		}
	}

	return ret
}

func (ar *Resource) AddItemAction(action Action) {
	ar.resourceItemActions = append(ar.resourceItemActions, action)
}

func (ar *Resource) AddAction(action Action) {
	ar.resourceActions = append(ar.resourceActions, action)
}
