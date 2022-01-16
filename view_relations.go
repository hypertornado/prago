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

func (resource *resource) getAutoRelationsView(id int, inValues interface{}, user *user) (ret []view) {

	for _, v := range resource.relations {
		if !resource.app.authorize(user, v.resource.canView) {
			continue
		}

		q := v.resource.query().is(v.field, fmt.Sprintf("%d", id))

		filteredCount, err := q.count()
		must(err)

		var vi = view{}

		name := v.listName(user.Locale)
		vi.Name = name
		vi.Subname = messages.ItemsCount(filteredCount, user.Locale)

		vi.Navigation = append(vi.Navigation, tab{
			Name: messages.GetNameFunction("admin_table")(user.Locale),
			URL:  v.listURL(int64(id)),
		})

		if resource.app.authorize(user, v.resource.canUpdate) {
			vi.Navigation = append(vi.Navigation, tab{
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

func generateRelationListAPIHandler(app *App) func(*Request) {
	return func(request *Request) {

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
		if !app.authorize(request.user, sourceResource.canView) {
			panic("cant authorize source resource")
		}

		targetResource := app.getResourceByName(listRequest.TargetResource)
		if !app.authorize(request.user, targetResource.canView) {
			panic("cant authorize target resource")
		}

		q := targetResource.query().is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
		if targetResource.orderDesc {
			q = q.orderDesc(targetResource.orderByColumn)
		} else {
			q = q.order(targetResource.orderByColumn)
		}

		limit := listRequest.Count
		if limit > 10 {
			limit = 10
		}
		q = q.limit(limit)

		q.offset(listRequest.Offset)

		//var rowItems interface{}
		//targetResource.newArrayOfItems(&rowItems)

		rowItems, err := q.list()
		if err != nil {
			panic(err)
		}

		vv := reflect.ValueOf(rowItems)
		var data []interface{}
		for i := 0; i < vv.Len(); i++ {
			data = append(
				data,
				targetResource.itemToRelationData(vv.Index(i).Interface(), request.user, sourceResource),
			)
		}

		request.SetData("data", data)
		request.RenderView("admin_item_view_relationlist_response")
	}
}
