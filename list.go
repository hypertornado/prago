package prago

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type list struct {
	Name                 string
	TypeID               string
	Colspan              int64
	Header               []listHeaderItem
	VisibleColumns       string
	Columns              string
	CanChangeOrder       bool
	CanExport            bool
	OrderColumn          string
	OrderDesc            bool
	Locale               string
	ItemsPerPage         int64
	PaginationData       []listPaginationData
	StatsLimitSelectData []listPaginationData
	MultipleActions      []listMultipleAction
}

type listPaginationData struct {
	Name     string
	Value    int64
	Selected bool
}

type listHeaderItem struct {
	Name         string
	NameHuman    string
	ColumnName   string
	CanOrder     bool
	DefaultShow  bool
	FilterLayout string
	Field        field
	FilterData   interface{}
}

type listContent struct {
	TotalCountStr string
	Rows          []listRow
	Colspan       int64
	Stats         *listStats
	Message       string
	Pagination    pagination
}

type listRow struct {
	ID                    int64
	URL                   string
	Items                 []listCell
	Actions               listItemActions
	AllowsMultipleActions bool
}

type listCell struct {
	OrderedBy     bool
	Template      string
	Value         interface{}
	OriginalValue interface{}
}

type listItemActions struct {
	VisibleButtons  []buttonData
	ShowOrderButton bool
	MenuButtons     []buttonData
}

type pagination struct {
	TotalPages   int64
	SelectedPage int64
}

type listMultipleAction struct {
	ID       string
	Name     string
	IsDelete bool
}

func (resource *resource) getListHeader(user *user) (list list, err error) {
	lang := user.Locale

	list.Colspan = 1
	list.TypeID = resource.id
	list.VisibleColumns = resource.defaultVisibleFieldsStr(user)
	list.Columns = resource.fieldsStr(user)

	list.OrderColumn = resource.orderByColumn
	list.OrderDesc = resource.orderDesc
	list.Locale = user.Locale

	list.ItemsPerPage = resource.defaultItemsPerPage
	list.PaginationData = resource.getPaginationData(user)

	list.StatsLimitSelectData = getStatsLimitSelectData(user.Locale)
	list.MultipleActions = resource.getMultipleActions(user)

	orderField, ok := resource.fieldMap[resource.orderByColumn]
	if !ok || !orderField.CanOrder {
		err = ErrItemNotFound
		return
	}

	list.Name = resource.name(lang)

	if resource.orderField != nil {
		list.CanChangeOrder = true
	}
	list.CanExport = resource.app.authorize(user, resource.canExport)

	for _, v := range resource.fieldArrays {
		if v.authorizeView(user) {
			headerItem := (*v).getListHeaderItem(user)
			if headerItem.DefaultShow {
				list.Colspan++
			}
			list.Header = append(list.Header, headerItem)
		}
	}
	return
}

func (resource *resource) defaultVisibleFieldsStr(user *user) string {
	ret := []string{}
	for _, v := range resource.fieldArrays {
		if !v.authorizeView(user) {
			continue
		}
		if v.DefaultShow {
			ret = append(ret, v.ColumnName)
		}
	}
	r := strings.Join(ret, ",")
	return r
}

func (resource *resource) fieldsStr(user *user) string {
	ret := []string{}
	for _, v := range resource.fieldArrays {
		if !v.authorizeView(user) {
			continue
		}
		ret = append(ret, v.ColumnName)
	}
	return strings.Join(ret, ",")
}

func (field field) getListHeaderItem(user *user) listHeaderItem {
	headerItem := listHeaderItem{
		Name:        field.Name,
		NameHuman:   field.HumanName(user.Locale),
		ColumnName:  field.ColumnName,
		DefaultShow: field.DefaultShow,
		Field:       field,
	}

	headerItem.FilterLayout = field.filterLayout()

	if headerItem.FilterLayout == "filter_layout_boolean" {
		headerItem.FilterData = []string{
			messages.Get(user.Locale, "yes"),
			messages.Get(user.Locale, "no"),
		}
	}

	if headerItem.FilterLayout == "filter_layout_select" {
		fn := headerItem.Field.fieldType.filterLayoutDataSource
		if fn == nil {
			fn = headerItem.Field.fieldType.formDataSource
		}
		headerItem.FilterData = fn(field, user)
	}

	if field.CanOrder {
		headerItem.CanOrder = true
	}

	return headerItem
}

func (field *field) filterLayout() string {
	if field == nil {
		return ""
	}

	if field.fieldType.filterLayoutTemplate != "" {
		return field.fieldType.filterLayoutTemplate
	}

	if field.Typ.Kind() == reflect.String &&
		(field.Tags["prago-type"] == "" || field.Tags["prago-type"] == "text" || field.Tags["prago-type"] == "markdown") {
		return "filter_layout_text"
	}

	if field.Typ.Kind() == reflect.Int64 || field.Typ.Kind() == reflect.Int {
		if field.Tags["prago-type"] == "relation" {
			return "filter_layout_relation"
		}
		return "filter_layout_number"
	}

	if field.Typ.Kind() == reflect.Bool {
		return "filter_layout_boolean"
	}

	if field.Typ == reflect.TypeOf(time.Now()) {
		return "filter_layout_date"
	}

	return ""
}

func (resource *resource) addFilterParamsToQuery(q query, params url.Values) query {
	filter := map[string]string{}
	for _, v := range resource.fieldMap {
		key := v.ColumnName
		val := params.Get(key)
		if val != "" {
			filter[key] = val
		}
	}
	return resource.addFilterToQuery(q, filter)
}

func (resource *resource) addFilterToQuery(q query, filter map[string]string) query {
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
			q.where(str, v)
		case "filter_layout_number":
			var hasPrefix string
			v = strings.Replace(v, " ", "", -1)
			for _, prefix := range []string{">=", "<=", ">", "<"} {
				if strings.HasPrefix(v, prefix) {
					v = v[len(prefix):]
					hasPrefix = prefix
					break
				}
			}
			numVal, err := strconv.Atoi(v)
			if err == nil {
				if hasPrefix == "" {
					q.is(k, numVal)
				} else {
					q.where(
						fmt.Sprintf("%s %s ?", field.ColumnName, hasPrefix),
						numVal,
					)
				}
			}
		case "filter_layout_relation":
			v = strings.Trim(v, " ")
			numVal, err := strconv.Atoi(v)
			if err == nil {
				q.is(k, numVal)
			}
		case "filter_layout_boolean":
			switch v {
			case "true":
				q.is(k, true)
			case "false":
				q.is(k, false)
			}
		case "filter_layout_select":
			if field.Tags["prago-type"] == "file" || field.Tags["prago-type"] == "image" || field.Tags["prago-type"] == "cdnfile" {
				if v == "true" {
					q.where(fmt.Sprintf("%s !=''", field.ColumnName))
				}
				if v == "false" {
					q.where(fmt.Sprintf("%s =''", field.ColumnName))
				}
				continue
			}
			if v != "" {
				q.is(k, v)
			}
		case "filter_layout_date":
			v = strings.Trim(v, " ")
			var fromStr, toStr string
			fields := strings.Split(v, ",")
			if len(fields) == 1 {
				if strings.HasPrefix(v, ",") {
					toStr = fields[0]
				} else {
					fromStr = fields[0]
				}
			}

			if len(fields) == 2 {
				fromStr = fields[0]
				toStr = fields[1]
			}

			k = strings.Replace(k, "`", "", -1)
			if fromStr != "" {
				q.where(fmt.Sprintf("`%s` >= ?", k), fromStr)
			}
			if toStr != "" {
				q.where(fmt.Sprintf("`%s` <= ?", k), toStr)
			}
		}
	}
	return q
}

func (resource *resource) getListContent(user *user, params url.Values) (ret listContent, err error) {

	if !resource.app.authorize(user, resource.canView) {
		return listContent{}, errors.New("access denied")
	}

	var listHeader list
	listHeader, err = resource.getListHeader(user)
	if err != nil {
		return
	}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resource.defaultVisibleFieldsStr(user)
	}

	columnsAr := strings.Split(columnsStr, ",")
	columnsMap := map[string]bool{}
	for _, v := range columnsAr {
		columnsMap[v] = true
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

	q := resource.query()
	if orderDesc {
		q = q.orderDesc(orderBy)
	} else {
		q = q.order(orderBy)
	}

	var count int64
	//var item interface{}
	//resource.newItem(&item)
	countQuery := resource.query()
	countQuery = resource.addFilterParamsToQuery(countQuery, params)
	count, err = countQuery.count()
	if err != nil {
		return
	}

	var totalCount = resource.count()
	resource.updateCachedCount()

	if count == totalCount {
		ret.TotalCountStr = messages.ItemsCount(count, user.Locale)
	} else {
		ret.TotalCountStr = fmt.Sprintf("%s z %s", humanizeNumber(count), messages.ItemsCount(totalCount, user.Locale))
	}

	var itemsPerPage = resource.defaultItemsPerPage
	if params.Get("_pagesize") != "" {
		pageSize, err := strconv.Atoi(params.Get("_pagesize"))
		if err == nil && pageSize > 0 && pageSize <= 1000000 {
			itemsPerPage = int64(pageSize)
		}
	}

	totalPages := (count / itemsPerPage)
	if count%itemsPerPage != 0 {
		totalPages += +1
	}

	currentPage, _ := strconv.Atoi(params.Get("_page"))
	if currentPage < 1 {
		currentPage = 1
	}

	ret.Pagination = pagination{
		TotalPages:   totalPages,
		SelectedPage: int64(currentPage),
	}

	q = resource.addFilterParamsToQuery(q, params)
	q = q.offset((int64(currentPage) - 1) * itemsPerPage)
	q = q.limit(itemsPerPage)

	/*var rowItems interface{}
	resource.newArrayOfItems(&rowItems)
	q.get(rowItems)*/
	//fmt.Println(reflect.TypeOf(rowItems))

	//var rowItems interface{}
	//resource.newArrayOfItems(&rowItems)
	rowItems, err := q.list()
	if err != nil {
		return
	}

	val := reflect.ValueOf(rowItems)
	for i := 0; i < val.Len(); i++ {
		row := listRow{}
		itemVal := val.Index(i).Elem()

		for _, v := range listHeader.Header {
			if columnsMap[v.ColumnName] {
				fieldVal := itemVal.FieldByName(v.Name)
				var isOrderedBy bool
				if v.ColumnName == orderBy {
					isOrderedBy = true
				}
				row.Items = append(row.Items, resource.valueToListCell(user, v.Field, fieldVal, isOrderedBy))
			}
		}

		row.ID = itemVal.FieldByName("ID").Int()
		row.URL = resource.getURL(fmt.Sprintf("%d", row.ID))

		row.Actions = resource.app.getListItemActions(user, val.Index(i).Interface(), row.ID, *resource)
		row.AllowsMultipleActions = resource.allowsMultipleActions(user)
		ret.Rows = append(ret.Rows, row)
	}

	if count == 0 {
		ret.Message = messages.Get(user.Locale, "admin_list_empty")
	}
	ret.Colspan = int64(len(columnsMap)) + 1

	if params.Get("_stats") == "true" {
		ret.Stats = getListStats(resource, user, params)
	}

	return
}

type listContentJSON struct {
	Content  string
	CountStr string
	StatsStr string
}

func (resource *resource) getListContentJSON(user *user, params url.Values) (ret *listContentJSON, err error) {
	listData, err := resource.getListContent(user, params)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = resource.app.ExecuteTemplate(buf, "admin_list_cells", map[string]interface{}{
		"admin_list": listData,
	})
	if err != nil {
		return nil, err
	}

	var statsStr string
	if listData.Stats != nil {
		bufStats := new(bytes.Buffer)
		err = resource.app.ExecuteTemplate(bufStats, "admin_stats", listData.Stats)
		if err != nil {
			return nil, err
		}
		statsStr = bufStats.String()
	}

	return &listContentJSON{
		Content:  buf.String(),
		CountStr: listData.TotalCountStr,
		StatsStr: statsStr,
	}, nil

}

func (resource resource) valueToListCell(user *user, f field, val reflect.Value, isOrderedBy bool) listCell {
	if !f.authorizeView(user) {
		panic(fmt.Sprintf("can't access field '%s'", f.Name))
	}
	var item interface{}
	reflect.ValueOf(&item).Elem().Set(val)
	var cell listCell
	cell.Template = f.fieldType.listCellTemplate
	cell.Value = f.fieldType.listCellDataSource(user, f, item)
	cell.OriginalValue = val.Interface()
	cell.OrderedBy = isOrderedBy
	return cell
}
