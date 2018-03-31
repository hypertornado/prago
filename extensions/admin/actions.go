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

type ResourceAction struct {
	Name         func(string) string
	Auth         Authenticatizer
	Method       string
	Url          string
	Handler      func(Admin, Resource, prago.Request, User)
	ButtonParams map[string]string
}

func (ra *ResourceAction) GetName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.Url
}

var ActionList = ResourceAction{
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		navigation := admin.getResourceNavigation(resource, user, "")
		navigation.Wide = true
		request.SetData("navigation", navigation)

		listData, err := resource.getListHeader(admin, user)
		if err != nil {
			if err == ErrItemNotFound {
				render404(request)
				return
			}
			panic(err)
		}

		if resource.BeforeList != nil {
			if !resource.BeforeList(request, listData) {
				return
			}
		}

		request.SetData("admin_title", navigation.GetPageTitle())
		request.SetData("admin_list", listData)
		request.SetData("admin_yield", "admin_list")
		prago.Render(request, 200, "admin_layout")
	},
}

var ActionNew = ResourceAction{
	Url: "new",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		var item interface{}
		resource.newItem(&item)

		if resource.BeforeNew != nil {
			if !resource.BeforeNew(request, item) {
				return
			}
		}

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

var ActionCreate = ResourceAction{
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

		if resource.BeforeCreate != nil {
			if !resource.BeforeCreate(request, item) {
				return
			}
		}

		prago.Must(admin.Create(item))

		if resource.ActivityLog {
			admin.createNewActivityLog(resource, user, item)
		}

		if resource.AfterCreate != nil {
			if !resource.AfterCreate(request, item) {
				return
			}
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, admin.GetItemURL(resource, item, ""))
	},
}

var ActionView = ResourceAction{
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

var ActionEdit = ResourceAction{
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

		if resource.BeforeDetail != nil {
			if !resource.BeforeDetail(request, item) {
				return
			}
		}

		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(resource, user, item, id, "edit"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var ActionUpdate = ResourceAction{
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
			if err != nil {
				panic(err)
			}
		}

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.getFilter())
		prago.Must(err)

		if resource.BeforeUpdate != nil {
			if !resource.BeforeUpdate(request, item) {
				return
			}
		}
		prago.Must(admin.Save(item))

		if resource.ActivityLog {
			afterData, err := json.Marshal(item)
			if err != nil {
				panic(err)
			}

			admin.createEditActivityLog(resource, user, int64(id), beforeData, afterData)
		}

		if resource.AfterUpdate != nil {
			if !resource.AfterUpdate(request, item) {
				return
			}
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))

		if request.Params().Get("_submit_and_stay_clicked") == "true" {
			prago.Redirect(request, request.Request().URL.RequestURI())
		} else {
			prago.Redirect(request, admin.GetURL(&resource, fmt.Sprintf("%d", id)))
		}

	},
}

var ActionHistory = ResourceAction{
	Url: "history",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(resource, user, "history"),
			PageTemplate: "admin_history",
			PageData:     admin.getHistory(&resource, 0, 0),
		})
	},
}

var ActionItemHistory = ResourceAction{
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

var ActionExport = CreateNavigationalAction(
	"export",
	messages.Messages.GetNameFunction("admin_export"),
	"admin_export",
	func(admin Admin, resource Resource, request prago.Request, user User) interface{} {
		return nil
	},
)

var ActionDelete = CreateNavigationalItemAction(
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

var ActionDoDelete = ResourceAction{
	Url:    "delete",
	Method: "post",
	Handler: func(admin Admin, resource Resource, request prago.Request, user User) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		if resource.BeforeDelete != nil {
			if !resource.BeforeDelete(request, id) {
				return
			}
		}

		var item interface{}
		resource.newItem(&item)
		_, err = admin.Query().WhereIs("id", int64(id)).Delete(item)
		prago.Must(err)

		if resource.AfterDelete != nil {
			if !resource.AfterDelete(request, id) {
				return
			}
		}

		if resource.ActivityLog {
			admin.createDeleteActivityLog(resource, user, int64(id), item)
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		prago.Redirect(request, admin.GetURL(&resource, ""))
	},
}

var ActionOrder = ResourceAction{
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

func bindResourceAction(a *Admin, resource *Resource, action ResourceAction) error {
	return bindAction(a, resource, action, false)
}

func bindResourceItemAction(a *Admin, resource *Resource, action ResourceAction) error {
	return bindAction(a, resource, action, true)
}

func bindAction(a *Admin, resource *Resource, action ResourceAction, isItemAction bool) error {
	var url string
	if isItemAction {
		url = a.Prefix + "/" + resource.ID + "/:id"
		if len(action.Url) > 0 {
			url += "/" + action.Url
		}
	} else {
		url = a.GetURL(resource, action.Url)
	}

	method := strings.ToLower(action.Method)
	controller := resource.ResourceController

	var fn func(request prago.Request) = func(request prago.Request) {
		user := request.GetData("currentuser").(*User)
		if action.Auth != nil {
			if !action.Auth(user) {
				render403(request)
				return
			}
		}
		action.Handler(*a, *resource, request, *user)
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

//InitResourceDefault is default resource initializer
func InitResourceDefault(a *Admin, resource *Resource) error {
	for _, v := range resource.ResourceActions {
		bindResourceAction(a, resource, v)
	}

	for _, v := range resource.ResourceItemActions {
		bindResourceItemAction(a, resource, v)
	}

	if !resource.HasModel || !resource.HasView {
		return nil
	}

	bindResourceAction(a, resource, ActionList)
	bindResourceAction(a, resource, ActionOrder)

	if resource.ActivityLog {
		bindResourceAction(a, resource, ActionHistory)
	}

	if resource.CanCreate {
		bindResourceAction(a, resource, ActionNew)
		bindResourceAction(a, resource, ActionCreate)
	}

	bindResourceItemAction(a, resource, ActionView)

	if resource.CanEdit {
		bindResourceItemAction(a, resource, ActionEdit)
		bindResourceItemAction(a, resource, ActionUpdate)
		bindResourceItemAction(a, resource, ActionDelete)
		bindResourceItemAction(a, resource, ActionDoDelete)
	}

	if resource.CanExport {
		bindResourceAction(a, resource, ActionExport)
	}

	if resource.ActivityLog {
		bindResourceItemAction(a, resource, ActionItemHistory)
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

	for _, v := range resource.ResourceItemActions {
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
					Name:   name,
					Url:    prefix + "/" + v.Url,
					Params: v.ButtonParams,
				})
			}
		}
	}

	return ret
}

func (ar *Resource) AddResourceItemAction(action ResourceAction) {
	ar.ResourceItemActions = append(ar.ResourceItemActions, action)
}

func (ar *Resource) AddResourceAction(action ResourceAction) {
	ar.ResourceActions = append(ar.ResourceActions, action)
}
