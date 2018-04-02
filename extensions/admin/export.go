package admin

import (
	"fmt"
	"github.com/hypertornado/prago"
)

type exportFormData struct {
	Formats []string
	Fields  []exportFormDataField
}

type exportFormDataField struct {
	NameHuman  string
	ColumnName string
}

func (cache structCache) getExportFormData(user User, visible structFieldFilter) exportFormData {
	ret := exportFormData{
		Formats: []string{"csv"},
	}

	for _, v := range cache.fieldArrays {
		field := exportFormDataField{
			NameHuman:  v.humanName(user.Locale),
			ColumnName: v.ColumnName,
		}

		ret.Fields = append(ret.Fields, field)
	}

	return ret
}

func exportHandler(admin Admin, resource Resource, request prago.Request, user User) {

	if request.Params().Get("_format") == "csv" {
		exportHandlerCSV(admin, resource, request, user)
		return
	}
	panic("wrong format of export")
}

func exportHandlerCSV(admin Admin, resource Resource, request prago.Request, user User) {
	formData := resource.StructCache.getExportFormData(user, resource.VisibilityFilter)

	allowedFields := map[string]bool{}
	for _, v := range formData.Fields {
		allowedFields[v.ColumnName] = true
	}

	usedFields := []string{}
	fields := request.Request().PostForm["_field"]
	for _, v := range fields {
		if allowedFields[v] {
			usedFields = append(usedFields, v)
		}
	}

	filter := map[string]string{}
	for k, _ := range allowedFields {
		filter[k] = request.Request().PostForm.Get(k)
	}

	q := admin.Query()
	q = resource.addFilterToQuery(q, filter)

	var rowItems interface{}
	resource.newItems(&rowItems)
	q.Get(rowItems)

	fmt.Println(filter)
	panic("EXPOOOOORT")
}
