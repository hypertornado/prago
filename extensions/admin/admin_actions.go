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

		tableData, err := resource.ListTableItems(request)
		prago.Must(err)

		request.SetData("admin_list_table_data", tableData)
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
		prago.Must(resource.Create(item))

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

		err = resource.Save(item)
		prago.Must(err)

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDelete(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id/delete"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		prago.Must(err)

		_, err = resource.Query().Where(map[string]interface{}{"id": int64(id)}).Delete()
		prago.Must(err)

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
	for _, v := range defaultActions {
		action := resource.Actions[v]
		if action != nil {
			action(a, resource)
			delete(resource.Actions, v)
		}
	}

	for _, v := range resource.Actions {
		if v != nil {
			v(a, resource)
		}
	}
	return nil
}
