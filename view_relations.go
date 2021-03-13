package prago

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
)

type viewRelation struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Count          int64
}

func (resource *Resource) getAutoRelationsView(id int, inValues interface{}, user User) (ret []view) {

	for _, v := range resource.autoRelations {
		if !resource.app.Authorize(user, v.resource.CanView) {
			continue
		}

		var rowItem interface{}
		v.resource.newItem(&rowItem)

		q := resource.app.Query()
		q = q.WhereIs(v.field, fmt.Sprintf("%d", id))

		filteredCount, err := q.Count(rowItem)
		must(err)

		var vi = view{}

		name := v.listName(user.Locale)
		vi.Name = name
		vi.Subname = messages.ItemsCount(filteredCount, user.Locale)

		vi.Navigation = append(vi.Navigation, navigationTab{
			Name: messages.GetNameFunction("admin_table")(user.Locale),
			URL:  v.listURL(int64(id)),
		})

		if resource.app.Authorize(user, v.resource.CanEdit) {
			vi.Navigation = append(vi.Navigation, navigationTab{
				Name: messages.GetNameFunction("admin_new")(user.Locale),
				URL:  v.addURL(int64(id)),
			})
		}

		vi.Relation = &viewRelation{
			SourceResource: resource.id,
			TargetResource: v.resource.id,
			TargetField:    v.field,
			IDValue:        int64(id),
			Count:          filteredCount,
		}

		ret = append(ret, vi)
	}
	return
}

type relationListRequest struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Offset         int64
	Count          int64
}

func generateRelationListAPIHandler(app *App) func(Request) {
	return func(request Request) {

		user := request.GetUser()

		defer request.Request().Body.Close()

		reqData, err := ioutil.ReadAll(request.Request().Body)
		if err != nil {
			panic("relationListAPIHandler parsing json request: " + err.Error())
		}

		var listRequest relationListRequest
		err = json.Unmarshal(reqData, &listRequest)
		if err != nil {
			panic("Unmarshalling " + err.Error())
		}

		sourceResource := app.getResourceByName(listRequest.SourceResource)
		if !app.Authorize(user, sourceResource.CanView) {
			panic("cant authorize source resource")
		}

		targetResource := app.getResourceByName(listRequest.TargetResource)
		if !app.Authorize(user, targetResource.CanView) {
			panic("cant authorize target resource")
		}

		var rowItems interface{}
		targetResource.newArrayOfItems(&rowItems)

		q := app.Query()
		q = q.WhereIs(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
		if targetResource.OrderDesc {
			q = q.OrderDesc(targetResource.OrderByColumn)
		} else {
			q = q.Order(targetResource.OrderByColumn)
		}

		limit := listRequest.Count
		if limit > 10 {
			limit = 10
		}
		q = q.Limit(limit)

		q.Offset(listRequest.Offset)

		err = q.Get(rowItems)
		if err != nil {
			panic(err)
		}

		vv := reflect.ValueOf(rowItems).Elem()
		var data []interface{}
		for i := 0; i < vv.Len(); i++ {
			data = append(
				data,
				targetResource.itemToRelationData(vv.Index(i).Interface(), user, sourceResource),
			)
		}

		request.SetData("data", data)
		request.RenderView("admin_item_view_relationlist_response")
	}
}
