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
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(request.r.Context(), request, request.Request().URL.Query())
				must(err)
				request.WriteJSON(200, listDataJSON)
				return
			}
			if request.Request().URL.Query().Get("_format") == "xlsx" {
				if !request.Authorize(resource.canExport) {
					renderErrorPage(request, 403)
					return
				}

				q := request.Request().URL.Query()
				q.Set("_page", "1")
				q.Set("_pagesize", fmt.Sprintf("%d", 9223372036854775807))

				listData, err := resource.getListContent(request.r.Context(), request, q)
				must(err)

				file := xlsx.NewFile()
				sheet, err := file.AddSheet("List 1")
				must(err)

				row := sheet.AddRow()
				columnsStr := request.Request().URL.Query().Get("_columns")
				if columnsStr == "" {
					columnsStr = resource.defaultVisibleFieldsStr(request)
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
						cell.SetValue(v2.Name)
					}
				}
				file.Write(request.Response())
				return
			}
			panic("unkown format")
		},
	)

	resource.api("preview-relation/:id").Handler(
		func(request *Request) {
			item := resource.query(request.r.Context()).ID(request.Param("id"))
			if item == nil {
				renderErrorPage(request, 404)
				return
			}

			request.WriteJSON(
				200,
				resource.previewer(request, item).Preview(request.r.Context(), nil),
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

	resource.api("searchresource").Handler(
		func(request *Request) {
			q := request.Param("q")

			usedIDs := map[int64]bool{}

			ret := []preview{}

			id, err := strconv.Atoi(q)
			if err == nil {
				item := resource.query(request.r.Context()).ID(id)
				if item != nil {
					relationItem := resource.previewer(request, item).Preview(request.r.Context(), nil)
					if relationItem != nil {
						usedIDs[relationItem.ID] = true
						ret = append(ret, *relationItem)
					}
				}
			}

			filter := "%" + q + "%"
			for _, v := range []string{"name", "description"} {
				field := resource.Field(v)
				if field == nil {
					continue
				}
				items, err := resource.query(request.r.Context()).Limit(5).where(v+" LIKE ?", filter).OrderDesc("id").list()
				if err != nil {
					panic(err)
				}

				itemVals := reflect.ValueOf(items)
				itemLen := itemVals.Len()
				for i := 0; i < itemLen; i++ {
					viewItem := resource.previewer(request, itemVals.Index(i).Interface()).Preview(request.r.Context(), nil)
					if viewItem != nil && !usedIDs[viewItem.ID] {
						usedIDs[viewItem.ID] = true
						ret = append(ret, *viewItem)
					}
				}
			}

			if len(ret) > 5 {
				ret = ret[0:5]
			}

			for k := range ret {
				ret[k].Description = crop(ret[k].Description, 100)
			}

			request.WriteJSON(200, ret)
		},
	)

	resource.api("multipleaction").Method("POST").Handler(
		func(request *Request) {
			var ids []int64

			idsStr := strings.Split(request.Param("ids"), ",")
			for _, v := range idsStr {
				id, err := strconv.Atoi(v)
				if err != nil {
					panic(fmt.Sprintf("can't convert str '%s' to int", v))
				}
				ids = append(ids, int64(id))
			}

			actionName := request.Param("action")

			switch actionName {
			case "clone":
				if !request.Authorize(resource.canCreate) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					item := resource.query(request.r.Context()).ID(v)
					if item == nil {
						panic(fmt.Sprintf("can't get item for clone with id %d", v))
					}
					val := reflect.ValueOf(item).Elem()
					val.FieldByName("ID").SetInt(0)
					timeVal := reflect.ValueOf(time.Now())
					for _, fieldName := range []string{"CreatedAt", "UpdatedAt"} {
						field := val.FieldByName(fieldName)
						if field.IsValid() && field.CanSet() && field.Type() == timeVal.Type() {
							field.Set(timeVal)
						}
					}

					//TODO: log for creation
					err := resource.create(request.r.Context(), item)
					if err != nil {
						panic(fmt.Sprintf("can't create item for clone with id %d: %s", v, err))
					}

					if resource.activityLog {
						must(
							resource.logActivity(request, nil, item),
						)
					}
				}

				request.app.Notification(fmt.Sprintf(
					"%d položek naklonováno", len(ids),
				)).Flash(request)

			case "delete":
				if !request.Authorize(resource.canDelete) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					var values url.Values = make(map[string][]string)
					values.Add("id", fmt.Sprintf("%d", v))

					valValidation := newValuesValidation(request.r.Context(), request.app, request, values)
					for _, v := range resource.deleteValidations {
						v(valValidation)
					}

					if !valValidation.Valid() {
						request.WriteJSON(
							403,
							valValidation.validation.TextErrorReport(v, request.Locale()),
						)
						return
					}

					item := resource.query(request.r.Context()).ID(v)
					if item == nil {
						panic("can't find item to delete")
					}

					err := resource.deleteWithLog(item, request)
					must(err)
				}
			default:
				panic("did not find action")
			}
		},
	)

	type MultipleEditData struct {
		Form      *Form
		CSRFToken string
		IDs       string
	}

	resource.api("multiple_edit").Permission(resource.canUpdate).Method("GET").Handler(
		func(request *Request) {
			form := NewForm(
				resource.getURL("api/multiple_edit"),
			)

			var item interface{} = reflect.New(resource.typ).Interface()
			form.addResourceItems(resource, item, request)

			data := &MultipleEditData{
				Form:      form,
				CSRFToken: request.csrfToken(),
				IDs:       request.Param("ids"),
			}
			request.WriteHTML(200, "multiple_edit", data)
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

				_, validation, err := resource.editItemWithLogAndValues(
					request,
					usedValues,
				)

				if err == errValidation {
					report := validation.Validation().TextErrorReport(int64(id), request.Locale())
					request.WriteJSON(
						403,
						report,
					)
					return
				}
				must(err)
			}

		},
	)
}
