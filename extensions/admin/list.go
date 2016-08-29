package admin

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

type List struct {
	Header     []ListHeader
	Rows       []ListRow
	Pagination Pagination
	Order      bool
	HasDelete  bool
	HasNew     bool
}

type ListHeader struct {
	Name      string
	NameHuman string
}

type ListRow struct {
	ID    int64
	Items []ListCell
}

type ListCell struct {
	TemplateName string
	Value        string
	Url          string
}

type Pagination struct {
	Prev  Page
	Next  Page
	Pages []Page
}

type Page struct {
	Name    string
	Url     string
	Current bool
}

func (resource *AdminResource) GetList(lang string, path string, requestQuery url.Values) (list List, err error) {
	q := resource.Query()
	if resource.OrderDesc {
		q = q.OrderDesc(resource.OrderByColumn)
	} else {
		q = q.Order(resource.OrderByColumn)
	}

	_, list.HasDelete = resource.Actions["delete"]
	_, list.HasNew = resource.Actions["new"]

	if resource.StructCache.OrderColumnName == resource.OrderByColumn && !resource.OrderDesc {
		list.Order = true
	}

	var count int64
	count, err = q.Count()
	if err != nil {
		return
	}

	totalPages := (count / resource.Pagination)
	if count%resource.Pagination != 0 {
		totalPages += +1
	}

	var currentPage int64 = 1
	queryPage := requestQuery.Get("p")
	if len(queryPage) > 0 {
		convertedPage, err := strconv.Atoi(queryPage)
		if err == nil && convertedPage > 1 {
			currentPage = int64(convertedPage)
		}
	}

	if totalPages > 1 {
		for i := int64(1); i <= totalPages; i++ {
			p := Page{}
			p.Name = fmt.Sprintf("%d", i)
			if i == currentPage {
				p.Current = true
			}

			p.Url = path
			if i > 1 {
				newUrlValues := make(url.Values)
				newUrlValues.Set("p", fmt.Sprintf("%d", i))
				p.Url += "?" + newUrlValues.Encode()
			}

			list.Pagination.Pages = append(list.Pagination.Pages, p)
		}
	}

	q.Offset((currentPage - 1) * resource.Pagination)
	q.Limit(resource.Pagination)

	rowItems, err := q.List()

	for _, v := range resource.StructCache.fieldArrays {
		show := false
		if v.Name == "ID" || v.Name == "Name" {
			show = true
		}
		showTag := v.Tags["prago-preview"]
		if showTag == "true" {
			show = true
		}
		if showTag == "false" {
			show = false
		}

		if show {
			list.Header = append(list.Header, ListHeader{Name: v.Name, NameHuman: v.humanName(lang)})
		}
	}

	val := reflect.ValueOf(rowItems)
	for i := 0; i < val.Len(); i++ {
		row := ListRow{}
		itemVal := val.Index(i).Elem()

		for _, h := range list.Header {
			structField, _ := resource.Typ.FieldByName(h.Name)
			fieldVal := itemVal.FieldByName(h.Name)
			row.Items = append(row.Items, resource.ValueToCell(structField, fieldVal))
		}
		row.ID = itemVal.FieldByName("ID").Int()
		list.Rows = append(list.Rows, row)
	}
	return
}

func (resource *AdminResource) ValueToCell(field reflect.StructField, val reflect.Value) (cell ListCell) {
	cell.TemplateName = "admin_string"
	var item interface{}
	reflect.ValueOf(&item).Elem().Set(val)

	switch item.(type) {
	case string:
		cell.Value = item.(string)
	case bool:
		if item.(bool) {
			cell.Value = "✅"
		}
	case int64:
		cell.Value = fmt.Sprintf("%d", item.(int64))
		if field.Tag.Get("prago-type") == "relation" {
			relationResource := resource.admin.GetResourceByName(field.Name)
			relationItem, err := relationResource.Query().Where(map[string]interface{}{"id": item.(int64)}).First()
			if err != nil {
				panic(err)
			}

			nameField := reflect.ValueOf(relationItem).Elem().FieldByName("Name")
			cell.Value = nameField.String()
			cell.TemplateName = "admin_link"
			cell.Url = fmt.Sprintf("%s/%d", relationResource.ID, item.(int64))
			return
		}
	}

	if field.Tag.Get("prago-type") == "image" {
		cell.TemplateName = "admin_image"
	}

	if val.Type() == reflect.TypeOf(time.Now()) {
		var tm time.Time
		reflect.ValueOf(&tm).Elem().Set(val)
		cell.Value = tm.Format("2006-01-02 15:04:05")
	}

	if len(field.Tag.Get("prago-preview-type")) > 0 {
		cell.TemplateName = field.Tag.Get("prago-preview-type")
	}

	return
}
