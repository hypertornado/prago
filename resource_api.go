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

func (resourceData *resourceData) initDefaultResourceAPIs() {
	resourceData.API("list").Handler(
		func(request *Request) {
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resourceData.getListContentJSON(request.r.Context(), request, request.Request().URL.Query())
				must(err)
				request.WriteJSON(200, listDataJSON)
				return
			}
			if request.Request().URL.Query().Get("_format") == "xlsx" {
				if !request.Authorize(resourceData.canExport) {
					render403(request)
					return
				}

				q := request.Request().URL.Query()
				q.Set("_page", "1")
				q.Set("_pagesize", fmt.Sprintf("%d", 9223372036854775807))

				listData, err := resourceData.getListContent(request.r.Context(), request, q)
				must(err)

				file := xlsx.NewFile()
				sheet, err := file.AddSheet("List 1")
				must(err)

				row := sheet.AddRow()
				columnsStr := request.Request().URL.Query().Get("_columns")
				if columnsStr == "" {
					columnsStr = resourceData.defaultVisibleFieldsStr(request)
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
						/*
							if reflect.TypeOf(v2.OriginalValue) == reflect.TypeOf(time.Now()) {
								t := v2.OriginalValue.(time.Time)
								cell.SetString(t.Format("2006-01-02"))
							} else {
								cell.SetValue(v2.OriginalValue)
							}*/
					}
				}
				file.Write(request.Response())
				return
			}
			panic("unkown format")
		},
	)

	resourceData.API("quick-action").Handler(func(request *Request) {
		quickActionAPIHandler(resourceData, request)
	}).Method("POST")

	resourceData.API("preview-relation/:id").Handler(
		func(request *Request) {
			item := resourceData.query(request.r.Context()).ID(request.Param("id"))
			if item == nil {
				render404(request)
				return
			}

			request.WriteJSON(
				200,
				resourceData.previewer(request, item).Preview(request.r.Context(), nil),
			)
		},
	)

	resourceData.API("set-order").Permission(resourceData.canUpdate).Method("POST").Handler(
		func(request *Request) {
			if resourceData.orderField == nil {
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
				item := resourceData.query(request.r.Context()).ID(id)
				resourceData.setOrderPosition(item, int64(i))
				err := resourceData.Update(request.r.Context(), item)
				must(err)
			}
			request.WriteJSON(200, true)
		},
	)

	resourceData.API("searchresource").Handler(
		func(request *Request) {
			q := request.Param("q")

			usedIDs := map[int64]bool{}

			ret := []preview{}

			id, err := strconv.Atoi(q)
			if err == nil {
				item := resourceData.query(request.r.Context()).ID(id)
				if item != nil {
					relationItem := resourceData.previewer(request, item).Preview(request.r.Context(), nil)
					if relationItem != nil {
						usedIDs[relationItem.ID] = true
						ret = append(ret, *relationItem)
					}
				}
			}

			filter := "%" + q + "%"
			for _, v := range []string{"name", "description"} {
				field := resourceData.Field(v)
				if field == nil {
					continue
				}
				items, err := resourceData.query(request.r.Context()).Limit(5).where(v+" LIKE ?", filter).OrderDesc("id").list()
				if err != nil {
					panic(err)
				}

				itemVals := reflect.ValueOf(items)
				itemLen := itemVals.Len()
				for i := 0; i < itemLen; i++ {
					viewItem := resourceData.previewer(request, itemVals.Index(i).Interface()).Preview(request.r.Context(), nil)
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

	resourceData.API("multipleaction").Method("POST").Handler(
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
				if !request.Authorize(resourceData.canCreate) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					app := request.app
					item := resourceData.query(request.r.Context()).ID(v)
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
					err := resourceData.Create(request.r.Context(), item)
					if err != nil {
						panic(fmt.Sprintf("can't create item for clone with id %d: %s", v, err))
					}

					if app.search != nil {
						go func() {
							err := resourceData.saveSearchItem(request.r.Context(), item)
							if err != nil {
								app.Log().Println(fmt.Errorf("%s", err))
							}
							app.search.flush()
						}()
					}

					if resourceData.activityLog {
						must(
							resourceData.LogActivity(request, nil, item),
						)
					}
				}

				request.app.Notification(fmt.Sprintf(
					"%d položek naklonováno", len(ids),
				)).Flash(request)

			case "delete":
				if !request.Authorize(resourceData.canDelete) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					var values url.Values = make(map[string][]string)
					values.Add("id", fmt.Sprintf("%d", v))

					valValidation := newValuesValidation(request.r.Context(), request.app, request, values)
					for _, v := range resourceData.deleteValidations {
						v(valValidation)
					}

					if !valValidation.Valid() {
						request.WriteJSON(
							403,
							valValidation.validation.TextErrorReport(v, request.Locale()),
						)
						return
					}

					item := resourceData.query(request.r.Context()).ID(v)
					if item == nil {
						panic("can't find item to delete")
					}

					err := resourceData.DeleteWithLog(item, request)
					must(err)
				}
			default:
				var foundAction bool
				for _, action := range resourceData.quickActions {
					if action.url == actionName {
						foundAction = true
						if !request.Authorize(action.permission) {
							panic("don't have access")
						}
						for _, id := range ids {
							item := resourceData.query(request.r.Context()).ID(id)
							err := action.handler(item, request)
							must(err)
						}

					}
				}
				if !foundAction {
					panic("did not find action")
				}
			}
		},
	)

	type MultipleEditData struct {
		Form      *Form
		CSRFToken string
		IDs       string
	}

	resourceData.API("multiple_edit").Permission(resourceData.canUpdate).Method("GET").Handler(
		func(request *Request) {
			form := NewForm(
				resourceData.getURL("api/multiple_edit"),
			)

			var item interface{} = reflect.New(resourceData.typ).Interface()
			resourceData.addFormItems(item, request, form)

			data := &MultipleEditData{
				Form:      form,
				CSRFToken: request.csrfToken(),
				IDs:       request.Param("ids"),
			}
			request.WriteHTML(200, "multiple_edit", data)
		},
	)

	resourceData.API("multiple_edit").Permission(resourceData.canUpdate).Method("POST").Handler(
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

				_, validation, err := resourceData.editItemWithLogAndValues(
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
