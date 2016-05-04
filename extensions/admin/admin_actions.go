package admin

import (
	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/extensions/admin/messages"
	"strconv"
)

func BindList(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ""), func(request prago.Request) {

		tableData, err := resource.ListTableItems(GetLocale(request))
		if err != nil {
			panic(err)
		}

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
		if err != nil {
			panic(err)
		}

		form.Action = "../" + resource.ID
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_create"))
		AddCSRFToken(form, request)

		request.SetData("admin_form", form)
		request.SetData("admin_yield", "admin_new")
		prago.Render(request, 200, "admin_layout")
	})
}

func BindCreate(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ""), func(request prago.Request) {
		ValidateCSRF(request)
		item, err := resource.NewItem()
		if err != nil {
			panic(err)
		}
		resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, resource.BindDataFilter)
		err = resource.Create(item)
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_created"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDetail(a *Admin, resource *AdminResource) {
	resource.ResourceController.Get(a.GetURL(resource, ":id"), func(request prago.Request) {
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		if err != nil {
			panic(err)
		}

		form, err := resource.StructCache.GetForm(item, GetLocale(request), resource.VisibilityFilter, resource.EditabilityFilter)
		if err != nil {
			panic(err)
		}

		form.Action = request.Params().Get("id")
		form.AddSubmit("_submit", messages.Messages.Get(GetLocale(request), "admin_edit"))
		AddCSRFToken(form, request)

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
		if err != nil {
			panic(err)
		}

		item, err := resource.Query().Where(map[string]interface{}{"id": int64(id)}).First()
		if err != nil {
			panic(err)
		}

		err = resource.StructCache.BindData(item, request.Params(), request.Request().MultipartForm, resource.BindDataFilter)
		if err != nil {
			panic(err)
		}

		err = resource.Save(item)
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_edited"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func BindDelete(a *Admin, resource *AdminResource) {
	resource.ResourceController.Post(a.GetURL(resource, ":id/delete"), func(request prago.Request) {
		ValidateCSRF(request)
		id, err := strconv.Atoi(request.Params().Get("id"))
		if err != nil {
			panic(err)
		}

		_, err = resource.Query().Where(map[string]interface{}{"id": int64(id)}).Delete()
		if err != nil {
			panic(err)
		}

		FlashMessage(request, messages.Messages.Get(GetLocale(request), "admin_item_deleted"))
		prago.Redirect(request, a.Prefix+"/"+resource.ID)
	})
}

func AdminInitResourceDefault(a *Admin, resource *AdminResource) error {
	BindList(a, resource)
	BindNew(a, resource)
	BindCreate(a, resource)
	BindDetail(a, resource)
	BindUpdate(a, resource)
	BindDelete(a, resource)
	return nil
}
