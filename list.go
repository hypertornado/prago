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
	//Field        *Field
	RelatedResourceID string
	FilterData        interface{}
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

func (resourceData *resourceData) getListHeader(user *user) (list list, err error) {
	lang := user.Locale

	list.Colspan = 1
	list.TypeID = resourceData.id
	list.VisibleColumns = resourceData.defaultVisibleFieldsStr(user)
	list.Columns = resourceData.fieldsStr(user)

	list.OrderColumn = resourceData.orderByColumn
	list.OrderDesc = resourceData.orderDesc
	list.Locale = user.Locale

	list.ItemsPerPage = resourceData.defaultItemsPerPage
	list.PaginationData = resourceData.getPaginationData(user)

	list.StatsLimitSelectData = getStatsLimitSelectData(user.Locale)
	list.MultipleActions = resourceData.getMultipleActions(user)

	orderField, ok := resourceData.fieldMap[resourceData.orderByColumn]
	if !ok || !orderField.canOrder {
		err = ErrItemNotFound
		return
	}

	list.Name = resourceData.pluralName(lang)

	if resourceData.orderField != nil {
		list.CanChangeOrder = true
	}
	list.CanExport = resourceData.app.authorize(user, resourceData.canExport)

	for _, v := range resourceData.fields {
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

func (resourceData *resourceData) defaultVisibleFieldsStr(user *user) string {
	ret := []string{}
	for _, v := range resourceData.fields {
		if !v.authorizeView(user) {
			continue
		}
		if v.defaultShow {
			ret = append(ret, v.id)
		}
	}
	r := strings.Join(ret, ",")
	return r
}

func (resourceData *resourceData) fieldsStr(user *user) string {
	ret := []string{}
	for _, v := range resourceData.fields {
		if !v.authorizeView(user) {
			continue
		}
		ret = append(ret, v.id)
	}
	return strings.Join(ret, ",")
}

func (field *Field) getListHeaderItem(user *user) listHeaderItem {
	var relatedResourceID string
	if field.relatedResource != nil {
		relatedResourceID = field.relatedResource.getData().getID()
	}

	headerItem := listHeaderItem{
		Name:              field.fieldClassName,
		NameHuman:         field.name(user.Locale),
		ColumnName:        field.id,
		DefaultShow:       field.defaultShow,
		RelatedResourceID: relatedResourceID,
	}

	headerItem.FilterLayout = field.filterLayout()

	if headerItem.FilterLayout == "filter_layout_boolean" {
		headerItem.FilterData = []string{
			messages.Get(user.Locale, "yes"),
			messages.Get(user.Locale, "no"),
		}
	}

	if headerItem.FilterLayout == "filter_layout_select" {
		fn := field.fieldType.filterLayoutDataSource
		if fn == nil {
			fn = field.fieldType.formDataSource
		}
		headerItem.FilterData = fn(field, user)
	}

	if field.canOrder {
		headerItem.CanOrder = true
	}

	return headerItem
}

func (field *Field) filterLayout() string {
	if field == nil {
		return ""
	}

	if field.fieldType.filterLayoutTemplate != "" {
		return field.fieldType.filterLayoutTemplate
	}

	if field.typ.Kind() == reflect.String &&
		(field.tags["prago-type"] == "" || field.tags["prago-type"] == "text" || field.tags["prago-type"] == "markdown") {
		return "filter_layout_text"
	}

	if field.typ.Kind() == reflect.Int64 || field.typ.Kind() == reflect.Int {
		if field.tags["prago-type"] == "relation" {
			return "filter_layout_relation"
		}
		return "filter_layout_number"
	}

	if field.typ.Kind() == reflect.Bool {
		return "filter_layout_boolean"
	}

	if field.typ == reflect.TypeOf(time.Now()) {
		return "filter_layout_date"
	}

	return ""
}

func (resource *Resource[T]) addFilterParamsToQuery(q *Query[T], params url.Values) *Query[T] {
	filter := map[string]string{}
	for _, v := range resource.data.fieldMap {
		key := v.id
		val := params.Get(key)
		if val != "" {
			filter[key] = val
		}
	}
	return resource.addFilterToQuery(q, filter)
}

func (resource *Resource[T]) addFilterToQuery(q *Query[T], filter map[string]string) *Query[T] {
	for k, v := range filter {
		field := resource.data.fieldMap[k]
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
					q.Is(k, numVal)
				} else {
					q.Where(
						fmt.Sprintf("%s %s ?", field.id, hasPrefix),
						numVal,
					)
				}
			}
		case "filter_layout_relation":
			v = strings.Trim(v, " ")
			numVal, err := strconv.Atoi(v)
			if err == nil {
				q.Is(k, numVal)
			}
		case "filter_layout_boolean":
			switch v {
			case "true":
				q.Is(k, true)
			case "false":
				q.Is(k, false)
			}
		case "filter_layout_select":
			if field.tags["prago-type"] == "file" || field.tags["prago-type"] == "image" || field.tags["prago-type"] == "cdnfile" {
				if v == "true" {
					q.Where(fmt.Sprintf("%s !=''", field.id))
				}
				if v == "false" {
					q.Where(fmt.Sprintf("%s =''", field.id))
				}
				continue
			}
			if v != "" {
				q.Is(k, v)
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
				q.Where(fmt.Sprintf("`%s` >= ?", k), fromStr)
			}
			if toStr != "" {
				q.Where(fmt.Sprintf("`%s` <= ?", k), toStr)
			}
		}
	}
	return q
}

func (res *Resource[T]) getListContent(user *user, params url.Values) (ret listContent, err error) {
	resourceData := res.data
	if !resourceData.app.authorize(user, resourceData.canView) {
		return listContent{}, errors.New("access denied")
	}

	var listHeader list
	listHeader, err = resourceData.getListHeader(user)
	if err != nil {
		return
	}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resourceData.defaultVisibleFieldsStr(user)
	}

	columnsAr := strings.Split(columnsStr, ",")
	columnsMap := map[string]bool{}
	for _, v := range columnsAr {
		columnsMap[v] = true
	}

	orderBy := resourceData.orderByColumn
	if params.Get("_order") != "" {
		orderBy = params.Get("_order")
	}
	orderDesc := resourceData.orderDesc
	if params.Get("_desc") == "true" {
		orderDesc = true
	}
	if params.Get("_desc") == "false" {
		orderDesc = false
	}

	q := res.Query()
	if orderDesc {
		q = q.OrderDesc(orderBy)
	} else {
		q = q.Order(orderBy)
	}

	var count int64
	countQuery := res.Query()
	countQuery = res.addFilterParamsToQuery(countQuery, params)
	count, err = countQuery.Count()
	if err != nil {
		return
	}

	totalCount, _ := res.Query().Count()
	res.data.updateCachedCount()

	if count == totalCount {
		ret.TotalCountStr = messages.ItemsCount(count, user.Locale)
	} else {
		ret.TotalCountStr = fmt.Sprintf("%s z %s", humanizeNumber(count), messages.ItemsCount(totalCount, user.Locale))
	}

	var itemsPerPage = res.data.defaultItemsPerPage
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

	q = res.addFilterParamsToQuery(q, params)
	q = q.Offset((int64(currentPage) - 1) * itemsPerPage)
	q = q.Limit(itemsPerPage)

	rowItems := q.List()

	for _, item := range rowItems {
		row := listRow{}
		itemVal := reflect.ValueOf(item).Elem()

		for _, v := range listHeader.Header {
			if columnsMap[v.ColumnName] {
				fieldVal := itemVal.FieldByName(v.Name)
				var isOrderedBy bool
				if v.ColumnName == orderBy {
					isOrderedBy = true
				}
				row.Items = append(row.Items, valueToListCell(user, res.Field(v.ColumnName), fieldVal, isOrderedBy))
			}
		}

		//TODO: better find id
		row.ID = itemVal.FieldByName("ID").Int()
		row.URL = res.getData().getURL(fmt.Sprintf("%d", row.ID))

		row.Actions = res.getData().getListItemActions(user, item, row.ID)
		row.AllowsMultipleActions = res.data.allowsMultipleActions(user)
		ret.Rows = append(ret.Rows, row)
	}

	if count == 0 {
		ret.Message = messages.Get(user.Locale, "admin_list_empty")
	}
	ret.Colspan = int64(len(columnsMap)) + 1

	if params.Get("_stats") == "true" {
		ret.Stats = res.getListStats(user, params)
	}

	return
}

type listContentJSON struct {
	Content  string
	CountStr string
	StatsStr string
}

func (resource *Resource[T]) getListContentJSON(user *user, params url.Values) (ret *listContentJSON, err error) {
	listData, err := resource.getListContent(user, params)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = resource.data.app.ExecuteTemplate(buf, "admin_list_cells", map[string]interface{}{
		"admin_list": listData,
	})
	if err != nil {
		return nil, err
	}

	var statsStr string
	if listData.Stats != nil {
		bufStats := new(bytes.Buffer)
		err = resource.data.app.ExecuteTemplate(bufStats, "admin_stats", listData.Stats)
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

func valueToListCell(user *user, f *Field, val reflect.Value, isOrderedBy bool) listCell {
	if !f.authorizeView(user) {
		panic(fmt.Sprintf("can't access field '%s'", f.name("en")))
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

func (resourceData *resourceData) getPaginationData(user *user) (ret []listPaginationData) {
	var ints []int64
	var used bool

	for _, v := range []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000} {
		if !used {
			if v == resourceData.defaultItemsPerPage {
				used = true
			}
			if resourceData.defaultItemsPerPage < v {
				used = true
				ints = append(ints, resourceData.defaultItemsPerPage)
			}
		}
		ints = append(ints, v)
	}

	if resourceData.defaultItemsPerPage > ints[len(ints)-1] {
		ints = append(ints, resourceData.defaultItemsPerPage)
	}

	for _, v := range ints {
		var selected bool
		if v == resourceData.defaultItemsPerPage {
			selected = true
		}

		ret = append(ret, listPaginationData{
			Name:     messages.ItemsCount(v, user.Locale),
			Value:    v,
			Selected: selected,
		})
	}

	return
}
