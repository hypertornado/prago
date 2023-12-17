package prago

import (
	"encoding/json"
	"fmt"
	"reflect"

	"golang.org/x/net/context"
)

type viewRelation struct {
	SourceResource string
	TargetResource string
	TargetField    string
	IDValue        int64
	Count          int64
}

func (resource *Resource) getRelationViews(ctx context.Context, id int64, request *Request) (ret []*view) {
	for _, v := range resource.relations {
		vi := resource.getRelationView(ctx, id, v, request)
		if vi != nil {
			ret = append(ret, vi)
		}
	}
	return
}

func (resource *Resource) getRelationView(ctx context.Context, id int64, field *relatedField, request *Request) *view {
	if !request.Authorize(field.resource.canView) {
		return nil
	}

	filteredCount := field.resource.itemWithRelationCount(ctx, field.id, int64(id))

	ret := &view{}

	icon := iconResource
	if field.resource.icon != "" {
		icon = field.resource.icon
	}
	ret.Icon = icon

	name := field.listName(request.Locale())
	ret.Name = name
	ret.Subname = messages.ItemsCount(filteredCount, request.Locale())

	ret.Navigation = append(ret.Navigation, viewButton{
		Icon: iconTable,
		URL:  field.listURL(int64(id)),
	})

	if request.Authorize(field.resource.canUpdate) {
		ret.Navigation = append(ret.Navigation, viewButton{
			Icon: iconAdd,
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

func (resource *Resource) itemWithRelationCount(ctx context.Context, fieldID string, id int64) int64 {
	filteredCount, err := resource.query(ctx).Is(fieldID, id).count()
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
	data := targetResource.getPreviews(request.r.Context(), listRequest, request)
	request.WriteHTML(200, "admin_item_view_relationlist_response", data)
}

func (resource *Resource) getPreviews(ctx context.Context, listRequest relationListRequest, request *Request) []*preview {
	sourceResource := resource.app.getResourceByID(listRequest.SourceResource)
	if !request.Authorize(sourceResource.canView) {
		panic("cant authorize source resource")
	}

	if !request.Authorize(resource.canView) {
		panic("cant authorize target resource")
	}

	q := resource.query(ctx).Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
	if resource.orderDesc {
		q.addOrder(resource.orderByColumn, true)
	} else {
		q.addOrder(resource.orderByColumn, false)
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
			resource.previewer(request, itemVals.Index(i).Interface()).Preview(ctx, sourceResource),
		)
	}

	return ret
}
