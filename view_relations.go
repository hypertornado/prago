package prago

import (
	"encoding/json"
	"fmt"
)

type viewRelation struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Count          int64
}

func (resource *Resource[T]) getRelationViews(id int64, user *user) (ret []view) {
	for _, v := range resource.data.relations {
		vi := resource.getRelationView(id, v, user)
		if vi != nil {
			ret = append(ret, *vi)
		}
	}
	return
}

func (resource *Resource[T]) getRelationView(id int64, field *relatedField, user *user) *view {
	if !resource.data.app.authorize(user, field.resource.getData().canView) {
		return nil
	}

	filteredCount := field.resource.itemWithRelationCount(field.id, int64(id))

	ret := &view{}

	name := field.listName(user.Locale)
	ret.Name = name
	ret.Subname = messages.ItemsCount(filteredCount, user.Locale)

	ret.Navigation = append(ret.Navigation, tab{
		Name: "☰",
		URL:  field.listURL(int64(id)),
	})

	if resource.data.app.authorize(user, field.resource.getData().canUpdate) {
		ret.Navigation = append(ret.Navigation, tab{
			Name: "+",
			URL:  field.addURL(int64(id)),
		})
	}

	ret.Relation = &viewRelation{
		SourceResource: resource.getData().getID(),
		TargetResource: field.resource.getData().getID(),
		TargetField:    field.id,
		IDValue:        int64(id),
		Count:          filteredCount,
	}
	return ret
}

func (resource *Resource[T]) itemWithRelationCount(fieldID string, id int64) int64 {
	filteredCount, err := resource.Is(fieldID, id).Count()
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

	request.SetData("data", targetResource.getPreviews(listRequest, request.user))
	request.RenderView("admin_item_view_relationlist_response")
}

func (resource *Resource[T]) getPreviews(listRequest relationListRequest, user *user) []*preview {
	sourceResource := resource.data.app.getResourceByID(listRequest.SourceResource)
	if !resource.data.app.authorize(user, sourceResource.getData().canView) {
		panic("cant authorize source resource")
	}

	if !resource.data.app.authorize(user, resource.getData().canView) {
		panic("cant authorize target resource")
	}

	q := resource.Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
	if resource.data.orderDesc {
		q = q.OrderDesc(resource.getData().orderByColumn)
	} else {
		q = q.Order(resource.getData().orderByColumn)
	}

	limit := listRequest.Count
	if limit > 10 {
		limit = 10
	}
	q = q.Limit(limit)
	q.Offset(listRequest.Offset)

	rowItems := q.List()

	var ret []*preview
	for _, item := range rowItems {
		ret = append(
			ret,
			resource.getPreview(item, user, sourceResource),
		)
	}
	return ret
}
