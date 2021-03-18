package prago

import (
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hypertornado/prago/utils"
	"github.com/tealeg/xlsx"
)

func initResourceAPIs(resource *Resource) {
	resource.API("list").Handler(
		func(request Request) {
			user := request.getUser()
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(user, request.Request().URL.Query())
				must(err)
				request.RenderJSON(listDataJSON)
				return
			}
			if request.Request().URL.Query().Get("_format") == "xlsx" {
				if !resource.app.authorize(user, resource.canExport) {
					render403(request)
					return
				}
				listData, err := resource.getListContent(user, request.Request().URL.Query())
				must(err)

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
		func(request Request) {
			user := request.getUser()

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
				resource.itemToRelationData(item, user, nil),
			)
		},
	)

	resource.API("searchresource").Handler(
		func(request Request) {
			user := request.getUser()
			q := request.Params().Get("q")

			usedIDs := map[int64]bool{}

			ret := []viewRelationData{}

			id, err := strconv.Atoi(q)
			if err == nil {
				var item interface{}
				resource.newItem(&item)
				err := resource.app.Query().WhereIs("id", id).Get(item)
				if err == nil {
					relationItem := resource.itemToRelationData(item, user, nil)
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
						var item interface{}
						item = itemsVal.Index(i).Interface()
						viewItem := resource.itemToRelationData(item, user, nil)
						if viewItem != nil && usedIDs[viewItem.ID] == false {
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
				ret[k].Description = utils.Crop(ret[k].Description, 100)
			}

			request.RenderJSON(ret)
		},
	)
}