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

func (resource *Resource[T]) initDefaultResourceAPIs() {
	newResourceAPI(resource, "list").Handler(
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

	newResourceAPI(resource, "preview-relation/:id").Handler(
		func(request *Request) {
			item := resource.Is("id", request.Params().Get("id")).First()
			if item == nil {
				render404(request)
				return
			}

			request.RenderJSON(
				resource.itemToRelationData(item, request.user, nil),
			)
		},
	)

	newResourceAPI(resource, "set-order").Permission(resource.canUpdate).Method("POST").Handler(
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
				item := resource.Is("id", int64(id)).First()
				resource.setOrderPosition(item, int64(i))
				err := resource.Update(item)
				must(err)
			}
			request.RenderJSON(true)
		},
	)

	newResourceAPI(resource, "searchresource").Handler(
		func(request *Request) {
			q := request.Params().Get("q")

			usedIDs := map[int64]bool{}

			ret := []preview{}

			id, err := strconv.Atoi(q)
			if err == nil {
				item := resource.Is("id", id).First()
				if item != nil {
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
				items := resource.Query().Limit(5).Where(v+" LIKE ?", filter).List()
				itemsVal := reflect.ValueOf(items)
				for i := 0; i < itemsVal.Len(); i++ {
					item := itemsVal.Index(i).Interface()
					viewItem := resource.itemToRelationData(item, request.user, nil)
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

			request.RenderJSON(ret)
		},
	)

	newResourceAPI(resource, "multipleaction").Method("POST").Handler(
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
			case "clone":
				if !request.app.authorize(request.user, resource.canCreate) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					app := request.app
					item := resource.Is("id", v).First()
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
					err := resource.Create(item)
					if err != nil {
						panic(fmt.Sprintf("can't create item for clone with id %d: %s", v, err))
					}

					if app.search != nil {
						go func() {
							err := app.search.saveItem(resource, item)
							if err != nil {
								app.Log().Println(fmt.Errorf("%s", err))
							}
							app.search.flush()
						}()
					}

					if resource.activityLog {
						must(
							resource.LogActivity(request.user, nil, item),
						)
					}
				}

				request.app.Notification(fmt.Sprintf(
					"%d položek naklonováno", len(ids),
				)).Flash(request)

			case "delete":
				if !request.app.authorize(request.user, resource.canDelete) {
					renderAPINotAuthorized(request)
					return
				}
				for _, v := range ids {
					var values url.Values = make(map[string][]string)
					values.Add("id", fmt.Sprintf("%d", v))

					valValidation := newValuesValidation(request.user.Locale, values)
					for _, v := range resource.deleteValidations {
						v(valValidation)
					}

					if !valValidation.Valid() {
						request.RenderJSONWithCode(
							valValidation.validation.TextErrorReport(v, request.user.Locale),
							403,
						)
						return
					}

					item := resource.Is("id", v).First()
					if item == nil {
						panic("can't find item to delete")
					}

					err := resource.DeleteWithLog(item, request)
					must(err)
				}
			default:
				panic(fmt.Sprintf("unknown action: %s", request.Params().Get("action")))
			}
		},
	)

	newResourceAPI(resource, "multiple_edit").Permission(resource.canUpdate).Method("GET").Handler(
		func(request *Request) {
			var item T
			form := NewForm(
				resource.getURL("api/multiple_edit"),
			)
			resource.addFormItems(&item, request.user, form)
			request.SetData("form", form)

			request.SetData("CSRFToken", request.csrfToken())
			request.SetData("ids", request.Params().Get("ids"))

			request.RenderView("multiple_edit")
		},
	)

	newResourceAPI(resource, "multiple_edit").Permission(resource.canUpdate).Method("POST").Handler(
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
					report := validation.Validation().TextErrorReport(int64(id), request.user.Locale)
					request.RenderJSONWithCode(
						report,
						403,
					)
					return
				}
				must(err)
			}

		},
	)
}
