package prago

import (
	"encoding/csv"
	"fmt"
	"reflect"
	"time"
)

const exportCSVPageLimit = 1000

func bindResourceExportCSV(resourceData *Resource) {
	resourceData.api("export.csv").Permission(resourceData.canExport).Handler(func(request *Request) {
		if !request.Authorize(resourceData.canExport) {
			renderErrorPage(request, 403)
			return
		}

		var fieldNames []string
		var outputFields []*Field
		for _, v := range resourceData.fields {
			if v.authorizeView(request) {
				outputFields = append(outputFields, v)
				fieldNames = append(fieldNames, v.id)
			}
		}

		w := csv.NewWriter(request.w)
		w.Comma = ';'
		must(w.Write(fieldNames))

		iteration := 0

		for {
			q := resourceData.query(request.r.Context())
			q.offset = int64(iteration * exportCSVPageLimit)
			q.limit = exportCSVPageLimit
			items, err := q.list()
			must(err)

			itemsCount := reflect.ValueOf(items).Len()
			if itemsCount == 0 {
				break
			}

			for i := 0; i < itemsCount; i++ {
				val := reflect.ValueOf(items).Index(i).Elem()

				var strValuesRow []string

				for _, outputField := range outputFields {
					valIface := val.FieldByName(outputField.fieldClassName).Interface()
					strVal := fmt.Sprintf("%v", valIface)
					if t, ok := valIface.(time.Time); ok {
						strVal = t.Format(time.RFC3339)
					}
					strValuesRow = append(strValuesRow, strVal)
				}
				must(w.Write(strValuesRow))

			}
			iteration++

		}
		w.Flush()
	})
}
