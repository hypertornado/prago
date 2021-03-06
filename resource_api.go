package prago

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

func initDefaultResourceAPIs(resource *Resource) {
	resource.API("list").Handler(
		func(request *Request) {
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(request.user, request.Request().URL.Query())
				must(err)
				request.RenderJSON(listDataJSON)
				return
			}
			if request.Request().URL.Query().Get("_format") == "xlsx" {
				if !resource.app.authorize(request.user, resource.canExport) {
					render403(request)
					return
				}
				listData, err := resource.getListContent(request.user, request.Request().URL.Query())
				must(err)

				file := xlsx.NewFile()
				sheet, err := file.AddSheet("List 1")
				must(err)

				row := sheet.AddRow()
				columnsStr := request.Request().URL.Query().Get("_columns")
				if columnsStr == "" {
					columnsStr = resource.defaultVisibleFieldsStr(request.user)
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
						if reflect.TypeOf(v2.OriginalValue) == reflect.TypeOf(time.Now()) {
							t := v2.OriginalValue.(time.Time)
							cell.SetString(t.Format("2006-01-02"))
						} else {
							cell.SetValue(v2.OriginalValue)
						}
					}
				}
				file.Write(request.Response())
				return
			}
			panic("unkown format")
		},
	)

	resource.API("preview-relation/:id").Handler(
		func(request *Request) {
			var item interface{}
			resource.newItem(&item)
			err := resource.app.Query().WhereIs("id", request.Params().Get("id")).Get(item)
			if err == ErrItemNotFound {
				render404(request)
				return
			}
			if err != nil {
				panic(err)
			}

			request.RenderJSON(
				resource.itemToRelationData(item, request.user, nil),
			)
		},
	)

	resource.API("set-order").Permission(resource.canEdit).Method("POST").Handler(
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
				var item interface{}
				resource.newItem(&item)
				must(resource.app.Query().WhereIs("id", int64(id)).Get(item))
				must(resource.setOrderPosition(item, int64(i)))
				must(resource.app.Save(item))
			}
			request.RenderJSON(true)
		},
	)

	resource.API("searchresource").Handler(
		func(request *Request) {
			q := request.Params().Get("q")

			usedIDs := map[int64]bool{}

			ret := []viewRelationData{}

			id, err := strconv.Atoi(q)
			if err == nil {
				var item interface{}
				resource.newItem(&item)
				err := resource.app.Query().WhereIs("id", id).Get(item)
				if err == nil {
					relationItem := resource.itemToRelationData(item, request.user, nil)
					if relationItem != nil {
						usedIDs[relationItem.ID] = true
						ret = append(ret, *relationItem)
					}
				}
			}

			filter := "%" + q + "%"
			for _, v := range []string{"name", "description"} {
				field := resource.fieldMap[v]
				if field == nil {
					continue
				}
				var items interface{}
				resource.newArrayOfItems(&items)
				err := resource.app.Query().Limit(5).Where(v+" LIKE ?", filter).Get(items)
				if err == nil {
					itemsVal := reflect.ValueOf(items).Elem()
					for i := 0; i < itemsVal.Len(); i++ {
						item := itemsVal.Index(i).Interface()
						viewItem := resource.itemToRelationData(item, request.user, nil)
						if viewItem != nil && !usedIDs[viewItem.ID] {
							usedIDs[viewItem.ID] = true
							ret = append(ret, *viewItem)
						}
					}
				}
			}

			if len(ret) > 5 {
				ret = ret[0:5]
			}

			for k := range ret {
				ret[k].Description = crop(ret[k].Description, 100)
			}

			request.RenderJSON(ret)
		},
	)

	resource.API("multipleaction").Method("POST").Handler(
		func(request *Request) {
			var ids []int64

			idsStr := strings.Split(request.Params().Get("ids"), ",")
			for _, v := range idsStr {
				id, err := strconv.Atoi(v)
				if err != nil {
					panic(fmt.Sprintf("can't convert str '%s' to int", v))
				}
				ids = append(ids, int64(id))
			}

			switch request.Params().Get("action") {
			case "delete":
				if !request.app.authorize(request.user, resource.canDelete) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					err := resource.deleteItemWithLog(request.user, v)
					must(err)
				}
			default:
				panic(fmt.Sprintf("unknown action: %s", request.Params().Get("action")))
			}
		},
	)

	type MultipleEditFormItem struct {
		ID       string
		Name     string
		Template string
		Data     interface{}
	}

	resource.API("multiple_edit").Permission(resource.canEdit).Method("GET").Handler(
		func(request *Request) {
			var items []MultipleEditFormItem

			var item interface{}
			resource.newItem(&item)
			form, err := resource.getForm(item, request.user)
			form.Action = resource.getURL("api/multiple_edit")
			must(err)
			request.SetData("form", form)

			request.SetData("CSRFToken", request.csrfToken())
			request.SetData("ids", request.Params().Get("ids"))

			request.SetData("items", items)
			request.RenderView("multiple_edit")
		},
	)

	resource.API("multiple_edit").Permission(resource.canEdit).Method("POST").Handler(
		func(request *Request) {
			validateCSRF(request)

			idsStr := request.Request().PostForm["_ids"][0]
			fields := request.Request().PostForm["_fields"]

			var ids []int64

			for _, v := range strings.Split(idsStr, ",") {
				id, err := strconv.Atoi(v)
				must(err)
				ids = append(ids, int64(id))

			}

			var fieldsMap = map[string]bool{}
			for _, v := range fields {
				fieldsMap[v] = true
			}

			values := request.Request().PostForm

			_, err := resource.editItemsWithLog(request.user, ids, values, fieldsMap)
			must(err)

		},
	)
}
