package prago

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

//https://caldwell.org/projects/data/city-index

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
	Name              string
	Icon              string
	NameHuman         string
	ColumnName        string
	CanOrder          bool
	DefaultShow       bool
	FilterLayout      string
	FilterContent     template.HTML
	RelatedResourceID string
	FilterData        interface{}
	NaturalCellWidth  int64
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
	ID int64

	Name        string
	Description string
	ImageURL    string

	URL                   string
	Items                 []listCell
	Actions               listItemActions
	AllowsMultipleActions bool
}

type listItemActions struct {
	ShowOrderButton bool
	MenuButtons     []*buttonData
}

func (actions *listItemActions) JSON() template.HTMLAttr {
	b, err := json.Marshal(actions)
	must(err)
	return template.HTMLAttr(b)
}

type pagination struct {
	TotalPages   int64
	SelectedPage int64
}

type listMultipleAction struct {
	ID         string
	ActionType string
	Icon       string
	Name       string
	//IsDelete   bool
}

func (resource *Resource) getListHeader(userData UserData) (list list, err error) {
	lang := userData.Locale()

	list.Colspan = 1
	list.TypeID = resource.id
	list.VisibleColumns = resource.defaultVisibleFieldsStr(userData)
	list.Columns = resource.fieldsStr(userData)

	list.OrderColumn = resource.orderByColumn
	list.OrderDesc = resource.orderDesc
	list.Locale = userData.Locale()

	list.ItemsPerPage = resource.defaultItemsPerPage
	list.PaginationData = resource.getPaginationData(userData)

	list.StatsLimitSelectData = getStatsLimitSelectData(userData.Locale())
	list.MultipleActions = resource.getMultipleActions(userData)

	orderField, ok := resource.fieldMap[resource.orderByColumn]
	if !ok || !orderField.canOrder {
		err = ErrItemNotFound
		return
	}

	list.Name = resource.pluralName(lang)

	if resource.orderField != nil {
		list.CanChangeOrder = true
	}
	list.CanExport = userData.Authorize(resource.canExport)

	for _, v := range resource.fields {
		if v.authorizeView(userData) {
			headerItem := (*v).getListHeaderItem(userData)
			if headerItem.DefaultShow {
				list.Colspan++
			}
			list.Header = append(list.Header, headerItem)
		}
	}

	for k, stat := range resource.itemStats {
		if !userData.Authorize(stat.Permission) {
			continue
		}

		//who just 1 stat in table
		if k != 0 {
			continue
		}

		headerItem := listHeaderItem{
			Name:             stat.id,
			Icon:             "glyphicons-basic-43-stats-circle.svg",
			ColumnName:       stat.id,
			NameHuman:        stat.Name(userData.Locale()),
			CanOrder:         false,
			DefaultShow:      true,
			NaturalCellWidth: 150,
		}
		list.Header = append([]listHeaderItem{headerItem}, list.Header...)
	}

	return
}

func (resource *Resource) defaultVisibleFieldsStr(userData UserData) string {
	ret := []string{}
	for _, v := range resource.fields {
		if !v.authorizeView(userData) {
			continue
		}

		if !v.defaultHidden {
			ret = append(ret, v.id)
		}
	}
	for _, v := range resource.itemStats {
		if userData.Authorize(v.Permission) {
			ret = append(ret, v.id)
		}
	}
	r := strings.Join(ret, ",")
	return r
}

func (resource *Resource) fieldsStr(userData UserData) string {
	ret := []string{}
	for _, v := range resource.fields {
		if !v.authorizeView(userData) {
			continue
		}
		ret = append(ret, v.id)
	}
	return strings.Join(ret, ",")
}

const defaultNaturalCellWidth int64 = 100

func (field *Field) getNaturalCellWidth() int64 {
	ret := defaultNaturalCellWidth
	if field.fieldType.naturalCellWidth > 0 {
		ret = field.fieldType.naturalCellWidth
	}

	if field.fieldType.isRelation() {
		return 150
	}

	switch field.typ {
	case reflect.TypeOf(time.Now()):
		return 150
	case reflect.TypeOf(true):
		return 60
	case reflect.TypeOf(int64(0)):
		return 60
	case reflect.TypeOf(int(0)):
		return 60
	}

	return ret

}

func (field *Field) getListHeaderItem(userData UserData) listHeaderItem {
	var relatedResourceID string
	if field.relatedResource != nil {
		relatedResourceID = field.relatedResource.getID()
	}

	headerItem := listHeaderItem{
		Name:              field.fieldClassName,
		Icon:              field.getIcon(),
		NameHuman:         field.name(userData.Locale()),
		ColumnName:        field.id,
		DefaultShow:       !field.defaultHidden,
		RelatedResourceID: relatedResourceID,
		NaturalCellWidth:  field.getNaturalCellWidth(),
	}

	headerItem.FilterLayout = field.filterLayout()

	if headerItem.FilterLayout == "filter_layout_boolean" {
		headerItem.FilterData = []string{
			messages.Get(userData.Locale(), "yes"),
			messages.Get(userData.Locale(), "no"),
		}
	}

	if headerItem.FilterLayout == "filter_layout_select" {
		fn := field.fieldType.filterLayoutDataSource
		headerItem.FilterData = fn(field, userData)
	}

	if field.canOrder {
		headerItem.CanOrder = true
	}

	if headerItem.FilterLayout != "" {
		headerItem.FilterContent = field.resource.app.adminTemplates.
			ExecuteToHTML(headerItem.FilterLayout, headerItem)
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

	if field.tags["prago-type"] == "multirelation" {
		return "filter_layout_relation"
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

func (resource *Resource) addFilterParamsToQuery(listQuery *listQuery, params url.Values, userData UserData) *listQuery {
	filter := map[string]string{}
	for _, v := range resource.fieldMap {
		if userData.Authorize(v.canView) {
			key := v.id
			val := params.Get(key)
			if val != "" {
				filter[key] = val
			}
		}
	}
	return resource.addFilterToQuery(listQuery, filter)
}

func (resource *Resource) addFilterToQuery(listQuery *listQuery, filter map[string]string) *listQuery {
	for k, v := range filter {
		field := resource.fieldMap[k]
		if field == nil {
			continue
		}

		layout := field.filterLayout()

		switch layout {
		case "filter_layout_text":
			k = strings.Replace(k, "`", "", -1)
			if len(v) > 2 && strings.HasPrefix(v, "\"") && strings.HasSuffix(v, "\"") {
				qStr := v[1 : len(v)-1]
				str := fmt.Sprintf("`%s` = ?", k)
				listQuery.where(str, qStr)
			} else {
				v = "%" + v + "%"
				str := fmt.Sprintf("`%s` LIKE ?", k)
				listQuery.where(str, v)
			}
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
			//TODO: should not return anything where wrong filter
			if err == nil {
				if hasPrefix == "" {
					listQuery.Is(k, numVal)
				} else {
					listQuery.where(
						fmt.Sprintf("%s %s ?", field.id, hasPrefix),
						numVal,
					)
				}
			}
		case "filter_layout_relation":
			v = strings.Trim(v, " ")
			if field.typ.Kind() == reflect.String {
				v = "%;" + v + ";%"
				str := fmt.Sprintf("`%s` LIKE ?", k)
				listQuery.where(str, v)
			} else {
				numVal, err := strconv.Atoi(v)
				if err == nil {
					listQuery.Is(k, numVal)
				}
			}
		case "filter_layout_boolean":
			switch v {
			case "true":
				listQuery.Is(k, true)
			case "false":
				listQuery.Is(k, false)
			}
		case "filter_layout_select":
			if field.tags["prago-type"] == "file" || field.tags["prago-type"] == "image" || field.tags["prago-type"] == "cdnfile" {
				if v == "true" {
					listQuery.where(fmt.Sprintf("%s !=''", field.id))
				}
				if v == "false" {
					listQuery.where(fmt.Sprintf("%s =''", field.id))
				}
				continue
			}
			if v != "" {
				listQuery.Is(k, v)
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
				listQuery.where(fmt.Sprintf("`%s` >= ?", k), fromStr)
			}
			if toStr != "" {
				listQuery.where(fmt.Sprintf("`%s` <= ?", k), toStr)
			}
		}
	}
	return listQuery
}

func (resource *Resource) getPaginationData(userData UserData) (ret []listPaginationData) {
	var ints []int64
	var used bool

	for _, v := range []int64{10, 20, 100, 200, 500, 1000, 2000, 5000, 10000, 20000, 50000, 100000} {
		if !used {
			if v == resource.defaultItemsPerPage {
				used = true
			}
			if resource.defaultItemsPerPage < v {
				used = true
				ints = append(ints, resource.defaultItemsPerPage)
			}
		}
		ints = append(ints, v)
	}

	if resource.defaultItemsPerPage > ints[len(ints)-1] {
		ints = append(ints, resource.defaultItemsPerPage)
	}

	for _, v := range ints {
		var selected bool
		if v == resource.defaultItemsPerPage {
			selected = true
		}

		ret = append(ret, listPaginationData{
			Name:     messages.ItemsCount(v, userData.Locale()),
			Value:    v,
			Selected: selected,
		})
	}

	return
}
