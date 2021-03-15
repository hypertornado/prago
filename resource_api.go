package prago

import (
	"reflect"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

func initResourceAPIs(resource *Resource) {
	resource.API("list").Handler(
		func(request Request) {
			user := request.GetUser()
			if request.Request().URL.Query().Get("_format") == "json" {
				listDataJSON, err := resource.getListContentJSON(user, request.Request().URL.Query())
				must(err)
				request.RenderJSON(listDataJSON)
				return
			}
			if request.Request().URL.Query().Get("_format") == "xlsx" {
				if !resource.app.Authorize(user, resource.canExport) {
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

			user := request.GetUser()

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
}
