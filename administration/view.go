package administration

import (
	"fmt"
	"reflect"
	"time"

	"github.com/hypertornado/prago/administration/messages"
)

type view struct {
	Name       string
	Subname    string
	Navigation []navigationTab
	Items      []viewField
}

type viewField struct {
	Name     string
	Button   *viewAction
	Template string
	Value    interface{}
}

type viewAction struct {
	Name string
	URL  string
}

func (resource Resource) getViews(id int, inValues interface{}, user User) (ret []view) {
	ret = append(ret, resource.getBasicView(id, inValues, user))
	ret = append(ret, resource.getAutoRelationsView(id, inValues, user)...)
	return ret
}

func (resource Resource) getBasicView(id int, inValues interface{}, user User) view {
	visible := defaultVisibilityFilter
	ret := view{}
	for i, f := range resource.fieldArrays {
		if !visible(resource, user, *f) {
			continue
		}

		var ifaceVal interface{}
		reflect.ValueOf(&ifaceVal).Elem().Set(
			reflect.ValueOf(inValues).Elem().Field(i),
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name:     f.HumanName(user.Locale),
				Template: f.fieldType.ViewTemplate,
				Value:    f.fieldType.ViewDataSource(resource, user, *f, ifaceVal),
			},
		)
	}

	historyView := resource.Admin.getHistory(&resource, int64(id))

	if len(historyView.Items) > 0 {
		ret.Items = append(
			ret.Items,
			viewField{
				Name:     messages.Messages.Get(user.Locale, "admin_history_last"),
				Template: "admin_item_view_text",
				Value:    historyView.Items[0].UserName,
			},
		)

		ret.Items = append(
			ret.Items,
			viewField{
				Name: messages.Messages.Get(user.Locale, "admin_history_count"),
				Button: &viewAction{
					Name: messages.Messages.Get(user.Locale, "admin_history"),
					URL:  resource.GetURL(fmt.Sprintf("%d/history", id)),
				},
				Template: "admin_item_view_text",
				Value:    fmt.Sprintf("%d", len(historyView.Items)),
			},
		)

	}

	return ret
}

/*func (resource Resource) getAutoRelationsView(id int, inValues interface{}, user User) (ret []view) {
	return resource.relationsToView(resource.autoRelations, id, inValues, user)
}*/

func (resource *Resource) getAutoRelationsView(id int, inValues interface{}, user User) (ret []view) {
	for _, v := range resource.autoRelations {
		if !resource.Admin.Authorize(user, v.resource.CanView) {
			continue
		}

		var rowItem interface{}
		v.resource.newItem(&rowItem)

		totalCount, err := resource.Admin.Query().Count(rowItem)
		must(err)

		var rowItems interface{}
		v.resource.newArrayOfItems(&rowItems)

		var vi = view{}
		q := resource.Admin.Query()
		q = q.WhereIs(v.field, fmt.Sprintf("%d", id))
		//q = q.
		//q := resource.Admin.prefilterQuery(v.field, fmt.Sprintf("%d", id))
		if v.resource.OrderDesc {
			q = q.OrderDesc(v.resource.OrderByColumn)
		} else {
			q = q.Order(v.resource.OrderByColumn)
		}

		filteredCount, err := q.Count(rowItem)
		must(err)

		limit := resource.ItemsPerPage
		if limit > 10 {
			limit = 10
		}

		q = q.Limit(limit)
		q.Get(rowItems)

		vv := reflect.ValueOf(rowItems).Elem()
		var data []interface{}
		for i := 0; i < vv.Len(); i++ {
			data = append(
				data,
				v.resource.itemToRelationData(vv.Index(i).Interface(), user, resource),
			)
		}

		name := v.listName(user.Locale)
		vi.Name = name
		vi.Subname = fmt.Sprintf("(%d / %d / %d)", len(data), filteredCount, totalCount)

		vi.Navigation = append(vi.Navigation, navigationTab{
			Name: messages.Messages.GetNameFunction("admin_list")(user.Locale),
			URL:  v.listURL(int64(id)),
		})

		if resource.Admin.Authorize(user, v.resource.CanEdit) {
			vi.Navigation = append(vi.Navigation, navigationTab{
				Name: messages.Messages.GetNameFunction("admin_new")(user.Locale),
				URL:  v.addURL(int64(id)),
			})
		}

		vi.Items = append(
			vi.Items,
			viewField{
				Template: "admin_item_view_relations",
				Value:    data,
			},
		)
		ret = append(ret, vi)
	}
	return ret
}

func getDefaultViewTemplate(t reflect.Type) string {
	return "admin_item_view_text"
}

func getDefaultViewDataSource(f *Field) func(resource Resource, user User, f Field, value interface{}) interface{} {
	t := f.Typ
	if t == reflect.TypeOf(time.Now()) {
		if f.Tags["prago-type"] == "timestamp" || f.Name == "CreatedAt" || f.Name == "UpdatedAt" {
			return timestampViewDataSource
		}
		return timeViewDataSource
	}
	switch t.Kind() {
	case reflect.Bool:
		return boolViewDataSource
	default:
		return defaultViewDataSource
	}
}

func defaultViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return value
}

func timeViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return messages.Messages.Timestamp(
		user.Locale,
		value.(time.Time),
		false,
	)
}

func timestampViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	return messages.Messages.Timestamp(
		user.Locale,
		value.(time.Time),
		true,
	)
}

func boolViewDataSource(resource Resource, user User, f Field, value interface{}) interface{} {
	if value.(bool) {
		return messages.Messages.Get(user.Locale, "yes")
	}
	return messages.Messages.Get(user.Locale, "no")
}
