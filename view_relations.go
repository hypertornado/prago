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
	for _, v := range resource.relations {
		vi := resource.getRelationView(id, v, user)
		if vi != nil {
			ret = append(ret, *vi)
		}
	}
	return
}

func (resource *Resource[T]) getRelationView(id int64, field *relatedField, user *user) *view {
	if !resource.app.authorize(user, field.resource.getPermissionView()) {
		return nil
	}

	filteredCount := field.resource.itemWithRelationCount(field.id, int64(id))

	ret := &view{}

	name := field.listName(user.Locale)
	ret.Name = name
	ret.Subname = messages.ItemsCount(filteredCount, user.Locale)

	ret.Navigation = append(ret.Navigation, tab{
		Name: messages.GetNameFunction("admin_table")(user.Locale),
		URL:  field.listURL(int64(id)),
	})

	if resource.app.authorize(user, field.resource.getPermissionUpdate()) {
		ret.Navigation = append(ret.Navigation, tab{
			Name: messages.GetNameFunction("admin_new")(user.Locale),
			URL:  field.addURL(int64(id)),
		})
	}

	ret.Relation = &viewRelation{
		SourceResource: resource.getID(),
		TargetResource: field.resource.getID(),
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
	sourceResource := resource.app.getResourceByID(listRequest.SourceResource)
	if !resource.app.authorize(user, sourceResource.getPermissionView()) {
		panic("cant authorize source resource")
	}

	if !resource.app.authorize(user, resource.getPermissionView()) {
		panic("cant authorize target resource")
	}

	q := resource.Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
	if resource.isOrderDesc() {
		q = q.OrderDesc(resource.getOrderByColumn())
	} else {
		q = q.Order(resource.getOrderByColumn())
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
