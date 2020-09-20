package administration

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/tealeg/xlsx"

	"github.com/hypertornado/prago"
	"github.com/hypertornado/prago/administration/messages"
	"github.com/hypertornado/prago/utils"
)

type buttonData struct {
	Name   string
	URL    string
	Params map[string]string
}

//Action represents action
type Action struct {
	Name       func(string) string
	Permission Permission
	Method     string
	URL        string
	Handler    func(Resource, prago.Request, User)
}

func (ra *Action) getName(language string) string {
	if ra.Name != nil {
		return ra.Name(language)
	}
	return ra.URL
}

func actionList(resource *Resource) Action {
	return Action{
		Permission: resource.CanView,
		Name:       resource.HumanName,
		Handler: func(resource Resource, request prago.Request, user User) {
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(resource.Admin, user, request.Request().URL.Query())
				if err != nil {
					panic(err)
				}
				request.RenderJSON(listDataJSON)

				//request.Response().Header().Set("X-Count-Str", listData.TotalCountStr)

				/*buf := new(bytes.Buffer)
				err = resource.Admin.App.ExecuteTemplate(buf, "admin_list_cells", map[string]interface{}{
					"admin_list": listData,
				})
				if err != nil {
					panic(err)
				}

				ret := map[string]interface{}{
					"Content": string(buf.Bytes()),
				}

				request.RenderJSON(ret)*/

				return
			}

			if request.Request().URL.Query().Get("_format") == "xlsx" {
				listData, err := resource.getListContent(resource.Admin, user, request.Request().URL.Query())
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
						cell.SetValue(v2.OriginalValue)
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

			navigation := resource.Admin.getResourceNavigation(resource, user, "")
			navigation.Wide = true

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   navigation,
				PageTemplate: "admin_list",
				PageData:     listData,
			})
		},
	}
}

func actionNew(permission Permission) Action {
	return Action{
		Permission: permission,
		Name:       messages.Messages.GetNameFunction("admin_new"),
		URL:        "new",
		Handler: func(resource Resource, request prago.Request, user User) {
			var item interface{}
			resource.newItem(&item)

			resource.bindData(&item, user, request.Request().URL.Query(), defaultEditabilityFilter)

			form, err := resource.getForm(item, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "../" + resource.ID
			form.AddSubmit("_submit", messages.Messages.Get(user.Locale, "admin_save"))
			AddCSRFToken(form, request)

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getResourceNavigation(resource, user, "new"),
				PageTemplate: "admin_form",
				PageData:     form,
			})
		},
	}
}

func actionCreate(permission Permission) Action {
	return Action{
		Method:     "post",
		Permission: permission,
		URL:        "",
		Handler: func(resource Resource, request prago.Request, user User) {
			ValidateCSRF(request)
			var item interface{}
			resource.newItem(&item)

			form, err := resource.getForm(item, user)
			must(err)

			resource.bindData(item, user, request.Params(), form.getFilter())
			if resource.OrderFieldName != "" {
				resource.setOrderPosition(&item, resource.count()+1)
			}
			must(resource.Admin.Create(item))

			if resource.Admin.search != nil {
				err = resource.Admin.search.saveItem(&resource, item)
				if err != nil {
					request.Log().Println(fmt.Errorf("%s", err))
				}
				resource.Admin.search.Flush()
			}

			if resource.ActivityLog {
				resource.Admin.createNewActivityLog(resource, user, item)
			}

			AddFlashMessage(request, messages.Messages.Get(user.Locale, "admin_item_created"))
			request.Redirect(resource.GetItemURL(item, ""))
		},
	}
}

func actionView(resource *Resource) Action {
	return Action{
		Permission: resource.CanView,
		//Name:       messages.Messages.GetNameFunction("admin_view"),
		URL: "",
		Handler: func(resource Resource, request prago.Request, user User) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = resource.Admin.Query().WhereIs("id", int64(id)).Get(item)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getItemNavigation(resource, user, item, ""),
				PageTemplate: "admin_views",
				PageData:     resource.getViews(id, item, GetUser(request)),
				HideBox:      true,
			})
		},
	}
}

func actionEdit(permission Permission) Action {
	return Action{
		Name:       messages.Messages.GetNameFunction("admin_edit"),
		Permission: permission,
		URL:        "edit",
		Handler: func(resource Resource, request prago.Request, user User) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			err = resource.Admin.Query().WhereIs("id", int64(id)).Get(item)
			if err != nil {
				if err == ErrItemNotFound {
					render404(request)
					return
				}
				panic(err)
			}

			form, err := resource.getForm(item, user)
			must(err)

			form.Classes = append(form.Classes, "form_leavealert")
			form.Action = "edit"
			form.AddSubmit("_submit", messages.Messages.Get(user.Locale, "admin_save"))
			AddCSRFToken(form, request)

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getItemNavigation(resource, user, item, "edit"),
				PageTemplate: "admin_form",
				PageData:     form,
			})
		},
	}
}

func actionUpdate(permission Permission) Action {
	return Action{
		Permission: permission,
		URL:        "edit",
		Method:     "post",
		Handler: func(resource Resource, request prago.Request, user User) {
			ValidateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))

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
			must(resource.Admin.Save(item))

			if resource.Admin.search != nil {
				err = resource.Admin.search.saveItem(&resource, item)
				if err != nil {
					request.Log().Println(fmt.Errorf("%s", err))
				}
				resource.Admin.search.Flush()
			}

			if resource.ActivityLog {
				afterData, err := json.Marshal(item)
				if err != nil {
					panic(err)
				}

				resource.Admin.createEditActivityLog(resource, user, int64(id), beforeData, afterData)
			}

			AddFlashMessage(request, messages.Messages.Get(user.Locale, "admin_item_edited"))
			request.Redirect(resource.GetURL(fmt.Sprintf("%d", id)))
		},
	}
}

func actionHistory(permission Permission) Action {
	return Action{
		Name:       messages.Messages.GetNameFunction("admin_history"),
		Permission: permission,
		URL:        "history",
		Handler: func(resource Resource, request prago.Request, user User) {
			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getResourceNavigation(resource, user, "history"),
				PageTemplate: "admin_history",
				PageData:     resource.Admin.getHistory(&resource, 0),
			})
		},
	}
}

func actionItemHistory(permission Permission) Action {
	return Action{
		Name:       messages.Messages.GetNameFunction("admin_history"),
		Permission: permission,
		URL:        "history",
		Handler: func(resource Resource, request prago.Request, user User) {
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", int64(id)).Get(item))

			renderNavigationPage(request, adminNavigationPage{
				Navigation:   resource.Admin.getItemNavigation(resource, user, item, "history"),
				PageTemplate: "admin_history",
				PageData:     resource.Admin.getHistory(&resource, int64(id)),
			})
		},
	}
}

func actionDelete(permission Permission) Action {
	ret := CreateNavigationalItemAction(
		"delete",
		messages.Messages.GetNameFunction("admin_delete"),
		"admin_delete",
		func(resource Resource, request prago.Request, user User) interface{} {
			ret := map[string]interface{}{}
			form := NewForm()
			form.Method = "POST"
			AddCSRFToken(form, request)
			form.AddDeleteSubmit("send", messages.Messages.Get(user.Locale, "admin_delete"))
			ret["form"] = form

			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			itemName := getItemName(item)
			ret["delete_title"] = fmt.Sprintf("Chcete smazat položku %s?", itemName)
			ret["delete_title"] = messages.Messages.Get(user.Locale, "admin_delete_confirmation_name", itemName)
			fmt.Println(itemName)

			return ret
		},
	)
	ret.Permission = permission
	return ret
}

func actionDoDelete(permission Permission) Action {
	return Action{
		Permission: permission,
		URL:        "delete",
		Method:     "post",
		Handler: func(resource Resource, request prago.Request, user User) {
			ValidateCSRF(request)
			id, err := strconv.Atoi(request.Params().Get("id"))
			must(err)

			var item interface{}
			resource.newItem(&item)
			_, err = resource.Admin.Query().WhereIs("id", int64(id)).Delete(item)
			must(err)

			if resource.Admin.search != nil {
				err = resource.Admin.search.deleteItem(&resource, int64(id))
				if err != nil {
					request.Log().Println(fmt.Errorf("%s", err))
				}
				resource.Admin.search.Flush()
			}

			if resource.ActivityLog {
				resource.Admin.createDeleteActivityLog(resource, user, int64(id), item)
			}

			AddFlashMessage(request, messages.Messages.Get(user.Locale, "admin_item_deleted"))
			request.Redirect(resource.GetURL(""))
		},
	}
}

func actionPreview(permission Permission) Action {
	return Action{
		Name:       messages.Messages.GetNameFunction("admin_preview"),
		Permission: permission,
		URL:        "preview",
		Handler: func(resource Resource, request prago.Request, user User) {
			var item interface{}
			resource.newItem(&item)
			must(resource.Admin.Query().WhereIs("id", request.Params().Get("id")).Get(item))
			request.Redirect(
				resource.PreviewURLFunction(item),
			)
		},
	}
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

	var fn = func(request prago.Request) {
		user := GetUser(request)
		if !admin.Authorize(user, action.Permission) {
			render403(request)
			return
		}
		if resource != nil {
			action.Handler(*resource, request, user)
		} else {
			//TODO: ugly hack
			action.Handler(Resource{Admin: admin}, request, user)
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
	if resource.CanCreate == "" {
		resource.CanCreate = resource.CanEdit
	}
	if resource.CanDelete == "" {
		resource.CanDelete = resource.CanEdit
	}

	resourceActions := []Action{
		actionList(resource),
		actionNew(resource.CanCreate),
		actionCreate(resource.CanCreate),
		//actionStats(resource.CanView),
		//actionExport(resource.CanExport),
		//actionDoExport(resource.CanExport),
	}
	if resource.ActivityLog {
		resourceActions = append(resourceActions, actionHistory(resource.CanEdit))
	}
	/*for _, v := range resource.relations {
		resource.bindRelationActions(v)
	}*/
	resource.actions = append(resourceActions, resource.actions...)
	for _, v := range resource.actions {
		bindResourceAction(a, resource, v)
	}

	itemActions := []Action{
		actionView(resource),
	}

	itemActions = append(itemActions,
		actionEdit(resource.CanEdit),
		actionUpdate(resource.CanEdit),
		actionDelete(resource.CanDelete),
		actionDoDelete(resource.CanDelete),
	)

	if resource.PreviewURLFunction != nil {
		itemActions = append(itemActions, actionPreview(resource.CanView))
	}

	if resource.ActivityLog {
		itemActions = append(itemActions, actionItemHistory(resource.CanView))
	}
	resource.itemActions = append(itemActions, resource.itemActions...)

	for _, v := range resource.itemActions {
		bindResourceItemAction(a, resource, v)
	}
}

func (resource *Resource) getResourceActionsButtonData(user User, admin *Administration) (ret []buttonData) {
	navigation := admin.getResourceNavigation(*resource, user, "")
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

	if admin.Authorize(user, resource.CanEdit) && resource.OrderColumnName != "" {
		ret.ShowOrderButton = true
	}

	return ret
}

//AddItemAction adds item action
func (resource *Resource) AddItemAction(action Action) {
	resource.itemActions = append(resource.itemActions, action)
}

//AddAction adds action
func (resource *Resource) AddAction(action Action) {
	resource.actions = append(resource.actions, action)
}
