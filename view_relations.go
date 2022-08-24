package prago

import (
	"encoding/json"
	"fmt"
	"reflect"
)

type viewRelation struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Count          int64
}

func (resourceData *resourceData) getRelationViews(id int64, user *user) (ret []view) {
	for _, v := range resourceData.relations {
		vi := resourceData.getRelationView(id, v, user)
		if vi != nil {
			ret = append(ret, *vi)
		}
	}
	return
}

func (resourceData *resourceData) getRelationView(id int64, field *relatedField, user *user) *view {
	if !resourceData.app.authorize(user, field.resource.getData().canView) {
		return nil
	}

	filteredCount := field.resource.getData().itemWithRelationCount(field.id, int64(id))

	ret := &view{}

	name := field.listName(user.Locale)
	ret.Name = name
	ret.Subname = messages.ItemsCount(filteredCount, user.Locale)

	ret.Navigation = append(ret.Navigation, tab{
		Name: "☰",
		URL:  field.listURL(int64(id)),
	})

	if resourceData.app.authorize(user, field.resource.getData().canUpdate) {
		ret.Navigation = append(ret.Navigation, tab{
			Name: "+",
			URL:  field.addURL(int64(id)),
		})
	}

	ret.Relation = &viewRelation{
		SourceResource: resourceData.getID(),
		TargetResource: field.resource.getData().getID(),
		TargetField:    field.id,
		IDValue:        int64(id),
		Count:          filteredCount,
	}
	return ret
}

func (resourceData *resourceData) itemWithRelationCount(fieldID string, id int64) int64 {
	filteredCount, err := resourceData.Is(fieldID, id).count()
	if err != nil {
		panic(err)
	}
	return filteredCount
}

type relationListRequest struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Offset         int64
	Count          int64
}

func generateRelationListAPIHandler(request *Request) {
	decoder := json.NewDecoder(request.Request().Body)
	var listRequest relationListRequest
	decoder.Decode(&listRequest)
	defer request.Request().Body.Close()

	targetResource := request.app.getResourceByID(listRequest.TargetResource)

	request.SetData("data", targetResource.getData().getPreviews(listRequest, request.user))
	request.RenderView("admin_item_view_relationlist_response")
}

func (resourceData *resourceData) getPreviews(listRequest relationListRequest, user *user) []*preview {
	sourceResource := resourceData.app.getResourceByID(listRequest.SourceResource)
	if !resourceData.app.authorize(user, sourceResource.getData().canView) {
		panic("cant authorize source resource")
	}

	if !resourceData.app.authorize(user, resourceData.canView) {
		panic("cant authorize target resource")
	}

	q := resourceData.query().Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
	if resourceData.orderDesc {
		q.addOrder(resourceData.orderByColumn, true)
	} else {
		q.addOrder(resourceData.orderByColumn, false)
	}

	limit := listRequest.Count
	if limit > 10 {
		limit = 10
	}
	q = q.Limit(limit)
	q.Offset(listRequest.Offset)

	rowItems, err := q.list()
	must(err)

	itemVals := reflect.ValueOf(rowItems)

	itemLen := itemVals.Len()

	var ret []*preview

	for i := 0; i < itemLen; i++ {
		ret = append(
			ret,
			resourceData.getPreview(itemVals.Index(i).Interface(), user, sourceResource),
		)
	}

	/*for _, item := range rowItems {
		ret = append(
			ret,
			resourceData.getPreview(item, user, sourceResource),
		)
	}*/
	return ret
}
