package prago

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/net/context"
)

func (resource *Resource) getListContent(ctx context.Context, userData UserData, params url.Values) (ret listContent, err error) {
	if !userData.Authorize(resource.canView) {
		return listContent{}, errors.New("access denied")
	}

	var listHeader list
	listHeader, err = resource.getListHeader(userData)
	if err != nil {
		return
	}

	columnsStr := params.Get("_columns")
	if columnsStr == "" {
		columnsStr = resource.defaultVisibleFieldsStr(userData)
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

	q := resource.query(ctx)
	if orderDesc {
		q = q.OrderDesc(orderBy)
	} else {
		q = q.Order(orderBy)
	}

	var count int64
	countQuery := resource.query(ctx)
	countQuery = resource.addFilterParamsToQuery(countQuery, params, userData)
	count, err = countQuery.count()
	if err != nil {
		return
	}

	totalCount, _ := resource.query(ctx).count()
	resource.updateCachedCount()

	if count == totalCount {
		ret.TotalCountStr = messages.ItemsCount(count, userData.Locale())
	} else {
		ret.TotalCountStr = fmt.Sprintf("%s z %s", humanizeNumber(count), messages.ItemsCount(totalCount, userData.Locale()))
	}

	var itemsPerPage = resource.defaultItemsPerPage
	if params.Get("_pagesize") != "" {
		pageSize, err := strconv.Atoi(params.Get("_pagesize"))
		if err == nil && pageSize > 0 /*&& pageSize <= 1000000*/ {
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

	q = resource.addFilterParamsToQuery(q, params, userData)
	q = q.Offset((int64(currentPage) - 1) * itemsPerPage)
	q = q.Limit(itemsPerPage)

	rowItems, err := q.list()
	if err != nil {
		panic(err)
	}

	itemVals := reflect.ValueOf(rowItems)
	itemLen := itemVals.Len()
	for i := 0; i < itemLen; i++ {
		row := listRow{}
		itemVal := itemVals.Index(i).Elem()

		pw := resource.previewer(userData, itemVals.Index(i).Interface())
		row.Name = pw.Name()
		row.Description = pw.DescriptionBasic(nil)
		row.ImageURL = pw.ImageURL(context.Background())

		for _, v := range listHeader.Header {
			if columnsMap[v.ColumnName] {
				if resource.Field(v.Name) != nil {
					fieldVal := itemVal.FieldByName(v.Name)
					row.Items = append(row.Items, getCellViewData(userData, resource.Field(v.ColumnName), fieldVal.Interface()))
				} else {
					cell := listCell{}
					for _, stat := range resource.itemStats {
						if stat.id == v.Name {
							cell.Name = "⏳"

							var urlData url.Values = map[string][]string{}
							urlData.Add("resource_id", resource.id)
							urlData.Add("stat_id", stat.id)
							itemID := itemVal.FieldByName("ID").Int()
							urlData.Add("item_id", fmt.Sprintf("%d", itemID))

							cell.FetchURL = "/admin/api/resource-item-stats?" + urlData.Encode()
						}
					}
					row.Items = append(row.Items, cell)
				}
			}
		}

		previewer := resource.previewer(userData, itemVal.Addr().Interface())
		row.ID = previewer.ID()
		row.URL = previewer.URL("")

		row.Actions = resource.getListItemActions(userData, itemVal.Addr().Interface(), row.ID)
		row.AllowsMultipleActions = resource.allowsMultipleActions(userData)
		ret.Rows = append(ret.Rows, row)

	}

	ret.Language = userData.Locale()
	ret.Message = ret.TotalCountStr
	ret.Colspan = int64(len(columnsMap)) + 1

	return
}

type listContentJSON struct {
	Content string
	//CountStr  string
	//StatsStr  string
	FooterStr string
}

func (resource *Resource) getListContentJSON(ctx context.Context, userData UserData, params url.Values) *listContentJSON {
	listData, err := resource.getListContent(ctx, userData, params)
	must(err)

	return &listContentJSON{
		Content: resource.app.adminTemplates.ExecuteToString("list_cells", map[string]interface{}{
			"admin_list": listData,
		}),
		//CountStr: listData.TotalCountStr,
		//StatsStr: statsStr,
		FooterStr: resource.app.adminTemplates.ExecuteToString("list_footer", map[string]interface{}{
			"admin_list": listData,
		}),
	}
}
