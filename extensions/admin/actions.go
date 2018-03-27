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
	Handler      func(*Admin, *Resource, prago.Request)
	ButtonParams map[string]string
}

func (ra *ResourceAction) GetName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.Url
}

var ActionList = ResourceAction{
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
		user := request.GetData("currentuser").(*User)
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

		request.SetData("admin_title", resource.Name(GetLocale(request)))
		request.SetData("admin_list", listData)
		request.SetData("admin_yield", "admin_list")
		prago.Render(request, 200, "admin_layout")
	},
}

var ActionNew = ResourceAction{
	Url: "new",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
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

		user := request.GetData("currentuser").(*User)
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getResourceNavigation(*resource, *user, "new"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var ActionCreate = ResourceAction{
	Method: "post",
	Url:    "",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
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

		if resource.AfterCreate != nil {
			if !resource.AfterCreate(request, item) {
				return
			}
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))

		prago.Redirect(request, admin.GetItemURL(*resource, item, ""))
		//prago.Redirect(request, admin.Prefix+"/"+resource.ID)
	},
}

var ActionView = ResourceAction{
	Url: "",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {

		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

		view, err := resource.StructCache.getView(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		user := request.GetData("currentuser").(*User)
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(*resource, *user, item, id, ""),
			PageTemplate: "admin_view",
			PageData:     view,
		})
	},
}

var ActionEdit = ResourceAction{
	Url: "edit",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		var item interface{}
		resource.newItem(&item)
		prago.Must(admin.Query().WhereIs("id", int64(id)).Get(item))

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

		user := request.GetData("currentuser").(*User)
		renderNavigationPage(request, AdminNavigationPage{
			Navigation:   admin.getItemNavigation(*resource, *user, item, id, "edit"),
			PageTemplate: "admin_form",
			PageData:     form,
		})
	},
}

var ActionUpdate = ResourceAction{
	Url:    "edit",
	Method: "post",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
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

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.getFilter())
		prago.Must(err)

		if resource.BeforeUpdate != nil {
			if !resource.BeforeUpdate(request, item) {
				return
			}
		}
		prago.Must(admin.Save(item))

		if resource.AfterUpdate != nil {
			if !resource.AfterUpdate(request, item) {
				return
			}
		}

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))

		if request.Params().Get("_submit_and_stay_clicked") == "true" {
			prago.Redirect(request, request.Request().URL.RequestURI())
		} else {
			prago.Redirect(request, admin.GetURL(resource, fmt.Sprintf("%d", id)))
		}

	},
}

var ActionDelete = ResourceAction{
	Url:    "delete",
	Method: "post",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
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

		AddFlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		prago.WriteAPI(request, true, 200)
	},
}

var ActionOrder = ResourceAction{
	Url:    "order",
	Method: "post",
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
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
		if action.Auth != nil {
			user := request.GetData("currentuser").(*User)
			if !action.Auth(user) {
				render403(request)
				return
			}
		}
		action.Handler(a, resource, request)
	}

	switch method {
	case "post":
		controller.Post(url, fn)
	default:
		controller.Get(url, fn)
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

	if resource.CanCreate {
		bindResourceAction(a, resource, ActionNew)
		bindResourceAction(a, resource, ActionCreate)
	}

	bindResourceItemAction(a, resource, ActionView)

	if resource.CanEdit {
		bindResourceItemAction(a, resource, ActionEdit)
		bindResourceItemAction(a, resource, ActionUpdate)
		bindResourceItemAction(a, resource, ActionDelete)
	}

	return nil
}

func (ar *Resource) ResourceActionsButtonData(user *User, admin *Admin) []ButtonData {
	ret := []ButtonData{}
	if ar.CanCreate {
		ret = append(ret, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_new"),
			Url:  admin.GetURL(ar, "new"),
		})
	}

	for _, v := range ar.ResourceActions {
		if v.Url == "" {
			continue
		}
		name := v.Url
		if v.Name != nil {
			name = v.Name(user.Locale)
		}

		if v.Auth == nil || v.Auth(user) {
			ret = append(ret, ButtonData{
				Name: name,
				Url:  admin.GetURL(ar, v.Url),
			})
		}
	}
	return ret
}

func (ar *Resource) ResourceItemActionsButtonData(user *User, id int64, admin *Admin) []ButtonData {
	prefix := admin.GetURL(ar, fmt.Sprintf("%d", id))

	ret := []ButtonData{}

	ret = append(ret, ButtonData{
		Name: messages.Messages.Get(user.Locale, "admin_view"),
		Url:  prefix,
	})

	if ar.PreviewURLFunction != nil {
		var item interface{}
		ar.newItem(&item)
		err := admin.Query().WhereIs("id", id).Get(item)
		if err == nil {
			url := ar.PreviewURLFunction(item)
			if url != "" {
				ret = append(ret, ButtonData{
					Name: messages.Messages.Get(user.Locale, "admin_preview"),
					Url:  url,
					Params: map[string]string{
						"target": "_blank",
					},
				})
			}
		}
	}
	if ar.CanEdit {
		ret = append(ret, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_edit"),
			Url:  prefix + "/edit",
		})

		ret = append(ret, ButtonData{
			Name: messages.Messages.Get(user.Locale, "admin_delete"),
			Url:  "",
			Params: map[string]string{
				"class":                "btn admin-action-delete",
				"data-action":          fmt.Sprintf("%s/%d/delete?_csrfToken=", ar.ID, id),
				"data-confirm-message": messages.Messages.Get(user.Locale, "admin_delete_confirmation"),
			},
		})

		if ar.StructCache.OrderColumnName != "" {
			ret = append(ret, ButtonData{
				Name: "☰",
				Url:  "",
				Params: map[string]string{
					"class": "btn admin-action-order",
				},
			})
		}
	}

	for _, v := range ar.ResourceItemActions {
		if v.Name == nil {
			continue
		}
		name := v.Url
		if v.Name != nil {
			name = v.Name(user.Locale)
		}

		if v.Method == "" || v.Method == "get" || v.Method == "GET" {
			if v.Auth == nil || v.Auth(user) {
				ret = append(ret, ButtonData{
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
