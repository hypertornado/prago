package prago

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

func (resource *Resource) initDefaultResourceAPIs() {
	resource.api("list").Handler(
		func(request *Request) {
			request.WriteJSON(200,
				resource.getListContentJSON(
					request.r.Context(),
					request,
					request.Request().URL.Query(),
				),
			)
		},
	)

	resource.api("export").Handler(func(request *Request) {
		if !request.Authorize(resource.canExport) {
			renderErrorPage(request, 403)
			return
		}

		cdValue := fmt.Sprintf("attachment; filename=\"export_%s_%s.xlsx\"", resource.id, time.Now().Format("2006-01-02 15:04:05"))
		request.Response().Header().Set("Content-Disposition", cdValue)

		file := xlsx.NewFile()
		sheet, err := file.AddSheet(resource.pluralName(request.Locale()))
		must(err)

		row := sheet.AddRow()

		for _, field := range resource.fields {
			if request.Authorize(field.canView) {
				row.AddCell().Value = field.name(request.Locale())
			}
		}

		ids := strings.Split(request.Param("ids"), ",")
		for _, v := range ids {

			item := resource.query(request.r.Context()).ID(v)

			itemVal := reflect.ValueOf(item).Elem()

			row = sheet.AddRow()

			for _, field := range resource.fields {
				if !request.Authorize(field.canView) {
					continue
				}
				fieldValue := itemVal.FieldByName(field.fieldClassName)
				row.AddCell().SetValue(fieldValue.Interface())
			}
		}

		file.Write(request.Response())

	})

	resource.api("list-stats").Handler(func(request *Request) {
		var data = resource.getListStats(request.r.Context(), request, request.Params())
		request.WriteHTML(200, resource.app.adminTemplates, "list_stats_content", data)
	})

	resource.api("preview-relation/:ids").Handler(
		func(request *Request) {

			var previews []*Preview = []*Preview{}

			ids := strings.Split(request.Param("ids"), ";")

			for _, id := range ids {
				if id == "" || id == "0" {
					continue
				}
				item := resource.query(request.r.Context()).ID(id)
				if item == nil {
					panic(fmt.Sprintf("cant find resource item id %s", id))
				}

				previews = append(previews,
					resource.previewer(request, item).Preview(nil),
				)
			}

			request.WriteJSON(
				200,
				previews,
			)
		},
	)

	resource.api("set-order").Permission(resource.canUpdate).Method("POST").Handler(
		func(request *Request) {
			if resource.orderField == nil {
				panic("can't order")
			}

			decoder := json.NewDecoder(request.Request().Body)
			var t = map[string][]int{}
			must(decoder.Decode(&t))

			order, ok := t["order"]
			if !ok {
				panic("wrong format")
			}

			for i, id := range order {
				item := resource.query(request.r.Context()).ID(id)
				resource.setOrderPosition(item, int64(i))
				err := resource.update(request.r.Context(), item)
				must(err)
			}
			request.WriteJSON(200, true)
		},
	)

	resource.api("searchresource").Handler(func(request *Request) {
		searchResource(request, resource)
	})

	resource.api("multipleaction").Method("POST").Handler(
		func(request *Request) {
			var items []any

			idsStr := strings.Split(request.Param("ids"), ",")
			for _, v := range idsStr {
				id, err := strconv.Atoi(v)
				if err != nil {
					panic(fmt.Sprintf("can't convert str '%s' to int", v))
				}
				item := resource.query(request.r.Context()).ID(id)
				if item == nil {
					panic(fmt.Sprintf("can't find item %d", id))
				}
				items = append(items, item)
			}

			actionName := request.Param("action")

			var multiItemAction *MultipleItemAction
			for _, action := range resource.multipleActions {
				if action.ID == actionName {
					multiItemAction = action
				}
			}

			if !request.Authorize(multiItemAction.Permission) {
				renderAPINotAuthorized(request)
				return
			}

			response := &MultipleItemActionResponse{}
			multiItemAction.Handler(items, request, response)
			request.WriteJSON(200, response)
		},
	)

	type MultipleEditData struct {
		Form      *Form
		CSRFToken string
		IDs       string
	}

	resource.api("multiple_edit").Permission(resource.canUpdate).Method("GET").Handler(
		func(request *Request) {
			form := resource.app.NewForm(
				resource.getURL("api/multiple_edit"),
			)

			var item interface{} = reflect.New(resource.typ).Interface()
			form.initWithResourceItem(resource, item, request)

			data := &MultipleEditData{
				Form:      form,
				CSRFToken: request.csrfToken(),
				IDs:       request.Param("ids"),
			}
			request.WriteHTML(200, request.app.adminTemplates, "multiple_edit", data)
		},
	)

	resource.api("multiple_edit").Permission(resource.canUpdate).Method("POST").Handler(
		func(request *Request) {

			validateCSRF(request)

			idsStr := request.Request().PostForm["_ids"][0]
			fields := request.Request().PostForm["_fields"]

			var fieldsMap = map[string]bool{}
			for _, v := range fields {
				fieldsMap[v] = true
			}

			for _, idStr := range strings.Split(idsStr, ",") {
				id, err := strconv.Atoi(idStr)
				must(err)

				var usedValues url.Values = make(map[string][]string)
				usedValues.Set("id", idStr)
				for k := range fieldsMap {
					usedValues.Add(k, request.Request().PostForm.Get(k))
				}

				_, validation := resource.editItemWithLogAndValues(
					request,
					usedValues,
				)

				if !validation.Valid() {
					report := validation.TextErrorReport(int64(id), request.Locale())
					request.WriteJSON(
						403,
						report,
					)
					return
				}
			}

		},
	)
}
