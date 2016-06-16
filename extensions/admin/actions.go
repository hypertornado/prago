package admin

import (
	"encoding/json"
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"strconv"
)

type ActionBinder func(a *Admin, resource *AdminResource)

func BindList(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ""), func(request prago.Request) {

		listData, err := resource.GetList(GetLocale(request), request.Request().URL.Path, request.Request().URL.Query())
		prago.Must(err)

		if resource.BeforeList != nil {
			if !resource.BeforeList(request, listData) {
				return
			}
		}

		request.SetData("admin_list", listData)
		request.SetData("admin_yield", "admin_list")
		prago.Render(request, 200, "admin_layout")

	})
}

func BindNew(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, "new"), func(request prago.Request) {

		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}

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

		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindCreate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		ValidateCSRF(request)
		item, err := resource.NewItem()
		prago.Must(err)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, true)
		}

		resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.GetFilter())

		if resource.BeforeCreate != nil {
			if !resource.BeforeCreate(request, item) {
				return
			}
		}

		prago.Must(resource.Create(item))

		if resource.AfterCreate != nil {
			if !resource.AfterCreate(request, item) {
				return
			}
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDetail(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		prago.Must(err)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		form.Action = request.Params().Get("id")
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

		request.SetData("admin_item", item)
		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_edit")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindUpdate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		prago.Must(err)

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		prago.Must(err)

		if resource.AfterFormCreated != nil {
			form = resource.AfterFormCreated(form, request, false)
		}

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, form.GetFilter())
		prago.Must(err)

		if resource.BeforeUpdate != nil {
			if !resource.BeforeUpdate(request, item) {
				return
			}
		}

		err = resource.Save(item)
		prago.Must(err)

		if resource.AfterUpdate != nil {
			if !resource.AfterUpdate(request, item) {
				return
			}
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDelete(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id/delete"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		if resource.BeforeDelete != nil {
			if !resource.BeforeDelete(request, id) {
				return
			}
		}

		_, err = resource.Query().Where(map[string]interface{}{"id": int64(id)}).Delete()
		prago.Must(err)

		if resource.AfterDelete != nil {
			if !resource.AfterDelete(request, id) {
				return
			}
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindOrder(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, "order"), func(request prago.Request) {
		decoder := json.NewDecoder(request.Request().Body)
		var t = map[string][]int{}
		err := decoder.Decode(&t)
		prago.Must(err)

		order, ok := t["order"]
		if !ok {
			panic("wrong format")
		}

		for i, id := range order {
			item, err := resource.Query().Where(id).First()
			prago.Must(err)
			prago.Must(resource.StructCache.BindOrder(item, int64(i)))
			prago.Must(resource.Save(item))
		}

		WriteApi(request, true, 200)
	})
}

func AdminInitResourceDefault(a *Admin, resource *AdminResource) error {
	defaultActions := []string{"list", "order", "new", "create", "detail", "update", "delete"}
	usedActions := make(map[string]bool)
	for _, v := range defaultActions {
		action := resource.Actions[v]
		if action != nil {
			action(a, resource)
			usedActions[v] = true
		}
	}

	for k, v := range resource.Actions {
		if v != nil {
			if usedActions[k] == false {
				v(a, resource)
			}
		}
	}
	return nil
}
