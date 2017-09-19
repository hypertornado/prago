package admin

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type list struct {
	Name           string
	TypeID         string
	Actions        []ButtonData
	ItemActions    []ButtonData
	Colspan        int64
	Header         []listHeader
	CanChangeOrder bool
	OrderColumn    string
	OrderDesc      bool
}

type listHeader struct {
	Name         string
	NameHuman    string
	ColumnName   string
	CanOrder     bool
	FilterLayout string
}

type listContent struct {
	TotalCount int64
	Rows       []listRow
	Pagination pagination
	Colspan    int64
}

type listRow struct {
	ID      int64
	Items   []listCell
	Actions []ButtonData
}

type listCell struct {
	TemplateName string
	Value        string
	URL          string
}

type pagination struct {
	Pages []page
}

type page struct {
	Page    int64
	Current bool
}

type listRequest struct {
	Page      int64
	OrderBy   string
	OrderDesc bool
	Filter    map[string]string
}

func (resource *Resource) getListHeader(admin *Admin, path string, user *User) (list list, err error) {
	lang := user.Locale

	list.Colspan = 1
	list.TypeID = resource.ID
	list.Actions = resource.ResourceActionsButtonData(user, admin)

	list.OrderColumn = resource.OrderByColumn
	list.OrderDesc = resource.OrderDesc

	orderField, ok := resource.StructCache.fieldMap[resource.OrderByColumn]
	if !ok || !orderField.CanOrder {
		err = ErrItemNotFound
		return
	}

	list.Name = resource.Name(lang)

	if resource.StructCache.OrderColumnName == list.OrderColumn && !list.OrderDesc {
		list.CanChangeOrder = true
	}

	for _, v := range resource.StructCache.fieldArrays {
		if v.canShow() {
			list.Colspan++
			headerItem := listHeader{
				Name:       v.Name,
				NameHuman:  v.humanName(lang),
				ColumnName: v.ColumnName,
			}

			headerItem.FilterLayout = v.filterLayout()

			if v.CanOrder {
				headerItem.CanOrder = true
			}
			list.Header = append(list.Header, headerItem)
		}
	}

	return
}

func (sf *structField) filterLayout() string {
	if sf == nil {
		return ""
	}
	if sf.Typ.Kind() == reflect.String &&
		(sf.Tags["prago-type"] == "" || sf.Tags["prago-type"] == "text" || sf.Tags["prago-type"] == "markdown") {
		return "filter_layout_text"
	}

	if sf.Typ.Kind() == reflect.Int64 || sf.Typ.Kind() == reflect.Int {
		return "filter_layout_number"
	}

	if sf.Typ.Kind() == reflect.Bool {
		return "filter_layout_boolean"
	}
	return ""
}

func (resource *Resource) addFilterToQuery(q *Query, filter map[string]string) {
	for k, v := range filter {
		field := resource.StructCache.fieldMap[k]
		if field == nil {
			continue
		}

		layout := field.filterLayout()

		switch layout {
		case "filter_layout_text":
			v = "%" + v + "%"
			k = strings.Replace(k, "`", "", -1)
			str := fmt.Sprintf("`%s` LIKE ?", k)
			q = q.Where(str, v)
		case "filter_layout_number":
			v = strings.Trim(v, " ")
			numVal, err := strconv.Atoi(v)
			if err == nil {
				q.WhereIs(k, numVal)
			}
		case "filter_layout_boolean":
			switch v {
			case "true":
				q.WhereIs(k, true)
			case "false":
				q.WhereIs(k, false)
			}
		}
	}
}

func (resource *Resource) getListContent(admin *Admin, path string, requestQuery *listRequest, user *User) (list listContent, err error) {
	q := admin.Query()
	if requestQuery.OrderDesc {
		q.OrderDesc(requestQuery.OrderBy)
	} else {
		q.Order(requestQuery.OrderBy)
	}

	var count int64
	var item interface{}
	resource.newItem(&item)
	countQuery := admin.Query()
	resource.addFilterToQuery(countQuery, requestQuery.Filter)
	count, err = countQuery.Count(item)
	if err != nil {
		return
	}
	list.TotalCount = count

	totalPages := (count / resource.Pagination)
	if count%resource.Pagination != 0 {
		totalPages += +1
	}

	var currentPage int64 = requestQuery.Page

	if totalPages > 1 {
		for i := int64(1); i <= totalPages; i++ {
			p := page{
				Page: i,
			}
			if i == currentPage {
				p.Current = true
			}

			list.Pagination.Pages = append(list.Pagination.Pages, p)
		}
	}

	resource.addFilterToQuery(q, requestQuery.Filter)
	q.Offset((currentPage - 1) * resource.Pagination)
	q.Limit(resource.Pagination)

	var rowItems interface{}
	resource.newItems(&rowItems)
	q.Get(rowItems)

	val := reflect.ValueOf(rowItems).Elem()
	for i := 0; i < val.Len(); i++ {
		row := listRow{}
		itemVal := val.Index(i).Elem()

		for _, v := range resource.StructCache.fieldArrays {
			if v.canShow() {
				structField, _ := resource.Typ.FieldByName(v.Name)
				fieldVal := itemVal.FieldByName(v.Name)
				row.Items = append(row.Items, resource.valueToCell(admin, structField, fieldVal))
			}
		}

		row.ID = itemVal.FieldByName("ID").Int()
		row.Actions = resource.ResourceItemActionsButtonData(user, row.ID, admin)
		list.Rows = append(list.Rows, row)
		list.Colspan = int64(len(row.Items)) + 1
	}

	return
}

func (resource *Resource) valueToCell(admin *Admin, field reflect.StructField, val reflect.Value) (cell listCell) {
	cell.TemplateName = "admin_string"
	var item interface{}
	reflect.ValueOf(&item).Elem().Set(val)

	switch item.(type) {
	case string:
		cell.Value = item.(string)
	case bool:
		cell.TemplateName = "admin_cell_checkbox"
		if item.(bool) {
			cell.Value = "true"
		}
	case int64:
		cell.Value = fmt.Sprintf("%d", item.(int64))
		if field.Tag.Get("prago-type") == "relation" {
			relationResource := resource.admin.getResourceByName(field.Name)

			var relationItem interface{}
			relationResource.newItem(&relationItem)
			err := admin.Query().WhereIs("id", item.(int64)).Get(relationItem)
			if err != nil {
				if err == ErrItemNotFound {
					cell.Value = ""
					return
				}
				panic(err)
			}

			ifaceItemName, ok := relationItem.(interface {
				AdminItemName(string) string
			})
			if ok {
				cell.Value = ifaceItemName.AdminItemName("cs")
				cell.TemplateName = "admin_link"
				cell.URL = fmt.Sprintf("%s/%d/edit", relationResource.ID, item.(int64))
				return
			}

			nameField := reflect.ValueOf(relationItem).Elem().FieldByName("Name")

			cell.Value = nameField.String()
			cell.TemplateName = "admin_link"
			cell.URL = fmt.Sprintf("%s/%d/edit", relationResource.ID, item.(int64))
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
