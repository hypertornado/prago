package admin

import (
	"encoding/json"
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
	Name    func(string) string
	Method  string
	Url     string
	Handler func(*Admin, *Resource, prago.Request)
}

func (ra *ResourceAction) GetName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.Url
}

var ActionList = ResourceAction{
	Handler: func(admin *Admin, resource *Resource, request prago.Request) {
		listData, err := resource.getList(admin, GetLocale(request), request.Request().URL.Path, request.Request().URL.Query())
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

		form.Action = "../" + resource.ID
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
		AddCSRFToken(form, request)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, true)
		}

		request.SetData("admin_title", messages.Messages.Get(GetLocale(request), "admin_new")+" ⏤ "+resource.Name(GetLocale(request)))

		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
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
		prago.Redirect(request, admin.Prefix+"/"+resource.ID)
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

		form.Action = "edit"
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
		AddCSRFToken(form, request)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, false)
		}

		if resource.BeforeDetail != nil {
			if !resource.BeforeDetail(request, item) {
				return
			}
		}

		request.SetData("admin_title", messages.Messages.Get(GetLocale(request), "admin_edit")+" ⏤ "+resource.Name(GetLocale(request)))
		request.SetData("admin_item", item)
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
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
		prago.Redirect(request, admin.Prefix+"/"+resource.ID)
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

func BindResourceAction(a *Admin, resource *Resource, action ResourceAction) error {
	return BindAction(a, resource, action, false)
}

func BindResourceItemAction(a *Admin, resource *Resource, action ResourceAction) error {
	return BindAction(a, resource, action, true)
}

func BindAction(a *Admin, resource *Resource, action ResourceAction, isItemAction bool) error {
	var url string
	if isItemAction {
		url = a.GetItemURL(resource, action.Url)
	} else {
		url = a.GetURL(resource, action.Url)
	}

	method := strings.ToLower(action.Method)
	controller := resource.ResourceController

	var fn func(request prago.Request) = func(request prago.Request) {
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
		BindResourceAction(a, resource, v)
	}

	if !resource.HasModel || !resource.HasView {
		return nil
	}

	BindResourceAction(a, resource, ActionList)
	BindResourceAction(a, resource, ActionOrder)

	if resource.CanCreate {
		BindResourceAction(a, resource, ActionNew)
		BindResourceAction(a, resource, ActionCreate)
	}

	if resource.CanEdit {
		BindResourceItemAction(a, resource, ActionEdit)
		BindResourceItemAction(a, resource, ActionUpdate)
		BindResourceItemAction(a, resource, ActionDelete)
	}

	return nil
}
