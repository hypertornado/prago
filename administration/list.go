package administration

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/hypertornado/prago/administration/messages"
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
	Locale         string
}

type listHeaderItem struct {
	Name         string
	NameHuman    string
	ColumnName   string
	CanOrder     bool
	DefaultShow  bool
	FilterLayout string
	Field        Field
	FilterData   interface{}
}

type listContent struct {
	Count      int64
	TotalCount int64
	Rows       []listRow
	Pagination pagination
	Colspan    int64
	Message    string
}

type listRow struct {
	ID      int64
	URL     string
	Items   []listCell
	Actions listItemActions
}

type listCell struct {
	OrderedBy bool
	Template  string
	Value     interface{}
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
	Columns        map[string]bool
}

func (resource *Resource) getListHeader(user User) (list list, err error) {
	lang := user.Locale

	list.Colspan = 1
	list.TypeID = resource.ID

	list.OrderColumn = resource.OrderByColumn
	list.OrderDesc = resource.OrderDesc
	list.Locale = user.Locale

	orderField, ok := resource.fieldMap[resource.OrderByColumn]
	if !ok || !orderField.CanOrder {
		err = ErrItemNotFound
		return
	}

	list.Name = resource.HumanName(lang)

	if resource.OrderColumnName == list.OrderColumn && !list.OrderDesc {
		list.CanChangeOrder = true
	}

	for _, v := range resource.fieldArrays {
		if defaultVisibilityFilter(*resource, user, *v) {
			headerItem := (*v).getListHeaderItem(user)
			if headerItem.DefaultShow {
				list.Colspan++
			}
			list.Header = append(list.Header, headerItem)
		}
	}
	return
}

func (v Field) getListHeaderItem(user User) listHeaderItem {
	headerItem := listHeaderItem{
		Name:        v.Name,
		NameHuman:   v.HumanName(user.Locale),
		ColumnName:  v.ColumnName,
		DefaultShow: v.shouldShow(),
		Field:       v,
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

func (sf *Field) filterLayout() string {
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
		field := resource.fieldMap[k]
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

func (resource *Resource) getListContent(admin *Administration, requestQuery *listRequest, user User) (ret listContent, err error) {
	var listHeader list
	listHeader, err = resource.getListHeader(user)
	if err != nil {
		return
	}

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
	ret.Count = count

	var totalCount int64
	resource.newItem(&item)
	countQuery = admin.prefilterQuery(requestQuery.PrefilterField, requestQuery.PrefilterValue)
	totalCount, err = countQuery.Count(item)
	if err != nil {
		return
	}
	ret.TotalCount = totalCount

	totalPages := (count / resource.ItemsPerPage)
	if count%resource.ItemsPerPage != 0 {
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

			ret.Pagination.Pages = append(ret.Pagination.Pages, p)
		}
	}

	q = resource.addFilterToQuery(q, requestQuery.Filter)
	q = q.Offset((currentPage - 1) * resource.ItemsPerPage)
	q = q.Limit(resource.ItemsPerPage)

	var rowItems interface{}
	resource.newArrayOfItems(&rowItems)
	q.Get(rowItems)

	val := reflect.ValueOf(rowItems).Elem()
	for i := 0; i < val.Len(); i++ {
		row := listRow{}
		itemVal := val.Index(i).Elem()

		for _, v := range listHeader.Header {
			if requestQuery.Columns[v.ColumnName] {
				fieldVal := itemVal.FieldByName(v.Name)
				row.Items = append(row.Items, resource.valueToCell(user, v.Field, fieldVal, requestQuery))
			}
		}

		row.ID = itemVal.FieldByName("ID").Int()
		row.URL = resource.GetURL(fmt.Sprintf("%d", row.ID))

		row.Actions = admin.getListItemActions(user, val.Index(i).Interface(), row.ID, *resource)
		ret.Rows = append(ret.Rows, row)
	}

	if ret.Count == 0 {
		ret.Message = messages.Messages.Get(user.Locale, "admin_list_empty")
	}
	ret.Colspan = int64(len(requestQuery.Columns)) + 1

	return
}

func (resource Resource) valueToCell(user User, f Field, val reflect.Value, requestQuery *listRequest) listCell {
	var item interface{}
	reflect.ValueOf(&item).Elem().Set(val)

	var cell listCell
	cell.Template = f.fieldType.ListCellTemplate
	cell.Value = f.fieldType.ListCellDataSource(resource, user, f, item)
	if requestQuery.OrderBy == strings.ToLower(f.Name) {
		cell.OrderedBy = true
	}
	return cell
}
