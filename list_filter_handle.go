package prago

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/net/context"
)

func listFilterFormHandle(fv FormValidation, field *Field, request *Request) {

	var val string

	val = listFilterFormGetValues(fv, field, request)
	if !fv.Valid() {
		return
	}

	fv.Data(listFilterGetResponse(val, field, request))

	filterLayout := field.fieldType.filterLayoutTemplate

	if filterLayout == "filter_layout_select" {
		listFilterFormSelectHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_relation" {
		listFilterFormRelationHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_date" {
		listFilterFormDateHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_boolean" {
		listFilterFormBooleanHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_text" {
		listFilterFormTextHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_number" {
		listFilterFormNumberHandle(fv, field, request)
	}

}

func listFilterFormGetValues(fv FormValidation, field *Field, request *Request) string {

	filterLayout := field.fieldType.filterLayoutTemplate

	if filterLayout == "filter_layout_select" {
		return listFilterFormSelectHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_relation" {
		return listFilterFormRelationHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_date" {
		return listFilterFormDateHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_boolean" {
		return listFilterFormBooleanHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_text" {
		return listFilterFormTextHandle(fv, field, request)
	}

	if filterLayout == "filter_layout_number" {
		return listFilterFormNumberHandle(fv, field, request)
	}
	panic("unknown filter layout " + filterLayout)
}

func listFilterFormSelectHandle(fv FormValidation, field *Field, request *Request) string {
	var values []string
	var dataItems = field.fieldType.filterLayoutDataSource(field, request).([]*FormOption)
	for _, item := range dataItems {
		if item.ID == "" {
			continue
		}
		if request.Param(field.id+"_"+item.ID) == "on" {
			values = append(values, item.ID)
		}
	}

	return strings.Join(values, ",")
}

func listFilterFormRelationHandle(fv FormValidation, field *Field, request *Request) string {
	var values []string
	ints := MultirelationStringToArray(request.Param(field.id))
	for _, i := range ints {
		values = append(values, fmt.Sprintf("%d", i))
	}
	return strings.Join(values, ",")
}

func listFilterFormDateHandle(fv FormValidation, field *Field, request *Request) string {
	fromID := field.id + "_from"
	toID := field.id + "_to"

	if request.Param(fromID) != "" {
		_, err := time.Parse("2006-01-02", request.Param(fromID))
		if err != nil {
			fv.AddItemError(fromID, "Neplatný formát")
		}
	}
	if request.Param(toID) != "" {
		_, err := time.Parse("2006-01-02", request.Param(toID))
		if err != nil {
			fv.AddItemError(toID, "Neplatný formát")
		}
	}

	if !fv.Valid() {
		return ""
	}

	val := fmt.Sprintf("%s,%s", request.Param(fromID), request.Param(toID))
	if val == "," {
		val = ""
	}
	return val
}

func listFilterFormBooleanHandle(fv FormValidation, field *Field, request *Request) string {
	return request.Param(field.id)
}

func listFilterFormTextHandle(fv FormValidation, field *Field, request *Request) string {
	return request.Param(field.id)
}

func listFilterFormNumberHandle(fv FormValidation, field *Field, request *Request) string {
	return request.Param(field.id)
}

func listFilterGetResponse(value string, field *Field, request *Request) (ret *ListFilterResponse) {
	ret = &ListFilterResponse{}
	ret.ID = value
	ret.Name = value

	filterLayout := field.fieldType.filterLayoutTemplate

	if filterLayout == "filter_layout_select" {
		var names []string
		nameMap := map[string]string{}
		var dataItems = field.fieldType.filterLayoutDataSource(field, request).([]*FormOption)
		for _, item := range dataItems {
			nameMap[item.ID] = item.Name

		}
		valItems := strings.Split(value, ",")
		for _, valItem := range valItems {
			names = append(names, nameMap[valItem])
		}
		ret.Name = strings.Join(names, " nebo ")
	}

	if filterLayout == "filter_layout_relation" {
		var names []string
		if value != "" {
			ids := strings.Split(value, ",")
			for _, id := range ids {
				var name = fmt.Sprintf("#%s", id)
				item := field.relatedResource.query(context.Background()).ID(id)
				if item != nil {
					prev := field.relatedResource.previewer(request, item)
					name = prev.Name()
				}
				names = append(names, name)
			}
		}
		ret.Name = strings.Join(names, " nebo ")
	}

	if filterLayout == "filter_layout_date" {
		var names []string
		if ret.ID != "" {
			fields := strings.Split(ret.ID, ",")
			if fields[0] != "" {
				t, err := time.Parse("2006-01-02", fields[0])
				if err == nil {
					names = append(names, fmt.Sprintf("od %s", t.Format("2. 1. 2006")))
				}
			}
			if fields[1] != "" {
				t, err := time.Parse("2006-01-02", fields[1])
				if err == nil {
					names = append(names, fmt.Sprintf("do %s", t.Format("2. 1. 2006")))
				}
			}
		}
		ret.Name = strings.Join(names, " · ")
	}

	if filterLayout == "filter_layout_boolean" {
		if value == "true" {
			ret.Name = "✅ ano"
		}
		if value == "false" {
			ret.Name = "ne"
		}
	}
	return
}
