package prago

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const exportCSVPageLimit = 1000

func bindResourceExportCSV(resource *Resource) {

	app := resource.app

	PopupForm(app, "_list-export-csv", func(form *Form, request *Request) {
		form.AddHidden("_params").Value = request.Param("_params")
		form.AddSubmit("Stáhnout export csv")

	}, func(fv FormValidation, request *Request) {
		var params map[string]string
		if err := json.Unmarshal([]byte(request.Param("_params")), &params); err != nil {
			fv.AddError("Invalid params")
			return
		}

		resource := app.getResourceByID(params["_resource"])
		if !request.Authorize(resource.canExport) {
			fv.AddError("Not allowed")
			return
		}

		q := url.Values{}
		for k, v := range params {
			q.Set(k, v)
		}
		fv.Redirect(fmt.Sprintf("/admin/%s/api/export.csv?%s", resource.id, q.Encode()))

	}).Permission(loggedPermission).Icon(iconDownload).Name(unlocalized("Export"))

	resource.api("export.csv").Permission(resource.canExport).Handler(func(request *Request) {
		if !request.Authorize(resource.canExport) {
			renderErrorPage(request, 403)
			return
		}

		params := request.Request().URL.Query()

		cdValue := fmt.Sprintf("attachment; filename=\"export_%s_%s.csv\"", resource.id, time.Now().Format("2006-01-02 15:04:05"))
		request.Response().Header().Set("Content-Disposition", cdValue)

		var fieldNames []string
		var outputFields []*Field

		columnsParam := params.Get("_columns")
		var allowedColumns map[string]bool
		if columnsParam != "" {
			allowedColumns = make(map[string]bool)
			for _, col := range strings.Split(columnsParam, ",") {
				allowedColumns[strings.TrimSpace(col)] = true
			}
		}

		for _, v := range resource.fields {
			if v.authorizeView(request) {
				if allowedColumns != nil && !allowedColumns[v.id] {
					continue
				}
				outputFields = append(outputFields, v)
				fieldNames = append(fieldNames, v.id)
			}
		}

		orderBy := resource.orderByColumn
		if params.Get("_order") != "" {
			orderBy = params.Get("_order")
		}
		orderDesc := resource.orderDesc
		if params.Get("_desc") == "true" {
			orderDesc = true
		}
		if params.Get("_desc") == "false" {
			orderDesc = false
		}

		w := csv.NewWriter(request.w)
		w.Comma = ';'
		must(w.Write(fieldNames))

		err := resource.forEach(request.r.Context(), func(q *listQuery) {
			q = resource.addFilterParamsToQuery(q, params, request)
			if orderDesc {
				q = q.OrderDesc(orderBy)
			} else {
				q = q.Order(orderBy)
			}
		}, func(item any) error {

			val := reflect.ValueOf(item).Elem()

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

			return nil

		})
		must(err)
		w.Flush()
	})
}
