package administration

import (
	"fmt"
	"github.com/hypertornado/prago/administration/messages"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type list struct {
	Name           string
	TypeID         string
	Colspan        int64
	Header         []listHeaderItem
	CanChangeOrder bool
	OrderColumn    string
	OrderDesc      bool
	PrefilterField string
	PrefilterValue string
}

type listHeaderItem struct {
	Name         string
	NameHuman    string
	ColumnName   string
	CanOrder     bool
	ShouldShow   bool
	FilterLayout string
	FilterData   interface{}
}

type listContent struct {
	Count      int64
	TotalCount int64
	Rows       []listRow
	Pagination pagination
	Colspan    int64
}

type listRow struct {
	ID      int64
	URL     string
	Items   []listCell
	Actions listItemActions
}

type listCell struct {
	TemplateName string
	Value        string
	URL          string
}

type listItemActions struct {
	VisibleButtons  []buttonData
	ShowOrderButton bool
	MenuButtons     []buttonData
}

type pagination struct {
	Pages []page
}

type page struct {
	Page    int64
	Current bool
}

type listRequest struct {
	Page           int64
	OrderBy        string
	OrderDesc      bool
	PrefilterField string
	PrefilterValue string
	Filter         map[string]string
}

func (resource *Resource) getListHeader(user User) (list list, err error) {
	lang := user.Locale

	list.Colspan = 1
	list.TypeID = resource.ID

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
		headerItem := getListHeaderItem(*v, user)
		if headerItem.ShouldShow {
			list.Colspan++
		}
		list.Header = append(list.Header, headerItem)
	}
	return
}

func getListHeaderItem(v structField, user User) listHeaderItem {
	headerItem := listHeaderItem{
		Name:       v.Name,
		NameHuman:  v.humanName(user.Locale),
		ColumnName: v.ColumnName,
		ShouldShow: v.canShow(),
	}

	headerItem.FilterLayout = v.filterLayout()

	if headerItem.FilterLayout == "filter_layout_relation" {
		if v.Tags["prago-relation"] != "" {
			headerItem.ColumnName = v.Tags["prago-relation"]
		}
	}

	if headerItem.FilterLayout == "filter_layout_boolean" {
		headerItem.FilterData = []string{
			messages.Messages.Get(user.Locale, "yes"),
			messages.Messages.Get(user.Locale, "no"),
		}
	}

	if v.CanOrder {
		headerItem.CanOrder = true
	}

	return headerItem
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
		if sf.Tags["prago-type"] == "relation" {
			return "filter_layout_relation"
		}
		return "filter_layout_number"
	}

	if sf.Typ.Kind() == reflect.Bool {
		return "filter_layout_boolean"
	}

	if sf.Typ == reflect.TypeOf(time.Now()) {
		return "filter_layout_date"
	}

	return ""
}

func (resource *Resource) addFilterToQuery(q Query, filter map[string]string) Query {
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
			q.Where(str, v)
		case "filter_layout_number", "filter_layout_relation":
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
		case "filter_layout_date":
			fields := strings.Split(v, " - ")
			k = strings.Replace(k, "`", "", -1)
			if len(fields) == 2 {
				var str string

				str = fmt.Sprintf("`%s` >= ?", k)
				q.Where(str, fields[0])

				str = fmt.Sprintf("`%s` <= ?", k)
				q.Where(str, fields[1])
			}
		}
	}
	return q
}

func (admin *Administration) prefilterQuery(field, value string) Query {
	ret := admin.Query()
	if field != "" {
		ret = ret.WhereIs(field, value)
	}
	return ret
}

func (resource *Resource) getListContent(admin *Administration, requestQuery *listRequest, user User) (list listContent, err error) {
	q := admin.prefilterQuery(requestQuery.PrefilterField, requestQuery.PrefilterValue)
	if requestQuery.OrderDesc {
		q = q.OrderDesc(requestQuery.OrderBy)
	} else {
		q = q.Order(requestQuery.OrderBy)
	}

	var count int64
	var item interface{}
	resource.newItem(&item)
	countQuery := admin.prefilterQuery(requestQuery.PrefilterField, requestQuery.PrefilterValue)
	countQuery = resource.addFilterToQuery(countQuery, requestQuery.Filter)
	count, err = countQuery.Count(item)
	if err != nil {
		return
	}
	list.Count = count

	var totalCount int64
	resource.newItem(&item)
	countQuery = admin.prefilterQuery(requestQuery.PrefilterField, requestQuery.PrefilterValue)
	totalCount, err = countQuery.Count(item)
	if err != nil {
		return
	}
	list.TotalCount = totalCount

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

	q = resource.addFilterToQuery(q, requestQuery.Filter)
	q = q.Offset((currentPage - 1) * resource.Pagination)
	q = q.Limit(resource.Pagination)

	var rowItems interface{}
	resource.newArrayOfItems(&rowItems)
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
		row.URL = resource.GetURL(fmt.Sprintf("%d", row.ID))

		row.Actions = admin.getListItemActions(user, val.Index(i).Interface(), row.ID, *resource)
		list.Rows = append(list.Rows, row)
		list.Colspan = int64(len(row.Items)) + 1
	}

	return
}

func (resource *Resource) valueToCell(admin *Administration, field reflect.StructField, val reflect.Value) (cell listCell) {
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
			resourceName := field.Name
			if field.Tag.Get("prago-relation") != "" {
				resourceName = field.Tag.Get("prago-relation")
			}

			relationResource := resource.Admin.getResourceByName(resourceName)

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
				//TODO: localize
				cell.Value = ifaceItemName.AdminItemName("cs")
				cell.TemplateName = "admin_link"
				cell.URL = fmt.Sprintf("%s/%d", relationResource.ID, item.(int64))
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
