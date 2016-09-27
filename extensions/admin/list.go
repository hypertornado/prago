package admin

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

type list struct {
	Header     []listHeader
	Rows       []listRow
	Pagination pagination
	Order      bool
	HasDelete  bool
	HasNew     bool
}

type listHeader struct {
	Name        string
	NameHuman   string
	CanOrder    bool
	Ordered     bool
	OrderedDesc bool
	OrderPath   string
}

type listRow struct {
	ID    int64
	Items []listCell
}

type listCell struct {
	TemplateName string
	Value        string
	URL          string
}

type pagination struct {
	Prev  page
	Next  page
	Pages []page
}

type page struct {
	Name    string
	URL     string
	Current bool
}

func (resource *Resource) getList(lang string, path string, requestQuery url.Values) (list list, err error) {
	orderItem := resource.OrderByColumn
	orderDesc := resource.OrderDesc
	isDefaultOrder := true
	wasSomeOrderSet := false

	var qOrder string
	qOrder = requestQuery.Get("order")
	if len(qOrder) > 0 {
		orderItem = qOrder
		orderDesc = false
		wasSomeOrderSet = true
	}

	qOrder = requestQuery.Get("orderdesc")
	if len(qOrder) > 0 {
		orderItem = qOrder
		orderDesc = true
		wasSomeOrderSet = true
	}

	if orderItem != resource.OrderByColumn || orderDesc != resource.OrderDesc {
		isDefaultOrder = false
	}

	if wasSomeOrderSet && isDefaultOrder {
		err = ErrItemNotFound
		return
	}

	orderField, ok := resource.StructCache.fieldMap[orderItem]
	if !ok || !orderField.CanOrder {
		err = ErrItemNotFound
		return
	}

	if (orderItem != resource.OrderByColumn) && !orderField.canShow() {
		err = ErrItemNotFound
		return
	}

	q := resource.Query()
	if orderDesc {
		q.OrderDesc(orderItem)
	} else {
		q.Order(orderItem)
	}

	_, list.HasDelete = resource.Actions["delete"]
	_, list.HasNew = resource.Actions["new"]

	if resource.StructCache.OrderColumnName == orderItem && !orderDesc {
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
			p := page{}
			p.Name = fmt.Sprintf("%d", i)
			if i == currentPage {
				p.Current = true
			}

			p.URL = path
			if i > 1 {
				newURLValues := make(url.Values)
				newURLValues.Set("p", fmt.Sprintf("%d", i))
				if !isDefaultOrder {
					if orderDesc {
						newURLValues.Set("orderdesc", orderItem)
					} else {
						newURLValues.Set("order", orderItem)
					}
				}
				p.URL += "?" + newURLValues.Encode()
			}

			list.Pagination.Pages = append(list.Pagination.Pages, p)
		}
	}

	q.Offset((currentPage - 1) * resource.Pagination)
	q.Limit(resource.Pagination)

	rowItems, err := q.List()

	for _, v := range resource.StructCache.fieldArrays {
		if v.canShow() {
			headerItem := listHeader{
				Name:      v.Name,
				NameHuman: v.humanName(lang),
			}

			if v.CanOrder {
				headerItem.CanOrder = true
				shouldOrderDesc := false

				if orderItem == v.ColumnName {
					headerItem.Ordered = true
					headerItem.OrderedDesc = orderDesc
					if !orderDesc {
						shouldOrderDesc = true
					}
				}

				newURLValues := make(url.Values)
				if currentPage > 1 {
					newURLValues.Set("p", fmt.Sprintf("%d", currentPage))
				}

				if !(v.ColumnName == resource.OrderByColumn && shouldOrderDesc == resource.OrderDesc) {
					if shouldOrderDesc {
						newURLValues.Set("orderdesc", v.ColumnName)
					} else {
						newURLValues.Set("order", v.ColumnName)
					}
				}
				encodedValue := newURLValues.Encode()
				headerItem.OrderPath = path
				if encodedValue != "" {
					headerItem.OrderPath += "?" + newURLValues.Encode()
				}
			}

			list.Header = append(list.Header, headerItem)
		}
	}

	val := reflect.ValueOf(rowItems)
	for i := 0; i < val.Len(); i++ {
		row := listRow{}
		itemVal := val.Index(i).Elem()

		for _, h := range list.Header {
			structField, _ := resource.Typ.FieldByName(h.Name)
			fieldVal := itemVal.FieldByName(h.Name)
			row.Items = append(row.Items, resource.valueToCell(structField, fieldVal))
		}
		row.ID = itemVal.FieldByName("ID").Int()
		list.Rows = append(list.Rows, row)
	}
	return
}

func (resource *Resource) valueToCell(field reflect.StructField, val reflect.Value) (cell listCell) {
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
			relationResource := resource.admin.getResourceByName(field.Name)
			relationItem, err := relationResource.Query().Where(map[string]interface{}{"id": item.(int64)}).First()
			if err != nil {
				panic(err)
			}

			ifaceItemName, ok := relationItem.(interface {
				AdminItemName(string) string
			})
			if ok {
				cell.Value = ifaceItemName.AdminItemName("cs")
				cell.TemplateName = "admin_link"
				cell.URL = fmt.Sprintf("%s/%d", relationResource.ID, item.(int64))
				return
			}

			nameField := reflect.ValueOf(relationItem).Elem().FieldByName("Name")

			cell.Value = nameField.String()
			cell.TemplateName = "admin_link"
			cell.URL = fmt.Sprintf("%s/%d", relationResource.ID, item.(int64))
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
