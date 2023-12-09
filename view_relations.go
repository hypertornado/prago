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

func (resourceData *resourceData) getRelationViews(ctx context.Context, id int64, request *Request) (ret []*view) {
	for _, v := range resourceData.relations {
		vi := resourceData.getRelationView(ctx, id, v, request)
		if vi != nil {
			ret = append(ret, vi)
		}
	}
	return
}

func (resourceData *resourceData) getRelationView(ctx context.Context, id int64, field *relatedField, request *Request) *view {
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
		SourceResource: resourceData.getID(),
		TargetResource: field.resource.getID(),
		TargetField:    field.id,
		IDValue:        int64(id),
		Count:          filteredCount,
	}
	return ret
}

func (resourceData *resourceData) itemWithRelationCount(ctx context.Context, fieldID string, id int64) int64 {
	filteredCount, err := resourceData.query(ctx).Is(fieldID, id).count()
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

func (resourceData *resourceData) getPreviews(ctx context.Context, listRequest relationListRequest, request *Request) []*preview {
	sourceResource := resourceData.app.getResourceByID(listRequest.SourceResource)
	if !request.Authorize(sourceResource.canView) {
		panic("cant authorize source resource")
	}

	if !request.Authorize(resourceData.canView) {
		panic("cant authorize target resource")
	}

	q := resourceData.query(ctx).Is(listRequest.TargetField, fmt.Sprintf("%d", listRequest.IDValue))
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
			resourceData.previewer(request, itemVals.Index(i).Interface()).Preview(ctx, sourceResource),
		)
	}

	return ret
}
