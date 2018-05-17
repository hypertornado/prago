package administration

import (
	"encoding/csv"
	"fmt"
	"github.com/hypertornado/prago"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type exportFormData struct {
	Formats []string
	Fields  []listHeaderItem

	DefaultOrderColumnName string
	DefaultOrderDesc       bool
}

func (resource Resource) getExportFormData(user User) exportFormData {
	visible := defaultVisibilityFilter
	ret := exportFormData{
		DefaultOrderColumnName: resource.OrderByColumn,
		DefaultOrderDesc:       resource.OrderDesc,
	}

	for _, v := range resource.fieldArrays {
		if visible(resource, user, *v) {
			ret.Fields = append(ret.Fields, (*v).getListHeaderItem(user))
		}
	}

	return ret
}

func exportHandler(resource Resource, request prago.Request, user User) {
	formData := resource.getExportFormData(user)
	allowedFields := map[string]bool{}
	for _, v := range formData.Fields {
		allowedFields[v.ColumnName] = true
	}

	resource.Admin.createExportActivityLog(resource, user, formData)

	usedFields := []string{}
	usedFieldsMap := map[string]bool{}
	fields := request.Request().PostForm["_field"]
	for _, v := range fields {
		if allowedFields[v] {
			usedFields = append(usedFields, v)
			usedFieldsMap[v] = true
		}
	}

	filter := map[string]string{}
	for k, _ := range allowedFields {
		filter[k] = request.Request().PostForm.Get(k)
	}

	q := resource.Admin.Query()
	orderField := request.Request().PostForm.Get("_order")
	if request.Request().PostForm.Get("_desc") == "on" {
		q.OrderDesc(orderField)
	} else {
		q.Order(orderField)
	}
	resource.addFilterToQuery(q, filter)

	limit, err := strconv.Atoi(request.Request().PostForm.Get("_limit"))
	if err == nil && limit >= 0 {
		q.Limit(int64(limit))
	}

	var rowItems interface{}
	resource.newArrayOfItems(&rowItems)
	q.Get(rowItems)

	writer := csv.NewWriter(request.Response())
	request.Response().Header().Set("Content-Type", "text/csv")

	header := []string{}
	for _, field := range resource.fieldArrays {
		if usedFieldsMap[field.ColumnName] {
			header = append(header, field.Name)
		}
	}
	err = writer.Write(header)
	if err != nil {
		panic(err)
	}

	val := reflect.ValueOf(rowItems).Elem()
	for i := 0; i < val.Len(); i++ {
		itemVal := val.Index(i).Elem()
		row := []string{}
		for _, field := range resource.fieldArrays {
			if usedFieldsMap[field.ColumnName] {
				fieldVal := itemVal.FieldByName(field.Name)
				row = append(row, removeLines(exportFieldToString(fieldVal)))
			}
		}

		err := writer.Write(row)
		if err != nil {
			panic(err)
		}
	}
	writer.Flush()
}

func exportFieldToString(value reflect.Value) string {
	switch value.Kind() {
	case reflect.String:
		return value.String()
	case reflect.Int, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%v", value.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%v", value.Float())
	case reflect.Bool:
		return fmt.Sprintf("%v", value.Bool())
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Now()) {
			tm := value.Interface().(time.Time)
			return tm.Format("2006-01-02 15:04")
		}
	}
	return "<undefined export>"
}

func removeLines(in string) string {
	ret := strings.Replace(in, "\r\n", " ", -1)
	ret = strings.Replace(ret, "\n", " ", -1)
	return ret
}
